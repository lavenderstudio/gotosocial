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

package users

import (
	"net/http"

	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gopkg/log"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"

	errorsv2 "codeberg.org/gruf/go-errors/v2"
)

// InboxPOSTHandler deals with incoming POST requests to an actor's inbox.
// Eg., POST to https://example.org/users/whatever/inbox.
func (m *Module) InboxPOSTHandler(c *httputil.Context) {
	_, err := m.processor.Fedi().InboxPost(c, &c.W, c.R)
	if err != nil {
		errWithCode := errorsv2.AsV2[gtserror.WithCode](err)

		if errWithCode == nil {
			// Something else went wrong, and someone forgot to return
			// an errWithCode! It's chill though. Log the error but don't
			// return it as-is to the caller, to avoid leaking internals.
			log.Errorf(c, "returning Bad Request to caller, err was: %q", err)
			errWithCode = gtserror.NewErrorBadRequest(err)
		}

		// Pass along confirmed error with code to the main error handler
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	httputil.Data(c, http.StatusAccepted, apiutil.AppJSON, apiutil.StatusAcceptedJSON)
}
