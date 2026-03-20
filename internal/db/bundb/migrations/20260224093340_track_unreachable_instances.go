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
	"errors"
	"fmt"
	"strings"

	"code.superseriousbusiness.org/gopkg/log"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	dbpkg "code.superseriousbusiness.org/gotosocial/internal/db"
	newmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20260224093340_track_unreachable_instances/newmodel"
	oldmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20260224093340_track_unreachable_instances/oldmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		log.Info(ctx, "migrating instances table, this may take a little while...")

		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

			// Create federation errors table.
			if _, err := tx.
				NewCreateTable().
				Model((*newmodel.FederationError)(nil)).
				Exec(ctx); err != nil {
				return err
			}

			// Index federation errors table.
			//
			// This index allows doing selects by instance ID and
			// type, ID descending (ie., newest errors to oldest).
			if err := createIndex(ctx, tx,
				"federation_errors_instance_id_type_idx",
				"federation_errors",
				dbpkg.BunExpr{
					"?, ?, ? DESC",
					dbpkg.Idents(
						"instance_id",
						"type",
						"id",
					)},
			); err != nil {
				return err
			}

			// Create instance settings table.
			if _, err := tx.
				NewCreateTable().
				Model((*newmodel.InstanceSettings)(nil)).
				Exec(ctx); err != nil {
				return err
			}

			// If there's an instance entry for our
			// own instance, take it out and use it
			// create the instance settings entry.
			host := config.GetHost()
			oldInstance := new(oldmodel.Instance)
			err := tx.
				NewSelect().
				Table("instances").
				Where("? = ?", bun.Ident("domain"), host).
				Scan(ctx, oldInstance)
			if err != nil && !errors.Is(err, dbpkg.ErrNoEntries) {
				return err
			}

			if oldInstance.ID != "" {
				// We had an instance entry stored for our own
				// instance, create an instance settings from it.
				settings := &newmodel.InstanceSettings{
					// Use time for ID so we can potentially do something cool
					// in future like "instance existed since blah blah blah time".
					ID:                     id.NewULIDFromTime(oldInstance.CreatedAt),
					Title:                  oldInstance.Title,
					ShortDescription:       oldInstance.ShortDescription,
					ShortDescriptionText:   oldInstance.ShortDescriptionText,
					Description:            oldInstance.Description,
					DescriptionText:        oldInstance.DescriptionText,
					CustomCSS:              oldInstance.CustomCSS,
					Terms:                  oldInstance.Terms,
					TermsText:              oldInstance.TermsText,
					ContactEmail:           oldInstance.ContactEmail,
					ContactAccountUsername: oldInstance.ContactAccountUsername,
					ContactAccountID:       oldInstance.ContactAccountID,
				}
				if _, err := tx.
					NewInsert().
					Model(settings).
					Exec(ctx); err != nil {
					return err
				}

				// Remove this entry from the existing instances
				// table, as we don't want it affecting the count
				// of instances we need to update in a minute.
				if _, err := tx.
					NewDelete().
					Table("instances").
					Where("? = ?", bun.Ident("id"), oldInstance.ID).
					Exec(ctx); err != nil {
					return err
				}
			}

			var (
				// ID for paging.
				maxID string = id.Highest

				// Batch size for
				// selecting + updating.
				batchsz = 100

				// Number of instances
				// updated so far.
				updated int
			)

			// Create the new instances table.
			if _, err := tx.
				NewCreateTable().
				ModelTableExpr("new_instances").
				Model((*newmodel.Instance)(nil)).
				Exec(ctx); err != nil {
				return err
			}

			// Count number of instances we need to update.
			// This will exclude our own instance entry.
			total, err := tx.
				NewSelect().
				Table("instances").
				Count(ctx)
			if err != nil && !errors.Is(err, dbpkg.ErrNoEntries) {
				return err
			}

			for {
				// Select old instances.
				oldInstances := make([]*oldmodel.Instance, 0, batchsz)
				if err := tx.
					NewSelect().
					Model(&oldInstances).
					Where("? < ?", bun.Ident("id"), maxID).
					OrderExpr("? DESC", bun.Ident("id")).
					Limit(batchsz).
					Scan(ctx); err != nil {
					return err
				}

				l := len(oldInstances)
				if len(oldInstances) == 0 {
					// Nothing left
					// to update.
					break
				}

				// Convert old model
				// instances into new ones.
				newInstances := make([]*newmodel.Instance, 0, l)
				for _, oldInstance := range oldInstances {
					newInstances = append(newInstances, &newmodel.Instance{
						// Use ID from time instead of random ID like
						// previously, so we can sort by first-seen.
						ID:     id.NewULIDFromTime(oldInstance.CreatedAt),
						Domain: oldInstance.Domain,
						// Take just software without version, as that
						// changes so it's kinda pointless storing it.
						Software: strings.Split(oldInstance.Version, " ")[0],
					})
				}

				// Insert this batch of instances.
				// We don't care about return values.
				res, err := tx.
					NewInsert().
					Model(&newInstances).
					ModelTableExpr("new_instances").
					Returning("").
					Exec(ctx)
				if err != nil {
					return err
				}

				// Add rows affected to updated count.
				rowsAffected, err := res.RowsAffected()
				if err != nil {
					return err
				}
				updated += int(rowsAffected)
				if updated == total {
					// Done.
					break
				}

				// Set next page.
				maxID = oldInstances[l-1].ID

				// Log helpful message to admin.
				log.Infof(ctx,
					"migrated %d of %d instances (next page will be from %s)",
					updated, total, maxID,
				)
			}

			if total != int(updated) {
				// Return error here in order to rollback the whole transaction.
				return fmt.Errorf("total=%d does not match updated=%d", total, updated)
			}

			log.Infof(ctx, "finished migrating %d instances", total)

			// Drop the old table.
			log.Info(ctx, "dropping old instances table")
			if _, err := tx.
				NewDropTable().
				Table("instances").
				Exec(ctx); err != nil {
				return err
			}

			// Rename new table to old table.
			log.Info(ctx, "renaming new instances table")
			if _, err := tx.
				ExecContext(
					ctx,
					"ALTER TABLE ? RENAME TO ?",
					bun.Ident("new_instances"),
					bun.Ident("instances"),
				); err != nil {
				return err
			}

			if tx.Dialect().Name() == dialect.PG {
				log.Info(ctx, "moving postgres constraints from old table to new table")

				type spec struct {
					old     string
					new     string
					columns []string
				}

				// Rename uniqueness constraints from
				// "new_instances_*" to "instances_*".
				for _, spec := range []spec{
					{
						old:     "new_instances_pkey",
						new:     "instances_pkey",
						columns: []string{"id"},
					},
					{
						old:     "new_instances_domain_key",
						new:     "instances_domain_key",
						columns: []string{"domain"},
					},
				} {
					if _, err := tx.ExecContext(
						ctx,
						"ALTER TABLE ? DROP CONSTRAINT IF EXISTS ?",
						bun.Ident("instances"),
						bun.Safe(spec.old),
					); err != nil {
						return err
					}

					if _, err := tx.ExecContext(
						ctx,
						"ALTER TABLE ? ADD CONSTRAINT ? UNIQUE(?)",
						bun.Ident("instances"),
						bun.Safe(spec.new),
						bun.Safe(strings.Join(spec.columns, ",")),
					); err != nil {
						return err
					}
				}
			}

			return nil
		})
	}

	down := func(ctx context.Context, db *bun.DB) error {
		return nil
	}

	if err := Migrations.Register(up, down); err != nil {
		panic(err)
	}
}
