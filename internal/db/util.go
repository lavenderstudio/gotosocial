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
	"context"
	"database/sql"
	"database/sql/driver"

	"code.superseriousbusiness.org/gopkg/buffers"
	"code.superseriousbusiness.org/gopkg/xslices"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/schema"
)

// BunExpr encompasses the arguments
// that get passed to a bun._Expr() type
// function, also usefully to Where()!
type BunExpr struct {
	Fmt string
	Arg []any
}

// BunQueryable defines a bun type that
// permits starting a new database query.
type BunQueryable interface {
	Dialect() schema.Dialect
	NewValues(model any) *bun.ValuesQuery
	NewSelect() *bun.SelectQuery
	NewInsert() *bun.InsertQuery
	NewUpdate() *bun.UpdateQuery
	NewDelete() *bun.DeleteQuery
	NewRaw(query string, args ...any) *bun.RawQuery
}

// BunQueryBuilder defines a bun query builder type.
type BunQueryBuilder[QueryType any] interface {
	BunQueryable
	Table(tables ...string) QueryType
	TableExpr(query string, args ...any) QueryType
	Column(columns ...string) QueryType
	ColumnExpr(query string, args ...any) QueryType
	Where(query string, args ...any) QueryType
	Limit(n int) QueryType
	Order(orders ...string) QueryType
	OrderExpr(query string, args ...any) QueryType
	Scan(ctx context.Context, args ...any) error
}

