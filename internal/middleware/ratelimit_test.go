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

package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strconv"
	"testing"
	"time"

	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gotosocial/internal/middleware"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/stretchr/testify/suite"
)

type RateLimitTestSuite struct {
	suite.Suite
}

func (suite *RateLimitTestSuite) TestRateLimit() {
	const (
		trustedPlatform = "X-Test-IP"
		rlLimit         = "X-RateLimit-Limit"
		rlRemaining     = "X-RateLimit-Remaining"
		rlReset         = "X-RateLimit-Reset"
	)

	type rlTest struct {
		limit        int
		exceptions   []netip.Prefix
		clientIP     string
		shouldExcept bool
	}

	for _, test := range []rlTest{
		{
			limit:        10,
			exceptions:   nil,
			clientIP:     "192.0.2.0",
			shouldExcept: false,
		},
		{
			limit:        10,
			exceptions:   nil,
			clientIP:     "192.0.2.0",
			shouldExcept: false,
		},
		{
			limit:        10,
			exceptions:   []netip.Prefix{netip.MustParsePrefix("192.0.2.0/24")},
			clientIP:     "192.0.2.0",
			shouldExcept: true,
		},
		{
			limit:        10,
			exceptions:   []netip.Prefix{netip.MustParsePrefix("192.0.2.0/32")},
			clientIP:     "192.0.2.1",
			shouldExcept: false,
		},
	} {
		// Compile handler from required client IP
		// determining middlware, and rate limiter.
		h := httputil.Compile([]httputil.Middleware{
			middleware.WithClientIP(&httputil.ProxyConfiguration{RemoteIPHeaders: []string{trustedPlatform}}),
			middleware.RateLimit(test.limit, test.exceptions),
		},
			func(c *httputil.Context) { c.W.WriteHeader(200) },
		)

		// Approximate time when limiter will reset.
		resetAt := time.Now().Add(5 * time.Minute)

		// Make requests up to +
		// just over the limit.
		limitedAt := test.limit + 1
		for requestsCount := 1; requestsCount < limitedAt; requestsCount++ {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/example", nil)
			req.Header.Add(trustedPlatform, test.clientIP)
			c := httputil.ToContext(rec, req)

			// Call rate
			// limiter.
			h(c)

			var (
				// Fetch RL headers if present.
				limitStr     = rec.Header().Get(rlLimit)
				remainingStr = rec.Header().Get(rlRemaining)
				resetStr     = rec.Header().Get(rlReset)
			)

			if test.shouldExcept {
				// Request should be allowed through,
				// no rate-limit headers should be written.
				suite.Equal(http.StatusOK, rec.Code)
				suite.Empty(limitStr)
				suite.Empty(remainingStr)
				suite.Empty(resetStr)
				continue
			}

			suite.Equal(strconv.Itoa(test.limit), limitStr)
			suite.Equal(strconv.Itoa(test.limit-requestsCount), remainingStr)

			// Ensure reset is ISO8601, and resets at
			// approximate reset time (+/- 10 seconds).
			reset, err := util.ParseISO8601(resetStr)
			suite.NoError(err)
			suite.WithinDuration(resetAt, reset, 10*time.Second)

			if requestsCount < limitedAt {
				// Request should be allowed through.
				suite.Equal(http.StatusOK, rec.Code)
				continue
			}

			// Request should be denied.
			suite.Equal(http.StatusTooManyRequests, rec.Code)

			// Make a final request with an unrelated IP to
			// ensure it's only the one IP being limited.
			req = httptest.NewRequest("GET", "/example", nil)
			req.Header.Add(trustedPlatform, "192.0.2.255")
			rec = httptest.NewRecorder()
			c = httputil.ToContext(rec, req)

			// Call rate
			// limiter.
			h(c)

			// Request should be allowed through.
			suite.Equal(http.StatusOK, rec.Code)
		}
	}
}

func TestRateLimitTestSuite(t *testing.T) {
	suite.Run(t, new(RateLimitTestSuite))
}
