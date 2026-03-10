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

package db

import (
	"database/sql/driver"

	"github.com/uptrace/bun"
)

// ToNamedValues converts older driver.Value types to driver.NamedValue types.
func ToNamedValues(args []driver.Value) []driver.NamedValue {
	if args == nil {
		return nil
	}
	args2 := make([]driver.NamedValue, len(args))
	for i := range args {
		args2[i] = driver.NamedValue{
			Ordinal: i + 1,
			Value:   args[i],
		}
	}
	return args2
}

// BitNotSet returns a set of arguments for a bun
// expression indicating whether a bit flag IS set.
func BitIsSet[Type ~int16](col string, value Type) (
	string, // query string
	bun.Ident, // column ident
	Type, // bit flag
) {
	return "? & ? != 0", bun.Ident(col), value
}

// BitNotSet returns a set of arguments for a bun
// expression indicating whether a bit flag IS NOT set.
func BitNotSet[Type ~int16](col string, value Type) (
	string, // query string
	bun.Ident, // column ident
	Type, // bit flag
) {
	return "? & ? = 0", bun.Ident(col), value
}
