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

package bundb

import (
	"context"
	"errors"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type statusPinDB struct {
	db    *bun.DB
	state *state.State
}

func (s *statusPinDB) GetAccountStatusPins(ctx context.Context, accountID string) ([]string, error) {
	return s.state.Caches.DB.StatusPinnedIDs.Load(accountID, func() ([]string, error) {
		return getAccountStatusPins(ctx, s.db, accountID)
	})
}

func (s *statusPinDB) CountAccountStatusPins(ctx context.Context, accountID string) (int, error) {
	return s.state.Caches.DB.StatusPinnedIDs.Count(accountID, func() ([]string, error) {
		return getAccountStatusPins(ctx, s.db, accountID)
	})
}

func getAccountStatusPins(ctx context.Context, bundb *bun.DB, accountID string) ([]string, error) {
	var statusIDs []string
	err := bundb.NewSelect().
		Model((*gtsmodel.StatusPin)(nil)).
		Column("status_id").
		Where("? = ?", bun.Ident("account_id"), accountID).
		Order("created_at DESC").
		Scan(ctx, &statusIDs)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, err
	}
	return statusIDs, err
}

func (s *statusPinDB) DeleteAccountStatusPins(ctx context.Context, accountID string) error {
	_, err := s.db.NewDelete().
		Model((*gtsmodel.StatusPin)(nil)).
		Where("? = ?", bun.Ident("account_id"), accountID).
		Exec(ctx)
	s.state.Caches.DB.StatusPinnedIDs.Invalidate(accountID)
	return err
}

func (s *statusPinDB) IsStatusPinned(ctx context.Context, accountID, statusID string) (bool, error) {
	statusIDs, err := s.GetAccountStatusPins(ctx, accountID)
	if err != nil {
		return false, err
	}
	for _, id := range statusIDs {
		if id == statusID {
			return true, nil
		}
	}
	return false, nil
}

func (s *statusPinDB) PutStatusPin(ctx context.Context, pin *gtsmodel.StatusPin) error {
	_, err := s.db.NewInsert().
		Model(pin).
		Exec(ctx)
	s.state.Caches.DB.StatusPinnedIDs.Invalidate(pin.AccountID)
	return err
}

func (s *statusPinDB) DeleteStatusPin(ctx context.Context, statusID string) error {
	var accountID string
	_, err := s.db.NewDelete().
		Model((*gtsmodel.StatusPin)(nil)).
		Where("? = ?", bun.Ident("status_id"), statusID).
		Returning("?", bun.Ident("account_id")).
		Exec(ctx, &accountID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return err
	}
	s.state.Caches.DB.StatusPinnedIDs.Invalidate(accountID)
	return nil
}
