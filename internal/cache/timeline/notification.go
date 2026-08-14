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
	"slices"

	"code.superseriousbusiness.org/gopkg/log"
	"code.superseriousbusiness.org/gopkg/xslices"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"codeberg.org/gruf/go-structr"
)

// NotificationMeta contains minimum viable metadata
// about a notifcation in order to cache a timeline.
type NotificationMeta struct {
	ID              string
	OriginAccountID string
	StatusOrEditID  string

	// loaded is a temporary field that may be
	// set for a newly loaded timeline notifs
	// so notifications don't need to be loaded
	// from the database twice in succession.
	//
	// i.e. this will only be set if the notif
	// was newly inserted into the timeline cache.
	// for existing cache items this will be nil.
	loaded *gtsmodel.Notification
}

// NotificationTimeline provides a concurrency-safe sliding
// window cache of the freshest notifications in a timeline.
// Internally, only NotificationMeta{} objects themselves are
// stored, loading the actual notifications when necessary.
//
// See StatusTimeline{} for notes on design.
type NotificationTimeline struct {

	// underlying timeline cache of *NotificationMeta{},
	// primary-keyed by ID, with extra indices below.
	cache structr.Timeline[*NotificationMeta, string]

	// preloader synchronizes preload
	// state of the timeline cache.
	preloader preloader

	// fast-access cache indices.
	idx_ID              *structr.Index // nolint:revive
	idx_OriginAccountID *structr.Index // nolint:revive
	idx_StatusOrEditID  *structr.Index // nolint:revive

	// cutoff and maximum item lengths.
	// the timeline is trimmed back to
	// cutoff on each call to Trim(),
	// and maximum len triggers a Trim().
	//
	// the timeline itself does not
	// limit items due to complexities
	// it would introduce, so we apply
	// a 'cut-off' at regular intervals.
	cut, max int
}

// Init will initialize the timeline for usage,
// by preparing internal indices etc. This also
// sets the given max capacity for Trim() operations.
func (t *NotificationTimeline) Init(cap int) {
	t.cache.Init(structr.TimelineConfig[*NotificationMeta, string]{

		// Timeline item primary key field.
		PKey: structr.IndexConfig{Fields: "ID"},

		// Additional indexed fields.
		Indices: []structr.IndexConfig{
			{Fields: "OriginAccountID", Multiple: true},
			{Fields: "StatusOrEditID", Multiple: true},
		},

		// Timeline item copy function.
		Copy: func(n *NotificationMeta) *NotificationMeta {
			return &NotificationMeta{
				ID:              n.ID,
				OriginAccountID: n.OriginAccountID,
				StatusOrEditID:  n.StatusOrEditID,
				loaded:          nil, // NEVER stored
			}
		},
	})

	// Get fast index lookup ptrs.
	t.idx_ID = t.cache.Index("ID")
	t.idx_OriginAccountID = t.cache.Index("OriginAccountID")
	t.idx_StatusOrEditID = t.cache.Index("StatusOrEditID")

	// Set maximum capacity and
	// cutoff threshold we trim to.
	t.cut = int(0.60 * float64(cap))
	t.max = cap
}

// Preload will fill the NotificationTimeline{} cache with
// the latest sliding window of notification metadata for the
// timeline type returned by database 'loadPage' function.
//
// This function is concurrency-safe and repeated calls to
// it when already preloaded will be no-ops. To trigger a
// preload as being required, call .Clear().
func (t *NotificationTimeline) Preload(

	// loadPage should load the timeline of given page for cache hydration.
	loadPage func(*paging.Page) (notifs []*gtsmodel.Notification, err error),

	// filter can be used to perform filtering of returned
	// notifs BEFORE insert into cache. i.e. this will effect
	// what actually gets stored in the timeline cache.
	filter func(each *gtsmodel.Notification) (delete bool),
) (
	n int,
	err error,
) {
	err = t.preloader.CheckPreload(func() error {
		n, err = t.preload(loadPage, filter)
		return err
	})
	return
}

