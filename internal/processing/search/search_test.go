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

package search_test

import (
	"code.superseriousbusiness.org/gotosocial/internal/filter/mutes"
	"code.superseriousbusiness.org/gotosocial/internal/filter/status"
	"code.superseriousbusiness.org/gotosocial/internal/filter/visibility"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/processing/common"
	"code.superseriousbusiness.org/gotosocial/internal/processing/search"
	"code.superseriousbusiness.org/gotosocial/testrig"
)

const (
	rMediaPath    = "../../../testrig/media"
	rTemplatePath = "../../../web/template"
)

type SearchTestSuite struct {
	proc     *search.Processor
	test     *testrig.TestStructs
	accounts map[string]*gtsmodel.Account
	statuses map[string]*gtsmodel.Status
	hashtags map[string]*gtsmodel.Tag
}

func (suite *SearchTestSuite) SetupTest() {
	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.test = testrig.SetupTestStructs(rMediaPath, rTemplatePath)

	media := testrig.NewTestMediaManager(suite.test.State)

	federator := testrig.NewTestFederator(suite.test.State, suite.test.TransportController, media)

	visibility := visibility.NewFilter(suite.test.State)
	mutes := mutes.NewFilter(suite.test.State)
	status := status.NewFilter(suite.test.State)

	common := common.New(suite.test.State, media, suite.test.TypeConverter, federator, visibility, mutes, status, nil, suite.test.Surfacer)
	search := search.New(&common, suite.test.State, federator, suite.test.TypeConverter, visibility, suite.test.Surfacer)
	suite.proc = &search

	suite.accounts = testrig.NewTestAccounts()
	suite.statuses = testrig.NewTestStatuses()
}

func (suite *SearchTestSuite) TeardownTest() {
	testrig.TearDownTestStructs(suite.test)
}
