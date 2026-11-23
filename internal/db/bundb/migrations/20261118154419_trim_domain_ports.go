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

package migrations

import (
	"context"

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/netutil"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		for _, table := range []string{
			"domain_allows",
			"domain_blocks",
			"domain_permission_drafts",
			"domain_permission_excludes",
		} {
			if err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
				var domains []string

				// Select table domains.
				if err := tx.NewSelect().
					Table(table).
					Column("domain").
					Scan(ctx, &domains); err != nil {
					return gtserror.Newf("error scanning domains: %w", err)
				}

				// Track changed domains
				// for later update query.
				changed := make([]struct {
					old string `bun:"old"`
					new string `bun:"new"`
				}, 0, len(domains))

				// Trim all domains of their port,
				// if any, storing those that change.
				for _, domain := range domains {
					if d := netutil.TrimPort(domain); //
					d != domain {
						changed = append(changed, struct {
							old string `bun:"old"`
							new string `bun:"new"`
						}{
							old: domain,
							new: d,
						})
					}
				}

				if len(changed) > 0 {
					// Update all changed entries.
					if _, err := tx.NewUpdate().
						With("changed", tx.NewValues(changed)).
						Set("domain = changed.new").
						Where("domain = changed.old").
						Exec(ctx); err != nil {
						return gtserror.Newf("error updating domain(s): %w", err)
					}
				}

				return nil
			}); err != nil {
				return gtserror.Newf("error running %s migration: %w", table, err)
			}
		}
		return nil
	}

	down := func(ctx context.Context, db *bun.DB) error {
		return nil
	}

	if err := Migrations.Register(up, down); err != nil {
		panic(err)
	}
}
