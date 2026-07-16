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

package regexes

import (
	"regexp"

	"code.superseriousbusiness.org/gopkg/buffers"
	"codeberg.org/gruf/go-byteutil"
	xurls "mvdan.cc/xurls/v2"
)

const (
	users     = "users"
	relays    = "relays"
	statuses  = "statuses"
	inbox     = "inbox"
	outbox    = "outbox"
	followers = "followers"
	following = "following"
	liked     = "liked"
	publicKey = "main-key"
	follow    = "follow"
	blocks    = "blocks"
	reports   = "reports"
	accepts   = "accepts"

	usersOrRelaysGrp         = `(` + users + `|` + relays + `)`
	alphaNumeric             = `\p{L}\p{M}*|\p{N}`                                       // A single number or script character in any language, including chars with accents.
	usernameGrp              = `(?:` + alphaNumeric + `|\.|\-|\_|` + unicodeEmoji + `)`  // Non-capturing group that matches against a single valid username character, including emojis.
	domainGrp                = `(?:` + alphaNumeric + `|\.|\-|\:|` + unicodeEmoji + `)`  // Non-capturing group that matches against a single valid domain character, including emojis.
	mentionName              = `^@(` + usernameGrp + `+)(?:@(` + domainGrp + `+))?$`     // Extract parts of one mention, maybe including domain.
	mentionFinder            = `(?:^|\s)(@` + usernameGrp + `+(?:@` + domainGrp + `+)?)` // Extract all mentions from a text, each mention may include domain.
	emojiShortcode           = `\w{1,30}`                                                // Pattern for emoji shortcodes. maximumEmojiShortcodeLength = 30
	emojiFinder              = `(?:\b)?:(` + emojiShortcode + `):(?:\b)?`                // Extract all emoji shortcodes from a text.
	emojiValidator           = `^` + emojiShortcode + `$`                                // Validate a single emoji shortcode.
	doubleSpaceFinder        = `(?:\b)?(?:\s|^) (?:\b)?`                                 // Extract all double whitespaces from a text.
	usernameStrict           = `^[a-z0-9_]{1,64}$`                                       // Pattern for usernames on THIS instance. maximumUsernameLength = 64
	usernameRelaxed          = `[a-z0-9_\-\.]{1,}`                                       // Relaxed version of username that can match instance accounts too (includes characters commonly found in domain names).
	misskeyReportNotesFinder = `(?m)(?:^Note: ((?:http|https):\/\/.*)$)`                 // Extract reported Note URIs from the text of a Misskey report/flag.
	ulid                     = `[0123456789ABCDEFGHJKMNPQRSTVWXYZ]{26}`                  // Pattern for ULID.
	ulidValidate             = `^` + ulid + `$`                                          // Validate one ULID.

	/*
		Path parts / capture.
	*/

	actorPathPrefix    = `^/?` + usersOrRelaysGrp + `/(` + usernameRelaxed + `)`
	actorPath          = actorPathPrefix + `$`
	actorWebPathPrefix = `^/?` + `@(` + usernameRelaxed + `)`
	actorWebPath       = actorWebPathPrefix + `$`
	publicKeyPath      = actorPathPrefix + `/` + publicKey + `$`
	inboxPath          = actorPathPrefix + `/` + inbox + `$`
	outboxPath         = actorPathPrefix + `/` + outbox + `$`
	followersPath      = actorPathPrefix + `/` + followers + `$`
	followingPath      = actorPathPrefix + `/` + following + `$`
	likedPath          = actorPathPrefix + `/` + liked + `$`
	followPath         = actorPathPrefix + `/` + follow + `/(` + ulid + `)$`
	likePath           = actorPathPrefix + `/` + liked + `/(` + ulid + `)$`
	statusesPath       = actorPathPrefix + `/` + statuses + `/(` + ulid + `)$`
	acceptsPath        = actorPathPrefix + `/` + accepts + `/(` + ulid + `)$`
	blockPath          = actorPathPrefix + `/` + blocks + `/(` + ulid + `)$`
	reportPath         = `^/?` + reports + `/(` + ulid + `)$`
	filePath           = `^/?(` + ulid + `)/([a-z]+)/([a-z]+)/(` + ulid + `)\.([a-z0-9]+)$`
)

