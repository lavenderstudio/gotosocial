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

package status

import (
	"context"
	"errors"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/messages"
)

// Delete processes the delete of a given status, returning the deleted status if the delete goes through.
func (p *Processor) Delete(ctx context.Context, requester *gtsmodel.Account, statusID string) (*apimodel.Status, gtserror.WithCode) {

	// Fetch status with given ID, ensuring it is requester's status.
	status, errWithCode := p.c.GetOwnStatus(ctx, requester, statusID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Check if it is deleted.
	if status.Flags.Deleted() {
		const text = "target status not found"
		return nil, gtserror.NewErrorNotFound(
			errors.New(text),
			text,
		)
	}

	// Convert the status to API model BEFORE deleting it.
	apiStatus, errWithCode := p.c.GetAPIStatus(ctx, requester, status)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Replace content warning with raw
	// version if it's available, to make
	// delete + redraft work nicer.
	if status.ContentWarningText != "" {
		apiStatus.SpoilerText = status.ContentWarningText
	}

	// Process delete side effects.
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		APObjectType:   ap.ObjectNote,
		APActivityType: ap.ActivityDelete,
		GTSModel:       status,
		Origin:         requester,
		Target:         requester,
	})

	return apiStatus, nil
}
