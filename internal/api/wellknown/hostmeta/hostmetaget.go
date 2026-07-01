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

package hostmeta

import (
	"net/http"

	"code.superseriousbusiness.org/gopkg/httputil"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
)

// HostMetaGETHandler swagger:operation GET /.well-known/host-meta hostMetaGet
//
// Returns a compliant hostmeta response to web host metadata queries.
//
// See: https://www.rfc-editor.org/rfc/rfc6415.html
//
//	---
//	tags:
//	- .well-known
//
//	produces:
//	- application/xrd+xml"
//
//	responses:
//		'200':
//			schema:
//				"$ref": "#/definitions/hostmeta"
func (m *Module) HostMetaGETHandler(c *httputil.Context) {
	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.HostMetaHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	hostMeta := m.processor.Fedi().HostMetaGet()

	// Encode XML HTTP response.
	httputil.EncodeXMLResponse(c,
		http.StatusOK,
		HostMetaContentType,
		hostMeta,
	)
}
