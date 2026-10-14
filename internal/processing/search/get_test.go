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
	"slices"
	"strings"
	"testing"

	"code.superseriousbusiness.org/gopkg/log"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"github.com/stretchr/testify/assert"
)

func TestGetAccounts(t *testing.T) {
	suite := new(SearchTestSuite)
	suite.SetupTest()
	defer suite.TeardownTest()
	suite.TestGetAccounts(t)
}

func TestGetStatuses(t *testing.T) {
	suite := new(SearchTestSuite)
	suite.SetupTest()
	defer suite.TeardownTest()
	suite.TestGetStatuses(t)
}

func (suite *SearchTestSuite) TestGetAccounts(t *testing.T) {
	type testcase struct {
		requester string   // testrig account model key
		query     string   //
		resolve   bool     //
		follows   bool     //
		expect    []string // testrig account model keys
	}

	// Test performs a single instance of testcase{}.
	test := func(t *testing.T, testcase testcase) {
		t.Helper()
		accounts := suite.testGetAccounts(t,
			testcase.requester,
			testcase.query,
			testcase.resolve,
			testcase.follows,
		)
		must(assert.Len(t, accounts, len(testcase.expect)))
		for _, key := range testcase.expect {
			id := suite.accounts[key].ID
			must(slices.ContainsFunc(accounts,
				func(a *apimodel.Account) bool {
					return a.ID == id
				}))
		}
	}

	// Perform a series of username
	// specific get accounts tests.
	// Where for the same query we
	// test multiple times with query
	// formatting that should all be equal.
	for _, each := range []struct {
		model    string // testrig model
		username string // username query
		domain   string // domain query
		local    bool   // local account?
	}{
		{
			model:    "local_account_1",
			username: "zork",
			local:    true,
		},
		{
			model:    "local_account_1",
			username: "the_mighty",
			domain:   "localhost:8080",
			local:    true,
		},
		{
			model:    "local_account_1",
			username: "the_mighty_zork",
			local:    true,
		},

		{
			model:    "local_account_2",
			username: "turtle",
			local:    true,
		},
		{
			model:    "local_account_2",
			username: "1happy",
			local:    true,
		},
		{
			model:    "local_account_2",
			username: "happyturtle",
			local:    true,
		},
		{
			model:    "local_account_2",
			username: "1happyturtle",
			domain:   "localhost:8080",
			local:    true,
		},

		{
			model:    "local_account_3",
			username: "media",
			local:    true,
		},
		{
			model:    "local_account_3",
			username: "mogul",
			local:    true,
		},
		{
			model:    "local_account_3",
			username: "media_mogul",
			domain:   "localhost:8080",
			local:    true,
		},

		{
			model:    "remote_account_1",
			username: "foss",
			local:    false,
		},
		{
			model:    "remote_account_1",
			username: "satan",
			domain:   "fossbros",
			local:    false,
		},
		{
			model:    "remote_account_1",
			username: "foss_satan",
			domain:   "fossbros-anonymous.io",
			local:    false,
		},

		{
			model:    "remote_account_2",
			username: "some",
			local:    false,
		},
		{
			model:    "remote_account_2",
			username: "user",
			domain:   "example",
			local:    false,
		},
		{
			model:    "remote_account_2",
			username: "some_user",
			domain:   "example.org",
			local:    false,
		},

		{
			model:    "remote_account_3",
			username: "fuckin_maj",
			local:    false,
		},
		{
			model:    "remote_account_3",
			username: "her",
			domain:   "thequeen",
			local:    false,
		},
		{
			model:    "remote_account_3",
			username: "her_fuckin_maj",
			domain:   "thequeenisstillalive.technology",
			local:    false,
		},

		{
			model:    "remote_account_4",
			username: "üser",
			local:    false,
		},
		{
			model:    "remote_account_4",
			username: "üser",
			domain:   "xn--xample-ova.org",
			local:    false,
		},
	} {
		orig := test

		// Wrap the test function to always include
		// the domain search for this user query test.
		//
		// We only bother making username permutations
		// as they both follow the same logic / code paths.
		test := func(t *testing.T, testcase testcase) {
			testcase.query += " domain:" + each.domain
			orig(t, testcase)
		}

		// Perform each testcase in a separate sub-test instance.
		t.Run(log.Formatted(each).String(), func(t *testing.T) {
			testcase := testcase{requester: "local_account_1"}
			testcase.expect = []string{each.model}

			// Perform the actual username query string.
			testcase.query = "username:" + each.username
			test(t, testcase)

			// Ensure that quotes are appropriately stripped.
			testcase.query = "username:\"" + each.username + "\""
			test(t, testcase)

			// Ensure that spaces are appropriately ignored.
			testcase.query = "username:\" " + each.username + " \""
			test(t, testcase)

			// Ensure that casing doesn't matter for search querying.
			testcase.query = "username:" + strings.ToUpper(each.username)
			test(t, testcase)

			if each.local {
				// Verify for remotes that is:local doesn't filter it out.
				testcase.query = "username:" + each.username + " is:local"
				test(t, testcase)

				// Verify for local that is:remote *does* filter it out.
				testcase.query = "username:" + each.username + " is:remote"
				testcase.expect = nil
				test(t, testcase)
			} else {
				// Verify for remotes that is:remote doesn't filter it out.
				testcase.query = "username:" + each.username + " is:remote"
				test(t, testcase)

				// Verify for remote that is:local *does* filter it out.
				testcase.query = "username:" + each.username + " is:local"
				testcase.expect = nil
				test(t, testcase)
			}
		})
	}

	// Perform account note searches.
	for _, each := range []struct {
		model string // testrig model
		note  string // note query
	}{
		{
			model: "local_account_1",
			note:  "yo",
		},
		{
			model: "local_account_1",
			note:  "this is my profile",
		},

		{
			model: "local_account_2",
			note:  "concern me",
		},

		{
			model: "local_account_3",
			note:  "shitload",
		},
		{
			model: "local_account_3",
			note:  "gallery",
		},

		{
			model: "remote_account_1",
			note:  "i dunno",
		},
		{
			model: "remote_account_1",
			note:  "whatever",
		},

		{
			model: "remote_account_2",
			note:  "real",
		},
		{
			model: "remote_account_2",
			note:  "son of a gun",
		},

		{
			model: "remote_account_3",
			note:  "blame charles",
		},
		{
			model: "remote_account_3",
			note:  "if i die",
		},
		{
			model: "remote_account_3",
			note:  "king",
		},
	} {
		// Perform each testcase in a separate sub-test instance.
		t.Run(log.Formatted(each).String(), func(t *testing.T) {
			testcase := testcase{requester: "local_account_1"}
			testcase.expect = []string{each.model}

			testcase.query = "note:\"" + each.note + "\""
			test(t, testcase)
		})
	}

	// Perform a series of direct
	// lookups by username@domain / URL.
	for _, each := range []struct {
		model string // testrig model
		query string // direct query
	}{
		{
			model: "local_account_1",
			query: "the_mighty_zork@localhost:8080",
		},
		{
			model: "local_account_1",
			query: "@the_mighty_zork@localhost:8080",
		},
		{
			model: "local_account_1",
			query: "http://localhost:8080/users/the_mighty_zork",
		},
		{
			model: "local_account_1",
			query: "http://localhost:8080/@the_mighty_zork",
		},

		{
			model: "local_account_2",
			query: "1happyturtle@localhost:8080",
		},
		{
			model: "local_account_2",
			query: "@1happyturtle@localhost:8080",
		},
		{
			model: "local_account_2",
			query: "http://localhost:8080/users/1happyturtle",
		},
		{
			model: "local_account_2",
			query: "http://localhost:8080/@1happyturtle",
		},

		{
			model: "local_account_3",
			query: "media_mogul@localhost:8080",
		},
		{
			model: "local_account_3",
			query: "@media_mogul@localhost:8080",
		},
		{
			model: "local_account_3",
			query: "http://localhost:8080/users/media_mogul",
		},
		{
			model: "local_account_3",
			query: "http://localhost:8080/@media_mogul",
		},

		{
			model: "remote_account_1",
			query: "foss_satan@fossbros-anonymous.io",
		},
		{
			model: "remote_account_1",
			query: "@foss_satan@fossbros-anonymous.io",
		},
		{
			model: "remote_account_1",
			query: "http://fossbros-anonymous.io/users/foss_satan",
		},
		{
			model: "remote_account_1",
			query: "http://fossbros-anonymous.io/@foss_satan",
		},

		{
			model: "remote_account_2",
			query: "Some_User@example.org",
		},
		{
			model: "remote_account_2",
			query: "@Some_User@example.org",
		},
		{
			model: "remote_account_2",
			query: "http://example.org/users/Some_User",
		},
		{
			model: "remote_account_2",
			query: "http://example.org/@Some_User",
		},

		{
			model: "remote_account_3",
			query: "her_fuckin_maj@thequeenisstillalive.technology",
		},
		{
			model: "remote_account_3",
			query: "@her_fuckin_maj@thequeenisstillalive.technology",
		},
		{
			model: "remote_account_3",
			query: "http://thequeenisstillalive.technology/users/her_fuckin_maj",
		},
		{
			model: "remote_account_3",
			query: "http://thequeenisstillalive.technology/@her_fuckin_maj",
		},

		{
			model: "remote_account_4",
			query: "üser@xn--xample-ova.org",
		},
		{
			model: "remote_account_4",
			query: "@üser@xn--xample-ova.org",
		},
		{
			model: "remote_account_4",
			query: "https://xn--xample-ova.org/users/%C3%BCser",
		},
	} {
		// Perform each testcase in a separate sub-test instance.
		t.Run(log.Formatted(each).String(), func(t *testing.T) {
			testcase := testcase{requester: "local_account_1"}
			testcase.expect = []string{each.model}
			testcase.query = each.query
			test(t, testcase)
		})
	}
}

