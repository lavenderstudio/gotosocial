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
	"net/url"

	"code.superseriousbusiness.org/gopkg/httputil"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/templates"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
)

// authorizeInteractionGETHandler handles redirects from remote
// (usually Mastodon) instances when a user tries to do a
// "remote interaction" and gives their GoToSocial account/domain.
// We use this handler instead of serving a generic 404 page.
func (m *Module) authorizeInteractionGETHandler(c *httputil.Context) {
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

	// Redirects to the "authorize_interaction"
	// endpoint should contain the URI of the
	// object that the user is trying to interact
	// with in the 'uri' query param.
	uriStr := c.Query("uri")
	if uriStr == "" {
		const text = "no uri query parameter found in string"
		errWithCode := gtserror.NewWithCode(http.StatusNotFound, text)
		apiutil.WebErrorHandler(c, m.templates, errWithCode)
	}

	// Try to parse the object URI.
	interactionURI, err := url.Parse(uriStr)
	if err != nil {
		err := gtserror.Newf("interaction URI could not be parsed: %w", err)
		errWithCode := gtserror.NewErrorBadRequest(err, err.Error())
		apiutil.WebErrorHandler(c, m.templates, errWithCode)
	}

	// Pass to template renderer.
	m.templates.RenderPage(c,
		http.StatusOK,
		templates.WebPage{
			Template:    "authorize-interaction.tmpl",
			Stylesheets: []string{cssAbout},
			Javascript: []templates.JavascriptEntry{
				{
					Src:   jsFrontend,
					Async: true,
					Defer: true,
				},
			},
			Extra: map[string]any{
				"interactionURI": interactionURI,
				"instance":       instance,
				"ogMeta":         typeutils.OpenGraphBase(instance),
			},
		},
	)
}
