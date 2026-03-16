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

package bundb

import (
	"context"
	"time"

	"code.superseriousbusiness.org/gopkg/log"
	"codeberg.org/gruf/go-kv/v2"
	"github.com/uptrace/bun"
)

type queryLogger struct{}

func (queryLogger) BeforeQuery(ctx context.Context, _ *bun.QueryEvent) context.Context {
	return ctx // noop
}

func (h queryLogger) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	if dur := time.Since(event.StartTime); dur > time.Second {
		// Log at WARN level with slow query msg.
		log.WarnKVs(ctx, kv.Fields{
			{"duration", dur},
			{"query", event.Query},
			{"msg", "SLOW DATABASE QUERY"},
		}...)
	} else {
		// Log query at a 'faux'
		// trace level without
		// context, for speed 😎.
		log.PrintKVs(kv.Fields{
			{"level", "TRACE"},
			{"duration", dur},
			{"query", event.Query},
		}...)
	}
}

type slowQueryLogger struct{}

func (slowQueryLogger) BeforeQuery(ctx context.Context, _ *bun.QueryEvent) context.Context {
	return ctx // do nothing
}

func (slowQueryLogger) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	if dur := time.Since(event.StartTime); dur > time.Minute {
		// Only WARN log slow queries.
		log.WarnKVs(ctx, kv.Fields{
			{"duration", dur},
			{"query", event.Query},
			{"msg", "SLOW DATABASE QUERY"},
		}...)
	}
}
