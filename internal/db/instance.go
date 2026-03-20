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
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

// Instance contains functions for instance-level actions (counting instance users etc.).
type Instance interface {
	// CountInstanceAccounts returns the number of non-suspended accounts on this instance.
	CountInstanceAccounts(ctx context.Context) (int, error)

	// CountInstancePeers returns the number of statuses on this instance.
	CountInstanceStatuses(ctx context.Context) (int, error)

	// CountInstancePeers returns the number of instances that this instance peers aka federates with.
	CountInstancePeers(ctx context.Context) (int, error)

	// GetInstance returns the instance entry for the given domain, if it exists.
	GetInstance(ctx context.Context, domain string) (*gtsmodel.Instance, error)

	// GetInstanceByID returns the instance entry corresponding to the given id, if it exists.
	GetInstanceByID(ctx context.Context, id string) (*gtsmodel.Instance, error)

	// PutInstance inserts the given instance into the database.
	PutInstance(ctx context.Context, instance *gtsmodel.Instance) error

	// GetInstancesPage gets a page of instances with the given parameters.
	GetInstancesPage(ctx context.Context, page *paging.Page, domain string, orderBy gtsmodel.InstanceOrderBy, withErrorsOnly bool) ([]*gtsmodel.Instance, error)

	// AddInstanceDeliveryError adds the given instance delivery error message
	// to the instance delivery errors field for the given domain, if it exists.
	AddInstanceDeliveryError(ctx context.Context, domain string, errMsg string) error

	// SetInstanceSuccessfulDelivery updates the LatestSuccessfulDelivery time on the
	// instance entry for the given domain to time.Now() and clears stored delivery errors.
	SetInstanceSuccessfulDelivery(ctx context.Context, domain string) error

	// GetInstanceAccounts returns a slice of accounts from the given instance, arranged by ID.
	GetInstanceAccounts(ctx context.Context, domain string, maxID string, limit int) ([]*gtsmodel.Account, error)

	// GetInstancePeers returns a slice of instances that the host instance knows about.
	GetInstancePeers(ctx context.Context, includeSuspended bool) ([]*gtsmodel.Instance, error)

	// GetInstanceModeratorAddresses returns a slice of email addresses belonging to active
	// (as in, not suspended) moderators + admins on this instance.
	GetInstanceModeratorAddresses(ctx context.Context) ([]string, error)

	// GetInstanceModerators returns a slice of accounts belonging to active
	// (as in, non suspended) moderators + admins on this instance.
	GetInstanceModerators(ctx context.Context) ([]*gtsmodel.Account, error)
}
