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

package db

import (
	"context"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// StatusPin encapsulates methods
// for (un)?pinning statuses.
type StatusPin interface {

	// IsStatusPinned returns whether given account ID has pinned given status ID.
	IsStatusPinned(ctx context.Context, accountID, statusID string) (bool, error)

	// PutStatusPin inserts a new status pin for given account and status ID.
	PutStatusPin(ctx context.Context, pin *gtsmodel.StatusPin) error

	// DeleteStatusPin deletes the given status pin matching status ID.
	DeleteStatusPin(ctx context.Context, statusID string) error

	// CountAccountStatusPins returns the number of statuses that given account ID has pinned.
	CountAccountStatusPins(ctx context.Context, accountID string) (int, error)

	// GetAccountStatusPins returns the status IDs that given account ID has pinned.
	GetAccountStatusPins(ctx context.Context, accountID string) ([]string, error)

	// DeleteAccountStatusPins unpins all statuses for given account ID.
	DeleteAccountStatusPins(ctx context.Context, accountID string) error
}
