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

import "mime/multipart"

// RelayActor models a local
// admin-created relay actor.
//
// swagger:model relayActor
type RelayActor struct {
	// ID of this item.
	// example: 01KMQFY8C9P2049NN09R9CCMSR
	ID string `json:"id"`

	// The date when this relay actor was created (ISO 8601 Datetime).
	// example: 2021-07-30T09:20:25+00:00
	CreatedAt string `json:"created_at"`

	// ID of the admin account that created this relay actor.
	// example: 01KMQFY8C9P2049NN09R9CCMSR
	CreatedByAccountID string `json:"created_by_account_id"`

	// Account model for this actor.
	Account *Account `json:"account"`

	// Show followers of this relay actor on the web view of the relay actor account.
	WebShowFollowers bool `json:"web_show_followers"`

	// Matchers that apply to this relay actor.
	Matchers []RelayMatcher `json:"matchers"`

	RelayFlags
}

// swagger:ignore
type RelayActorCreateRequest struct {
	RelayActorUpdateRequest

	// Username to use for the relay actor.
	//
	// Will be prepended with "relay." to
	// form the full relay username.
	Username string `json:"username" form:"username" binding:"required"`
}

// Comply with interface WithFieldsAttributes (api/model/common.go).
func (form *RelayActorCreateRequest) GetFieldsAttributes() *[]UpdateField {
	return form.FieldsAttributes
}

// Comply with interface WithFieldsAttributes (api/model/common.go).
func (form *RelayActorCreateRequest) SetFieldsAttributes(v *[]UpdateField) {
	form.FieldsAttributes = v
}

// Comply with interface WithFieldsAttributes (api/model/common.go).
func (form *RelayActorCreateRequest) GetJSONFieldsAttributes() *map[string]UpdateField {
	return form.JSONFieldsAttributes
}

// Comply with interface WithFieldsAttributes (api/model/common.go).
func (form *RelayActorCreateRequest) SetJSONFieldsAttributes(v *map[string]UpdateField) {
	form.JSONFieldsAttributes = v
}

// swagger:ignore
type RelayActorUpdateRequest struct {
	RelayFlagsForm

	// Relay actor account should be made discoverable
	// and shown in the profile directory (if enabled).
	Discoverable *bool `form:"discoverable" json:"discoverable"`

	// The display name to use for the relay actor account.
	DisplayName *string `form:"display_name" json:"display_name"`

	// Bio/description of this relay actor account.
	Note *string `form:"note" json:"note"`

	// Avatar image encoded using multipart/form-data.
	Avatar *multipart.FileHeader `form:"avatar" json:"-"`

	// Description of the avatar image, for alt-text.
	AvatarDescription *string `form:"avatar_description" json:"avatar_description"`

	// Header image encoded using multipart/form-data
	Header *multipart.FileHeader `form:"header" json:"-"`

	// Description of the header image, for alt-text.
	HeaderDescription *string `form:"header_description" json:"header_description"`

	// Require manual approval of follow requests.
	Locked *bool `form:"locked" json:"locked"`

	// Show approved followers of this relay on the relay actor account's web view.
	WebShowFollowers *bool `form:"web_show_followers" json:"web_show_followers"`

	// Profile metadata names and values.
	FieldsAttributes *[]UpdateField `form:"fields_attributes" json:"-"`

	// Profile metadata names and values, parsed from JSON.
	JSONFieldsAttributes *map[string]UpdateField `form:"-" json:"fields_attributes"`
}

// Comply with interface WithFieldsAttributes (api/model/common.go).
func (form *RelayActorUpdateRequest) GetFieldsAttributes() *[]UpdateField {
	return form.FieldsAttributes
}

// Comply with interface WithFieldsAttributes (api/model/common.go).
func (form *RelayActorUpdateRequest) SetFieldsAttributes(v *[]UpdateField) {
	form.FieldsAttributes = v
}

// Comply with interface WithFieldsAttributes (api/model/common.go).
func (form *RelayActorUpdateRequest) GetJSONFieldsAttributes() *map[string]UpdateField {
	return form.JSONFieldsAttributes
}

// Comply with interface WithFieldsAttributes (api/model/common.go).
func (form *RelayActorUpdateRequest) SetJSONFieldsAttributes(v *map[string]UpdateField) {
	form.JSONFieldsAttributes = v
}

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

	RelayFlags
}

// swagger:model relayFlags
type RelayFlags struct {
	// Include public posts when relaying.
	Public bool `json:"public"`

	// Include unlisted/unlocked posts when relaying.
	Unlisted bool `json:"unlisted"`

	// Controls whether a relay should match included, non-ignored statuses by default.
	// If set true, and no "exclude"-type matchers are set, then all included, non-ignored statuses will be relayed.
	MatchByDefault bool `json:"match_by_default"`

	// Ignore sensitive posts when relaying.
	IgnoreSensitive bool `json:"ignore_sensitive"`

	// Ignore posts with media attachments when relaying.
	IgnoreMedia bool `json:"ignore_media"`

	// Ignore replies to other accounts when relaying.
	IgnoreReplies bool `json:"ignore_replies"`
}

// RelayFlagsForm models a flags create or update
// request for a relay push, subscription, or actor.
//
// swagger:ignore
type RelayFlagsForm struct {
	// Include public posts when relaying.
	Public *bool `json:"public" form:"public" xml:"public"`

	// Include unlisted/unlocked posts when relaying.
	Unlisted *bool `json:"unlisted" form:"unlisted" xml:"unlisted"`

	// Controls whether a relay entity should match included, non-ignored statuses by default.
	// If set true, and no "exclude"-type matchers are set on the relay entity, then all included, non-ignored statuses will be relayed.
	MatchByDefault *bool `json:"match_by_default" form:"match_by_default" xml:"match_by_default"`

	// Ignore sensitive posts when relaying.
	IgnoreSensitive *bool `json:"ignore_sensitive" form:"ignore_sensitive" xml:"ignore_sensitive"`

	// Ignore posts with media attachments when relaying.
	IgnoreMedia *bool `json:"ignore_media" form:"ignore_media" xml:"ignore_media"`

	// Ignore replies to other accounts when relaying.
	IgnoreReplies *bool `json:"ignore_replies" form:"ignore_replies" xml:"ignore_replies"`
}

// RelayConnectionCreateRequest models a create
// request for a relay push or relay subscription.
//
// swagger:ignore
type RelayConnectionCreateRequest struct {
	RelayFlagsForm

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
