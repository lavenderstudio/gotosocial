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

package account

import (
	"context"
	"errors"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/messages"
)

// RemoveFromFollowers removes targetAccountID from
// requesting account's followers collection (if present).
func (p *Processor) RemoveFromFollowers(
	ctx context.Context,
	requester *gtsmodel.Account,
	targetAccountID string,
) (*apimodel.Relationship, gtserror.WithCode) {
	// Get target account from the db.
	targetAccount, err := p.state.DB.GetAccountByID(ctx, targetAccountID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("db error getting account: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if targetAccount == nil {
		err := gtserror.New("account not found")
		return nil, gtserror.NewErrorNotFound(err)
	}

	// Check if a follow exists from
	// targetAccountID -> requester.
	follow, err := p.state.DB.GetFollow(
		gtscontext.SetBarebones(ctx),
		targetAccountID,
		requester.ID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("db error checking existing follow: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if follow != nil {
		// Follow exists, remove it.
		if err := p.state.DB.DeleteFollow(ctx,
			follow.AccountID,
			follow.TargetAccountID,
		); err != nil {
			err = gtserror.Newf("db error removing follow: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		// Handle side effects async.
		p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
			APObjectType:   ap.ActivityFollow,
			APActivityType: ap.ActivityReject,
			GTSModel:       follow,
			Origin:         requester,
			Target:         targetAccount,
		})
	}

	// Return the (changed) relationship state.
	return p.RelationshipGet(ctx, requester, targetAccountID)
}
