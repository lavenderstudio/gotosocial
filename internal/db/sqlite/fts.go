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

package sqlite

import (
	"context"
	"slices"
	"strings"

	"code.superseriousbusiness.org/gopkg/log"
	"code.superseriousbusiness.org/gopkg/xslices"
	dbpkg "code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

// CreateFts5Tables will create and populate new
// FTS5 virtual tables according to var dbpkg.AllFts().
func CreateFts5Tables(ctx context.Context, db *bun.DB) error {
	if db.Dialect().Name() != dialect.SQLite {
		return gtserror.New("incorrect database type")
	}
	for _, table := range dbpkg.AllFts() {
		if err := createFts5VirtualTable(ctx, db,
			table.Name,
			table.Tokenizer,
			table.Table,
			table.Columns...,
		); err != nil {
			return err
		}
	}
	return nil
}

// DropFts5Tables will drop any existing
// FTS5 virtual tables according to dbpkg.AllFts().
func DropFts5Tables(ctx context.Context, db *bun.DB) error {
	if db.Dialect().Name() != dialect.SQLite {
		return gtserror.New("incorrect database type")
	}
	for _, table := range dbpkg.AllFts() {
		if _, err := db.NewRaw("DROP TABLE IF EXISTS ?",
			table.Name).Exec(ctx); err != nil {
			return gtserror.Newf("error dropping table %s: %w", table.Name, err)
		}
		for _, trigger := range []string{
			"INSERT",
			"UPDATE",
			"DELETE",
		} {
			triggerName := table.Name + "_after_"
			triggerName += bun.Ident(strings.ToLower(trigger))
			if _, err := db.NewRaw("DROP TRIGGER IF EXISTS ?",
				triggerName).Exec(ctx); err != nil {
				return gtserror.Newf("error dropping trigger %s: %w", triggerName, err)
			}
		}
	}
	return nil
}

// RebuildFts5Tables will rebuild existing FTS5 virtual tables according to dbpkg.FtsTables.
func RebuildFts5Tables(ctx context.Context, db *bun.DB) error {
	if db.Dialect().Name() != dialect.SQLite {
		return gtserror.New("incorrect database type")
	}
	for _, table := range dbpkg.AllFts() {
		_, err := db.NewRaw("INSERT INTO ? (?) VALUES ('rebuild')",
			table.Name, table.Name).Exec(ctx)
		if err != nil {
			return gtserror.Newf("error rebuilding fts5 table %s: %w", table.Name, err)
		}
	}
	return nil
}

// createFts5VirtualTable wraps the creation of an FTS5 virtual
// table with given details, including any necessary triggers on
// the external content table, and inserting any initial data.
func createFts5VirtualTable(
	ctx context.Context,
	db *bun.DB,

	// name of the
	// FTS5 table.
	name bun.Ident,

	// tokenizer type
	// to use for table.
	tokenizer int,

	// external content table
	// that FTS5 table will be
	// indexing columns from.
	extTable bun.Ident,

	// external table columns
	// that will be tokenized
	// and stored in this table.
	extCols ...bun.Ident,
) error {
	if len(extCols) == 0 {
		panic("no columns given")
	}

	// define a shorthand for
	// sqlite's default rowid column.
	const rowid = bun.Ident("rowid")

	log.Infof(ctx, "creating fts5 table: %s", name)

	// Get column bun placeholder arguments.
	colArgs := dbpkg.Placeholders(len(extCols))

	// Start preparing FTS5 virtual table query.
	stmt := "CREATE VIRTUAL TABLE ? USING fts5("
	stmt += colArgs

	// Finish statement declaring external content table,
	// and tokenizer implementation in use for driver.
	stmt += ", content=?, tokenize=?);"

	// Prepare args w/ name, cols, and external content table.
	args := append([]any{name}, xslices.ToAny(extCols)...)
	args = append(args, extTable)

	// Add tokenizer arg.
	switch tokenizer {
	case dbpkg.TokenizerTypeWord:
		args = append(args, WordTokenizer)
	case dbpkg.TokenizerTypeTrigram:
		args = append(args, "trigram")
	}

	// Execute the prepared statement with args.
	_, err := db.NewRaw(stmt, args...).Exec(ctx)
	if err != nil {
		return gtserror.Newf("error creating table %s: %w", name, err)
	}

	// Create a set of bun arguments of FTS indexed
	// column names, starting with internal `rowid`,
	// all prefixed with "new." for INSERT triggers.
	newColArgs := []any{bun.Ident("new.rowid")}
	for _, col := range extCols {
		newColArgs = append(newColArgs, "new."+col)
	}

	// Create a set of bun arguments of FTS indexed
	// column names, starting with internal `rowid`,
	// all prefixed with "old." for DELETE triggers.
	oldColArgs := []any{bun.Ident("old.rowid")}
	for _, col := range extCols {
		oldColArgs = append(oldColArgs, "old."+col)
	}

	// Get external columns as 'any' slice,
	// used in later query argument building.
	anyExtCols := xslices.ToAny(extCols)

	// Create series of triggers to cover
	// INSERT, UPDATE and DELETE events to
	// the external content table, to ensure
	// that the new FTS5 table stays in sync.
	for _, trigger := range []struct {

		// after is the SQL method
		// which activates trigger.
		after string

		// do is the list of statements
		// which are performed on trigger.
		do []dbpkg.BunExpr
	}{
		{
			// CREATE TRIGGER table_fts_after_insert AFTER INSERT ON table
			//     WHEN ...
			// BEGIN
			//     INSERT INTO table_fts (rowid, col1, col2, col3) VALUES (new.rowid, new.col1, new.col2, new.col3);
			// END;
			after: "INSERT",
			do: []dbpkg.BunExpr{
				{
					Fmt: "INSERT INTO ? (?, " + colArgs + ") VALUES (?, " + colArgs + ")",
					Arg: slices.Concat([]any{name, rowid}, anyExtCols, newColArgs),
				},
			},
		},

		{
			// CREATE TRIGGER table_fts_after_update AFTER UPDATE ON table BEGIN
			//     INSERT INTO table_fts (table_fts, rowid, col1, col2, col3) VALUES ('delete', old.rowid, old.col1, old.col2, old.col3);
			//     INSERT INTO table_fts (rowid, col1, col2, col3) VALUES (new.rowid, new.col1, new.col2, new.col3);
			// END;
			after: "UPDATE",
			do: []dbpkg.BunExpr{
				{
					Fmt: "INSERT INTO ? (?, ?, " + colArgs + ") VALUES (?, ?, " + colArgs + ")",
					Arg: slices.Concat([]any{name, name, rowid}, anyExtCols, []any{"delete"}, oldColArgs),
				},
				{
					Fmt: "INSERT INTO ? (?, " + colArgs + ") VALUES (?, " + colArgs + ")",
					Arg: slices.Concat([]any{name, rowid}, anyExtCols, newColArgs),
				},
			},
		},

		{
			// CREATE TRIGGER table_fts_after_delete AFTER DELETE ON table BEGIN
			//     INSERT INTO table_fts (table_fts, rowid, col1, col2, col3) VALUES ('delete', old.rowid, old.col1, old.col2, old.col3);
			// END;
			after: "DELETE",
			do: []dbpkg.BunExpr{
				{
					Fmt: "INSERT INTO ? (?, ?, " + colArgs + ") VALUES (?, ?, " + colArgs + ")",
					Arg: slices.Concat([]any{name, name, rowid}, anyExtCols, []any{"delete"}, oldColArgs),
				},
			},
		},
	} {

		// Generate trigger name based
		// on fts table name and action.
		triggerName := name + "_after_"
		triggerName += bun.Ident(strings.ToLower(trigger.after))

		// Start preparing trigger creation SQL statement and args slice.
		stmt := "CREATE TRIGGER ? AFTER " + trigger.after + " ON ? BEGIN "
		args := []any{triggerName, extTable}

		// For each of the trigger actions
		// append to statement and args.
		for _, do := range trigger.do {
			args = append(args, do.Arg...)
			stmt += do.Fmt + "; "
		}

		// Finish stmt.
		stmt += "END;"

		log.Infof(ctx, "creating trigger: %s", triggerName)

		// Execute the prepared statement with args.
		_, err := db.NewRaw(stmt, args...).Exec(ctx)
		if err != nil {
			return gtserror.Newf("error creating trigger %s: %w", triggerName, err)
		}
	}

	log.Infof(ctx, "populating fts5 table: %s", name)

	// As the FTS5 table is based on external content, populate it with a rebuild.
	_, err = db.NewRaw("INSERT INTO ? (?) VALUES ('rebuild')", name, name).Exec(ctx)
	if err != nil {
		return gtserror.Newf("error populating table %s: %w", name, err)
	}

	return nil
}
