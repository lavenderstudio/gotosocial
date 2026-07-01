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

package api

import (
	"time"

	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/accounts"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/admin"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/announcements"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/apps"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/blocks"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/bookmarks"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/conversations"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/customemojis"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/debug"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/directory"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/exports"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/favourites"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/featuredtags"
	filtersV1 "code.superseriousbusiness.org/gotosocial/internal/api/client/filters/v1"
	filtersV2 "code.superseriousbusiness.org/gotosocial/internal/api/client/filters/v2"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/followedtags"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/followrequests"
	importdata "code.superseriousbusiness.org/gotosocial/internal/api/client/import"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/instance"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/interactionpolicies"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/interactionrequests"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/lists"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/markers"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/media"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/mutes"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/notifications"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/polls"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/preferences"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/push"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/relaypushes"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/reports"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/scheduledstatuses"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/search"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/statuses"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/streaming"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/suggestions"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/tags"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/timelines"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/tokens"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/trends"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/user"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/middleware"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"code.superseriousbusiness.org/gotosocial/internal/router"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/templates"
)

type Client struct {
	processor *processing.Processor
	db        db.DB

	accounts            *accounts.Module            // api/v1/accounts, api/v1/profile
	admin               *admin.Module               // api/v1/admin
	announcements       *announcements.Module       // api/v1/announcements
	apps                *apps.Module                // api/v1/apps
	blocks              *blocks.Module              // api/v1/blocks
	bookmarks           *bookmarks.Module           // api/v1/bookmarks
	conversations       *conversations.Module       // api/v1/conversations
	customEmojis        *customemojis.Module        // api/v1/custom_emojis
	debug               *debug.Module               // api/v1/debug
	directory           *directory.Module           // api/v1/directory
	exports             *exports.Module             // api/v1/exports
	favourites          *favourites.Module          // api/v1/favourites
	featuredTags        *featuredtags.Module        // api/v1/featured_tags
	filtersV1           *filtersV1.Module           // api/v1/filters
	filtersV2           *filtersV2.Module           // api/v2/filters
	followRequests      *followrequests.Module      // api/v1/follow_requests
	followedTags        *followedtags.Module        // api/v1/followed_tags
	importData          *importdata.Module          // api/v1/import
	instance            *instance.Module            // api/v1/instance
	interactionPolicies *interactionpolicies.Module // api/v1/interaction_policies
	interactionRequests *interactionrequests.Module // api/v1/interaction_requests
	lists               *lists.Module               // api/v1/lists
	markers             *markers.Module             // api/v1/markers
	media               *media.Module               // api/v1/media, api/v2/media
	mutes               *mutes.Module               // api/v1/mutes
	notifications       *notifications.Module       // api/v1/notifications
	polls               *polls.Module               // api/v1/polls
	preferences         *preferences.Module         // api/v1/preferences
	push                *push.Module                // api/v1/push
	relayPushes         *relaypushes.Module         // api/v1/relay_pushes
	reports             *reports.Module             // api/v1/reports
	scheduledStatuses   *scheduledstatuses.Module   // api/v1/scheduled_statuses
	search              *search.Module              // api/v1/search, api/v2/search
	statuses            *statuses.Module            // api/v1/statuses
	streaming           *streaming.Module           // api/v1/streaming
	suggestions         *suggestions.Module         // api/v2/suggestions
	tags                *tags.Module                // api/v1/tags
	timelines           *timelines.Module           // api/v1/timelines
	tokens              *tokens.Module              // api/v1/tokens
	trends              *trends.Module              // api/v1/trends
	user                *user.Module                // api/v1/user
}

