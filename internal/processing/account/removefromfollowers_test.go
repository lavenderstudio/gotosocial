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
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"github.com/stretchr/testify/suite"
)

type RemoveFromFollowersTestSuite struct {
	AccountStandardTestSuite
}

func (suite *RemoveFromFollowersTestSuite) TestRemoveFromFollowers() {
	var (
		ctx       = suite.T().Context()
		requester = suite.testAccounts["local_account_1"]
		follower  = suite.testAccounts["remote_account_1"]
	)

	// Have remote account 1 follow zork.
	follow := &gtsmodel.Follow{
		ID:              "01KXK81BF1K7V18KDGQJFZD8H5",
		URI:             "https://fossbros-anonymous.io/activities/019f6682-42fe-7cad-870b-d561133d615e",
		AccountID:       follower.ID,
		TargetAccountID: requester.ID,
	}
	if err := suite.state.DB.PutFollow(ctx, follow); err != nil {
		suite.FailNow(err.Error())
	}

	// Have zork remove the follower.
	relationship, errWithCode := suite.accountProcessor.RemoveFromFollowers(ctx, requester, follower.ID)
	if errWithCode != nil {
		suite.FailNow(errWithCode.Error())
	}
	suite.False(relationship.FollowedBy)

	// There should be a message going to the worker.
	cMsg, _ := suite.getClientMsg(5 * time.Second)
	suite.Equal(ap.ActivityReject, cMsg.APActivityType)
	suite.Equal(ap.ActivityFollow, cMsg.APObjectType)
	suite.Equal(requester.ID, cMsg.Origin.ID)
	suite.Equal(follower.ID, cMsg.Target.ID)
}

func TestRemoveFromFollowersTestSuite(t *testing.T) {
	suite.Run(t, new(RemoveFromFollowersTestSuite))
}
