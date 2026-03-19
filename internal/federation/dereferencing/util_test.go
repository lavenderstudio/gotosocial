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

package dereferencing

import (
	"strconv"
	"testing"
)

func TestKeyedList(t *testing.T) {
	var list keyedList[string]

	for i := 0; i < 100; i++ {
		str := strconv.Itoa(i)
		list.put(str, str)
	}

	for i := 0; i < 100; i++ {
		str := strconv.Itoa(i)
		if list.get(str) != str {
			t.Fatal()
		}
	}

	for i := 0; len(list) > 1; i++ {
		str := strconv.Itoa(i)
		before := len(list)
		list.delete(str)
		if len(list) == before {
			t.Fatal("failed to decrease list size for key " + str)
		}
	}

	list.delete(list[0].k)
	if cap(list) != 0 {
		t.Fatal()
	}
}
