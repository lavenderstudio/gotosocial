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

package relaypushes

import (
	"errors"
	"net/http"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/gin-gonic/gin"
)

// RelayPushMatcherPOSTHandler swagger:operation POST /api/v1/relay_pushes/{id}/matchers relayPushMatcherPost
//
// Add a relay matcher to a relay push.
//
// Returns the relay push with the given matcher added to it.
//
//	---
//	tags:
//	- relay_pushes
//
//	consumes:
//	- application/json
//	- application/xml
//	- application/x-www-form-urlencoded
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
//		description: ID of the relay push.
//	-
//		name: keyword
//		in: formData
//		required: true
//		description: The text to be matched.
//		type: string
//	-
//		name: whole_word
//		in: formData
//		description: Matcher should consider word boundaries.
//		type: boolean
//		default: false
//	-
//		name: exclude
//		in: formData
//		description: Matcher should cause matched posts to be excluded from relaying rather than included.
//		type: boolean
//		default: false
//
//	security:
//	- OAuth2 Bearer:
//		- write:relays
//
//	responses:
//		'200':
//			name: relay push
//			description: Relay push the newly-created matcher belongs to.
//			schema:
//				"$ref": "#/definitions/relayConnection"
//		'400':
//			schema:
//				"$ref": "#/definitions/error"
//			description: bad request
//		'401':
//			schema:
//				"$ref": "#/definitions/error"
//			description: unauthorized
//		'403':
//			schema:
//				"$ref": "#/definitions/error"
//			description: forbidden to moved accounts
//		'404':
//			schema:
//				"$ref": "#/definitions/error"
//			description: not found
//		'406':
//			schema:
//				"$ref": "#/definitions/error"
//			description: not acceptable
//		'422':
//			schema:
//				"$ref": "#/definitions/error"
//			description: unprocessable content
//		'500':
//			schema:
//				"$ref": "#/definitions/error"
//			description: internal server error
func (m *Module) RelayPushMatcherPOSTHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeWriteRelays,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if authed.Account.IsMoving() {
		apiutil.ForbiddenAfterMove(c)
		return
	}

	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	relayPushID, errWithCode := apiutil.ParseID(c.Param(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	form := new(apimodel.RelayMatcherCreateUpdateRequest)
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	// Ensure keyword is set.
	if form.Keyword == nil || *form.Keyword == "" {
		const errText = "keyword not provided"
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(errors.New(errText), errText), m.processor.InstanceGetV1)
		return
	}

	resp, errWithCode := m.processor.RelayPush().MatcherCreate(
		c.Request.Context(),
		authed,
		relayPushID,
		*form.Keyword,
		util.PtrOrZero(form.WholeWord),
		util.PtrOrZero(form.Exclude),
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, resp)
}

// RelayPushMatcherDELETEHandler swagger:operation DELETE /api/v1/relay_pushes/{id}/matchers/{matcher_id} relayPushMatcherDelete
//
// Remove a relay matcher from a relay push.
//
// Returns the relay push with the given matcher removed from it.
//
//	---
//	tags:
//	- relay_pushes
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
//		description: ID of the relay push.
//	-
//		name: matcher_id
//		type: string
//		description: ID of the relay matcher.
//		in: path
//		required: true
//
//	security:
//	- OAuth2 Bearer:
//		- write:relays
//
//	responses:
//		'200':
//			name: relay push
//			description: Relay push the now-deleted matcher belonged to.
//			schema:
//				"$ref": "#/definitions/relayConnection"
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
func (m *Module) RelayPushMatcherDELETEHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeWriteRelays,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if authed.Account.IsMoving() {
		apiutil.ForbiddenAfterMove(c)
		return
	}

	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	relayID, errWithCode := apiutil.ParseID(c.Param(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	matcherID, errWithCode := apiutil.ParseRelayMatcherID(c.Param(apiutil.RelayMatcherIDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	resp, errWithCode := m.processor.RelayPush().MatcherDelete(
		c.Request.Context(),
		authed,
		relayID,
		matcherID,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, resp)
}

// RelayPushMatcherPUTHandler swagger:operation PUT /api/v1/relay_pushes/{id}/matchers/{matcher_id} relayPushMatcherPut
//
// Update a relay matcher on a relay push.
//
//	---
//	tags:
//	- relay_pushes
//
//	consumes:
//	- application/json
//	- application/xml
//	- application/x-www-form-urlencoded
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
//		description: ID of the relay push.
//	-
//		name: matcher_id
//		type: string
//		description: ID of the relay matcher.
//		in: path
//		required: true
//	-
//		name: keyword
//		in: formData
//		required: true
//		description: The text to be matched.
//		type: string
//	-
//		name: whole_word
//		in: formData
//		description: Matcher should consider word boundaries.
//		type: boolean
//		default: false
//	-
//		name: exclude
//		in: formData
//		description: Matcher should cause matched posts to be excluded from relaying rather than included.
//		type: boolean
//		default: false
//
//	security:
//	- OAuth2 Bearer:
//		- write:relays
//
//	responses:
//		'200':
//			name: relay push
//			description: Relay push the now-updated matcher belongs to.
//			schema:
//				"$ref": "#/definitions/relayConnection"
//		'400':
//			schema:
//				"$ref": "#/definitions/error"
//			description: bad request
//		'401':
//			schema:
//				"$ref": "#/definitions/error"
//			description: unauthorized
//		'403':
//			schema:
//				"$ref": "#/definitions/error"
//			description: forbidden to moved accounts
//		'404':
//			schema:
//				"$ref": "#/definitions/error"
//			description: not found
//		'406':
//			schema:
//				"$ref": "#/definitions/error"
//			description: not acceptable
//		'409':
//			schema:
//				"$ref": "#/definitions/error"
//			description: conflict (duplicate keyword)
//		'422':
//			schema:
//				"$ref": "#/definitions/error"
//			description: unprocessable content
//		'500':
//			schema:
//				"$ref": "#/definitions/error"
//			description: internal server error
func (m *Module) RelayPushMatcherPUTHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeWriteRelays,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if authed.Account.IsMoving() {
		apiutil.ForbiddenAfterMove(c)
		return
	}

	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	relayID, errWithCode := apiutil.ParseID(c.Param(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	matcherID, errWithCode := apiutil.ParseRelayMatcherID(c.Param(apiutil.RelayMatcherIDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	form := new(apimodel.RelayMatcherCreateUpdateRequest)
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	resp, errWithCode := m.processor.RelayPush().MatcherUpdate(
		c.Request.Context(),
		authed,
		relayID,
		matcherID,
		form.Keyword,
		form.WholeWord,
		form.Exclude,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, resp)
}
