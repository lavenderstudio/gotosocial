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

package xencoding

import (
	"errors"
	"io"
)

var emptydst = &struct{}{}

// Decode performs a decode of given reader, for given new decoder allocator, into destination.
// This also performs a second decode into an empty destination to ensure no trailing garbage.
func Decode[Decoder interface{ Decode(any) error }](r io.Reader, new func(io.Reader) Decoder, dst any) error {

	// Create
	// decoder.
	dec := new(r)

	// Attempt main decode into destination.
	if err := dec.Decode(dst); err != nil {
		return err
	}

	// Ensure there isn't trailing data after decode.
	if err := dec.Decode(emptydst); err != io.EOF {
		return errors.New("data remaining after first decode")
	}

	return nil
}
