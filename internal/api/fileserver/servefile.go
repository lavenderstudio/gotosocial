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

package fileserver

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gopkg/log"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
)

// ServeFile is for serving attachments, headers, and avatars to the requester from instance storage.
//
// Note: to mitigate scraping attempts, no information should be given out on a bad request except "404 page not found".
// Don't give away account ids or media ids or anything like that; callers shouldn't be able to infer anything.
func (m *Module) ServeFile(c *httputil.Context) {
	authed, errWithCode := apiutil.TokenAuth(c, apiutil.AuthRequirements{
		Token:   false,
		App:     false,
		User:    false,
		Account: false,
		Scope:   nil,
	})
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	// We use request params to check what to pull out of the database/storage so check everything. A request URL should be formatted as follows:
	// "https://example.org/fileserver/[ACCOUNT_ID]/[MEDIA_TYPE]/[MEDIA_SIZE]/[FILE_NAME]"
	// "FILE_NAME" consists of two parts, the attachment's database id, a period, and the file extension.
	accountID := c.PathValue(AccountIDKey)
	if accountID == "" {
		err := fmt.Errorf("missing %s from request", AccountIDKey)
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorNotFound(err))
		return
	}

	mediaType := c.PathValue(MediaTypeKey)
	if mediaType == "" {
		err := fmt.Errorf("missing %s from request", MediaTypeKey)
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorNotFound(err))
		return
	}

	mediaSize := c.PathValue(MediaSizeKey)
	if mediaSize == "" {
		err := fmt.Errorf("missing %s from request", MediaSizeKey)
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorNotFound(err))
		return
	}

	fileName := c.PathValue(FileNameKey)
	if fileName == "" {
		err := fmt.Errorf("missing %s from request", FileNameKey)
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorNotFound(err))
		return
	}

	content, errWithCode := m.processor.Media().GetFile(c, authed.Account, &apimodel.GetContentRequestForm{
		AccountID: accountID,
		MediaType: mediaType,
		MediaSize: mediaSize,
		FileName:  fileName,
	})
	if errWithCode != nil {
		apiutil.ErrorHandler(c, m.templates, errWithCode)
		return
	}

	if content.URL != nil {
		// This is a non-local, non-proxied S3 file we're redirecting to. Derive
		// the max-age value from how long the link has left until it expires.
		maxAge := int(time.Until(content.URL.Expiry).Seconds())
		c.W.Header().Set("Cache-Control", "private, max-age="+strconv.Itoa(maxAge)+", immutable")
		httputil.Redirect(c, http.StatusFound, content.URL.String())
		return
	}

	defer func() {
		// Close content when we're done, catch errors.
		if err := content.Content.Close(); err != nil {
			log.Errorf(c, "ServeFile: error closing readcloser: %s", err)
		}
	}()

	// TODO: if the requester only accepts text/html we should try to serve them *something*.
	// This is mostly needed because when sharing a link to a gts-hosted file on something like mastodon, the masto servers will
	// attempt to look up the content to provide a preview of the link, and they ask for text/html.
	_, err := apiutil.NegotiateAccept(c, content.ContentType)
	if err != nil {
		apiutil.ErrorHandler(c, m.templates, gtserror.NewErrorNotAcceptable(err, err.Error()))
		return
	}

	// if this is a head request, just
	// return info + throw the reader away.
	if c.R.Method == http.MethodHead {
		c.W.Header().Set("Content-Type", content.ContentType)
		c.W.Header().Set("Content-Length", strconv.FormatInt(content.ContentLength, 10))
		c.W.WriteHeader(http.StatusOK)
		return
	}

	// Set known media content type and serve the file.
	c.W.Header().Set("Content-Type", content.ContentType)
	httputil.ServeFile(c,
		content.Content,
		content.ContentLength,

		// Only serve actual file content
		// if this isn't a HEAD request.
		c.R.Method != "HEAD",
	)
}
