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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/accounts"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type BlockTestSuite struct {
	AccountStandardTestSuite
}

func (suite *BlockTestSuite) TestBlockSelf() {
	testAcct := suite.testAccounts["local_account_1"]

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080%s", strings.Replace(accounts.BlockPath, ":id", testAcct.ID, 1)), nil)
	recorder := httptest.NewRecorder()
	c := httputil.ToContext(recorder, req)
	c.V.Set(oauth.SessionAuthorizedAccount, testAcct)
	c.V.Set(oauth.SessionAuthorizedToken, oauth.DBTokenToToken(suite.testTokens["local_account_1"]))
	c.V.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	c.V.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])

	c.SetPathValue(apiutil.IDKey, testAcct.ID)

	suite.accountsModule.AccountBlockPOSTHandler(c)

	// 1. status should be Not Acceptable due to attempted self-block
	suite.Equal(http.StatusNotAcceptable, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := io.ReadAll(result.Body)
	_ = b
	assert.NoError(suite.T(), err)
}

func TestBlockTestSuite(t *testing.T) {
	suite.Run(t, new(BlockTestSuite))
}
