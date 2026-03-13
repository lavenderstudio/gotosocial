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
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"github.com/uptrace/bun"

	// we haven't changed anything on the status model in regards to the
	// database since the last migration, but we still need a snapshot so
	// just use the status model used in the previous migtration here.
	gtsmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20260221171254_add_flags_column/new"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		// Drop existing indices that
		// don't account for deleted flag.
		for _, index := range []string{
			"statuses_local_idx",
			"statuses_polls_scheduler_index",
			"statuses_profile_web_view_idx",
			"statuses_profile_web_view_including_boosts_idx",
			"statuses_public_timeline_idx",

			// we don't recreate this one
			// as i can't even get it currently
			// to use this index, nore create it
			// in any manner that doesn't juse
			// use statuses_account_id_id_idx.
			"statuses_account_view_idx",
		} {
			if err := dropIndex(ctx, db, index); err != nil {
				return err
			}
		}

		// Drop the existing status
		// view counting local-only statuses.
		if _, err := db.NewRaw("DROP VIEW ?",
			bun.Ident("statuses_local_count_view")).
			Exec(ctx); err != nil {
			return gtserror.Newf("error dropping statuses_local_count_view: %w", err)
		}

		// Recreate each of the indices using bitwise AND
		// operations on the new "flags" column to check if
		// the appropriate gtsmodel.StatusFlag bits are set,
		// now accounting for deleted status flags.
		for _, index := range []struct {
			Name  string
			Cols  dbpkg.BunExpr
			Where []dbpkg.BunExpr
		}{
			{
				/*
					Confirmed working here:
					sqlite> CREATE INDEX "statuses_local_idx" ON "statuses" ("visibility", "id" DESC) WHERE "flags" & 8 != 0 AND "flags" & 32 = 0 AND "flags" & 2 = 0;
					sqlite> EXPLAIN QUERY PLAN SELECT id FROM statuses WHERE boost_of_id IS NULL AND id < 'ZZZZZZZZZZZZZZZZZZZZZZZZZZ' AND id > '00000000000000000000000000' AND visibility = 2 AND flags & 8 != 0 AND flags & 32 = 0 AND flags & 2 = 0 ORDER BY id DESC;
					QUERY PLAN
					`--SEARCH statuses USING INDEX statuses_local_idx (visibility=? AND id>? AND id<?)
				*/
				Name: "statuses_local_idx",
				Cols: dbpkg.BunExpr{
					"?, ? DESC",
					dbpkg.Idents(
						"visibility",
						"id",
					),
				},
				Where: []dbpkg.BunExpr{

					// i.e. "local" = true
					{"? & ? != 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagLocal}},

					// i.e. "pending_approval" = false
					{"? & ? = 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagPendingApproval}},

					// i.e. "deleted" = false
					{"? & ? = 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagDeleted}},
				},
			},

			{
				/*
					Confirmed working here:
					sqlite> CREATE INDEX "statuses_polls_scheduler_index" ON "statuses" ("poll_id") WHERE "flags" & 8 != 0 AND "flags" & 2 = 0;
					sqlite> EXPLAIN QUERY PLAN SELECT "polls"."id" FROM "polls" JOIN "statuses" ON "polls"."id" = "statuses"."poll_id" WHERE ("statuses"."flags" & 8 != 0) AND ("statuses"."flags" & 2 = 0) AND ("polls"."expires_at" IS NOT NULL) AND ("polls"."closed_at" IS NULL);
					QUERY PLAN
					|--SCAN polls
					`--SEARCH statuses USING INDEX statuses_polls_scheduler_index (poll_id=?)
				*/
				Name: "statuses_polls_scheduler_index",
				Cols: dbpkg.BunExpr{"?", dbpkg.Idents("poll_id")},
				Where: []dbpkg.BunExpr{

					// i.e. "local" = true
					{"? & ? != 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagLocal}},

					// i.e. "deleted" = false
					{"? & ? = 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagDeleted}},
				},
			},

			{
				/*
					Confirmed working here:
					sqlite> CREATE INDEX "statuses_profile_web_view_idx" ON "statuses" ("account_id", "visibility", "id" DESC) WHERE "in_reply_to_uri" IS NULL AND "boost_of_id" IS NULL AND "flags" & 8 != 0 AND "flags" & 2 = 0 AND "flags" & 16 != 0;
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

					// i.e. "local" = true
					{"? & ? != 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagLocal}},

					// i.e. "federated" = true
					{"? & ? != 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagFederated}},

					// i.e. "deleted" = false
					{"? & ? = 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagDeleted}},
				},
			},

			{
				/*
					Confirmed working here:
					sqlite> CREATE INDEX "statuses_profile_web_view_including_boosts_idx" ON "statuses" ("account_id", "visibility", "id" DESC) WHERE "in_reply_to_uri" IS NULL AND "flags" & 8 != 0 AND "flags" & 2 = 0 AND "flags" & 16 != 0;
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

					// i.e. "local" = true
					{"? & ? != 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagLocal}},

					// i.e. "federated" = true
					{"? & ? != 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagFederated}},

					// i.e. "deleted" = false
					{"? & ? = 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagDeleted}},
				},
			},

			{
				/*
					Confirmed working here:
					sqlite> CREATE INDEX "statuses_public_timeline_idx" ON "statuses" ("visibility", "id" DESC) WHERE "visibility" = 2 AND "boost_of_id" IS NULL AND "flags" & 32 = 0 AND "flags" & 2 = 0;
					sqlite> EXPLAIN QUERY PLAN SELECT id FROM statuses WHERE id < 'ZZZZZZZZZZZZZZZZZZZZZZZZZZ' AND id > '00000000000000000000000000' AND visibility = 2 AND boost_of_id IS NULL AND flags & 32 = 0 AND
					flags & 2 = 0 ORDER BY id DESC;
					QUERY PLAN
					`--SEARCH statuses USING INDEX statuses_public_timeline_idx (visibility=? AND id>? AND id<?)
				*/
				Name: "statuses_public_timeline_idx",
				Cols: dbpkg.BunExpr{
					"?, ? DESC",
					dbpkg.Idents(
						"visibility",
						"id",
					),
				},
				Where: []dbpkg.BunExpr{
					{"? = ?", []any{bun.Ident("visibility"), gtsmodel.VisibilityPublic}},
					{"? IS NULL", dbpkg.Idents("boost_of_id")},

					// i.e. "pending_approval" = false
					{"? & ? = 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagPendingApproval}},

					// i.e. "deleted" = false
					{"? & ? = 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagDeleted}},
				},
			},

			{
				/*
					Confirmed working here:
					sqlite> CREATE INDEX "statuses_deleted_idx" ON "statuses" ("id" DESC) WHERE "flags" & 2 != 0;
					sqlite> EXPLAIN QUERY PLAN SELECT id FROM statuses WHERE id < 'ZZZZZZZZZZZZZZZZZZZZZZZZZZ' AND id > '00000000000000000000000000' AND (SELECT COUNT(1) FROM "statuses" AS "sub" WHERE "statuses"."id" = "sub"."in_reply_to_id") = 0 AND flags & 2 != 0 ORDER BY id DESC;
					QUERY PLAN
					|--SEARCH statuses USING INDEX statuses_deleted_idx (id>? AND id<?)
					`--CORRELATED SCALAR SUBQUERY 1
					   `--SEARCH sub USING COVERING INDEX statuses_in_reply_to_id_idx (in_reply_to_id=?)
				*/
				Name: "statuses_deleted_idx",
				Cols: dbpkg.BunExpr{"? DESC", dbpkg.Idents("id")},
				Where: []dbpkg.BunExpr{

					// i.e. "deleted" = true
					{"? & ? != 0", []any{bun.Ident("flags"), gtsmodel.StatusFlagDeleted}},
				},
			},
		} {
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

		// Create new local statuses count view,
		// not taking into account deleted statuses.
		if _, err := db.NewRaw("CREATE VIEW ? AS "+
			"SELECT COUNT(1) FROM ? "+
			"WHERE (? & ? != 0) "+ // local
			"AND (? & ? = 0) "+ // NOT pending approval
			"AND (? & ? = 0) "+ // NOT deleted
			"AND (? IN (?))",
			bun.Ident("statuses_local_count_view"),
			bun.Ident("statuses"),
			bun.Ident("flags"), gtsmodel.StatusFlagLocal,
			bun.Ident("flags"), gtsmodel.StatusFlagPendingApproval,
			bun.Ident("flags"), gtsmodel.StatusFlagDeleted,
			bun.Ident("visibility"), bun.List([]gtsmodel.Visibility{
				gtsmodel.VisibilityPublic,
				gtsmodel.VisibilityUnlocked,
				gtsmodel.VisibilityFollowersOnly,
				gtsmodel.VisibilityMutualsOnly,
			}),
		).Exec(ctx); err != nil {
			return err
		}

		return nil
	}

	down := func(ctx context.Context, db *bun.DB) error {
		return nil
	}

	if err := Migrations.Register(up, down); err != nil {
		panic(err)
	}
}
