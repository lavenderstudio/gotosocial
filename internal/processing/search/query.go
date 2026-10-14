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

package search

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"code.superseriousbusiness.org/gopkg/buffers"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"codeberg.org/gruf/go-longdur"
)

// errNoOutput is returned when conflicting search
// parameters exist such that no output would be returned.
var errNoOutput = errors.New("search produces no output")

type searchQuery struct {
	db.CommonSearchParams

	Before time.Time
	After  time.Time

	// Remaining string
	// after all params
	// are extracted.
	Remain string

	// account search params
	UsernameQuery string
	DomainQuery   string
	NoteQuery     string

	// status search params
	AuthorID            string
	From                string
	Reply               db.OnlyBool
	Sensitive           db.OnlyBool
	Media               db.OnlyBool
	Poll                db.OnlyBool
	ContentQuery        string
	ContentWarningQuery string
}

func parseSearchQuery(
	query string,
	follows bool,
	authorID string,
) (
	*searchQuery,
	error,
) {
	var err error
	searchQuery := new(searchQuery)
	searchQuery.Remain, err = parseQueryTokens(query, func(token string) (used bool, err error) {
		used = true
		switch {

		// query filter:
		// "show statuses only by given account"
		case strings.HasPrefix(token, "from:"):
			token = token[len("from:"):]
			token = remoteQuotes(token)
			if token != "" {
				searchQuery.From = token
			}

		// query filter:
		// "show accounts / statuses created before"
		case strings.HasPrefix(token, "before:"):
			token = token[len("before:"):]
			token = remoteQuotes(token)
			searchQuery.Before, err = parseTime(token)

		// query filter:
		// "show accounts / statuses created after"
		case strings.HasPrefix(token, "after:"):
			token = token[len("after:"):]
			token = remoteQuotes(token)
			searchQuery.After, err = parseTime(token)

		// query filter:
		// "show accounts matching username query"
		case strings.HasPrefix(token, "username:"):
			token = token[len("username:"):]
			token = remoteQuotes(token)
			if token != "" {
				searchQuery.UsernameQuery = token
			}

		// query filter:
		// "show accounts matching domain query"
		case strings.HasPrefix(token, "domain:"):
			token = token[len("domain:"):]
			token = remoteQuotes(token)
			if token != "" {
				searchQuery.DomainQuery = token
			}

		// query filter:
		// "show accounts matching note query"
		case strings.HasPrefix(token, "note:"):
			token = token[len("note"):]
			token = remoteQuotes(token)
			if token != "" {
				searchQuery.NoteQuery = token
			}

		// query filter:
		// "show statuses matching content query"
		case strings.HasPrefix(token, "content:"):
			token = token[len("content:"):]
			token = remoteQuotes(token)
			if token != "" {
				searchQuery.ContentQuery = token
			}

		// query filter:
		// "show statuses matching content warning query"
		case strings.HasPrefix(token, "content_warning:"):
			token = token[len("content_warning:"):]
			token = remoteQuotes(token)
			if token != "" {
				searchQuery.ContentWarningQuery = token
			}

		// query filter:
		// "show accounts / statuses that only are ___"
		case strings.HasPrefix(token, "is:"):
			token = token[len("is:"):]
			token = remoteQuotes(token)
			token = strings.ToLower(token)
			switch token {

			// is remote
			case "remote":
				switch searchQuery.Location {
				case db.LocationLocal:
					err = errNoOutput
				case db.LocationAny:
					searchQuery.Location = db.LocationRemote
				}

			// is local
			case "local":
				switch searchQuery.Location {
				case db.LocationRemote:
					err = errNoOutput
				case db.LocationAny:
					searchQuery.Location = db.LocationLocal
				}

			// is reply.
			case "reply":
				switch searchQuery.Reply {
				case db.OnlyFalse:
					err = errNoOutput
				default:
					searchQuery.Reply = db.OnlyTrue
				}

			// is not reply.
			case "notreply":
				switch searchQuery.Reply {
				case db.OnlyTrue:
					err = errNoOutput
				default:
					searchQuery.Reply = db.OnlyFalse
				}

			// is sensitive.
			case "sensitive":
				switch searchQuery.Sensitive {
				case db.OnlyFalse:
					err = errNoOutput
				default:
					searchQuery.Sensitive = db.OnlyTrue
				}

			// is not sensitive.
			case "notsensitive":
				switch searchQuery.Sensitive {
				case db.OnlyTrue:
					err = errNoOutput
				default:
					searchQuery.Sensitive = db.OnlyFalse
				}

			default:
				// Set token as unused, gets
				// returned below as 'remain'.
				used = false
			}

		// query filter:
		// "show statuses only with ___"
		case strings.HasPrefix(token, "has:"):
			token = token[len("has:"):]
			token = remoteQuotes(token)
			token = strings.ToLower(token)
			switch token {

			// with media.
			case "media":
				switch searchQuery.Media {
				case db.OnlyFalse:
					err = errNoOutput
				default:
					searchQuery.Media = db.OnlyTrue
				}

			// without media.
			case "nomedia":
				switch searchQuery.Media {
				case db.OnlyTrue:
					err = errNoOutput
				default:
					searchQuery.Media = db.OnlyFalse
				}

			// with poll.
			case "poll":
				switch searchQuery.Poll {
				case db.OnlyTrue:
					err = errNoOutput
				default:
					searchQuery.Poll = db.OnlyTrue
				}

			// without poll.
			case "nopoll":
				switch searchQuery.Poll {
				case db.OnlyTrue:
					err = errNoOutput
				default:
					searchQuery.Poll = db.OnlyFalse
				}

			default:
				// Set token as unused, gets
				// returned below as 'remain'.
				used = false
			}

		// query filter:
		// "show accounts / statuses only by relation ___"
		case strings.HasPrefix(token, "relation:"):
			token = token[len("relation:"):]
			token = remoteQuotes(token)
			token = strings.ToLower(token)
			switch token {

			// user is following
			case "following":
				switch searchQuery.Relationship {
				case db.RelationFollowers:
					err = errNoOutput
					return
				case db.RelationAny:
					searchQuery.Relationship = db.RelationFollowing
				}

			// user followed by
			case "followers":
				switch searchQuery.Relationship {
				case db.RelationFollowing:
					err = errNoOutput
				case db.RelationAny:
					searchQuery.Relationship = db.RelationFollowers
				}

			default:
				// Set token as unused, gets
				// returned below as 'remain'.
				used = false
			}

		default:
			// Set token as unused, gets
			// returned below as 'remain'.
			used = false
		}
		return
	})
	if err != nil {
		return nil, err
	}

	// Set author ID query from URL.
	searchQuery.AuthorID = authorID

	if follows {
		// If the separate URL query parameter
		// "follows" was set, ensure it doesn't
		// conflict with other provided params.
		switch searchQuery.Relationship {
		case db.RelationFollowers:
			return nil, errNoOutput
		case db.RelationAny:
			searchQuery.Relationship = db.RelationFollowing
		}
	}

	// If 'before:' and 'after:' were
	// provided, ensure they don't conflict.
	if !searchQuery.Before.IsZero() &&
		!searchQuery.After.IsZero() &&
		searchQuery.Before.After(searchQuery.After) {
		return nil, errNoOutput
	}

	// If author was provided,
	// ensure 'is:$location' or 'relation:' were NOT provided.
	if (searchQuery.AuthorID != "" || searchQuery.From != "") &&
		(searchQuery.Location != 0 || searchQuery.Relationship != 0) {
		return nil, errNoOutput
	}

	return searchQuery, nil
}

