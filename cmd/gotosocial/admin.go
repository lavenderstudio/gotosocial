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
	"code.superseriousbusiness.org/gotosocial/cmd/gotosocial/action/admin/account"
	"code.superseriousbusiness.org/gotosocial/cmd/gotosocial/action/admin/media"
	"code.superseriousbusiness.org/gotosocial/cmd/gotosocial/action/admin/statuses"
	"code.superseriousbusiness.org/gotosocial/cmd/gotosocial/action/admin/trans"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"github.com/spf13/cobra"
)

func adminCommands() *cobra.Command {
	adminCmd := &cobra.Command{
		Use:   "admin",
		Short: "gotosocial admin-related tasks",
	}

	/*
	   ADMIN ACCOUNT COMMANDS
	*/

	adminAccountCmd := add(adminCmd, &cobra.Command{
		Use:   "account",
		Short: "admin commands related to local (this instance) accounts",
	})

	_ = add(adminAccountCmd, &cobra.Command{
		Use:   "create",
		Short: "create a new local account",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), account.Create)
		},
	},
		config.AddAdminAccountCreate,
	)

	_ = add(adminAccountCmd, &cobra.Command{
		Use:   "list",
		Short: "list all existing local accounts",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), account.List)
		},
	})

	_ = add(adminAccountCmd, &cobra.Command{
		Use:   "confirm",
		Short: "confirm an existing local account manually, thereby skipping email confirmation",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), account.Confirm)
		},
	},
		config.AddAdminAccount,
	)

	_ = add(adminAccountCmd, &cobra.Command{
		Use:   "promote",
		Short: "promote a local account to admin",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), account.Promote)
		},
	},
		config.AddAdminAccount,
	)

	_ = add(adminAccountCmd, &cobra.Command{
		Use:   "demote",
		Short: "demote a local account from admin to normal user",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), account.Demote)
		},
	},
		config.AddAdminAccount,
	)

	_ = add(adminAccountCmd, &cobra.Command{
		Use:   "disable",
		Short: "set 'disabled' to true on a local account to prevent it from signing in or posting etc, but don't delete anything",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), account.Disable)
		},
	},
		config.AddAdminAccount,
	)

	_ = add(adminAccountCmd, &cobra.Command{
		Use:   "enable",
		Short: "undo a previous disable command by setting 'disabled' to false on a local account",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), account.Enable)
		},
	},
		config.AddAdminAccount,
	)

	_ = add(adminAccountCmd, &cobra.Command{
		Use:   "password",
		Short: "set a new password for the given local account",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), account.Password)
		},
	},
		config.AddAdminAccount,
		config.AddAdminAccountPassword,
	)

	_ = add(adminAccountCmd, &cobra.Command{
		Use:   "disable-2fa",
		Short: "disable 2fa for the given local account",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), account.Disable2FA)
		},
	},
		config.AddAdminAccount,
	)

	/*
	   ADMIN IMPORT/EXPORT COMMANDS
	*/

	_ = add(adminCmd, &cobra.Command{
		Use:   "export",
		Short: "export data from the database to file at the given path",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), trans.Export)
		},
	},
		config.AddAdminTrans,
	)

	_ = add(adminCmd, &cobra.Command{
		Use:   "import",
		Short: "import data from a file into the database",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), trans.Import)
		},
	},
		config.AddAdminTrans,
	)

	/*
		ADMIN MEDIA COMMANDS
	*/

	adminMediaCmd := add(adminCmd, &cobra.Command{
		Use:   "media",
		Short: "admin commands related to stored media / emojis",
	})

	/*
		ADMIN MEDIA LIST COMMANDS
	*/

	_ = add(adminMediaCmd, &cobra.Command{
		Use:   "list-attachments",
		Short: "list local, remote, or all attachments",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), media.ListAttachments)
		},
	},
		config.AddAdminMediaList,
	)

	_ = add(adminMediaCmd, &cobra.Command{
		Use:   "list-emojis",
		Short: "list local, remote, or all emojis",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), media.ListEmojis)
		},
	},
		config.AddAdminMediaList,
	)

	/*
		ADMIN MEDIA PRUNE COMMANDS
	*/
	adminMediaPruneCmd := add(adminMediaCmd, &cobra.Command{
		Use:   "prune",
		Short: "admin commands for pruning media from storage",
	})

	_ = add(adminMediaPruneCmd, &cobra.Command{
		Use:   "orphaned",
		Short: "prune orphaned media from storage",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), media.PruneOrphaned)
		},
	},
		config.AddAdminMediaPrune,
	)

	_ = add(adminMediaPruneCmd, &cobra.Command{
		Use:   "remote",
		Short: "prune unused / stale media from storage, older than given duration",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), media.PruneRemote)
		},
	},
		config.AddAdminMediaPrune,
	)

	_ = add(adminMediaPruneCmd, &cobra.Command{
		Use:   "all",
		Short: "perform all media and emoji prune / cleaning commands",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), media.PruneAll)
		},
	},
		config.AddAdminMediaPrune,
	)

	/*
		ADMIN STATUS COMMANDS
	*/

	adminStatusesCmd := add(adminCmd, &cobra.Command{
		Use:   "statuses",
		Short: "admin commands related to stored statuses",
	})

	/*
		ADMIN STATUS PRUNE COMMANDS
	*/

	adminStatusesPruneCmd := add(adminStatusesCmd, &cobra.Command{
		Use:   "prune",
		Short: "admin commands for pruning statuses from the database",
	})

	_ = add(adminStatusesPruneCmd, &cobra.Command{
		Use:   "remote",
		Short: "prune old, locally-not-interacted-with remote status threads from the database",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRun(preRunArgs{cmd: cmd})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), statuses.PruneOldRemote)
		},
	})

	return adminCmd
}
