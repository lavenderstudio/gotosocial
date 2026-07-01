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

package web

import (
	"net/http"

	"code.superseriousbusiness.org/gopkg/httputil"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
)

const textCSSUTF8 = string(apiutil.TextCSS + "; charset=utf-8")

func (m *Module) customCSSGETHandler(c *httputil.Context) {
	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.TextCSS); errWithCode != nil {
		apiutil.WebErrorHandler(c, m.templates, errWithCode)
		return
	}

	requestedUser, errWithCode := apiutil.ParseUsername(c.PathValue(apiutil.UsernameKey))
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, m.templates, errWithCode)
		return
	}

	// Retrieve customCSS if enabled on the instance.
	// Else use an empty string, to help with caching
	// when custom CSS gets toggled on or off.
	var customCSS string
	if config.GetAccountsAllowCustomCSS() {
		customCSS, errWithCode = m.processor.Account().GetCustomCSSForUsername(c, requestedUser)
		if errWithCode != nil {
			apiutil.WebErrorHandler(c, m.templates, errWithCode)
			return
		}
	}

	c.W.Header().Set(cacheControlHeader, cacheControlNoCache)
	httputil.String(c, http.StatusOK, textCSSUTF8, customCSS)
}

func (m *Module) instanceCustomCSSGETHandler(c *httputil.Context) {
	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.TextCSS); errWithCode != nil {
		apiutil.WebErrorHandler(c, m.templates, errWithCode)
		return
	}

	instanceV1, errWithCode := m.processor.InstanceGetV1(c)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, m.templates, errWithCode)
		return
	}

	instanceCustomCSS := instanceV1.CustomCSS

	c.W.Header().Set(cacheControlHeader, cacheControlNoCache)
	httputil.String(c, http.StatusOK, textCSSUTF8, instanceCustomCSS)
}
