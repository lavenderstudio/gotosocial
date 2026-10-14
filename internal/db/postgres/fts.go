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

package postgres

import (
	"context"

	"code.superseriousbusiness.org/gopkg/log"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"

	dbpkg "code.superseriousbusiness.org/gotosocial/internal/db"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

// CreateFtsIndices creates new FTS indices according to dbpkg.AllFts().
// NOTE: this expects the pg_trgm extension has been enabled by its migration.
func CreateFtsIndices(ctx context.Context, db *bun.DB) error {
	if db.Dialect().Name() != dialect.PG {
		return gtserror.New("incorrect database type")
	}
	for _, table := range dbpkg.AllFts() {
		if err := createFtsIndices(ctx, db,
			table.Tokenizer,
			table.Table,
			table.Columns,
		); err != nil {
			return err
		}
	}
	return nil
}

// DropFtsIndices drops any existing FTS indices according to dbpkg.AllFts().
func DropFtsIndices(ctx context.Context, db *bun.DB) error {
	if db.Dialect().Name() != dialect.PG {
		return gtserror.New("incorrect database type")
	}
	for _, table := range dbpkg.AllFts() {
		for _, col := range table.Columns {
			name := ftsIndexName(table.Name, col)
			if _, err := db.NewRaw("DROP INDEX IF EXISTS ?", name).Exec(ctx); err != nil {
				return gtserror.Newf("error dropping index %s: %w", name, err)
			}
		}
	}
	return nil
}

// createFtsIndices creates a new postgres
// FTS GIN per-column given, on the given table,
// with appropriate tokenizing function.
func createFtsIndices(
	ctx context.Context,
	db *bun.DB,
	tokenizer dbpkg.TokenizerType,
	table bun.Ident,
	cols []bun.Ident,
) error {
	if len(cols) == 0 {
		panic("no columns given")
	}

	for _, col := range cols {
		// Generate index name for column.
		name := ftsIndexName(table, col)
		log.Infof(ctx, "creating fts index: %s", name)

		switch tokenizer {
		case dbpkg.TokenizerTypeWord:

			// Create a GIN index on this table using case insensitive simple dictionary tokenizing.
			_, err := db.NewRaw("CREATE INDEX ? ON ? USING GIN (to_tsvector('simple', lower(?)))",
				name, table, col).Exec(ctx)
			if err != nil {
				return gtserror.Newf("error creating index %s: %w", name, err)
			}

		case dbpkg.TokenizerTypeTrigram:

			// Create a GIN index on this table using case insensitive trigrams tokenizing.
			_, err := db.NewRaw("CREATE INDEX ? ON ? USING GIN (lower(?) gin_trgm_ops)",
				name, table, col).Exec(ctx)
			if err != nil {
				return gtserror.Newf("error creating index %s: %w", name, err)
			}
		}
	}

	return nil
}

// ftsIndexName generates FTS index name to use for column on table.
func ftsIndexName(table bun.Ident, col bun.Ident) bun.Ident {
	return table + "_" + col + "_fts_idx"
}
