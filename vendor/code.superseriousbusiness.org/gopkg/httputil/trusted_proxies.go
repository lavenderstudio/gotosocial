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

package httputil

import (
	"net/http"
	"net/netip"
	"strings"
)

// ProxyConfiguration combines a set of IP prefixes that can
// be trusted to be provide remote IP information, and a set of
// header names from which remote IP information can be determined.
// Together, these can be used to determine upstream remote IP of
// an incoming request for a given reverse proxy configuration.
type ProxyConfiguration struct {
	RemoteIPHeaders []string
	TrustedPrefixes []netip.Prefix
}

// DefaultProxyConfiguration returns a valid default
// ProxyConfiguration{} with X-Forwarded-For remote IP header set.
func DefaultProxyConfiguration() ProxyConfiguration {
	return ProxyConfiguration{RemoteIPHeaders: []string{"X-Forwarded-For"}}
}

// GetRemoteIP parses the http.Request{}.RemoteAddr value
// as netip.AddrPort{}, returning the netip.Addr{} portion.
func GetRemoteIP(r *http.Request) netip.Addr {
	addrport, _ := netip.ParseAddrPort(r.RemoteAddr)
	return addrport.Addr()
}

// GetClientIP attempts to fetch a valid client IP for the given request given
// the receiving ProxyConfiguration{}. Else, falls back to http.Request{}.RemoteAddr.
func (cfg *ProxyConfiguration) GetClientIP(r *http.Request) netip.Addr {
	ip := GetRemoteIP(r)
	if cfg == nil {
		return ip
	}
	if cfg.IsTrustedProxy(ip) {
		for _, name := range cfg.RemoteIPHeaders {
			for _, v := range r.Header.Values(name) {
				ip := getTrustedRemoteIP(cfg.TrustedPrefixes, v)
				if ip.IsValid() {
					return ip
				}
			}
		}
	}
	return ip
}

// GetTrustedRemoteIP attempts to find the first trusted remote IP from a given header value,
// (i.e. comma-separated list of IPs as found in an X-Real-IP or X-Forwarded-For header),
// working backwards through list and finding first outside of configured trusted prefixes.
func (cfg *ProxyConfiguration) GetTrustedRemoteIP(hdrvalue string) netip.Addr {
	return getTrustedRemoteIP(cfg.TrustedPrefixes, hdrvalue)
}

// IsTrustedProxy returns whether given netip.Addr{} is a trusted
// proxy according to receiving trusted IP prefixes configuration.
func (cfg *ProxyConfiguration) IsTrustedProxy(addr netip.Addr) bool {
	return doPrefixesContain(cfg.TrustedPrefixes, addr)
}

func getTrustedRemoteIP(prefixes []netip.Prefix, value string) netip.Addr {
	for len(value) > 0 {
		var each string

		// Split string by comma delimiters.
		i := strings.LastIndexByte(value, ',')
		if i >= 0 {

			// Reslice up to delim.
			each = value[i+1:]
			value = value[:i]
		} else {

			// Each is the
			// entire string.
			each = value
			value = ""
		}

		// Trim extra spaces.
		each = trimSpace(each)

		// Attempt to parse IP from string.
		ip, err := netip.ParseAddr(each)
		if err != nil {
			break
		}

		// Check if IP is trusted proxy,
		// if so keep looking for next IP.
		if doPrefixesContain(prefixes, ip) {
			continue
		}

		return ip
	}

	// i.e. none found.
	return netip.Addr{}
}

// doPrefixesContain returns whether addr is contained within prefixes.
func doPrefixesContain(prefixes []netip.Prefix, addr netip.Addr) bool {
	for _, prefix := range prefixes {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

// trimSpace trims leading and trailing ' ' from s.
func trimSpace(s string) string {
	for len(s) > 0 && s[0] == ' ' {
		s = s[1:]
	}
	for len(s) > 0 && s[len(s)-1] == ' ' {
		s = s[:len(s)-1]
	}
	return s
}
