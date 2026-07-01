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

package tags

import (
	"net/http"

	"code.superseriousbusiness.org/gopkg/httputil"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
)

// UnfollowTagPOSTHandler swagger:operation POST /api/v1/tags/{tag_name}/unfollow unfollowTag
//
// Unfollow a hashtag.
//
// Idempotent: if you are not following the tag, this call will still succeed.
//
//	---
//	tags:
//	- tags
//
//	produces:
//	- application/json
//
//	security:
//	- OAuth2 Bearer:
//		- write:follows
//
//	parameters:
//	-
//		name: tag_name
//		type: string
//		description: Name of the tag (no leading `#`)
//		in: path
//		required: true
//
//	responses:
//		'200':
//			description: "Info about the tag."
//			schema:
//				"$ref": "#/definitions/tag"
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
//			description: unauthorized
//		'500':
//			schema:
//				"$ref": "#/definitions/error"
//			description: internal server error
func (m *Module) UnfollowTagPOSTHandler(c *httputil.Context) {
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

	name, errWithCode := apiutil.ParseTagName(c.PathValue(apiutil.TagNameKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	apiTag, errWithCode := m.processor.Tags().Unfollow(c, authed.Account, name)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.JSON(c, http.StatusOK, apiTag)
}
