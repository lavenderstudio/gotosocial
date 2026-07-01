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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/instance"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type InstancePeersGetTestSuite struct {
	InstanceStandardTestSuite
}

func (suite *InstancePeersGetTestSuite) TestInstancePeersGetNoParams() {
	testStructs := testrig.SetupTestStructs(rMediaPath, rTemplatePath)
	defer testrig.TearDownTestStructs(testStructs)
	instanceModule := instance.New(testStructs.Processor, testStructs.Templates)

	baseURI := fmt.Sprintf("%s://%s", config.GetProtocol(), config.GetHost())
	requestURI := fmt.Sprintf("%s/%s", baseURI, instance.InstancePeersPath)
	req := httptest.NewRequest(http.MethodGet, requestURI, nil)

	recorder := httptest.NewRecorder()
	c := httputil.ToContext(recorder, req)

	instanceModule.InstancePeersGETHandler(c)

	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	out := testrig.MustJSONStringFromBytes(b)
	suite.Equal(`[
  "example.org",
  "fossbros-anonymous.io",
  "thequeenisstillalive.technology",
  "ëxample.org"
]`, out)
}

func (suite *InstancePeersGetTestSuite) TestInstancePeersGetNoParamsUnauthorized() {
	testStructs := testrig.SetupTestStructs(rMediaPath, rTemplatePath)
	defer testrig.TearDownTestStructs(testStructs)
	instanceModule := instance.New(testStructs.Processor, testStructs.Templates)

	config.SetInstanceExposePeers(false)

	recorder := httptest.NewRecorder()
	baseURI := fmt.Sprintf("%s://%s", config.GetProtocol(), config.GetHost())
	requestURI := fmt.Sprintf("%s/%s", baseURI, instance.InstancePeersPath)
	ctx := suite.newContext(recorder, http.MethodGet, requestURI, nil, "", false)

	instanceModule.InstancePeersGETHandler(ctx)

	suite.Equal(http.StatusUnauthorized, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"Unauthorized: peers open query requires an authenticated account/user"}`, string(b))
}

func (suite *InstancePeersGetTestSuite) TestInstancePeersGetNoParamsAuthorized() {
	testStructs := testrig.SetupTestStructs(rMediaPath, rTemplatePath)
	defer testrig.TearDownTestStructs(testStructs)
	instanceModule := instance.New(testStructs.Processor, testStructs.Templates)

	config.SetInstanceExposePeers(false)

	recorder := httptest.NewRecorder()
	baseURI := fmt.Sprintf("%s://%s", config.GetProtocol(), config.GetHost())
	requestURI := fmt.Sprintf("%s/%s", baseURI, instance.InstancePeersPath)
	ctx := suite.newContext(recorder, http.MethodGet, requestURI, nil, "", true)

	instanceModule.InstancePeersGETHandler(ctx)

	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	out := testrig.MustJSONStringFromBytes(b)
	suite.Equal(`[
  "example.org",
  "fossbros-anonymous.io",
  "thequeenisstillalive.technology",
  "ëxample.org"
]`, out)
}

func (suite *InstancePeersGetTestSuite) TestInstancePeersGetOnlySuspended() {
	testStructs := testrig.SetupTestStructs(rMediaPath, rTemplatePath)
	defer testrig.TearDownTestStructs(testStructs)
	instanceModule := instance.New(testStructs.Processor, testStructs.Templates)

	recorder := httptest.NewRecorder()
	baseURI := fmt.Sprintf("%s://%s", config.GetProtocol(), config.GetHost())
	requestURI := fmt.Sprintf("%s/%s?filter=suspended", baseURI, instance.InstancePeersPath)
	ctx := suite.newContext(recorder, http.MethodGet, requestURI, nil, "", false)

	instanceModule.InstancePeersGETHandler(ctx)

	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	out := testrig.MustJSONStringFromBytes(b)
	suite.Equal(`[
  {
    "comment": "reply-guying to tech posts",
    "domain": "replyguys.com",
    "severity": "suspend",
    "suspended_at": "2020-05-13T13:29:12.000Z"
  }
]`, out)
}

func (suite *InstancePeersGetTestSuite) TestInstancePeersGetOnlySuspendedUnauthorized() {
	testStructs := testrig.SetupTestStructs(rMediaPath, rTemplatePath)
	defer testrig.TearDownTestStructs(testStructs)
	instanceModule := instance.New(testStructs.Processor, testStructs.Templates)

	config.SetInstanceExposeBlocklist(false)

	recorder := httptest.NewRecorder()
	baseURI := fmt.Sprintf("%s://%s", config.GetProtocol(), config.GetHost())
	requestURI := fmt.Sprintf("%s/%s?filter=suspended", baseURI, instance.InstancePeersPath)
	ctx := suite.newContext(recorder, http.MethodGet, requestURI, nil, "", false)

	instanceModule.InstancePeersGETHandler(ctx)

	suite.Equal(http.StatusUnauthorized, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"Unauthorized: peers blocked query requires an authenticated account/user"}`, string(b))
}

