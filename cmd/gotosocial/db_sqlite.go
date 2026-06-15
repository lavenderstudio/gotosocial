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

//go:build !nosqlite

package main

import (
	"code.superseriousbusiness.org/gotosocial/cmd/gotosocial/action/database"
	"github.com/spf13/cobra"
)

// sqliteCommands returns the 'sqlite' subcommand
func sqliteCommands() *cobra.Command {
	sqliteCmd := &cobra.Command{
		Use:   "sqlite",
		Short: "sqlite-specific tasks",
	}

	// sqliteCmd.AddCommand(&cobra.Command{
	// 	Use: "vacuum",
	// 	Short: "performs an sqlite \"vacuum\" operation; " +
	// 		"please note this requires AT LEAST db size free disk-space",
	// 	PreRunE: func(cmd *cobra.Command, args []string) error {
	// 		return preRun(preRunArgs{cmd: cmd})
	// 	},
	// 	RunE: func(cmd *cobra.Command, args []string) error {
	// 		return run(cmd.Context(), database.SQLiteVacuum)
	// 	},
	// })

	sqliteCmd.AddCommand(&cobra.Command{
		Use:   "analyze",
		Short: "performs an sqlite \"analyze\" operation",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), database.SQLiteAnalyze)
		},
	})

	return sqliteCmd
}
