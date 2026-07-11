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
	"code.superseriousbusiness.org/gopkg/httputil/binding"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

// RelaySubscriptionMatcherPOSTHandler swagger:operation POST /api/v1/admin/relay_subscriptions/{id}/matchers relaySubscriptionMatcherPost
//
// Add a relay matcher to a relay subscription.
//
// Returns the relay subscription with the given matcher added to it.
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
//		in: path
//		type: string
//		required: true
//		description: ID of the relay subscription.
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
//		- admin:write:relays
//
//	responses:
//		'200':
//			name: relay subscription
//			description: Relay subscription the newly-created matcher belongs to.
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
func (m *Module) RelaySubscriptionMatcherPOSTHandler(c *httputil.Context) {
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

	if authed.Account.IsMoving() {
		apiutil.ForbiddenAfterMove(c)
		return
	}

	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	relaySubscriptionID, errWithCode := apiutil.ParseID(c.PathValue(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	form := new(apimodel.RelayMatcherCreateUpdateRequest)
	if err := binding.ShouldBind(c, form, int64(config.GetHTTPServerMaxMultipartMemory())); err != nil { // nolint
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorBadRequest(err, err.Error()))
		return
	}

	// Ensure keyword is set.
	if form.Keyword == nil || *form.Keyword == "" {
		const errText = "keyword not provided"
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorBadRequest(errors.New(errText), errText))
		return
	}

	resp, errWithCode := m.processor.Admin().RelaySubscriptionMatcherCreate(
		c,
		relaySubscriptionID,
		*form.Keyword,
		util.PtrOrZero(form.WholeWord),
		util.PtrOrZero(form.Exclude),
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.JSON(c, http.StatusOK, resp)
}

// RelaySubscriptionMatcherDELETEHandler swagger:operation DELETE /api/v1/admin/relay_subscriptions/{id}/matchers/{matcher_id} relaySubscriptionMatcherDelete
//
// Remove a relay matcher from a relay subscription.
//
// Returns the relay subscription with the given matcher removed from it.
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
//		description: ID of the relay subscription.
//	-
//		name: matcher_id
//		type: string
//		description: ID of the relay matcher.
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
//			description: Relay subscription the now-deleted matcher belonged to.
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
func (m *Module) RelaySubscriptionMatcherDELETEHandler(c *httputil.Context) {
	authed, errWithCode := apiutil.TokenAuth(c, apiutil.AuthRequirements{
		Token:   true,
		App:     true,
		User:    true,
		Account: true,
		Scope:   []apiutil.Scope{apiutil.ScopeWriteRelays},
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

	if authed.Account.IsMoving() {
		apiutil.ForbiddenAfterMove(c)
		return
	}

	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	relayID, errWithCode := apiutil.ParseID(c.PathValue(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	matcherID, errWithCode := apiutil.ParseRelayMatcherID(c.PathValue(apiutil.RelayMatcherIDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	resp, errWithCode := m.processor.Admin().RelaySubscriptionMatcherDelete(
		c,
		relayID,
		matcherID,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.JSON(c, http.StatusOK, resp)
}

// RelaySubscriptionMatcherPUTHandler swagger:operation PUT /api/v1/admin/relay_subscriptions/{id}/matchers/{matcher_id} relaySubscriptionMatcherPut
//
// Update a relay matcher on a relay subscription.
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
//		in: path
//		type: string
//		required: true
//		description: ID of the relay subscription.
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
//		- admin:write:relays
//
//	responses:
//		'200':
//			name: relay subscription
//			description: Relay subscription the now-updated matcher belongs to.
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
func (m *Module) RelaySubscriptionMatcherPUTHandler(c *httputil.Context) {
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

	if authed.Account.IsMoving() {
		apiutil.ForbiddenAfterMove(c)
		return
	}

	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	relayID, errWithCode := apiutil.ParseID(c.PathValue(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	matcherID, errWithCode := apiutil.ParseRelayMatcherID(c.PathValue(apiutil.RelayMatcherIDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	form := new(apimodel.RelayMatcherCreateUpdateRequest)
	if err := binding.ShouldBind(c, form, int64(config.GetHTTPServerMaxMultipartMemory())); err != nil { // nolint
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorBadRequest(err, err.Error()))
		return
	}

	resp, errWithCode := m.processor.Admin().RelaySubscriptionMatcherUpdate(
		c,
		relayID,
		matcherID,
		form.Keyword,
		form.WholeWord,
		form.Exclude,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.JSON(c, http.StatusOK, resp)
}
