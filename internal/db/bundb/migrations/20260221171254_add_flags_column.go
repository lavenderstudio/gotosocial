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
	"database/sql"
	"errors"
	"time"

	"code.superseriousbusiness.org/gopkg/log"
	"code.superseriousbusiness.org/gopkg/xslices"
	newmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20260221171254_add_flags_column/new"
	oldmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20260221171254_add_flags_column/old"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		// Add new flags column to statuses table. Its default of 0 is safe.
		if err := addColumn(ctx, db, (*newmodel.Status)(nil), "Flags"); err != nil {
			return err
		}

		log.Info(ctx, "creating new table: status_pins")

		// Add new StatusPin model table.
		if _, err := db.NewCreateTable().
			Model((*newmodel.StatusPin)(nil)).
			Exec(ctx); err != nil {
			return gtserror.Newf("error adding status_pins table: %w", err)
		}

		// Merge WAL file to minimize size ahead of big tx.
		if err := doWALCheckpoint(ctx, db); err != nil {
			return err
		}

		// Get a total count of all statuses before migration.
		total, err := db.NewSelect().Table("statuses").Count(ctx)
		if err != nil {
			return gtserror.Newf("error getting status table count: %w", err)
		}

		// Start at largest
		// possible ULID value.
		maxID := id.Highest

		log.Warnf(ctx, "migrating status flags to new column, this may take a *long* time")

		// Open initial transaction.
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		// Total updated count.
		var updatedTotal int64

		var statuses []*oldmodel.Status
		var statusIDs []string
		var statusPins []*newmodel.StatusPin
		for i := 1; ; i++ {

			// Reset slices.
			clear(statuses)
			clear(statusIDs)
			clear(statusPins)
			statuses = statuses[:0]
			statusIDs = statusIDs[:0]
			statusPins = statusPins[:0]

			// Mark batch start.
			start := time.Now()

			// Select from statuses.
			if err := tx.NewSelect().
				Model(&statuses).
				Column("id", "account_id", "pinned_at").
				Where("? < ?", bun.Ident("id"), maxID).
				OrderExpr("? DESC", bun.Ident("id")).
				Limit(500).
				Scan(ctx); err != nil && !errors.Is(err, sql.ErrNoRows) {
				return gtserror.Newf("error selecting statuses: %w", err)
			}

			if len(statuses) == 0 {
				// No more statuses!
				//
				// Transaction will be closed
				// after leaving the loop.
				break

			} else if i%200 == 0 {
				// Begin a new transaction every
				// 200 batches (~100,000 statuses),
				// to avoid massive commits.

				// Close existing db transaction.
				if err := tx.Commit(); err != nil {
					return err
				}

				// Merge WAL file to try minimize its size.
				if err := doWALCheckpoint(ctx, db); err != nil {
					return err
				}

				// Start a new db transaction.
				tx, err = db.BeginTx(ctx, nil)
				if err != nil {
					return err
				}
			}

			// Set next maxID value from statuses.
			maxID = statuses[len(statuses)-1].ID

			// Gather all IDs from selected statuses.
			statusIDs = xslices.Gather(statusIDs,
				statuses, func(s *oldmodel.Status) string {
					return s.ID
				})

			// Status IDs as a bun.List() value.
			inStatusIDs := bun.List(statusIDs)

			// Perform an UPDATE query for each new possible
			// status flag bit field value, performing a bitwise
			// OR on "flags" to set the bit for matching WHERE clause.
			for _, q := range []struct {
				Bit   newmodel.StatusFlag
				Where BunExpr
			}{
				{Bit: newmodel.StatusFlagSensitive, Where: BunExpr{"? = true", idents("sensitive")}},
				{Bit: newmodel.StatusFlagLocal, Where: BunExpr{"? = true", idents("local")}},
				{Bit: newmodel.StatusFlagFederated, Where: BunExpr{"? = true", idents("federated")}},
				{Bit: newmodel.StatusFlagPendingApproval, Where: BunExpr{"? = true", idents("pending_approval")}},
			} {
				if _, err := tx.NewUpdate().
					Table("statuses").

					// Only operating on status IDs in selected batch.
					Where("? IN (?)", bun.Ident("id"), inStatusIDs).

					// Updating "flags" via OR to set the current 'bit' flag value.
					Set("? = (?|?)", bun.Ident("flags"), bun.Ident("flags"), q.Bit).

					// Only on given WHERE clause.
					Where(q.Where.Fmt, q.Where.Arg...).
					Exec(ctx); err != nil {
					return gtserror.Newf("error setting \"flags\" value = %s: %w", q.Bit.String(), err)
				}
			}

			// Create new status pins models
			// to replace existing pinned col.
			for _, status := range statuses {
				if !status.PinnedAt.IsZero() {
					statusPins = append(statusPins, &newmodel.StatusPin{
						CreatedAt: status.PinnedAt,
						AccountID: status.AccountID,
						StatusID:  status.ID,
					})
				}
			}

			// Insert new pinned models.
			if _, err := tx.NewInsert().
				Model(&statusPins).
				Exec(ctx); err != nil {
				return gtserror.Newf("error inserting status pins: %w", err)
			}

			// Increment updated total by ID count.
			updatedTotal += int64(len(statuses))

			// Calculate rows / second tx speed.
			timeTaken := time.Since(start).Seconds()
			secsPerRow := float64(timeTaken) / float64(len(statuses))
			rowsPerSec := float64(1) / float64(secsPerRow)

			// Calculate percentage of all statuses updated so far.
			perc := (float64(updatedTotal) / float64(total)) * 100

			log.Infof(ctx, "[~%.2f%% done; ~%.0f rows/s] migrating status flags",
				perc, rowsPerSec)
		}

		// Close the final db transaction.
		if err := tx.Commit(); err != nil {
			return err
		}

		// Merge WAL file to try minimize its size.
		if err := doWALCheckpoint(ctx, db); err != nil {
			return err
		}

		// Drop existing indices that
		// rely on the old column types.
		for _, index := range []string{
			"statuses_account_id_pinned_at_idx",
			"statuses_local_idx",
			"statuses_pending_approval_idx",
			"statuses_polls_scheduler_index",
			"statuses_profile_web_view_idx",
			"statuses_profile_web_view_including_boosts_idx",
			"statuses_public_timeline_idx",
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
		// the appropriate newmodel.StatusFlag bits are set.
		for _, index := range []struct {
			Name  string
			Cols  BunExpr
			Where []BunExpr
		}{
			{
				Name: "statuses_local_idx",
				Cols: BunExpr{"?, ? DESC", idents("visibility", "id")},
				Where: []BunExpr{

					// i.e. "local" = true
					{"? & ? != 0", []any{bun.Ident("flags"), newmodel.StatusFlagLocal}},

					// i.e. "pending_approval" = false
					{"? & ? = 0", []any{bun.Ident("flags"), newmodel.StatusFlagPendingApproval}},
				},
			},

			{
				Name: "statuses_account_id_pending_approval_idx",
				Cols: BunExpr{"?", idents("account_id")},
				Where: []BunExpr{

					// i.e. "pending_approval" = true
					{"? & ? != 0", []any{bun.Ident("flags"), newmodel.StatusFlagPendingApproval}},
				},
			},

			{
				Name: "statuses_polls_scheduler_index",
				Cols: BunExpr{"?", idents("poll_id")},
				Where: []BunExpr{

					// i.e. "local" = true
					{"? & ? != 0", []any{bun.Ident("flags"), newmodel.StatusFlagLocal}},
				},
			},

			{
				Name: "statuses_profile_web_view_idx",
				Cols: BunExpr{"?, ?, ? DESC", idents("account_id", "visibility", "id")},
				Where: []BunExpr{
					{"? IS NULL", idents("boost_of_id")},
					{"? IS NULL", idents("in_reply_to_uri")},

					// i.e. "local" = true
					{"? & ? != 0", []any{bun.Ident("flags"), newmodel.StatusFlagLocal}},

					// i.e. "federated" = true
					{"? & ? != 0", []any{bun.Ident("flags"), newmodel.StatusFlagFederated}},
				},
			},

			{
				Name: "statuses_profile_web_view_including_boosts_idx",
				Cols: BunExpr{"?, ?, ? DESC", idents("account_id", "visibility", "id")},
				Where: []BunExpr{
					{"? IS NULL", idents("in_reply_to_uri")},

					// i.e. "local" = true
					{"? & ? != 0", []any{bun.Ident("flags"), newmodel.StatusFlagLocal}},

					// i.e. "federated" = true
					{"? & ? != 0", []any{bun.Ident("flags"), newmodel.StatusFlagFederated}},
				},
			},

			{
				Name: "statuses_public_timeline_idx",
				Cols: BunExpr{"? DESC", idents("id")},
				Where: []BunExpr{
					{"? = ?", []any{bun.Ident("visibility"), newmodel.VisibilityPublic}},
					{"? IS NULL", idents("boost_of_id")},

					// i.e. "pending_approval" = false
					{"? & ? = 0", []any{bun.Ident("flags"), newmodel.StatusFlagPendingApproval}},
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

		// Create new local statuses count view.
		if _, err := db.NewRaw("CREATE VIEW ? AS "+
			"SELECT COUNT(1) FROM ? "+
			"WHERE (? & ? != 0) "+ // local
			"AND (? & ? = 0) "+ // NOT pending approval
			"AND (? IN (?))",
			bun.Ident("statuses_local_count_view"),
			bun.Ident("statuses"),
			bun.Ident("flags"), newmodel.StatusFlagLocal,
			bun.Ident("flags"), newmodel.StatusFlagPendingApproval,
			bun.Ident("visibility"), bun.List([]newmodel.Visibility{
				newmodel.VisibilityPublic,
				newmodel.VisibilityUnlocked,
				newmodel.VisibilityFollowersOnly,
				newmodel.VisibilityMutualsOnly,
			}),
		).Exec(ctx); err != nil {
			return err
		}

		// Drop unused columns from database.
		for _, field := range []string{
			"Sensitive",
			"PinnedAt",
			"Local",
			"Federated",
			"PendingApproval",
		} {
			if err := dropColumn(ctx, db,
				(*oldmodel.Status)(nil),
				field,
			); err != nil {
				return err
			}

			// WAL merge after each drop to minimize WAL size.
			if err := doWALCheckpoint(ctx, db); err != nil {
				return err
			}
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
