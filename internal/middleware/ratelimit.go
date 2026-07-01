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
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"time"

	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gopkg/log"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
)

const rateLimitPeriod = 5 * time.Minute

// RateLimit returns a middleware that will automatically rate
// limit caller (by IP address), and enrich the response header with
// the following headers:
//
//   - `X-Ratelimit-Limit`     - max requests allowed per time period (fixed).
//   - `X-Ratelimit-Remaining` - requests remaining for this IP before reset.
//   - `X-Ratelimit-Reset`     - ISO8601 timestamp when the rate limit will reset.
//
// If `X-Ratelimit-Limit` is exceeded, the request is aborted and an
// HTTP 429 TooManyRequests status is returned.
//
// If the config AdvancedRateLimitRequests value is <= 0, then a noop
// handler will be returned, which performs no rate limiting.
func RateLimit(limit int, except []netip.Prefix) httputil.MiddlewareFunc {
	if limit <= 0 {
		// Rate limiting is disabled.
		// Return noop middleware.
		return nil
	}

	// Prepare limiter instance.
	limiter := limiter.New(
		memory.NewStore(),
		limiter.Rate{
			Period: rateLimitPeriod,
			Limit:  int64(limit),
		},
	)

	// It's prettymuch impossible to effectively
	// rate limit the immense IPv6 address space
	// unless we mask some of the bytes.
	//
	// This mask is pretty coarse, and puts IPv6
	// blocking on more or less the same footing
	// as IPv4 blocking in terms of how likely it
	// is to prevent abuse while still allowing
	// legit users access to the service.
	ipv6Mask := net.CIDRMask(64, 128)

	return func(h func(*httputil.Context)) func(*httputil.Context) {
		if h == nil {
			panic("nil func")
		}

		return func(c *httputil.Context) {
			// Use our heuristic for determining
			// clientIP, which accounts for reverse
			// proxies and trusted proxies setting.
			ip := gtscontext.ClientIP(c)

			// ClientIP
			// must be set.
			if ip == nil {
				log.Warn(c,
					"cannot do rate limiting for this request as client IP was empty;"+
						" your upstream reverse proxy may be misconfigured")

				// Pass on
				// to next.
				h(c)
				return
			}

			// Check if this IP is exempt from rate
			// limits and skip further checks if so.
			for _, prefix := range except {
				if prefix.Contains(*ip) {

					// Pass on
					// to next.
					h(c)
					return
				}
			}

			if ip.Is6() {
				// Convert to older "net"
				// package IP for masking.
				asIP := net.IP(ip.AsSlice())

				// Apply coarse IPv6 mask.
				asIP = asIP.Mask(ipv6Mask)

				// Convert back to netip.Addr type.
				addr, _ := netip.AddrFromSlice(asIP)
				ip = &addr
			}

			// Fetch rate limit info for (masked) clientIP.
			context, err := limiter.Get(c, ip.String())
			if err != nil {
				respondInternalServerError(c, err)
				return
			}

			// Provide reset in same format used by
			// Mastodon. There's no real standard as
			// to what format X-RateLimit-Reset SHOULD
			// use, but since most clients interacting
			// with us will expect the Mastodon version,
			// it makes sense to take this.
			resetT := time.Unix(context.Reset, 0)
			reset := util.FormatISO8601(resetT)

			c.W.Header().Set("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
			c.W.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
			c.W.Header().Set("X-RateLimit-Reset", reset)

			if context.Reached {
				// Return JSON error message for
				// consistency with other endpoints.
				httputil.Data(c,
					http.StatusTooManyRequests,
					apiutil.AppJSON,
					apiutil.ErrorRateLimited,
				)
				return
			}

			// Pass on
			// to next.
			h(c)
		}
	}
}
