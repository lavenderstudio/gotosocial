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
	"context"
	"errors"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

// FollowRequestAccept accepts a follow request on behalf of
// requester, where follow requester is the given accountID.
func (p *Processor) FollowRequestAccept(
	ctx context.Context,
	requester *gtsmodel.Account,
	accountID string,
) (*apimodel.Relationship, gtserror.WithCode) {
	return p.c.FollowRequestAccept(ctx, requester, accountID)
}

// FollowRequestReject rejects a follow request on behalf of
// requester, where follow requester is the given accountID.
func (p *Processor) FollowRequestReject(
	ctx context.Context,
	requester *gtsmodel.Account,
	accountID string,
) (*apimodel.Relationship, gtserror.WithCode) {
	return p.c.FollowRequestReject(ctx, requester, accountID)
}

// OutgoingFollowRequestsGet returns a page of accounts
// which target the requester with follow requests.
//
// The pagePath param should be set to the API path that's being
// used to call this function, eg "/api/v1/follow_requests".
func (p *Processor) FollowRequestsGet(
	ctx context.Context,
	requester *gtsmodel.Account,
	page *paging.Page,
	pagePath string,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	return p.c.FollowRequestsGet(ctx, requester, page, pagePath)
}

// OutgoingFollowRequestsGet returns a page of accounts
// targeted by follow requests owned by the requester.
//
// The pagePath param should be set to the API path that's being
// used to call this function, eg "/api/v1/follow_requesting".
func (p *Processor) OutgoingFollowRequestsGet(
	ctx context.Context,
	requester *gtsmodel.Account,
	page *paging.Page,
	pagePath string,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	// Fetch follow requests originating from the given requesting account model.
	followRequests, err := p.state.DB.GetAccountFollowRequesting(ctx, requester.ID, page)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Check for empty response.
	count := len(followRequests)
	if count == 0 {
		return paging.EmptyResponse(), nil
	}

	// Get the lowest and highest
	// ID values, used for paging.
	lo := followRequests[count-1].ID
	hi := followRequests[0].ID

	// Func to fetch follow source at index.
	getIdx := func(i int) *gtsmodel.Account {
		return followRequests[i].TargetAccount
	}

	// Get a filtered slice of public API account models.
	items := p.c.GetVisibleAPIAccountsPaged(ctx,
		requester,
		getIdx,
		count,
	)

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  pagePath,
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
	}), nil
}
