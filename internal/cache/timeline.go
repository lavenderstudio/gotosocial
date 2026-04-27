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

package cache

import (
	"code.superseriousbusiness.org/gopkg/log"
	"code.superseriousbusiness.org/gotosocial/internal/cache/timeline"
	"code.superseriousbusiness.org/gotosocial/internal/config"
)

type TimelineCaches struct {
	// Public provides an instance-level
	// cache of the public status timeline.
	Public timeline.StatusTimeline

	// Local provides an instance-level
	// cache of the local status timeline.
	Local timeline.StatusTimeline

	// Home provides a concurrency-safe map of status timeline
	// caches for home timelines, keyed by home's account ID.
	Home timeline.StatusTimelines

	// List provides a concurrency-safe map of status
	// timeline caches for lists, keyed by list ID.
	List timeline.StatusTimelines

	// Tag provides a concurrency-safe map of status
	// timeline caches for tags, keyed by tag ID.
	Tag timeline.StatusTimelines
}

func (c *Caches) initPublicTimeline() {
	// TODO: configurable
	cap := 800

	log.Infof(nil, "cache size = %d", cap)

	c.Timelines.Public.Init(cap)
}

func (c *Caches) initLocalTimeline() {
	// TODO: configurable
	cap := 800

	log.Infof(nil, "cache size = %d", cap)

	c.Timelines.Local.Init(cap)
}

func (c *Caches) initHomeTimelines() {
	cap := config.GetCacheHomeTimelineSize()
	cap = max(100, cap) // clamp to min=100

	timeout := config.GetCacheHomeTimelineTimeout()
	log.Infof(nil, "cache size = %d, timeout = %s", cap, timeout)

	c.Timelines.Home.Init(int(cap), timeout)
}

func (c *Caches) initListTimelines() {
	cap := config.GetCacheListTimelineSize()
	cap = max(100, cap) // clamp to min=100

	timeout := config.GetCacheListTimelineTimeout()
	log.Infof(nil, "cache size = %d, timeout = %s", cap, timeout)

	c.Timelines.List.Init(int(cap), timeout)
}

func (c *Caches) initTagTimelines() {
	cap := config.GetCacheTagTimelineSize()
	cap = max(50, cap) // clamp to min=50

	timeout := config.GetCacheTagTimelineTimeout()
	log.Infof(nil, "cache size = %d, timeout = %s", cap, timeout)

	c.Timelines.Tag.Init(int(cap), timeout)
}
