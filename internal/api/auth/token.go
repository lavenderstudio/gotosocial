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

package auth

import (
	"net/http"
	"net/url"

	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gopkg/httputil/binding"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
)

type tokenRequestForm struct {
	GrantType    *string `form:"grant_type" json:"grant_type" xml:"grant_type"`
	Code         *string `form:"code" json:"code" xml:"code"`
	RedirectURI  *string `form:"redirect_uri" json:"redirect_uri" xml:"redirect_uri"`
	ClientID     *string `form:"client_id" json:"client_id" xml:"client_id"`
	ClientSecret *string `form:"client_secret" json:"client_secret" xml:"client_secret"`
	Scope        *string `form:"scope" json:"scope" xml:"scope"`
}

// TokenPOSTHandler should be served as a POST at https://example.org/oauth/token
// The idea here is to serve an oauth access token to a user, which can be used for authorizing against non-public APIs.
func (m *Module) TokenPOSTHandler(c *httputil.Context) {
	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	help := []string{}

	form := &tokenRequestForm{}
	if err := binding.ShouldBind(c, form, int64(config.GetHTTPServerMaxMultipartMemory())); err != nil { // nolint
		apiutil.OAuthErrorHandler(c, gtserror.NewErrorBadRequest(oauth.ErrInvalidRequest, err.Error()))
		return
	}

	c.R.Form = url.Values{}

	var grantType string
	if form.GrantType != nil {
		grantType = *form.GrantType
		c.R.Form.Set("grant_type", grantType)
	} else {
		help = append(help, "grant_type was not set in the token request form, but must be set to authorization_code or client_credentials")
	}

	if form.ClientID != nil {
		c.R.Form.Set("client_id", *form.ClientID)
	} else {
		help = append(help, "client_id was not set in the token request form")
	}

	if form.ClientSecret != nil {
		c.R.Form.Set("client_secret", *form.ClientSecret)
	} else {
		help = append(help, "client_secret was not set in the token request form")
	}

	if form.RedirectURI != nil {
		c.R.Form.Set("redirect_uri", *form.RedirectURI)
	} else {
		help = append(help, "redirect_uri was not set in the token request form")
	}

	var code string
	if form.Code != nil {
		if grantType != "authorization_code" {
			help = append(help, "a code was provided in the token request form, but grant_type was not set to authorization_code")
		} else {
			code = *form.Code
			c.R.Form.Set("code", code)
		}
	} else if grantType == "authorization_code" {
		help = append(help, "code was not set in the token request form, but must be set since grant_type is authorization_code")
	}

	if form.Scope != nil {
		c.R.Form.Set("scope", *form.Scope)
	}

	if len(help) != 0 {
		apiutil.OAuthErrorHandler(c, gtserror.NewErrorBadRequest(oauth.ErrInvalidRequest, help...))
		return
	}

	token, errWithCode := m.processor.OAuthHandleTokenRequest(c.R)
	if errWithCode != nil {
		apiutil.OAuthErrorHandler(c, errWithCode)
		return
	}

	c.W.Header().Set("Cache-Control", "no-store")
	c.W.Header().Set("Pragma", "no-cache")
	httputil.EncodeJSONResponse(c,
		http.StatusOK,
		apiutil.AppJSON,
		token,
	)
}
