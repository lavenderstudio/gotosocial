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

package util

import (
	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
)

// JSONAcceptHeaders is a slice of offers that just contains application/json types.
var JSONAcceptHeaders = []string{
	AppJSON,
}

// WebfingerJSONAcceptHeaders is a slice of offers that prefers the
// jrd+json content type, but will be chill and fall back to app/json.
// This is to be used specifically for webfinger responses.
// See https://www.rfc-editor.org/rfc/rfc7033#section-10.2
var WebfingerJSONAcceptHeaders = []string{
	AppJRDJSON,
	AppJSON,
}

// JSONOrHTMLAcceptHeaders is a slice of offers that prefers AppJSON and will
// fall back to HTML if necessary. This is useful for error handling, since it can
// be used to serve a nice HTML page if the caller accepts that, or just JSON if not.
var JSONOrHTMLAcceptHeaders = []string{
	AppJSON,
	TextHTML,
}

// HTMLAcceptHeaders is a slice of offers that just contains text/html types.
var HTMLAcceptHeaders = []string{
	TextHTML,
}

// HTMLOrActivityPubHeaders matches text/html first, then activitypub types.
// This is useful for user URLs that a user might go to in their browser,
// but which should also be able to serve ActivityPub as a fallback.
//
// https://www.w3.org/TR/activitypub/#retrieving-objects
var HTMLOrActivityPubHeaders = []string{
	TextHTML,
	AppActivityLDJSON,
	AppActivityJSON,
}

// ActivityPubOrHTMLHeaders matches activitypub types first, then text/html.
// This is useful for URLs that should serve ActivityPub by default, but
// which a user might also go to in their browser sometimes.
//
// https://www.w3.org/TR/activitypub/#retrieving-objects
var ActivityPubOrHTMLHeaders = []string{
	AppActivityLDJSON,
	AppActivityJSON,
	TextHTML,
}

// ActivityPubHeaders matches only activitypub Accept headers.
// This is useful for URLs should only serve ActivityPub.
//
// https://www.w3.org/TR/activitypub/#retrieving-objects
var ActivityPubHeaders = []string{
	AppActivityLDJSON,
	AppActivityJSON,
}

var HostMetaHeaders = []string{
	AppXMLXRD,
	AppXML,
}

// CSVHeaders just contains the text/csv
// MIME type, used for import/export.
var CSVHeaders = []string{
	TextCSV,
}

// NegotiateAccept wraps httputil.NegotiateAccept() to include a gtserror.WithCode{} status code.
func NegotiateAccept(c *httputil.Context, offers ...string) (string, gtserror.WithCode) {
	contentType, err := httputil.NegotiateAccept(c, offers...)
	if err != nil {
		return "", gtserror.NewErrorNotAcceptable(err, err.Error())
	}
	return contentType, nil
}
