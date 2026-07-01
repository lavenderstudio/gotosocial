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
	"net/http"

	"code.superseriousbusiness.org/gopkg/httputil"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
)

// AccountFollowPOSTHandler swagger:operation POST /api/v1/accounts/{id}/follow accountFollow
//
// Follow account with id.
//
// The parameters can also be given in the body of the request, as JSON, if the content-type is set to 'application/json'.
// The parameters can also be given in the body of the request, as XML, if the content-type is set to 'application/xml'.
//
// If you already follow (request) the given account, then the follow (request) will be updated instead using the
// `reblogs` and `notify` parameters.
//
//	---
//	tags:
//	- accounts
//
//	consumes:
//	- application/json
//	- application/xml
//	- application/x-www-form-urlencoded
//
//	parameters:
//	-
//		name: id
//		required: true
//		in: path
//		description: ID of the account to follow.
//		type: string
//	-
//		name: reblogs
//		type: boolean
//		default: true
//		description: Show reblogs from this account.
//		in: formData
//	-
//		name: notify
//		type: boolean
//		default: false
//		description: Notify when this account posts.
//		in: formData
//
//	produces:
//	- application/json
//
//	security:
//	- OAuth2 Bearer:
//		- write:follows
//
//	responses:
//		'200':
//			name: account relationship
//			description: Your relationship to this account.
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
func (m *Module) AccountFollowPOSTHandler(c *httputil.Context) {
	authed, errWithCode := apiutil.TokenAuth(c, apiutil.AuthRequirements{
		Token:   true,
		App:     true,
		User:    true,
		Account: true,
		Scope:   []apiutil.Scope{apiutil.ScopeWriteFollows},
	})
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
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

	targetAcctID, errWithCode := apiutil.ParseID(c.PathValue(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	form := &apimodel.AccountFollowRequest{}
	if err := httputil.ShouldBind(c, form, int64(config.GetHTTPServerMaxMultipartMemory())); err != nil { // nolint
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorBadRequest(err, err.Error()))
		return
	}
	form.ID = targetAcctID

	relationship, errWithCode := m.processor.Account().FollowCreate(c, authed.Account, form)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.JSON(c, http.StatusOK, relationship)
}
