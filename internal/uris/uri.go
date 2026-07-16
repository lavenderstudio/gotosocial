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

package uris

import (
	"fmt"
	"net/url"
	"strings"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/regexes"
)

// RelayUsernamePrefix is the prefix
// for local relay actor usernames.
const RelayUsernamePrefix = "relay."

// PathComponent is a string type alias
// representing a path within a URI,
// eg., "users", "accepts", etc.
type PathComponent = string

const (
	// UsersPath is for serving user actors.
	UsersPath PathComponent = "users"

	// RelaysPath is like UsersPath
	// but for our local relay actors.
	RelaysPath PathComponent = "relays"

	// StatusesPath is for serving statuses.
	StatusesPath PathComponent = "statuses"

	// InboxPath represents the
	// ActivityPub inbox location.
	InboxPath PathComponent = "inbox"

	// OutboxPath represents the
	// ActivityPub outbox location.
	OutboxPath PathComponent = "outbox"

	// FollowersPath represents the
	// ActivityPub followers location.
	FollowersPath PathComponent = "followers"

	// FollowingPath represents the
	// ActivityPub following location.
	FollowingPath PathComponent = "following"

	// LikedPath represents the
	// ActivityPub liked location.
	LikedPath PathComponent = "liked"

	// CollectionsPath represents the
	// ActivityPub collections location.
	CollectionsPath PathComponent = "collections"

	// FeaturedPath represents the
	// ActivityPub featured location.
	FeaturedPath PathComponent = "featured"

	// PublicKeyPath is for serving
	// an actor's public RSA key.
	PublicKeyPath PathComponent = "main-key"

	// FollowPath is used to generate the
	// URI for a Follow (request) Activity.
	FollowPath PathComponent = "follow"

	// UpdatePath is used to generate
	// the URI for an Update activity.
	UpdatePath PathComponent = "updates"

	// BlocksPath is used to generate
	// the URI for a Block activity.
	BlocksPath PathComponent = "blocks"

	// MovesPath is used to generate
	// the URI for a Move activity.
	MovesPath PathComponent = "moves"

	// ReportsPath is used to generate the
	// URI for a Flag (ie., report) activity.
	ReportsPath PathComponent = "reports"

	// ConfirmEmailPath is used to generate
	// the URI for an email confirmation link.
	ConfirmEmailPath PathComponent = "confirm_email"

	// FileserverPath is a path component
	// for serving attachments + emoji files.
	FileserverPath PathComponent = "fileserver"

	// EmojiPath represents the
	// ActivityPub Emoji location.
	EmojiPath PathComponent = "emoji"

	// TagsPath represents the
	// ActivityPub Tags location.
	TagsPath PathComponent = "tags"

	// AcceptsPath represents the location
	// of an ActivityPub Accept activity.
	AcceptsPath PathComponent = "accepts"

	// RejectsPath represents the location
	// of an ActivityPub Reject activity.
	RejectsPath PathComponent = "rejects"

	// AuthorizationsPath represents the location of
	// an Authorization type such as LikeAuthorization,
	// ReplyAuthorization, AnnounceAuthorization, etc.
	AuthorizationsPath PathComponent = "authorizations"

	// LikeRequestsPath is used to generate
	// the URI for a LikeRequest activity.
	LikeRequestsPath PathComponent = "like_requests"

	// ReplyRequestsPath is used to generate
	// the URI for a ReplyRequest activity.
	ReplyRequestsPath PathComponent = "reply_requests"

	// LikeRequestsPath is used to generate
	// the URI for an AnnounceRequest activity.
	AnnounceRequestsPath PathComponent = "announce_requests"
)

