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
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/gin-gonic/gin"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
)

// RelaySubscriptionsGETHandler swagger:operation GET /api/v1/admin/relay_subscriptions adminRelaySubscriptions
//
// View relay subscriptions.
//
// The subscriptions will be returned in descending chronological order (newest first), with sequential IDs (bigger = newer).
//
//	---
//	tags:
//	- admin
//
//	produces:
//	- application/json
//
//	security:
//	- OAuth2 Bearer:
//		- admin:read:relays
//
//	responses:
//		'200':
//			name: relay subscriptions
//			description: Array of relay subscriptions.
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
func (m *Module) RelaySubscriptionsGETHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeAdminReadRelays,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if !*authed.User.Admin {
		err := fmt.Errorf("user %s not an admin", authed.User.ID)
		apiutil.ErrorHandler(c, gtserror.NewErrorForbidden(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	resp, errWithCode := m.processor.Admin().RelaySubscriptionsGet(c.Request.Context())
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if resp.LinkHeader != "" {
		c.Header("Link", resp.LinkHeader)
	}

	apiutil.JSON(c, http.StatusOK, resp.Items)
}

// RelaySubscriptionGETHandler swagger:operation GET /api/v1/admin/relay_subscriptions/{id} adminRelaySubscriptionGet
//
// View relay subscription with the given ID.
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
//		type: string
//		description: The id of the relay subscription.
//		in: path
//		required: true
//
//	security:
//	- OAuth2 Bearer:
//		- admin:read:relays
//
//	responses:
//		'200':
//			name: relay subscription
//			description: The requested relay subscription.
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
func (m *Module) RelaySubscriptionGETHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeAdminReadRelays,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if !*authed.User.Admin {
		err := fmt.Errorf("user %s not an admin", authed.User.ID)
		apiutil.ErrorHandler(c, gtserror.NewErrorForbidden(err, err.Error()), m.processor.InstanceGetV1)
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

	resp, errWithCode := m.processor.Admin().RelaySubscriptionGet(c.Request.Context(), id)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, resp)
}

// RelaySubscriptionPOSTHandler swagger:operation POST /api/v1/admin/relay_subscriptions relaySubscriptionCreate
//
// Create a new relay subscription targeting a remote relay actor URI.
//
// The parameters can also be given in the body of the request, as JSON, if the content-type is set to 'application/json'.
// The parameters can also be given in the body of the request, as XML, if the content-type is set to 'application/xml'.
//
//	---
//	tags:
//	- admin
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
//		description: Ingest public posts. If false, never ingest public posts via this subscription.
//		type: boolean
//		default: true
//	-
//		name: unlisted
//		in: formData
//		description: Ingest unlisted posts. If false, never ingest unlisted posts via this subscription.
//		type: boolean
//	-
//		name: match_by_default
//		in: formData
//		description: >-
//			Controls whether the relay subscription should ingest all non-ignored posts by default.
//			If set true, and no "exclude"-type matchers are set on the subscription, then all included, non-ignored posts will be ingested.
//		type: boolean
//	-
//		name: ignore_sensitive
//		in: formData
//		description: Never ingest sensitive posts via this subscription.
//		type: boolean
//	-
//		name: ignore_media
//		in: formData
//		description: Never ingest posts with media attachments via this subscription.
//		type: boolean
//	-
//		name: ignore_replies
//		in: formData
//		description: Never ingest non-self-replies (ie., comments) via this subscription.
//		type: boolean
//
//	security:
//	- OAuth2 Bearer:
//		- admin:write:relays
//
//	responses:
//		'200':
//			description: The newly-created relay subscription.
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
func (m *Module) RelaySubscriptionPOSTHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeAdminWriteRelays,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if !*authed.User.Admin {
		err := fmt.Errorf("user %s not an admin", authed.User.ID)
		apiutil.ErrorHandler(c, gtserror.NewErrorForbidden(err, err.Error()), m.processor.InstanceGetV1)
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

	resp, errWithCode := m.processor.Admin().RelaySubscriptionCreate(
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

// RelaySubscriptionPUTHandler swagger:operation PUT /api/v1/admin/relay_subscriptions/{id} relaySubscriptionUpdate
//
// Update a relay subscription.
//
// The parameters can also be given in the body of the request, as JSON, if the content-type is set to 'application/json'.
// The parameters can also be given in the body of the request, as XML, if the content-type is set to 'application/xml'.
//
//	---
//	tags:
//	- admin
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
//		description: The id of the relay subscription.
//		in: path
//		required: true
//	-
//		name: public
//		in: formData
//		description: Ingest public posts. If false, never ingest public posts via this subscription.
//		type: boolean
//	-
//		name: unlisted
//		in: formData
//		description: Ingest unlisted posts. If false, never ingest unlisted posts via this subscription.
//		type: boolean
//	-
//		name: match_by_default
//		in: formData
//		description: >-
//			Controls whether the relay subscription should ingest all non-ignored posts by default.
//			If set true, and no "exclude"-type matchers are set on the subscription, then all included, non-ignored posts will be ingested.
//		type: boolean
//	-
//		name: ignore_sensitive
//		in: formData
//		description: Never ingest sensitive posts via this subscription.
//		type: boolean
//	-
//		name: ignore_media
//		in: formData
//		description: Never ingest posts with media attachments via this subscription.
//		type: boolean
//	-
//		name: ignore_replies
//		in: formData
//		description: Never ingest non-self-replies (ie., comments) via this subscription.
//		type: boolean
//
//	security:
//	- OAuth2 Bearer:
//		- admin:write:relays
//
//	responses:
//		'200':
//			description: The newly-created relay subscription.
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
func (m *Module) RelaySubscriptionPUTHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeAdminWriteRelays,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if !*authed.User.Admin {
		err := fmt.Errorf("user %s not an admin", authed.User.ID)
		apiutil.ErrorHandler(c, gtserror.NewErrorForbidden(err, err.Error()), m.processor.InstanceGetV1)
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
	form := new(apimodel.RelayConnectionUpdateRequest)
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

	resp, errWithCode := m.processor.Admin().RelaySubscriptionUpdate(
		c.Request.Context(),
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

// RelaySubscriptionDELETEHandler swagger:operation DELETE /api/v1/admin/relay_subscriptions/{id} adminRelaySubscriptionDelete
//
// Delete relay subscription with the given ID.
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
//		type: string
//		description: The id of the relay subscription.
//		in: path
//		required: true
//
//	security:
//	- OAuth2 Bearer:
//		- admin:write:relays
//
//	responses:
//		'200':
//			name: relay subscription
//			description: The deleted relay subscription.
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
func (m *Module) RelaySubscriptionDELETEHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeAdminWriteRelays,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if !*authed.User.Admin {
		err := fmt.Errorf("user %s not an admin", authed.User.ID)
		apiutil.ErrorHandler(c, gtserror.NewErrorForbidden(err, err.Error()), m.processor.InstanceGetV1)
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

	resp, errWithCode := m.processor.Admin().RelaySubscriptionDelete(c.Request.Context(), id)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, resp)
}
