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
	"sync/atomic"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// StatusTimelines is a concurrency safe map of StatusTimeline{}
// objects, optimizing *very heavily* for reads over writes.
type StatusTimelines struct {

	// atomic cache map pointer, RO outside CAS
	ptr atomic.Pointer[map[string]*_StatusTimeline]

	// ...
	timeout time.Duration

	// new StatusTimeline{}
	// init arguments.
	cap int
}

// a simple wrapper around StatusTimeline
// to add a last-use-time tracking value.
type _StatusTimeline struct {
	StatusTimeline
	last atomic.Pointer[time.Time]
}

// Init stores the given argument(s) such that any created StatusTimeline{}
// objects by MustGet() will initialize them with the given arguments.
func (t *StatusTimelines) Init(cap int, timeout time.Duration) {
	t.timeout = timeout
	t.cap = cap
}

// MustGet will attempt to fetch StatusTimeline{} stored under key, else creating one.
func (t *StatusTimelines) MustGet(key string) *StatusTimeline {
	var tt *_StatusTimeline

	// Perform load and (potential) store operation within main loadAndCAS() function loop.
	t.loadAndCAS(func(m map[string]*_StatusTimeline) (map[string]*_StatusTimeline, bool) {

		// Look for an existing
		// timeline object in cache.
		if tt = m[key]; tt != nil {

			// i.e. no change.
			return nil, false
		}

		// Get map clone
		// before changes.
		m = clone(m)

		// Allocate new timeline.
		tt = new(_StatusTimeline)
		tt.Init(t.cap)

		// Store timeline
		// in new map.
		m[key] = tt

		// i.e. changed
		return m, true
	})

	if t.timeout > 0 {
		// Update timeline
		// last use time.
		now := time.Now()
		tt.last.Store(&now)
	}

	// Return embedded timeline.
	return &tt.StatusTimeline
}

// InsertOne attempts to call StatusTimeline{}.InsertOne() on timeline under key, only if it exists.
func (t *StatusTimelines) InsertOne(key string, status *gtsmodel.Status) bool {
	if p := t.ptr.Load(); p != nil {
		if tt := (*p)[key]; tt != nil {
			return tt.InsertOne(status)
		}
	}
	return false
}

// Delete will delete the stored StatusTimeline{} under key, if any.
func (t *StatusTimelines) Delete(key string) {
	t.loadAndCAS(func(m map[string]*_StatusTimeline) (map[string]*_StatusTimeline, bool) {
		if m[key] == nil {

			// i.e. no change.
			return nil, false
		}

		// Get map clone
		// before changes.
		m = clone(m)

		// Delete ID.
		delete(m, key)

		// i.e. changed
		return m, true
	})
}

// RemoveByStatusIDs calls RemoveByStatusIDs() for each of the stored StatusTimeline{}s.
func (t *StatusTimelines) RemoveByStatusIDs(statusIDs ...string) {
	if p := t.ptr.Load(); p != nil {
		for _, tt := range *p {
			tt.RemoveByStatusIDs(statusIDs...)
		}
	}
}

// RemoveByAccountIDs calls RemoveByAccountIDs() for each of the stored StatusTimeline{}s.
func (t *StatusTimelines) RemoveByAccountIDs(accountIDs ...string) {
	if p := t.ptr.Load(); p != nil {
		for _, tt := range *p {
			tt.RemoveByAccountIDs(accountIDs...)
		}
	}
}

// Trim calls Trim() for each of the stored StatusTimeline{}s,
// clearing and / or dropping timelines beyond timeout time.
func (t *StatusTimelines) Trim() {
	if t.timeout <= 0 {
		// No timeout is set, perform
		// a simple trim of timelines.
		if p := t.ptr.Load(); p != nil {
			for _, tt := range *p {
				tt.Trim()
			}
		}
		return
	}

	// Perform a more complex
	// timeout based trimming.
	t.trim()
}

