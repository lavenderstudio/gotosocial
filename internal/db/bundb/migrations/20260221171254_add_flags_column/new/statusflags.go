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

package gtsmodel

import (
	"fmt"
	"strconv"

	"codeberg.org/gruf/go-byteutil"
)

// StatusFlag is the bit type for
// indivial StatusFlags members.
type StatusFlag bitFieldType

const (
	// NOTE: THE FOLLOWING VALUES SHOULD NEVER
	// BE CHANGED WITHOUT PERFORMING A DATABASE
	// MIGRATION TO UPDATE OLD -> NEW BIT VALUES.

	// StatusFlagDeleted indices whether status is marked as deleted,
	// (and thus should be stubbed-out and replaced with placeholders).
	//
	// TODO: NOT YET IMPLEMENTED
	StatusFlagDeleted StatusFlag = 1 << 1

	// StatusFlagSensitive indicates whether status is marked as sensitive.
	StatusFlagSensitive StatusFlag = 1 << 2

	// StatusFlagLocal indicates whether status was authored by local account.
	StatusFlagLocal StatusFlag = 1 << 3

	// StatusFlagFederated indicates whether status is federated beyond local timeline(s).
	StatusFlagFederated StatusFlag = 1 << 4

	// StatusFlagPendingApproval indicates whether status is a reply / boost wrapper
	// that must be approved by the reply-ee or boost-ee before being fully distributed.
	StatusFlagPendingApproval StatusFlag = 1 << 5
)

// String returns a human-readable form of StatusFlag.
func (f StatusFlag) String() string {
	switch f {
	case 0:
		return "unset"
	case StatusFlagDeleted:
		return "deleted"
	case StatusFlagSensitive:
		return "sensitive"
	case StatusFlagLocal:
		return "local"
	case StatusFlagFederated:
		return "federated"
	case StatusFlagPendingApproval:
		return "pending_approval"
	default:
		panic(fmt.Sprintf("invalid status flag: %d", f))
	}
}

// StatusFlags stores a variety of different boolean
// flags for attached status model, stored in smallint
// bit field type for memory and database efficiency.
type StatusFlags bitFieldType

// Deleted returns whether StatusFlagDeleted is set.
func (f StatusFlags) Deleted() bool {
	return f&StatusFlags(StatusFlagDeleted) != 0
}

// SetDeleted sets / unsets the StatusFlagDeleted bit.
func (f *StatusFlags) SetDeleted(ok bool) {
	if ok {
		*f |= StatusFlags(StatusFlagDeleted)
	} else {
		*f &= ^StatusFlags(StatusFlagDeleted)
	}
}

// Sensitive returns whether StatusFlagSensitive is set.
func (f StatusFlags) Sensitive() bool {
	return f&StatusFlags(StatusFlagSensitive) != 0
}

// SetSensitive sets / unsets the StatusFlagSensitive bit.
func (f *StatusFlags) SetSensitive(ok bool) {
	if ok {
		*f |= StatusFlags(StatusFlagSensitive)
	} else {
		*f &= ^StatusFlags(StatusFlagSensitive)
	}
}

// Local returns whether StatusFlagLocal is set.
func (f StatusFlags) Local() bool {
	return f&StatusFlags(StatusFlagLocal) != 0
}

// SetLocal sets / unsets the StatusFlagLocal bit.
func (f *StatusFlags) SetLocal(ok bool) {
	if ok {
		*f |= StatusFlags(StatusFlagLocal)
	} else {
		*f &= ^StatusFlags(StatusFlagLocal)
	}
}

// Federated returns whether StatusFlagFederated is set.
func (f StatusFlags) Federated() bool {
	return f&StatusFlags(StatusFlagFederated) != 0
}

// SetFederated sets / unsets the StatusFlagFederated bit.
func (f *StatusFlags) SetFederated(ok bool) {
	if ok {
		*f |= StatusFlags(StatusFlagFederated)
	} else {
		*f &= ^StatusFlags(StatusFlagFederated)
	}
}

// PendingApproval returns whether StatusFlagPendingApproval is set.
func (f StatusFlags) PendingApproval() bool {
	return f&StatusFlags(StatusFlagPendingApproval) != 0
}

// SetPendingApproval sets / unsets the StatusFlagPendingApproval bit.
func (f *StatusFlags) SetPendingApproval(ok bool) {
	if ok {
		*f |= StatusFlags(StatusFlagPendingApproval)
	} else {
		*f &= ^StatusFlags(StatusFlagPendingApproval)
	}
}

// String returns a single human-readable form of StatusFlags.
func (f StatusFlags) String() string {
	var buf byteutil.Buffer
	buf.B = append(buf.B, '{')
	buf.B = append(buf.B, "deleted="...)
	buf.B = strconv.AppendBool(buf.B, f.Deleted())
	buf.B = append(buf.B, ',')
	buf.B = append(buf.B, "sensitive="...)
	buf.B = strconv.AppendBool(buf.B, f.Sensitive())
	buf.B = append(buf.B, ',')
	buf.B = append(buf.B, "federated="...)
	buf.B = strconv.AppendBool(buf.B, f.Federated())
	buf.B = append(buf.B, ',')
	buf.B = append(buf.B, "pending_approval="...)
	buf.B = strconv.AppendBool(buf.B, f.PendingApproval())
	buf.B = append(buf.B, '}')
	return buf.String()
}
