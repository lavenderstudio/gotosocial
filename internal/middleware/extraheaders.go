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
	"code.superseriousbusiness.org/gopkg/httputil"
)

// ExtraHeaders returns a new gin middleware
// which adds various extra headers to the response.
func ExtraHeaders() httputil.FlatMiddlewareFunc {
	return func(c *httputil.Context) {

		// Inform all callers which server implementation this is.
		c.W.Header().Set("Server", "gotosocial")

		// Equivalent to CSP frame-ancestors for older browsers
		c.W.Header().Set("X-Frame-Options", "DENY")

		// Don't do MIME type sniffing
		c.W.Header().Set("X-Content-Type-Options", "nosniff")

		// Only send Referer header for URLs matching our protocol, hostname and port
		c.W.Header().Set("Referrer-Policy", "same-origin")

		// Prevent google chrome cohort tracking. Originally this was referred
		// to as FlocBlock. Floc was replaced by Topics in 2022 and the spec says
		// that interest-cohort will also block Topics (as of 2022-Nov).
		//
		// See: https://smartframe.io/blog/google-topics-api-everything-you-need-to-know
		//
		// See: https://github.com/patcg-individual-drafts/topics
		c.W.Header().Set("Permissions-Policy", "browsing-topics=()")
	}
}
