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

package v2_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"

	"code.superseriousbusiness.org/gopkg/httputil"
	filtersV2 "code.superseriousbusiness.org/gotosocial/internal/api/client/filters/v2"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/internal/stream"
)

func (suite *FiltersTestSuite) putFilterKeyword(
	filterKeywordID string,
	keyword *string,
	wholeWord *bool,
	requestJson *string,
	expectedHTTPStatus int,
	expectedBody string,
) (*apimodel.FilterKeyword, error) {
	// create the request
	req := httptest.NewRequest(http.MethodPut, config.GetProtocol()+"://"+config.GetHost()+"/api/"+filtersV2.KeywordPath+"/"+filterKeywordID, nil)
	req.Header.Set("accept", "application/json")

	// instantiate recorder + test context
	recorder := httptest.NewRecorder()
	c := httputil.ToContext(recorder, req)
	c.V.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])
	c.V.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["local_account_1"]))
	c.V.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	c.V.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])

	if requestJson != nil {
		c.R.Header.Set("content-type", "application/json")
		c.R.Body = io.NopCloser(strings.NewReader(*requestJson))
	} else {
		c.R.Form = make(url.Values)
		if keyword != nil {
			c.R.Form["keyword"] = []string{*keyword}
		}
		if wholeWord != nil {
			c.R.Form["whole_word"] = []string{strconv.FormatBool(*wholeWord)}
		}
	}

	c.SetPathValue("id", filterKeywordID)

	// trigger the handler
	suite.filtersModule.FilterKeywordPUTHandler(c)

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

	resp := &apimodel.FilterKeyword{}
	if err := json.Unmarshal(b, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (suite *FiltersTestSuite) TestPutFilterKeywordFull() {
	homeStream := suite.openHomeStream(suite.testAccounts["local_account_1"])

	filterKeywordID := suite.testFilterKeywords["local_account_1_filter_1_keyword_1"].ID
	keyword := "fnords"
	wholeWord := true
	filterKeyword, err := suite.putFilterKeyword(filterKeywordID, &keyword, &wholeWord, nil, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(keyword, filterKeyword.Keyword)
	suite.Equal(wholeWord, filterKeyword.WholeWord)

	suite.checkStreamed(homeStream, true, "", stream.EventTypeFiltersChanged)
}

func (suite *FiltersTestSuite) TestPutFilterKeywordFullJSON() {
	homeStream := suite.openHomeStream(suite.testAccounts["local_account_1"])

	filterKeywordID := suite.testFilterKeywords["local_account_1_filter_1_keyword_1"].ID
	requestJson := `{
		"keyword": "fnords",
		"whole_word": true
	}`
	filterKeyword, err := suite.putFilterKeyword(filterKeywordID, nil, nil, &requestJson, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal("fnords", filterKeyword.Keyword)
	suite.True(filterKeyword.WholeWord)

	suite.checkStreamed(homeStream, true, "", stream.EventTypeFiltersChanged)
}

func (suite *FiltersTestSuite) TestPutFilterKeywordMinimal() {
	homeStream := suite.openHomeStream(suite.testAccounts["local_account_1"])

	filterKeywordID := suite.testFilterKeywords["local_account_1_filter_1_keyword_1"].ID
	keyword := "fnords"
	filterKeyword, err := suite.putFilterKeyword(filterKeywordID, &keyword, nil, nil, http.StatusOK, "")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(keyword, filterKeyword.Keyword)
	suite.False(filterKeyword.WholeWord)

	suite.checkStreamed(homeStream, true, "", stream.EventTypeFiltersChanged)
}

func (suite *FiltersTestSuite) TestPutFilterKeywordEmptyKeyword() {
	filterKeywordID := suite.testFilterKeywords["local_account_1_filter_1_keyword_1"].ID
	keyword := ""
	_, err := suite.putFilterKeyword(filterKeywordID, &keyword, nil, nil, http.StatusUnprocessableEntity, `{"error":"Unprocessable Entity: filter keyword must be provided, and must be no more than 40 chars"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *FiltersTestSuite) TestPutFilterKeywordMissingKeyword() {
	filterKeywordID := suite.testFilterKeywords["local_account_1_filter_1_keyword_1"].ID
	_, err := suite.putFilterKeyword(filterKeywordID, nil, nil, nil, http.StatusUnprocessableEntity, `{"error":"Unprocessable Entity: filter keyword must be provided, and must be no more than 40 chars"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

// Changing our filter keyword to the same keyword as another filter keyword in the same filter should fail.
func (suite *FiltersTestSuite) TestPutFilterKeywordKeywordConflict() {
	filterKeywordID := suite.testFilterKeywords["local_account_1_filter_2_keyword_1"].ID
	conflictingKeyword := suite.testFilterKeywords["local_account_1_filter_2_keyword_2"].Keyword
	_, err := suite.putFilterKeyword(filterKeywordID, &conflictingKeyword, nil, nil, http.StatusConflict, `{"error":"Conflict: duplicate keyword"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *FiltersTestSuite) TestPutFilterKeywordAnotherAccountsFilterKeyword() {
	filterKeywordID := suite.testFilterKeywords["local_account_2_filter_1_keyword_1"].ID
	keyword := "fnord"
	_, err := suite.putFilterKeyword(filterKeywordID, &keyword, nil, nil, http.StatusNotFound, `{"error":"Not Found: filter not found"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *FiltersTestSuite) TestPutFilterKeywordNonexistentFilterKeyword() {
	filterKeywordID := "not_even_a_real_ULID"
	keyword := "fnord"
	_, err := suite.putFilterKeyword(filterKeywordID, &keyword, nil, nil, http.StatusNotFound, `{"error":"Not Found: filter keyword not found"}`)
	if err != nil {
		suite.FailNow(err.Error())
	}
}
