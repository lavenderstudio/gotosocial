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

// Instance represents a
// single federated instance.
type Instance struct {
	// ID of this item in the database.
	ID string `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`

	// Instance domain,
	// eg., example.org
	Domain string `bun:",nullzero,notnull,unique"`

	// Software deployed for this
	// instance, eg., "mastodon".
	Software string `bun:",nullzero"`

	// Time of latest *SUCCESSFUL* attempt
	// to deliver a message to this instance.
	LatestSuccessfulDelivery time.Time `bun:"type:timestamptz,nullzero"`

	// Recent delivery errors.
	//
	// Not stored in the db.
	DeliveryErrors []*FederationError `bun:"-"`
}

// InstanceOrderBy is for doing db
// queries for admin view of instances
type InstanceOrderBy enumType

const (
	InstanceOrderByUnknown InstanceOrderBy = iota
	// Order alphabetically (a -> z).
	InstanceOrderByAlphabetical
	// Order by date instance first seen (newest -> oldest).
	InstanceOrderByFirstSeen
)

func (d InstanceOrderBy) String() string {
	switch d {
	case InstanceOrderByAlphabetical:
		return "alphabetical"
	case InstanceOrderByFirstSeen:
		return "first_seen"
	default:
		return "unknown"
	}
}