func (suite *InstancePeersGetTestSuite) TestInstancePeersGetOnlySuspendedAuthorized() {
	testStructs := testrig.SetupTestStructs(rMediaPath, rTemplatePath)
	defer testrig.TearDownTestStructs(testStructs)
	instanceModule := instance.New(testStructs.Processor, testStructs.Templates)

	config.SetInstanceExposeBlocklist(false)

	recorder := httptest.NewRecorder()
	baseURI := fmt.Sprintf("%s://%s", config.GetProtocol(), config.GetHost())
	requestURI := fmt.Sprintf("%s/%s?filter=suspended", baseURI, instance.InstancePeersPath)
	ctx := suite.newContext(recorder, http.MethodGet, requestURI, nil, "", true)

	instanceModule.InstancePeersGETHandler(ctx)

	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	out := testrig.MustJSONStringFromBytes(b)
	suite.Equal(`[
  {
    "comment": "reply-guying to tech posts",
    "domain": "replyguys.com",
    "severity": "suspend",
    "suspended_at": "2020-05-13T13:29:12.000Z"
  }
]`, out)
}

func (suite *InstancePeersGetTestSuite) TestInstancePeersGetAll() {
	testStructs := testrig.SetupTestStructs(rMediaPath, rTemplatePath)
	defer testrig.TearDownTestStructs(testStructs)
	instanceModule := instance.New(testStructs.Processor, testStructs.Templates)

	recorder := httptest.NewRecorder()
	baseURI := fmt.Sprintf("%s://%s", config.GetProtocol(), config.GetHost())
	requestURI := fmt.Sprintf("%s/%s?filter=suspended,open", baseURI, instance.InstancePeersPath)
	ctx := suite.newContext(recorder, http.MethodGet, requestURI, nil, "", false)

	instanceModule.InstancePeersGETHandler(ctx)

	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	out := testrig.MustJSONStringFromBytes(b)
	suite.Equal(`[
  {
    "domain": "example.org"
  },
  {
    "domain": "fossbros-anonymous.io"
  },
  {
    "comment": "reply-guying to tech posts",
    "domain": "replyguys.com",
    "severity": "suspend",
    "suspended_at": "2020-05-13T13:29:12.000Z"
  },
  {
    "domain": "thequeenisstillalive.technology"
  },
  {
    "domain": "ëxample.org"
  }
]`, out)
}

func (suite *InstancePeersGetTestSuite) TestInstancePeersGetAllowed() {
	testStructs := testrig.SetupTestStructs(rMediaPath, rTemplatePath)
	defer testrig.TearDownTestStructs(testStructs)
	instanceModule := instance.New(testStructs.Processor, testStructs.Templates)

	recorder := httptest.NewRecorder()
	baseURI := fmt.Sprintf("%s://%s", config.GetProtocol(), config.GetHost())
	requestURI := fmt.Sprintf("%s/%s?filter=allowed", baseURI, instance.InstancePeersPath)
	ctx := suite.newContext(recorder, http.MethodGet, requestURI, nil, "", false)

	instanceModule.InstancePeersGETHandler(ctx)

	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	out := testrig.MustJSONStringFromBytes(b)
	suite.Equal(`[]`, out)
}

