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
	"code.superseriousbusiness.org/gotosocial/internal/templates"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
)

const (
	loginPath = "/login"
)

func (m *Module) loginGETHandler(c *httputil.Context) {
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
			Template:    "login-info.tmpl",
			Stylesheets: []string{cssAbout, cssLoginInfo},
			Extra: map[string]any{
				"instance": instance,
				"ogMeta":   typeutils.OpenGraphBase(instance),
			},
		},
	)
}
