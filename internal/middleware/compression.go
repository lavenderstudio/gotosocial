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

	"code.superseriousbusiness.org/gopkg/httputil"
	"github.com/klauspost/compress/gzhttp"
)

// WithCompression ...
func WithCompression() httputil.MiddlewareFunc {
	return func(h func(*httputil.Context)) func(*httputil.Context) {
		if h == nil {
			panic("nil func")
		}

		hh := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			c := httputil.UnwrapContext(r.Context())
			c.W.RW = rw
			h(c)
		})
		gzh := gzhttp.GzipHandler(hh)
		if gzh == nil {
			panic("nil handler")
		}

		return func(c *httputil.Context) {
			gzh.ServeHTTP(c.W.RW, c.R)
		}
	}
}
