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
	"sync/atomic"

	"code.superseriousbusiness.org/gopkg/httputil"
)

// WithThrottling returns a middleware func that performs throttling of incoming requests,
// ensuring that only a certain number of requests are handled concurrently, to reduce
// congestion of the server.
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
// RetryAfter determines the Retry-After header value to be sent to throttled requests.
//
// Useful links:
//
//   - https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Retry-After
//   - https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/503
func WithThrottling(limit, queueLimit int, errh httputil.HandlerFunc) httputil.MiddlewareFunc {
	if limit <= 0 || queueLimit <= 0 {
		// throttling is disabled,
		// return a noop middleware
		return nil
	}

	if errh == nil {
		panic("nil error func")
	}

	// token represents request
	// that is being processed.
	type token struct{}

	// request queue token channel.
	tokens := make(chan token, limit)

	// current requester count.
	var inflight atomic.Int64

	// prefill the token channel.
	for i := 0; i < limit; i++ {
		tokens <- token{}
	}

	return func(h httputil.HandlerFunc) httputil.HandlerFunc {
		if h == nil {
			panic("nil func")
		}

		return func(c *httputil.Context) {
			// Always decrement
			// request counter.
			defer inflight.Add(-1)

			// Increment count.
			n := inflight.Add(1)

			// Check whether the request
			// count is over queue limit.
			if n > int64(queueLimit) {
				errh(c)
				return
			}

			// Sit and wait in the
			// queue for free token.
			select {

			case <-c.R.Context().Done():
				// request context has
				// been canceled already.
				return

			case tok := <-tokens:
				// caller has successfully
				// received a token, allowing
				// request to be processed.

				defer func() {
					// when we're finished, return
					// this token to the bucket.
					tokens <- tok
				}()

				// Pass on
				// to next.
				h(c)
			}
		}
	}
}
