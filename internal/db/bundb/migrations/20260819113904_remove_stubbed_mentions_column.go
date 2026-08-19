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
		var model *gtsmodel.Status
		if _, err := db.NewUpdate().
			Model(model).
			Where(dbpkg.BitIsSet("flags", gtsmodel.StatusFlagDeleted)).
			Where(dbpkg.WhereNotArrayIsNullOrEmpty(db, "mentions")).
			Set("? = NULL", bun.Ident("mentions")).
			Exec(ctx); err != nil {
			return gtserror.Newf("error dropping mentions from stubbed statuses: %w", err)
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
