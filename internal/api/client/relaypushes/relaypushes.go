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
	"fmt"
	"net/http"
	"net/url"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/gin-gonic/gin"
)

const (
	BasePath                    = "/v1/relay_pushes"
	WithID                      = "/:" + apiutil.IDKey
	BasePathWithID              = BasePath + WithID
	RelayPushMatchersPath       = BasePathWithID + "/matchers"
	RelayPushMatchersPathWithID = RelayPushMatchersPath + "/:" + apiutil.RelayMatcherIDKey
)

type Module struct {
	processor *processing.Processor
}

func New(processor *processing.Processor) *Module {
	return &Module{
		processor: processor,
	}
}

func (m *Module) Route(attachHandler func(method string, path string, f ...gin.HandlerFunc) gin.IRoutes) {
	attachHandler(http.MethodGet, BasePath, m.RelayPushesGETHandler)
	attachHandler(http.MethodGet, BasePathWithID, m.RelayPushGETHandler)
	attachHandler(http.MethodPost, BasePath, m.RelayPushPOSTHandler)
	attachHandler(http.MethodPut, BasePathWithID, m.RelayPushPUTHandler)
	attachHandler(http.MethodDelete, BasePathWithID, m.RelayPushDELETEHandler)
	attachHandler(http.MethodPost, RelayPushMatchersPath, m.RelayPushMatcherPOSTHandler)
	attachHandler(http.MethodDelete, RelayPushMatchersPathWithID, m.RelayPushMatcherDELETEHandler)
	attachHandler(http.MethodPut, RelayPushMatchersPathWithID, m.RelayPushMatcherPUTHandler)
}

// RelayPushesGETHandler swagger:operation GET /api/v1/relay_pushes relayPushes
//
// View relay pushes.
//
// The pushes will be returned in descending chronological order (newest first), with sequential IDs (bigger = newer).
//
//	---
//	tags:
//	- relay_pushes
//
//	produces:
//	- application/json
//
//	security:
//	- OAuth2 Bearer:
//		- read:relays
//
//	responses:
//		'200':
//			name: relay pushes
//			description: Array of relay pushes.
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/relayConnection"
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
func (m *Module) RelayPushesGETHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeReadRelays,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	resp, errWithCode := m.processor.RelayPush().RelayPushesGet(c.Request.Context(), authed.Account.ID)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if resp.LinkHeader != "" {
		c.Header("Link", resp.LinkHeader)
	}

	apiutil.JSON(c, http.StatusOK, resp.Items)
}

