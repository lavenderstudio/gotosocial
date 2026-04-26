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

package model

// RelayConnection models a relay push or relay subscription targeting a relay actor.
//
// swagger:model relayConnection
type RelayConnection struct {
	// ID of this item.
	// example: 01KMQFY8C9P2049NN09R9CCMSR
	ID string `json:"id"`

	// The date when this relay connection was created (ISO 8601 Datetime).
	// example: 2021-07-30T09:20:25+00:00
	CreatedAt string `json:"created_at"`

	// ID of the account that created this relay connection.
	// Will only be set for relay subscriptions, not relay pushes.
	// example: 01KMQFRR8PDEVBH0PWKR23E2YB
	AccountID string `json:"account_id,omitempty"`

	// ActivityPub URI of the relay service actor.
	// example: https://relay.activitypub.ca/actor
	RelayActorURI string `json:"relay_actor_uri"`

	// Matchers that apply to this relay connection.
	Matchers []RelayMatcher `json:"matchers"`

	// True if this relay connection has been approved by the relay actor.
	Approved bool `json:"approved"`

	// Include public posts when relaying via this connection.
	Public bool `json:"public"`

	// Include unlisted/unlocked posts when relaying via this connection.
	Unlisted bool `json:"unlisted"`

	// Controls whether a relay connection should match included, non-ignored statuses by default.
	// If set true, and no "exclude"-type matchers are set on the relay connection, then all included, non-ignored statuses will be relayed.
	MatchByDefault bool `json:"match_by_default"`

	// Ignore sensitive posts when relaying via this connection.
	IgnoreSensitive bool `json:"ignore_sensitive"`

	// Ignore posts with media attachments when relaying via this connection.
	IgnoreMedia bool `json:"ignore_media"`

	// Ignore replies to other accounts when relaying via this connection.
	IgnoreReplies bool `json:"ignore_replies"`
}

// RelayConnectionUpdateRequest models an update request for a relay push or relay subscription.
//
// swagger:ignore
type RelayConnectionUpdateRequest struct {
	// Include public posts when relaying via this connection.
	Public *bool `json:"public" form:"public" xml:"public"`

	// Include unlisted/unlocked posts when relaying via this connection.
	Unlisted *bool `json:"unlisted" form:"unlisted" xml:"unlisted"`

	// Controls whether a relay connection should match included, non-ignored statuses by default.
	// If set true, and no "exclude"-type matchers are set on the relay connection, then all included, non-ignored statuses will be relayed.
	MatchByDefault *bool `json:"match_by_default" form:"match_by_default" xml:"match_by_default"`

	// Ignore sensitive posts when relaying via this connection.
	IgnoreSensitive *bool `json:"ignore_sensitive" form:"ignore_sensitive" xml:"ignore_sensitive"`

	// Ignore posts with media attachments when relaying via this connection.
	IgnoreMedia *bool `json:"ignore_media" form:"ignore_media" xml:"ignore_media"`

	// Ignore connection owner's replies to other accounts when relaying via this connection.
	IgnoreReplies *bool `json:"ignore_replies" form:"ignore_replies" xml:"ignore_replies"`
}

// RelayConnectionCreateRequest models an create request for a relay push or relay subscription.
//
// swagger:ignore
type RelayConnectionCreateRequest struct {
	RelayConnectionUpdateRequest

	// ActivityPub URI of the relay service actor.
	// example: https://relay.activitypub.ca/actor
	RelayActorURI string `json:"relay_actor_uri" form:"relay_actor_uri" xml:"relay_actor_uri" binding:"required"`
}

// RelayMatcher models a relay matcher used to filter what is + isn't pushed / subscribed to by a relay connection.
//
// swagger:model relayMatcher
type RelayMatcher struct {
	// ID of this item.
	// example: 01KMQFYQHEZ6WCNCMN4629NBV8
	ID string `json:"id"`

	// The text to be matched.
	//
	// Example: whatever
	Keyword string `json:"keyword"`

	// Consider word boundaries when matching.
	WholeWord bool `json:"whole_word"`

	// If true, this relay matcher will cause matches to be EXCLUDED from relaying rather than INCLUDED in relaying.
	Exclude bool `json:"exclude"`
}

// RelayMatcherCreateUpdateRequest models a request to create or update a relay matcher for a relay connection.
//
// swagger:ignore
type RelayMatcherCreateUpdateRequest struct {
	// The text to be matched.
	Keyword *string `json:"keyword" form:"keyword" xml:"keyword"`

	// Consider word boundaries when matching.
	WholeWord *bool `json:"whole_word" form:"whole_word" xml:"whole_word"`

	// If true, this relay matcher will cause matches to be EXCLUDED from relaying rather than INCLUDED in relaying.
	Exclude *bool `json:"exclude" form:"exclude" xml:"exclude"`
}
