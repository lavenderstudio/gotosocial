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
	"net/url"

	"code.superseriousbusiness.org/gopkg/log"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

func (p *Processor) DirectoryGet(
	ctx context.Context,
	requester *gtsmodel.Account,
	page *paging.Page,
	offset int,
	orderBy gtsmodel.DirectoryOrderBy,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	// Get specified page of accounts.
	accounts, err := p.state.DB.GetDirectoryPage(ctx, page, offset, orderBy)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting accounts: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Check items length.
	count := len(accounts)
	if count == 0 {
		return paging.EmptyResponse(), nil
	}

	var (
		// Preallocate expected items.
		items = make([]any, 0, count)

		// Set paging low / high IDs.
		lo = accounts[count-1].ID
		hi = accounts[0].ID
	)

	// Logic for converting accounts to items varies
	// depending on whether the request is authenticated.
	if requester != nil {
		// Convert fetched accounts to API accounts,
		// ignoring accounts that block the requester.
		for _, a := range accounts {
			blocked, err := p.state.DB.IsBlocked(ctx, a.ID, requester.ID)
			if err != nil {
				log.Errorf(ctx, "db error checking block: %v", err)
				continue
			}

			if blocked {
				// Just skip
				// this one.
				continue
			}

			item, err := p.converter.AccountToAPIAccountPublic(ctx, a)
			if err != nil {
				log.Errorf(ctx, "error converting to web status: %v", err)
				continue
			}

			items = append(items, item)
		}
	} else {
		// Just convert fetched accounts to API accounts,
		// there's no requester to check for a block against.
		for _, a := range accounts {
			item, err := p.converter.AccountToAPIAccountPublic(ctx, a)
			if err != nil {
				log.Errorf(ctx, "error convering to web status: %v", err)
				continue
			}

			items = append(items, item)
		}
	}

	// Prepare response.
	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/api/v1/directory",
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
		Query: url.Values{
			apiutil.DirectoryOrderKey: []string{orderBy.String()},
		},
	}), nil
}

func (p *Processor) WebDirectoryGet(
	ctx context.Context,
	page *paging.Page,
	orderBy gtsmodel.DirectoryOrderBy,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	// Get specified page of accounts.
	accounts, err := p.state.DB.GetDirectoryPage(ctx, page, 0, orderBy)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting accounts: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Check items length.
	count := len(accounts)
	if count == 0 {
		return paging.EmptyResponse(), nil
	}

	var (
		// Preallocate expected items.
		items = make([]any, 0, count)

		// Set paging low / high IDs.
		lo = accounts[count-1].ID
		hi = accounts[0].ID
	)

	// Populate items.
	for _, a := range accounts {
		apiAcct, err := p.converter.AccountToAPIAccountPublic(ctx, a)
		if err != nil {
			log.Errorf(ctx, "error convering to api account: %v", err)
			continue
		}

		item, err := p.converter.AccountToWebAccount(ctx, a, apiAcct)
		if err != nil {
			log.Errorf(ctx, "error convering to web account: %v", err)
			continue
		}

		items = append(items, item)
	}

	// Prepare response.
	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/directory",
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
		Query: url.Values{
			apiutil.DirectoryOrderKey: []string{orderBy.String()},
		},
	}), nil
}
