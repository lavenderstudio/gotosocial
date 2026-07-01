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

package middleware

import (
	"context"
	"net/netip"

	"code.superseriousbusiness.org/gopkg/httputil"
)

type clientipkey struct{}

// GetClientIP searches given context for netip.Addr{} stored by WithClientIP().
func GetClientIP(ctx context.Context) *netip.Addr {
	addr, _ := ctx.Value(clientipkey{}).(*netip.Addr)
	return addr
}

// WithClientIP adds a middleware that stores a netip.Addr{} client
// IP within request context based on the given proxy configuration.
// A nil proxy configuration is acceptable, in which case remote addr is used.
func WithClientIP(cfg *httputil.ProxyConfiguration) httputil.FlatMiddlewareFunc {
	if cfg == nil {
		return func(c *httputil.Context) {
			addrport, _ := netip.ParseAddrPort(c.R.RemoteAddr)
			addr := addrport.Addr()
			c.V.Set(clientipkey{}, &addr)
		}
	}
	return func(c *httputil.Context) {
		ip := cfg.GetClientIP(c.R)
		c.V.Set(clientipkey{}, &ip)
	}
}
