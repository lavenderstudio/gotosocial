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
	"net/http"
	"runtime"
	"strconv"
	"time"

	"code.superseriousbusiness.org/gopkg/httputil"
	extmiddleware "code.superseriousbusiness.org/gopkg/httputil/middleware"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
)

// Throttle returns a middleware that performs throttling of incoming requests,
// ensuring that only a certain number of requests are handled concurrently, to reduce
// congestion of the server.
//
// Limits are configured using available CPUs and the given cpuMultiplier value.
// Open request limit is available CPUs * multiplier; backlog limit is limit * multiplier.
//
// Example values for multiplier 8:
//
//	1 cpu = 08 open, 064 backlog
//	2 cpu = 16 open, 128 backlog
//	4 cpu = 32 open, 256 backlog
//
// Example values for multiplier 4:
//
//	1 cpu = 04 open, 016 backlog
//	2 cpu = 08 open, 032 backlog
//	4 cpu = 16 open, 064 backlog
//
// Callers will first attempt to get a backlog token. Once they have that, they will
// wait in the backlog queue until they can get a token to allow their request to be
// processed.
//
// If the backlog queue is full, the request context is closed, or the caller has been
// waiting in the backlog for too long, this function will abort the request chain,
// write a JSON error into the response, set an appropriate Retry-After value, and set
// the HTTP response code to 503: Service Unavailable.
//
// If the multiplier is <= 0, a noop middleware will be returned instead.
//
// RetryAfter determines the Retry-After header value to be sent to throttled requests.
//
// Useful links:
//
//   - https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Retry-After
//   - https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/503
func Throttle(cpuMultiplier int, retryAfter time.Duration) httputil.MiddlewareFunc {
	limit := runtime.GOMAXPROCS(0) * cpuMultiplier
	queueLimit := limit * cpuMultiplier
	retryAfter = max(retryAfter, 5*time.Second)                             // clamp to a minimum
	retryAfterStr := strconv.FormatUint(uint64(retryAfter/time.Second), 10) // #nosec G115 -- Checked right above
	return extmiddleware.WithThrottling(limit, queueLimit, func(c *httputil.Context) {
		c.W.Header().Set("Retry-After", retryAfterStr)
		httputil.Data(c,
			http.StatusTooManyRequests,
			apiutil.AppJSON,
			apiutil.ErrorCapacityExceeded,
		)
	})
}
