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

	"code.superseriousbusiness.org/gopkg/log"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

			log.Info(ctx, "creating instance directory indexes, this may take a little while...")

			// Create index for local accounts only.
			//
			// This should be very small as it's a partial
			// index on '"domain" IS NULL' so it ignores
			// huge amounts of the accounts table.
			if _, err := tx.NewCreateIndex().
				Table("accounts").
				Index("accounts_directory_local_only_idx").
				Column(
					"domain",
					"discoverable",
					"suspended_at",
				).
				ColumnExpr("? DESC", bun.Ident("created_at")).
				Where("? IS NULL", bun.Ident("domain")).
				Where("? = ?", bun.Ident("discoverable"), true).
				Where("? IS NULL", bun.Ident("suspended_at")).
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			// Create index on account stats
			// last_status_at, as this is
			// needed for ordering.
			if _, err := tx.NewCreateIndex().
				Table("account_stats").
				Index("account_stats_last_status_at_idx").
				ColumnExpr("? DESC", bun.Ident("last_status_at")).
				Where("? IS NOT NULL", bun.Ident("last_status_at")).
				IfNotExists().
				Exec(ctx); err != nil {
				return err
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
