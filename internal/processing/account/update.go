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

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// Update processes the update of
// an account with the given form.
func (p *Processor) Update(
	ctx context.Context,
	account *gtsmodel.Account,
	form *apimodel.UpdateAccountRequest,
) (*apimodel.Account, gtserror.WithCode) {
	// Need settings set
	// for StatusContentType.
	if account.Settings == nil {
		var err error
		account.Settings, err = p.state.DB.GetAccountSettings(ctx, account.ID)
		if err != nil {
			err := gtserror.Newf("db error getting account settings: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	if errWithCode := p.c.UpdateAccount(ctx,
		account,
		form,
		account.Settings.StatusContentType,
	); errWithCode != nil {
		return nil, errWithCode
	}

	acctSensitive, err := p.converter.AccountToAPIAccountSensitive(ctx, account)
	if err != nil {
		err := gtserror.Newf("error converting account: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return acctSensitive, nil
}
