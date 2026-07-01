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

package streaming

import (
	"net/http"
	"time"

	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"code.superseriousbusiness.org/gotosocial/internal/templates"
	"github.com/gorilla/websocket"
)

const (
	BasePath            = "/v1/streaming"          // path for the streaming api, minus the 'api' prefix
	StreamQueryKey      = "stream"                 // type of stream being requested
	StreamListKey       = "list"                   // id of list being requested
	StreamTagKey        = "tag"                    // name of tag being requested
	AccessTokenQueryKey = "access_token"           // oauth access token
	AccessTokenHeader   = "Sec-Websocket-Protocol" //nolint:gosec
)

type Module struct {
	templates *templates.Templates
	processor *processing.Processor
	wsupgrade websocket.Upgrader
	pingfreq  time.Duration
}

func New(processor *processing.Processor, templates *templates.Templates, pingFreq time.Duration) *Module {
	return &Module{
		templates: templates,
		processor: processor,
		pingfreq:  pingFreq,
		wsupgrade: websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,

			// We expect CORS requests for websockets,
			// (via eg., semaphore.social) so be lenient.
			// TODO: make this customizable?
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (m *Module) Route(g *httputil.RouteGroup) {
	g.GET(BasePath, m.StreamGETHandler)
}
