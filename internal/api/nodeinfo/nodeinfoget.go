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

package nodeinfo

import (
	"errors"
	"net/http"

	"code.superseriousbusiness.org/gopkg/httputil"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
)

// NodeInfo2GETHandler swagger:operation GET /nodeinfo/{schema_version} nodeInfoGet
//
// Returns a compliant nodeinfo response to node info queries.
//
// See: https://nodeinfo.diaspora.software/schema.html
//
//	---
//	tags:
//	- nodeinfo
//
//	parameters:
//	-
//		name: schema_version
//		type: string
//		description: Schema version of nodeinfo to request. 2.0 and 2.1 are currently supported.
//		in: path
//		required: true
//
//	produces:
//	- application/json; profile="http://nodeinfo.diaspora.software/ns/schema/2.0#"
//	- application/json; profile="http://nodeinfo.diaspora.software/ns/schema/2.1#"
//
//	responses:
//		'200':
//			schema:
//				"$ref": "#/definitions/nodeinfo"
func (m *Module) NodeInfo2GETHandler(c *httputil.Context) {
	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	var (
		contentType   string
		schemaVersion = c.PathValue(NodeInfoSchema)
	)

	switch schemaVersion {
	case NodeInfo20:
		contentType = NodeInfo20ContentType
	case NodeInfo21:
		contentType = NodeInfo21ContentType
	default:
		const errText = "only nodeinfo 2.0 and 2.1 are supported"
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorNotFound(errors.New(errText), errText))
		return
	}

	nodeInfo, errWithCode := m.processor.Fedi().NodeInfoGet(c, schemaVersion)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	// Encode JSON response.
	httputil.JSONType(c,
		http.StatusOK,
		contentType,
		nodeInfo,
	)
}