// parseQueryTokens is an internal function to be used by wrapping query parser functions, that
// parses tokens by whitespace separation, accounting for quoted sections, passing each token to
// callback. the returned string contains unquoted input minus any "used" tokens according to callback.
func parseQueryTokens(s string, each func(string) (used bool, err error)) (string, gtserror.WithCode) {
	if each == nil {
		panic("nil function")
	}

	var delim bool
	var quoted bool

	// Acquire buffer to
	// gather unused tokens
	// to return as remain.
	buf := buffers.Get1k()
	buf.B = buf.B[:0]
	buf.Guarantee(len(s))

	// Iterate through input string split by unicode spaces,
	// accounting for whether current rune is within double quotes.
	for token := range strings.FieldsFuncSeq(s, func(r rune) bool {
		switch r {
		case '\\':
			delim = !delim
			return false
		case '"':
			if !delim {
				quoted = !quoted
				return false
			} else {
				delim = false
				return false
			}
		}
		return !quoted && unicode.IsSpace(r)
	}) {
		// Pass to callback func.
		used, err := each(token)
		switch {
		case err != nil:
			return "", gtserror.NewErrorBadRequest(err, err.Error())
		case used:
			// ignore
		default:
			// Append to remain string.
			token = remoteQuotes(token)
			buf.B = append(buf.B, token...)
			buf.B = append(buf.B, ' ')
		}
	}

	// Drop last space.
	if len(buf.B) > 0 {
		buf.B = buf.B[:len(buf.B)-1]
	}

	// Get copy of
	// buffer string.
	s = string(buf.B)

	// Replace buffer.
	buffers.Put1k(buf)

	return s, nil
}

// parseTime attempts to parse a time value from given
// string, supporting a variety of time formats, and
// supports duration strings to represent "since" times.
func parseTime(s string) (time.Time, error) {
	if strings.ContainsRune(s, '.') {

		// Try parse a regular duration,
		// for a "since" defined stamp.
		d, err := time.ParseDuration(s)
		if err == nil {
			return time.Now().Add(-d), nil
		}
	} else {

		// Try parse a "long" duration,
		// for a "since" defined stamp.
		d, err := longdur.Parse(s)
		if err == nil {
			return d.SubTime(time.Now()), nil
		}
	}

	// Try parsing many time formats
	// to find one they have provided.
	for _, layout := range []string{
		time.DateTime,
		time.DateOnly,
		time.Stamp,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
	} {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t.Local(), nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid time: %q", s)
}

// removeQuotes removes OUTER double
// quote characters, nothing more.
//
// we leave most of it up to the database.
func remoteQuotes(s string) string {
	s = strings.TrimPrefix(s, `"`)
	s = strings.TrimSuffix(s, `"`)
	return s
}
