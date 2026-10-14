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

package search

import (
	"context"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

// dbSearchForAccounts wraps the database SearchForAccounts()
// function to support passing our own accountQuery type.
func (p *Processor) dbSearchForAccounts(
	ctx context.Context,
	requester *gtsmodel.Account,
	query *searchQuery,
	page *paging.Page,
) (
	[]*gtsmodel.Account,
	error,
) {
	switch query.DomainQuery {
	case config.GetHost(),
		config.GetAccountDomain():
		// Normalize domain according
		// to database column values.
		query.DomainQuery = ""
	}

	// Update paging with before / after times if set.
	setBeforeAfterParams(page, query.Before, query.After)

	// Set requesting account's ID.
	query.RequesterID = requester.ID

	// Perform the database search query using params.
	accounts, err := p.state.DB.SearchForAccounts(ctx,
		db.AccountSearchParams{
			CommonSearchParams:    query.CommonSearchParams,
			UsernameOrDomainQuery: query.Remain,
			UsernameQuery:         query.UsernameQuery,
			DomainQuery:           query.DomainQuery,
			NoteQuery:             query.NoteQuery,
		},
		page,
	)
	if err != nil {
		return nil, gtserror.Newf("db error searching for accounts: %w", err)
	}

	return accounts, nil
}

// dbSearchForStatuses wraps the database SearchForStatuses()
// function to support passing our own statusQuery type.
func (p *Processor) dbSearchForStatuses(
	ctx context.Context,
	requester *gtsmodel.Account,
	query *searchQuery,
	page *paging.Page,
) (
	[]*gtsmodel.Status,
	error,
) {
	// Update paging with before / after times if set.
	setBeforeAfterParams(page, query.Before, query.After)

	// Set requesting account's ID.
	query.RequesterID = requester.ID

	// Perform the database search query using params.
	statuses, err := p.state.DB.SearchForStatuses(ctx,
		db.StatusSearchParams{
			CommonSearchParams:    query.CommonSearchParams,
			ContentOrWarningQuery: query.Remain,
			ContentWarningQuery:   query.ContentWarningQuery,
			ContentQuery:          query.ContentQuery,
			AuthorID:              query.AuthorID,
			Reply:                 query.Reply,
			Sensitive:             query.Sensitive,
			Media:                 query.Media,
			Poll:                  query.Poll,
		},
		page,
	)
	if err != nil {
		return nil, gtserror.Newf("db error searching for statuses: %w", err)
	}

	return statuses, nil
}

// dbSearchForTags wraps the
// database SearchForTags() function.
func (p *Processor) dbSearchForTags(
	ctx context.Context,
	query string,
	page *paging.Page,
) (
	[]*gtsmodel.Tag,
	error,
) {
	// Perform the database tag search query.
	tags, err := p.state.DB.SearchForTags(ctx,
		query,
		page,
	)
	if err != nil {
		return nil, gtserror.Newf("db error searching for tags: %w", err)
	}

	return tags, nil
}

// setBeforeAfterParams updates given page minimum and maximum ID values according
// to whether before|after times provide a narrower scope than correctly set ULIDs.
func setBeforeAfterParams(page *paging.Page, before, after time.Time) {

	if !before.IsZero() {
		// If a before time is provided, apply this
		// limit via ID paging, but only if minimum
		// ID doesn't provide a more narrow scope.
		beforeID := id.NewULIDFromTime(before)
		if beforeID > page.Min.Value {
			page.Min.Value = beforeID
		}
	}

	if !after.IsZero() {
		// If an after time is provided, apply this
		// limit via ID paging, but only if maximum
		// ID doesn't provide a more narrow scope.
		afterID := id.NewULIDFromTime(after)
		if afterID < page.Max.Value {
			page.Max.Value = afterID
		}
	}
}
