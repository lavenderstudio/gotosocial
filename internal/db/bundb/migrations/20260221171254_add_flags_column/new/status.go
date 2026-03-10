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

package gtsmodel

import (
	"time"
)

// Status represents a user-created 'post' or
// 'status' in the database, either remote or local
type Status struct {

	// Primary ID of this item in the database.
	ID string `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`

	// When was the status created.
	CreatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`

	// When this status was last edited (if set).
	EditedAt time.Time `bun:"type:timestamptz,nullzero"`

	// When was item (remote) last fetched.
	FetchedAt time.Time `bun:"type:timestamptz,nullzero"`

	// Activitypub URI of this status.
	URI string `bun:",unique,nullzero,notnull"`

	// Web url for viewing this status.
	URL string `bun:",nullzero"`

	// Content HTML for this status.
	Content string `bun:""`

	// Database IDs of any media attachments associated with this
	// status, and the attachments corresponding to attachmentIDs.
	AttachmentIDs []string `bun:"attachments,array"`

	// Database IDs of any tags used in this status, and the
	// tags corresponding to tagIDs. https://bun.uptrace.dev/guide/relations.html#many-to-many-relation
	TagIDs []string `bun:"tags,array"`

	// Database IDs of any mentions in this status,
	// and the mentions corresponding to mentionIDs.
	MentionIDs []string `bun:"mentions,array"`

	// Database IDs of any emojis used in this status, and the
	// emojis corresponding to emojiIDs. https://bun.uptrace.dev/guide/relations.html#many-to-many-relation
	EmojiIDs []string `bun:"emojis,array"`

	// ID of the account that posted this status, the
	// activitypub URI of the account that posted this status,
	// and the Account model corresponding to the accountID.
	AccountID  string `bun:"type:CHAR(26),nullzero,notnull"`
	AccountURI string `bun:",nullzero,notnull"`

	// ID of the status this is in reply to (or NULL), the
	// activitypub URI of the status this is in reply to (or NULL),
	// and the Status model corresponding to the statusID (if set).
	InReplyToID  string `bun:"type:CHAR(26),nullzero"`
	InReplyToURI string `bun:",nullzero"`

	// ID of the account that this status replies to,
	// account corresponding to inReplyToAccountID.
	InReplyToAccountID string `bun:"type:CHAR(26),nullzero"`

	// ID of the status this status is a boost of,
	// the URI of the status this status is a boost of
	// (not inserted in the db, just for dereferencing purposes),
	// and the status that corresponds to boostOfID.
	BoostOfID  string `bun:"type:CHAR(26),nullzero"`
	BoostOfURI string `bun:"-"`

	// ID of the account that owns the boosted status,
	// and account that corresponds to boostOfAccountID
	BoostOfAccountID string `bun:"type:CHAR(26),nullzero"`

	// ID of the thread to which this status belongs.
	ThreadID string `bun:"type:CHAR(26),nullzero,notnull,default:'00000000000000000000000000'"`

	// IDs of status edits for this status, ordered from
	// smallest (oldest) -> largest (newest) ID. Edits of
	// this status, ordered from oldest -> newest edit.
	EditIDs []string `bun:"edits,array"`

	// ID of the poll attached to this status,
	// and the Poll that corresponds to pollID.
	PollID string `bun:"type:CHAR(26),nullzero"`

	// Content warning HTML for this status, and the
	// original text of the content warning without formatting
	ContentWarning     string `bun:",nullzero"`
	ContentWarningText string `bun:""`

	// Flags contains numerous status boolean flags.
	Flags StatusFlags `bun:",notnull,default:0"`

	// Visibility entry for this status.
	Visibility Visibility `bun:",nullzero,notnull"`

	// What language is this status written in?
	Language string `bun:",nullzero"`

	// Which application was used to create this status? And
	// the application corresponding to createdWithApplicationID.
	CreatedWithApplicationID string `bun:"type:CHAR(26),nullzero"`

	// What is the activitystreams type of this status?
	// See: https://www.w3.org/TR/activitystreams-vocabulary/#object-types.
	//
	// Will probably almost always be Note but who knows!.
	ActivityStreamsType string `bun:",nullzero,notnull"`

	// Original text of the status without formatting.
	Text string `bun:""`

	// Content type used to process the original text of the status.
	ContentType StatusContentType `bun:",nullzero"`

	// InteractionPolicy for this status. If null then the default InteractionPolicy
	// should be assumed for this status's Visibility. Always null for boost wrappers.
	InteractionPolicy *InteractionPolicy `bun:""`

	// If true, then status is a reply to or boost wrapper of a status on
	// our instance, has permission to do the interaction, and an Accept
	// should be sent out for it immediately. Field not stored in the DB.
	PreApproved bool `bun:"-"`

	// URI of *either* an Accept Activity, or a ReplyAuthorization or
	// AnnounceAuthorization, which approves the Announce, Create or
	// interaction request Activity that this status was/will be attached to.
	ApprovedByURI string `bun:",nullzero"`
}

// Visibility represents the
// visibility granularity of a status.
type Visibility enumType

const (
	// VisibilityNone means nobody can see this.
	// It's only used for web status visibility.
	VisibilityNone Visibility = 1

	// VisibilityPublic means this status will
	// be visible to everyone on all timelines.
	VisibilityPublic Visibility = 2

	// VisibilityUnlocked means this status will be visible to everyone,
	// but will only show on home timeline to followers, and in lists.
	VisibilityUnlocked Visibility = 3

	// VisibilityFollowersOnly means this status is viewable to followers only.
	VisibilityFollowersOnly Visibility = 4

	// VisibilityMutualsOnly means this status
	// is visible to mutual followers only.
	VisibilityMutualsOnly Visibility = 5

	// VisibilityDirect means this status is
	// visible only to mentioned recipients.
	VisibilityDirect Visibility = 6

	// VisibilityDefault is used when no other setting can be found.
	VisibilityDefault Visibility = VisibilityUnlocked
)

// StatusContentType is the content type with which a status's text is
// parsed. Can be either plain or markdown. Empty will default to plain.
type StatusContentType enumType

const (
	StatusContentTypePlain    StatusContentType = 1
	StatusContentTypeMarkdown StatusContentType = 2
	StatusContentTypeDefault                    = StatusContentTypePlain
)
