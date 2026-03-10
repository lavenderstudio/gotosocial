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

// StatusPin represents a Status
// that has been pinned by its author.
type StatusPin struct {

	// The database ID of the status that is pinned,
	// only one pin can exist for status so is PK.
	StatusID string `bun:"type:CHAR(26),pk,nullzero,notnull"`

	// The database ID of the account pinning the status,
	// will always be `statuses.account_id`, but makes
	// select and delete queries easier and faster.
	AccountID string `bun:"type:CHAR(26),nullzero,notnull"`

	// Creation time of the status pin.
	CreatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`
}