// ActorURIs encapsulates a bunch of URIs
// and URLs for a user, host, account, etc.
type ActorURIs struct {

	// The web URL of the instance
	// host, eg https://example.org
	HostURL string

	// The web URL of the actor,
	// eg., https://example.org/@example_user
	ActorURL string

	// The web URL for statuses of this actor,
	// eg., https://example.org/@example_user/statuses
	StatusesURL string

	// The ActivityPub URI/ID of the actor,
	// eg., https://example.org/users/example_user
	ActorURI string

	// The ActivityPub URI for this actor's statuses,
	// eg., https://example.org/users/example_user/statuses
	StatusesURI string

	// The ActivityPub URI for this actor's ActivityPub inbox,
	// eg., https://example.org/users/example_user/inbox
	InboxURI string

	// The ActivityPub URI for this actor's ActivityPub outbox,
	// eg., https://example.org/users/example_user/outbox
	OutboxURI string

	// The ActivityPub URI for this actor's followers,
	// eg., https://example.org/users/example_user/followers
	FollowersURI string

	// The ActivityPub URI for this actor's following,
	// eg., https://example.org/users/example_user/following
	FollowingURI string

	// The ActivityPub URI for this actor's liked posts.
	// eg., https://example.org/users/example_user/liked
	LikedURI string

	// The ActivityPub URI for this actor's featured collections,
	// eg., https://example.org/users/example_user/collections/featured
	FeaturedCollectionURI string

	// The URI for this actor's public key,
	// eg., https://example.org/users/example_user/publickey
	PublicKeyURI string
}

// ensure path prefix is as expected, or panic.
func checkPathPrefix(pc PathComponent) PathComponent {
	switch pc {
	case UsersPath, RelaysPath:
		return pc // OK.
	default:
		panic("unusable pathPrefix")
	}
}

// GenerateURIForFollow returns the AP URI for a new follow -- something like:
// https://example.org/users/whatever_user/follow/01F7XTH1QGBAPMGF49WJZ91XGC
func GenerateURIForFollow(pathPrefix PathComponent, username string, id string) string {
	proto := config.GetProtocol()
	host := config.GetHost()
	return buildURL4(proto,
		host,
		checkPathPrefix(pathPrefix),
		username,
		FollowPath,
		id,
	)
}

// GenerateURIForLike returns the AP URI for a new like/fave -- something like:
// https://example.org/users/whatever_user/liked/01F7XTH1QGBAPMGF49WJZ91XGC
func GenerateURIForLike(pathPrefix PathComponent, username string, id string) string {
	proto := config.GetProtocol()
	host := config.GetHost()
	return buildURL4(proto,
		host,
		checkPathPrefix(pathPrefix),
		username,
		LikedPath,
		id,
	)
}

// GenerateURIForUpdate returns the AP URI for a new update activity -- something like:
// https://example.org/users/whatever_user#updates/01F7XTH1QGBAPMGF49WJZ91XGC
func GenerateURIForUpdate(pathPrefix PathComponent, username string, thisUpdateID string) string {
	proto := config.GetProtocol()
	host := config.GetHost()
	return buildURL4(proto,
		host,
		checkPathPrefix(pathPrefix),
		username,
		UpdatePath,
		thisUpdateID,
	)
}

// GenerateURIForBlock returns the AP URI for a new block activity -- something like:
// https://example.org/users/whatever_user/blocks/01F7XTH1QGBAPMGF49WJZ91XGC
func GenerateURIForBlock(pathPrefix PathComponent, username string, thisBlockID string) string {
	proto := config.GetProtocol()
	host := config.GetHost()
	return buildURL4(proto,
		host,
		checkPathPrefix(pathPrefix),
		username,
		BlocksPath,
		thisBlockID,
	)
}

// GenerateURIForMove returns the AP URI for a new Move activity -- something like:
// https://example.org/users/whatever_user/moves/01F7XTH1QGBAPMGF49WJZ91XGC
func GenerateURIForMove(pathPrefix PathComponent, username string, thisMoveID string) string {
	proto := config.GetProtocol()
	host := config.GetHost()
	return buildURL4(proto,
		host,
		checkPathPrefix(pathPrefix),
		username,
		MovesPath,
		thisMoveID,
	)
}

// GenerateURIForReport returns the API URI for a new Flag activity -- something like:
// https://example.org/reports/01GP3AWY4CRDVRNZKW0TEAMB5R
//
// This path specifically doesn't contain any info about the user who did the reporting,
// to protect their privacy.
func GenerateURIForReport(thisReportID string) string {
	proto := config.GetProtocol()
	host := config.GetHost()
	return buildURL2(proto,
		host,
		ReportsPath,
		thisReportID,
	)
}

