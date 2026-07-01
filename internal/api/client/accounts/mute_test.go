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
	"net/url"
	"strconv"
	"strings"
	"testing"

	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/accounts"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"github.com/stretchr/testify/suite"
)

type MuteTestSuite struct {
	AccountStandardTestSuite
}

func (suite *MuteTestSuite) postMute(
	accountID string,
	notifications *bool,
	duration *int,
	requestJson *string,
	expectedHTTPStatus int,
	expectedBody string,
) (*apimodel.Relationship, error) {
	// instantiate recorder + test context
	req := httptest.NewRequest(http.MethodPut, config.GetProtocol()+"://"+config.GetHost()+"/api/"+accounts.BasePath+"/"+accountID+"/mute", nil)
	recorder := httptest.NewRecorder()
	c := httputil.ToContext(recorder, req)
	c.V.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	c.V.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["local_account_1"]))
	c.V.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	c.V.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])

	c.R.Header.Set("accept", "application/json")
	if requestJson != nil {
		c.R.Header.Set("content-type", "application/json")
		c.R.Body = io.NopCloser(strings.NewReader(*requestJson))
	} else {
		c.R.Form = make(url.Values)
		if notifications != nil {
			c.R.Form["notifications"] = []string{strconv.FormatBool(*notifications)}
		}
		if duration != nil {
			c.R.Form["duration"] = []string{strconv.Itoa(*duration)}
		}
	}

	c.SetPathValue("id", accountID)

	// trigger the handler
	suite.accountsModule.AccountMutePOSTHandler(c)

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

func (suite *MuteTestSuite) TestPostMuteFull() {
	accountID := suite.testAccounts["remote_account_1"].ID
	notifications := true
	duration := 86400
	relationship, err := suite.postMute(accountID, &notifications, &duration, nil, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.True(relationship.Muting)
	suite.Equal(notifications, relationship.MutingNotifications)
}

func (suite *MuteTestSuite) TestPostMuteFullJSON() {
	accountID := suite.testAccounts["remote_account_2"].ID
	// Use a numeric literal with a fractional part to test the JSON-specific handling for non-integer "duration".
	requestJson := `{
		"notifications": true,
		"duration": 86400.1
	}`
	relationship, err := suite.postMute(accountID, nil, nil, &requestJson, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.True(relationship.Muting)
	suite.True(relationship.MutingNotifications)
}

func (suite *MuteTestSuite) TestPostMuteMinimal() {
	accountID := suite.testAccounts["remote_account_3"].ID
	relationship, err := suite.postMute(accountID, nil, nil, nil, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.True(relationship.Muting)
	suite.False(relationship.MutingNotifications)
}

func (suite *MuteTestSuite) TestPostMuteSelf() {
	accountID := suite.testAccounts["local_account_1"].ID
	_, err := suite.postMute(accountID, nil, nil, nil, http.StatusNotAcceptable, `{"error":"Not Acceptable: getMuteTarget: account 01F8MH1H7YV1Z7D2C8K2730QBF cannot mute or unmute itself"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *MuteTestSuite) TestPostMuteNonexistentAccount() {
	accountID := "not_even_a_real_ULID"
	// Even though we pass account ID as mixed-case,
	// accountIDs always get converted to uppercase
	// (since ULIDs are always uppercase) in apiutil.ParseID().
	_, err := suite.postMute(accountID, nil, nil, nil, http.StatusNotFound, `{"error":"Not Found: getMuteTarget: target account NOT_EVEN_A_REAL_ULID not found in the db"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func TestMuteTestSuite(t *testing.T) {
	suite.Run(t, new(MuteTestSuite))
}
