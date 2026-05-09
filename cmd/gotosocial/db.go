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

package main

import (
	"code.superseriousbusiness.org/gotosocial/cmd/gotosocial/action/database"
	"github.com/spf13/cobra"
)

// databaseCommands returns the 'database' subcommand.
func databaseCommands() *cobra.Command {
	databaseCmd := &cobra.Command{
		Use:   "database",
		Short: "gotosocial database-releated tasks",
	}

	// SQLite commands only set if !nosqlite tag.
	if sqlite := sqliteCommands(); sqlite != nil {
		databaseCmd.AddCommand(sqlite)
	}

	databaseCmd.AddCommand(&cobra.Command{
		Use:   "ping",
		Short: "perform a database \"ping\"",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), database.Ping)
		},
	})

	return databaseCmd
}
