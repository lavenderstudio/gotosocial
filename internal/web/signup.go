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
	"errors"
	"net/http"

	"code.superseriousbusiness.org/gopkg/httputil"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/templates"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
	"code.superseriousbusiness.org/gotosocial/internal/validate"
)

func (m *Module) signupGETHandler(c *httputil.Context) {
	// We'll need the instance later, and we can also use it
	// before then to make it easier to return a web error.
	instance, errWithCode := m.processor.InstanceGetV1(c)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, m.templates, errWithCode)
		return
	}

	// We only serve text/html at this endpoint.
	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.TextHTML); errWithCode != nil {
		apiutil.WebErrorHandler(c, m.templates, errWithCode)
		return
	}

	// Pass to template renderer.
	m.templates.RenderPage(c,
		http.StatusOK,
		templates.WebPage{
			Template: "sign-up.tmpl",
			Extra: map[string]any{
				"oidcEnabled":      config.GetOIDCEnabled(),
				"registrationOpen": config.GetAccountsRegistrationOpen(),
				"reasonRequired":   config.GetAccountsReasonRequired(),
				"instance":         instance,
				"ogMeta":           typeutils.OpenGraphBase(instance),
			},
		},
	)
}

func (m *Module) signupPOSTHandler(c *httputil.Context) {
	// We'll need the instance later, and we can also use it
	// before then to make it easier to return a web error.
	instance, errWithCode := m.processor.InstanceGetV1(c)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, m.templates, errWithCode)
		return
	}

	// We only serve text/html at this endpoint.
	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.TextHTML); errWithCode != nil {
		apiutil.WebErrorHandler(c, m.templates, errWithCode)
		return
	}

	form := &apimodel.AccountCreateRequest{}
	if err := httputil.ShouldBind(c, form, int64(config.GetHTTPServerMaxMultipartMemory())); err != nil { // nolint
		apiutil.WebErrorHandler(c, m.templates, gtserror.NewErrorBadRequest(err, err.Error()))
		return
	}

	if err := validate.CreateAccount(form); err != nil {
		apiutil.WebErrorHandler(c, m.templates, gtserror.NewErrorBadRequest(err, err.Error()))
		return
	}

	clientIP := gtscontext.ClientIP(c)
	if clientIP == nil {
		err := errors.New("ip address could not be parsed from request")
		apiutil.WebErrorHandler(c, m.templates, gtserror.NewErrorBadRequest(err, err.Error()))
		return
	}

	form.IP = clientIP.AsSlice()

	// We have all the info we need, call user+account create
	// (this will also trigger side effects like sending emails etc).
	user, errWithCode := m.processor.User().Create(
		c,
		// nil to use
		// instance app.
		nil,
		form,
	)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, m.templates, errWithCode)
		return
	}

	// Serve a page informing the
	// user that they've signed up.
	// Pass to template renderer.
	m.templates.RenderPage(c,
		http.StatusOK,
		templates.WebPage{
			Template: "signed-up.tmpl",
			Extra: map[string]any{
				"email":    user.UnconfirmedEmail,
				"username": user.Account.Username,
				"instance": instance,
				"ogMeta":   typeutils.OpenGraphBase(instance),
			},
		},
	)
}
