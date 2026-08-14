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

package main

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"codeberg.org/gruf/go-byteutil"
	"codeberg.org/gruf/go-kv/v2"
	"codeberg.org/gruf/go-kv/v2/format"
	"github.com/ncruces/go-sqlite3"
	"github.com/openengineer/go-repl"
)

func main() {
	fmt.Print("please note this is a utility for gotosocial developers\n" +
		"it provides no warranty or data safety guarantees\n" +
		"ensure you have backups!!1!\n\n")

	// Open connection to database at args.
	conn, err := sqlite3.Open(os.Args[1])
	if err != nil {
		panic(err)
	}

	// Close on exit.
	defer conn.Close()

	// Prepare new SQLite REPL.
	sqliteREPL := new(sqliteREPL)
	sqliteREPL.db = conn

	// Pass as handler to REPL lib.
	repl := repl.NewRepl(sqliteREPL)
	sqliteREPL.rpl = repl

	// Start the main REPL library loop.
	if err := repl.Loop(); err != nil {
		panic(err)
	}
}

var args = func() format.Args {
	args := format.DefaultArgs()
	args.SetLogfmt()
	return args
}()

type sqliteREPL struct {
	db  *sqlite3.Conn
	rpl *repl.Repl
	dst []any
	buf byteutil.Buffer
}

func (r *sqliteREPL) Prompt() string {
	return "> "
}

func (r *sqliteREPL) Eval(buffer string) string {
	var lines []string

	switch buffer { //nolint
	case ".q":
		// Finish up with
		// database conn.
		_ = r.db.Close()

		// Exit REPL,
		// calls os.Exit().
		r.rpl.Quit()
	}

	// Prepare statement provided in buffer.
	stmt, tail, err := r.db.Prepare(buffer)
	defer stmt.Close() //nolint

	if err != nil {
		return fmt.Sprintf("sql error: %v", err)
	} else if tail != "" {
		return "sql error: multi queries per line not supported"
	}

	// Get no. returned columns.
	cols := stmt.ColumnCount()

	// Get names for these columns.
	names := make([]string, cols)
	for i := range names {
		names[i] = stmt.ColumnName(i)
	}

	// Prepare destination slice.
	clear(r.dst[0:cap(r.dst)])
	r.dst = slices.Grow(r.dst[:0], cols)
	r.dst = r.dst[0:cols]

	if len(names) != len(r.dst) {
		panic("bound check elimination")
	}

	// Start scanning.
	for stmt.Step() {

		// Reset dest on each iter.
		clear(r.dst[0:len(r.dst)])

		// Scan columns into our slice.
		err := stmt.ColumnsRaw(r.dst...)
		if err != nil {
			return fmt.Sprintf("scan error: %v", err)
		}

		// Format column results
		// with names into buffer.
		for i := range cols {
			kv.AppendQuoteString(&r.buf, names[i])
			r.buf.B = append(r.buf.B, '=')
			r.buf.B = format.Global.Append(r.buf.B, r.dst[i], args)
			r.buf.B = append(r.buf.B, ',', ' ')
		}

		// Drop last comma
		// space separator.
		r.buf.Truncate(2)

		// Append buffer to raw line output.
		lines = append(lines, string(r.buf.B))
		r.buf.Reset()
	}

	// Return raw lines output.
	return strings.Join(lines, "\n")
}

func (r *sqliteREPL) Tab(buffer string) string {
	return ""
}
