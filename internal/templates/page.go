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

package templates

import (
	"net/netip"

	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gopkg/log"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
)

// WebPage encapsulates variables for
// rendering an HTML template within
// a standard GtS "page" template.
type WebPage struct {

	// Name of the template for rendering
	// the page. Eg., "example.tmpl".
	Template string

	// Paths to CSS files to add to
	// the page as "stylesheet" entries.
	// Can be nil.
	Stylesheets []string

	// JS files to add to the
	// page as "script" entries.
	// Can be nil.
	Javascript []JavascriptEntry

	// Extra parameters to pass to
	// the template for rendering,
	// eg., "account": *Account etc.
	// Can be nil.
	Extra map[string]any
}

type JavascriptEntry struct {
	// Insert <script> tag at the end
	// of <body> rather than in <head>.
	Bottom bool

	// Path to the js file.
	Src string

	// Use async="" attribute.
	Async bool

	// Use defer="" attribute.
	Defer bool
}

// RenderPage renders the given HTML template and
// page params within the standard GtS "page" template.
//
// ogMeta, stylesheets, javascript, and any extra
// properties will be provided to the template if
// set, but can all be nil.
//
// RenderPage also checks whether the requesting
// clientIP is 127.0.0.1 or within a private IP range.
// If so, it injects a suggestion into the page header
// about setting trusted-proxies correctly.
func (t *Templates) RenderPage(c *httputil.Context, statusCode int, page WebPage) {

	// Use their map (if given)
	// as template data storage.
	data := page.Extra
	if data == nil {
		data = make(map[string]any, 6)
	}

	// If rate limiting is disabled entirely
	// there's no point in giving a trusted
	// proxies rec, as proper clientIP is
	// basically only used for rate limiting.
	if config.GetAdvancedRateLimitRequests() > 0 {
		t.injectTrustedProxiesRec(c, data)
	}

	// Add per-page template variables.
	data["stylesheets"] = page.Stylesheets
	data["javascript"] = page.Javascript

	// Pass to main template page rendering function.
	t.renderPage(c, statusCode, page.Template, data)
}

// RenderErrorPage renders a generic error page template with the given message.
func (t *Templates) RenderErrorPage(c *httputil.Context, statusCode int, msg string) {
	data := make(map[string]any, 4)
	data["code"] = statusCode
	data["error"] = msg
	data["requestID"] = gtscontext.RequestID(c)
	t.renderPage(c, statusCode, "error.tmpl", data)
}

// RenderNotVisiblePage renders a page
// explaining that the item at the requested
// URL is not visible on the web.
func (t *Templates) RenderNotVisiblePage(
	c *httputil.Context,
	statusCode int,
) {
	data := make(map[string]any, 6)
	data["url"] = config.GetProtocol() + "://" + config.GetHost() + c.R.URL.String()
	data["javascript"] = []JavascriptEntry{
		{
			// include frontend JS so url can
			// be copy / pasted client-side.
			Src:   "/assets/dist/frontend.js",
			Async: true,
			Defer: true,
		},
	}
	data["requestID"] = gtscontext.RequestID(c)
	t.renderPage(c, 404, "item_not_visible.tmpl", data)
}

// RenderDeletedPage renders a page
// explaining that the item at the
// requested URL has been deleted.
func (t *Templates) RenderDeletedPage(
	c *httputil.Context,
	statusCode int,
) {
	data := make(map[string]any, 4)
	data["requestID"] = gtscontext.RequestID(c)
	t.renderPage(c, 404, "item_deleted.tmpl", data)
}

