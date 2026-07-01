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
	"errors"
	"slices"
	"strings"

	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/oauth2/v4"
)

// Auth wraps an authorized token, application, user, and account.
// It is used in the functions GetAuthed and MustAuth.
// Because the user might *not* be authed, any of the fields in this struct
// might be nil, so make sure to check that when you're using this struct anywhere.
type Auth struct {
	Token       oauth2.TokenInfo
	Application *gtsmodel.Application
	User        *gtsmodel.User
	Account     *gtsmodel.Account
}

type AuthRequirements struct {
	Token   bool
	App     bool
	User    bool
	Account bool
	Scope   []Scope
}

// TokenAuth is a convenience function for returning an TokenAuth struct from a gin context.
// In essence, it tries to extract a token, application, user, and account from the context,
// and then sets them on a struct for convenience.
//
// If any are not present in the context, they will be set to nil on the returned TokenAuth struct.
//
// If *ALL* are not present, then nil and an error will be returned.
//
// If something goes wrong during parsing, then nil and an error will be returned (consider this not authed).
// TokenAuth is like GetAuthed, but will fail if one of the requirements is not met.
func TokenAuth(c *httputil.Context, r AuthRequirements) (*Auth, gtserror.WithCode) {
	a := new(Auth)

	a.Token, _ = c.V.Get(oauth.SessionAuthorizedToken).(oauth2.TokenInfo)
	a.Application, _ = c.V.Get(oauth.SessionAuthorizedApplication).(*gtsmodel.Application)
	a.User, _ = c.V.Get(oauth.SessionAuthorizedUser).(*gtsmodel.User)
	a.Account, _ = c.V.Get(oauth.SessionAuthorizedAccount).(*gtsmodel.Account)

	if r.Token && a.Token == nil {
		const errText = "token not supplied"
		return nil, gtserror.NewErrorUnauthorized(errors.New(errText), errText)
	}

	if r.App && a.Application == nil {
		const errText = "application not supplied"
		return nil, gtserror.NewErrorUnauthorized(errors.New(errText), errText)
	}

	if r.User && a.User == nil {
		const errText = "user not supplied or not authorized"
		return nil, gtserror.NewErrorUnauthorized(errors.New(errText), errText)
	}

	if r.Account && a.Account == nil {
		const errText = "account not supplied or not authorized"
		return nil, gtserror.NewErrorUnauthorized(errors.New(errText), errText)
	}

	if len(r.Scope) != 0 {
		// We need to match one of the
		// required scopes, check if we can.
		hasScopes := strings.Split(a.Token.GetScope(), " ")
		scopeOK := slices.ContainsFunc(
			hasScopes,
			func(hasScope string) bool {
				for _, requiredScope := range r.Scope {
					if Scope(hasScope).Permits(requiredScope) {
						// Got it.
						return true
					}
				}
				return false
			},
		)

		if !scopeOK {
			const errText = "token has insufficient scope permission"
			return nil, gtserror.NewErrorForbidden(errors.New(errText), errText)
		}
	}

	return a, nil
}
