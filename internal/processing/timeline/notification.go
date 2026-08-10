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

package timeline

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"slices"

	"code.superseriousbusiness.org/gopkg/log"
	"code.superseriousbusiness.org/gopkg/xslices"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

// NotificationsGet ...
func (p *Processor) NotificationsGet(
	ctx context.Context,
	requester *gtsmodel.Account,
	page *paging.Page,
	includeTypes []gtsmodel.NotificationType,
	excludeTypes []gtsmodel.NotificationType,
) (
	*apimodel.PageableResponse,
	gtserror.WithCode,
) {
	var err error

	// Ensure we have valid
	// input paging cursor.
	id.ValidatePage(page)

	// Get notification timeline for requesting account.
	timeline := p.state.Caches.Timelines.Notifications.
		MustGet(requester.ID)

	// Load status page via timeline cache, also
	// getting lo, hi values for next, prev pages.
	//
	// NOTE: this safely handles the case of a nil
	// input timeline, i.e. uncached timeline type.
	apiNotifs, lo, hi, err := timeline.Load(ctx,

		// Notif page
		// to load.
		page,

		// Database notification page loading function.
		func(page *paging.Page) ([]*gtsmodel.Notification, error) {
			return p.state.DB.GetAccountNotifications(ctx, requester.ID, page)
		},

		// Notification load function for cached timeline entries.
		func(ids []string) ([]*gtsmodel.Notification, error) {
			return p.state.DB.GetNotificationsByIDs(ctx, ids)
		},

		// Pre cache insert notification filtering.
		func(notif *gtsmodel.Notification) (delete bool) {
			if notif.OriginAccount != nil {

				// If new local account sign-up, skip normal filtering
				// because origin account won't be confirmed yet in DB.
				if notif.NotificationType == gtsmodel.NotificationAdminSignup {
					return false
				}

				// Check if notif origin account visible to requester.
				visible, err := p.visFilter.AccountVisible(ctx,
					requester,
					notif.OriginAccount,
				)
				if err != nil {
					log.Errorf(ctx, "error checking account visibility: %v", err)
					return true
				}

				if !visible {
					return true
				}

				// Check if notification origin account muted by requester.
				muted, err := p.muteFilter.AccountNotificationsMuted(ctx,
					requester.ID,
					notif.OriginAccountID,
				)
				if err != nil {
					log.Errorf(ctx, "error checking account mute: %v", err)
					return true
				}

				if muted {
					return true
				}
			}

			if notif.Status != nil {
				// Check if notif status visible to requester.
				visible, err := p.visFilter.StatusVisible(ctx,
					requester,
					notif.Status,
				)
				if err != nil {
					log.Errorf(ctx, "error checking status visibility: %v", err)
					return true
				}

				if !visible {
					return true
				}

				// Check if notification status is muted to requester.
				muted, err := p.muteFilter.StatusNotificationsMuted(ctx,
					requester,
					notif.Status,
				)
				if err != nil {
					log.Errorf(ctx, "error checking status mute: %v", err)
					return true
				}

				if muted {
					return true
				}
			}

			return false
		},

		// Frontend API model preparation function.
		func(notif *gtsmodel.Notification) (*apimodel.Notification, error) {
			var filters []apimodel.FilterResult

			// If include types were provided, check notification *is* part of
			// the included list. Or if exclude types were provided, check
			// notification *isn't* part of the excluded list. Else, return nil.
			//
			// TODO: this should perhaps be moved to a separate postFilter
			// function in the cache, but it's difficult when status filtering
			// (as below) also returns results used in API model preparation.
			if (len(includeTypes) > 0 && !slices.Contains(includeTypes, notif.NotificationType)) ||
				(len(excludeTypes) > 0 && slices.Contains(excludeTypes, notif.NotificationType)) {
				return nil, nil
			}

			if notif.Status != nil {
				var hide bool

				// Check whether this status is filtered by requester in this context.
				filters, hide, err = p.statusFilter.StatusFilterResultsInContext(ctx,
					requester,
					notif.Status,
					gtsmodel.FilterContextNotifications,
				)
				if err != nil {
					return nil, err
				} else if hide {
					return nil, nil
				}
			}

			// Finally, pass notification to get converted to frontend API model.
			apiNotif, err := p.converter.NotificationToAPINotification(ctx, notif)
			if err != nil {
				return nil, err
			}

			if apiNotif.Status != nil {
				// Set any filters on notif status.
				apiNotif.Status.Filtered = filters
			}

			return apiNotif, nil
		},
	)

	if err != nil {
		err := gtserror.Newf("error loading timeline: %w", err)
		return nil, gtserror.WrapWithCode(http.StatusInternalServerError, err)
	}

	// Prepare timeline query.
	query := make(url.Values, 2)
	if len(includeTypes) > 0 {
		query["types[]"] = notificationTypes(includeTypes)
	}
	if len(excludeTypes) > 0 {
		query["exclude_types[]"] = notificationTypes(excludeTypes)
	}

	// Package returned API statuses as pageable response.
	return paging.PackageResponse(paging.ResponseParams{
		Items: xslices.ToAny(apiNotifs),
		Path:  "/api/v1/notifications",
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
		Query: query,
	}), nil
}

func (p *Processor) NotificationGet(ctx context.Context, account *gtsmodel.Account, targetNotifID string) (*apimodel.Notification, gtserror.WithCode) {
	notif, err := p.state.DB.GetNotificationByID(ctx, targetNotifID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("error getting from db: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if notif == nil {
		const text = "notification not found"
		return nil, gtserror.NewErrorNotFound(
			errors.New(text),
			text,
		)
	}

	if notif.TargetAccountID != account.ID {
		err := gtserror.New("requester does not match notification target")
		return nil, gtserror.NewErrorNotFound(err)
	}

	// NOTE: we specifically don't do any filtering
	// or mute checking for a notification directly
	// fetched by ID. only from timelines etc.

	apiNotif, err := p.converter.NotificationToAPINotification(ctx, notif)
	if err != nil {
		err := gtserror.Newf("error converting to api model: %w", err)
		return nil, gtserror.WrapWithCode(http.StatusInternalServerError, err)
	}

	return apiNotif, nil
}

func (p *Processor) NotificationsClear(ctx context.Context, authed *apiutil.Auth) gtserror.WithCode {
	// Delete all notifications of all types that target the authorized account.
	if err := p.state.DB.DeleteNotifications(ctx, nil, authed.Account.ID, ""); err != nil && !errors.Is(err, db.ErrNoEntries) {
		return gtserror.NewErrorInternalError(err)
	}

	return nil
}

// notificationTypes returns given notification types as string values.
func notificationTypes(types []gtsmodel.NotificationType) []string {
	typestrs := make([]string, len(types))
	if len(typestrs) != len(types) {
		panic(gtserror.New("BCE"))
	}
	for i, typ := range types {
		typestrs[i] = typ.String()
	}
	return typestrs
}