func (t *StatusTimelines) trim() {
	// A longer duration than timeout
	// after which we mark an unused
	// timeline as stale and *delete*
	// from the timelines cache map.
	var staleout time.Duration

	// Clamp staleout check time to a minimum 1 hour.
	if staleout = 10 * t.timeout; staleout < time.Hour {
		staleout = time.Hour
	}

	// Load current
	// cache map ptr.
	p := t.ptr.Load()
	if p == nil {
		return
	}

	var stale lazyset

	// Get current time.
	now := time.Now()

	// Range all timelines.
	for key, tt := range *p {

		// Load last use time.
		last := *tt.last.Load()

		// Determine how much
		// time has passed since
		// timeline last used.
		diff := now.Sub(last)

		switch {
		case diff >= staleout:
			// If timeline hasn't been used since
			// 'staleout' threshold, it's time to
			// delete it from the map. Due to our
			// heavy optimization for reads, this
			// may be relatively expensive, hence
			// why 'staleout' is clamped to a min.
			stale.Add(key)

		case diff >= t.timeout:
			// If timeline hasn't been used since
			// 'timeout', simply drop the entire
			// thing from memory. There's no need
			// to delete it as the entire structure
			// is fairly small in-memory and saves
			// us needing to rewrite the RO map.
			tt.Clear()

		default:
			// Else, simply
			// trim to 'cut'.
			tt.Trim()
		}
	}

	// If no stale keys found,
	// no need to continue.
	if len(stale) == 0 {
		return
	}

	// Within the main load / CAS loop, clone current map and drop all stale keys from it.
	t.loadAndCAS(func(m map[string]*_StatusTimeline) (map[string]*_StatusTimeline, bool) {
		clone := make(map[string]*_StatusTimeline, len(m)-len(stale))
		for key, tt := range m {

			// Check if marked as stale.
			if _, ok := stale[key]; ok {

				// Weed-out race conditions by performing
				// a final staleness check on last-use time.
				if now.Sub(*tt.last.Load()) >= staleout {

					// Timeline definitely
					// stale, skip adding.
					continue
				}
			}

			// Add to clone.
			clone[key] = tt
		}

		// Return map clone, and
		// determine if it changed.
		changed := len(clone) != len(m)
		return clone, changed
	})
}

// Clear attempts to call Clear() for StatusTimeline{} under key.
func (t *StatusTimelines) Clear(key string) {
	if p := t.ptr.Load(); p != nil {
		if tt := (*p)[key]; tt != nil {
			tt.Clear()
		}
	}
}

// ClearAll calls Clear() for each of the stored StatusTimeline{}s.
func (t *StatusTimelines) ClearAll() {
	if p := t.ptr.Load(); p != nil {
		for _, tt := range *p {
			tt.Clear()
		}
	}
}

func (t *StatusTimelines) loadAndCAS(fn func(current map[string]*_StatusTimeline) (new map[string]*_StatusTimeline, changed bool)) {
	if fn == nil {
		panic("nil func")
	}
	for {
		// Load current ptr.
		cur := t.ptr.Load()

		// Get timeline map to work on.
		var m map[string]*_StatusTimeline
		if cur != nil {
			m = (*cur)
		}

		// Pass to fn.
		m, ok := fn(m)
		if !ok {

			// Nothing
			// changed.
			return
		}

		// Attempt to update the map ptr.
		if !t.ptr.CompareAndSwap(cur, &m) {

			// We failed the
			// CAS, reloop.
			continue
		}
	}
}

type lazyset map[string]struct{}

func (s *lazyset) Add(key string) {
	if *s == nil {
		(*s) = make(lazyset)
	}
	(*s)[key] = struct{}{}
}

// clone is functionally similar to maps.Clone(),
// except a nil input will return initialized output.
func clone[T any](m map[string]T) map[string]T {
	m2 := make(map[string]T, len(m))
	for key, val := range m {
		m2[key] = val
	}
	return m2
}
