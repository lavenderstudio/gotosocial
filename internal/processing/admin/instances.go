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

package admin

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	"code.superseriousbusiness.org/gopkg/log"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

func (p *Processor) InstancesGet(
	ctx context.Context,
	page *paging.Page,
	domain string,
	orderBy gtsmodel.InstanceOrderBy,
	withErrorsOnly bool,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	// Get specified page of instances.
	instances, err := p.state.DB.GetInstancesPage(ctx, page, domain, orderBy, withErrorsOnly)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting instances: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Check items length.
	count := len(instances)
	if count == 0 {
		return paging.EmptyResponse(), nil
	}

	var (
		// Preallocate expected items.
		items = make([]any, 0, count)

		// Set paging low / high IDs.
		lo = instances[count-1].ID
		hi = instances[0].ID
	)

	for _, a := range instances {
		item, err := p.converter.InstanceToAdminAPIInstance(ctx, a)
		if err != nil {
			log.Errorf(ctx, "error converting to admin API instance: %v", err)
			continue
		}

		items = append(items, item)
	}

	// Prepare paging query kvs.
	query := url.Values{
		apiutil.OrderKey:               []string{orderBy.String()},
		apiutil.AdminWithErrorsOnlyKey: []string{strconv.FormatBool(withErrorsOnly)},
	}
	if domain != "" {
		query.Add(apiutil.DomainKey, domain)
	}

	// Prepare response.
	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/api/v1/admin/instances",
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
		Query: query,
	}), nil
}

func (p *Processor) InstanceGet(ctx context.Context, id string) (*apimodel.AdminInstance, gtserror.WithCode) {
	// Get instance with specified ID.
	instance, err := p.state.DB.GetInstanceByID(ctx, id)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting instance: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if instance == nil {
		err := gtserror.Newf("instance not found in the db: %w", err)
		return nil, gtserror.NewErrorNotFound(err)
	}

	item, err := p.converter.InstanceToAdminAPIInstance(ctx, instance)
	if err != nil {
		err := gtserror.Newf("error converting to admin API instance: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return item, nil
}
