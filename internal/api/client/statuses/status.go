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

package statuses

import (
	"code.superseriousbusiness.org/gopkg/httputil"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"code.superseriousbusiness.org/gotosocial/internal/templates"
)

const (
	// BasePath is the base path for serving the statuses API, minus the 'api' prefix
	BasePath = "/v1/statuses"
	// BasePathWithID is just the base path with the ID key in it.
	// Use this anywhere you need to know the ID of the status being queried.
	BasePathWithID = BasePath + "/:" + apiutil.IDKey

	// FavouritedPath is for seeing who's faved a given status
	FavouritedPath = BasePathWithID + "/favourited_by"
	// FavouritePath is for posting a fave on a status
	FavouritePath = BasePathWithID + "/favourite"
	// UnfavouritePath is for removing a fave from a status
	UnfavouritePath = BasePathWithID + "/unfavourite"

	// RebloggedPath is for seeing who's boosted a given status
	RebloggedPath = BasePathWithID + "/reblogged_by"
	// ReblogPath is for boosting/reblogging a given status
	ReblogPath = BasePathWithID + "/reblog"
	// UnreblogPath is for undoing a boost/reblog of a given status
	UnreblogPath = BasePathWithID + "/unreblog"

	// BookmarkPath is for creating a bookmark on a given status
	BookmarkPath = BasePathWithID + "/bookmark"
	// UnbookmarkPath is for removing a bookmark from a given status
	UnbookmarkPath = BasePathWithID + "/unbookmark"

	// MutePath is for muting a given status so that notifications will no longer be received about it.
	MutePath = BasePathWithID + "/mute"
	// UnmutePath is for undoing an existing mute
	UnmutePath = BasePathWithID + "/unmute"

	// PinPath is for pinning a status to an account profile so that it's the first thing people see
	PinPath = BasePathWithID + "/pin"
	// UnpinPath is for undoing a pin and returning a status to the ever-swirling drain of time and entropy
	UnpinPath = BasePathWithID + "/unpin"

	// ContextPath is used for fetching context of posts
	ContextPath = BasePathWithID + "/context"

	// HistoryPath is used for fetching history of posts.
	HistoryPath = BasePathWithID + "/history"

	// SourcePath is used for fetching source of a post.
	SourcePath = BasePathWithID + "/source"
)

type Module struct {
	templates *templates.Templates
	processor *processing.Processor
}

func New(processor *processing.Processor, templates *templates.Templates) *Module {
	return &Module{
		processor: processor,
	}
}

func (m *Module) Route(g *httputil.RouteGroup) {
	// create / get / edit / delete status
	g.POST(BasePath, m.StatusCreatePOSTHandler)
	g.GET(BasePathWithID, m.StatusGETHandler)
	g.GET(BasePath, m.StatusesGETHandler)
	g.PUT(BasePathWithID, m.StatusEditPUTHandler)
	g.DELETE(BasePathWithID, m.StatusDELETEHandler)

	// fave stuff
	g.POST(FavouritePath, m.StatusFavePOSTHandler)
	g.POST(UnfavouritePath, m.StatusUnfavePOSTHandler)
	g.GET(FavouritedPath, m.StatusFavedByGETHandler)

	// pin stuff
	g.POST(PinPath, m.StatusPinPOSTHandler)
	g.POST(UnpinPath, m.StatusUnpinPOSTHandler)

	// mute stuff
	g.POST(MutePath, m.StatusMutePOSTHandler)
	g.POST(UnmutePath, m.StatusUnmutePOSTHandler)

	// reblog stuff
	g.POST(ReblogPath, m.StatusBoostPOSTHandler)
	g.POST(UnreblogPath, m.StatusUnboostPOSTHandler)
	g.GET(RebloggedPath, m.StatusBoostedByGETHandler)
	g.POST(BookmarkPath, m.StatusBookmarkPOSTHandler)
	g.POST(UnbookmarkPath, m.StatusUnbookmarkPOSTHandler)

	// context / status thread
	g.GET(ContextPath, m.StatusContextGETHandler)

	// history/edit stuff
	g.GET(HistoryPath, m.StatusHistoryGETHandler)
	g.GET(SourcePath, m.StatusSourceGETHandler)
}
