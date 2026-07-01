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
	"time"

	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gotosocial/internal/api/health"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/router"
	"code.superseriousbusiness.org/gotosocial/internal/templates"
	"codeberg.org/gruf/go-cache/v3"
)

type MaintenanceModule struct {
	templates *templates.Templates
	eTagCache cache.Cache[string, eTagCacheEntry]
}

// NewMaintenance returns a module that routes only
// static assets, and returns a code 503 maintenance
// message template to all other requests.
func NewMaintenance() *MaintenanceModule {
	templates, err := templates.Load(nil, nil, nil)
	if err != nil {
		panic(err)
	}
	return &MaintenanceModule{
		templates: templates,
		eTagCache: newETagCache(),
	}
}

// ETagCache implements withETagCache.
func (m *MaintenanceModule) ETagCache() cache.Cache[string, eTagCacheEntry] {
	return m.eTagCache
}

func (m *MaintenanceModule) Route(r *router.Router, mi ...httputil.Middleware) {

	// Add static assets to router.
	routeAssets(m, r, mi...)

	// Serve OK in response to live
	// requests, but not ready requests.
	r.GET(health.LivePath, func(c *httputil.Context) { c.W.WriteHeader(http.StatusOK) })
	r.HEAD(health.LivePath, func(c *httputil.Context) { c.W.WriteHeader(http.StatusOK) })

	// For everything else, serve maintenance template.
	r.Handle("", "/", func(c *httputil.Context) {
		retryAfter := time.Now().Add(120 * time.Second).UTC()
		c.W.Header().Add("Retry-After", "120")
		c.W.Header().Add("Retry-After", retryAfter.Format(http.TimeFormat))
		c.W.Header().Set("Cache-Control", "no-store")
		m.templates.RenderPage(c, http.StatusServiceUnavailable, templates.WebPage{
			Template: "maintenance.tmpl",
			Extra: map[string]any{
				"host": config.GetHost(),

				// empty instance model, to prevent
				// templater from fetching a template,
				// which fail with state of database.
				"instance": &apimodel.InstanceV1{},
			},
		})
	})
}
