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
	"errors"
	"net/netip"
	"strings"

	"github.com/hashicorp/cronexpr"
)

// Deprecated is a placeholder type
// for use with config fields that have
// the "deprecated-by" field tag set.
type Deprecated string

// CronExpression is a wrapper for cronexpr.Expression
// to allow parsing by CLI "flag"-like utilities.
type CronExpression struct {
	*cronexpr.Expression
	Expr string
}

func MustParseCron(expr string) (cron CronExpression) {
	if err := cron.Set(expr); err != nil {
		panic(err)
	}
	return
}

func (expr *CronExpression) Set(in string) (err error) {
	if in == "" {
		return
	}
	expr.Expr = in // set the raw expression string
	expr.Expression, err = cronexpr.Parse(in)
	return
}

func (expr *CronExpression) MarshalText() ([]byte, error) {
	return []byte(expr.Expr), nil
}

func (expr *CronExpression) UnmarshalText(text []byte) error {
	return expr.Set(string(text))
}

func (expr *CronExpression) String() string {
	return expr.Expr
}

// IPPrefixes is a type-alias for []netip.Prefix
// to allow parsing by CLI "flag"-like utilities.
type IPPrefixes []netip.Prefix

func (p *IPPrefixes) Set(in string) error {
	prefix, err := netip.ParsePrefix(in)
	if err != nil {
		return err
	}
	(*p) = append((*p), prefix)
	return nil
}

func (p *IPPrefixes) Strings() []string {
	if len(*p) == 0 {
		return nil
	}
	strs := make([]string, len(*p))
	for i, prefix := range *p {
		strs[i] = prefix.String()
	}
	return strs
}

type InstanceDirectoryMode int16

const (
	InstanceDirectoryModeUnknown InstanceDirectoryMode = iota
	InstanceDirectoryModeOff
	InstanceDirectoryModeWebOnly
	InstanceDirectoryModeOpen
)

// MarshalText implements encoding.TextMarshaler{}.
func (i InstanceDirectoryMode) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler{}.
func (i *InstanceDirectoryMode) UnmarshalText(text []byte) error {
	return i.Set(string(text))
}

func (i *InstanceDirectoryMode) Set(in string) error {
	switch strings.ToLower(in) {
	case "off":
		*i = InstanceDirectoryModeOff
		return nil
	case "webonly", "":
		*i = InstanceDirectoryModeWebOnly
		return nil
	case "open":
		*i = InstanceDirectoryModeOpen
		return nil
	default:
		return errors.New("unrecognized instance directory mode '" + in + "'")
	}
}

func (i InstanceDirectoryMode) String() string {
	switch i {
	case InstanceDirectoryModeOff:
		return "off"
	case InstanceDirectoryModeWebOnly:
		return "webonly"
	case InstanceDirectoryModeOpen:
		return "open"
	default:
		return "unknown"
	}
}