// GenerateURIForEmailConfirm returns a link for email confirmation -- something like:
// https://example.org/confirm_email?token=490e337c-0162-454f-ac48-4b22bb92a205
func GenerateURIForEmailConfirm(token string) string {
	proto := config.GetProtocol()
	host := config.GetHost()
	return buildURL1(proto, host, ConfirmEmailPath) + "?token=" + token
}

// GenerateURIForAccept returns the AP URI for a new Accept activity -- something like:
// https://example.org/users/whatever_user/accepts/01F7XTH1QGBAPMGF49WJZ91XGC
func GenerateURIForAccept(pathPrefix PathComponent, username string, thisAcceptID string) string {
	proto := config.GetProtocol()
	host := config.GetHost()
	return buildURL4(proto,
		host,
		checkPathPrefix(pathPrefix),
		username,
		AcceptsPath,
		thisAcceptID,
	)
}

// GenerateURIForAuthorization returns the AP URI for a new Authorization object,
// ie., LikeAuthorization, ReplyAuthorization, or AnnounceAuthorization.
// Eg., https://example.org/users/whatever_user/authorizations/01F7XTH1QGBAPMGF49WJZ91XGC
func GenerateURIForAuthorization(pathPrefix PathComponent, username string, id string) string {
	proto := config.GetProtocol()
	host := config.GetHost()
	return buildURL4(proto,
		host,
		checkPathPrefix(pathPrefix),
		username,
		AuthorizationsPath,
		id,
	)
}

// GenerateURIForLikeRequest returns the AP URI for a new LikeRequest object,
// Eg., https://example.org/users/whatever_user/like_requests/01F7XTH1QGBAPMGF49WJZ91XGC
func GenerateURIForLikeRequest(pathPrefix PathComponent, username string, id string) string {
	proto := config.GetProtocol()
	host := config.GetHost()
	return buildURL4(proto,
		host,
		checkPathPrefix(pathPrefix),
		username,
		LikeRequestsPath,
		id,
	)
}

// GenerateURIForReplyRequest returns the AP URI for a new ReplyRequest object,
// Eg., https://example.org/users/whatever_user/reply_requests/01F7XTH1QGBAPMGF49WJZ91XGC
func GenerateURIForReplyRequest(pathPrefix PathComponent, username string, id string) string {
	proto := config.GetProtocol()
	host := config.GetHost()
	return buildURL4(proto,
		host,
		checkPathPrefix(pathPrefix),
		username,
		ReplyRequestsPath,
		id,
	)
}

// GenerateURIForAnnounceRequest returns the AP URI for a new AnnounceRequest object,
// Eg., https://example.org/users/whatever_user/announce_requests/01F7XTH1QGBAPMGF49WJZ91XGC
func GenerateURIForAnnounceRequest(pathPrefix PathComponent, username string, id string) string {
	proto := config.GetProtocol()
	host := config.GetHost()
	return buildURL4(proto,
		host,
		checkPathPrefix(pathPrefix),
		username,
		AnnounceRequestsPath,
		id,
	)
}

// GenerateURIForReject returns the AP URI for a new Reject activity -- something like:
// https://example.org/users/whatever_user/rejects/01F7XTH1QGBAPMGF49WJZ91XGC
func GenerateURIForReject(pathPrefix PathComponent, username string, thisRejectID string) string {
	proto := config.GetProtocol()
	host := config.GetHost()
	return buildURL4(proto,
		host,
		checkPathPrefix(pathPrefix),
		username,
		RejectsPath,
		thisRejectID,
	)
}