// RelayPushGETHandler swagger:operation GET /api/v1/relay_pushes/{id} relayPushGet
//
// View relay push with the given ID.
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
//		type: string
//		description: The id of the relay push.
//		in: path
//		required: true
//
//	security:
//	- OAuth2 Bearer:
//		- read:relays
//
//	responses:
//		'200':
//			name: relay push
//			description: The requested relay push.
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
func (m *Module) RelayPushGETHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeReadRelays,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	id, errWithCode := apiutil.ParseID(c.Param(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	resp, errWithCode := m.processor.RelayPush().RelayPushGet(
		c.Request.Context(),
		authed,
		id,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, resp)
}

// RelayPushPOSTHandler swagger:operation POST /api/v1/relay_pushes relayPushCreate
//
// Create a new relay push targeting a remote relay actor URI.
//
// The parameters can also be given in the body of the request, as JSON, if the content-type is set to 'application/json'.
// The parameters can also be given in the body of the request, as XML, if the content-type is set to 'application/xml'.
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
//		name: relay_actor_uri
//		in: formData
//		description: The ActivityPub URI of the remote relay actor.
//		type: string
//		required: true
//	-
//		name: public
//		in: formData
//		description: Push public posts. If false, never send public posts to this relay.
//		type: boolean
//		default: true
//	-
//		name: unlisted
//		in: formData
//		description: Push unlisted posts. If false, never send unlisted posts to this relay.
//		type: boolean
//	-
//		name: match_by_default
//		in: formData
//		description: >-
//			Controls whether the relay push should send all non-ignored posts by default.
//			If set true, and no "exclude"-type matchers are set on the push, then all included, non-ignored posts will be sent.
//		type: boolean
//	-
//		name: ignore_sensitive
//		in: formData
//		description: Never send sensitive posts to this relay.
//		type: boolean
//	-
//		name: ignore_media
//		in: formData
//		description: Never send posts with media attachments to this relay.
//		type: boolean
//	-
//		name: ignore_replies
//		in: formData
//		description: Never send non-self-replies (ie., comments) to this relay.
//		type: boolean
//
//	security:
//	- OAuth2 Bearer:
//		- write:relays
//
//	responses:
//		'200':
//			description: The newly-created relay push.
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
//			description: forbidden
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
//			description: unprocessable -- remote actor URI could not be dereferenced, or remote actor host is blocked
//		'500':
//			schema:
//				"$ref": "#/definitions/error"
//			description: internal server error
func (m *Module) RelayPushPOSTHandler(c *gin.Context) {
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

	form := new(apimodel.RelayConnectionCreateRequest)
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	// Ensure relayActorURI is parseable.
	relayActorURI, err := url.Parse(form.RelayActorURI)
	if err != nil {
		err := fmt.Errorf("relay_actor_uri not parseable: %w", err)
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	resp, errWithCode := m.processor.RelayPush().RelayPushCreate(
		c.Request.Context(),
		authed,
		relayActorURI,
		util.PtrOrValue(form.Public, true),   // default true
		util.PtrOrZero(form.Unlisted),        // default false
		util.PtrOrZero(form.MatchByDefault),  // default false
		util.PtrOrZero(form.IgnoreSensitive), // default false
		util.PtrOrZero(form.IgnoreMedia),     // default false
		util.PtrOrZero(form.IgnoreReplies),   // default false
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, resp)
}

// RelayPushPUTHandler swagger:operation PUT /api/v1/relay_pushes/{id} relayPushUpdate
//
// Update a relay push.
//
// The parameters can also be given in the body of the request, as JSON, if the content-type is set to 'application/json'.
// The parameters can also be given in the body of the request, as XML, if the content-type is set to 'application/xml'.
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
//		type: string
//		description: The id of the relay push.
//		in: path
//		required: true
//	-
//		name: public
//		in: formData
//		description: Push public posts. If false, never send public posts to this relay.
//		type: boolean
//	-
//		name: unlisted
//		in: formData
//		description: Push unlisted posts. If false, never send unlisted posts to this relay.
//		type: boolean
//	-
//		name: match_by_default
//		in: formData
//		description: >-
//			Controls whether the relay push should send all non-ignored posts by default.
//			If set true, and no "exclude"-type matchers are set on the push, then all included, non-ignored posts will be sent.
//		type: boolean
//	-
//		name: ignore_sensitive
//		in: formData
//		description: Never send sensitive posts to this relay.
//		type: boolean
//	-
//		name: ignore_media
//		in: formData
//		description: Never send posts with media attachments to this relay.
//		type: boolean
//	-
//		name: ignore_replies
//		in: formData
//		description: Never send non-self-replies (ie., comments) to this relay.
//		type: boolean
//
//	security:
//	- OAuth2 Bearer:
//		- write:relays
//
//	responses:
//		'200':
//			description: The newly-created relay push.
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
//			description: forbidden
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
//			description: unprocessable -- remote actor URI could not be dereferenced, or remote actor host is blocked
//		'500':
//			schema:
//				"$ref": "#/definitions/error"
//			description: internal server error
func (m *Module) RelayPushPUTHandler(c *gin.Context) {
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

	id, errWithCode := apiutil.ParseID(c.Param(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	// Parse form.
	form := new(apimodel.RelayConnectionCreateRequest)
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	// Ensure something is being updated.
	if form.Public == nil &&
		form.Unlisted == nil &&
		form.MatchByDefault == nil &&
		form.IgnoreSensitive == nil &&
		form.IgnoreMedia == nil &&
		form.IgnoreReplies == nil {
		const errText = "no update fields provided"
		errWithCode := gtserror.NewErrorBadRequest(errors.New(errText), errText)
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	resp, errWithCode := m.processor.RelayPush().RelayPushUpdate(
		c.Request.Context(),
		authed,
		id,
		form.Public,
		form.Unlisted,
		form.MatchByDefault,
		form.IgnoreSensitive,
		form.IgnoreMedia,
		form.IgnoreReplies,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, resp)
}

// RelayPushDELETEHandler swagger:operation DELETE /api/v1/relay_pushes/{id} relayPushDelete
//
// Delete relay push with the given ID.
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
//		type: string
//		description: The id of the relay push.
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
//			description: The deleted relay push.
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
func (m *Module) RelayPushDELETEHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeWriteRelays,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	id, errWithCode := apiutil.ParseID(c.Param(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	resp, errWithCode := m.processor.RelayPush().RelayPushDelete(c.Request.Context(), authed, id)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, resp)
}