// preload contains the core logic of
// Preload(), without t.preloader checks.
func (t *NotificationTimeline) preload(
	loadPage func(page *paging.Page) (notifs []*gtsmodel.Notification, err error),
	filter func(each *gtsmodel.Notification) (delete bool),
) (
	int,
	error,
) {
	if loadPage == nil {
		panic("nil load page func")
	}

	// Clear timeline
	// before preload.
	t.cache.Clear()

	// Our starting, page at the top
	// of the possible timeline.
	page := new(paging.Page)
	order := paging.OrderDescending
	page.Max.Order = order
	page.Max.Value = plus1hULID()
	page.Min.Order = order
	page.Min.Value = ""
	page.Limit = 100

	// Prepare a slice for gathering notification meta.
	metas := make([]*NotificationMeta, 0, page.Limit)

	var n int
	for n < t.cut {
		// Load page of timeline notifs.
		notifs, err := loadPage(page)
		if err != nil {
			return n, gtserror.Newf("error loading notifications: %w", err)
		}

		// No more notifs from
		// load function = at end.
		if len(notifs) == 0 {
			break
		}

		// Update our next page cursor from notifs.
		page.Max.Value = notifs[len(notifs)-1].ID

		// Perform any filtering on newly loaded notifs.
		notifs = doFilterNotifications(notifs, filter)

		// After filtering no more
		// notifs remain, retry.
		if len(notifs) == 0 {
			continue
		}

		// Convert notifications to meta and insert,
		// setting 'n' to current timeline length.
		metas = toNotificationMeta(metas[:0], notifs)
		n = t.cache.Insert(metas...)
	}

	return n, nil
}

// Load will load given page of notification timeline. First it
// will prioritize fetching notifications from the sliding window
// that is the timeline cache of latest notifications, else it will
// fall back to loading from the database using callback funcs.
// The returned string values are the low / high notification ID
// paging values, used in calculating next / prev page links.
func (t *NotificationTimeline) Load(
	ctx context.Context,
	page *paging.Page,

	// loadPage should load the timeline of given page for cache hydration.
	loadPage func(page *paging.Page) (notifs []*gtsmodel.Notification, err error),

	// loadIDs should load notifications with given IDs, this is used
	// to load notifications of already cached entries in the timeline.
	loadIDs func(ids []string) (notif []*gtsmodel.Notification, err error),

	// filter performs filtering of returned
	// notifs BEFORE they are inserted into the cache.
	filter func(each *gtsmodel.Notification) (delete bool),

	// prepareAPI should prepare internal notification model to frontend API model.
	prepareAPI func(notif *gtsmodel.Notification) (apiNotif *apimodel.Notification, err error),
) (
	[]*apimodel.Notification,
	string, // lo
	string, // hi
	error,
) {
	var err error

	// Get paging details.
	lo := page.Min.Value
	hi := page.Max.Value
	limit := page.Limit
	order := page.Order()
	dir := toDirection(order)
	if limit <= 0 {

		// a page limit MUST be set!
		// this shouldn't be possible
		// but we check anyway to stop
		// chance of limitless db calls!
		panic("invalid page limit")
	}

	// Ensure timeline has been preloaded.
	_, err = t.Preload(loadPage, filter)
	if err != nil {
		return nil, "", "", err
	}

	// Use a copy of current page so
	// we can repeatedly update it.
	nextPg := new(paging.Page)
	*nextPg = *page
	nextPg.Min.Value = lo
	nextPg.Max.Value = hi

	// Load a little more than limit to
	// reduce chance of db calls below.
	limitPtr := util.Ptr(limit + 10)

	// First we attempt to load status
	// metadata entries from the timeline
	// cache, up to given limit.
	metas := t.cache.Select(
		util.PtrIf(lo),
		util.PtrIf(hi),
		limitPtr,
		dir,
	)

	// Preallocate slice of interstitial models.
	metas = slices.Grow(metas, *limitPtr-len(metas))

	// Preallocate slice of required notif API models.
	apiNotifs := make([]*apimodel.Notification, 0, limit)

	if len(metas) > 0 {
		// Before we can do any filtering, we need
		// to load notif models for cached entries.
		err = loadNotifications(metas, loadIDs)
		if err != nil {
			return nil, "", "", gtserror.Newf("error loading notifications: %w", err)
		}

		// Update nextPg cursor parameter for database query.
		nextPageParams(nextPg, metas[len(metas)-1].ID, order)

		// Prepare frontend API models for
		// the loaded notifs. For now this
		// also does its own extra filtering.
		apiNotifs = prepareNotifications(ctx,
			metas,
			prepareAPI,
			apiNotifs,
			limit,
		)
	}

	// If not enough cached timeline
	// notifs were found for page,
	// we need to call to database.
	if len(apiNotifs) < limit {

		// Pass to main timeline database load function.
		apiNotifs, err = loadNotificationTimeline(ctx,
			nextPg,
			metas,
			apiNotifs,
			loadPage,
			filter,
			prepareAPI,
		)
		if err != nil {
			return nil, "", "", err
		}
	}

	// Reset values.
	lo, hi = "", ""

	if len(apiNotifs) > 0 {
		// Set returned lo, hi paging values.
		lo = apiNotifs[len(apiNotifs)-1].ID
		hi = apiNotifs[0].ID
	}

	if order.Ascending() {
		// The caller always expects notifs
		// to be returned in DESC order, but we
		// build the notif slice in paging order.
		// If paging ASC, we need to reverse the
		// returned notifs and paging values.
		slices.Reverse(apiNotifs)
		lo, hi = hi, lo
	}

	return apiNotifs, lo, hi, nil
}

