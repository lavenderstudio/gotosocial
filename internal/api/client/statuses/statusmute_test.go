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

type StatusMuteTestSuite struct {
	StatusStandardTestSuite
}

func (suite *StatusMuteTestSuite) post(path string, handler func(*httputil.Context), targetStatusID string) (int, string) {
	t := suite.testTokens["local_account_1"]
	oauthToken := oauth.DBTokenToToken(t)

	req := httptest.NewRequest(http.MethodPost, path, nil) // the endpoint we're hitting
	req.Header.Set("accept", "application/json")
	recorder := httptest.NewRecorder()
	c := httputil.ToContext(recorder, req)
	c.V.Set(oauth.SessionAuthorizedApplication, suite.testApplications["application_1"])
	c.V.Set(oauth.SessionAuthorizedToken, oauthToken)
	c.V.Set(oauth.SessionAuthorizedUser, suite.testUsers["local_account_1"])
	c.V.Set(oauth.SessionAuthorizedAccount, suite.testAccounts["local_account_1"])

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	c.SetPathValue(apiutil.IDKey, targetStatusID)

	handler(c)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	return recorder.Code, testrig.MustJSONStringFromBytes(b)
}

func (suite *StatusMuteTestSuite) TestMuteUnmuteStatus() {
	var (
		targetStatus = suite.testStatuses["local_account_1_status_1"]
		path         = fmt.Sprintf("http://localhost:8080/api%s", strings.ReplaceAll(statuses.MutePath, ":id", targetStatus.ID))
	)

	// Mute the status, ensure `muted` is `true`.
	code, muted := suite.post(path, suite.statusModule.StatusMutePOSTHandler, targetStatus.ID)
	suite.Equal(http.StatusOK, code)
	suite.Equal(`{
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
  "application": {
    "name": "really cool gts application",
    "website": "https://reallycool.app"
  },
  "bookmarked": false,
  "card": null,
  "content": "<p>hello everyone!</p>",
  "content_type": "text/plain",
  "created_at": "2021-10-20T10:40:37.000Z",
  "edited_at": null,
  "emojis": [],
  "favourited": false,
  "favourites_count": 1,
  "id": "01F8MHAMCHF6Y650WCRSCP4WMY",
  "in_reply_to_account_id": null,
  "in_reply_to_id": null,
  "interaction_policy": {
    "can_favourite": {
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": []
    },
    "can_reblog": {
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": []
    },
    "can_reply": {
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": []
    }
  },
  "language": "en",
  "media_attachments": [],
  "mentions": [],
  "muted": true,
  "pinned": false,
  "poll": null,
  "reblog": null,
  "reblogged": false,
  "reblogs_count": 1,
  "replies_count": 2,
  "sensitive": true,
  "spoiler_text": "introduction post",
  "tags": [],
  "text": "hello everyone!",
  "uri": "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
  "url": "http://localhost:8080/@the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
  "visibility": "public"
}`, muted)

	// Unmute the status, ensure `muted` is `false`.
	code, unmuted := suite.post(path, suite.statusModule.StatusUnmutePOSTHandler, targetStatus.ID)
	suite.Equal(http.StatusOK, code)
	suite.Equal(`{
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
  "application": {
    "name": "really cool gts application",
    "website": "https://reallycool.app"
  },
  "bookmarked": false,
  "card": null,
  "content": "<p>hello everyone!</p>",
  "content_type": "text/plain",
  "created_at": "2021-10-20T10:40:37.000Z",
  "edited_at": null,
  "emojis": [],
  "favourited": false,
  "favourites_count": 1,
  "id": "01F8MHAMCHF6Y650WCRSCP4WMY",
  "in_reply_to_account_id": null,
  "in_reply_to_id": null,
  "interaction_policy": {
    "can_favourite": {
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": []
    },
    "can_reblog": {
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": []
    },
    "can_reply": {
      "automatic_approval": [
        "public",
        "me"
      ],
      "manual_approval": []
    }
  },
  "language": "en",
  "media_attachments": [],
  "mentions": [],
  "muted": false,
  "pinned": false,
  "poll": null,
  "reblog": null,
  "reblogged": false,
  "reblogs_count": 1,
  "replies_count": 2,
  "sensitive": true,
  "spoiler_text": "introduction post",
  "tags": [],
  "text": "hello everyone!",
  "uri": "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
  "url": "http://localhost:8080/@the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
  "visibility": "public"
}`, unmuted)
}

func TestStatusMuteTestSuite(t *testing.T) {
	suite.Run(t, new(StatusMuteTestSuite))
}
