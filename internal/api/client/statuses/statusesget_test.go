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
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/statuses"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type StatusesGetTestSuite struct {
	StatusStandardTestSuite
}

func (suite *StatusesGetTestSuite) getStatuses(
	requestingAccount *gtsmodel.Account,
	token *gtsmodel.Token,
	user *gtsmodel.User,
	statusIDs []string,
	expectedHTTPStatus int,
	expectedBody string,
) ([]apimodel.Status, error) {
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)

	requestURL := testrig.URLMustParse("/api" + statuses.BasePath)
	query := url.Values{}
	for _, statusID := range statusIDs {
		query.Add("id[]", statusID)
	}

	requestURL.RawQuery = query.Encode()
	ctx.Request = httptest.NewRequest(http.MethodGet, requestURL.String(), nil)
	ctx.Request.Header.Set("accept", "application/json")

	if token != nil {
		ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(token))
		ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	}
	if requestingAccount != nil {
		ctx.Set(oauth.SessionAuthorizedAccount, requestingAccount)
	}
	if user != nil {
		ctx.Set(oauth.SessionAuthorizedUser, user)
	}

	suite.statusModule.StatusesGETHandler(ctx)

	result := recorder.Result()
	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(expectedHTTPStatus, recorder.Code, string(body))

	if expectedBody != "" {
		suite.Equal(expectedBody, string(body))
		return nil, nil
	}

	apiStatuses := make([]apimodel.Status, 0)
	if err := json.Unmarshal(body, &apiStatuses); err != nil {
		suite.FailNow(err.Error(), string(body))
	}

	return apiStatuses, nil
}

func (suite *StatusesGetTestSuite) TestGetStatusesAuthenticatedOK() {
	apiStatuses, err := suite.getStatuses(
		suite.testAccounts["local_account_1"],
		suite.testTokens["local_account_1"],
		suite.testUsers["local_account_1"],
		[]string{
			suite.testStatuses["admin_account_status_1"].ID,
			suite.testStatuses["admin_account_status_2"].ID,
			"01ZZZZZZZZZZZZZZZZZZZZZZZZ",
		},
		http.StatusOK,
		"",
	)
	suite.NoError(err)
	suite.Len(apiStatuses, 2)
	suite.Equal(suite.testStatuses["admin_account_status_1"].ID, apiStatuses[0].ID)
	suite.Equal(suite.testStatuses["admin_account_status_2"].ID, apiStatuses[1].ID)
}

func (suite *StatusesGetTestSuite) TestGetStatusesNoIDsBadRequest() {
	_, err := suite.getStatuses(
		suite.testAccounts["local_account_1"],
		suite.testTokens["local_account_1"],
		suite.testUsers["local_account_1"],
		nil,
		http.StatusBadRequest,
		`{"error":"Bad Request: no status ids specified"}`,
	)
	suite.NoError(err)
}

func (suite *StatusesGetTestSuite) TestGetStatusesTooManyIDs() {
	_, err := suite.getStatuses(
		suite.testAccounts["local_account_1"],
		suite.testTokens["local_account_1"],
		suite.testUsers["local_account_1"],
		make([]string, 101),
		http.StatusUnprocessableEntity,
		`{"error":"Unprocessable Entity: too many status ids specified"}`,
	)
	suite.NoError(err)
}

func (suite *StatusesGetTestSuite) TestGetStatusesInsufficientScope() {
	_, err := suite.getStatuses(
		suite.testAccounts["local_account_1"],
		suite.testTokens["local_account_1_push_only"],
		suite.testUsers["local_account_1"],
		[]string{suite.testStatuses["admin_account_status_1"].ID},
		http.StatusForbidden,
		`{"error":"Forbidden: token has insufficient scope permission"}`,
	)
	suite.NoError(err)
}

func (suite *StatusesGetTestSuite) TestGetStatusesDeduplicatesAndKeepsOrder() {
	apiStatuses, err := suite.getStatuses(
		suite.testAccounts["local_account_1"],
		suite.testTokens["local_account_1"],
		suite.testUsers["local_account_1"],
		[]string{
			suite.testStatuses["admin_account_status_2"].ID,
			suite.testStatuses["admin_account_status_1"].ID,
			suite.testStatuses["admin_account_status_2"].ID,
		},
		http.StatusOK,
		"",
	)
	suite.NoError(err)
	suite.Len(apiStatuses, 2)
	suite.Equal(suite.testStatuses["admin_account_status_2"].ID, apiStatuses[0].ID)
	suite.Equal(suite.testStatuses["admin_account_status_1"].ID, apiStatuses[1].ID)
}

func TestStatusesGetTestSuite(t *testing.T) {
	suite.Run(t, new(StatusesGetTestSuite))
}
