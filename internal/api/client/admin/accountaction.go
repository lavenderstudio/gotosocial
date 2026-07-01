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
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
)

// AccountActionPOSTHandler swagger:operation POST /api/v1/admin/accounts/{id}/action adminAccountAction
//
// Perform an admin action on an account.
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
//		name: id
//		required: true
//		in: path
//		description: ID of the account.
//		type: string
//	-
//		name: type
//		in: formData
//		description: Type of action to be taken, currently only supports `suspend`.
//		type: string
//		required: true
//	-
//		name: text
//		in: formData
//		description: Optional text describing why this action was taken.
//		type: string
//
//	security:
//	- OAuth2 Bearer:
//		- admin:write:accounts
//
//	responses:
//		'200':
//			description: OK
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
//		'409':
//			schema:
//				"$ref": "#/definitions/error"
//			description: >-
//				Conflict: There is already an admin action running that conflicts with this action.
//				Check the error message in the response body for more information. This is a temporary
//				error; it should be possible to process this action if you try again in a bit.
//		'500':
//			schema:
//				"$ref": "#/definitions/error"
//			description: internal server error
func (m *Module) AccountActionPOSTHandler(c *httputil.Context) {
	authed, errWithCode := apiutil.TokenAuth(c, apiutil.AuthRequirements{
		Token:   true,
		App:     true,
		User:    true,
		Account: true,
		Scope:   []apiutil.Scope{apiutil.ScopeAdminWriteAccounts},
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

	form := &apimodel.AdminActionRequest{}
	if err := httputil.ShouldBind(c, form, int64(config.GetHTTPServerMaxMultipartMemory())); err != nil { // nolint
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorBadRequest(err, err.Error()))
		return
	}

	if form.Type == "" {
		err := errors.New("no type specified")
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorBadRequest(err, err.Error()))
		return
	}

	targetAcctID, errWithCode := apiutil.ParseID(c.PathValue(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}
	form.TargetID = targetAcctID

	if _, errWithCode := m.processor.Admin().AccountAction(c, authed.Account, form); errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.JSON(c, http.StatusOK, map[string]string{
		"message": "OK",
	})
}
