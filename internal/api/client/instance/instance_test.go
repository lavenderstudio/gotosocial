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

package instance_test

import (
	"bytes"
	"fmt"
	"net/http/httptest"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type InstanceStandardTestSuite struct {
	suite.Suite

	testTokens       map[string]*gtsmodel.Token
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account
	testAttachments  map[string]*gtsmodel.MediaAttachment
}

const (
	rMediaPath    = "../../../../testrig/media"
	rTemplatePath = "../../../../web/template/"
)

func (suite *InstanceStandardTestSuite) SetupSuite() {
	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.testTokens = testrig.NewTestTokens()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
}

func (suite *InstanceStandardTestSuite) newContext(
	recorder *httptest.ResponseRecorder,
	method string,
	path string,
	body []byte,
	contentType string,
	auth bool,
) *gin.Context {
	protocol := config.GetProtocol()
	host := config.GetHost()

	baseURI := fmt.Sprintf("%s://%s", protocol, host)
	requestURI := fmt.Sprintf("%s/%s", baseURI, path)

	req := httptest.NewRequest(method, requestURI, bytes.NewReader(body)) // the endpoint we're hitting

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	req.Header.Set("accept", "application/json")

	ctx, _ := testrig.CreateGinTestContext(recorder, req)

	if auth {
		ctx.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["admin_account"])
		ctx.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["admin_account"]))
		ctx.Set(oauth.SessionAuthorizedApplication, suite.testApplications["admin_account"])
		ctx.Set(oauth.SessionAuthorizedUser, suite.testUsers["admin_account"])
	}

	return ctx
}
