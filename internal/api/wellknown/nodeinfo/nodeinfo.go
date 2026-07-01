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

package nodeinfo

import (
	"net/http"

	"code.superseriousbusiness.org/gopkg/httputil"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/middleware"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"code.superseriousbusiness.org/gotosocial/internal/templates"
)

const (
	// NodeInfoWellKnownPath is the base path for serving responses
	// to nodeinfo lookup requests, minus the '.well-known' prefix.
	NodeInfoWellKnownPath = "/nodeinfo"
)

type Module struct {
	templates *templates.Templates
	processor *processing.Processor
}

// New returns a new nodeinfo module
func New(processor *processing.Processor, templates *templates.Templates) *Module {
	return &Module{
		templates: templates,
		processor: processor,
	}
}

func (m *Module) Route(g *httputil.RouteGroup) {
	// If instance is configured to serve instance stats
	// faithfully at nodeinfo, we should allow robots to
	// crawl nodeinfo endpoints in a limited capacity.
	// In all other cases, disallow everything.
	var robots httputil.FlatMiddlewareFunc
	if config.GetInstanceStatsMode() == config.InstanceStatsModeServe {
		robots = middleware.RobotsHeaders(middleware.RobotsHeadersModeAllowSome)
	} else {
		robots = middleware.RobotsHeaders(middleware.RobotsHeadersModeDefault)
	}

	// Attach handler, injecting
	// robots http header middleware.
	g.Use(robots)
	g.GET(NodeInfoWellKnownPath, m.NodeInfoWellKnownGETHandler)
}

// NodeInfoWellKnownGETHandler swagger:operation GET /.well-known/nodeinfo nodeInfoWellKnownGet
//
// Returns a well-known response which redirects callers to `/nodeinfo/2.0`.
//
// eg. `{"links":[{"rel":"http://nodeinfo.diaspora.software/ns/schema/2.0","href":"http://example.org/nodeinfo/2.0"}]}`
// See: https://nodeinfo.diaspora.software/protocol.html
//
//	---
//	tags:
//	- .well-known
//
//	produces:
//	- application/json
//
//	responses:
//		'200':
//			schema:
//				"$ref": "#/definitions/wellKnownResponse"
func (m *Module) NodeInfoWellKnownGETHandler(c *httputil.Context) {
	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	resp, errWithCode := m.processor.Fedi().NodeInfoRelGet(c)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	// Encode JSON response.
	httputil.JSONType(c,
		http.StatusOK,
		apiutil.AppJSON,
		resp,
	)
}
