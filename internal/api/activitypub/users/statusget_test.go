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

package users_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"code.superseriousbusiness.org/gopkg/httputil"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type StatusGetTestSuite struct {
	UserStandardTestSuite
}

func (suite *StatusGetTestSuite) TestGetStatus() {
	// the dereference we're gonna use
	derefRequests := testrig.NewTestDereferenceRequests(suite.testAccounts)
	signedRequest := derefRequests["foss_satan_dereference_local_account_1_status_1"]
	targetAccount := suite.testAccounts["local_account_1"]
	targetStatus := suite.testStatuses["local_account_1_status_1"]

	// setup request
	req := httptest.NewRequest(http.MethodGet, targetStatus.URI, nil) // the endpoint we're hitting
	recorder := httptest.NewRecorder()
	c := httputil.ToContext(recorder, req)
	c.R.Header.Set("accept", "application/activity+json")
	c.R.Header.Set("Signature", signedRequest.SignatureHeader)
	c.R.Header.Set("Date", signedRequest.DateHeader)

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	c.SetPathValue(apiutil.UsernameKey, targetAccount.Username)
	c.SetPathValue(apiutil.IDKey, targetStatus.ID)

	// trigger the function being tested, first passing through sigcheck.
	suite.signatureCheck.Compile(suite.userModule.StatusGETHandler)(c)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	// should be a Note
	m := make(map[string]any)
	err = json.Unmarshal(b, &m)
	suite.NoError(err)

	t, err := streams.ToType(suite.T().Context(), m)
	suite.NoError(err)

	note, ok := t.(vocab.ActivityStreamsNote)
	suite.True(ok)

	// convert note to status
	a, err := suite.tc.ASStatusToStatus(suite.T().Context(), note)
	suite.NoError(err)
	suite.EqualValues(targetStatus.Content, a.Content)
}

func (suite *StatusGetTestSuite) TestGetStatusLowercase() {
	// the dereference we're gonna use
	derefRequests := testrig.NewTestDereferenceRequests(suite.testAccounts)
	signedRequest := derefRequests["foss_satan_dereference_local_account_1_status_1_lowercase"]
	targetAccount := suite.testAccounts["local_account_1"]
	targetStatus := suite.testStatuses["local_account_1_status_1"]

	// setup request
	req := httptest.NewRequest(http.MethodGet, strings.ToLower(targetStatus.URI), nil) // the endpoint we're hitting
	recorder := httptest.NewRecorder()
	c := httputil.ToContext(recorder, req)
	c.R.Header.Set("accept", "application/activity+json")
	c.R.Header.Set("Signature", signedRequest.SignatureHeader)
	c.R.Header.Set("Date", signedRequest.DateHeader)

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	c.SetPathValue(apiutil.UsernameKey, strings.ToLower(targetAccount.Username))
	c.SetPathValue(apiutil.IDKey, strings.ToLower(targetStatus.ID))

	// trigger the function being tested, first passing through sigcheck.
	suite.signatureCheck.Compile(suite.userModule.StatusGETHandler)(c)

	// check response
	suite.EqualValues(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	// should be a Note
	m := make(map[string]any)
	err = json.Unmarshal(b, &m)
	suite.NoError(err)

	t, err := streams.ToType(suite.T().Context(), m)
	suite.NoError(err)

	note, ok := t.(vocab.ActivityStreamsNote)
	suite.True(ok)

	// convert note to status
	a, err := suite.tc.ASStatusToStatus(suite.T().Context(), note)
	suite.NoError(err)
	suite.EqualValues(targetStatus.Content, a.Content)
}

func TestStatusGetTestSuite(t *testing.T) {
	suite.Run(t, new(StatusGetTestSuite))
}
