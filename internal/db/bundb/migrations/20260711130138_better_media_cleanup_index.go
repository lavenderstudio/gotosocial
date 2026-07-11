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
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

			// Drop existing media "cleanup" index which unfortunately isn't doing much.
			if err := dropIndex(ctx, tx, "media_attachments_cleanup_idx"); err != nil {
				return err
			}

			// Create new selected indices on media
			// attachments and emojis on ID specifically
			// where remote indicating columns are set.
			for _, index := range []struct {
				name  string
				table string
				cols  dbpkg.BunExpr
				where []dbpkg.BunExpr
			}{
				{
					name:  "media_attachments_remotes_idx",
					table: "media_attachments",
					cols:  dbpkg.BunExpr{"? DESC", dbpkg.Idents("id")},
					where: []dbpkg.BunExpr{
						{"? IS NOT NULL", dbpkg.Idents("remote_url")},
					},
				},
				{
					name:  "emojis_remotes_idx",
					table: "emojis",
					cols:  dbpkg.BunExpr{"? DESC", dbpkg.Idents("id")},
					where: []dbpkg.BunExpr{
						{"? IS NOT NULL", dbpkg.Idents("domain")},
					},
				},
			} {
				if err := createIndex(ctx, tx,
					index.name,
					index.table,
					index.cols,
					index.where...,
				); err != nil {
					return err
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