// loadNotificationTimeline encapsulates the logic of iteratively
// attempt to load a notification timeline page from the database,
// that is in the form of given callback functions. these will
// then be prepared to frontend API models for return.
func loadNotificationTimeline(
	ctx context.Context,
	nextPg *paging.Page,
	metas []*NotificationMeta,
	apiNotifs []*apimodel.Notification,
	loadPage func(page *paging.Page) (notifs []*gtsmodel.Notification, err error),
	filter func(each *gtsmodel.Notification) (delete bool),
	prepareAPI func(notif *gtsmodel.Notification) (apiNotif *apimodel.Notification, err error),
) (
	[]*apimodel.Notification,
	error,
) {
	if loadPage == nil {
		panic("nil load page func")
	}

	// Extract paging params, in particular
	// limit is used separate to nextPg to
	// determine the *expected* return limit,
	// not just what we use in db queries.
	returnLimit := nextPg.Limit
	order := nextPg.Order()

	// Perform maximum of 5 load
	// attempts fetching notifs.
	for i := 0; i < 5; i++ {

		// Update page limit to the *remaining*
		// limit of total we're expected to return.
		nextPg.Limit = returnLimit - len(apiNotifs)
		if nextPg.Limit <= 0 {
			break
		}

		// But load a bit more than
		// limit to reduce db calls.
		nextPg.Limit += 10

		// Load next timeline notifis.
		notifs, err := loadPage(nextPg)
		if err != nil {
			return nil, gtserror.Newf("error loading timeline: %w", err)
		}

		// No more notifs from
		// load function = at end.
		if len(notifs) == 0 {
			break
		}

		// Update nextPg cursor parameter for next database query.
		nextPageParams(nextPg, notifs[len(notifs)-1].ID, order)

		// Perform filtering on newly loaded notifs.
		notifs = doFilterNotifications(notifs, filter)

		// After filtering no more
		// notifs remain, retry load.
		if len(notifs) == 0 {
			continue
		}

		// Convert to our interstitial notif meta type.
		metas = toNotificationMeta(metas[:0], notifs)

		// Prepare frontend API models for
		// the loaded notifs. For now this
		// also does its own extra filtering.
		apiNotifs = prepareNotifications(ctx,
			metas,
			prepareAPI,
			apiNotifs,
			returnLimit,
		)
	}

	return apiNotifs, nil
}

// InsertOne allows you to insert a single notification into the timeline.
func (t *NotificationTimeline) InsertOne(notif *gtsmodel.Notification) {

	// If timeline no preloaded, i.e.
	// no-one using it, don't insert.
	if !t.preloader.Check() {
		return
	}

	// If item is beyond end of the
	// timeline, don't bother adding.
	if tailID := t.cache.Tail(); //
	tailID == nil || notif.ID < *tailID {
		return
	}

	// Insert new timeline notification.
	t.cache.Insert(&NotificationMeta{
		ID:              notif.ID,
		OriginAccountID: notif.OriginAccountID,
		StatusOrEditID:  notif.StatusOrEditID,
	})
}

// RemoveByNotificationIDs ...
func (t *NotificationTimeline) RemoveByNotificationIDs(notifIDs ...string) {
	keys := make([]structr.Key, len(notifIDs))
	if len(keys) != len(notifIDs) {
		panic(gtserror.New("BCE"))
	}

	// Nil check indices outside loop.
	if t.idx_OriginAccountID == nil {
		panic("indices are nil")
	}

	// Convert IDs to index keys.
	for i, id := range notifIDs {
		keys[i] = structr.MakeKey(id)
	}

	// Invalidate cached entries with IDs.
	t.cache.Invalidate(t.idx_ID, keys...)
}

// RemoveByOriginAccountIDs removes all cached timeline entries originating from account IDs.
func (t *NotificationTimeline) RemoveByOriginAccountIDs(accountIDs ...string) {
	keys := make([]structr.Key, len(accountIDs))
	if len(keys) != len(accountIDs) {
		panic(gtserror.New("BCE"))
	}

	// Nil check indices outside loop.
	if t.idx_OriginAccountID == nil {
		panic("indices are nil")
	}

	// Convert accountIDs to index keys.
	for i, id := range accountIDs {
		keys[i] = structr.MakeKey(id)
	}

	// Invalidate all cached entries with account IDs.
	t.cache.Invalidate(t.idx_OriginAccountID, keys...)
}

