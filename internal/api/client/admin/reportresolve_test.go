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

package admin_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/admin"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/stretchr/testify/suite"
)

type ReportResolveTestSuite struct {
	AdminStandardTestSuite
}

func (suite *ReportResolveTestSuite) resolveReport(
	account *gtsmodel.Account,
	token *gtsmodel.Token,
	user *gtsmodel.User,
	targetReportID string,
	expectedHTTPStatus int,
	expectedBody string,
	actionTakenComment *string,
) (*apimodel.AdminReport, error) {
	// create the request URI
	requestPath := admin.ReportsPath + "/" + targetReportID + "/resolve"
	baseURI := config.GetProtocol() + "://" + config.GetHost()
	requestURI := baseURI + "/api/" + requestPath

	// create the request
	req := httptest.NewRequest(http.MethodPost, requestURI, nil)
	req.Header.Set("accept", "application/json")
	if actionTakenComment != nil {
		req.Form = url.Values{"action_taken_comment": {*actionTakenComment}}
	}

	// instantiate recorder + test context
	recorder := httptest.NewRecorder()
	c := httputil.ToContext(recorder, req)
	c.V.Set(oauth.SessionAuthorizedAccount, account)
	c.V.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(token))
	c.V.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	c.V.Set(oauth.SessionAuthorizedUser, user)

	c.SetPathValue(apiutil.IDKey, targetReportID)

	// trigger the handler
	suite.adminModule.ReportResolvePOSTHandler(c)

	// read the response
	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	errs := gtserror.NewMultiError(2)

	if resultCode := recorder.Code; expectedHTTPStatus != resultCode {
		errs.Appendf("expected %d got %d", expectedHTTPStatus, resultCode)
	}

	// if we got an expected body, return early
	if expectedBody != "" {
		if string(b) != expectedBody {
			errs.Appendf("expected %s got %s", expectedBody, string(b))
		}
		return nil, errs.Combine()
	}

	resp := &apimodel.AdminReport{}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (suite *ReportResolveTestSuite) TestReportResolve1() {
	testAccount := suite.testAccounts["admin_account"]
	testToken := suite.testTokens["admin_account"]
	testUser := suite.testUsers["admin_account"]
	testReportID := suite.testReports["local_account_2_report_remote_account_1"].ID
	var actionTakenComment *string = nil

	report, err := suite.resolveReport(testAccount, testToken, testUser, testReportID, http.StatusOK, "", actionTakenComment)
	suite.NoError(err)
	suite.NotEmpty(report)

	// report should be resolved
	suite.True(report.ActionTaken)
	actionTime, err := util.ParseISO8601(*report.ActionTakenAt)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.WithinDuration(time.Now(), actionTime, 1*time.Minute)
	updatedTime, err := util.ParseISO8601(report.UpdatedAt)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.WithinDuration(time.Now(), updatedTime, 1*time.Minute)
	suite.Equal(report.ActionTakenByAccount.ID, testAccount.ID)
	suite.EqualValues(report.ActionTakenComment, actionTakenComment)
	suite.EqualValues(report.AssignedAccount.ID, testAccount.ID)
}

func (suite *ReportResolveTestSuite) TestReportResolve2() {
	testAccount := suite.testAccounts["admin_account"]
	testToken := suite.testTokens["admin_account"]
	testUser := suite.testUsers["admin_account"]
	testReportID := suite.testReports["local_account_2_report_remote_account_1"].ID
	var actionTakenComment *string = util.Ptr("no action was taken, this is a frivolous report you boob")

	report, err := suite.resolveReport(testAccount, testToken, testUser, testReportID, http.StatusOK, "", actionTakenComment)
	suite.NoError(err)
	suite.NotEmpty(report)

	// report should be resolved
	suite.True(report.ActionTaken)
	actionTime, err := util.ParseISO8601(*report.ActionTakenAt)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.WithinDuration(time.Now(), actionTime, 1*time.Minute)
	updatedTime, err := util.ParseISO8601(report.UpdatedAt)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.WithinDuration(time.Now(), updatedTime, 1*time.Minute)
	suite.Equal(report.ActionTakenByAccount.ID, testAccount.ID)
	suite.EqualValues(report.ActionTakenComment, actionTakenComment)
	suite.EqualValues(report.AssignedAccount.ID, testAccount.ID)
}

func TestReportResolveTestSuite(t *testing.T) {
	suite.Run(t, &ReportResolveTestSuite{})
}
