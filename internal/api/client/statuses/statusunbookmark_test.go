/*
   GoToSocial
   Copyright (C) GoToSocial Authors admin@gotosocial.org
   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.
   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

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
	"code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"github.com/stretchr/testify/suite"
)

type StatusUnbookmarkTestSuite struct {
	StatusStandardTestSuite
}

func (suite *StatusUnbookmarkTestSuite) TestPostUnbookmark() {
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)

	targetStatus := suite.testStatuses["admin_account_status_1"]

	// setup
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080%s", strings.Replace(statuses.UnbookmarkPath, ":id", targetStatus.ID, 1)), nil) // the endpoint we're hitting
	req.Header.Set("accept", "application/json")
	recorder := httptest.NewRecorder()
	c := httputil.ToContext(recorder, req)
	c.V.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	c.V.Set(oauth.SessionAuthorizedToken, oauthToken)
	c.V.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	c.V.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])

	c.SetPathValue(apiutil.IDKey, targetStatus.ID)

	suite.statusModule.StatusUnbookmarkPOSTHandler(c)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	statusReply := &model.Status{}
	err = json.Unmarshal(b, statusReply)
	suite.NoError(err)

	suite.False(statusReply.Bookmarked)
}

func TestStatusUnbookmarkTestSuite(t *testing.T) {
	suite.Run(t, new(StatusUnbookmarkTestSuite))
}