var (
	// URLLike captures anything that resembles a URL. This includes URLs
	// with or without a scheme, and emails.
	URLLike = xurls.Relaxed()

	// MentionName captures the username and domain part from
	// a mention string such as @whatever_user@example.org,
	// returning whatever_user and example.org (without the @ symbols).
	// Will also work for characters with umlauts and other accents.
	// See: https://regex101.com/r/9tjNUy/1 for explanation and examples.
	MentionName = regexp.MustCompile(mentionName)

	// MentionFinder extracts whole mentions from a piece of text.
	MentionFinder = regexp.MustCompile(mentionFinder)

	// EmojiValidator validates an emoji shortcode.
	EmojiValidator = regexp.MustCompile(emojiValidator)

	// EmojiFinder extracts emoji strings from a piece of text.
	// See: https://regex101.com/r/478XGM/1
	EmojiFinder = regexp.MustCompile(emojiFinder)

	// DoubleSpaceFinder extracts double whitespaces from a piece of text.
	DoubleSpaceFinder = regexp.MustCompile(doubleSpaceFinder)

	// Username can be used to validate
	// usernames of new signups on this instance.
	Username = regexp.MustCompile(usernameStrict)

	// MisskeyReportNotes captures a list of Note
	// URIs from report content created by Misskey.
	//
	// See: https://regex101.com/r/EnTOBV/1
	MisskeyReportNotes = regexp.MustCompile(misskeyReportNotesFinder)

	// ULID parses and validate a ULID.
	ULID = regexp.MustCompile(ulidValidate)

	// ActorPath validates and captures the prefix and
	// username parts from eg /[users|relays]/example_username
	ActorPath = regexp.MustCompile(actorPath)

	// ActorWebPath validates and captures the prefix
	// and username parts from eg /@example_username.
	ActorWebPath = regexp.MustCompile(actorWebPath)

	// PublicKeyPath validates and captures the prefix and username
	// parts from eg /[users|relays]/example_username/main-key
	PublicKeyPath = regexp.MustCompile(publicKeyPath)

	// InboxPath validates and captures the prefix and username
	// parts from eg /[users|relays]/example_username/inbox
	InboxPath = regexp.MustCompile(inboxPath)

	// OutboxPath validates and captures the prefix and username
	// parts from eg /[users|relays]/example_username/outbox
	OutboxPath = regexp.MustCompile(outboxPath)

	// FollowersPath validates and captures the prefix and username
	// parts from eg /[users|relays]/example_username/followers
	FollowersPath = regexp.MustCompile(followersPath)

	// FollowingPath validates and captures the prefix and username
	// parts from eg /[users|relays]/example_username/following
	FollowingPath = regexp.MustCompile(followingPath)

	// LikedPath validates and captures the prefix and username
	// parts from eg /[users|relays]/example_username/liked
	LikedPath = regexp.MustCompile(likedPath)

	// FollowPath validates and captures the prefix, username, and ULID parts
	// from eg /[users|relays]/example_username/follow/01F7XT5JZW1WMVSW1KADS8PVDH
	FollowPath = regexp.MustCompile(followPath)

	// LikePath validates and captures the prefix, username, and ULID parts
	// from eg /[users|relays]/example_username/liked/01F7XT5JZW1WMVSW1KADS8PVDH
	LikePath = regexp.MustCompile(likePath)

	// StatusesPath validates and captures the prefix, username, and ULID parts
	// from eg /[users|relays]/example_username/statuses/01F7XT5JZW1WMVSW1KADS8PVDH
	//
	// The regex can be played with here: https://regex101.com/r/G9zuxQ/1
	StatusesPath = regexp.MustCompile(statusesPath)

	// BlockPath validates and captures the prefix, username, and ULID parts
	// from eg /[users|relays]/example_username/blocks/01F7XT5JZW1WMVSW1KADS8PVDH
	BlockPath = regexp.MustCompile(blockPath)

	// ReportPath validates and captures the ULID
	// part from eg /reports/01GP3AWY4CRDVRNZKW0TEAMB5R
	ReportPath = regexp.MustCompile(reportPath)

	// ReportPath validates and captures the prefix, username, and ULID parts
	// from eg /[users|relays]/example_username/accepts/01GP3AWY4CRDVRNZKW0TEAMB5R
	AcceptsPath = regexp.MustCompile(acceptsPath)

	// FilePath parses a file storage path of the form [ACCOUNT_ID]/[MEDIA_TYPE]/[MEDIA_SIZE]/[FILE_NAME]
	// eg 01F8MH1H7YV1Z7D2C8K2730QBF/attachment/small/01F8MH8RMYQ6MSNY3JM2XT1CQ5.jpeg
	//
	// It captures the account id, media type, media size, file name, and file extension, eg
	// `01F8MH1H7YV1Z7D2C8K2730QBF`, `attachment`, `small`, `01F8MH8RMYQ6MSNY3JM2XT1CQ5`, `jpeg`.
	FilePath = regexp.MustCompile(filePath)
)

// bufpool is a memory pool of byte buffers
// for use in our regex utility functions.
var bufpool = buffers.Pool(1024)

// ReplaceAllStringFunc will call through to .ReplaceAllStringFunc in the provided regex, but provide you a clean byte buffer for optimized string writes.
func ReplaceAllStringFunc(rgx *regexp.Regexp, src string, repl func(match string, buf *byteutil.Buffer) string) string {
	buf := bufpool.Get()
	defer bufpool.Put(buf)
	return rgx.ReplaceAllStringFunc(src, func(match string) string {
		buf.Reset() // reset use
		return repl(match, buf)
	})
}
