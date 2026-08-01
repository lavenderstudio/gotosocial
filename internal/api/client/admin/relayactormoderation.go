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

// RelayActorFollowRequestsGETHandler swagger:operation GET /api/v1/admin/relay_actors/{id}/follow_requests relayActorFollowRequests
//
// See follow requests targeting relay actor account.
//
// The next and previous queries can be parsed from the returned Link header.
// Example:
//
// ```
// <https://example.org/api/v1/admin/relay_actors/0657WMDEC3KQDTD6NZ4XJZBK4M/follow_requests?limit=80&max_id=01FC0SKA48HNSVR6YKZCQGS2V8>; rel="next", <https://example.org/api/v1/admin/relay_actors/0657WMDEC3KQDTD6NZ4XJZBK4M/follow_requests?limit=80&min_id=01FC0SKW5JK2Q4EVAV2B462YY0>; rel="prev"
// ````
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
//		name: id
//		in: path
//		type: string
//		required: true
//		description: ID of the relay actor.
//	-
//		name: max_id
//		type: string
//		description: >-
//			Return only items *OLDER* than the given max ID.
//			The item with the specified ID will not be included in the response.
//			NOTE: the ID is of the internal follow request, NOT any of the returned accounts.
//		in: query
//		required: false
//	-
//		name: since_id
//		type: string
//		description: >-
//			Return only items *NEWER* than the given since ID.
//			The item with the specified ID will not be included in the response.
//			NOTE: the ID is of the internal follow request, NOT any of the returned accounts.
//		in: query
//		required: false
//	-
//		name: min_id
//		type: string
//		description: >-
//			Return only items *IMMEDIATELY NEWER* than the given min ID.
//			The item with the specified ID will not be included in the response.
//			NOTE: the ID is of the internal follow request, NOT any of the returned accounts.
//		in: query
//		required: false
//	-
//		name: limit
//		type: integer
//		description: Number of items to return.
//		default: 40
//		minimum: 1
//		maximum: 80
//		in: query
//		required: false
//
//	security:
//	- OAuth2 Bearer:
//		- admin:read:relays
//
//	responses:
//		'200':
//			name: accounts
//			description: Array of accounts that follow request the relay actor account.
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/account"
//			headers:
//				Link:
//					type: string
//					description: Links to the next and previous queries.
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
func (m *Module) RelayActorFollowRequestsGETHandler(c *httputil.Context) {
	_, errWithCode := apiutil.TokenAuth(c, apiutil.AuthRequirements{
		Token:   true,
		App:     true,
		User:    true,
		Account: true,
		Scope:   []apiutil.Scope{apiutil.ScopeAdminReadRelays},
	})
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	id, errWithCode := apiutil.ParseID(c.PathValue(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	page, errWithCode := paging.ParseIDPage(c,
		1,  // min limit
		80, // max limit
		40, // default limit
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	resp, errWithCode := m.processor.Admin().RelayActorFollowRequestsGet(c, id, page)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	if resp.LinkHeader != "" {
		c.W.Header().Set("Link", resp.LinkHeader)
	}

	httputil.JSON(c, http.StatusOK, resp.Items)
}

// RelayActorFollowRequestAuthorizePOSTHandler swagger:operation POST /api/v1/admin/relay_actors/{id}/follow_requests/{target_account_id}/authorize relayActorAuthorizeFollowRequest
//
// Accept/authorize follow request from the given account ID.
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
//		name: id
//		in: path
//		type: string
//		required: true
//		description: ID of the relay actor.
//	-
//		name: target_account_id
//		in: path
//		type: string
//		required: true
//		description: ID of the account whose follow request to accept.
//
//	security:
//	- OAuth2 Bearer:
//		- admin:write:relays
//
//	responses:
//		'200':
//			name: account relationship
//			description: The relay actor account's updated relationship to the target account.
//			schema:
//				"$ref": "#/definitions/accountRelationship"
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
func (m *Module) RelayActorFollowRequestAuthorizePOSTHandler(c *httputil.Context) {
	authed, errWithCode := apiutil.TokenAuth(c, apiutil.AuthRequirements{
		Token:   true,
		App:     true,
		User:    true,
		Account: true,
		Scope:   []apiutil.Scope{apiutil.ScopeAdminWriteRelays},
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

	id, errWithCode := apiutil.ParseID(c.PathValue(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	targetAccountID, errWithCode := apiutil.ParseTargetAccountID(c.PathValue(apiutil.TargetAccountIDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	resp, errWithCode := m.processor.Admin().RelayActorFollowRequestAccept(c, id, targetAccountID)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.JSON(c, http.StatusOK, resp)
}

// RelayActorFollowRequestRejectPOSTHandler swagger:operation POST /api/v1/admin/relay_actors/{id}/follow_requests/{target_account_id}/reject relayActorRejectFollowRequest
//
// Reject follow request from the given account ID.
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
//		name: id
//		in: path
//		type: string
//		required: true
//		description: ID of the relay actor.
//	-
//		name: target_account_id
//		in: path
//		type: string
//		required: true
//		description: ID of the account whose follow request to reject.
//
//	security:
//	- OAuth2 Bearer:
//		- admin:write:relays
//
//	responses:
//		'200':
//			name: account relationship
//			description: The relay actor account's updated relationship to the target account.
//			schema:
//				"$ref": "#/definitions/accountRelationship"
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
func (m *Module) RelayActorFollowRequestRejectPOSTHandler(c *httputil.Context) {
	authed, errWithCode := apiutil.TokenAuth(c, apiutil.AuthRequirements{
		Token:   true,
		App:     true,
		User:    true,
		Account: true,
		Scope:   []apiutil.Scope{apiutil.ScopeAdminWriteRelays},
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

	id, errWithCode := apiutil.ParseID(c.PathValue(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	targetAccountID, errWithCode := apiutil.ParseTargetAccountID(c.PathValue(apiutil.TargetAccountIDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	resp, errWithCode := m.processor.Admin().RelayActorFollowRequestReject(c, id, targetAccountID)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.JSON(c, http.StatusOK, resp)
}

// RelayActorFollowersGETHandler swagger:operation GET /api/v1/admin/relay_actors/{id}/followers relayActorFollowers
//
// See followers of account with given relay actor id.
//
// The next and previous queries can be parsed from the returned Link header.
// Example:
//
// ```
// <https://example.org/api/v1/admin/relay_actors/0657WMDEC3KQDTD6NZ4XJZBK4M/followers?limit=80&max_id=01FC0SKA48HNSVR6YKZCQGS2V8>; rel="next", <https://example.org/api/v1/admin/relay_actors/0657WMDEC3KQDTD6NZ4XJZBK4M/followers?limit=80&min_id=01FC0SKW5JK2Q4EVAV2B462YY0>; rel="prev"
// ````
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
//		name: id
//		in: path
//		type: string
//		required: true
//		description: ID of the relay actor.
//	-
//		name: max_id
//		type: string
//		description: >-
//			Return only items *OLDER* than the given max ID.
//			The item with the specified ID will not be included in the response.
//			NOTE: the ID is of the internal follow, NOT any of the returned accounts.
//		in: query
//		required: false
//	-
//		name: since_id
//		type: string
//		description: >-
//			Return only items *NEWER* than the given since ID.
//			The item with the specified ID will not be included in the response.
//			NOTE: the ID is of the internal follow, NOT any of the returned accounts.
//		in: query
//		required: false
//	-
//		name: min_id
//		type: string
//		description: >-
//			Return only items *IMMEDIATELY NEWER* than the given min ID.
//			The item with the specified ID will not be included in the response.
//			NOTE: the ID is of the internal follow, NOT any of the returned accounts.
//		in: query
//		required: false
//	-
//		name: limit
//		type: integer
//		description: Number of items to return.
//		default: 40
//		minimum: 1
//		maximum: 80
//		in: query
//		required: false
//
//	security:
//	- OAuth2 Bearer:
//		- admin:read:relays
//
//	responses:
//		'200':
//			name: accounts
//			description: Array of accounts that follow the relay actor account.
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/account"
//			headers:
//				Link:
//					type: string
//					description: Links to the next and previous queries.
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
func (m *Module) RelayActorFollowersGETHandler(c *httputil.Context) {
	_, errWithCode := apiutil.TokenAuth(c, apiutil.AuthRequirements{
		Token:   true,
		App:     true,
		User:    true,
		Account: true,
		Scope:   []apiutil.Scope{apiutil.ScopeAdminReadRelays},
	})
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	id, errWithCode := apiutil.ParseID(c.PathValue(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	page, errWithCode := paging.ParseIDPage(c,
		1,  // min limit
		80, // max limit
		40, // default limit
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	resp, errWithCode := m.processor.Admin().RelayActorFollowersGet(c, id, page)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	if resp.LinkHeader != "" {
		c.W.Header().Set("Link", resp.LinkHeader)
	}

	httputil.JSON(c, http.StatusOK, resp.Items)
}

// RelayActorRemoveFromFollowersPOSTHandler swagger:operation POST /api/v1/admin/relay_actors/{id}/accounts/{target_account_id}/remove_from_followers relayActorAccountRemoveFromFollowers
//
// Remove the given account from the followers of the relay actor.
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
//		name: id
//		in: path
//		type: string
//		required: true
//		description: ID of the relay actor.
//	-
//		name: target_account_id
//		in: path
//		type: string
//		required: true
//		description: ID of the account to remove from followers.
//
//	security:
//	- OAuth2 Bearer:
//		- admin:write:relays
//
//	responses:
//		'200':
//			name: account relationship
//			description: The relay actor account's updated relationship to the target account.
//			schema:
//				"$ref": "#/definitions/accountRelationship"
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
func (m *Module) RelayActorRemoveFromFollowersPOSTHandler(c *httputil.Context) {
	authed, errWithCode := apiutil.TokenAuth(c, apiutil.AuthRequirements{
		Token:   true,
		App:     true,
		User:    true,
		Account: true,
		Scope:   []apiutil.Scope{apiutil.ScopeAdminWriteRelays},
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

	id, errWithCode := apiutil.ParseID(c.PathValue(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	targetAccountID, errWithCode := apiutil.ParseTargetAccountID(c.PathValue(apiutil.TargetAccountIDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	resp, errWithCode := m.processor.Admin().RelayActorFollowerRemove(c, id, targetAccountID)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.JSON(c, http.StatusOK, resp)
}

// RelayActorBlocksGETHandler swagger:operation GET /api/v1/admin/relay_actors/{id}/blocks relayActorBlocks
//
// See accounts blocked by the relay actor account.
//
// The next and previous queries can be parsed from the returned Link header.
// Example:
//
// ```
// <https://example.org/api/v1/admin/relay_actors/0657WMDEC3KQDTD6NZ4XJZBK4M/blocks?limit=80&max_id=01FC0SKA48HNSVR6YKZCQGS2V8>; rel="next", <https://example.org/api/v1/admin/relay_actors/0657WMDEC3KQDTD6NZ4XJZBK4M/blocks?limit=80&min_id=01FC0SKW5JK2Q4EVAV2B462YY0>; rel="prev"
// ````
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
//		name: id
//		in: path
//		type: string
//		required: true
//		description: ID of the relay actor.
//	-
//		name: max_id
//		type: string
//		description: >-
//			Return only items *OLDER* than the given max ID.
//			The item with the specified ID will not be included in the response.
//			NOTE: the ID is of the internal block, NOT any of the returned accounts.
//		in: query
//		required: false
//	-
//		name: since_id
//		type: string
//		description: >-
//			Return only items *NEWER* than the given since ID.
//			The item with the specified ID will not be included in the response.
//			NOTE: the ID is of the internal block, NOT any of the returned accounts.
//		in: query
//		required: false
//	-
//		name: min_id
//		type: string
//		description: >-
//			Return only items *IMMEDIATELY NEWER* than the given min ID.
//			The item with the specified ID will not be included in the response.
//			NOTE: the ID is of the internal block, NOT any of the returned accounts.
//		in: query
//		required: false
//	-
//		name: limit
//		type: integer
//		description: Number of items to return.
//		default: 40
//		minimum: 1
//		maximum: 80
//		in: query
//		required: false
//
//	security:
//	- OAuth2 Bearer:
//		- admin:read:relays
//
//	responses:
//		'200':
//			name: accounts
//			description: Array of accounts blocked by the relay actor account.
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/account"
//			headers:
//				Link:
//					type: string
//					description: Links to the next and previous queries.
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
func (m *Module) RelayActorBlocksGETHandler(c *httputil.Context) {
	_, errWithCode := apiutil.TokenAuth(c, apiutil.AuthRequirements{
		Token:   true,
		App:     true,
		User:    true,
		Account: true,
		Scope:   []apiutil.Scope{apiutil.ScopeAdminReadRelays},
	})
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	id, errWithCode := apiutil.ParseID(c.PathValue(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	page, errWithCode := paging.ParseIDPage(c,
		1,  // min limit
		80, // max limit
		40, // default limit
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	resp, errWithCode := m.processor.Admin().RelayActorBlocksGet(c, id, page)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	if resp.LinkHeader != "" {
		c.W.Header().Set("Link", resp.LinkHeader)
	}

	httputil.JSON(c, http.StatusOK, resp.Items)
}

// RelayActorBlockPOSTHandler swagger:operation POST /api/v1/admin/relay_actors/{id}/accounts/{target_account_id}/block relayActorAccountBlock
//
// Create a block targeting the given account on behalf of the relay actor.
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
//		name: id
//		in: path
//		type: string
//		required: true
//		description: ID of the relay actor.
//	-
//		name: target_account_id
//		in: path
//		type: string
//		required: true
//		description: ID of the account to block.
//
//	security:
//	- OAuth2 Bearer:
//		- admin:write:relays
//
//	responses:
//		'200':
//			name: account relationship
//			description: The relay actor account's updated relationship to the target account.
//			schema:
//				"$ref": "#/definitions/accountRelationship"
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
func (m *Module) RelayActorBlockPOSTHandler(c *httputil.Context) {
	authed, errWithCode := apiutil.TokenAuth(c, apiutil.AuthRequirements{
		Token:   true,
		App:     true,
		User:    true,
		Account: true,
		Scope:   []apiutil.Scope{apiutil.ScopeAdminWriteRelays},
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

	id, errWithCode := apiutil.ParseID(c.PathValue(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	targetAccountID, errWithCode := apiutil.ParseTargetAccountID(c.PathValue(apiutil.TargetAccountIDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	resp, errWithCode := m.processor.Admin().RelayActorBlock(c, id, targetAccountID)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.JSON(c, http.StatusOK, resp)
}

// RelayActorUnblockPOSTHandler swagger:operation POST /api/v1/admin/relay_actors/{id}/accounts/{target_account_id}/unblock relayActorAccountUnblock
//
// Remove a block targeting the given account on behalf of the relay actor.
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
//		name: id
//		in: path
//		type: string
//		required: true
//		description: ID of the relay actor.
//	-
//		name: target_account_id
//		in: path
//		type: string
//		required: true
//		description: ID of the account to block.
//
//	security:
//	- OAuth2 Bearer:
//		- admin:write:relays
//
//	responses:
//		'200':
//			name: account relationship
//			description: The relay actor account's updated relationship to the target account.
//			schema:
//				"$ref": "#/definitions/accountRelationship"
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
func (m *Module) RelayActorUnblockPOSTHandler(c *httputil.Context) {
	authed, errWithCode := apiutil.TokenAuth(c, apiutil.AuthRequirements{
		Token:   true,
		App:     true,
		User:    true,
		Account: true,
		Scope:   []apiutil.Scope{apiutil.ScopeAdminWriteRelays},
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

	id, errWithCode := apiutil.ParseID(c.PathValue(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	targetAccountID, errWithCode := apiutil.ParseTargetAccountID(c.PathValue(apiutil.TargetAccountIDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	resp, errWithCode := m.processor.Admin().RelayActorUnblock(c, id, targetAccountID)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.JSON(c, http.StatusOK, resp)
}
