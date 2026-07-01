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

package gtscontext

import (
	"context"

	"code.superseriousbusiness.org/gopkg/log"
	"codeberg.org/gruf/go-kv/v2"
)

func init() {
	// Add our required logging hooks on application initialization.
	log.AddHook(func(ctx context.Context, kvs []kv.Field) []kv.Field {

		// Incoming HTTP request ID, if set.
		if id := RequestID(ctx); id != "" {
			kvs = append(kvs, kv.Field{K: "requestID", V: id})
		}

		// Incoming client's signing pubkey ID (to us).
		if id := HTTPSignaturePubKeyID(ctx); id != nil {
			kvs = append(kvs, kv.Field{K: "pubKeyID", V: id})
		}

		// Outgoing client signing pubkey ID (from us).
		if id := OutgoingPublicKeyID(ctx); id != "" {
			kvs = append(kvs, kv.Field{K: "pubKeyID", V: id})
		}

		return kvs
	})
}
