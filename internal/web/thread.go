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

package web

import (
	"net/http"

	"code.superseriousbusiness.org/gopkg/httputil"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/templates"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
)

func (m *Module) threadGETHandler(c *httputil.Context) {
	// Parse account requestedUser and status ID from the URL.
	requestedUser, errWithCode := apiutil.ParseUsername(c.PathValue(apiutil.UsernameKey))
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, m.templates, errWithCode)
		return
	}

	statusID, errWithCode := apiutil.ParseID(c.PathValue(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, m.templates, errWithCode)
		return
	}

	// Check what type of content is being requested. If we're getting an AP
	// request on this endpoint we should render the AP representation instead.
	accept, err := apiutil.NegotiateAccept(c, apiutil.HTMLOrActivityPubHeaders...)
	if err != nil {
		apiutil.WebErrorHandler(c, m.templates, gtserror.NewErrorNotAcceptable(err, err.Error()))
		return
	}

	if apiutil.ASContentType(accept) {
		// AP status representation has been requested.
		status, errWithCode := m.processor.Fedi().StatusGet(c, requestedUser, statusID)
		if errWithCode != nil {
			apiutil.WebErrorHandler(c, m.templates, errWithCode)
			return
		}

		httputil.JSONType(c, http.StatusOK, accept, status)
		return
	}

	// text/html has been requested. Proceed with getting the web view of the status.

	// Fetch the target account so we can do some checks on it.
	acct, errWithCode := m.processor.Account().GetWeb(c, requestedUser)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, m.templates, errWithCode)
		return
	}

	// If requested account is suspended, this page should not be visible.
	if acct.Suspended {
		err := gtserror.Newf("account %s is suspended", requestedUser)
		apiutil.WebErrorHandler(c, m.templates, gtserror.NewErrorNotFound(err))
		return
	}

	// Get the thread context. This will fetch the status as well.
	context, errWithCode := m.processor.Status().WebContextGet(c, statusID)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, m.templates, errWithCode)
		return
	}

	// Ensure status actually belongs to requested account.
	if context.Status.Account.ID != acct.ID {
		err := gtserror.Newf("account %s does not own status %s", requestedUser, statusID)
		apiutil.WebErrorHandler(c, m.templates, gtserror.NewErrorNotFound(err))
		return
	}

	// If every account in the thread is indexable, then we allow the thread page to be indexed.
	var robotsMeta string
	if context.Indexable {
		robotsMeta = apiutil.RobotsDirectivesAllowSome
	}

	// Prepare stylesheets for thread.
	stylesheets := make([]string, 0, 6)

	// Basic thread stylesheets.
	stylesheets = append(
		stylesheets,
		[]string{
			cssFA,
			cssStatus,
			cssThread,
		}...,
	)

	// User-selected theme if set.
	if theme := acct.Theme; theme != "" {
		stylesheets = append(
			stylesheets,
			themesPathPrefix+"/"+theme,
		)
	}

	// Custom CSS for this user last in cascade.
	stylesheets = append(
		stylesheets,
		"/@"+acct.Username+"/custom.css",
	)

	// Fetch instance details for status open graph details.
	instance, errWithCode := m.processor.InstanceGetV1(c)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, m.templates, errWithCode)
		return
	}

	// Pass to template renderer.
	m.templates.RenderPage(c,
		http.StatusOK,
		templates.WebPage{
			Template:    "thread.tmpl",
			Stylesheets: stylesheets,
			Javascript: []templates.JavascriptEntry{
				{
					Src:   jsFrontend,
					Async: true,
					Defer: true,
				},
				{
					Bottom: true,
					Src:    jsFrontendPrerender,
				},
			},
			Extra: map[string]any{
				"context":    context,
				"robotsMeta": robotsMeta,
				"instance":   instance,
				"ogMeta":     typeutils.OpenGraphStatus(instance, acct, context.Status),
			},
		},
	)
}
