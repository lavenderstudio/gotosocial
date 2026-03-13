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

package bundb

import (
	"context"
	"database/sql"
	"slices"
	"strings"

	"code.superseriousbusiness.org/gopkg/log"
	"code.superseriousbusiness.org/gotosocial/internal/cache"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// likeEscaper is a thread-safe string replacer which escapes
// common SQLite + Postgres `LIKE` wildcard chars using the
// escape character `\`. Initialized as a var in this package
// so it can be reused.
var likeEscaper = strings.NewReplacer(
	`\`, `\\`, // Escape char.
	`%`, `\%`, // Zero or more char.
	`_`, `\_`, // Exactly one char.
)

// likeOperator returns an appropriate LIKE or
// ILIKE operator for the given query's dialect.
func likeOperator(query *bun.SelectQuery) string {
	const (
		like  = "LIKE"
		ilike = "ILIKE"
	)

	d := query.Dialect().Name()
	if d == dialect.SQLite {
		return like
	} else if d == dialect.PG {
		return ilike
	}

	log.Panicf(nil, "db conn %s was neither pg nor sqlite", d)
	return ""
}

// bunArrayType wraps the given type in a pgdialect.Array
// if needed, which postgres wants for serializing arrays.
func bunArrayType(db bun.IDB, arr any) any {
	switch db.Dialect().Name() {
	case dialect.SQLite:
		return arr // return as-is
	case dialect.PG:
		return pgdialect.Array(arr)
	default:
		panic("unreachable")
	}
}

// whereLike appends a WHERE clause to the
// given SelectQuery, which searches for
// matches of `search` in the given subQuery
// using LIKE (SQLite) or ILIKE (Postgres).
func whereLike(
	query *bun.SelectQuery,
	subject interface{},
	search string,
) *bun.SelectQuery {
	// Escape existing wildcard + escape
	// chars in the search query string.
	search = likeEscaper.Replace(search)

	// Add our own wildcards back in; search
	// zero or more chars around the query.
	search = `%` + search + `%`

	// Get appropriate operator.
	like := likeOperator(query)

	// Append resulting WHERE
	// clause to the main query.
	return query.Where(
		"(?) ? ? ESCAPE ?",
		subject, bun.Safe(like), search, `\`,
	)
}

// whereStartsLike is like whereLike,
// but only searches for strings that
// START WITH `search`.
func whereStartsLike(
	query *bun.SelectQuery,
	subject interface{},
	search string,
) *bun.SelectQuery {
	// Escape existing wildcard + escape
	// chars in the search query string.
	search = likeEscaper.Replace(search)

	// Add our own wildcards back in; search
	// zero or more chars after the query.
	search += `%`

	// Get appropriate operator.
	like := likeOperator(query)

	// Append resulting WHERE
	// clause to the main query.
	return query.Where(
		"(?) ? ? ESCAPE ?",
		subject, bun.Safe(like), search, `\`,
	)
}

// exists checks the results of a SelectQuery for the existence of the data in question, masking ErrNoEntries errors.
func exists(ctx context.Context, query *bun.SelectQuery) (bool, error) {
	exists, err := query.Exists(ctx)
	switch err {
	case nil:
		return exists, nil
	case sql.ErrNoRows:
		return false, nil
	default:
		return false, err
	}
}

// notExists checks the results of a SelectQuery for the non-existence of the data in question, masking ErrNoEntries errors.
func notExists(ctx context.Context, query *bun.SelectQuery) (bool, error) {
	exists, err := exists(ctx, query)
	return !exists, err
}

// loadPagedIDs loads a page of IDs from given SliceCache by `key`, resorting to `loadDESC` if required. Uses `page` to sort + page resulting IDs.
// NOTE: IDs returned from `cache` / `loadDESC` MUST be in descending order, otherwise paging will not work correctly / return things out of order.
func loadPagedIDs(cache *cache.SliceCache[string], key string, page *paging.Page, loadDESC func() ([]string, error)) ([]string, error) {

	// Check cache for IDs, else load.
	ids, err := cache.Load(key, loadDESC)
	if err != nil {
		return nil, err
	}

	// Our cached / selected IDs are ALWAYS
	// fetched from `loadDESC` in descending
	// order. Depending on the paging requested
	// this may be an unexpected order.
	if page.GetOrder().Ascending() {
		slices.Reverse(ids)
	}

	// Page the resulting IDs.
	ids = page.Page(ids)

	return ids, nil
}

// incrementAccountStats will increment the given column in the `account_stats` table matching `account_id`.
func incrementAccountStats(ctx context.Context, tx bun.Tx, col bun.Ident, accountID string) error {
	if _, err := tx.NewUpdate().
		Model((*gtsmodel.AccountStats)(nil)).
		Where("? = ?", bun.Ident("account_id"), accountID).
		Set("? = (? + 1)", bun.Ident(col), bun.Ident(col)).
		Exec(ctx); err != nil {
		return gtserror.Newf("error updating %s: %w", col, err)
	}
	return nil
}

// decrementAccountStats will decrement the given column in the `account_stats`
// table matching `account_id`, not allowing the value to decrement below 0.
func decrementAccountStats(ctx context.Context, tx bun.Tx, col bun.Ident, accountID string) error {
	q := tx.NewUpdate().
		Model((*gtsmodel.AccountStats)(nil)).
		Where("? = ?", bun.Ident("account_id"), accountID)

	// Set either to zero or to [col]-1 using
	// MAX (sqlite) or GREATEST (postgres) funcs.
	switch tx.Dialect().Name() {
	case dialect.SQLite:
		// https://sqlite.org/lang_corefunc.html#max_scalar
		q = q.Set("? = MAX(0, ? - 1)", bun.Ident(col), bun.Ident(col))
	case dialect.PG:
		// https://www.postgresql.org/docs/current/functions-conditional.html#FUNCTIONS-GREATEST-LEAST
		q = q.Set("? = GREATEST(0, ? - 1)", bun.Ident(col), bun.Ident(col))
	default:
		panic("unreachable")
	}

	if _, err := q.Exec(ctx); err != nil {
		return gtserror.Newf("error updating %s: %w", col, err)
	}

	return nil
}

// updateWhere parses []db.Where and adds it to the given update query.
func updateWhere(q *bun.UpdateQuery, where []db.Where) {
	for _, w := range where {
		query, args := parseWhere(w)
		q = q.Where(query, args...)
	}
}

// selectWhere parses []db.Where and adds it to the given select query.
func selectWhere(q *bun.SelectQuery, where []db.Where) {
	for _, w := range where {
		query, args := parseWhere(w)
		q = q.Where(query, args...)
	}
}

// deleteWhere parses []db.Where and adds it to the given where query.
func deleteWhere(q *bun.DeleteQuery, where []db.Where) {
	for _, w := range where {
		query, args := parseWhere(w)
		q = q.Where(query, args...)
	}
}

// parseWhere looks through the options on a single db.Where entry, and
// returns the appropriate query string and arguments.
func parseWhere(w db.Where) (query string, args []interface{}) {
	if w.Not {
		if w.Value == nil {
			query = "? IS NOT NULL"
			args = []interface{}{bun.Ident(w.Key)}
			return
		}

		query = "? != ?"
		args = []interface{}{bun.Ident(w.Key), w.Value}
		return
	}

	if w.Value == nil {
		query = "? IS NULL"
		args = []interface{}{bun.Ident(w.Key)}
		return
	}

	query = "? = ?"
	args = []interface{}{bun.Ident(w.Key), w.Value}
	return
}

// accountIDValue is a convenience struct for using
// CTE's to provide accountIDs to select statuses of.
type accountIDValue struct {
	AccountID string `bun:"type:CHAR(26)"`
}
