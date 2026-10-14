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
	"strings"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// Lookup does a quick, non-resolving search for accounts that
// match the given query. It expects input that looks like a
// namestring, and will normalize plaintext to look more like
// a namestring. Will only ever return one account, and only on
// an exact match.
//
// This behavior aligns more or less with Mastodon's API.
// See https://docs.joinmastodon.org/methods/accounts/#lookup
func (p *Processor) Lookup(
	ctx context.Context,
	requester *gtsmodel.Account,
	query string,
) (
	*apimodel.Account,
	gtserror.WithCode,
) {
	var ok bool
	var username, domain string

	// If this looks like a namestring,
	// attempt to separate into parts.
	if strings.Contains(query, "@") {
		username, domain, ok = tryParseUsernameDomain(query)
	}

	if !ok {
		// Else query for
		// local username.
		username = query
		domain = ""
	}

	// Try lookup the queried account by its username @ domain.
	account, err := p.state.DB.GetAccountByUsernameDomain(ctx,
		username,
		domain,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting account %s: %w", query, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if account == nil {
		const text = "target account not found"
		return nil, gtserror.NewErrorNotFound(
			errors.New(text),
			text,
		)
	}

	// Convert to frontend API account model.
	apiAccount, err := p.convertAPIAccount(ctx,
		requester,
		account,

		// not direct, don't
		// show blocked etc.
		false,
	)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Specifically check
	// for an unreturnable
	// account, i.e. deleted
	// or suspended etc.
	if apiAccount == nil {
		const text = "target account not found"
		return nil, gtserror.NewErrorNotFound(
			errors.New(text),
			text,
		)
	}

	return apiAccount, nil
}
