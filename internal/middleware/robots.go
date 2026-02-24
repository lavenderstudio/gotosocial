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
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"github.com/gin-gonic/gin"
)

type RobotsHeadersMode int

const (
	// Disallow indexing via noindex, nofollow.
	// Includes anti-"ai" headers.
	RobotsHeadersModeDefault RobotsHeadersMode = iota
	// Just set anti-"ai" headers and
	// leave the other headers be.
	RobotsHeadersModeDisallowAIOnly
	// Allow some limited indexing.
	// Includes anti-"ai" headers.
	RobotsHeadersModeAllowSome
)

// RobotsHeaders adds robots directives to the X-Robots-Tag HTTP header.
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Robots-Tag
//
// If mode == "aiOnly" then only the noai and noimageai values will be set,
// and other headers will be left alone (for route groups / handlers to set).
//
// If mode == "allowSome" then noai, noimageai, and some indexing will be set.
//
// If mode == "" then noai, noimageai, noindex, and nofollow will be set
// (ie., as restrictive as possible).
func RobotsHeaders(mode RobotsHeadersMode) gin.HandlerFunc {
	const key = "X-Robots-Tag"

	switch mode {
	case RobotsHeadersModeDisallowAIOnly:
		return func(c *gin.Context) {
			c.Writer.Header().Set(key, apiutil.RobotsDirectiveDisallowAI)
		}

	case RobotsHeadersModeAllowSome:
		return func(c *gin.Context) {
			c.Writer.Header().Set(key, apiutil.RobotsDirectivesAllowSome)
		}

	default:
		return func(c *gin.Context) {
			c.Writer.Header().Set(key, apiutil.RobotsDirectivesDisallowIndex)
		}
	}
}
