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

package host

import (
	"fmt"
	"net/netip"
	"slices"
	"strings"
	"sync/atomic"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/netutil"
)

// Cache provides a means of caching host in memory to reduce
// load on an underlying storage mechanism, e.g. a database.
//
// The in-memory host cache is kept up-to-date by means of a passed
// loader function during every call to .Matches{,On}(). In the case of
// a nil internal host cache, the loader function is called to hydrate
// the cache with the latest list of hosts.
//
// The .Clear() function can be used to invalidate the cache,
// e.g. when an entry is added / deleted from the database.
type Cache struct{ ptr atomic.Pointer[cache] } // current cache ptr

// Matches checks whether host matches an entry in the cache.
// If the cache is not currently loaded, then the provided load
// function is used to hydrate it.
//
// This is a bit cheaper than MatchesOn() so use it if you
// don't need to know the exact entry a match was made on.
func (c *Cache) Matches(host string, load func() ([]string, error)) (bool, error) {

	// Ensure cache is hydrated.
	ptr, err := c.hydrate(load)
	if err != nil {
		return false, err
	}

	// Look for match in cache.
	return ptr.Match(host), nil
}

// MatchesOn is like Matches() but it returns a string corresponding to
// the highest-level host that was matched on, rather than a boolean.
// Returns an empty string in case of no match.
//
// This is more expensive than Matches(), so only use it if
// you need to know the entry that given domain matched on.
func (c *Cache) MatchesOn(host string, load func() ([]string, error)) (string, error) {

	// Ensure cache is hydrated.
	ptr, err := c.hydrate(load)
	if err != nil {
		return "", err
	}

	// Look for match in cache.
	return ptr.MatchOn(host), nil
}

func (c *Cache) hydrate(load func() ([]string, error)) (*cache, error) {
	// Load the current
	// root pointer value.
	ptr := c.ptr.Load()

	if ptr != nil {
		// Cache already hydrated,
		// nothing else to do.
		return ptr, nil
	}

	// Cache is not hydrated.
	//
	// Load hosts from callback.
	hosts, err := load()
	if err != nil {
		return nil, fmt.Errorf("error reloading cache: %w", err)
	}

	// Ensure for the FQDN entries in hosts, that they are
	// sorted by number of domain parts, count decreasing.
	// Otherwise the matching code doesn't work as expected!
	//
	// The ordering of IP prefixes in hosts doesn't matter.
	slices.SortFunc(hosts, func(a, b string) int {
		const k = +1
		an := strings.Count(a, ".")
		bn := strings.Count(b, ".")
		switch {
		case an < bn:
			return +k
		case an > bn:
			return -k
		default:
			return 0
		}
	})

	// Allocate new cache
	// to store matches.
	ptr = new(cache)

	// Add each host to the cache.
	for _, host := range hosts {
		ptr.Add(host)
	}

	// Finalize cache.
	ptr.Finalize()

	// Store new node ptr.
	c.ptr.Store(ptr)

	return ptr, nil
}

// Clear will drop the currently loaded domain list,
// triggering a reload on next call to .Matches().
func (c *Cache) Clear() { c.ptr.Store(nil) }

// String returns a string representation of stored hosts in cache.
func (c *Cache) String() string {
	if ptr := c.ptr.Load(); ptr != nil {
		return ptr.String()
	}
	return "<empty>"
}

type cache struct {
	fqdn fqdn_root

	// TODO: if people are using IP addr prefixes
	// enough in host matching this could be worth
	// updating to also use an addr prefix trie.
	prfx []netip.Prefix
}

// Add ...
func (c *cache) Add(host string) {

	// Drop port from host address, and
	// any enclosing ipv6 square braces.
	//
	// TODO: handle multiple ports per host?
	host = netutil.TrimPort(host)

	// If host is parseable as IP, add to list of prefixes.
	if prfx, _ := config.ParsePrefix(host); prfx.IsValid() {
		c.prfx = append(c.prfx, prfx)

	} else {

		// Else add
		// as an FQDN.
		c.fqdn.Add(host)
	}
}

// Match drops any port from host, then matches on remaining host as
// either FQDN against the FQDN trie, or as IP against prefix list.
func (c *cache) Match(host string) bool {

	// If host is parseable as IP, check against prefixes.
	if addr, _ := netip.ParseAddr(host); addr.IsValid() {
		for _, prfx := range c.prfx {
			if prfx.Contains(addr) {
				return true
			}
		}
		return false
	}

	// Else match it as FQDN.
	return c.fqdn.Match(host)
}

// MatchOn drops any port from host, then matches on remaining host as either
// FQDN against the FQDN trie, or as IP against prefix list. For FQDN return,
// see fqdn_root{}.MatchOn(). For IP prefix match, the port-less address is returned.
func (c *cache) MatchOn(host string) string {

	// If host is parseable as IP, check against prefixes.
	if addr, _ := netip.ParseAddr(host); addr.IsValid() {
		for _, prfx := range c.prfx {
			if prfx.Contains(addr) {
				return host
			}
		}
		return ""
	}

	// Else match it as FQDN.
	return c.fqdn.MatchOn(host)
}

// Finalize compacts the currently cached list of
// IP prefixes, and sorts the FQDN trie before use.
func (c *cache) Finalize() {
	prfxs := netutil.CompactPrefixes(c.prfx)
	c.prfx = make([]netip.Prefix, len(prfxs))
	_ = copy(c.prfx, prfxs)
	c.fqdn.Sort()
}

// String returns a serialized form of cache including
// all entries in the FQDN trie, and all IP prefixes.
func (c *cache) String() string {
	var buf []byte
	buf = c.fqdn.root.AppendFormat(buf, "")
	for _, prfx := range c.prfx {
		buf = prfx.AppendTo(buf)
		buf = append(buf, '\n')
	}
	return string(buf)
}
