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

	"code.superseriousbusiness.org/gopkg/log"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/federation"
	"code.superseriousbusiness.org/gotosocial/internal/filter/visibility"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/processing/common"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/surfacing"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
)

type Processor struct {
	// common processor logic
	c *common.Processor

	state     *state.State
	federator *federation.Federator
	converter *typeutils.Converter
	visFilter *visibility.Filter
	surfacer  *surfacing.Surfacer
}

func New(
	common *common.Processor,
	state *state.State,
	federator *federation.Federator,
	converter *typeutils.Converter,
	visFilter *visibility.Filter,
	surfacer *surfacing.Surfacer,
) Processor {
	return Processor{
		c:         common,
		state:     state,
		federator: federator,
		converter: converter,
		visFilter: visFilter,
		surfacer:  surfacer,
	}
}

// convertAPIAccount performs visibility checks on the given
// account without requesting account checks, to ensure it isn't
// suspended, unconfirmed, etc. then performs block checks between
// the two accounts and appropriately returns either a 'public'
// API model response, or a stubbed-out blocked model response.

func (p *Processor) convertAPIAccount(
	ctx context.Context,
	requester *gtsmodel.Account,
	account *gtsmodel.Account,
	direct bool,
) (
	*apimodel.Account,
	error,
) {
	if !direct && account.IsInstance() {
		// Don't show any
		// instance accounts.
		return nil, nil
	}

	// Check whether account is visible to requesting user,
	// this accounts for suspended, unconfirmed, blocks, etc.
	visible, err := p.visFilter.AccountVisible(ctx, requester,
		account)
	if err != nil {
		return nil, gtserror.Newf("error checking account %s visibility: %w", account.URI, err)
	}

	var apiAccount *apimodel.Account

	if !visible {
		if direct {
			// Account should not be fully visible, only return a stubbed-out model.
			apiAccount, err = p.converter.AccountToAPIAccountBlocked(ctx, account)
		}
	} else {
		// Else, convert and return the account as a regular public model.
		apiAccount, err = p.converter.AccountToAPIAccountPublic(ctx, account)
	}

	if err != nil {
		return nil, gtserror.Newf("error converting account %s: %w", account.URI, err)
	}

	return apiAccount, nil
}

// convertAPIAccounts is similar to GetVisibleAPIAccounts()
// except that it supports returning stubbed-out accounts
// when they fail visibility checks if 'direct' = true.
func (p *Processor) convertAPIAccounts(
	ctx context.Context,
	requester *gtsmodel.Account,
	accounts []*gtsmodel.Account,
	direct bool,
) []*apimodel.Account {
	apiAccounts := make([]*apimodel.Account, 0, len(accounts))
	for _, account := range accounts {

		// Convert to frontend API account model.
		apiAccount, err := p.convertAPIAccount(ctx,
			requester,
			account,
			direct,
		)
		if err != nil {
			log.Error(ctx, err)
			continue
		}

		if apiAccount == nil {
			// Account was suspended,
			// user not confirmed etc.
			continue
		}

		// Append converted account to result slice.
		apiAccounts = append(apiAccounts, apiAccount)
	}

	return apiAccounts
}
