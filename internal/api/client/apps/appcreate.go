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

package apps

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gopkg/httputil/binding"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
)

// these consts are used to ensure users can't spam huge entries into our database
const (
	formFieldLen    = 1024
	formRedirectLen = 2056
)

// AppsPOSTHandler swagger:operation POST /api/v1/apps appCreate
//
// Register a new application on this instance.
//
// The registered application can be used to obtain an application token.
// This can then be used to register a new account, or (through user auth) obtain an access token.
//
// If the application was registered with a Bearer token passed in the Authorization header, the created application will be managed by the authenticated user (must have scope write:applications).
//
// Parameters can also be given in the body of the request, as JSON, if the content-type is set to 'application/json'.
// Parameters can also be given in the body of the request, as XML, if the content-type is set to 'application/xml'.
//
//	---
//	tags:
//	- apps
//
//	consumes:
//	- application/json
//	- application/xml
//	- application/x-www-form-urlencoded
//
//	produces:
//	- application/json
//
//	responses:
//		'200':
//			description: "The newly-created application."
//			schema:
//				"$ref": "#/definitions/application"
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
func (m *Module) AppsPOSTHandler(c *httputil.Context) {
	authed, errWithCode := apiutil.TokenAuth(c, apiutil.AuthRequirements{
		Token:   false,
		App:     false,
		User:    false,
		Account: false,
		Scope:   nil,
	})
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	if authed.Token != nil {
		// If a token has been passed, user
		// needs write perm on applications.
		if !slices.ContainsFunc(
			strings.Split(authed.Token.GetScope(), " "),
			func(hasScope string) bool {
				return apiutil.Scope(hasScope).Permits(apiutil.ScopeWriteApplications)
			},
		) {
			const errText = "token has insufficient scope permission"
			errWithCode := gtserror.NewErrorForbidden(errors.New(errText), errText)
			apiutil.ErrorHandler(c, m.templates, errWithCode)
			return
		}
	}

	if authed.Account != nil && authed.Account.IsMoving() {
		apiutil.ForbiddenAfterMove(c)
		return
	}

	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	form := &apimodel.ApplicationCreateRequest{}
	if err := binding.ShouldBind(c, form, int64(config.GetHTTPServerMaxMultipartMemory())); err != nil { // nolint
		errWithCode := gtserror.NewErrorBadRequest(err, err.Error())
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	if l := len([]rune(form.ClientName)); l > formFieldLen {
		m.fieldTooLong(c, "client_name", formFieldLen, l)
		return
	}

	if l := len([]rune(form.RedirectURIs)); l > formRedirectLen {
		m.fieldTooLong(c, "redirect_uris", formRedirectLen, l)
		return
	}

	if l := len([]rune(form.Scopes)); l > formFieldLen {
		m.fieldTooLong(c, "scopes", formFieldLen, l)
		return
	}

	if l := len([]rune(form.Website)); l > formFieldLen {
		m.fieldTooLong(c, "website", formFieldLen, l)
		return
	}

	var managedByUserID string
	if authed.User != nil {
		managedByUserID = authed.User.ID
	}

	apiApp, errWithCode := m.processor.Application().Create(c, managedByUserID, form)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.JSON(c, http.StatusOK, apiApp)
}

func (m *Module) fieldTooLong(c *httputil.Context, fieldName string, max int, actual int) {
	errText := fmt.Sprintf(
		"%s must be less than %d characters, provided %s was %d characters",
		fieldName, max, fieldName, actual,
	)

	errWithCode := gtserror.NewErrorBadRequest(errors.New(errText), errText)
	apiutil.ErrorHandler(c, m.templates, errWithCode)
}
