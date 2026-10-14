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
	"net/url"
	"strings"

	"code.superseriousbusiness.org/gotosocial/internal/util"
)

func tryParseURL(s string) *url.URL {
	if !strings.Contains(s, "://") {
		// a lot of strings can parse as a valid URL,
		// so we limit it to only those with a scheme.
		return nil
	}

	// Parse as a URL.
	u, _ := url.Parse(s)
	return u
}

func tryParseUsernameDomain(s string) (string, string, bool) {
	if !strings.Contains(s, "@") {
		// only consider it a namestring
		// if it contains an '@'
		return "", "", false
	}

	// So it for-sure contains an
	// '@', ensure a prefixing one.
	if !strings.HasPrefix(s, "@") {
		s = "@" + s
	}

	username, domain, err := util.ExtractNamestringParts(s)
	return username, domain, (err == nil)
}
