// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package migrations

import (
	"context"

	dbpkg "code.superseriousbusiness.org/gotosocial/internal/db"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"

	// we haven't changed anything on the status model in regards to the
	// database since the last migration, but we still need a snapshot so
	// just use the status model used in the previous migtration here.
	gtsmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20260221171254_add_flags_column/new"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		// Drop existing indices we're
		// going to be recreating below.
		for _, index := range []string{
			"statuses_public_timeline_idx",
			"statuses_profile_web_view_idx",
			"statuses_profile_web_view_including_boosts_idx",
		} {
			if err := dropIndex(ctx, db, index); err != nil {
				return err
			}
		}

		// Recreate some more of the indices,
		// further narrowing down some existing
		// ones, and some new additional indices.
		for _, index := range []struct {
			Name     string
			Cols     dbpkg.BunExpr
			Where    []dbpkg.BunExpr
			Sqlite   bool
			Postgres bool
		}{
			{
				/*
					Confirmed working here:
					sqlite> CREATE INDEX "statuses_home_timeline_idx" ON "statuses" ("account_id", "id" DESC) WHERE "flags" & 32 = 0 AND "flags" & 2 = 0;
					sqlite> ANALYZE;
					sqlite> EXPLAIN QUERY PLAN WITH "_data" ("account_id") AS (VALUES (...) AND ("status"."id" > '01KKV73XVGWZQJHMT0CWGPYNTM') ORDER BY "status"."id" DESC LIMIT 25
					QUERY PLAN
					|--CO-ROUTINE _data
					|  `--SCAN 626 CONSTANT ROWS
					|--SCAN _data
					|--SEARCH status USING INDEX statuses_home_timeline_idx (account_id=? AND id>? AND id<?)
					`--USE TEMP B-TREE FOR ORDER BY

					credit to cdn0x12 for this!
				*/
				Name: "statuses_home_timeline_idx",
				Cols: dbpkg.BunExpr{
					"?, ? DESC",
					dbpkg.Idents(
						"account_id",
						"id",
					),
				},
				Where: []dbpkg.BunExpr{

					// i.e. "pending_approval" = false
					{"? & ? = 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagPendingApproval}},

					// i.e. "deleted" = false
					{"? & ? = 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagDeleted}},
				},

				// This index works for both SQLite
				// and Postgres, though depending
				// on query size sometimes SQLite
				// prefers statuses_account_id_id_idx.
				Postgres: true,
				Sqlite:   true,
			},

			{
				/*
					Confirmed working here:
					sqlite> CREATE INDEX "statuses_local_timeline_idx" ON "statuses" ("visibility", "id" DESC) WHERE "boost_of_id" IS NULL AND "visibility" = 2 AND "flags" & 8 != 0 AND "flags" & 32 = 0 AND "flags" & 2 = 0;
					sqlite> ANALYZE;
					sqlite> EXPLAIN QUERY PLAN SELECT "status"."id" FROM "statuses" AS "status" WHERE ("status"."flags" & 8 != 0) AND ("status"."visibility" = 2) AND ("status"."flags" & 32 = 0) AND ("flags" & 2 = 0) AND ("status"."boost_of_id" IS NULL) AND ("status"."id" < '01GB7PBGAK38VAX6DFRBZBH305') ORDER BY "status"."id" DESC LIMIT 100;
					QUERY PLAN
					`--SEARCH status USING INDEX statuses_local_timeline_idx (visibility=? AND id<?)

					credit to cdn0x12 for this!
				*/

				// we don't *need* visibility here as
				// an indexable variable here, but it's
				// the only way to get SQLite to use it.
				Name: "statuses_local_timeline_idx",
				Cols: dbpkg.BunExpr{
					"?, ? DESC",
					dbpkg.Idents(
						"visibility",
						"id",
					),
				},
				Where: []dbpkg.BunExpr{
					{"? IS NULL", dbpkg.Idents("boost_of_id")},
					{"? = ?", []any{bun.Ident("visibility"), gtsmodel.VisibilityPublic}},

					// i.e. "local" = true
					{"? & ? != 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagLocal}},

					// i.e. "pending_approval" = false
					{"? & ? = 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagPendingApproval}},

					// i.e. "deleted" = false
					{"? & ? = 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagDeleted}},
				},

				// Used fine by
				// both indices.
				Postgres: true,
				Sqlite:   true,
			},

			{
				/*
					Confirmed working here:
					sqlite> CREATE INDEX "statuses_profile_web_view_idx" ON "statuses" ("account_id", "visibility", "id" DESC) WHERE "in_reply_to_uri" IS NULL AND ("visibility" = 2 OR "visibility" = 3) AND "boost_of_id" IS NULL AND "flags" & 8 != 0 AND "flags" & 2 = 0 AND "flags" & 16 != 0;
					sqlite> analyze;
					sqlite> EXPLAIN QUERY PLAN SELECT "status"."id" FROM "statuses" AS "status" WHERE ("status"."account_id" = '01F8MH17FWEB39HZJ76B6VXSKF') AND ("status"."visibility" = 2) AND ("status"."in_reply_to_uri" IS NULL) AND ("status"."boost_of_id" IS NULL) AND ("status"."flags" & 8 != 0) AND ("status"."flags" & 2 = 0) AND ("status"."flags" & 16 != 0) ORDER BY "status"."id" DESC LIMIT 20;
					QUERY PLAN
					`--SEARCH status USING INDEX statuses_profile_web_view_idx (account_id=? AND visibility=?)
				*/
				Name: "statuses_profile_web_view_idx",
				Cols: dbpkg.BunExpr{
					"?, ?, ? DESC",
					dbpkg.Idents(
						"account_id",
						"visibility",
						"id",
					)},
				Where: []dbpkg.BunExpr{
					{"? IS NULL", dbpkg.Idents("in_reply_to_uri")},
					{"? IS NULL", dbpkg.Idents("boost_of_id")},

					// We only accept one of two
					// possible visiblities for the
					// profile view, so we can limit
					// the possible index size here.
					{"? = ? OR ? = ?", []any{
						bun.Ident("visibility"), gtsmodel.VisibilityPublic,
						bun.Ident("visibility"), gtsmodel.VisibilityUnlocked,
					}},

					// i.e. "local" = true
					{"? & ? != 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagLocal}},

					// i.e. "federated" = true
					{"? & ? != 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagFederated}},

					// i.e. "deleted" = false
					{"? & ? = 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagDeleted}},
				},

				// Used fine by
				// both indices.
				Postgres: true,
				Sqlite:   true,
			},

			{
				/*
					Confirmed working here:
					sqlite> CREATE INDEX "statuses_profile_web_view_including_boosts_idx" ON "statuses" ("account_id", "visibility", "id" DESC) WHERE "in_reply_to_uri" IS NULL AND ("visibility" = 2 OR "visibility" = 3) AND "flags" & 8 != 0 AND "flags" & 2 = 0 AND "flags" & 16 != 0;
					sqlite> analyze;
					sqlite> EXPLAIN QUERY PLAN SELECT "status"."id" FROM "statuses" AS "status" LEFT JOIN "accounts" AS "boost_of_account" ON "status"."boost_of_account_id" = "boost_of_account"."id" LEFT JOIN "statuses" AS "boost_of" ON "status"."boost_of_id" = "boost_of"."id" WHERE ("status"."account_id" = '01F8MH17FWEB39HZJ76B6VXSKF') AND ("status"."visibility" = 2) AND ("status"."in_reply_to_uri" IS NULL) AND (("status"."boost_of_id" IS NULL) OR (("boost_of"."visibility" = 2) AND ("boost_of"."flags" & 16 != 0) AND ("boost_of_account"."hides_to_public_from_unauthed_web" = FALSE))) AND ("status"."flags" & 8 != 0) AND ("status"."flags" & 2 = 0) AND ("status"."flags" & 16 != 0) ORDER BY "status"."id" DESC LIMIT 20;
					QUERY PLAN
					|--SEARCH status USING INDEX statuses_profile_web_view_including_boosts_idx (account_id=? AND visibility=?)
					|--SEARCH boost_of_account USING INDEX sqlite_autoindex_accounts_1 (id=?) LEFT-JOIN
					`--SEARCH boost_of USING INDEX sqlite_autoindex_statuses_1 (id=?) LEFT-JOIN
					sqlite> EXPLAIN QUERY PLAN SELECT "status"."id" FROM "statuses" AS "status" LEFT JOIN "accounts" AS "boost_of_account" ON "status"."boost_of_account_id" = "boost_of_account"."id" LEFT JOIN "statuses" AS "boost_of" ON "status"."boost_of_id" = "boost_of"."id" WHERE ("status"."account_id" = '01F8MH17FWEB39HZJ76B6VXSKF') AND ("status"."visibility" = 2) AND ("status"."in_reply_to_uri" IS NULL) AND (("status"."boost_of_id" IS NULL) OR (("boost_of"."visibility" = 2) AND ("boost_of"."flags" & 16 != 0) AND ("boost_of_account"."hides_to_public_from_unauthed_web" = FALSE))) AND ("status"."flags" & 8 != 0) AND ("status"."flags" & 2 = 0) AND ("status"."flags" & 16 != 0) AND (("status"."attachments" IS NOT NULL AND "status"."attachments" != 'null' AND "status"."attachments" != '[]') OR ("boost_of"."attachments" IS NOT NULL AND "boost_of"."attachments" != 'null' AND "boost_of"."attachments" != '[]')) ORDER BY "status"."id" DESC LIMIT 20;
					QUERY PLAN
					|--SEARCH status USING INDEX statuses_profile_web_view_including_boosts_idx (account_id=? AND visibility=?)
					|--SEARCH boost_of_account USING INDEX sqlite_autoindex_accounts_1 (id=?) LEFT-JOIN
					`--SEARCH boost_of USING INDEX sqlite_autoindex_statuses_1 (id=?) LEFT-JOIN
				*/
				Name: "statuses_profile_web_view_including_boosts_idx",
				Cols: dbpkg.BunExpr{
					"?, ?, ? DESC",
					dbpkg.Idents(
						"account_id",
						"visibility",
						"id",
					)},
				Where: []dbpkg.BunExpr{
					{"? IS NULL", dbpkg.Idents("in_reply_to_uri")},

					// We only accept one of two
					// possible visiblities for the
					// profile view, so we can limit
					// the possible index size here.
					{"? = ? OR ? = ?", []any{
						bun.Ident("visibility"), gtsmodel.VisibilityPublic,
						bun.Ident("visibility"), gtsmodel.VisibilityUnlocked,
					}},

					// i.e. "local" = true
					{"? & ? != 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagLocal}},

					// i.e. "federated" = true
					{"? & ? != 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagFederated}},

					// i.e. "deleted" = false
					{"? & ? = 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagDeleted}},
				},

				// Used fine by
				// both indices.
				Postgres: true,
				Sqlite:   true,
			},

			{
				/*
					Confirmed working here:
					sqlite> CREATE INDEX "statuses_public_timeline_idx" ON "statuses" ("id" DESC) WHERE "visibility" = 2 AND "boost_of_id" IS NULL AND "flags" & 32 = 0 AND "flags" & 2 = 0;
					sqlite> analyze;
					sqlite> EXPLAIN QUERY PLAN SELECT id FROM statuses WHERE id < 'ZZZZZZZZZZZZZZZZZZZZZZZZZZ' AND id > '00000000000000000000000000' AND visibility = 2 AND boost_of_id IS NULL AND flags & 32 = 0 AND flags & 2 = 0 ORDER BY id DESC;
					QUERY PLAN
					`--SEARCH statuses USING INDEX statuses_public_timeline_idx (id>? AND id<?)

					TODO: test on postgres
				*/
				Name: "statuses_public_timeline_idx",
				Cols: dbpkg.BunExpr{"? DESC", dbpkg.Idents("id")},
				Where: []dbpkg.BunExpr{
					{"? = ?", []any{bun.Ident("visibility"), gtsmodel.VisibilityPublic}},
					{"? IS NULL", dbpkg.Idents("boost_of_id")},

					// i.e. "pending_approval" = false
					{"? & ? = 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagPendingApproval}},

					// i.e. "deleted" = false
					{"? & ? = 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagDeleted}},
				},

				// Used fine by
				// both indices.
				Postgres: true,
				Sqlite:   true,
			},
		} {
			switch d := db.Dialect().Name(); {
			case !index.Sqlite && d == dialect.SQLite:
				// index not required for sqlite

			case !index.Postgres && d == dialect.PG:
				// index not required for postgres

			default:
				// Create the prepared index.
				if err := createIndex(ctx, db,
					index.Name,
					"statuses",
					index.Cols,
					index.Where...,
				); err != nil {
					return err
				}
			}
		}

		// SHRINK 👏 THAT 👏 WAL 👏 !!
		err := doWALCheckpoint(ctx, db)
		return err
	}

	down := func(ctx context.Context, db *bun.DB) error {
		return nil
	}

	if err := Migrations.Register(up, down); err != nil {
		panic(err)
	}
}
