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

import "sync/atomic"

// lazyset implements a set from a string-keyed
// map, only allocating itself on first addition.
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

// atomic_map wraps an atomic pointer to a map to provide a utility
// load-followed-by-compare-and-swap function to support a copy on
// write model of concurrency for a map experiencing heavy read usage.
//
// nolint:revive
type atomic_map[K comparable, V any] struct{ atomic.Pointer[map[K]V] }

func (ptr *atomic_map[K, V]) loadAndCAS(fn func(old map[K]V) (new map[K]V, changed bool)) {
	if fn == nil {
		panic("nil func")
	}
	for {
		// Load cur ptr.
		cur := ptr.Load()

		// Get map
		// to work on.
		var m map[K]V
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
		if !ptr.CompareAndSwap(cur, &m) {

			// We failed the
			// CAS, reloop.
			continue
		}
	}
}
