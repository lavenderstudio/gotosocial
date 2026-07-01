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
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/templates"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
)

const (
	domainBlocklistPath = aboutPath + "/domain_blocks"
	domainAllowlistPath = aboutPath + "/domain_allows"
)

func (m *Module) domainBlocklistGETHandler(c *httputil.Context) {
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

	if !config.GetInstanceExposeBlocklistWeb() {
		const errText = "this instance does not expose its blocklist via the web"
		errWithCode := gtserror.NewErrorUnauthorized(errors.New(errText), errText)
		apiutil.WebErrorHandler(c, m.templates, errWithCode)
		return
	}

	domainBlocks, errWithCode := m.processor.InstancePeersGet(
		c,
		true,  // Include blocked.
		false, // Don't include allowed.
		false, // Don't include open.
		false, // Don't flatten list.
		false, // Don't include severity.
	)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, m.templates, errWithCode)
		return
	}

	// Pass to template renderer.
	m.templates.RenderPage(c,
		http.StatusOK,
		templates.WebPage{
			Template:    "domain-blocklist.tmpl",
			Stylesheets: []string{cssFA},
			Extra: map[string]any{
				"blocklist": domainBlocks,
				"instance":  instance,
				"ogMeta":    typeutils.OpenGraphBase(instance),
			},
		},
	)
}

func (m *Module) domainAllowlistGETHandler(c *httputil.Context) {
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

	if !config.GetInstanceExposeAllowlistWeb() {
		const errText = "this instance does not expose its allowlist via the web"
		errWithCode := gtserror.NewErrorUnauthorized(errors.New(errText), errText)
		apiutil.WebErrorHandler(c, m.templates, errWithCode)
		return
	}

	domainAllows, errWithCode := m.processor.InstancePeersGet(
		c,
		false, // Don't include blocked.
		true,  // Include allowed.
		false, // Don't include open.
		false, // Don't flatten list.
		false, // Don't include severity.
	)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, m.templates, errWithCode)
		return
	}

	// Pass to template renderer.
	m.templates.RenderPage(c,
		http.StatusOK,
		templates.WebPage{
			Template:    "domain-allowlist.tmpl",
			Stylesheets: []string{cssFA},
			Extra: map[string]any{
				"allowlist": domainAllows,
				"instance":  instance,
				"ogMeta":    typeutils.OpenGraphBase(instance),
			},
		},
	)
}
