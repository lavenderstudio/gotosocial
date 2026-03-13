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

package db

import (
	"database/sql/driver"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// BunExpr encompasses the arguments
// that get passed to a bun._Expr() type
// function, also usefully to Where()!
type BunExpr struct {
	Fmt string
	Arg []any
}

// ToNamedValues converts older driver.Value types to driver.NamedValue types.
func ToNamedValues(args []driver.Value) []driver.NamedValue {
	if args == nil {
		return nil
	}
	args2 := make([]driver.NamedValue, len(args))
	for i := range args {
		args2[i] = driver.NamedValue{
			Ordinal: i + 1,
			Value:   args[i],
		}
	}
	return args2
}

// BitNotSet returns a set of arguments for a bun
// expression indicating whether a bit flag IS set.
func BitIsSet[Type ~int16](col string, value Type) (
	string, // query string
	bun.Ident, // column ident
	Type, // bit flag
) {
	return "? & ? != 0", bun.Ident(col), value
}

// BitNotSet returns a set of arguments for a bun
// expression indicating whether a bit flag IS NOT set.
func BitNotSet[Type ~int16](col string, value Type) (
	string, // query string
	bun.Ident, // column ident
	Type, // bit flag
) {
	return "? & ? = 0", bun.Ident(col), value
}

// BitSetExpr ...
func BitSetExpr[Type ~int16](col string, value Type) BunExpr {
	return BunExpr{"? & ? != 0", []any{bun.Ident(col), value}}
}

// BitNotSetExpr ...
func BitNotSetExpr[Type ~int16](col string, value Type) BunExpr {
	return BunExpr{"? & ? = 0", []any{bun.Ident(col), value}}
}

// Idents is syntactic sugar for
// converting a variable slice of
// string arguments to bun.Ident()
// types for use as bun []any args.
//
// This allocates an extra slice of []any
// type so do not use in performance critical
// pathways, ideally only in migrations!
func Idents(s ...string) []any {
	a := make([]any, len(s))
	for i, s := range s {
		a[i] = bun.Ident(s)
	}
	return a
}

// ArrayType wraps the given type in a pgdialect.Array
// if needed, which postgres wants for serializing arrays.
func ArrayType(db bun.IDB, arr any) any {
	switch db.Dialect().Name() {
	case dialect.SQLite:
		return arr // return as-is
	case dialect.PG:
		return pgdialect.Array(arr)
	default:
		panic("unreachable")
	}
}

// WhereArrayIsNullOrEmpty returns a BunExpr checking whether value contained in
// 'col' is NULL or is an empty JSON array, depending on current database type.
func WhereArrayIsNullOrEmpty(db bun.IDB, col string) (string, bun.Ident, bun.Ident) {
	var query string
	switch db.Dialect().Name() {
	case dialect.SQLite:
		query = "(? IS NULL) OR (json_array_length(?) = 0)"
	case dialect.PG:
		query = "(? IS NULL) OR (CARDINALITY(?) = 0)"
	default:
		panic("unreachable")
	}
	return query, bun.Ident(col), bun.Ident(col)
}

// WhereArrayIsNullOrEmpty returns a BunExpr checking whether value contained in
// 'col' is NULL or is an empty JSON array, depending on current database type.
func WhereArrayIsNullOrEmptyExpr(db bun.IDB, col string) (expr BunExpr) {
	switch db.Dialect().Name() {
	case dialect.SQLite:
		expr.Fmt = "(? IS NULL) OR (json_array_length(?) = 0)"
	case dialect.PG:
		expr.Fmt = "(? IS NULL) OR (CARDINALITY(?) = 0)"
	default:
		panic("unreachable")
	}
	expr.Arg = []any{bun.Ident(col), bun.Ident(col)}
	return
}
