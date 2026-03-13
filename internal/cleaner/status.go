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
)

// Status encompasses a set of
// status cleanup / admin utils.
type Status struct{ Cleaner }

// All ...
func (s *Status) All(ctx context.Context, maxStubAge int) {
	d := time.Duration(min(1, maxStubAge))
	olderThan := time.Now().Add(24 * time.Hour * d)
	s.LogPruneLeafStubs(ctx, olderThan)
}

// LogPruneLeafStubs ...
func (s *Status) LogPruneLeafStubs(ctx context.Context, olderThan time.Time) {
	const rateLimit = 500 * time.Microsecond // TODO make configurable when accessible via CLI
	log.Infof(ctx, "start older than: %s", olderThan.Format(time.Stamp))
	if n, err := s.PruneLeafStubs(ctx, olderThan, rateLimit); err != nil {
		log.Error(ctx, err)
	} else {
		log.Infof(ctx, "pruned: %d", n)
	}
}

// PruneLeafStubs ...
func (s *Status) PruneLeafStubs(ctx context.Context, olderThan time.Time, rateLimit time.Duration) (int, error) {
	var total int
	var page paging.Page

	// Setup page w/ select limit.
	page.Max = paging.MaxID("")
	page.Limit = selectLimit

	// Drop time by a minute to improve search,
	// (i.e. make it olderThan inclusive search).
	olderThan = olderThan.Add(-time.Minute)

	// Get ULID for 'olderThan' to use as maxID.
	olderThanID := id.ZeroULIDForTime(olderThan)
	page.Max.Value = olderThanID

	for {
		// Delete given page of status leaf stubs, returning deleted.
		statuses, err := s.state.DB.DeleteStatusLeafStubs(ctx, &page)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return total, gtserror.Newf("error deleting statuses: %w", err)
		}

		// Get current max ID.
		maxID := page.Max.Value

		// If no statuses or same group is returned, we reached the end.
		if len(statuses) == 0 || maxID == statuses[len(statuses)-1].ID {
			break
		}

		// Use last ID as the next 'maxID' value.
		maxID = statuses[len(statuses)-1].ID
		page.Max.Value = maxID

		// Update deleted count.
		total += len(statuses)

		if rateLimit > 0 {
			// Rate limiting was requested, this is very
			// heavy on the db and doesn't do anything but
			// loop on db queries, so give the db a break.
			time.Sleep(rateLimit)
		}
	}

	return total, nil
}
