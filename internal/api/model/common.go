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

package model

import (
	"errors"
	"strconv"
	"unsafe"

	"codeberg.org/gruf/go-longdur"
)

// DurationOrDays wraps a longdur.Duration type to also
// support unmarshaling direct integers as day counts.
// The zero value is (un)marshaled as JSON null or empty.
//
// NOTE: this type largely exists as a transitionary type
// to handle our older API calls requiring integer number
// of days, while a "longdur" is much more flexible.
type DurationOrDays struct{ longdur.Duration }

// MarshalJSON: implements json.Marshaler{}.
func (d DurationOrDays) MarshalJSON() ([]byte, error) {
	if d.Duration == 0 {
		return []byte("null"), nil
	}
	return []byte("\"" + d.String() + "\""), nil
}

// UnmarshalJSON: implements json.Unmarshaler{}.
func (d *DurationOrDays) UnmarshalJSON(data []byte) error {
	if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
		data = data[1 : len(data)-1]
	}
	return d.unmarshalb(data)
}

// MarshalText: implements encoding.TextMarshaler{}.
func (d DurationOrDays) MarshalText() ([]byte, error) {
	if d.Duration == 0 {
		return []byte(""), nil
	}
	return d.Duration.MarshalText()
}

// UnmarshalText: implements encoding.TextUmarshaler{}.
func (d *DurationOrDays) UnmarshalText(text []byte) error {
	return d.unmarshalb(text)
}

// UnmarshalParam: implements binding.BindUnmarshaler{}.
func (d *DurationOrDays) UnmarshalParam(param string) error {
	return d.unmarshal(param)
}

// unmarshalb converts byte slice to string (with length check) and calls d.unmarshal().
func (d *DurationOrDays) unmarshalb(b []byte) error {
	if len(b) == 0 {
		return errors.New("invalid duration")
	} else if string(b) == "null" {
		d.Duration = 0
		return nil
	}
	return d.unmarshal(unsafe.String(&b[0], len(b)))
}

// unmarshal attempts to unmarshal string as either integer, or string encoded duration.
func (d *DurationOrDays) unmarshal(str string) error {

	// Initially, try to parse as an integer.
	i, err := strconv.ParseUint(str, 10, 64)
	if err == nil {

		// Set number of days from integer provided.
		d.Duration = longdur.Duration(i) * longdur.Day
		return nil
	}

	// Parse this as a duration.
	return d.Duration.Set(str)
}
