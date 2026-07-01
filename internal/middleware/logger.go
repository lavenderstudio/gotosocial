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
	"fmt"
	"net/http"
	"strings"
	"time"

	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gopkg/httputil/middleware"
	"code.superseriousbusiness.org/gopkg/log"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"codeberg.org/gruf/go-bytesize"
	"codeberg.org/gruf/go-kv/v2"
)

// Logger returns a gin middleware which provides request logging and panic recovery.
func Logger(logClientIP bool) httputil.MiddlewareFunc {
	return func(h func(*httputil.Context)) func(*httputil.Context) {
		if h == nil {
			panic("nil func")
		}

		return func(c *httputil.Context) {
			// Determine time
			// before pass-off.
			before := time.Now()

			// defer so that we log *after
			// the request has completed*.
			defer func() {

				// Recover from any panics
				// and dump stack to stderr.
				if r := util.Recover(); r != nil {

					if c.W.StatusCode == 0 {
						// No response written, send generic Internal Error.
						c.W.WriteHeader(http.StatusInternalServerError)
					}

					// Append panic information to the request.
					err := fmt.Errorf("recovered panic: %v", r)
					c.Error(err)
				}

				// Initialize the logging fields.
				fields := make(kv.Fields, 5, 8)
				if len(fields) < 5 {
					panic(gtserror.New("bound check elimination"))
				}

				// Set request logging fields.
				fields[0] = kv.Field{"latency", time.Since(before)}
				fields[1] = kv.Field{"userAgent", c.R.UserAgent()}
				fields[2] = kv.Field{"method", c.R.Method}
				fields[3] = kv.Field{"statusCode", c.W.StatusCode}

				// If the request contains sensitive query
				// data only log path, else log entire URI.
				if sensitiveQuery(c.R.URL.RawQuery) {
					path := c.R.URL.Path
					fields[4] = kv.Field{"uri", path}
				} else {
					uri := c.R.RequestURI
					fields[4] = kv.Field{"uri", uri}
				}

				if logClientIP {
					if ip := middleware.GetClientIP(c); ip != nil {
						// Append IP only if configured to.
						fields = append(fields, kv.Field{
							"clientIP", ip.String(),
						})
					}
				}

				if errs := c.Errors(); len(errs) > 0 {
					// Append any extra log fields
					// attached to request errors.
					for _, err := range errs {
						extra := gtserror.LogFields(err)
						fields = append(fields, extra...)
					}

					// Always attach any found errors.
					fields = append(fields, kv.Field{
						"errors", errs,
					})
				}

				// Create entry
				// with fields.
				l := log.New().
					WithContext(c).
					WithFields(fields...)

				// Default level.
				lvl := log.INFO

				if c.W.StatusCode >= 500 {
					// Actual error.
					lvl = log.ERROR
				}

				// Get appropriate text for this status code.
				statusText := http.StatusText(c.W.StatusCode)
				if statusText == "" {

					// Look for other codes.
					switch c.W.StatusCode {
					case gtserror.StatusClientClosedRequest:
						statusText = gtserror.StatusTextClientClosedRequest
					case httputil.StatusHijacked:
						statusText = "Switching Protocols"
					default:
						statusText = "Unknown Status"
					}
				}

				// Generate nice looking bytecount.
				size := bytesize.Size(c.W.Written) // #nosec G115 -- Just logging

				// Write log entry with status text + body size.
				l.Logf(lvl, "%s: wrote %s", statusText, size)
			}()

			// Pass to
			// next h.
			h(c)
		}
	}
}

// sensitiveQuery checks whether given query string
// contains sensitive data that shouldn't be logged.
func sensitiveQuery(query string) bool {
	return strings.Contains(query, "token")
}
