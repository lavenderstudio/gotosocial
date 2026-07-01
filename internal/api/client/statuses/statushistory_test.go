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

type StatusHistoryTestSuite struct {
	StatusStandardTestSuite
}

func (suite *StatusHistoryTestSuite) TestGetHistory() {
	var (
		testApplication = suite.testApplications["application_1"]
		testAccount     = suite.testAccounts["local_account_1"]
		testUser        = suite.testUsers["local_account_1"]
		testToken       = oauth.DBTokenToToken(suite.testTokens["local_account_1"])
		targetStatusID  = suite.testStatuses["local_account_1_status_1"].ID
		target          = fmt.Sprintf("http://localhost:8080%s", strings.ReplaceAll(statuses.HistoryPath, ":id", targetStatusID))
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
	suite.statusModule.StatusHistoryGETHandler(c)

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
	suite.Equal(`[
  {
    "account": {
      "acct": "the_mighty_zork",
      "avatar": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/original/01F8MH58A357CV5K7R7TJMSH6S.jpg",
      "avatar_description": "a green goblin looking nasty",
      "avatar_media_id": "01F8MH58A357CV5K7R7TJMSH6S",
      "avatar_static": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/small/01F8MH58A357CV5K7R7TJMSH6S.webp",
      "bot": false,
      "created_at": "2022-05-20T11:09:18.000Z",
      "discoverable": true,
      "display_name": "original zork (he/they)",
      "emojis": [],
      "enable_rss": true,
      "fields": [],
      "followers_count": 2,
      "following_count": 2,
      "group": false,
      "header": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg",
      "header_description": "A very old-school screenshot of the original team fortress mod for quake",
      "header_media_id": "01PFPMWK2FF0D9WMHEJHR07C3Q",
      "header_static": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/small/01PFPMWK2FF0D9WMHEJHR07C3Q.webp",
      "id": "01F8MH1H7YV1Z7D2C8K2730QBF",
      "indexable": true,
      "last_status_at": "2024-11-01",
      "locked": false,
      "noindex": false,
      "note": "<p>hey yo this is my profile!</p>",
      "statuses_count": 9,
      "url": "http://localhost:8080/@the_mighty_zork",
      "username": "the_mighty_zork"
    },
    "content": "<p>hello everyone!</p>",
    "created_at": "2021-10-20T10:40:37.000Z",
    "emojis": [],
    "media_attachments": [],
    "poll": null,
    "sensitive": true,
    "spoiler_text": "introduction post"
  }
]`, out)
}

func TestStatusHistoryTestSuite(t *testing.T) {
	suite.Run(t, new(StatusHistoryTestSuite))
}
