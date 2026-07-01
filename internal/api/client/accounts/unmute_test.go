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

package accounts_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/accounts"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
)

func (suite *MuteTestSuite) postUnmute(
	accountID string,
	expectedHTTPStatus int,
	expectedBody string,
) (*apimodel.Relationship, error) {
	// create the request
	req := httptest.NewRequest(http.MethodPut, config.GetProtocol()+"://"+config.GetHost()+"/api/"+accounts.BasePath+"/"+accountID+"/unmute", nil)
	req.Header.Set("accept", "application/json")

	// instantiate recorder + test context
	recorder := httptest.NewRecorder()
	c := httputil.ToContext(recorder, req)
	c.V.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	c.V.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["local_account_1"]))
	c.V.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	c.V.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])

	c.SetPathValue(apiutil.IDKey, accountID)

	// trigger the handler
	suite.accountsModule.AccountUnmutePOSTHandler(c)

	// read the response
	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	errs := gtserror.NewMultiError(2)

	// check code + body
	if resultCode := recorder.Code; expectedHTTPStatus != resultCode {
		errs.Appendf("expected %d got %d", expectedHTTPStatus, resultCode)
		if expectedBody == "" {
			return nil, errs.Combine()
		}
	}

	// if we got an expected body, return early
	if expectedBody != "" {
		if string(b) != expectedBody {
			errs.Appendf("expected %s got %s", expectedBody, string(b))
		}
		return nil, errs.Combine()
	}

	resp := &apimodel.Relationship{}
	if err := json.Unmarshal(b, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (suite *MuteTestSuite) TestPostUnmuteWithoutPreviousMute() {
	accountID := suite.testAccounts["remote_account_4"].ID
	relationship, err := suite.postUnmute(accountID, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.False(relationship.Muting)
	suite.False(relationship.MutingNotifications)
}

func (suite *MuteTestSuite) TestPostWithPreviousMute() {
	accountID := suite.testAccounts["local_account_2"].ID

	relationship, err := suite.postMute(accountID, nil, nil, nil, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.True(relationship.Muting)
	suite.False(relationship.MutingNotifications)

	relationship, err = suite.postUnmute(accountID, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.False(relationship.Muting)
	suite.False(relationship.MutingNotifications)
}

func (suite *MuteTestSuite) TestPostUnmuteSelf() {
	accountID := suite.testAccounts["local_account_1"].ID
	_, err := suite.postUnmute(accountID, http.StatusNotAcceptable, `{"error":"Not Acceptable: getMuteTarget: account 01F8MH1H7YV1Z7D2C8K2730QBF cannot mute or unmute itself"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *MuteTestSuite) TestPostUnmuteNonexistentAccount() {
	accountID := "not_even_a_real_ULID"
	// Even though we pass account ID as mixed-case,
	// accountIDs always get converted to uppercase
	// (since ULIDs are always uppercase) in apiutil.ParseID().
	_, err := suite.postUnmute(accountID, http.StatusNotFound, `{"error":"Not Found: getMuteTarget: target account NOT_EVEN_A_REAL_ULID not found in the db"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}
