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

package middleware

import (
	"fmt"
	"net/url"

	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gopkg/log"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/gorilla/sessions"
	"github.com/quasoft/memstore"
)

type sessionkey struct{}

// GetSession fetches stored session from provided context.
func GetSession(c *httputil.Context) *sessions.Session {
	session, _ := c.V.Get(sessionkey{}).(*sessions.Session)
	return session
}

// SessionOptions returns the standard set of options to use for each session.
func SessionOptions(cookiePolicy apiutil.CookiePolicy) *sessions.Options {
	return &sessions.Options{
		Path:     "/",
		Domain:   cookiePolicy.Domain,
		MaxAge:   120, // 2 minutes
		Secure:   cookiePolicy.Secure,
		HttpOnly: cookiePolicy.HTTPOnly,
		SameSite: cookiePolicy.SameSite,
	}
}

// SessionName is a utility function that derives
// an appropriate session name from the hostname.
func SessionName() (string, error) {

	// Parse the combined protocol + host.
	u, err := url.Parse(config.GetProtocol() + "://" + config.GetHost())
	if err != nil {
		return "", err
	}

	// Use hostname
	// without port.
	host := u.Hostname()
	if host == "" {
		return "", fmt.Errorf("could not derive hostname without port from %s://%s", config.GetProtocol(), config.GetHost())
	}

	// make sure IDNs are converted to punycode or the cookie library breaks:
	// see https://en.wikipedia.org/wiki/Punycode
	punyhost, err := util.Punify(host)
	if err != nil {
		return "", fmt.Errorf("could not convert %q to punycode: %v", host, err)
	}

	return "gotosocial-" + punyhost, nil
}

// Session returns a new gin middleware that implements session cookies using the given sessionName, authentication
// key, and encryption key. Session name can be derived from the SessionName utility function in this package.
func Session(sessionName string, auth []byte, crypt []byte, cookiePolicy apiutil.CookiePolicy) httputil.FlatMiddlewareFunc {
	store := memstore.NewMemStore(auth, crypt)
	store.Options = SessionOptions(cookiePolicy)
	return func(c *httputil.Context) {

		// Get / create new session with given name.
		session, err := store.Get(c.R, sessionName)
		if err != nil {
			log.Errorf(c, "error getting session: %v", err)
			return
		}

		// Store session in request ctx.
		c.V.Set(sessionkey{}, session)
	}
}