// Placeholders returns bun argument
// placeholders up to a maximum of 40.
//
// any more than this will panic but
// god i cannot think of any queries
// we write by hand that get this large.
func Placeholders(n int) string {
	const placeholders = "?, ?, ?, ?, ?, ?, ?, ?, ?, ?, " +
		"?, ?, ?, ?, ?, ?, ?, ?, ?, ?, " +
		"?, ?, ?, ?, ?, ?, ?, ?, ?, ?, " +
		"?, ?, ?, ?, ?, ?, ?, ?, ?, ?, "
	return placeholders[:(3*n)-2]
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

// Where is syntactic sugar for returning a
// BunExpr{} with format string and arguments.
func Where(fmt string, arg ...any) BunExpr {
	return BunExpr{Fmt: fmt, Arg: arg}
}

// WhereGroup returns a BunExpr{} WHERE clause that is
// a combination of the given BunExpr{} joined by 'sep'.
func WhereGroup(sep string, wheres ...BunExpr) (expr BunExpr) {
	if len(wheres) == 0 {
		panic("no wheres")
	}

	// Pre-allocate a slice of arguments of length of all combined arguments.
	expr.Arg = make([]any, 0, xslices.Count(wheres, func(where BunExpr) int {
		return len(where.Arg)
	}))

	// Acquire a byte buffer.
	buf := buffers.Get1k()
	buf.B = buf.B[:0]

	// Append each where expression
	// to buffer, separated by 'sep'.
	for _, where := range wheres {
		buf.B = append(buf.B, '(')
		buf.B = append(buf.B, where.Fmt...)
		buf.B = append(buf.B, ')')
		buf.B = append(buf.B, sep...)
		expr.Arg = append(expr.Arg, where.Arg...)
	}

	// Truncate trailing separator,
	// obtain buffered string copy,
	// return buffer to pool.
	buf.Truncate(len(sep))
	expr.Fmt = string(buf.B)
	buffers.Put1k(buf)

	return
}

// WhereGroupQuery calls WhereGroup() and appends
// the resulting BunExpr{} to given query builder.
func WhereGroupQuery[

	// this is some very confusing generic parameter-
	// -isation here, but you can check the type info
	// shown by the linter for uses of this function
	// and you'll see it collapses without interfaces
	//
	// e.g.
	// WhereGroup(*bun.SelectQuery, ...) *bun.SelectQuery
	QueryType any,
	Builder BunQueryBuilder[QueryType],
](
	q Builder,
	sep string,
	wheres ...BunExpr,
) QueryType {
	expr := WhereGroup(sep, wheres...)
	return q.Where(expr.Fmt, expr.Arg...)
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

// BitSetExpr returns the results of BitSet() as a BunExpr{}.
func BitSetExpr[Type ~int16](col string, value Type) BunExpr {
	return BunExpr{"? & ? != 0", []any{bun.Ident(col), value}}
}

// BitNotSetExpr returns the results of BitNotSet() as a BunExpr{}.
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

// GroupArrayExpr returns the appropriate expression for database
// type for grouping an array of column values into a single array.
func GroupArrayExpr(db bun.IDB) string {
	switch db.Dialect().Name() {
	case dialect.SQLite:
		return "json_group_array(?)"
	case dialect.PG:
		return "array_ag(?)"
	default:
		panic("unreachable")
	}
}

// GroupArray returns a grouped selection of column values as an array type parseable by bun.
func GroupArray(db bun.IDB, col string) (string, bun.Ident) {
	return GroupArrayExpr(db), bun.Ident(col)
}

// WhereArrayIsNullOrEmpty returns bun expr arguments checking whether value contained
// in 'col' is NULL or is an empty JSON array, depending on current database type.
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

// WhereNotArrayIsNullOrEmpty returns bun expr arguments checking whether value contained
// in 'col' is NOT (NULL or is an empty JSON array), depending on current database type.
func WhereNotArrayIsNullOrEmpty(db bun.IDB, col string) (string, bun.Ident, bun.Ident) {
	query, col1, col2 := WhereArrayIsNullOrEmpty(db, col)
	return "NOT (" + query + ")", col1, col2
}

// WhereArrayIsNullOrEmptyExpr returns a BunExpr checking whether value contained in
// 'col' is NULL or is an empty JSON array, depending on current database type.
func WhereArrayIsNullOrEmptyExpr(db bun.IDB, col string) (expr BunExpr) {
	query, col1, col2 := WhereArrayIsNullOrEmpty(db, col)
	expr.Fmt = query
	expr.Arg = []any{col1, col2}
	return
}

// WhereNotArrayIsNullOrEmptyExpr returns a BunExpr checking whether value contained in
// 'col' is NOT (NULL or is an empty JSON array), depending on current database type.
func WhereNotArrayIsNullOrEmptyExpr(db bun.IDB, col string) (expr BunExpr) {
	query, col1, col2 := WhereNotArrayIsNullOrEmpty(db, col)
	expr.Fmt = query
	expr.Arg = []any{col1, col2}
	return
}

// Scannable defines a type that
// provides separate SQLite and
// PostgreSQL scanning functions.
type Scannable interface {
	ScanRowPG(context.Context, *sql.Rows) error
	ScanRowSQLite(context.Context, *sql.Rows) error

	ScanRowsPG(context.Context, *sql.Rows) (int, error)
	ScanRowsSQLite(context.Context, *sql.Rows) (int, error)
}

// Scan calls Scan() by wrapping type T to use the appropriate scanning function for current db.
// This should only be required for types that bun otherwise gets confused about trying to scan.
func Scan[Q any, T Scannable](ctx context.Context, q BunQueryBuilder[Q], dst T) error {
	switch q.Dialect().Name() {
	case dialect.PG:
		return q.Scan(ctx, &asPGScanner[T]{dst})
	case dialect.SQLite:
		return q.Scan(ctx, &asSQLiteScanner[T]{dst})
	default:
		panic("unreachable")
	}
}

// asPGScanner is a type-alias for
// ScannableValue that calls ScanRow(s)?PG().
type asPGScanner[T Scannable] struct{ T T }

func (v *asPGScanner[T]) ScanRow(ctx context.Context, rows *sql.Rows) error {
	return v.T.ScanRowPG(ctx, rows)
}

func (v *asPGScanner[T]) ScanRows(ctx context.Context, rows *sql.Rows) (int, error) {
	return v.T.ScanRowsPG(ctx, rows)
}

func (v *asPGScanner[T]) Value() any {
	return v.T
}

// asSQLiteScanner is a type-alias for
// ScannableValue that calls ScanRow(s)?SQLite().
type asSQLiteScanner[T Scannable] struct{ T T }

func (v *asSQLiteScanner[T]) ScanRow(ctx context.Context, rows *sql.Rows) error {
	return v.T.ScanRowSQLite(ctx, rows)
}

func (v *asSQLiteScanner[T]) ScanRows(ctx context.Context, rows *sql.Rows) (int, error) {
	return v.T.ScanRowsSQLite(ctx, rows)
}

func (v *asSQLiteScanner[T]) Value() any {
	return v.T
}
