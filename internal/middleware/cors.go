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
	"net/url"
	"time"

	"code.superseriousbusiness.org/gopkg/httputil"
	"github.com/rs/cors"
)

// CORS returns a new gin middleware which allows CORS requests to be processed.
// This is necessary in order for web/browser-based clients like Semaphore to work.
func CORS() httputil.FlatMiddlewareFunc {
	cors := cors.New(cors.Options{
		// Allow all origins with expected schema.
		AllowOriginFunc: func(origin string) bool {
			u, err := url.Parse(origin)
			if err != nil {
				return false
			}
			switch u.Scheme {
			case "http", "https":
			case "chrome-extension":
			case "safari-extension":
			case "moz-extension":
			case "ms-browser-extension":
			case "ws", "wss":
			default:
				return false
			}
			return true
		},

		AllowedMethods: []string{
			"HEAD",
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},

		AllowedHeaders: []string{
			// basic cors stuff
			"Content-Length",
			"Content-Type",
			"Origin",

			// needed to pass
			// oauth bearer tokens
			"Authorization",

			// Some clients require this; see:
			//   - https://docs.joinmastodon.org/methods/statuses/#headers
			//   - https://codeberg.org/superseriousbusiness/gotosocial/issues/1664
			"Idempotency-Key",

			// needed for websocket upgrade requests
			"Upgrade",
			"Sec-WebSocket-Extensions",
			"Sec-WebSocket-Key",
			"Sec-WebSocket-Protocol",
			"Sec-WebSocket-Version",
			"Connection",
		},

		ExposedHeaders: []string{
			// needed for accessing next/prev links
			// when making GET timeline requests
			"Link",

			// needed so clients can handle rate limits
			"X-RateLimit-Reset",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-Request-Id",

			// websocket stuff
			"Connection",
			"Sec-WebSocket-Accept",
			"Upgrade",
		},

		MaxAge: int((2 * time.Minute) / time.Second),
	})

	if cors == nil {
		panic("nil cors")
	}

	return func(c *httputil.Context) {
		cors.HandlerFunc(&c.W, c.R)
	}
}
