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

package followrequests

import (
	"code.superseriousbusiness.org/gopkg/httputil"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"code.superseriousbusiness.org/gotosocial/internal/templates"
)

const (
	// BasePath is the base path for serving the follow request API, minus the 'api' prefix
	BasePath = "/v1/follow_requests"
	// BasePathWithID is just the base path with the ID key in it.
	// Use this anywhere you need to know the ID of the account that owns the follow request being queried.
	BasePathWithID = BasePath + "/:" + apiutil.IDKey
	// AuthorizePath is used for authorizing follow requests
	AuthorizePath = BasePathWithID + "/authorize"
	// RejectPath is used for rejecting follow requests
	RejectPath = BasePathWithID + "/reject"
	// OutgoingPath is used for fetching the list of accounts you requested to follow.
	OutgoingPath = BasePath + "/outgoing"
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
	g.GET(BasePath, m.FollowRequestGETHandler)
	g.GET(OutgoingPath, m.OutgoingFollowRequestGETHandler)
	g.POST(AuthorizePath, m.FollowRequestAuthorizePOSTHandler)
	g.POST(RejectPath, m.FollowRequestRejectPOSTHandler)
}