// GenerateActorURIs throws together a bunch
// of URIs for the actor with the given username.
func GenerateActorURIs(pathPrefix PathComponent, username string) ActorURIs {
	proto := config.GetProtocol()
	host := config.GetHost()

	// URLs for serving web requests.
	hostURL := proto + "://" + host
	var userURL string
	if pathPrefix == RelaysPath {
		userURL = hostURL + "/@relay." + username
	} else {
		userURL = hostURL + "/@" + username
	}
	statusesURL := userURL + "/" + StatusesPath

	// The below URIs are used in ActivityPub and Webfinger
	userURI := hostURL + "/" + string(checkPathPrefix(pathPrefix)) + "/" + username
	statusesURI := userURI + "/" + StatusesPath
	inboxURI := userURI + "/" + InboxPath
	outboxURI := userURI + "/" + OutboxPath
	followersURI := userURI + "/" + FollowersPath
	followingURI := userURI + "/" + FollowingPath
	likedURI := userURI + "/" + LikedPath
	collectionURI := userURI + "/" + CollectionsPath + "/" + FeaturedPath
	publicKeyURI := userURI + "/" + PublicKeyPath

	return ActorURIs{
		HostURL:               hostURL,
		ActorURL:              userURL,
		StatusesURL:           statusesURL,
		ActorURI:              userURI,
		StatusesURI:           statusesURI,
		InboxURI:              inboxURI,
		OutboxURI:             outboxURI,
		FollowersURI:          followersURI,
		FollowingURI:          followingURI,
		LikedURI:              likedURI,
		FeaturedCollectionURI: collectionURI,
		PublicKeyURI:          publicKeyURI,
	}
}

// URIForAttachment generates a URI
// for an attachment/emoji/header etc.
//
// Will produce something like:
//
//	"https://example.org/fileserver/01FPST95B8FC3HG3AGCDKPQNQ2/attachment/original/01FPST9QK4V5XWS3F9Z4F2G1X7.gif"
func URIForAttachment(
	accountID string,
	mediaType string,
	mediaSize string,
	mediaID string,
	extension string,
) string {
	proto := config.GetProtocol()
	host := config.GetHost()
	return buildURL5(proto,
		host,
		FileserverPath,
		accountID,
		mediaType,
		mediaSize,
		mediaID,
	) + "." + extension
}

// StoragePathForAttachment generates a storage
// path for an attachment/emoji/header etc.
//
// Will produce something like:
//
//	"01FPST95B8FC3HG3AGCDKPQNQ2/attachment/original/01FPST9QK4V5XWS3F9Z4F2G1X7.gif"
func StoragePathForAttachment(
	accountID string,
	mediaType string,
	mediaSize string,
	mediaID string,
	extension string,
) string {
	return buildPath4(
		accountID,
		mediaType,
		mediaSize,
		mediaID,
	) + "." + extension
}

// URIForEmoji generates an
// ActivityPub URI for an emoji.
//
// Will produce something like:
//
//	"https://example.org/emoji/01FPST9QK4V5XWS3F9Z4F2G1X7"
func URIForEmoji(emojiID string) string {
	proto := config.GetProtocol()
	host := config.GetHost()
	return buildURL2(proto,
		host,
		EmojiPath,
		emojiID,
	)
}

// URIForTag generates an ActivityPub uri for a tag.
func URIForTag(name string) string {
	proto := config.GetProtocol()
	host := config.GetHost()
	name = strings.ToLower(name)
	return buildURL2(proto,
		host,
		TagsPath,
		name,
	)
}

// IsActorPath returns true if the given URL path corresponds to eg /users/example_username
func IsActorPath(id *url.URL) bool {
	return regexes.ActorPath.MatchString(id.Path)
}

// IsUserWebPath returns true if the given URL path corresponds to eg /@example_username
func IsUserWebPath(id *url.URL) bool {
	return regexes.ActorWebPath.MatchString(id.Path)
}

// IsInboxPath returns true if the given URL path corresponds to eg /users/example_username/inbox
func IsInboxPath(id *url.URL) bool {
	return regexes.InboxPath.MatchString(id.Path)
}

// IsOutboxPath returns true if the given URL path corresponds to eg /users/example_username/outbox
func IsOutboxPath(id *url.URL) bool {
	return regexes.OutboxPath.MatchString(id.Path)
}

// IsFollowersPath returns true if the given URL path corresponds to eg /users/example_username/followers
func IsFollowersPath(id *url.URL) bool {
	return regexes.FollowersPath.MatchString(id.Path)
}

// IsFollowingPath returns true if the given URL path corresponds to eg /users/example_username/following
func IsFollowingPath(id *url.URL) bool {
	return regexes.FollowingPath.MatchString(id.Path)
}

