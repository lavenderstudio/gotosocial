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

// DomainPermissionDraftsPOSTHandler swagger:operation POST /api/v1/admin/domain_permission_drafts domainPermissionDraftCreate
//
// Create a domain permission draft with the given parameters.
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
//		name: domain
//		in: formData
//		description: Domain to create the permission draft for.
//		type: string
//	-
//		name: permission_type
//		in: formData
//		description: Create a draft "allow" or a draft "block".
//		type: string
//	-
//		name: obfuscate
//		in: formData
//		description: >-
//			Obfuscate the name of the domain when serving it publicly.
//			Eg., `example.org` becomes something like `ex***e.org`.
//		type: boolean
//	-
//		name: public_comment
//		in: formData
//		description: >-
//			Public comment about this domain permission.
//			This will be displayed alongside the domain permission if you choose to share permissions.
//		type: string
//	-
//		name: private_comment
//		in: formData
//		description: >-
//			Private comment about this domain permission. Will only be shown to other admins, so this
//			is a useful way of internally keeping track of why a certain domain ended up permissioned.
//		type: string
//
//	security:
//	- OAuth2 Bearer:
//		- admin:write
//
//	responses:
//		'200':
//			description: The newly created domain permission draft.
//			schema:
//				"$ref": "#/definitions/domainPermission"
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
func (m *Module) DomainPermissionDraftsPOSTHandler(c *httputil.Context) {
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

	// Parse + validate form.
	form := new(apimodel.DomainPermissionRequest)
	if err := binding.ShouldBind(c, form, int64(config.GetHTTPServerMaxMultipartMemory())); err != nil { // nolint
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorBadRequest(err, err.Error()))
		return
	}

	if form.Domain == "" {
		const errText = "domain must be set"
		errWithCode := gtserror.NewErrorBadRequest(errors.New(errText), errText)
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	permType, errWithCode := parseDomainPermissionType(form.PermissionType)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	permDraft, errWithCode := m.processor.Admin().DomainPermissionDraftCreate(
		c,
		authed.Account,
		form.Domain,
		permType,
		util.PtrOrZero(form.Obfuscate),
		util.PtrOrZero(form.PublicComment),
		util.PtrOrZero(form.PrivateComment),
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.JSON(c, http.StatusOK, permDraft)
}
