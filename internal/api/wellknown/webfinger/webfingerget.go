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

package webfinger

import (
	"errors"
	"fmt"
	"net/http"

	"code.superseriousbusiness.org/gopkg/httputil"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

// WebfingerGETRequest swagger:operation GET /.well-known/webfinger webfingerGet
//
// Handles webfinger account lookup requests.
//
// For example, a GET to `https://goblin.technology/.well-known/webfinger?resource=acct:tobi@goblin.technology` would return:
//
// ```
//
//	{"subject":"acct:tobi@goblin.technology","aliases":["https://goblin.technology/users/tobi","https://goblin.technology/@tobi"],"links":[{"rel":"http://webfinger.net/rel/profile-page","type":"text/html","href":"https://goblin.technology/@tobi"},{"rel":"self","type":"application/activity+json","href":"https://goblin.technology/users/tobi"}]}
//
// ```
//
// See: https://webfinger.net/
//
//	---
//	tags:
//	- .well-known
//
//	produces:
//	- application/jrd+json
//
//	responses:
//		'200':
//			schema:
//				"$ref": "#/definitions/wellKnownResponse"
func (m *Module) WebfingerGETRequest(c *httputil.Context) {
	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.WebfingerJSONAcceptHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	resourceQuery := c.Query("resource")
	if resourceQuery == "" {
		err := errors.New("no 'resource' in request query")
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorBadRequest(err, err.Error()))
		return
	}

	requestedUser, requestedHost, err := util.ExtractWebfingerParts(resourceQuery)
	if err != nil {
		err := fmt.Errorf("bad webfinger request with resource query %s: %w", resourceQuery, err)
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorBadRequest(err, err.Error()))
		return
	}

	if requestedHost != config.GetHost() && requestedHost != config.GetAccountDomain() {
		err := fmt.Errorf("requested host %s does not belong to this instance", requestedHost)
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorBadRequest(err, err.Error()))
		return
	}

	resp, errWithCode := m.processor.Fedi().WebfingerGet(c, requestedUser)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	// Encode JSON response.
	httputil.JSONType(c,
		http.StatusOK,
		apiutil.AppJRDJSON,
		resp,
	)
}
