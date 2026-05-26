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

import "time"

// Token is a translation of the gotosocial token
// with the ExpiresIn fields replaced with ExpiresAt.
type Token struct {
	// ID of this item in the database.
	ID string `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`

	// User-set name for this token.
	Name string `bun:",nullzero"`

	// Approximate time when this token was last used.
	LastUsed time.Time `bun:"type:timestamptz,nullzero"`

	// ID of the client who owns this token.
	ClientID string `bun:"type:CHAR(26),nullzero,notnull"`

	// ID of the user who owns this token.
	UserID string `bun:"type:CHAR(26),nullzero"`

	// Oauth redirect URI for this token.
	RedirectURI string `bun:",nullzero,notnull"`

	// Oauth scope.
	Scope string `bun:",nullzero,notnull,default:'read'"`

	// Code, if present.
	Code string `bun:",pk,nullzero,notnull,default:''"`

	// Code challenge, if code present.
	CodeChallenge string `bun:",nullzero"`

	// Code challenge method, if code present.
	CodeChallengeMethod string `bun:",nullzero"`

	// Code created time, if code present.
	CodeCreateAt time.Time `bun:"type:timestamptz,nullzero"`

	// Code expires at -- null means the code never expires.
	CodeExpiresAt time.Time `bun:"type:timestamptz,nullzero"`

	// User level access token, if present.
	Access string `bun:",pk,nullzero,notnull,default:''"`

	// User level access token created time, if access present.
	AccessCreateAt time.Time `bun:"type:timestamptz,nullzero"`

	// User level access token expires at -- null means the token never expires.
	AccessExpiresAt time.Time `bun:"type:timestamptz,nullzero"`

	// Refresh token, if present.
	Refresh string `bun:",pk,nullzero,notnull,default:''"`

	// Refresh created at, if refresh present.
	RefreshCreateAt time.Time `bun:"type:timestamptz,nullzero"`

	// Refresh expires at -- null means the refresh token never expires.
	RefreshExpiresAt time.Time `bun:"type:timestamptz,nullzero"`
}
