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

package account_test

import (
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"github.com/stretchr/testify/suite"
)

type RelationshipsTestSuite struct {
	AccountStandardTestSuite
}

func (suite *RelationshipsTestSuite) TestConnectedDomains() {
	ctx := suite.T().Context()

	// Have zork follow remote
	// account 1 for this test.
	if err := suite.state.DB.PutFollow(ctx,
		&gtsmodel.Follow{
			ID:              id.NewULID(),
			URI:             "whatever1",
			AccountID:       suite.testAccounts["local_account_1"].ID,
			TargetAccountID: suite.testAccounts["remote_account_1"].ID,
		},
	); err != nil {
		suite.FailNow(err.Error())
	}

	// Have remote account 2
	// follow zork for this test.
	if err := suite.state.DB.PutFollow(ctx,
		&gtsmodel.Follow{
			ID:              id.NewULID(),
			URI:             "whatever2",
			AccountID:       suite.testAccounts["remote_account_2"].ID,
			TargetAccountID: suite.testAccounts["local_account_1"].ID,
		},
	); err != nil {
		suite.FailNow(err.Error())
	}

	for _, test := range []struct {
		requester       *gtsmodel.Account
		acct            *gtsmodel.Account
		expectedDomains []string
	}{
		{
			requester:       nil,
			acct:            suite.testAccounts["admin_account"],
			expectedDomains: []string{"localhost:8080"},
		},
		{
			requester: nil,
			acct:      suite.testAccounts["local_account_1"],
			expectedDomains: []string{
				"example.org",
				"fossbros-anonymous.io",
				"localhost:8080",
			},
		},
		{
			requester: nil,
			acct:      suite.testAccounts["local_account_2"],
			// Hides collections.
			expectedDomains: nil,
		},
		{
			requester: suite.testAccounts["local_account_2"],
			acct:      suite.testAccounts["local_account_2"],
			// Show collections to self.
			expectedDomains: []string{"localhost:8080"},
		},
	} {
		domains, err := suite.accountProcessor.ConnectedDomainsGet(ctx,
			test.requester,
			test.acct.ID,
		)
		if err != nil {
			suite.FailNow(err.Error())
		}
		suite.EqualValues(test.expectedDomains, domains)
	}
}

func TestRelationshipsTestSuite(t *testing.T) {
	suite.Run(t, new(RelationshipsTestSuite))
}