// testGetAccounts searches for accounts on both the
// GetAccounts() and generic Get() endpoint ensuring
// they both return the same set of results, and returns those.
func (suite *SearchTestSuite) testGetAccounts(
	t *testing.T,
	requesterKey string,
	query string,
	resolve bool,
	follows bool,
) []*apimodel.Account {
	t.Helper()

	// Perform request as keyed test account.
	requester := suite.accounts[requesterKey]

	// Search for accounts according to params.
	res1, _, _, err := suite.proc.GetAccounts(
		t.Context(),
		requester,
		query,
		resolve,
		follows,
		&paging.Page{Limit: 100},
	)
	must(assert.NoError(t, err))

	// Perform the same query on the
	// generic 'get' search endpoint.
	res2, _, _, err := suite.proc.Get(
		t.Context(),
		requester,
		query,
		resolve,
		follows,
		"",
		&paging.Page{Limit: 100},
	)
	must(assert.NoError(t, err))

	// Ensure result accounts returned from both endpoints are equal.
	must(assert.True(t, slices.EqualFunc(res1.Accounts, res2.Accounts,
		func(a1, a2 *apimodel.Account) bool {
			return a1.ID == a2.ID
		})))

	// Return results.
	return res1.Accounts
}

func (suite *SearchTestSuite) TestGetStatuses(t *testing.T) {
	type testcase struct {
		requester string   // testrig status model key
		query     string   //
		resolve   bool     //
		follows   bool     //
		author    string   // testrig account model key
		expect    []string // testrig status model keys
	}

	// Test performs a single instance of testcase{}.
	test := func(t *testing.T, testcase testcase) {
		t.Helper()
		statuses := suite.testGetStatuses(t,
			testcase.requester,
			testcase.query,
			testcase.resolve,
			testcase.follows,
			testcase.author,
		)
		must(assert.Len(t, statuses, len(testcase.expect)))
		for _, key := range testcase.expect {
			id := suite.statuses[key].ID
			must(slices.ContainsFunc(statuses,
				func(s *apimodel.Status) bool {
					return s.ID == id
				}))
		}
	}

	for _, each := range []struct {
		content string   // content query
		warning string   // content warning query
		author  string   // author testrig model key
		expect  []string // expected testrig models
	}{
		{
			content: "hello world",
			warning: "",
			author:  "",
			expect:  []string{"admin_account_status_1"},
		},
		{
			content: "",
			warning: "puppies",
			author:  "",
			expect:  []string{"admin_account_status_2"},
		},
		{
			content: "hello",
			warning: "",
			author:  "",
			expect:  []string{"admin_account_status_1", "local_account_1_status_1"},
		},
		{
			content: "hello",
			warning: "",
			author:  "admin_account",
			expect:  []string{"admin_account_status_1"},
		},
		{
			content: "hello",
			warning: "",
			author:  "local_account_1",
			expect:  []string{"local_account_1_status_1"},
		},
		{
			// non-public status, but it is authored by
			// requester so should be visible in results.
			content: "very personal post",
			warning: "",
			author:  "",
			expect:  []string{"local_account_1_status_3"},
		},
		{
			// non-public status, authored by someone the
			// requester isn't following, shouldn't be returned.
			content: "some media for ya",
			warning: "",
			author:  "",
			expect:  []string{},
		},
	} {
		// Perform each testcase in a separate sub-test instance.
		t.Run(log.Formatted(each).String(), func(t *testing.T) {
			testcase := testcase{requester: "local_account_1"}
			testcase.expect = each.expect
			testcase.author = each.author
			testcase.query += " content:\"" + each.content + "\""
			testcase.query += " content_warning:\"" + each.warning + "\""
			test(t, testcase)
		})
	}
}

