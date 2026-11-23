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

package netutil

import (
	"net/netip"
	"slices"
)

var (
	// ipv6GlobalUnicast is the prefix set aside by IANA for global unicast assignments, i.e "the internet".
	// https://www.iana.org/assignments/ipv6-unicast-address-assignments/ipv6-unicast-address-assignments.xhtml
	ipv6GlobalUnicast = netip.MustParsePrefix("2000::/3")

	// ipv6Reserved contains IPv6 reserved IP prefixes that fall within ipv6GlobalUnicast.
	// https://www.iana.org/assignments/iana-ipv6-special-registry/iana-ipv6-special-registry.xhtml
	ipv6Reserved = [...]netip.Prefix{
		netip.MustParsePrefix("2001::/23"),         // IETF Protocol Assignments (RFC 2928)
		netip.MustParsePrefix("2001:db8::/32"),     // Documentation (RFC 3849)
		netip.MustParsePrefix("2002::/16"),         // 6to4 (RFC 3056)
		netip.MustParsePrefix("2620:4f:8000::/48"), // Direct Delegation AS112 Service (RFC 7534)
	}

	// ipv4Reserved contains IPv4 reserved IP prefixes.
	// https://www.iana.org/assignments/iana-ipv4-special-registry/iana-ipv4-special-registry.xhtml
	ipv4Reserved = [...]netip.Prefix{
		netip.MustParsePrefix("0.0.0.0/8"),       // Current network
		netip.MustParsePrefix("10.0.0.0/8"),      // Private
		netip.MustParsePrefix("100.64.0.0/10"),   // RFC6598
		netip.MustParsePrefix("127.0.0.0/8"),     // Loopback
		netip.MustParsePrefix("169.254.0.0/16"),  // Link-local
		netip.MustParsePrefix("172.16.0.0/12"),   // Private
		netip.MustParsePrefix("192.0.0.0/24"),    // RFC6890
		netip.MustParsePrefix("192.0.2.0/24"),    // Test, doc, examples
		netip.MustParsePrefix("192.31.196.0/24"), // AS112-v4, RFC 7535
		netip.MustParsePrefix("192.52.193.0/24"), // AMT, RFC 7450
		netip.MustParsePrefix("192.88.99.0/24"),  // IPv6 to IPv4 relay
		netip.MustParsePrefix("192.168.0.0/16"),  // Private
		netip.MustParsePrefix("192.175.48.0/24"), // Direct Delegation AS112 Service, RFC 7534
		netip.MustParsePrefix("198.18.0.0/15"),   // Benchmarking tests
		netip.MustParsePrefix("198.51.100.0/24"), // Test, doc, examples
		netip.MustParsePrefix("203.0.113.0/24"),  // Test, doc, examples
		netip.MustParsePrefix("224.0.0.0/4"),     // Multicast
		netip.MustParsePrefix("240.0.0.0/4"),     // Reserved (includes broadcast / 255.255.255.255)
	}
)

// IsSafeIP returns whether ip is an IPv4/6
// address in a non-reserved, public range.
func IsSafeIP(ip netip.Addr) bool {
	switch {

	// IPv4: check if IP
	// in IPv4 reserved.
	case ip.Is4():
		for _, reserved := range ipv4Reserved {
			if reserved.Contains(ip) {
				return false
			}
		}
		return true

	// IPv6: check if IP
	// in IPv6 reserved.
	case ip.Is6():
		if !ipv6GlobalUnicast.Contains(ip) {
			// Address is not globally routable,
			// ie., not "on the internet".
			return false
		}

		for _, reserved := range ipv6Reserved {
			if reserved.Contains(ip) {
				// Address is globally routable
				// but falls in a reserved range.
				return false
			}
		}

		return true

	// Assume malicious
	// by default.
	default:
		return false
	}
}

// CompactPrefixes compacts the given slice of prefixes by first sorting
// them in length order, ascending, then removing overlapping prefix entries.
func CompactPrefixes(prfxs []netip.Prefix) []netip.Prefix {

	// Trim any invalid prefixes.
	prfxs = slices.DeleteFunc(prfxs,
		func(p netip.Prefix) bool {
			return !p.IsValid()
		})

	// First, sort the prefixes in ascending order of length, where
	// a shorter length, broader matching prefix, comes first.
	slices.SortFunc(prfxs, func(p1, p2 netip.Prefix) int {
		b1, b2 := p1.Bits(), p2.Bits()
		switch {
		case b1 < b2:
			return -1
		case b1 > b2:
			return +1
		default:
			return 0
		}
	})

	// Then delete all larger prefixes already
	// covered by shorter ones earlier in the slice.
	for i := 0; i < len(prfxs); i++ {
		for j := i + 1; j < len(prfxs); {
			if !prfxs[i].Contains(prfxs[j].Addr()) {
				j++
				continue
			}
			_ = copy(prfxs[i+1:], prfxs[j+1:])
			prfxs = prfxs[:len(prfxs)-1]
		}
	}

	return prfxs
}
