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

package bundb_test

import (
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"github.com/stretchr/testify/suite"
)

type DirectoryTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *DirectoryTestSuite) TestDirectory() {
	ctx := suite.T().Context()

	// Ensure account stats populated
	// for each account in the test rig.
	for _, account := range suite.testAccounts {
		if err := suite.state.DB.PopulateAccountStats(ctx, account); err != nil {
			suite.FailNow(err.Error())
		}
	}

	// Get from the top.
	accounts, err := suite.state.DB.GetDirectoryPage(ctx,
		&paging.Page{
			Limit: 20,
		},
		0, // offset
		gtsmodel.DirectoryOrderByActive,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.Len(accounts, 2)

	// Get with offset of 1.
	accounts, err = suite.state.DB.GetDirectoryPage(ctx,
		&paging.Page{
			Limit: 20,
		},
		1, // offset
		gtsmodel.DirectoryOrderByActive,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.Len(accounts, 1)

	// Get with paging params.
	accounts, err = suite.state.DB.GetDirectoryPage(ctx,
		&paging.Page{
			Limit: 20,
			Max:   paging.MaxID(suite.testAccounts["local_account_1"].ID),
		},
		0, // offset
		gtsmodel.DirectoryOrderByActive,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.Len(accounts, 1)

	// Get ordered by creation time.
	accounts, err = suite.state.DB.GetDirectoryPage(ctx,
		&paging.Page{
			Limit: 20,
		},
		0, // offset
		gtsmodel.DirectoryOrderByNew,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.Len(accounts, 2)
}

func TestDirectoryTestSuite(t *testing.T) {
	suite.Run(t, new(DirectoryTestSuite))
}
