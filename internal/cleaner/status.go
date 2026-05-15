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

package cleaner

import (
	"context"
	"errors"
	"time"

	"code.superseriousbusiness.org/gopkg/log"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"codeberg.org/gruf/go-kv/v2"
	"codeberg.org/gruf/go-longdur"
)

// Status encompasses a set of
// status cleanup / admin utils.
type Status struct{ Cleaner }

// All will execute all cleaner.Status utilities synchronously, including output logging.
// NOTE: unlike other cleaner types, `gtscontext.DryRun()` is not checked or respected here.
func (s *Status) All(ctx context.Context, now time.Time, maxStubAge, maxRemoteAge longdur.Duration) {
	var dur time.Duration
	if _, dur = maxStubAge.Duration(); dur > 0 {
		const rateLimit = 500 * time.Millisecond
		s.LogPruneLeafStubs(ctx, now.Add(-dur), rateLimit)
	}
	if _, dur = maxRemoteAge.Duration(); dur > 0 {
		const rateLimit = 500 * time.Millisecond
		s.LogPruneOldRemote(ctx, now.Add(-dur), rateLimit)
	}
}

// LogPruneOldRemote performs PruneOldRemote(...), logging the start and outcome.
func (s *Status) LogPruneOldRemote(ctx context.Context, olderThan time.Time, rateLimit time.Duration) {
	log.Infof(ctx, "start older than: %s", olderThan.Format(stamp))
	if n, err := s.PruneOldRemote(ctx, olderThan, rateLimit); err != nil {
		log.Error(ctx, err)
	} else {
		log.Infof(ctx, "pruned: %d", n)
	}
}

// LogPruneLeafStubs performs PruneLeafStubs(...), logging the start and outcome.
func (s *Status) LogPruneLeafStubs(ctx context.Context, olderThan time.Time, rateLimit time.Duration) {
	log.Infof(ctx, "start older than: %s", olderThan.Format(stamp))
	if n, err := s.PruneLeafStubs(ctx, olderThan, rateLimit); err != nil {
		log.Error(ctx, err)
	} else {
		log.Infof(ctx, "pruned: %d", n)
	}
}

// PruneOldRemote will delete old status threads without any boosts, local favourites or local replies, older than given time.
// Rate limit is an optional (i.e. when > 0) parameter to limit DB load by sleeping between subsequent delete calls.
func (s *Status) PruneOldRemote(ctx context.Context, olderThan time.Time, rateLimit time.Duration) (int, error) {
	var total int
	page := new(paging.Page)

	// Setup page w/ select limit.
	page.Max = paging.MaxID("")
	page.Limit = selectLimit

	// Drop time by a minute to improve search,
	// (i.e. make it olderThan inclusive search).
	olderThan = olderThan.Add(+time.Minute)

	// Get binary ULID for 'olderThan' to use as maxID.
	olderThanID := id.ZeroBinaryULIDForTime(olderThan)
	page.Max.Value = olderThanID.String()

	for page != nil {
		if rateLimit > 0 {
			// Rate limiting was requested, this is very
			// heavy on the db and doesn't do anything but
			// loop on db queries, so give the db a break.
			time.Sleep(rateLimit)
		}

		// Delete given page of old remote status threads, returning deleted count.
		count, next, err := s.state.DB.DeleteOldRemoteStatuses(ctx, olderThanID, page)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return total, gtserror.Newf("error deleting statuses: %w", err)
		}

		log.DebugKVs(ctx, kv.Fields{
			{K: "count", V: count},
			{K: "page", V: page.Max.Value},
		}...)

		// Update count.
		total += count

		// Set next.
		page = next
	}

	return total, nil
}

// PruneLeafStubs will delete orphaned leaf status stubs older than given time, i.e. those marked as deleted and
// not helping to maintain status threading by being the middle connector between otherwise undeleted statuses.
// Rate limit is an optional (i.e. when > 0) parameter to limit DB load by sleeping between subsequent delete calls.
func (s *Status) PruneLeafStubs(ctx context.Context, olderThan time.Time, rateLimit time.Duration) (int, error) {
	var total int
	page := new(paging.Page)

	// Setup page w/ select limit.
	page.Max = paging.MaxID("")
	page.Limit = selectLimit

	// Drop time by a minute to improve search,
	// (i.e. make it olderThan inclusive search).
	olderThan = olderThan.Add(-time.Minute)

	// Get ULID for 'olderThan' to use as maxID.
	olderThanID := id.ZeroULIDForTime(olderThan)
	page.Max.Value = olderThanID

	for page != nil {
		if rateLimit > 0 {
			// Rate limiting was requested, this is very
			// heavy on the db and doesn't do anything but
			// loop on db queries, so give the db a break.
			time.Sleep(rateLimit)
		}

		// Delete given page of status leaf stubs, return delete count.
		count, next, err := s.state.DB.DeleteLeafStubStatuses(ctx, page)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return total, gtserror.Newf("error deleting statuses: %w", err)
		}

		log.DebugKVs(ctx, kv.Fields{
			{K: "count", V: count},
			{K: "page", V: page.Max.Value},
		}...)

		// Update count.
		total += count

		// Set next.
		page = next
	}

	return total, nil
}
