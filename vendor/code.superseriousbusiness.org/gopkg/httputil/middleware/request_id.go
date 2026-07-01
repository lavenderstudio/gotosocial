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

package middleware

import (
	"context"
	"encoding/base32"
	"encoding/binary"
	"sync/atomic"
	"time"

	"code.superseriousbusiness.org/gopkg/httputil"
)

type requestidkey struct{}

// GetRequestID searches given context for request ID string stored by WithRequestID().
func GetRequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestidkey{}).(string)
	return id
}

// WithRequestID returns a middleware that checks incoming request for
// an existing request ID with given header name, generating new where
// not set, or in the case of an empty name always generates a new ID.
//
// The ID is deterministic based on server start time and current request count
// count, so should not be used other than for log entry identification.
func WithRequestID(header string) httputil.FlatMiddlewareFunc {
	var (
		// request counter.
		count atomic.Uint32

		// server start time in milliseconds.
		start = uint64(time.Now().UnixMilli())

		// shorthand to binary.
		be = binary.BigEndian

		// b32 is a base 32 encoding based on a human-readable character set (no padding).
		b32 = base32.NewEncoding("0123456789abcdefghjkmnpqrstvwxyz").WithPadding(-1)
	)

	if header == "" {
		return func(c *httputil.Context) {
			var buf [12]byte
			be.PutUint64(buf[0:], start)
			be.PutUint32(buf[8:], count.Add(1))
			id := b32.EncodeToString(buf[:])
			c.V.Set(requestidkey{}, id)
		}
	}

	return func(c *httputil.Context) {
		var id string
		if id = c.R.Header.Get(header); id == "" {
			var buf [12]byte
			be.PutUint64(buf[0:], start)
			be.PutUint32(buf[8:], count.Add(1))
			id = b32.EncodeToString(buf[:])
		}
		c.V.Set(requestidkey{}, id)
		c.W.Header().Set(header, id)
	}
}
