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

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"github.com/stretchr/testify/suite"
)

type SearchTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *SearchTestSuite) TestSearchAccountsTurtle() {
	accounts, err := suite.db.SearchForAccounts(suite.T().Context(), db.AccountSearchParams{
		UsernameQuery: "turtle",
	}, toPage("", "", "", 10))
	suite.NoError(err)
	suite.Len(accounts, 1)
}

func (suite *SearchTestSuite) TestSearchAccounts1Happy() {
	accounts, err := suite.db.SearchForAccounts(suite.T().Context(), db.AccountSearchParams{
		UsernameQuery: "1happy",
	}, toPage("", "", "", 10))
	suite.NoError(err)
	suite.Len(accounts, 1)
}

func (suite *SearchTestSuite) TestSearchAccountsTurtleFollowing() {
	testAccount := suite.testAccounts["local_account_1"]
	accounts, err := suite.db.SearchForAccounts(suite.T().Context(), db.AccountSearchParams{
		CommonSearchParams: db.CommonSearchParams{
			RequesterID:  testAccount.ID,
			Relationship: db.RelationFollowing,
		},
		UsernameQuery: "turtle",
	}, toPage("", "", "", 10))
	suite.NoError(err)
	suite.Len(accounts, 1)
}

func (suite *SearchTestSuite) TestSearchAccountsTurtleFollowingUpper() {
	testAccount := suite.testAccounts["local_account_1"]
	accounts, err := suite.db.SearchForAccounts(suite.T().Context(), db.AccountSearchParams{
		CommonSearchParams: db.CommonSearchParams{
			RequesterID:  testAccount.ID,
			Relationship: db.RelationFollowing,
		},
		UsernameQuery: "TURTLE",
	}, toPage("", "", "", 10))
	suite.NoError(err)
	suite.Len(accounts, 1)
}

func (suite *SearchTestSuite) TestSearchAccountsPost() {
	accounts, err := suite.db.SearchForAccounts(suite.T().Context(), db.AccountSearchParams{
		NoteQuery: "post",
	}, toPage("", "", "", 10))
	suite.NoError(err)
	suite.Len(accounts, 2)
}

func (suite *SearchTestSuite) TestSearchAccountsFoss() {
	accounts, err := suite.db.SearchForAccounts(suite.T().Context(), db.AccountSearchParams{
		UsernameQuery: "foss",
	}, toPage("", "", "", 10))
	suite.NoError(err)
	suite.Len(accounts, 1)
}

func (suite *SearchTestSuite) TestSearchStatuses() {
	testAccount := suite.testAccounts["local_account_1"]
	statuses, err := suite.db.SearchForStatuses(suite.T().Context(), db.StatusSearchParams{
		CommonSearchParams: db.CommonSearchParams{RequesterID: testAccount.ID},
		ContentQuery:       "hello",
	}, toPage("", "", "", 10))
	suite.NoError(err)
	suite.Len(statuses, 2)
}

func (suite *SearchTestSuite) TestSearchStatusesFromAccount() {
	testAccount := suite.testAccounts["local_account_1"]
	fromAccount := suite.testAccounts["local_account_2"]
	fromAccountID := fromAccount.ID
	statuses, err := suite.db.SearchForStatuses(suite.T().Context(), db.StatusSearchParams{
		CommonSearchParams: db.CommonSearchParams{RequesterID: testAccount.ID},
		AuthorID:           fromAccountID,
		ContentQuery:       "hi",
	}, toPage("", "", "", 10))
	suite.NoError(err)
	suite.Len(statuses, 4)
	suite.Equal(fromAccount.ID, statuses[0].AccountID)
}

func (suite *SearchTestSuite) TestSearchTags() {
	// Search with full tag string.
	tags, err := suite.db.SearchForTags(suite.T().Context(), "welcome", toPage("", "", "", 10))
	suite.NoError(err)
	suite.Len(tags, 1)

	// Search with partial tag string.
	tags, err = suite.db.SearchForTags(suite.T().Context(), "wel", toPage("", "", "", 10))
	suite.NoError(err)
	suite.Len(tags, 1)

	// Search with end of tag string.
	tags, err = suite.db.SearchForTags(suite.T().Context(), "come", toPage("", "", "", 10))
	suite.NoError(err)
	suite.Len(tags, 1)
}

func TestSearchTestSuite(t *testing.T) {
	suite.Run(t, new(SearchTestSuite))
}
