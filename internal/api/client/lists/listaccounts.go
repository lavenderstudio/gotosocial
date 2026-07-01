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

package lists

import (
	"net/http"

	"code.superseriousbusiness.org/gopkg/httputil"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

// ListAccountsGETHandler swagger:operation GET /api/v1/lists/{id}/accounts listAccounts
//
// Page through accounts in this list.
//
// The returned Link header can be used to generate the previous and next queries when scrolling up or down a timeline.
//
// Example:
//
// ```
// <https://example.org/api/v1/list/01H0W619198FX7J54NF7EH1NG2/accounts?limit=20&max_id=01FC3GSQ8A3MMJ43BPZSGEG29M>; rel="next", <https://example.org/api/v1/list/01H0W619198FX7J54NF7EH1NG2/accounts?limit=20&min_id=01FC3KJW2GYXSDDRA6RWNDM46M>; rel="prev"
// ````
//
//	---
//	tags:
//	- lists
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
//			Return only list entries *OLDER* than the given max ID.
//			The account from the list entry with the specified ID will not be included in the response.
//		in: query
//		required: false
//	-
//		name: since_id
//		type: string
//		description: >-
//			Return only list entries *NEWER* than the given since ID.
//			The account from the list entry with the specified ID will not be included in the response.
//		in: query
//	-
//		name: min_id
//		type: string
//		description: >-
//			Return only list entries *IMMEDIATELY NEWER* than the given min ID.
//			The account from the list entry with the specified ID will not be included in the response.
//		in: query
//		required: false
//	-
//		name: limit
//		type: integer
//		description: >-
//			Number of accounts to return.
//			If set to 0 explicitly, all accounts in the list will be returned, and pagination headers will not be used.
//			This is a workaround for Mastodon API peculiarities: https://docs.joinmastodon.org/methods/lists/#query-parameters.
//		default: 40
//		minimum: 0
//		maximum: 80
//		in: query
//		required: false
//
//	security:
//	- OAuth2 Bearer:
//		- read:lists
//
//	responses:
//		'200':
//			headers:
//				Link:
//					type: string
//					description: Links to the next and previous queries.
//			name: accounts
//			description: Array of accounts.
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/account"
//		'400':
//			schema:
//				"$ref": "#/definitions/error"
//			description: bad request
//		'401':
//			schema:
//				"$ref": "#/definitions/error"
//			description: unauthorized
//		'404':
//			schema:
//				"$ref": "#/definitions/error"
//			description: not found
//		'406':
//			schema:
//				"$ref": "#/definitions/error"
//			description: not acceptable
//		'500':
//			schema:
//				"$ref": "#/definitions/error"
//			description: internal server error
func (m *Module) ListAccountsGETHandler(c *httputil.Context) {
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

	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	targetListID, errWithCode := apiutil.ParseID(c.PathValue(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	page, errWithCode := paging.ParseIDPage(c,
		1,  // min limit
		80, // max limit
		0,  // default = paging disabled
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	resp, errWithCode := m.processor.List().GetListAccounts(
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