// IsFollowPath returns true if the given URL path corresponds to eg /users/example_username/follow/SOME_ULID_OF_A_FOLLOW
func IsFollowPath(id *url.URL) bool {
	return regexes.FollowPath.MatchString(id.Path)
}

// IsLikedPath returns true if the given URL path corresponds to eg /users/example_username/liked
func IsLikedPath(id *url.URL) bool {
	return regexes.LikedPath.MatchString(id.Path)
}

// IsLikePath returns true if the given URL path corresponds to eg /users/example_username/liked/SOME_ULID_OF_A_STATUS
func IsLikePath(id *url.URL) bool {
	return regexes.LikePath.MatchString(id.Path)
}

// IsStatusesPath returns true if the given URL path corresponds to eg /users/example_username/statuses/SOME_ULID_OF_A_STATUS
func IsStatusesPath(id *url.URL) bool {
	return regexes.StatusesPath.MatchString(id.Path)
}

// IsPublicKeyPath returns true if the given URL path corresponds to eg /users/example_username/main-key
func IsPublicKeyPath(id *url.URL) bool {
	return regexes.PublicKeyPath.MatchString(id.Path)
}

// IsBlockPath returns true if the given URL path corresponds to eg /users/example_username/blocks/SOME_ULID_OF_A_BLOCK
func IsBlockPath(id *url.URL) bool {
	return regexes.BlockPath.MatchString(id.Path)
}

// IsReportPath returns true if the given URL path corresponds to eg /reports/SOME_ULID_OF_A_REPORT
func IsReportPath(id *url.URL) bool {
	return regexes.ReportPath.MatchString(id.Path)
}

// IsAcceptsPath returns true if the given URL path corresponds to eg /users/example_username/accepts/SOME_ULID_OF_AN_ACCEPT
func IsAcceptsPath(id *url.URL) bool {
	return regexes.AcceptsPath.MatchString(id.Path)
}

// ParseStatusesPath returns the path prefix, username, and ulid from a
// path such as /users/example_username/statuses/SOME_ULID_OF_A_STATUS.
//
// If the prefix is "relays" then "relay." will be
// automatically prepended to the returned username.
func ParseStatusesPath(id *url.URL) (
	pathPrefix PathComponent,
	username string,
	ulid string,
	err error,
) {
	matches := regexes.StatusesPath.FindStringSubmatch(id.Path)
	if len(matches) != 4 {
		err = fmt.Errorf("expected 4 matches but matches length was %d", len(matches))
		return
	}
	pathPrefix = matches[1]
	username = matches[2]
	if pathPrefix == RelaysPath {
		username = RelayUsernamePrefix + username
	}
	ulid = matches[3]
	return
}

// ParseActorPath returns the path prefix and username
// from a path such as /users/example_username
//
// If the prefix is "relays" then "relay." will be
// automatically prepended to the returned username.
func ParseActorPath(id *url.URL) (
	pathPrefix PathComponent,
	username string,
	err error,
) {
	matches := regexes.ActorPath.FindStringSubmatch(id.Path)
	if len(matches) != 3 {
		err = fmt.Errorf("expected 3 matches but matches length was %d", len(matches))
		return
	}
	pathPrefix = matches[1]
	username = matches[2]
	if pathPrefix == RelaysPath {
		username = RelayUsernamePrefix + username
	}
	return
}

// ParseUserPath returns the username from a path such as /@example_username
func ParseActorWebPath(id *url.URL) (username string, err error) {
	matches := regexes.ActorWebPath.FindStringSubmatch(id.Path)
	if len(matches) != 2 {
		err = fmt.Errorf("expected 2 matches but matches length was %d", len(matches))
		return
	}
	username = matches[1]
	return
}

// ParseInboxPath returns the path prefix and username
// from a path such as /users/example_username/inbox
//
// If the prefix is "relays" then "relay." will be
// automatically prepended to the returned username.
func ParseInboxPath(id *url.URL) (
	pathPrefix PathComponent,
	username string,
	err error,
) {
	matches := regexes.InboxPath.FindStringSubmatch(id.Path)
	if len(matches) != 3 {
		err = fmt.Errorf("expected 2 matches but matches length was %d", len(matches))
		return
	}
	pathPrefix = matches[1]
	username = matches[2]
	if pathPrefix == RelaysPath {
		username = RelayUsernamePrefix + username
	}
	return
}

