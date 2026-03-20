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

package directory

import (
	"math"
	"net/http"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"github.com/gin-gonic/gin"
)

// DirectoryGETHandler swagger:operation GET /api/v1/directory directoryGet
//
// Get an array of accounts **on this instance** that have marked themselves as being "discoverable" in the directory.
//
// If instance-directory-mode is set to "full" then this API endpoint will be available without authentication, otherwise a token must be provided.
//
// Unlike Mastodon and other fedi softwares, this endpoint only supports showing local accounts.
//
//	---
//	tags:
//	- directory
//
//	produces:
//	- application/json
//
//	security:
//	- OAuth2 Bearer:
//		- read:directory
//
//	parameters:
//	-
//		name: offset
//		type: integer
//		description: >-
//			Skip the first n results.
//			If offset is provided, other paging parameters will be ignored.
//		in: query
//		required: false
//	-
//		name: max_id
//		type: string
//		description: >-
//			Return only items after than the given max ID.
//			The item with the specified ID will not be included in the response.
//			Parameter ignored if offset is specified.
//		in: query
//		required: false
//	-
//		name: since_id
//		type: string
//		description: >-
//			Return only items before the given since ID.
//			The item with the specified ID will not be included in the response.
//			Parameter ignored if offset is specified.
//		in: query
//		required: false
//	-
//		name: min_id
//		type: string
//		description: >-
//			Return only items *IMMEDIATELY BEFORE* the given min ID.
//			The item with the specified ID will not be included in the response.
//			Parameter ignored if offset is specified.
//		in: query
//		required: false
//	-
//		name: limit
//		type: integer
//		description: Number of accounts to return.
//		default: 40
//		minimum: 1
//		maximum: 80
//		in: query
//		required: false
//	-
//		name: order
//		type: string
//		description: >-
//			Use 'active' to sort by most recently posted statuses (default),
//			or 'new' to sort by most recently created profiles.
//		in: query
//		required: false
//
//	responses:
//		'200':
//			description: Array of accounts.
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/account"
//			headers:
//				Link:
//					type: string
//					description: Links to the next and previous queries.
//		'401':
//			schema:
//				"$ref": "#/definitions/error"
//			description: unauthorized
//		'406':
//			schema:
//				"$ref": "#/definitions/error"
//			description: not acceptable
//		'500':
//			schema:
//				"$ref": "#/definitions/error"
//			description: internal server error
func (m *Module) DirectoryGETHandler(c *gin.Context) {
	// If directory is not exposed to unauthed
	// callers, fail if a token was not provided.
	var requester *gtsmodel.Account
	if config.GetInstanceDirectoryMode() != config.InstanceDirectoryModeOpen {
		auth, errWithCode := apiutil.TokenAuth(c,
			true, true, true, true,
			apiutil.ScopeReadDirectory,
		)
		if errWithCode != nil {
			apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
			return
		}
		requester = auth.Account
	}

	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	// Parse paging params.
	page, errWithCode := paging.ParseIDPage(c,
		1,  // min limit
		80, // max limit
		40, // default limit
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	// Parse order (default "active").
	orderBy, errWithCode := apiutil.ParseDirectoryOrder(
		c.Query(apiutil.OrderKey),
		gtsmodel.DirectoryOrderByActive,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	// Parse offset (default 0).
	offset, errWithCode := apiutil.ParseOffset(c.Query(apiutil.OffsetKey), 0, math.MaxInt, 0)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	resp, errWithCode := m.processor.Account().DirectoryGet(
		c.Request.Context(),
		requester,
		page,
		offset,
		orderBy,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if resp.LinkHeader != "" {
		c.Header("Link", resp.LinkHeader)
	}

	apiutil.JSON(c, http.StatusOK, resp.Items)
}
