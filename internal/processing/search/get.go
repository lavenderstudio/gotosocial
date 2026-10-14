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
	"errors"
	"slices"
	"strings"

	"code.superseriousbusiness.org/gopkg/xslices"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
)

// EmptyResult provides an initialized-once
// shared SearchResult{} instance for empty returns.
var EmptyResult = &apimodel.SearchResult{
	Accounts: []*apimodel.Account{},
	Statuses: []*apimodel.Status{},
	Hashtags: []any{},
}

// Get performs a search for known accounts, statuses and tags
// visible to requester, and returns the combined results. It
// also returns next / prev pages for link URL calculation.
func (p *Processor) Get(
	ctx context.Context,
	requester *gtsmodel.Account,
	query string,
	resolve bool,
	follows bool,
	authorID string,
	page *paging.Page,
) (
	result *apimodel.SearchResult,
	next *paging.Page,
	prev *paging.Page,
	errWithCode gtserror.WithCode,
) {
	// TODO (kim): right now we're doing a double parse
	// of raw query string for accounts and statuses, but
	// in time it would be nice to find a nice API that
	// manages doing a single parse and shares more code.

	// Perform an accounts search query with parameters.
	accounts, direct, errWithCode := p.queryAccounts(ctx,
		requester,
		query,
		resolve,
		follows,
		page,
	)
	switch {
	case errWithCode == nil:

	case errors.Is(errWithCode, errNoOutput):
		// query operators would
		// produce zero results.
		return EmptyResult, nil, nil, nil

	default:
		return nil, nil, nil, errWithCode
	}

	if direct {
		if len(accounts) == 0 {
			// Return empty without paging.
			return EmptyResult, nil, nil, nil
		}

		// Return the direct account
		// as result, without paging.
		return &apimodel.SearchResult{
			Hashtags: []any{},
			Statuses: []*apimodel.Status{},
			Accounts: p.convertAPIAccounts(ctx,
				requester,
				accounts,
				true, // direct
			),
		}, nil, nil, nil
	}

	// Perform a statuses search query with parameters.
	statuses, direct, errWithCode := p.queryStatuses(ctx,
		requester,
		query,
		resolve,
		follows,
		authorID,
		page,
	)
	switch {
	case errWithCode == nil:
	case errors.Is(errWithCode, errNoOutput):
		// query operators would
		// produce zero results.
		return EmptyResult, nil, nil, nil
	default:
		return nil, nil, nil, errWithCode
	}

	if direct {
		if len(statuses) == 0 {
			// Return empty without paging.
			return EmptyResult, nil, nil, nil
		}

		// Return the direct status
		// as result, without paging.
		result = &apimodel.SearchResult{
			Hashtags: []any{},
			Accounts: []*apimodel.Account{},
			Statuses: p.c.GetVisibleAPIStatuses(ctx,
				requester,
				statuses,
				gtsmodel.FilterContextNone,
			),
		}
		return
	}

	// Perform a tags name search query.
	tags, errWithCode := p.queryTags(ctx,
		query,
		page,
	)
	if errWithCode != nil {
		return nil, nil, nil, errWithCode
	}

	// Check if there were any results returned at all.
	total := len(accounts) + len(statuses) + len(tags)
	if total == 0 {
		return EmptyResult, nil, nil, nil
	}

	// Convert fetched accounts to API models.
	apiAccounts := p.convertAPIAccounts(ctx,
		requester,
		accounts,
		false,
	)

	// Convert fetched statuses to API models.
	apiStatuses := p.c.GetVisibleAPIStatuses(ctx,
		requester,
		statuses,
		gtsmodel.FilterContextNone,
	)

	// Convert fetched tags to API.
	apiTags := make([]any, len(tags))
	if len(apiTags) != len(tags) {
		panic(gtserror.New("BCE"))
	}
	for i, tag := range tags {
		apiTag := typeutils.TagToAPITag(tag, true, nil)
		apiTags[i] = &apiTag
	}

	// Package as final result type.
	result = &apimodel.SearchResult{
		Accounts: apiAccounts,
		Statuses: apiStatuses,
		Hashtags: apiTags,
	}

	// Allocate expect no. model IDs.
	ids := make([]string, 0, total)

	// Gather a list of all model IDs, in descending order, for next / prev page params.
	ids = xslices.Gather(ids, accounts, func(a *gtsmodel.Account) string { return a.ID })
	ids = xslices.Gather(ids, statuses, func(s *gtsmodel.Status) string { return s.ID })
	ids = xslices.Gather(ids, tags, func(t *gtsmodel.Tag) string { return t.ID })
	slices.Sort(ids)
	slices.Reverse(ids)

	// Generate next / previous page details.
	next = page.Next(ids[0], ids[len(ids)-1])
	prev = page.Prev(ids[0], ids[len(ids)-1])

	return
}

