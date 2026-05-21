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

package statuses

import (
	"errors"
	"net/http"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"github.com/gin-gonic/gin"
)

// StatusesGETHandler swagger:operation GET /api/v1/statuses statusesGet
//
// View multiple statuses with the given IDs.
//
//	---
//	tags:
//	- statuses
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: id[]
//		type: array
//		items:
//			type: string
//		description: Target status IDs.
//		in: query
//		collectionFormat: multi
//		required: true
//
//	security:
//	- OAuth2 Bearer:
//		- read:statuses
//
//	responses:
//		'200':
//			description: "The requested statuses."
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/status"
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
//		'422':
//			schema:
//				"$ref": "#/definitions/error"
//			description: unprocessable entity
//		'500':
//			schema:
//				"$ref": "#/definitions/error"
//			description: internal server error
func (m *Module) StatusesGETHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeReadStatuses,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	ids := c.QueryArray("id[]")
	if len(ids) == 0 {
		ids = c.QueryArray("id")
	}

	if len(ids) == 0 {
		const text = "no status ids specified"
		errWithCode := gtserror.NewErrorBadRequest(errors.New(text), text)
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	} else if len(ids) > 100 {
		const text = "too many status ids specified"
		errWithCode := gtserror.NewErrorUnprocessableEntity(errors.New(text), text)
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiStatuses, errWithCode := m.processor.Status().GetMultiple(c.Request.Context(), authed.Account, ids)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, apiStatuses)
}
