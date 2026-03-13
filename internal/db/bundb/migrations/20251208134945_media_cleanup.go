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
	newmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20251208134945_media_cleanup/newmodel"
	oldmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20251208134945_media_cleanup/oldmodel"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

			// Add new error columns to the database.
			for model, field := range map[any]string{
				(*newmodel.MediaAttachment)(nil): "Error",
				(*newmodel.Emoji)(nil):           "Error",
			} {
				if err := addColumn(ctx, tx, model, field); err != nil {
					return err
				}
			}

			// Drop old media cleanup index that relies on below dropped columns.
			if err := dropIndex(ctx, tx, "media_attachments_cleanup_idx"); err != nil {
				return err
			}

			// Create new media cleanup index.
			if err := createIndex(ctx, tx,
				"media_attachments_cleanup_idx",
				"media_attachments",
				dbpkg.BunExpr{"?, ?", []any{bun.Ident("file_path"), bun.Ident("thumbnail_path")}},
			); err != nil {
				return err
			}

			// Unset all file paths for media
			// attachments that are uncached,
			// to match new caching strategy.
			if _, err := tx.NewUpdate().
				Table("media_attachments").
				Where("? IS NULL OR ? = false", bun.Ident("cached"), bun.Ident("cached")).
				Set("? = ?", bun.Ident("thumbnail_path"), "").
				Set("? = ?", bun.Ident("file_path"), "").
				Exec(ctx); err != nil {
				return gtserror.Newf("error updating uncached media: %w", err)
			}

			// Unset all file paths for emoji
			// attachments that are uncached,
			// to match new caching strategy.
			if _, err := tx.NewUpdate().
				Table("emojis").
				Where("? IS NULL OR ? = false", bun.Ident("cached"), bun.Ident("cached")).
				Set("? = ?", bun.Ident("image_static_path"), "").
				Set("? = ?", bun.Ident("image_path"), "").
				Exec(ctx); err != nil {
				return gtserror.Newf("error updating uncached emojis: %w", err)
			}

			// Drop (now) unused columns from database.
			for model, fields := range map[any][]string{
				(*oldmodel.MediaAttachment)(nil): {"Cached", "Processing"},
				(*oldmodel.Emoji)(nil):           {"Cached"},
			} {
				for _, field := range fields {
					if err := dropColumn(ctx, tx, model, field); err != nil {
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
