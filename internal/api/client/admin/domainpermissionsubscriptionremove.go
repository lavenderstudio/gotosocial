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

	"code.superseriousbusiness.org/gopkg/httputil"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

// DomainPermissionSubscriptionRemovePOSTHandler swagger:operation POST /api/v1/admin/domain_permission_subscriptions/{id}/remove domainPermissionSubscriptionRemove
//
// Remove a domain permission subscription.
//
//	---
//	tags:
//	- admin
//
//	consumes:
//	- multipart/form-data
//	- application/json
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: id
//		required: true
//		in: path
//		description: ID of the domain permission subscription.
//		type: string
//	-
//		name: remove_children
//		in: formData
//		description: >-
//			When removing the domain permission subscription, also
//			remove children of this subscription, ie., domain permissions
//			that are managed by this subscription. If false, then children
//			will instead be orphaned but not removed.
//
//			Note that removed permissions may end up being created again later
//			by another domain permission subscription of lower priority than
//			the removed subscription. Likewise, orphaned children may be later
//			adopted by another subscription.
//		type: boolean
//		default: true
//
//	security:
//	- OAuth2 Bearer:
//		- admin:write
//
//	responses:
//		'200':
//			description: The removed domain permission subscription.
//			schema:
//				"$ref": "#/definitions/domainPermissionSubscription"
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
//		'409':
//			schema:
//				"$ref": "#/definitions/error"
//			description: conflict
//		'500':
//			schema:
//				"$ref": "#/definitions/error"
//			description: internal server error
func (m *Module) DomainPermissionSubscriptionRemovePOSTHandler(c *httputil.Context) {
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

	id, errWithCode := apiutil.ParseID(c.PathValue(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	type RemoveForm struct {
		RemoveChildren *bool `json:"remove_children" form:"remove_children"`
	}

	form := new(RemoveForm)
	if err := httputil.ShouldBind(c, form, int64(config.GetHTTPServerMaxMultipartMemory())); err != nil { // nolint
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorBadRequest(err, err.Error()))
		return
	}

	// Default removeChildren to true.
	removeChildren := util.PtrOrValue(form.RemoveChildren, true)

	permSub, errWithCode := m.processor.Admin().DomainPermissionSubscriptionRemove(
		c,
		id,
		removeChildren,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.JSON(c, http.StatusOK, permSub)
}
