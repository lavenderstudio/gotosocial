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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/statuses"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type StatusSourceTestSuite struct {
	StatusStandardTestSuite
}

func (suite *StatusSourceTestSuite) TestGetSource() {
	var (
		testApplication = suite.testApplications["application_1"]
		testAccount     = suite.testAccounts["local_account_1"]
		testUser        = suite.testUsers["local_account_1"]
		testToken       = oauth.DBTokenToToken(suite.testTokens["local_account_1"])
		targetStatusID  = suite.testStatuses["local_account_1_status_1"].ID
		target          = fmt.Sprintf("http://localhost:8080%s", strings.ReplaceAll(statuses.SourcePath, ":id", targetStatusID))
	)

	// Setup request.
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, target, nil)
	request.Header.Set("accept", "application/json")
	c := httputil.ToContext(recorder, request)

	// Set auth + path params.
	c.V.Set(oauth.SessionAuthorizedApplication, testApplication)
	c.V.Set(oauth.SessionAuthorizedToken, testToken)
	c.V.Set(oauth.SessionAuthorizedUser, testUser)
	c.V.Set(oauth.SessionAuthorizedAccount, testAccount)
	c.SetPathValue(apiutil.IDKey, targetStatusID)

	// Call the handler.
	suite.statusModule.StatusSourceGETHandler(c)

	// Check code.
	if code := recorder.Code; code != http.StatusOK {
		suite.FailNow("", "unexpected http code: %d", code)
	}

	// Read body.
	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	out := testrig.MustJSONStringFromBytes(b)
	suite.Equal(`{
  "content_type": "text/plain",
  "id": "01F8MHAMCHF6Y650WCRSCP4WMY",
  "spoiler_text": "introduction post",
  "text": "hello everyone!"
}`, out)
}

func TestStatusSourceTestSuite(t *testing.T) {
	suite.Run(t, new(StatusSourceTestSuite))
}
