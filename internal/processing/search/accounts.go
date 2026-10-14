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

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

// Accounts does a partial search for accounts that
// match the given query. It expects input that looks
// like a namestring, and will normalize plaintext to look
// more like a namestring. For queries that include domain,
// it will only return one match at most. For namestrings
// that exclude domain, multiple matches may be returned.
//
// This behavior aligns more or less with Mastodon's API.
// See https://docs.joinmastodon.org/methods/accounts/#search.
func (p *Processor) Accounts(
	ctx context.Context,
	requester *gtsmodel.Account,
	query string,
	resolve bool,
	follows bool,
	page *paging.Page,
) (
	[]*apimodel.Account,
	gtserror.WithCode,
) {
	// If it looks like a namestring, try to separate it.
	username, domain, ok := tryParseUsernameDomain(query)

	if !ok {
		// Else query
		// for username.
		username = query
		domain = ""
	}

	var accounts []*gtsmodel.Account

	// First try a direct match for username@domain.
	account, err := p.getAccountByUsernameDomain(ctx,
		requester,
		username,
		domain,
		resolve,
	)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	if account != nil {
		// Wrap the single result as slice type.
		accounts = []*gtsmodel.Account{account}
	} else {

		// Search for accounts ~username && ~domain.
		accounts, err = p.dbSearchForAccounts(ctx,
			requester,
			&searchQuery{
				CommonSearchParams: db.CommonSearchParams{
					RequesterID: requester.ID,
					Relationship: func() db.RelationType {
						if follows {
							return db.RelationFollowing
						} else { //nolint
							return 0
						}
					}(),
				},
				UsernameQuery: username,
				DomainQuery:   domain,
			},
			page,
		)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	// Convert fetched accounts to API models.
	apiAccounts := p.convertAPIAccounts(ctx,
		requester,
		accounts,
		false,
	)

	return apiAccounts, nil
}
