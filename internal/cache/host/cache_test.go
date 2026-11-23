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

package host_test

import (
	"errors"
	"strings"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/cache/host"
)

func hostCache(t *testing.T) (*host.Cache, func() ([]string, error)) {
	c := new(host.Cache)

	cached := []string{
		"google.com",               //
		"mail.google.com",          // should be ignored since covered above
		"dev.mail.google.com",      // same again
		"google.co.uk",             //
		"mail.google.co.uk",        //
		"pleroma.bad.host",         //
		"pleroma.still.a.bad.host", //
		"127.0.0.1",                //
		"127.0.0.1/32",             // should be ignored since covered above
		"127.0.1.0/24",             //
		"127.0.1.1",                // should be ignored since covered above
		"127.0.1.0/28",             // should be ignored since covered above
		"127.0.1.16/28",            // should be ignored since covered above
	}

	loader := func() ([]string, error) {
		t.Log("load: returning cached hosts")
		return cached, nil
	}

	return c, loader
}

func TestMatch(t *testing.T) {
	// Create test domain cache.
	c, loader := hostCache(t)

	// Check known matching hosts.
	for _, host := range []string{
		"google.com",
		"mail.google.com",
		"dev.mail.google.com",
		"google.co.uk",
		"mail.google.co.uk",
		"pleroma.bad.host",
		"dev.pleroma.bad.host",
		"pleroma.still.a.bad.host",
		"dev.pleroma.still.a.bad.host",
		"127.0.0.1",
		"127.0.1.0",
		"127.0.1.1",
		"127.0.1.255",
	} {
		t.Logf("checking matches: %s", host)
		if b, _ := c.Matches(host, loader); !b {
			t.Fatalf("should be matched: %s", host)
		}
	}

	// Check known unmatched hosts.
	for _, host := range []string{
		"askjeeves.com",
		"ask-kim.co.uk",
		"google.ie",
		"mail.google.ie",
		"gts.bad.host",
		"mastodon.bad.host",
		"akkoma.still.a.bad.host",
		"127.0.0.2",
	} {
		t.Logf("checking isn't matched: %s", host)
		if b, _ := c.Matches(host, loader); b {
			t.Fatalf("should not be matched: %s", host)
		}
	}

	t.Log(c.String())
}

func TestMatchesOn(t *testing.T) {
	// Create test domain cache.
	c, loader := hostCache(t)

	// Check known matching domains.
	for _, domain := range []string{
		"google.com",
		"mail.google.com",
		"dev.mail.google.com",
		"google.co.uk",
		"mail.google.co.uk",
		"pleroma.bad.host",
		"dev.pleroma.bad.host",
		"pleroma.still.a.bad.host",
		"dev.pleroma.still.a.bad.host",
	} {
		matchedOn, _ := c.MatchesOn(domain, loader)
		t.Logf("checking domain matches: domain=%s matchedOn=%s", domain, matchedOn)
		if !strings.HasSuffix(domain, matchedOn) {
			t.Fatalf("matched domain %s should match on %s", domain, matchedOn)
		}
	}
}

func TestHydrateError(t *testing.T) {
	var c host.Cache
	knownErr := errors.New("known error")

	// Check hydrate is actually performed and returns error.
	if _, err := c.Matches("", func() ([]string, error) {
		t.Log("load: returning known error")
		return nil, knownErr
	}); !errors.Is(err, knownErr) {
		t.Fatalf("matches did not return expected error: %v", err)
	}
}
