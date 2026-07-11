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
	"code.superseriousbusiness.org/gopkg/httputil/binding"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

// AccountMutePOSTHandler swagger:operation POST /api/v1/accounts/{id}/mute accountMute
//
// Mute account by ID.
//
// If account was already muted, succeeds anyway. This can be used to update the details of a mute.
//
//	---
//	tags:
//	- accounts
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: id
//		type: string
//		description: The ID of the account to block.
//		in: path
//		required: true
//	-
//		name: notifications
//		type: boolean
//		description: Mute notifications as well as posts.
//		in: formData
//		required: false
//		default: false
//	-
//		name: duration
//		type: number
//		description: How long the mute should last, in seconds. If 0 or not provided, mute lasts indefinitely.
//		in: formData
//		required: false
//		default: 0
//
//	security:
//	- OAuth2 Bearer:
//		- write:mutes
//
//	responses:
//		'200':
//			description: Your relationship to the account.
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
//		'500':
//			schema:
//				"$ref": "#/definitions/error"
//			description: internal server error
func (m *Module) AccountMutePOSTHandler(c *httputil.Context) {
	authed, errWithCode := apiutil.TokenAuth(c, apiutil.AuthRequirements{
		Token:   true,
		App:     true,
		User:    true,
		Account: true,
		Scope:   []apiutil.Scope{apiutil.ScopeWriteMutes},
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

	form := &apimodel.UserMuteCreateUpdateRequest{}
	if err := binding.ShouldBind(c, form, int64(config.GetHTTPServerMaxMultipartMemory())); err != nil { // nolint
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorBadRequest(err, err.Error()))
		return
	}

	if err := normalizeCreateUpdateMute(form); err != nil {
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorUnprocessableEntity(err, err.Error()))
		return
	}

	relationship, errWithCode := m.processor.Account().MuteCreate(
		c,
		authed.Account,
		targetAcctID,
		form,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.JSON(c, http.StatusOK, relationship)
}

func normalizeCreateUpdateMute(form *apimodel.UserMuteCreateUpdateRequest) error {
	// Apply defaults for missing fields.
	form.Notifications = util.Ptr(util.PtrOrValue(form.Notifications, false))

	// Normalize duration if necessary.
	if form.DurationI != nil {
		// If we parsed this as JSON, duration
		// may be either a float64 or a string.
		duration, err := apiutil.ParseDuration(form.DurationI, "duration")
		if err != nil {
			return err
		}
		form.Duration = duration
	}

	// Interpret zero as indefinite duration.
	if form.Duration != nil && *form.Duration == 0 {
		form.Duration = nil
	}

	return nil
}
