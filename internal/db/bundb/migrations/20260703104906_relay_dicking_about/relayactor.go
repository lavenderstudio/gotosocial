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

// smallint is the largest size supported
// by a PostgreSQL SMALLINT, since an SQLite
// SMALLINT is actually variable in size.
type smallint int16

// bitFieldType is the type we use
// for database int bit fields, at
// least where the smallest int size
// will suffice for number of fields.
type bitFieldType smallint

// RelayActor represents a *local* relay actor on
// this instance, created by an instance admin.
type RelayActor struct {
	// ID of this item in the database.
	// Creation time is encoded in the ID.
	ID string `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`

	// ID of the account that created this relay actor.
	CreatedByAccountID string `bun:"type:CHAR(26),notnull,nullzero"`

	// ActivityPub URI of the relay actor,
	// eg., `https://example.org/relays/some_relay_username`
	URI string `bun:",notnull,nullzero,unique"`

	// Flags contains numerous boolean
	// flags for this relay actor.
	// Default = relay public posts.
	Flags bitFieldType `bun:",notnull,default:2"`

	// Boolean flags for config
	// of the relay actor account.
	//
	// Default = WebShowFollowers true.
	ActorAccountFlags bitFieldType `bun:",notnull,default:2"`

	// IDs of matchers that apply
	// to this relay actor.
	MatcherIDs []string `bun:"matchers,array"`
}
