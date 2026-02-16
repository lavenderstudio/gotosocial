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
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/admin"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type AccountActionTestSuite struct {
	AdminStandardTestSuite
}

func (suite *AccountActionTestSuite) TestSuspendAccount() {
	testAccount := new(gtsmodel.Account)
	*testAccount = *suite.testAccounts["local_account_1"]

	// Set up the request.
	requestBody, w, err := testrig.CreateMultipartFormData(
		nil,
		map[string][]string{
			"type": {"suspend"},
			"text": {"yeet this nerd"},
		})
	if err != nil {
		panic(err)
	}
	bodyBytes := requestBody.Bytes()
	recorder := httptest.NewRecorder()
	ctx := suite.newContext(
		recorder,
		http.MethodPost,
		bodyBytes,
		admin.AccountsActionPath,
		w.FormDataContentType(),
	)
	ctx.AddParam(apiutil.IDKey, testAccount.ID)

	// Call the handler
	suite.adminModule.AccountActionPOSTHandler(ctx)

	// We should have no error
	// message in the result body.
	result := recorder.Result()
	defer result.Body.Close()

	// Check the response
	b, err := io.ReadAll(result.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// We should have OK because
	// our request was valid.
	if recorder.Code != http.StatusOK {
		suite.FailNow("",
			"expected code %s, got %d: %v",
			http.StatusOK, recorder.Code, string(b),
		)
	}

	// Account should be suspended.
	if !testrig.WaitFor(func() bool {
		account, err := suite.state.DB.GetAccountByID(ctx, testAccount.ID)
		return err == nil && !account.SuspendedAt.IsZero()
	}) {
		suite.FailNow("", "failed waiting for account to be suspended")
	}

	// Appropriate admin action
	// should be in the db.
	if !testrig.WaitFor(func() bool {
		adminActions, err := suite.state.DB.GetAdminActions(ctx)
		return err == nil && slices.ContainsFunc(
			adminActions,
			func(a *gtsmodel.AdminAction) bool {
				return a.TargetID == testAccount.ID &&
					a.Text == "yeet this nerd"
			},
		)
	}) {
		suite.FailNow("", "admin actions did not contain expected action")
	}
}

func TestAccountActionTestSuite(t *testing.T) {
	suite.Run(t, &AccountActionTestSuite{})
}
