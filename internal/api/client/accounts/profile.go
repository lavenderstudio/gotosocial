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

package accounts

import (
	"context"
	"net/http"

	"code.superseriousbusiness.org/gopkg/httputil"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// AccountAvatarDELETEHandler swagger:operation DELETE /api/v1/profile/avatar accountAvatarDelete
//
// Delete the authenticated account's avatar.
// If the account doesn't have an avatar, the call succeeds anyway.
//
//	---
//	tags:
//	- accounts
//
//	produces:
//	- application/json
//
//	security:
//	- OAuth2 Bearer:
//		- admin
//
//	responses:
//		'200':
//			description: The updated account, including profile source information.
//			schema:
//				"$ref": "#/definitions/account"
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
//		'406':
//			schema:
//				"$ref": "#/definitions/error"
//			description: not acceptable
//		'500':
//			schema:
//				"$ref": "#/definitions/error"
//			description: internal server error
func (m *Module) AccountAvatarDELETEHandler(c *httputil.Context) {
	m.accountDeleteProfileAttachment(c, m.processor.Media().DeleteAvatar)
}

// AccountHeaderDELETEHandler swagger:operation DELETE /api/v1/profile/header accountHeaderDelete
//
// Delete the authenticated account's header.
// If the account doesn't have a header, the call succeeds anyway.
//
//	---
//	tags:
//	- accounts
//
//	produces:
//	- application/json
//
//	security:
//	- OAuth2 Bearer:
//		- admin
//
//	responses:
//		'200':
//			description: The updated account, including profile source information.
//			schema:
//				"$ref": "#/definitions/account"
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
//		'406':
//			schema:
//				"$ref": "#/definitions/error"
//			description: not acceptable
//		'500':
//			schema:
//				"$ref": "#/definitions/error"
//			description: internal server error
func (m *Module) AccountHeaderDELETEHandler(c *httputil.Context) {
	m.accountDeleteProfileAttachment(c, m.processor.Media().DeleteHeader)
}

// accountDeleteProfileAttachment checks that an authenticated account is present and allowed to alter itself,
// runs an attachment deletion processor method, and returns the updated account.
func (m *Module) accountDeleteProfileAttachment(c *httputil.Context, processDelete func(context.Context, *gtsmodel.Account) (*apimodel.Account, gtserror.WithCode)) {
	authed, errWithCode := apiutil.TokenAuth(c, apiutil.AuthRequirements{
		Token:   true,
		App:     true,
		User:    true,
		Account: true,
		Scope:   []apiutil.Scope{apiutil.ScopeWriteAccounts},
	})
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	acctSensitive, errWithCode := processDelete(c, authed.Account)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.JSON(c, http.StatusOK, acctSensitive)
}
