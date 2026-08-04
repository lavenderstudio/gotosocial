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

	"code.superseriousbusiness.org/gopkg/httputil"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/validate"
)

// RelayActorsGETHandler swagger:operation GET /api/v1/admin/relay_actors adminRelayActors
//
// View relay actors.
//
// The actors will be returned in descending chronological order (newest first), with sequential IDs (bigger = newer).
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
//			name: relay actors
//			description: Array of relay actors.
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/relayActor"
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
func (m *Module) RelayActorsGETHandler(c *httputil.Context) {
	authed, errWithCode := apiutil.TokenAuth(c, apiutil.AuthRequirements{
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

	if !*authed.User.Admin {
		err := fmt.Errorf("user %s not an admin", authed.User.ID)
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorForbidden(err, err.Error()))
		return
	}

	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	resp, errWithCode := m.processor.Admin().RelayActorsGet(c)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.JSON(c, http.StatusOK, resp)
}

// RelayActorGETHandler swagger:operation GET /api/v1/admin/relay_actors/{id} adminRelayActorGet
//
// View relay actor with the given ID.
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
//		description: The id of the relay actor.
//		in: path
//		required: true
//
//	security:
//	- OAuth2 Bearer:
//		- admin:read:relays
//
//	responses:
//		'200':
//			name: relay actor
//			description: The requested relay actor.
//			schema:
//				"$ref": "#/definitions/relayActor"
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
func (m *Module) RelayActorGETHandler(c *httputil.Context) {
	authed, errWithCode := apiutil.TokenAuth(c, apiutil.AuthRequirements{
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

	resp, errWithCode := m.processor.Admin().RelayActorGet(c, id)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.JSON(c, http.StatusOK, resp)
}

// RelayActorPOSTHandler swagger:operation POST /api/v1/admin/relay_actors relayActorCreate
//
// Create a new relay actor.
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
//	security:
//	- OAuth2 Application:
//		- admin:write:relays
//
//	parameters:
//	-
//		name: username
//		in: formData
//		description: >-
//			The desired username for the relay actor account.
//			Will be prefixed with "relay." to form the final username.
//			Eg., "username=example" results in username "relay.example".
//		type: string
//	-
//		name: discoverable
//		in: formData
//		description: Relay actor account should be made discoverable and shown in the profile directory (if enabled).
//		type: boolean
//	-
//		name: display_name
//		in: formData
//		description: The display name to use for the account.
//		type: string
//		allowEmptyValue: true
//	-
//		name: note
//		in: formData
//		description: Bio/description of this account.
//		type: string
//		allowEmptyValue: true
//	-
//		name: avatar
//		in: formData
//		description: Avatar of the user.
//		type: file
//	-
//		name: avatar_description
//		in: formData
//		description: Description of avatar image, for alt-text.
//		type: string
//		allowEmptyValue: true
//	-
//		name: header
//		in: formData
//		description: Header of the user.
//		type: file
//	-
//		name: header_description
//		in: formData
//		description: Description of header image, for alt-text.
//		type: string
//		allowEmptyValue: true
//	-
//		name: locked
//		in: formData
//		description: Require manual approval of follow requests.
//		type: boolean
//	-
//		name: fields_attributes[0][name]
//		in: formData
//		description: Name of 1st profile field to be added to this account's profile.
//			(The index may be any string; add more indexes to send more fields.)
//		type: string
//	-
//		name: fields_attributes[0][value]
//		in: formData
//		description: Value of 1st profile field to be added to this account's profile.
//			(The index may be any string; add more indexes to send more fields.)
//		type: string
//	-
//		name: fields_attributes[1][name]
//		in: formData
//		description: Name of 2nd profile field to be added to this account's profile.
//		type: string
//	-
//		name: fields_attributes[1][value]
//		in: formData
//		description: Value of 2nd profile field to be added to this account's profile.
//		type: string
//	-
//		name: fields_attributes[2][name]
//		in: formData
//		description: Name of 3rd profile field to be added to this account's profile.
//		type: string
//	-
//		name: fields_attributes[2][value]
//		in: formData
//		description: Value of 3rd profile field to be added to this account's profile.
//		type: string
//	-
//		name: fields_attributes[3][name]
//		in: formData
//		description: Name of 4th profile field to be added to this account's profile.
//		type: string
//	-
//		name: fields_attributes[3][value]
//		in: formData
//		description: Value of 4th profile field to be added to this account's profile.
//		type: string
//	-
//		name: fields_attributes[4][name]
//		in: formData
//		description: Name of 5th profile field to be added to this account's profile.
//		type: string
//	-
//		name: fields_attributes[4][value]
//		in: formData
//		description: Value of 5th profile field to be added to this account's profile.
//		type: string
//	-
//		name: fields_attributes[5][name]
//		in: formData
//		description: Name of 6th profile field to be added to this account's profile.
//		type: string
//	-
//		name: fields_attributes[5][value]
//		in: formData
//		description: Value of 6th profile field to be added to this account's profile.
//		type: string
//	-
//		name: public
//		in: formData
//		description: Relay public posts. If false, never relay public posts via this actor.
//		type: boolean
//		default: true
//	-
//		name: unlisted
//		in: formData
//		description: Relay unlisted posts. If false, never relay unlisted posts via this actor.
//		type: boolean
//	-
//		name: match_by_default
//		in: formData
//		description: >-
//			Controls whether the relay actor should relay all non-ignored posts by default.
//			If set true, and no "exclude"-type matchers are set on the actor, then all included, non-ignored posts will be relayed.
//		type: boolean
//	-
//		name: ignore_sensitive
//		in: formData
//		description: Never relay sensitive posts via this actor.
//		type: boolean
//	-
//		name: ignore_media
//		in: formData
//		description: Never relay posts with media attachments via this actor.
//		type: boolean
//	-
//		name: ignore_replies
//		in: formData
//		description: Never relay non-self-replies (ie., comments) via this actor.
//		type: boolean
//
//	responses:
//		'200':
//			description: The newly-created relay actor.
//			schema:
//				"$ref": "#/definitions/relayActor"
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
//		'500':
//			schema:
//				"$ref": "#/definitions/error"
//			description: internal server error
func (m *Module) RelayActorPOSTHandler(c *httputil.Context) {
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

	// Parse form.
	f, err := apiutil.ParseWithFieldsAttributes(c, new(apimodel.RelayActorCreateRequest))
	if err != nil {
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorBadRequest(err, err.Error()))
		return
	}

	// Safe type assertion as we just
	// instantiated this ourselves above.
	form := f.(*apimodel.RelayActorCreateRequest)

	// Ensure username OK.
	if err := validate.Username(form.Username); err != nil {
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorBadRequest(err, err.Error()))
		return
	}

	resp, errWithCode := m.processor.Admin().RelayActorCreate(c, authed, form)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.JSON(c, http.StatusOK, resp)
}

// RelayActorPUTHandler swagger:operation PUT /api/v1/admin/relay_actors/{id} relayActorUpdate
//
// Update a relay actor.
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
//	security:
//	- OAuth2 Bearer:
//		- admin:write:relays
//
//	parameters:
//	-
//		name: id
//		type: string
//		description: The id of the relay actor.
//		in: path
//		required: true
//	-
//		name: discoverable
//		in: formData
//		description: Relay actor account should be made discoverable and shown in the profile directory (if enabled).
//		type: boolean
//	-
//		name: display_name
//		in: formData
//		description: The display name to use for the account.
//		type: string
//	-
//		name: note
//		in: formData
//		description: Bio/description of this account.
//		type: string
//	-
//		name: avatar
//		in: formData
//		description: Avatar of the user.
//		type: file
//	-
//		name: avatar_description
//		in: formData
//		description: Description of avatar image, for alt-text.
//		type: string
//	-
//		name: header
//		in: formData
//		description: Header of the user.
//		type: file
//	-
//		name: header_description
//		in: formData
//		description: Description of header image, for alt-text.
//		type: string
//	-
//		name: locked
//		in: formData
//		description: Require manual approval of follow requests.
//		type: boolean
//	-
//		name: fields_attributes[0][name]
//		in: formData
//		description: Name of 1st profile field to be added to this account's profile.
//			(The index may be any string; add more indexes to send more fields.)
//		type: string
//	-
//		name: fields_attributes[0][value]
//		in: formData
//		description: Value of 1st profile field to be added to this account's profile.
//			(The index may be any string; add more indexes to send more fields.)
//		type: string
//	-
//		name: fields_attributes[1][name]
//		in: formData
//		description: Name of 2nd profile field to be added to this account's profile.
//		type: string
//	-
//		name: fields_attributes[1][value]
//		in: formData
//		description: Value of 2nd profile field to be added to this account's profile.
//		type: string
//	-
//		name: fields_attributes[2][name]
//		in: formData
//		description: Name of 3rd profile field to be added to this account's profile.
//		type: string
//	-
//		name: fields_attributes[2][value]
//		in: formData
//		description: Value of 3rd profile field to be added to this account's profile.
//		type: string
//	-
//		name: fields_attributes[3][name]
//		in: formData
//		description: Name of 4th profile field to be added to this account's profile.
//		type: string
//	-
//		name: fields_attributes[3][value]
//		in: formData
//		description: Value of 4th profile field to be added to this account's profile.
//		type: string
//	-
//		name: fields_attributes[4][name]
//		in: formData
//		description: Name of 5th profile field to be added to this account's profile.
//		type: string
//	-
//		name: fields_attributes[4][value]
//		in: formData
//		description: Value of 5th profile field to be added to this account's profile.
//		type: string
//	-
//		name: fields_attributes[5][name]
//		in: formData
//		description: Name of 6th profile field to be added to this account's profile.
//		type: string
//	-
//		name: fields_attributes[5][value]
//		in: formData
//		description: Value of 6th profile field to be added to this account's profile.
//		type: string
//	-
//		name: public
//		in: formData
//		description: Relay public posts. If false, never relay public posts via this actor.
//		type: boolean
//		default: true
//	-
//		name: unlisted
//		in: formData
//		description: Relay unlisted posts. If false, never relay unlisted posts via this actor.
//		type: boolean
//	-
//		name: match_by_default
//		in: formData
//		description: >-
//			Controls whether the relay actor should relay all non-ignored posts by default.
//			If set true, and no "exclude"-type matchers are set on the actor, then all included, non-ignored posts will be relayed.
//		type: boolean
//	-
//		name: ignore_sensitive
//		in: formData
//		description: Never relay sensitive posts via this actor.
//		type: boolean
//	-
//		name: ignore_media
//		in: formData
//		description: Never relay posts with media attachments via this actor.
//		type: boolean
//	-
//		name: ignore_replies
//		in: formData
//		description: Never relay non-self-replies (ie., comments) via this actor.
//		type: boolean
//
//	responses:
//		'200':
//			description: The newly-created relay actor.
//			schema:
//				"$ref": "#/definitions/relayActor"
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
func (m *Module) RelayActorPUTHandler(c *httputil.Context) {
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

	// Parse form.
	f, err := apiutil.ParseWithFieldsAttributes(c, new(apimodel.RelayActorUpdateRequest))
	if err != nil {
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorBadRequest(err, err.Error()))
		return
	}

	// Safe type assertion as we just
	// instantiated this ourselves above.
	form := f.(*apimodel.RelayActorUpdateRequest)

	// Ensure something is being updated.
	if form.Discoverable == nil &&
		form.DisplayName == nil &&
		form.Note == nil &&
		form.Avatar == nil &&
		form.AvatarDescription == nil &&
		form.Header == nil &&
		form.HeaderDescription == nil &&
		form.Locked == nil &&
		form.FieldsAttributes == nil &&
		form.Public == nil &&
		form.Unlisted == nil &&
		form.MatchByDefault == nil &&
		form.IgnoreSensitive == nil &&
		form.IgnoreMedia == nil &&
		form.IgnoreReplies == nil {
		const errText = "no update fields provided"
		errWithCode := gtserror.NewErrorBadRequest(errors.New(errText), errText)
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	resp, errWithCode := m.processor.Admin().RelayActorUpdate(c, id, form)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.JSON(c, http.StatusOK, resp)
}

// RelayActorHeaderDELETEHandler swagger:operation DELETE /api/v1/admin/relay_actors/{id}/profile/header adminRelayActorHeaderDelete
//
// Delete header image of account of relay actor with the given ID.
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
//		description: The id of the relay actor.
//		in: path
//		required: true
//
//	security:
//	- OAuth2 Bearer:
//		- admin:write:relays
//
//	responses:
//		'200':
//			name: relay actor
//			description: The updated relay actor.
//			schema:
//				"$ref": "#/definitions/relayActor"
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
func (m *Module) RelayActorHeaderDELETEHandler(c *httputil.Context) {
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

	resp, errWithCode := m.processor.Admin().RelayActorDeleteHeader(c, id)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.JSON(c, http.StatusOK, resp)
}

// RelayActorAvatarDELETEHandler swagger:operation DELETE /api/v1/admin/relay_actors/{id}/profile/avatar adminRelayActorAvatarDelete
//
// Delete avatar image of account of relay actor with the given ID.
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
//		description: The id of the relay actor.
//		in: path
//		required: true
//
//	security:
//	- OAuth2 Bearer:
//		- admin:write:relays
//
//	responses:
//		'200':
//			name: relay actor
//			description: The updated relay actor.
//			schema:
//				"$ref": "#/definitions/relayActor"
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
func (m *Module) RelayActorAvatarDELETEHandler(c *httputil.Context) {
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

	resp, errWithCode := m.processor.Admin().RelayActorDeleteAvatar(c, id)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.JSON(c, http.StatusOK, resp)
}

// RelayActorDELETEHandler swagger:operation DELETE /api/v1/admin/relay_actors/{id} adminRelayActorDelete
//
// Delete relay actor with the given ID.
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
//		description: The id of the relay actor.
//		in: path
//		required: true
//
//	security:
//	- OAuth2 Bearer:
//		- admin:write:relays
//
//	responses:
//		'200':
//			name: relay actor
//			description: The deleted relay actor.
//			schema:
//				"$ref": "#/definitions/relayActor"
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
func (m *Module) RelayActorDELETEHandler(c *httputil.Context) {
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

	resp, errWithCode := m.processor.Admin().RelayActorDelete(c, authed, id)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.JSON(c, http.StatusOK, resp)
}
