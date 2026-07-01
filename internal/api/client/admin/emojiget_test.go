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
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/admin"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type EmojiGetTestSuite struct {
	AdminStandardTestSuite
}

func (suite *EmojiGetTestSuite) TestEmojiGet1() {
	recorder := httptest.NewRecorder()
	testEmoji := suite.testEmojis["rainbow"]

	path := admin.EmojiPathWithID
	c := suite.newContext(recorder, http.MethodGet, nil, path, "application/json")
	c.SetPathValue(apiutil.IDKey, testEmoji.ID)

	suite.adminModule.EmojiGETHandler(c)
	suite.Equal(http.StatusOK, recorder.Code)

	b, err := io.ReadAll(recorder.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	out := testrig.MustJSONStringFromBytes(b)
	suite.Equal(`{
  "category": "reactions",
  "content_type": "image/png",
  "disabled": false,
  "id": "01F8MH9H8E4VG3KDYJR9EGPXCQ",
  "shortcode": "rainbow",
  "static_url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/static/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
  "total_file_size": 42794,
  "updated_at": "2021-09-20T10:40:37.000Z",
  "uri": "http://localhost:8080/emoji/01F8MH9H8E4VG3KDYJR9EGPXCQ",
  "url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
  "visible_in_picker": true
}`, out)
}

func (suite *EmojiGetTestSuite) TestEmojiGet2() {
	recorder := httptest.NewRecorder()
	testEmoji := suite.testEmojis["yell"]

	path := admin.EmojiPathWithID
	c := suite.newContext(recorder, http.MethodGet, nil, path, "application/json")
	c.SetPathValue(apiutil.IDKey, testEmoji.ID)

	suite.adminModule.EmojiGETHandler(c)
	suite.Equal(http.StatusOK, recorder.Code)

	b, err := io.ReadAll(recorder.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	out := testrig.MustJSONStringFromBytes(b)
	suite.Equal(`{
  "content_type": "image/png",
  "disabled": false,
  "domain": "fossbros-anonymous.io",
  "id": "01GD5KP5CQEE1R3X43Y1EHS2CW",
  "shortcode": "yell",
  "static_url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/static/01GD5KP5CQEE1R3X43Y1EHS2CW.png",
  "total_file_size": 19854,
  "updated_at": "2020-03-18T12:12:00.000Z",
  "uri": "http://fossbros-anonymous.io/emoji/01GD5KP5CQEE1R3X43Y1EHS2CW",
  "url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01GD5KP5CQEE1R3X43Y1EHS2CW.png",
  "visible_in_picker": false
}`, out)
}

func (suite *EmojiGetTestSuite) TestEmojiGetNotFound() {
	recorder := httptest.NewRecorder()

	path := admin.EmojiPathWithID
	c := suite.newContext(recorder, http.MethodGet, nil, path, "application/json")
	c.SetPathValue(apiutil.IDKey, "01GF8VRXX1R00X7XH8973Z29R1")

	suite.adminModule.EmojiGETHandler(c)
	suite.Equal(http.StatusNotFound, recorder.Code)

	b, err := io.ReadAll(recorder.Body)
	suite.NoError(err)
	suite.NotNil(b)
	suite.Equal(`{"error":"Not Found"}`, string(b))
}

func TestEmojiGetTestSuite(t *testing.T) {
	suite.Run(t, &EmojiGetTestSuite{})
}
