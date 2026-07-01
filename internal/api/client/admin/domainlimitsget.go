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

package admin

import (
	"fmt"
	"net/http"

	"code.superseriousbusiness.org/gopkg/httputil"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

// DomainLimitsGETHandler swagger:operation GET /api/v1/admin/domain_limits domainLimitsGet
//
// View domain limits currently in place.
//
// By default, items will be returned in alphabetical order by domain.
//
// Paging is optional for this endpoint. To page, use the `limit`, `max_id` and/or `min_id` params. Else, all items will be returned.
//
// If paging, the next and previous queries can be parsed from the returned Link header.
//
// Example:
//
// ```
// <https://example.org/api/v1/admin/domain_limits?limit=20&max_id=01FC0SKA48HNSVR6YKZCQGS2V8>; rel="next", <https://example.org/api/v1/admin/domain_limits?limit=20&min_id=01FC0SKW5JK2Q4EVAV2B462YY0>; rel="prev"
// ````
//
// If paging, items will be returned in descending chronological order (newest first), with sequential IDs (bigger = newer).
//
//	---
//	tags:
//	- admin
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: max_id
//		type: string
//		description: >-
//			Return only items *OLDER* than the given max ID (for paging downwards).
//			The item with the specified ID will not be included in the response.
//		in: query
//	-
//		name: since_id
//		type: string
//		description: >-
//			Return only items *NEWER* than the given since ID.
//			The item with the specified ID will not be included in the response.
//		in: query
//	-
//		name: min_id
//		type: string
//		description: >-
//			Return only items immediately *NEWER* than the given min ID (for paging upwards).
//			The item with the specified ID will not be included in the response.
//		in: query
//	-
//		name: limit
//		type: integer
//		description: Number of items to return. Use 0 to return all (no paging).
//		default: 20
//		minimum: 0
//		maximum: 100
//		in: query
//
//	security:
//	- OAuth2 Bearer:
//		- admin:read:domain_limits
//
//	responses:
//		'200':
//			description: Domain limits.
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/domainLimit"
//			headers:
//				Link:
//					type: string
//					description: Links to the next and previous queries. Omitted if empty page or not paging.
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'403':
//			description: forbidden
//		'406':
//			description: not acceptable
//		'500':
//			description: internal server error
func (m *Module) DomainLimitsGETHandler(c *httputil.Context) {
	authed, errWithCode := apiutil.TokenAuth(c, apiutil.AuthRequirements{
		Token:   true,
		App:     true,
		User:    true,
		Account: true,
		Scope:   []apiutil.Scope{apiutil.ScopeAdminReadDomainLimits},
	})
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	if !*authed.User.Admin {
		err := fmt.Errorf("user %s not an admin", authed.User.ID)
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorForbidden(err, err.Error()))
		return
	}

	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	// Allow paging but don't
	// enforce or use it by default.
	page, errWithCode := paging.ParseIDPage(c,
		0,   // min items
		100, // max items
		0,   // default (no paging)
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	resp, errWithCode := m.processor.Admin().DomainLimitsGet(
		c,
		page,
	)

	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	if resp.LinkHeader != "" {
		c.W.Header().Set("Link", resp.LinkHeader)
	}

	httputil.JSON(c, http.StatusOK, resp.Items)
}
