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
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gopkg/log"
	"codeberg.org/gruf/go-bytesize"
	"codeberg.org/gruf/go-kv/v2"
)

// Logger returns a middleware which
// provides request logging and panic recovery.
func Logger() httputil.MiddlewareFunc {
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
				if r := recover(); r != nil {

					// Gather calling funcs and
					// log panic to stderr output.
					pcs := make([]uintptr, 10)
					n := runtime.Callers(3, pcs)
					i := runtime.CallersFrames(pcs[:n])
					_, _ = fmt.Fprintf(os.Stderr,
						"recovered panic: %v\n\n%s\n",
						r, gatherFrames(i, n).String())

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
					panic("bound check elimination")
				}

				// Set request logging fields.
				fields[0] = kv.Field{"latency", time.Since(before)}
				fields[1] = kv.Field{"userAgent", c.R.UserAgent()}
				fields[2] = kv.Field{"method", c.R.Method}
				fields[3] = kv.Field{"statusCode", c.W.StatusCode}
				fields[4] = kv.Field{"uri", c.R.URL.Path}

				if errs := c.Errors(); len(errs) > 0 {
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

type callers []runtime.Frame

// String will return a simple string
// representation of receiving Callers slice.
func (c callers) String() string {

	// Guess-timate to reduce allocs.
	buf := make([]byte, 0, 64*len(c))
	for i := 0; i < len(c); i++ {
		frame := c[i]

		// Append formatted caller info.
		fn := funcName(frame.Func)
		buf = append(buf, fn+"()\n\t"+frame.File+":"...)
		buf = strconv.AppendInt(buf, int64(frame.Line), 10)
		buf = append(buf, '\n')
	}

	return unsafe.String(&buf[0], len(buf))
}

// funcName formats a function name to a quickly-readable string.
func funcName(fn *runtime.Func) string {
	if fn == nil {
		return ""
	}

	// Get func name
	// for formatting.
	name := fn.Name()

	// Drop all but the package name and function name, no mod path
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}

	const params = `[...]`

	// Drop any generic type parameter markers
	if idx := strings.Index(name, params); idx >= 0 {
		name = name[:idx] + name[idx+len(params):]
	}

	return name
}

// gatherFrames collates runtime frames from a frame iterator.
func gatherFrames(iter *runtime.Frames, n int) callers {
	if iter == nil {
		return nil
	}
	frames := make([]runtime.Frame, 0, n)
	for {
		f, ok := iter.Next()
		if !ok {
			break
		}
		frames = append(frames, f)
	}
	return frames
}
