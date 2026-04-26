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
	newmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20260324103226_relay_subscription"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

			// Create new relay tables.
			for _, m := range []any{
				(*newmodel.RelaySubscription)(nil),
				(*newmodel.RelayPush)(nil),
				(*newmodel.RelayMatcher)(nil),
			} {
				if _, err := tx.
					NewCreateTable().
					Model(m).
					Exec(ctx); err != nil {
					return err
				}
			}

			// Index new tables on
			// account_id and relay_actor_uri.
			for _, table := range []string{
				"relay_subscriptions",
				"relay_pushes",
			} {
				for _, column := range []string{
					"account_id",
					"relay_actor_uri",
				} {
					if err := createIndex(ctx, tx,
						table+"_"+column+"_idx",
						table,
						dbpkg.BunExpr{"?", dbpkg.Idents(column)},
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
