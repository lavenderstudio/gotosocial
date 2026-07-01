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
	"strings"

	"code.superseriousbusiness.org/gopkg/httputil"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/templates"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
)

func (m *Module) indexHandler(c *httputil.Context) {
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

	// If a landingPageUser is set in the config, redirect to
	// that user's profile instead of rendering landing/index page.
	if landingPageUser := config.GetLandingPageUser(); landingPageUser != "" {
		httputil.Redirect(c, http.StatusFound, "/@"+strings.ToLower(landingPageUser))
		return
	}

	// If indexing is allowed, set robots
	// meta to more permissive setting.
	var robotsMeta string
	if config.GetInstanceRobotsAllowIndexing() {
		robotsMeta = apiutil.RobotsDirectivesAllowSome
	}

	// Pass to template renderer.
	m.templates.RenderPage(c,
		http.StatusOK,
		templates.WebPage{
			Template:    "index.tmpl",
			Stylesheets: []string{cssAbout, cssIndex},
			Extra: map[string]any{
				// Render "home to x
				// users [etc]" strap.
				"showStrap": true,

				// Show "log in" button
				// in top-right corner.
				"showLoginButton": true,

				// Allow limited indexing
				// or use empty string
				// for default restrictive.
				"robotsMeta": robotsMeta,

				"instance": instance,
				"ogMeta":   typeutils.OpenGraphBase(instance),
			},
		},
	)
}
