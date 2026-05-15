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

package media

import (
	"context"
	"fmt"

	"code.superseriousbusiness.org/gotosocial/internal/cleaner"
	"code.superseriousbusiness.org/gotosocial/internal/db/bundb"
	"code.superseriousbusiness.org/gotosocial/internal/media"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	gtsstorage "code.superseriousbusiness.org/gotosocial/internal/storage"
)

type prune struct {
	manager *media.Manager
	cleaner *cleaner.Cleaner
	state   *state.State
}

func setupPrune(ctx context.Context) (*prune, error) {
	var state state.State
	state.Caches.Init()

	err := state.Caches.Start()
	if err != nil {
		return nil, fmt.Errorf("error starting caches: %w", err)
	}

	// Set state DB connection.
	// Don't need Actions for this.
	state.DB, err = bundb.NewBunDBService(ctx, &state)
	if err != nil {
		return nil, fmt.Errorf("error starting database: %w", err)
	}

	// Scheduler is required for the
	// cleaner, but no other workers
	// are needed for this CLI action.
	state.Workers.StartScheduler()

	//nolint:contextcheck
	state.Storage, err = gtsstorage.AutoConfig()
	if err != nil {
		return nil, fmt.Errorf("error creating storage backend: %w", err)
	}

	//nolint:contextcheck
	manager := media.NewManager(&state)

	//nolint:contextcheck
	cleaner := cleaner.New(&state)

	return &prune{
		manager: manager,
		cleaner: cleaner,
		state:   &state,
	}, nil
}

func (p *prune) shutdown() error {
	var err error

	if err = p.state.DB.Close(); err != nil {
		err = fmt.Errorf("error stopping database: %w", err)
	}

	p.state.Workers.Scheduler.Stop()
	p.state.Caches.Stop()

	return err
}
