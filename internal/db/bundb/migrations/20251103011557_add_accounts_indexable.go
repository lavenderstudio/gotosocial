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
	"reflect"

	gtsmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20251103011557_add_accounts_indexable"

	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

			// Generate new Account.Indexable column definition from bun.
			accountType := reflect.TypeOf((*gtsmodel.Account)(nil))
			colDef, err := getBunColumnDef(tx, accountType, "Indexable")
			if err != nil {
				return err
			}

			// Add column to Account table. Its default of off is safe.
			if _, err = tx.NewAddColumn().
				Model((*gtsmodel.Account)(nil)).
				ColumnExpr(colDef).
				Exec(ctx); // nocollapse
			err != nil {
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
