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

package httpclient

import (
	"net/netip"
	"syscall"

	"code.superseriousbusiness.org/gotosocial/internal/netutil"
)

type Sanitizer struct {
	Allow []netip.Prefix
	Block []netip.Prefix
}

// Sanitize implements the required net.Dialer.Control function signature.
func (s *Sanitizer) Sanitize(ntwrk, addr string, _ syscall.RawConn) error {

	// Check network.
	switch ntwrk {
	case "tcp4", "tcp6":
	default:
		return ErrInvalidNetwork
	}

	// Parse IP and port from address string.
	ipport, err := netip.ParseAddrPort(addr)
	if err != nil {
		return err
	}

	// Separate the IP.
	ip := ipport.Addr()

	// Check if IP is explicitly allowed.
	for i := 0; i < len(s.Allow); i++ {
		if s.Allow[i].Contains(ip) {
			return nil
		}
	}

	// Check if IP is explicitly blocked.
	for i := 0; i < len(s.Block); i++ {
		if s.Block[i].Contains(ip) {
			return ErrReservedAddr
		}
	}

	// Validate as a safe IP.
	if !netutil.IsSafeIP(ip) {
		return ErrReservedAddr
	}

	return nil
}