// ParseOutboxPath returns the path prefix and username
// from a path such as /users/example_username/outbox
//
// If the prefix is "relays" then "relay." will be
// automatically prepended to the returned username.
func ParseOutboxPath(id *url.URL) (
	pathPrefix PathComponent,
	username string,
	err error,
) {
	matches := regexes.OutboxPath.FindStringSubmatch(id.Path)
	if len(matches) != 3 {
		err = fmt.Errorf("expected 3 matches but matches length was %d", len(matches))
		return
	}
	pathPrefix = matches[1]
	username = matches[2]
	if pathPrefix == RelaysPath {
		username = RelayUsernamePrefix + username
	}
	return
}

// ParseFollowersPath returns the path prefix and username
// from a path such as /users/example_username/followers
//
// If the prefix is "relays" then "relay." will be
// automatically prepended to the returned username.
func ParseFollowersPath(id *url.URL) (
	pathPrefix PathComponent,
	username string,
	err error,
) {
	matches := regexes.FollowersPath.FindStringSubmatch(id.Path)
	if len(matches) != 3 {
		err = fmt.Errorf("expected 3 matches but matches length was %d", len(matches))
		return
	}
	pathPrefix = matches[1]
	username = matches[2]
	if pathPrefix == RelaysPath {
		username = RelayUsernamePrefix + username
	}
	return
}

// ParseFollowingPath returns the path prefix and username
// from a path such as /users/example_username/following
//
// If the prefix is "relays" then "relay." will be
// automatically prepended to the returned username.
func ParseFollowingPath(id *url.URL) (
	pathPrefix PathComponent,
	username string,
	err error,
) {
	matches := regexes.FollowingPath.FindStringSubmatch(id.Path)
	if len(matches) != 3 {
		err = fmt.Errorf("expected 3 matches but matches length was %d", len(matches))
		return
	}
	pathPrefix = matches[1]
	username = matches[2]
	if pathPrefix == RelaysPath {
		username = RelayUsernamePrefix + username
	}
	return
}

// ParseLikedPath returns the path prefix, username, and ulid from a
// path such as /users/example_username/liked/SOME_ULID_OF_A_STATUS
//
// If the prefix is "relays" then "relay." will be
// automatically prepended to the returned username.
func ParseLikedPath(id *url.URL) (
	pathPrefix PathComponent,
	username string,
	ulid string,
	err error,
) {
	matches := regexes.LikePath.FindStringSubmatch(id.Path)
	if len(matches) != 4 {
		err = fmt.Errorf("expected 4 matches but matches length was %d", len(matches))
		return
	}
	pathPrefix = matches[1]
	username = matches[2]
	if pathPrefix == RelaysPath {
		username = RelayUsernamePrefix + username
	}
	ulid = matches[3]
	return
}

// ParseBlockPath returns the path prefix, username, and ulid from a
// path such as /users/example_username/blocks/SOME_ULID_OF_A_BLOCK
//
// If the prefix is "relays" then "relay." will be
// automatically prepended to the returned username.
func ParseBlockPath(id *url.URL) (
	pathPrefix PathComponent,
	username string,
	ulid string,
	err error,
) {
	matches := regexes.BlockPath.FindStringSubmatch(id.Path)
	if len(matches) != 4 {
		err = fmt.Errorf("expected 4 matches but matches length was %d", len(matches))
		return
	}
	pathPrefix = matches[1]
	username = matches[2]
	if pathPrefix == RelaysPath {
		username = RelayUsernamePrefix + username
	}
	ulid = matches[3]
	return
}

// ParseReportPath returns the ulid from a path such as /reports/SOME_ULID_OF_A_REPORT
func ParseReportPath(id *url.URL) (ulid string, err error) {
	matches := regexes.ReportPath.FindStringSubmatch(id.Path)
	if len(matches) != 2 {
		err = fmt.Errorf("expected 2 matches but matches length was %d", len(matches))
		return
	}
	ulid = matches[1]
	return
}
