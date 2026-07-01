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
	"code.superseriousbusiness.org/gopkg/log"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"github.com/gorilla/feeds"
)

const (
	charsetUTF8 = "; charset=utf-8"
	appRSSUTF8  = string(apiutil.AppRSSXML) + charsetUTF8
	appAtomUTF8 = string(apiutil.AppAtomXML) + charsetUTF8
	appJSONUTF8 = string(apiutil.AppFeedJSON) + charsetUTF8
)

func (m *Module) rssFeedGETHandler(c *httputil.Context) {
	contentType, err := httputil.NegotiateAccept(c,
		apiutil.AppRSSXML,
		apiutil.AppAtomXML,
		apiutil.AppFeedJSON,
		apiutil.AppJSON,
	)
	if err != nil {
		apiutil.WebErrorHandler(c, m.templates, gtserror.NewErrorNotAcceptable(err, err.Error()))
		return
	}

	// Fetch + normalize username from URL.
	username, errWithCode := apiutil.ParseUsername(c.PathValue(apiutil.UsernameKey))
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, m.templates, errWithCode)
		return
	}

	// Parse paging parameters from request.
	page, errWithCode := paging.ParseIDPage(c,
		1,  // min limit
		40, // max limit
		20, // default limit
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	getFunc, lastPostAt, errWithCode := m.processor.Account().GetRSSFeedForUsername(
		c,
		username,
		page,
	)
	if errWithCode != nil {
		apiutil.WebErrorHandler(c, m.templates, errWithCode)
		return
	}

	var feed *feeds.Feed

	// Key to use in etag cache (note content-type suffix).
	cacheKey := c.R.URL.Path + "#" + contentType

	// Check etag cache for an existing entry under key.
	cacheEntry, wasCached := m.eTagCache.Get(cacheKey)

	if !wasCached || unixAfter(lastPostAt, cacheEntry.lastModified) {
		// We either have no ETag cache entry for this account's feed,
		// or we have an expired cache entry (account has posted since
		// the cache entry was last generated).
		//
		// As such, we need to generate a new ETag, and for that we need
		// the string representation of the RSS feed.
		feed, errWithCode = getFunc()
		if errWithCode != nil {
			apiutil.WebErrorHandler(c, m.templates, errWithCode)
			return
		}

		etag, err := generateFeedETag(feed, contentType)
		if err != nil {
			errWithCode := gtserror.NewErrorInternalError(err)
			apiutil.WebErrorHandler(c, m.templates, errWithCode)
			return
		}

		// We never want lastModified to be zero, so if account
		// has never actually posted anything, just use Now as
		// the lastModified time instead for cache control.
		var lastModified time.Time
		if lastPostAt.IsZero() {
			lastModified = time.Now()
		} else {
			lastModified = lastPostAt
		}

		// Store the new cache entry.
		cacheEntry = eTagCacheEntry{
			eTag:         etag,
			lastModified: lastModified,
		}
		m.eTagCache.Set(cacheKey, cacheEntry)
	}

	// Set 'ETag' and 'Last-Modified' headers no matter what;
	// even if we return 304 in the next checks, caller may
	// want to cache these header values.
	c.W.Header().Set(eTagHeader, cacheEntry.eTag)
	c.W.Header().Set(lastModifiedHeader, cacheEntry.lastModified.Format(http.TimeFormat))

	// Instruct caller to validate the response with us before
	// each reuse, so that the 'ETag' and 'Last-Modified' headers
	// actually take effect.
	//
	// "The no-cache response directive indicates that the response
	// can be stored in caches, but the response must be validated
	// with the origin server before each reuse, even when the cache
	// is disconnected from the origin server."
	//
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control
	c.W.Header().Set(cacheControlHeader, cacheControlNoCache)

	// Check if caller submitted an ETag via 'If-None-Match'.
	// If they did + it matches what we have, that means they've
	// already seen the latest version of this feed, so just bail.
	ifNoneMatch := c.R.Header.Get(ifNoneMatchHeader)
	if ifNoneMatch == cacheEntry.eTag {
		c.W.WriteHeader(http.StatusNotModified)
		return
	}

	// Check if the caller submitted a time via 'If-Modified-Since'.
	// If they did, and our cached ETag entry is not newer than the
	// given time, this means the caller has already seen the latest
	// version of this feed, so just bail.
	ifModifiedSince := extractIfModifiedSince(c)
	if !ifModifiedSince.IsZero() &&
		!unixAfter(cacheEntry.lastModified, ifModifiedSince) {
		c.W.WriteHeader(http.StatusNotModified)
		return
	}

	// At this point we know that the client wants the newest
	// representation of the RSS feed, either because they didn't
	// submit any 'If-None-Match' / 'If-Modified-Since' cache headers,
	// or because they did but the account has posted more recently
	// than the values of the submitted headers would suggest.
	//
	// If we had a cache hit earlier, we may not have called the
	// getRSSFeed function yet; if that's the case then do call it
	// now because we definitely need it.
	if feed == nil {
		feed, errWithCode = getFunc()
		if errWithCode != nil {
			apiutil.WebErrorHandler(c, m.templates, errWithCode)
			return
		}
	}

	// Encode response.
	switch contentType {
	case apiutil.AppRSSXML:
		httputil.XMLType(c, http.StatusOK, appRSSUTF8, (&feeds.Rss{feed}).FeedXml())
	case apiutil.AppAtomXML:
		httputil.XMLType(c, http.StatusOK, appAtomUTF8, (&feeds.Atom{feed}).FeedXml())
	case apiutil.AppFeedJSON, apiutil.AppJSON:
		httputil.JSONType(c, http.StatusOK, appJSONUTF8, (&feeds.JSON{feed}).JSONFeed())
	}
}

// generateFeedETag generates feed etag for appropriate content-type encoding.
func generateFeedETag(feed *feeds.Feed, contentType string) (string, error) {
	switch contentType {
	case apiutil.AppRSSXML:
		return generateETagFrom(feed.WriteRss)
	case apiutil.AppAtomXML:
		return generateETagFrom(feed.WriteAtom)
	case apiutil.AppFeedJSON, apiutil.AppJSON:
		return generateETagFrom(feed.WriteJSON)
	default:
		panic("unreachable")
	}
}

// unixAfter returns true if the unix value of t1
// is greater than (ie., after) the unix value of t2.
func unixAfter(t1 time.Time, t2 time.Time) bool {
	if t1.IsZero() {
		// if t1 is zero then it cannot
		// possibly be greater than t2.
		return false
	}

	if t2.IsZero() {
		// t1 is not zero but t2 is,
		// so t1 is necessarily greater.
		return true
	}

	return t1.Unix() > t2.Unix()
}

// extractIfModifiedSince parses a time.Time from the
// 'If-Modified-Since' header of the given request.
//
// If no time was provided, or the provided time was
// not parseable, it will return a zero time.
func extractIfModifiedSince(c *httputil.Context) time.Time {
	val := c.R.Header.Get(ifModifiedSinceHeader)
	if val == "" {
		return time.Time{} // Nothing set.
	}

	ifModifiedSince, err := http.ParseTime(val)
	if err != nil {
		log.Errorf(c, "error parsing %q: %v", val, err)
		return time.Time{}
	}

	return ifModifiedSince
}
