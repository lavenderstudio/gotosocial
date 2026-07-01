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

package timelines

import (
	"net/http"

	"code.superseriousbusiness.org/gopkg/httputil"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

// ListTimelineGETHandler swagger:operation GET /api/v1/timelines/list/{id} listTimeline
//
// See statuses/posts from the given list timeline.
//
// The statuses will be returned in descending chronological order (newest first), with sequential IDs (bigger = newer).
//
// The returned Link header can be used to generate the previous and next queries when scrolling up or down a timeline.
//
// Example:
//
// ```
// <https://example.org/api/v1/timelines/list/01H0W619198FX7J54NF7EH1NG2?limit=20&max_id=01FC3GSQ8A3MMJ43BPZSGEG29M>; rel="next", <https://example.org/api/v1/timelines/list/01H0W619198FX7J54NF7EH1NG2?limit=20&min_id=01FC3KJW2GYXSDDRA6RWNDM46M>; rel="prev"
// ````
//
//	---
//	tags:
//	- timelines
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: id
//		type: string
//		description: ID of the list
//		in: path
//		required: true
//	-
//		name: max_id
//		type: string
//		description: >-
//			Return only statuses *OLDER* than the given max status ID.
//			The status with the specified ID will not be included in the response.
//		in: query
//		required: false
//	-
//		name: since_id
//		type: string
//		description: >-
//			Return only statuses *NEWER* than the given since status ID.
//			The status with the specified ID will not be included in the response.
//		in: query
//	-
//		name: min_id
//		type: string
//		description: >-
//			Return only statuses *NEWER* than the given since status ID.
//			The status with the specified ID will not be included in the response.
//		in: query
//		required: false
//	-
//		name: limit
//		type: integer
//		description: Number of statuses to return.
//		default: 20
//		in: query
//		required: false
//
//	security:
//	- OAuth2 Bearer:
//		- read:lists
//
//	responses:
//		'200':
//			name: statuses
//			description: Array of statuses.
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/status"
//			headers:
//				Link:
//					type: string
//					description: Links to the next and previous queries.
//		'401':
//			schema:
//				"$ref": "#/definitions/error"
//			description: unauthorized
//		'400':
//			schema:
//				"$ref": "#/definitions/error"
//			description: bad request
func (m *Module) ListTimelineGETHandler(c *httputil.Context) {
	authed, errWithCode := apiutil.TokenAuth(c, apiutil.AuthRequirements{
		Token:   true,
		App:     true,
		User:    true,
		Account: true,
		Scope:   []apiutil.Scope{apiutil.ScopeReadLists},
	})
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	if authed.Account.IsMoving() {
		// For moving/moved accounts, just return
		// empty to avoid breaking client apps.
		httputil.Data(c, http.StatusOK, apiutil.AppJSON, apiutil.EmptyJSONArray)
		return
	}

	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	targetListID, errWithCode := apiutil.ParseID(c.PathValue(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
	}

	page, errWithCode := paging.ParseIDPage(c,
		1,  // min limit
		40, // max limit
		20, // default limit
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	resp, errWithCode := m.processor.Timeline().ListTimelineGet(
		c,
		authed.Account,
		targetListID,
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
