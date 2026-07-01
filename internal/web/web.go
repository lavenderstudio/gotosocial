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
	"context"
	"net/http"
	"net/url"

	"code.superseriousbusiness.org/gopkg/httputil"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/middleware"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"code.superseriousbusiness.org/gotosocial/internal/router"
	"code.superseriousbusiness.org/gotosocial/internal/templates"
	"code.superseriousbusiness.org/gotosocial/internal/uris"
	"codeberg.org/gruf/go-cache/v3"
)

const (
	confirmEmailPath         = "/" + uris.ConfirmEmailPath
	profileGroupPath         = "/@:username"
	statusPath               = "/statuses/:" + apiutil.IDKey // leave out the '/@:username' prefix as this will be served within the profile group
	tagsPath                 = "/tags/:" + apiutil.TagNameKey
	customCSSPath            = profileGroupPath + "/custom.css"
	instanceCustomCSSPath    = "/custom.css"
	rssFeedPath              = profileGroupPath + "/feed.rss"
	assetsPathPrefix         = "/assets"
	distPathPrefix           = assetsPathPrefix + "/dist"
	themesPathPrefix         = assetsPathPrefix + "/themes"
	settingsPathPrefix       = "/settings"
	settingsPanelGlob        = settingsPathPrefix + "/*panel"
	userPanelPath            = settingsPathPrefix + "/user"
	adminPanelPath           = settingsPathPrefix + "/admin"
	signupPath               = "/signup"
	authorizeInteractionPath = "/authorize_interaction"

	cacheControlHeader    = "Cache-Control"     // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control
	cacheControlNoCache   = "no-cache"          // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control#response_directives
	ifModifiedSinceHeader = "If-Modified-Since" // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-Modified-Since
	ifNoneMatchHeader     = "If-None-Match"     // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-None-Match
	eTagHeader            = "ETag"              // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/ETag
	lastModifiedHeader    = "Last-Modified"     // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Last-Modified

	cssFA               = assetsPathPrefix + "/Fork-Awesome/css/fork-awesome.min.css"
	cssAbout            = distPathPrefix + "/about.css"
	cssIndex            = distPathPrefix + "/index.css"
	cssLoginInfo        = distPathPrefix + "/login-info.css"
	cssStatus           = distPathPrefix + "/status.css"
	cssThread           = distPathPrefix + "/thread.css"
	cssProfile          = distPathPrefix + "/profile.css"
	cssProfileWideStats = distPathPrefix + "/_profile-header-wide-stats.css"
	cssProfileGallery   = distPathPrefix + "/profile-gallery.css"
	cssSettings         = distPathPrefix + "/settings-style.css"
	cssTag              = distPathPrefix + "/tag.css"
	cssDirectory        = distPathPrefix + "/directory.css"

	jsFrontend          = distPathPrefix + "/frontend.js"           // Progressive enhancement frontend JS.
	jsFrontendPrerender = distPathPrefix + "/frontend_prerender.js" // Frontend JS that should run before page renders.
	jsSettings          = distPathPrefix + "/settings.js"           // Settings panel React application.
)

type Module struct {
	templates    *templates.Templates
	processor    *processing.Processor
	eTagCache    cache.Cache[string, eTagCacheEntry]
	cookiePolicy apiutil.CookiePolicy
	isURIBlocked func(context.Context, *url.URL) (bool, error)
}

func New(db db.DB, processor *processing.Processor, templates *templates.Templates, cookiePolicy apiutil.CookiePolicy) *Module {
	return &Module{
		templates:    templates,
		processor:    processor,
		eTagCache:    newETagCache(),
		cookiePolicy: cookiePolicy,
		isURIBlocked: db.IsURIBlocked,
	}
}

// ETagCache implements withETagCache.
func (m *Module) ETagCache() cache.Cache[string, eTagCacheEntry] {
	return m.eTagCache
}

// Route attaches the assets filesystem and profile,
// status, and other web handlers to the router.
func (m *Module) Route(r *router.Router, mi ...httputil.Middleware) {

	// Add static assets to router.
	routeAssets(m, r, mi...)

	// Handlers that serve profiles and statuses should use
	// the SignatureCheck middleware, so that requests with
	// content-type application/activity+json can be served.
	profileGroup := r.Group(profileGroupPath)
	profileGroup.Use(mi...)
	profileGroup.Use(middleware.ExtractSignature(m.isURIBlocked), middleware.CacheControl(middleware.CacheControlConfig{
		Directives: []string{"no-store"},
	}))
	profileGroup.GET("", m.profileGETHandler) // use empty path here since it's the base of the group
	profileGroup.GET(statusPath, m.threadGETHandler)

	// Group for all other web handlers.
	everythingElseGroup := r.Group("")
	everythingElseGroup.Use(mi...)
	everythingElseGroup.GET("/", m.indexHandler) // front-page
	everythingElseGroup.GET(settingsPathPrefix, m.SettingsPanelHandler)
	everythingElseGroup.GET(settingsPanelGlob, m.SettingsPanelHandler)
	everythingElseGroup.GET(customCSSPath, m.customCSSGETHandler)
	everythingElseGroup.GET(instanceCustomCSSPath, m.instanceCustomCSSGETHandler)
	everythingElseGroup.GET(rssFeedPath, m.rssFeedGETHandler)
	everythingElseGroup.GET(confirmEmailPath, m.confirmEmailGETHandler)
	everythingElseGroup.POST(confirmEmailPath, m.confirmEmailPOSTHandler)
	everythingElseGroup.GET(aboutPath, m.aboutGETHandler)
	everythingElseGroup.GET(loginPath, m.loginGETHandler)
	everythingElseGroup.GET(domainBlocklistPath, m.domainBlocklistGETHandler)
	everythingElseGroup.GET(domainAllowlistPath, m.domainAllowlistGETHandler)
	everythingElseGroup.GET(tagsPath, m.tagGETHandler)
	everythingElseGroup.GET(signupPath, m.signupGETHandler)
	everythingElseGroup.GET(authorizeInteractionPath, m.authorizeInteractionGETHandler)
	everythingElseGroup.POST(signupPath, m.signupPOSTHandler)
	everythingElseGroup.GET(directoryPath, m.directoryGETHandler)

	// Redirects from old endpoints for back compat.
	r.GET("/auth/edit", func(c *httputil.Context) { httputil.Redirect(c, http.StatusMovedPermanently, userPanelPath) })
	r.GET("/user", func(c *httputil.Context) { httputil.Redirect(c, http.StatusMovedPermanently, userPanelPath) })
	r.GET("/admin", func(c *httputil.Context) { httputil.Redirect(c, http.StatusMovedPermanently, adminPanelPath) })
}
