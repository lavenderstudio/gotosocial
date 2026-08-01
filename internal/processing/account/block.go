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

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

func (p *Processor) BlocksGet(
	ctx context.Context,
	requester *gtsmodel.Account,
	page *paging.Page,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	// When returning blocks via the client API,
	// stub blocked accounts so user doesn't have
	// to see profile pic, bio, etc. of blockees.
	path := "/api/v1/blocks"
	return p.c.BlocksGet(ctx,
		requester,
		true, // stub
		page,
		path,
	)
}

func (p *Processor) BlockCreate(
	ctx context.Context,
	requester *gtsmodel.Account,
	targetAccountID string,
) (*apimodel.Relationship, gtserror.WithCode) {
	return p.c.BlockCreate(ctx, requester, targetAccountID)
}

func (p *Processor) BlockRemove(
	ctx context.Context,
	requester *gtsmodel.Account,
	targetAccountID string,
) (*apimodel.Relationship, gtserror.WithCode) {
	return p.c.BlockRemove(ctx, requester, targetAccountID)
}
