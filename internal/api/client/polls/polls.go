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

package polls

import (
	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gotosocial/internal/api/util"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"code.superseriousbusiness.org/gotosocial/internal/templates"
)

const (
	BasePath        = "/:" + util.APIVersionKey + "/polls"       // BasePath is the base API path for making poll requests through v1 or v2 of the api (for mastodon API compatibility)
	PollWithID      = BasePath + "/:" + apiutil.IDKey            //
	PollVotesWithID = BasePath + "/:" + apiutil.IDKey + "/votes" //
)

type Module struct {
	templates *templates.Templates
	processor *processing.Processor
}

func New(processor *processing.Processor, templates *templates.Templates) *Module {
	return &Module{
		templates: templates,
		processor: processor,
	}
}

func (m *Module) Route(g *httputil.RouteGroup) {
	g.GET(PollWithID, m.PollGETHandler)
	g.POST(PollVotesWithID, m.PollVotePOSTHandler)
}
