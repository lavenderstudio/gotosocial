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

package config_test

import (
	"net/netip"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestParsePrefixSingleAddress(t *testing.T) {
	{
		prefix, err := config.ParsePrefix("127.0.0.1")
		assert.NoError(t, err)

		ip := netip.MustParseAddr("127.0.0.1")
		ip2 := netip.MustParseAddr("127.0.0.2")
		ip3 := netip.MustParseAddr("127.0.0.0")

		assert.True(t, prefix.Contains(ip))
		assert.False(t, prefix.Contains(ip2))
		assert.False(t, prefix.Contains(ip3))
	}

	{
		prefix, err := config.ParsePrefix("10.1.0.0")
		assert.NoError(t, err)

		ip := netip.MustParseAddr("10.1.0.0")
		ip2 := netip.MustParseAddr("10.1.0.1")
		ip3 := netip.MustParseAddr("10.2.0.0")
		ip4 := netip.MustParseAddr("10.0.0.0")

		assert.True(t, prefix.Contains(ip))
		assert.False(t, prefix.Contains(ip2))
		assert.False(t, prefix.Contains(ip3))
		assert.False(t, prefix.Contains(ip4))
	}
}
