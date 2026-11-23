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

package netutil_test

import (
	"net/netip"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/netutil"
)

func TestSafeIP(t *testing.T) {
	tests := []struct {
		name string
		ip   netip.Addr
	}{
		// IPv4 tests
		{
			name: "IPv4 this host on this network",
			ip:   netip.MustParseAddr("0.0.0.0"),
		},
		{
			name: "IPv4 dummy address",
			ip:   netip.MustParseAddr("192.0.0.8"),
		},
		{
			name: "IPv4 Port Control Protocol Anycast",
			ip:   netip.MustParseAddr("192.0.0.9"),
		},
		{
			name: "IPv4 Traversal Using Relays around NAT Anycast",
			ip:   netip.MustParseAddr("192.0.0.10"),
		},
		{
			name: "IPv4 NAT64/DNS64 Discovery 1",
			ip:   netip.MustParseAddr("192.0.0.17"),
		},
		{
			name: "IPv4 NAT64/DNS64 Discovery 2",
			ip:   netip.MustParseAddr("192.0.0.171"),
		},
		// IPv6 tests
		{
			name: "IPv4-mapped address",
			ip:   netip.MustParseAddr("::ffff:169.254.169.254"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if safe := netutil.IsSafeIP(tc.ip); safe {
				t.Fatalf("Expected IP %s to not safe (%t), got: %t", tc.ip, false, safe)
			}
		})
	}
}