func (suite *InstancePeersGetTestSuite) TestInstancePeersGetAllWithObfuscated() {
	testStructs := testrig.SetupTestStructs(rMediaPath, rTemplatePath)
	defer testrig.TearDownTestStructs(testStructs)
	instanceModule := instance.New(testStructs.Processor, testStructs.Templates)

	err := testStructs.State.DB.Put(suite.T().Context(), &gtsmodel.DomainBlock{
		ID:                 "01G633XTNK51GBADQZFZQDP6WR",
		CreatedAt:          testrig.TimeMustParse("2021-06-09T12:34:55+02:00"),
		UpdatedAt:          testrig.TimeMustParse("2021-06-09T12:34:55+02:00"),
		Domain:             "omg.just.the.worst.org.ever",
		CreatedByAccountID: "01F8MH17FWEB39HZJ76B6VXSKF",
		PublicComment:      "just absolutely the worst, wowza",
		Obfuscate:          util.Ptr(true),
	})
	suite.NoError(err)

	recorder := httptest.NewRecorder()
	baseURI := fmt.Sprintf("%s://%s", config.GetProtocol(), config.GetHost())
	requestURI := fmt.Sprintf("%s/%s?filter=suspended,open", baseURI, instance.InstancePeersPath)
	ctx := suite.newContext(recorder, http.MethodGet, requestURI, nil, "", false)

	instanceModule.InstancePeersGETHandler(ctx)

	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	out := testrig.MustJSONStringFromBytes(b)
	suite.Equal(`[
  {
    "domain": "example.org"
  },
  {
    "domain": "fossbros-anonymous.io"
  },
  {
    "comment": "just absolutely the worst, wowza",
    "domain": "o*g.*u**.t**.*or*t.*r**ev**",
    "severity": "suspend",
    "suspended_at": "2021-06-09T10:34:55.000Z"
  },
  {
    "comment": "reply-guying to tech posts",
    "domain": "replyguys.com",
    "severity": "suspend",
    "suspended_at": "2020-05-13T13:29:12.000Z"
  },
  {
    "domain": "thequeenisstillalive.technology"
  },
  {
    "domain": "ëxample.org"
  }
]`, out)
}

func (suite *InstancePeersGetTestSuite) TestInstancePeersGetAllWithObfuscatedFlat() {
	testStructs := testrig.SetupTestStructs(rMediaPath, rTemplatePath)
	defer testrig.TearDownTestStructs(testStructs)
	instanceModule := instance.New(testStructs.Processor, testStructs.Templates)

	err := testStructs.State.DB.Put(suite.T().Context(), &gtsmodel.DomainBlock{
		ID:                 "01G633XTNK51GBADQZFZQDP6WR",
		CreatedAt:          testrig.TimeMustParse("2021-06-09T12:34:55+02:00"),
		UpdatedAt:          testrig.TimeMustParse("2021-06-09T12:34:55+02:00"),
		Domain:             "omg.just.the.worst.org.ever",
		CreatedByAccountID: "01F8MH17FWEB39HZJ76B6VXSKF",
		PublicComment:      "just absolutely the worst, wowza",
		Obfuscate:          util.Ptr(true),
	})
	suite.NoError(err)

	recorder := httptest.NewRecorder()
	baseURI := fmt.Sprintf("%s://%s", config.GetProtocol(), config.GetHost())
	requestURI := fmt.Sprintf("%s/%s?filter=suspended,open&flat=true", baseURI, instance.InstancePeersPath)
	ctx := suite.newContext(recorder, http.MethodGet, requestURI, nil, "", false)

	instanceModule.InstancePeersGETHandler(ctx)

	suite.Equal(http.StatusOK, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	out := testrig.MustJSONStringFromBytes(b)
	suite.Equal(`[
  "example.org",
  "fossbros-anonymous.io",
  "o*g.*u**.t**.*or*t.*r**ev**",
  "replyguys.com",
  "thequeenisstillalive.technology",
  "ëxample.org"
]`, out)
}

func (suite *InstancePeersGetTestSuite) TestInstancePeersGetFunkyParams() {
	testStructs := testrig.SetupTestStructs(rMediaPath, rTemplatePath)
	defer testrig.TearDownTestStructs(testStructs)
	instanceModule := instance.New(testStructs.Processor, testStructs.Templates)

	recorder := httptest.NewRecorder()
	baseURI := fmt.Sprintf("%s://%s", config.GetProtocol(), config.GetHost())
	requestURI := fmt.Sprintf("%s/%s?filter=aaaaaaaaaaaaaaaaa,open", baseURI, instance.InstancePeersPath)
	ctx := suite.newContext(recorder, http.MethodGet, requestURI, nil, "", true)

	instanceModule.InstancePeersGETHandler(ctx)

	suite.Equal(http.StatusBadRequest, recorder.Code)

	result := recorder.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	suite.NoError(err)

	suite.Equal(`{"error":"Bad Request: filter aaaaaaaaaaaaaaaaa not recognized; accepted values are 'open', 'blocked', 'allowed', and 'suspended' (deprecated)"}`, string(b))
}

func TestInstancePeersGetTestSuite(t *testing.T) {
	suite.Run(t, &InstancePeersGetTestSuite{})
}