func (t *Templates) renderPage(
	c *httputil.Context,
	statusCode int,
	template string,
	data map[string]any,
) {
	// The database provided can
	// be nil during maintenance.
	if t.db != nil {

		// Check if instance model has been provided in data.
		instance, ok := data["instance"].(*apimodel.InstanceV1)
		if !ok {

			// Get our current instance settings model.
			settings, err := t.db.GetInstanceSettings(c)
			if err != nil {
				log.Errorf(c, "error getting settings: %v", err)
			} else {

				// Convert the fetched instance settings to an API v1 instance model.
				instance, err = t.typeConv.InstanceSettingsToAPIV1Instance(c, settings)
				if err != nil {
					log.Errorf(c, "error converting to instance: %v", err)
				}
			}

			// Set fetched instance data.
			data["instance"] = instance
		}
	}

	// Render template inside page.
	data["pageContent"] = template

	// Inject specific page class by trimming ".tmpl" suffix.
	// In the page template (see page.tmpl) this will be appended
	// with "-page", so "index.tmpl" gets class "page index-page".
	data["pageClass"] = template[:len(template)-5]

	// So the template knows if the account directory is enabled
	data["showAcctDir"] = config.GetInstanceDirectoryWebEnabled()

	// Render the page.
	httputil.RenderHTML(c,
		statusCode,
		t.baseTmpl,
		"page.tmpl",
		data,
	)
}

// DockerSubnet is a prefix that lets one make hazy guesses
// as to whether an address is within the ranges Docker
// uses for subnets, ie., 172.16.0.0 -> 172.31.255.255.
var dockerSubnet = netip.MustParsePrefix("172.16.0.0/12")

// ipv4 / ipv6 loopback / localhost addresses.
var ipv4Loopback = netip.MustParseAddr("127.0.0.1")
var ipv6Loopback1 = netip.MustParseAddr("::1")
var ipv6Loopback2 = netip.MustParseAddr("0:0:0:0:0:0:0:1")

func (t *Templates) injectTrustedProxiesRec(c *httputil.Context, data map[string]any) {
	const (
		ipv4CIDR       = "/32"
		ipv6CIDR       = "/128"
		dockerIPv4CIDR = "/16"
	)

	// clientIP = the client IP that gin
	// derives based on x-forwarded-for
	// and current trusted proxies.
	clientIP := gtscontext.ClientIP(c)
	if clientIP == nil {
		log.Warn(c, "empty client ip")
		return
	}

	switch {
	// Ensure client IP set.
	case clientIP == nil:
		log.Warn(c, "empty client ip")
		return

	// Check if set to loopback / localhost.
	case (*clientIP) == ipv6Loopback1 ||
		(*clientIP) == ipv6Loopback2:
		// Suggest precise ipv6 loopback.
		trustedProxiesRec := clientIP.String() + ipv6CIDR
		data["trustedProxiesRec"] = trustedProxiesRec
		return
	case (*clientIP) == ipv4Loopback:
		// Suggest precise ipv4 loopback.
		trustedProxiesRec := clientIP.String() + ipv4CIDR
		data["trustedProxiesRec"] = trustedProxiesRec
		return
	}

	var hasRemoteIPHeader bool

	// Check for a reverse-proxy-set IP in headers.
	for _, key := range t.proxyCfg.RemoteIPHeaders {
		if c.R.Header.Get(key) != "" {
			hasRemoteIPHeader = true
			break
		}
	}

	if !hasRemoteIPHeader {
		// Upstream hasn't set a
		// remote IP header so we're
		// probably not in a reverse
		// proxy setup, bail.
		return
	}

	if !clientIP.IsPrivate() {
		// Upstream set a remote IP
		// header but final clientIP
		// isn't private, so upstream
		// is probably already trusted.
		// Don't inject suggestion.
		return
	}

	for _, prefix := range config.GetAdvancedRateLimitExceptions() {
		if prefix.Contains(*clientIP) {
			// This ip is exempt from
			// rate limiting, so don't
			// inject the suggestion.
			return
		}
	}

	// Private IP, check if Docker subnet.
	if dockerSubnet.Contains(*clientIP) {

		// Suggest a CIDR that likely
		// covers this Docker subnet,
		// eg., 172.17.0.0 -> 172.17.255.255.
		trustedProxiesRec := clientIP.String() + dockerIPv4CIDR
		data["trustedProxiesRec"] = trustedProxiesRec
		return
	}

	// Private IP but we don't know
	// what it is. Suggest precise CIDR.
	var cidr string
	if clientIP.Is6() {
		cidr = ipv6CIDR
	} else {
		cidr = ipv4CIDR
	}

	trustedProxiesRec := clientIP.String() + cidr
	data["trustedProxiesRec"] = trustedProxiesRec
}
