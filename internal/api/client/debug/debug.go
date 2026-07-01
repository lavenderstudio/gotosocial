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

package debug

import (
	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/templates"
)

const (
	BasePath             = "/v1/debug"
	APUrlPath            = BasePath + "/apurl"
	ClearCachesPath      = BasePath + "/caches/clear"
	StatusVisibilityPath = BasePath + "/status/visibility"

	// endpoint clones to maintain
	// backwards compatibility with
	// previous gotosocial versions
	_CompatAPUrlPath       = "/v1/admin/debug/apurl"
	_CompatClearCachesPath = "/v1/admin/debug/caches/clear"
)

type Module struct {
	state     *state.State
	templates *templates.Templates
	processor *processing.Processor
}

func New(state *state.State, processor *processing.Processor, templates *templates.Templates) *Module {
	return &Module{
		state:     state,
		templates: templates,
		processor: processor,
	}
}

func (m *Module) Route(g *httputil.RouteGroup) {
	// activitypub debug endpoints.
	g.GET(APUrlPath, m.APUrlGETHandler)

	// cache debug endpoints.
	g.POST(ClearCachesPath, m.ClearCachesPOSTHandler)

	// status debug endpoints.
	g.GET(StatusVisibilityPath, m.StatusVisibilityGETHandler)

	// backwards compatibility endpoints
	g.GET(_CompatAPUrlPath, m.APUrlGETHandler)
	g.POST(_CompatClearCachesPath, m.ClearCachesPOSTHandler)
}