// GetAccounts performs a search for known accounts visible
// to requester, and returns the results with next / prev
// pages for caller to calculate next / prev URLs.
func (p *Processor) GetAccounts(
	ctx context.Context,
	requester *gtsmodel.Account,
	query string,
	resolve bool,
	follows bool,
	page *paging.Page,
) (
	result *apimodel.SearchResult,
	next *paging.Page,
	prev *paging.Page,
	errWithCode gtserror.WithCode,
) {
	// Perform an accounts search query with parameters.
	accounts, direct, errWithCode := p.queryAccounts(ctx,
		requester,
		query,
		resolve,
		follows,
		page,
	)
	if errWithCode != nil {
		return nil, nil, nil, errWithCode
	}

	if len(accounts) == 0 {
		return EmptyResult, nil, nil, nil
	}

	// Convert fetched accounts to API models.
	apiAccounts := p.convertAPIAccounts(ctx,
		requester,
		accounts,
		direct,
	)

	// Package as final result type.
	result = &apimodel.SearchResult{
		Accounts: apiAccounts,
		Statuses: []*apimodel.Status{},
		Hashtags: []any{},
	}

	// Direct URL matches
	// don't need paging.
	if !direct {

		// Get next / previous page details.
		hi := accounts[len(accounts)-1].ID
		lo := accounts[0].ID
		next = page.Next(lo, hi)
		prev = page.Prev(lo, hi)
	}

	return
}

// GetStatuses performs a search for known statuses visible
// to requester, and returns the results with next / prev
// pages for caller to calculate next / prev URLs.
func (p *Processor) GetStatuses(
	ctx context.Context,
	requester *gtsmodel.Account,
	query string,
	resolve bool,
	follows bool,
	authorID string,
	page *paging.Page,
) (
	result *apimodel.SearchResult,
	next *paging.Page,
	prev *paging.Page,
	errWithCode gtserror.WithCode,
) {
	// Perform a statuses search query with parameters.
	statuses, direct, errWithCode := p.queryStatuses(ctx,
		requester,
		query,
		resolve,
		follows,
		authorID,
		page,
	)
	if errWithCode != nil {
		return nil, nil, nil, errWithCode
	}

	if len(statuses) == 0 {
		return EmptyResult, nil, nil, nil
	}

	// Convert fetched statuses to API models.
	apiStatuses := p.c.GetVisibleAPIStatuses(ctx,
		requester,
		statuses,
		gtsmodel.FilterContextNone,
	)

	// Package as final result type.
	result = &apimodel.SearchResult{
		Accounts: []*apimodel.Account{},
		Statuses: apiStatuses,
		Hashtags: []any{},
	}

	// Direct URL matches
	// don't need paging.
	if !direct {

		// Get next / previous page details.
		hi := statuses[len(statuses)-1].ID
		lo := statuses[0].ID
		next = page.Next(lo, hi)
		prev = page.Prev(lo, hi)
	}

	return
}

// GetTags performs a search for known tags,
// and returns the results with next / prev pages
// for caller to calculate next / prev page URLs.
func (p *Processor) GetTags(
	ctx context.Context,
	query string,
	page *paging.Page,
) (
	result *apimodel.SearchResult,
	next *paging.Page,
	prev *paging.Page,
	errWithCode gtserror.WithCode,
) {
	// Perform a tags search with params.
	tags, errWithCode := p.queryTags(ctx,
		query,
		page,
	)
	if errWithCode != nil {
		return nil, nil, nil, errWithCode
	}

	if len(tags) == 0 {
		return EmptyResult, nil, nil, nil
	}

	// Convert fetched tags to API.
	apiTags := make([]any, len(tags))
	if len(apiTags) != len(tags) {
		panic(gtserror.New("BCE"))
	}
	for i, tag := range tags {
		apiTag := typeutils.TagToAPITag(tag, true, nil)
		apiTags[i] = &apiTag
	}

	// Package as final result type.
	result = &apimodel.SearchResult{
		Accounts: []*apimodel.Account{},
		Statuses: []*apimodel.Status{},
		Hashtags: apiTags,
	}

	// Get next / previous page.
	hi := tags[len(tags)-1].ID
	lo := tags[0].ID
	next = page.Next(lo, hi)
	prev = page.Prev(lo, hi)

	return
}

func (p *Processor) queryTags(
	ctx context.Context,
	query string,
	page *paging.Page,
) (
	[]*gtsmodel.Tag,
	gtserror.WithCode,
) {
	// Trim hash symbol prefix from query.
	query = strings.TrimPrefix(query, "#")

	// Search the database for tag page by query.
	tags, err := p.dbSearchForTags(ctx, query, page)
	if err != nil {
		err := gtserror.Newf("error searching for tags by query %q: %w", query, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return tags, nil
}
