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

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"

	"code.superseriousbusiness.org/gotosocial/internal/db/postgres"
	"code.superseriousbusiness.org/gotosocial/internal/db/sqlite"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		switch db.Dialect().Name() {
		case dialect.SQLite:
			return sqlite.CreateFts5Tables(ctx, db)
		case dialect.PG:
			// Load the pg_trgm extension for trigram-based tokenizing for search indices.
			if _, err := db.ExecContext(ctx, "CREATE EXTENSION pg_trgm;"); err != nil {
				return gtserror.Newf("gotosocial requires pg_trgm to be available: %w", err)
			}
			return postgres.CreateFtsIndices(ctx, db)
		default:
			panic("unreachable")
		}
	}

	down := func(ctx context.Context, db *bun.DB) error {
		return nil
	}

	if err := Migrations.Register(up, down); err != nil {
		panic(err)
	}
}
