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
	"net/mail"

	"code.superseriousbusiness.org/gopkg/httputil"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
)

// EmailTestPostHandler swagger:operation POST /api/v1/admin/email/test testEmailSend
//
// Send a generic test email to a specified email address.
//
// This can be used to validate an instance's SMTP configuration, and to debug any potential issues.
//
// If an error is returned by the SMTP connection, this handler will return code 422 to indicate that
// the request could not be processed, and the SMTP error will be returned to the caller.
//
//	---
//	tags:
//	- admin
//
//	consumes:
//	- multipart/form-data
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: email
//		in: formData
//		description: The email address that the test email should be sent to.
//		type: string
//		required: true
//	-
//		name: message
//		in: formData
//		description: Optional message to include in the email.
//		type: string
//
//	security:
//	- OAuth2 Bearer:
//		- admin:write
//
//	responses:
//		'202':
//			description: Test email was sent.
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
//			description: >-
//				An smtp occurred while the email attempt was in progress.
//				Check the returned json for more information. The smtp error
//				will be included, to help you debug communication with the
//				smtp server.
//		'500':
//			schema:
//				"$ref": "#/definitions/error"
//			description: internal server error
func (m *Module) EmailTestPOSTHandler(c *httputil.Context) {
	authed, errWithCode := apiutil.TokenAuth(c, apiutil.AuthRequirements{
		Token:   true,
		App:     true,
		User:    true,
		Account: true,
		Scope:   []apiutil.Scope{apiutil.ScopeAdminWrite},
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

	form := &apimodel.AdminSendTestEmailRequest{}
	if err := httputil.ShouldBind(c, form, int64(config.GetHTTPServerMaxMultipartMemory())); err != nil { // nolint
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorBadRequest(err, err.Error()))
		return
	}

	email, err := mail.ParseAddress(form.Email)
	if err != nil {
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorBadRequest(err, err.Error()))
		return
	}

	errWithCode = m.processor.Admin().EmailTest(
		c,
		authed.Account,
		email.Address,
		form.Message,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.JSON(c, http.StatusAccepted, map[string]string{
		"status": "test email sent",
	})
}
