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

package admin

import (
	"context"
	"fmt"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/messages"
)

func (p *Processor) AccountAction(
	ctx context.Context,
	admin *gtsmodel.Account,
	request *apimodel.AdminActionRequest,
) (
	string,
	gtserror.WithCode,
) {
	target, err := p.state.DB.GetAccountByID(ctx, request.TargetID)
	if err != nil {
		err := gtserror.Newf("db error getting target account: %w", err)
		return "", gtserror.NewErrorInternalError(err)
	}

	switch gtsmodel.ParseAdminActionType(request.Type) {
	case gtsmodel.AdminActionSuspend:
		return p.accountActionSuspend(ctx, admin, target, request.Text)

	default:
		// TODO: add more types to this slice when adding
		//       more types to the switch statement above.
		supportedTypes := []string{
			gtsmodel.AdminActionSuspend.String(),
		}

		err := fmt.Errorf(
			"admin action type %s is not supported for this endpoint, "+
				"currently supported types are: %q",
			request.Type, supportedTypes)

		return "", gtserror.NewErrorBadRequest(err, err.Error())
	}
}

func (p *Processor) accountActionSuspend(
	ctx context.Context,
	admin *gtsmodel.Account,
	target *gtsmodel.Account,
	text string,
) (
	string,
	gtserror.WithCode,
) {
	actionID := id.NewULID()
	errWithCode := p.state.AdminActions.Run(
		ctx,
		&gtsmodel.AdminAction{
			ID:             actionID,
			TargetCategory: gtsmodel.AdminActionCategoryAccount,
			TargetID:       target.ID,
			Target:         target,
			Type:           gtsmodel.AdminActionSuspend,
			AccountID:      admin.ID,
			Text:           text,
		},
		func(ctx context.Context) gtserror.MultiError {
			if err := p.state.Workers.Client.Process(
				ctx,
				&messages.FromClientAPI{
					APObjectType:   ap.ActorPerson,
					APActivityType: ap.ActivityDelete,
					Origin:         admin,
					Target:         target,
				},
			); err != nil {
				errs := gtserror.NewMultiError(1)
				errs.Append(err)
				return errs
			}
			return nil
		},
	)
	return actionID, errWithCode
}
