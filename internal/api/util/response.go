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

package util

import (
	"encoding/json"
	"net/http"
)

var (
	// Pre-preared response body data.
	StatusOKJSON = mustJSON(map[string]string{
		"status": http.StatusText(http.StatusOK),
	})
	StatusAcceptedJSON = mustJSON(map[string]string{
		"status": http.StatusText(http.StatusAccepted),
	})
	StatusForbiddenJSON = mustJSON(map[string]string{
		"status": http.StatusText(http.StatusForbidden),
	})
	StatusInternalServerErrorJSON = mustJSON(map[string]string{
		"status": http.StatusText(http.StatusInternalServerError),
	})
	ErrorCapacityExceeded = mustJSON(map[string]string{
		"error": "server capacity exceeded",
	})
	ErrorRateLimited = mustJSON(map[string]string{
		"error": "rate limit reached",
	})
	EmptyJSONObject = json.RawMessage(`{}`)
	EmptyJSONArray  = json.RawMessage(`[]`)
)

// mustJSON converts data to JSON, else panicking.
func mustJSON(data any) []byte {
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return b
}