// RemoveByStatusOrEditIDs removes all cached timeline entries pertaining to status or edit IDs.
func (t *NotificationTimeline) RemoveByStatusOrEditIDs(statusOrEditIDs ...string) {
	keys := make([]structr.Key, len(statusOrEditIDs))
	if len(keys) != len(statusOrEditIDs) {
		panic(gtserror.New("BCE"))
	}

	// Nil check indices outside loop.
	if t.idx_StatusOrEditID == nil {
		panic("indices are nil")
	}

	// Convert given IDs to index keys.
	for i, id := range statusOrEditIDs {
		keys[i] = structr.MakeKey(id)
	}

	// Invalidate all cached entries with given IDs.
	t.cache.Invalidate(t.idx_StatusOrEditID, keys...)
}

// Trim will ensure that receiving timeline is less than or
// equal in length to the given threshold percentage of the
// timeline's preconfigured maximum capacity. This will always
// trim from the bottom-up to prioritize streamed inserts.
func (t *NotificationTimeline) Trim() { t.cache.Trim(t.cut, structr.Asc) }

// Clear will mark the entire timeline as requiring preload,
// which will trigger a clear and reload of the entire thing.
func (t *NotificationTimeline) Clear() {

	// we clear the cache within the protection of
	// the "preloader" mechanism to ensure we don't
	// drop the cache while a preload is in progress.
	t.preloader.Clear(t.cache.Clear)
}

// prepareNotifications takes a slice of cached (or, freshly loaded!) NotificationMeta{}
// models, and uses given functions to return prepared frontend models of them.
func prepareNotifications(
	ctx context.Context,
	meta []*NotificationMeta,
	prepareAPI func(*gtsmodel.Notification) (*apimodel.Notification, error),
	apiNotifs []*apimodel.Notification,
	limit int,
) []*apimodel.Notification {
	switch { //nolint:gocritic
	case prepareAPI == nil:
		panic("nil prepare fn")
	}

	// Iterate the given StatusMeta objects for pre-prepared
	// frontend models, otherwise attempting to prepare them.
	for _, meta := range meta {

		// Check if we have prepared enough
		// API notifs for caller to return.
		if len(apiNotifs) >= limit {
			break
		}

		if meta.loaded == nil {
			// We failed loading this
			// status, skip preparing.
			continue
		}

		// Prepare provided notif for frontend,
		// (note can return nil from late filtering).
		prepared, err := prepareAPI(meta.loaded)
		if err != nil {
			log.Errorf(ctx, "error preparing notification %s: %v", meta.loaded.ID, err)
			continue
		}

		if prepared != nil {
			// Append notification to return slice.
			apiNotifs = append(apiNotifs, prepared)
		}
	}

	return apiNotifs
}

// loadStatuses loads notifications using provided callback
// for the notifications in meta slice that aren't yet loaded.
// the amount very much depends on whether meta objects are
// yet-to-be-cached (i.e. newly loaded with notification),
// or are from the timeline cache (i.e. unloaded notification).
func loadNotifications(
	metas []*NotificationMeta,
	loadIDs func([]string) ([]*gtsmodel.Notification, error),
) error {
	// Determine which of our passed notif
	// meta objects still need notifs loading.
	toLoadIDs := make([]string, len(metas))
	loadedMap := make(map[string]*NotificationMeta, len(metas))
	for i, meta := range metas {
		if meta.loaded == nil {
			toLoadIDs[i] = meta.ID
			loadedMap[meta.ID] = meta
		}
	}

	// Load notifications with given IDs.
	loaded, err := loadIDs(toLoadIDs)
	if err != nil {
		return gtserror.Newf("error loading notifications: %w", err)
	}

	// Update returned NotificationMeta objects
	// with newly loaded notifications by IDs.
	for i := range loaded {
		notif := loaded[i]
		meta := loadedMap[notif.ID]
		meta.loaded = notif
	}

	return nil
}

// toNotificationMeta converts a slice of database model notifications
// into our cache wrapper type, a slice of []NotificationMeta{}.
func toNotificationMeta(in []*NotificationMeta, notifs []*gtsmodel.Notification) []*NotificationMeta {
	return xslices.Gather(in, notifs, func(n *gtsmodel.Notification) *NotificationMeta {
		return &NotificationMeta{
			ID:              n.ID,
			OriginAccountID: n.OriginAccountID,
			StatusOrEditID:  n.StatusOrEditID,
			loaded:          n,
		}
	})
}

// doFilterNotifications performs given filter function on provided notifications.
func doFilterNotifications(notifs []*gtsmodel.Notification, filter func(*gtsmodel.Notification) bool) []*gtsmodel.Notification {

	// Check for provided
	// filter function.
	if filter == nil {
		return notifs
	}

	// Filter the provided input notifications.
	return slices.DeleteFunc(notifs, filter)
}
