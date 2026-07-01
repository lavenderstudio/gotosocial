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
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/templates"
)

func (m *Module) confirmEmailGETHandler(c *httputil.Context) {
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

	// If there's no token in the query,
	// just serve the 404 web handler.
	token := c.Query("token")
	if token == "" {
		errWithCode := gtserror.NewErrorNotFound(errors.New(http.StatusText(http.StatusNotFound)))
		apiutil.WebErrorHandler(c, m.templates, errWithCode)
		return
	}

	// Get user but don't confirm yet.
	user, errWithCode := m.processor.User().EmailGetUserForConfirmToken(c, token)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, m.templates, errWithCode)
		return
	}

	// They may have already confirmed before
	// and are visiting the link again for
	// whatever reason. This is fine, just make
	// sure we have an email address to show them.
	email := user.UnconfirmedEmail
	if email == "" {
		// Already confirmed, take
		// that address instead.
		email = user.Email
	}

	// Serve page where user can click button
	// to POST confirmation to same endpoint.
	// Pass to template renderer.
	m.templates.RenderPage(c,
		http.StatusOK,
		templates.WebPage{
			Template: "confirm-email.tmpl",
			Extra: map[string]any{
				"email":    email,
				"username": user.Account.Username,
				"token":    token,
				"instance": instance,
			},
		},
	)
}

func (m *Module) confirmEmailPOSTHandler(c *httputil.Context) {
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

	// If there's no token in the query,
	// just serve the 404 web handler.
	token := c.Query("token")
	if token == "" {
		errWithCode := gtserror.NewErrorNotFound(errors.New(http.StatusText(http.StatusNotFound)))
		apiutil.WebErrorHandler(c, m.templates, errWithCode)
		return
	}

	// Confirm email address for real this time.
	user, errWithCode := m.processor.User().EmailConfirm(c, token)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, m.templates, errWithCode)
		return
	}

	// Serve page informing user that their
	// email address is now confirmed.
	//
	// Pass to template renderer.
	m.templates.RenderPage(c,
		http.StatusOK,
		templates.WebPage{
			Template: "confirmed-email.tmpl",
			Extra: map[string]any{
				"email":    user.Email,
				"username": user.Account.Username,
				"token":    token,
				"approved": *user.Approved,
				"instance": instance,
			},
		},
	)
}