func (c *Client) Route(r *router.Router, m ...httputil.Middleware) {
	// create a new group on the
	// top level client 'api' prefix
	apiGroup := r.Group("api")

	// attach non-global middlewares
	// appropriate to the client api
	apiGroup.Use(m...)
	apiGroup.Use(
		middleware.TokenCheck(c.db, c.processor.OAuthValidateBearerToken),
		middleware.CacheControl(middleware.CacheControlConfig{
			// Never cache client api responses.
			Directives: []string{"no-store"},
		}),
	)

	c.accounts.Route(apiGroup)
	c.admin.Route(apiGroup)
	c.announcements.Route(apiGroup)
	c.apps.Route(apiGroup)
	c.blocks.Route(apiGroup)
	c.bookmarks.Route(apiGroup)
	c.conversations.Route(apiGroup)
	c.customEmojis.Route(apiGroup)
	c.debug.Route(apiGroup)
	c.directory.Route(apiGroup)
	c.exports.Route(apiGroup)
	c.favourites.Route(apiGroup)
	c.featuredTags.Route(apiGroup)
	c.filtersV1.Route(apiGroup)
	c.filtersV2.Route(apiGroup)
	c.followRequests.Route(apiGroup)
	c.followedTags.Route(apiGroup)
	c.importData.Route(apiGroup)
	c.instance.Route(apiGroup)
	c.interactionPolicies.Route(apiGroup)
	c.interactionRequests.Route(apiGroup)
	c.lists.Route(apiGroup)
	c.markers.Route(apiGroup)
	c.media.Route(apiGroup)
	c.mutes.Route(apiGroup)
	c.notifications.Route(apiGroup)
	c.polls.Route(apiGroup)
	c.preferences.Route(apiGroup)
	c.push.Route(apiGroup)
	c.relayPushes.Route(apiGroup)
	c.reports.Route(apiGroup)
	c.scheduledStatuses.Route(apiGroup)
	c.search.Route(apiGroup)
	c.statuses.Route(apiGroup)
	c.streaming.Route(apiGroup)
	c.suggestions.Route(apiGroup)
	c.tags.Route(apiGroup)
	c.timelines.Route(apiGroup)
	c.tokens.Route(apiGroup)
	c.trends.Route(apiGroup)
	c.user.Route(apiGroup)
}

func NewClient(state *state.State, process *processing.Processor, templates *templates.Templates) *Client {
	return &Client{
		processor: process,
		db:        state.DB,

		accounts:            accounts.New(process, templates),
		admin:               admin.New(state, process, templates),
		announcements:       announcements.New(process, templates),
		apps:                apps.New(process, templates),
		blocks:              blocks.New(process, templates),
		bookmarks:           bookmarks.New(process, templates),
		conversations:       conversations.New(process, templates),
		customEmojis:        customemojis.New(process, templates),
		debug:               debug.New(state, process, templates),
		directory:           directory.New(process, templates),
		exports:             exports.New(process, templates),
		favourites:          favourites.New(process, templates),
		featuredTags:        featuredtags.New(process, templates),
		filtersV1:           filtersV1.New(process, templates),
		filtersV2:           filtersV2.New(process, templates),
		followRequests:      followrequests.New(process, templates),
		followedTags:        followedtags.New(process, templates),
		importData:          importdata.New(process, templates),
		instance:            instance.New(process, templates),
		interactionPolicies: interactionpolicies.New(process, templates),
		interactionRequests: interactionrequests.New(process, templates),
		lists:               lists.New(process, templates),
		markers:             markers.New(process, templates),
		media:               media.New(process, templates),
		mutes:               mutes.New(process, templates),
		notifications:       notifications.New(process, templates),
		polls:               polls.New(process, templates),
		preferences:         preferences.New(process, templates),
		push:                push.New(process, templates),
		relayPushes:         relaypushes.New(process, templates),
		reports:             reports.New(process, templates),
		scheduledStatuses:   scheduledstatuses.New(process, templates),
		search:              search.New(process, templates),
		statuses:            statuses.New(process, templates),
		streaming:           streaming.New(process, templates, time.Second*30),
		suggestions:         suggestions.New(process, templates),
		tags:                tags.New(process, templates),
		timelines:           timelines.New(process, templates),
		tokens:              tokens.New(process, templates),
		trends:              trends.New(process, templates),
		user:                user.New(process, templates),
	}
}
