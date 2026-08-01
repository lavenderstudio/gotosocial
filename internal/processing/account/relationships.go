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

package account

import (
	"cmp"
	"context"
	"errors"
	"slices"

	"code.superseriousbusiness.org/gopkg/log"
	"code.superseriousbusiness.org/gopkg/xslices"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

// RelationshipGet returns the relationship
// between requester account and target account.
func (p *Processor) RelationshipGet(
	ctx context.Context,
	requester *gtsmodel.Account,
	targetAcctID string,
) (*apimodel.Relationship, gtserror.WithCode) {
	return p.c.APIRelationship(ctx, requester, targetAcctID)
}

// FollowersGet fetches a list of
// the target account's followers.
func (p *Processor) FollowersGet(
	ctx context.Context,
	requester *gtsmodel.Account,
	targetAccountID string,
	page *paging.Page,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	// Fetch target account to check it exists, and visibility of requester->target.
	target, errWithCode := p.c.GetVisibleTargetAccount(ctx, requester, targetAccountID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if target.IsInstance() || target.IsRelayActor() {
		// Hide for instance and relay actors.
		return paging.EmptyResponse(), nil
	}

	// If account isn't requesting its own followers list,
	// but instead the list for a local account that has
	// hide_followers set, just return an empty array.
	if targetAccountID != requester.ID &&
		target.IsLocal() &&
		*target.Settings.HideCollections {
		return paging.EmptyResponse(), nil
	}

	path := "/api/v1/accounts/" + targetAccountID + "/followers"
	return p.c.FollowersGet(ctx, target, requester, page, path)
}

// FollowingGet fetches a list of accounts
// that the target account is following.
func (p *Processor) FollowingGet(
	ctx context.Context,
	requester *gtsmodel.Account,
	targetAccountID string,
	page *paging.Page,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	// Fetch target account to check it exists, and visibility of requester->target.
	target, errWithCode := p.c.GetVisibleTargetAccount(ctx, requester, targetAccountID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if target.IsInstance() || target.IsRelayActor() {
		// Hide for instance and relay actors.
		return paging.EmptyResponse(), nil
	}

	// If account isn't requesting its own following list,
	// but instead the list for a local account that has
	// hide_followers set, just return an empty array.
	if targetAccountID != requester.ID &&
		target.IsLocal() &&
		*target.Settings.HideCollections {
		return paging.EmptyResponse(), nil
	}

	path := "/api/v1/accounts/" + targetAccountID + "/following"
	return p.c.FollowingGet(ctx, target, requester, page, path)
}

// ConnectedDomainsGet returns an alphabetical,
// deduplicated, depunified list of all domains
// that the target account is connected with
// via a following and/or follower relationship.
func (p *Processor) ConnectedDomainsGet(
	ctx context.Context,
	requester *gtsmodel.Account,
	targetAccountID string,
) ([]string, gtserror.WithCode) {
	// Fetch target account to check it exists, and visibility of requester->target.
	target, errWithCode := p.c.GetVisibleTargetAccount(ctx, requester, targetAccountID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if target.IsInstance() {
		// Instance actor only
		// follows relay actors,
		// just return nil here.
		return nil, nil
	}

	if target.IsLocalUserAccount() &&
		*target.Settings.HideCollections &&
		(requester == nil || requester.ID != target.ID) {
		// Target account
		// is not requester,
		// and hides collections.
		return nil, nil
	}

	// OK to show connected
	// domains to this requester.

	// Get all follows that target targetAccountID.
	followers, err := p.state.DB.GetAccountFollowers(ctx, targetAccountID, nil)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("db error getting followers: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Get all follows that are owned by targetAccountID.
	following, err := p.state.DB.GetAccountFollows(ctx, targetAccountID, nil)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("db error getting followers: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Map for deduplication.
	l := len(followers) + len(following)
	domainsMap := make(map[string]any, l)

	// Function to depunify and dedupe a domain.
	uniqueDomain := func(domain string) (string, bool) {
		// Depunify the domain.
		domain, err = util.DePunify(domain)
		if err != nil {
			log.Errorf(ctx, "error depunifying follower domain: %v", err)
			return "", false
		}

		// Check if already gathered.
		if _, gathered := domainsMap[domain]; gathered {
			// Do nothing.
			return "", false
		}

		// Mark domain in dedupe map.
		domainsMap[domain] = struct{}{}
		return domain, true
	}

	// Get our own domain
	// once outside the loops.
	ourDomain := config.GetAccountDomain()

	// Collect domains
	// of all followers.
	domains := xslices.GatherIf(
		nil,
		followers,
		func(f *gtsmodel.Follow) (string, bool) {
			// Use follow owner domain,
			// fall back to our domain
			// for local accounts.
			domain := cmp.Or(
				f.Account.Domain,
				ourDomain,
			)
			return uniqueDomain(domain)
		},
	)

	// Collect domains
	// of all following.
	domains = append(domains,
		xslices.GatherIf(
			nil,
			following,
			func(f *gtsmodel.Follow) (string, bool) {
				// Use follow target domain,
				// fall back to our domain
				// for local accounts.
				domain := cmp.Or(
					f.TargetAccount.Domain,
					ourDomain,
				)
				return uniqueDomain(domain)
			},
		)...,
	)

	// Sort alphabetically
	// before returning.
	slices.Sort(domains)
	return domains, nil
}
