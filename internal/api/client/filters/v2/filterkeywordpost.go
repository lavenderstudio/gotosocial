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

package v2

import (
	"net/http"

	"code.superseriousbusiness.org/gopkg/httputil"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"code.superseriousbusiness.org/gotosocial/internal/validate"
)

// FilterKeywordPOSTHandler swagger:operation POST /api/v2/filters/{id}/keywords filterKeywordPost
//
// Add a filter keyword to an existing filter.
//
//	---
//	tags:
//	- filters
//
//	consumes:
//	- application/json
//	- application/xml
//	- application/x-www-form-urlencoded
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: id
//		in: path
//		type: string
//		required: true
//		description: ID of the filter to add the filtered status to.
//	-
//		name: keyword
//		in: formData
//		required: true
//		description: |-
//			The text to be filtered
//
//			Sample: fnord
//		type: string
//		minLength: 1
//		maxLength: 40
//	-
//		name: whole_word
//		in: formData
//		description: |-
//			Should the filter consider word boundaries?
//
//			Sample: true
//		type: boolean
//		default: false
//
//	security:
//	- OAuth2 Bearer:
//		- write:filters
//
//	responses:
//		'200':
//			name: filterKeyword
//			description: New filter keyword.
//			schema:
//				"$ref": "#/definitions/filterKeyword"
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
//		'409':
//			schema:
//				"$ref": "#/definitions/error"
//			description: conflict (duplicate keyword)
//		'422':
//			schema:
//				"$ref": "#/definitions/error"
//			description: unprocessable content
//		'500':
//			schema:
//				"$ref": "#/definitions/error"
//			description: internal server error
func (m *Module) FilterKeywordPOSTHandler(c *httputil.Context) {
	authed, errWithCode := apiutil.TokenAuth(c, apiutil.AuthRequirements{
		Token:   true,
		App:     true,
		User:    true,
		Account: true,
		Scope:   []apiutil.Scope{apiutil.ScopeWriteFilters},
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

	filterID, errWithCode := apiutil.ParseID(c.PathValue(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	form := &apimodel.FilterKeywordCreateUpdateRequest{}
	if err := httputil.ShouldBind(c, form, int64(config.GetHTTPServerMaxMultipartMemory())); err != nil { // nolint
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorBadRequest(err, err.Error()))
		return
	}

	if err := validateNormalizeCreateUpdateFilterKeyword(form); err != nil {
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorUnprocessableEntity(err, err.Error()))
		return
	}

	apiFilter, errWithCode := m.processor.FiltersV2().KeywordCreate(c, authed.Account, filterID, form)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.JSON(c, http.StatusOK, apiFilter)
}

func validateNormalizeCreateUpdateFilterKeyword(form *apimodel.FilterKeywordCreateUpdateRequest) error {
	if err := validate.FilterKeyword(form.Keyword); err != nil {
		return err
	}

	form.WholeWord = util.Ptr(util.PtrOrValue(form.WholeWord, false))

	return nil
}
