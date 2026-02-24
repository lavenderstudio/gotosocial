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

	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type directoryDB struct {
	db    *bun.DB
	state *state.State
}

func (d *directoryDB) GetDirectoryPage(
	ctx context.Context,
	page *paging.Page,
	offset int,
	orderBy gtsmodel.DirectoryOrderBy,
) ([]*gtsmodel.Account, error) {
	// Ensure nobody's doing anything silly.
	if orderBy == gtsmodel.DirectoryOrderByUnknown {
		panic("invalid orderBy value")
	}

	var (
		// Make educated guess for slice size
		limit      = page.GetLimit()
		accountIDs = make([]string, 0, limit)
	)

	q := d.db.NewSelect().
		TableExpr("? AS ?", bun.Ident("accounts"), bun.Ident("account")).
		// Join on account status
		// to check last_status_at.
		Join(
			"JOIN ? ON ? = ?",
			bun.Ident("account_stats"),
			bun.Ident("account_stats.account_id"), bun.Ident("account.id"),
		).
		// Select only IDs from table.
		Column("account.id").
		// Select only local accounts.
		//
		// This is first so it can use the partial
		// query and filter out a shitload of remote
		// accounts we're not interested in.
		Where("? IS NULL", bun.Ident("account.domain")).
		// Select only discoverable accounts.
		Where("? = ?", bun.Ident("account.discoverable"), true).
		// Don't select suspended accounts.
		Where("? IS NULL", bun.Ident("account.suspended_at")).
		// Select only accounts that have posted at least once.
		Where("? IS NOT NULL", bun.Ident("account_stats.last_status_at"))

	// Add paging params
	// to the query.
	var err error
	q, err = d.pageDirectoryQuery(
		ctx,
		q,
		page,
		offset,
		orderBy,
	)
	if err != nil {
		return nil, err
	}

	if limit > 0 {
		// Limit amount of
		// accounts returned.
		q = q.Limit(limit)
	}

	if err := q.Scan(ctx, &accountIDs); err != nil {
		return nil, err
	}

	return d.state.DB.GetAccountsByIDs(ctx, accountIDs)
}

func (d *directoryDB) pageDirectoryQuery(
	ctx context.Context,
	q *bun.SelectQuery,
	page *paging.Page,
	offset int,
	orderBy gtsmodel.DirectoryOrderBy,
) (*bun.SelectQuery, error) {
	var (
		// Get paging params.
		minID = page.GetMin()
		maxID = page.GetMax()
		order = page.GetOrder()

		// We know orderBy is either Active or New, not Unknown.
		orderByActive = orderBy == gtsmodel.DirectoryOrderByActive
	)

	if offset != 0 {
		// If offset is set, just use
		// this for paging, ignoring
		// maxID, minID, etc.
		q = q.Offset(offset)
	} else {
		// Page using max + min ID.
		if maxID != "" {
			// Get account for max ID.
			maxIDAcct, err := d.state.DB.GetAccountByID(
				gtscontext.SetBarebones(ctx),
				maxID,
			)
			if err != nil {
				err := gtserror.NewfAt(3, "db error getting maxID account %s: %w", maxID, err)
				return nil, err
			}

			if orderByActive {
				// Get account stats and use last_status_at time.
				if err := d.state.DB.PopulateAccountStats(ctx, maxIDAcct); err != nil {
					err := gtserror.NewfAt(3, "db error getting stats for maxID account %s: %w", maxID, err)
					return nil, err
				}
				q = q.Where("? < ?", bun.Ident("account_stats.last_status_at"), maxIDAcct.Stats.LastStatusAt)
			} else {
				// Use account creation time.
				q = q.Where("? < ?", bun.Ident("account.created_at"), maxIDAcct.CreatedAt)
			}
		}

		if minID != "" {
			// Get account for min ID.
			minIDAcct, err := d.state.DB.GetAccountByID(
				gtscontext.SetBarebones(ctx),
				minID,
			)
			if err != nil {
				err := gtserror.NewfAt(3, "db error getting minID account %s: %w", minID, err)
				return nil, err
			}

			if orderByActive {
				// Get account stats and use last_status_at time.
				if err := d.state.DB.PopulateAccountStats(ctx, minIDAcct); err != nil {
					err := gtserror.NewfAt(3, "db error getting stats for minID account %s: %w", minID, err)
					return nil, err
				}
				q = q.Where("? > ?", bun.Ident("account_stats.last_status_at"), minIDAcct.Stats.LastStatusAt)
			} else {
				// Use account creation time.
				q = q.Where("? > ?", bun.Ident("account.created_at"), minIDAcct.CreatedAt)
			}
		}
	}

	if orderByActive {
		// Order by latest status.
		if order == paging.OrderAscending {
			q = q.OrderExpr("? ASC", bun.Ident("account_stats.last_status_at"))
		} else {
			q = q.OrderExpr("? DESC", bun.Ident("account_stats.last_status_at"))
		}
	} else {
		// Order by creation date.
		if order == paging.OrderAscending {
			q = q.OrderExpr("? ASC", bun.Ident("account.created_at"))
		} else {
			q = q.OrderExpr("? DESC", bun.Ident("account.created_at"))
		}
	}

	return q, nil
}