// testGetStatuses searches for statuses on both the
// GetStatuses() and generic Get() endpoint ensuring
// they both return the same set of results, and returns those.
func (suite *SearchTestSuite) testGetStatuses(
	t *testing.T,
	requesterKey string,
	query string,
	resolve bool,
	follows bool,
	authorKey string,
) []*apimodel.Status {
	t.Helper()

	// Perform request as keyed test account.
	requester := suite.accounts[requesterKey]

	var authorID string
	if authorKey != "" {
		// Look for statuses by keyed test account.
		authorID = suite.accounts[authorKey].ID
	}

	// Search for accounts according to params.
	res1, _, _, err := suite.proc.GetStatuses(
		t.Context(),
		requester,
		query,
		resolve,
		follows,
		authorID,
		&paging.Page{Limit: 100},
	)
	must(assert.NoError(t, err))

	// Perform the same query on the
	// generic 'get' search endpoint.
	res2, _, _, err := suite.proc.Get(
		t.Context(),
		requester,
		query,
		resolve,
		follows,
		authorID,
		&paging.Page{Limit: 100},
	)
	must(assert.NoError(t, err))

	// Ensure result statuses returned from both endpoints are equal.
	must(assert.True(t, slices.EqualFunc(res1.Statuses, res2.Statuses,
		func(s1, s2 *apimodel.Status) bool {
			return s1.ID == s2.ID
		})))

	// Return results.
	return res1.Statuses
}

func must(ok bool) {
	if !ok {
		panic("!ok")
	}
}
