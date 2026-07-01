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

package statuses_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/statuses"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type StatusUnfaveTestSuite struct {
	StatusStandardTestSuite
}

// unfave a status
func (suite *StatusUnfaveTestSuite) TestPostUnfave() {
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)

	// this is the status we wanna unfave: in the testrig it's already faved by this account
	targetStatus := suite.testStatuses["admin_account_status_1"]

	// setup
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080%s", strings.Replace(statuses.UnfavouritePath, ":id", targetStatus.ID, 1)), nil) // the endpoint we're hitting
	req.Header.Set("accept", "application/json")
	recorder := httptest.NewRecorder()
	c := httputil.ToContext(recorder, req)
	c.V.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	c.V.Set(oauth.SessionAuthorizedToken, oauthToken)
	c.V.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	c.V.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	c.SetPathValue(apiutil.IDKey, targetStatus.ID)

	suite.statusModule.StatusUnfavePOSTHandler(c)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := io.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	statusReply := &apimodel.Status{}
	err = json.Unmarshal(b, statusReply)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), targetStatus.ContentWarning, statusReply.SpoilerText)
	assert.Equal(suite.T(), targetStatus.Content, statusReply.Content)
	assert.False(suite.T(), statusReply.Sensitive)
	assert.Equal(suite.T(), apimodel.VisibilityPublic, statusReply.Visibility)
	assert.False(suite.T(), statusReply.Favourited)
	assert.Equal(suite.T(), 0, statusReply.FavouritesCount)
}

// try to unfave a status that's already not faved
func (suite *StatusUnfaveTestSuite) TestPostAlreadyNotFaved() {
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)

	// this is the status we wanna unfave: in the testrig it's not faved by this account
	targetStatus := suite.testStatuses["admin_account_status_2"]

	// setup
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080%s", strings.Replace(statuses.UnfavouritePath, ":id", targetStatus.ID, 1)), nil) // the endpoint we're hitting
	req.Header.Set("accept", "application/json")
	recorder := httptest.NewRecorder()
	c := httputil.ToContext(recorder, req)
	c.V.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	c.V.Set(oauth.SessionAuthorizedToken, oauthToken)
	c.V.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	c.V.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	c.SetPathValue(apiutil.IDKey, targetStatus.ID)

	suite.statusModule.StatusUnfavePOSTHandler(c)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := io.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	statusReply := &apimodel.Status{}
	err = json.Unmarshal(b, statusReply)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), targetStatus.Content, statusReply.Content)
	assert.True(suite.T(), statusReply.Sensitive)
	assert.Equal(suite.T(), apimodel.VisibilityPublic, statusReply.Visibility)
	assert.False(suite.T(), statusReply.Favourited)
	assert.Equal(suite.T(), 0, statusReply.FavouritesCount)
}

func TestStatusUnfaveTestSuite(t *testing.T) {
	suite.Run(t, new(StatusUnfaveTestSuite))
}
