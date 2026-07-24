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

package config

import (
	"fmt"

	"codeberg.org/gruf/go-split"
	"github.com/spf13/cast"
)

// joinFlag joins 2 not-empty flag parts with a hyphen '-',
// else just returns the single not-empty flag part.
func joinFlag(p1, p2 string) string {
	if p1 == "" {
		return p2
	} else if p2 == "" {
		return p1
	}
	return p1 + "-" + p2
}

// toStringSlice attempts to coerce input variable to a string slice,
// preferably by some form of casting, else splitting as comma-separated.
func toStringSlice(a any) ([]string, error) {
	switch a := a.(type) {
	case []string:
		return a, nil
	case string:
		return split.SplitStrings[string](a)
	case []any:
		ss := make([]string, len(a))
		for i, a := range a {
			var err error
			ss[i], err = cast.ToStringE(a)
			if err != nil {
				return nil, err
			}
		}
		return ss, nil
	default:
		return nil, fmt.Errorf("cannot cast %T to []string", a)
	}
}

// flattenConfigMap ...
func flattenConfigMap(m map[string]any) {
	var flatten func(src map[string]any, prefix string)
	flatten = func(src map[string]any, prefix string) {
		for k, v := range src {
			switch v := v.(type) {
			case map[string]any:
				flatten(v, joinFlag(prefix, k))
			default:
				m[joinFlag(prefix, k)] = v
			}
		}
	}
	for k, v := range m {
		switch v := v.(type) { //nolint
		case map[string]any:
			flatten(v, k)
			delete(m, k)
		}
	}
}
