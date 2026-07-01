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
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

// DomainLimitsPOSTHandler swagger:operation POST /api/v1/admin/domain_limits domainLimitCreate
//
// Create a domain limit.
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
//		description: Hostname of the domain to limit.
//		type: string
//		required: true
//	-
//		name: media_policy
//		in: formData
//		description: |-
//			Policy to apply to media files originating from the limited domain.
//			No action = default (not limited).
//			Mark sensitive = mark all media from the limited domain as sensitive.
//			Reject = do not download media from the limited domain. Serve a link to the media instead.
//		type: string
//		enum:
//			- no_action
//			- mark_sensitive
//			- reject
//		default: no_action
//	-
//		name: follows_policy
//		in: formData
//		description: |-
//			Policy to apply to follow (requests) originating from the limited domain.
//			No action = default (not limited).
//			Manual approval = require manual approval for all follows from limited domain.
//			Reject non mutual = automatically reject follows from the limited domain when they're not follow-backs.
//			Reject all = automatically reject all follows from the limited domain.
//		type: string
//		enum:
//			- no_action
//			- manual_approval
//			- reject_non_mutual
//			- reject_all
//		default: no_action
//	-
//		name: statuses_policy
//		in: formData
//		description: |-
//			Policy to apply to statuses of non-followed accounts on the limited domain.
//			No action = default (not limited).
//			Filter warn = trigger a warn filter pointing to this domain limit.
//			Filter hide = trigger a hide filter pointing to this domain limit.
//		type: string
//		enum:
//			- no_action
//			- filter_warn
//			- filter_hide
//		default: no_action
//	-
//		name: accounts_policy
//		in: formData
//		description: |-
//			Policy to apply to non-followed accounts on the limited domain.
//			No action = default (not limited).
//			Mute = mute all non-followed accounts on the limited domain.
//		type: string
//		enum:
//			- no_action
//			- mute
//		default: no_action
//	-
//		name: content_warning
//		in: formData
//		description: Content warning to prepend to posts from accounts on this instance.
//		type: string
//	-
//		name: public_comment
//		in: formData
//		description: >-
//			Public comment about this domain limit.
//			This will be displayed alongside the domain limit if you choose to share limits.
//		type: string
//	-
//		name: private_comment
//		in: formData
//		description: >-
//			Private comment about this domain limit. Will only be shown to other admins, so this
//			is a useful way of internally keeping track of why a certain domain ended up limited.
//		type: string
//
//	security:
//	- OAuth2 Bearer:
//		- admin:write:domain_limits
//
//	responses:
//		'200':
//			description: The newly created domain limit.
//			schema:
//				"$ref": "#/definitions/domainLimit"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'403':
//			description: forbidden
//		'404':
//			description: not found
//		'406':
//			description: not acceptable
//		'409':
//			description: There is already a limit in place for this domain.
//		'500':
//			description: internal server error
func (m *Module) DomainLimitsPOSTHandler(c *httputil.Context) {
	authed, errWithCode := apiutil.TokenAuth(c, apiutil.AuthRequirements{
		Token:   true,
		App:     true,
		User:    true,
		Account: true,
		Scope:   []apiutil.Scope{apiutil.ScopeAdminWriteDomainLimits},
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

	form := new(apimodel.DomainLimitRequest)
	if err := httputil.ShouldBind(c, form, int64(config.GetHTTPServerMaxMultipartMemory())); err != nil { // nolint
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorBadRequest(err, err.Error()))
		return
	}

	if form.Domain == "" {
		const errText = "domain must be set"
		errWithCode := gtserror.NewErrorBadRequest(errors.New(errText), errText)
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	domainLimit, errWithCode := m.processor.Admin().DomainLimitCreate(
		c,
		authed.Account,
		form.Domain,
		util.PtrOrValue(form.MediaPolicy, apimodel.MediaPolicyNoAction),
		util.PtrOrValue(form.FollowsPolicy, apimodel.FollowsPolicyNoAction),
		util.PtrOrValue(form.StatusesPolicy, apimodel.StatusesPolicyNoAction),
		util.PtrOrValue(form.AccountsPolicy, apimodel.AccountsPolicyNoAction),
		util.PtrOrZero(form.ContentWarning),
		util.PtrOrZero(form.PublicComment),
		util.PtrOrZero(form.PrivateComment),
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.JSON(c, http.StatusOK, domainLimit)
}
