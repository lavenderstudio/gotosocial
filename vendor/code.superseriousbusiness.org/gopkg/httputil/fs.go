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

package httputil

import (
	"io"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"

	"code.superseriousbusiness.org/gopkg/log"
)

// ServeFile ...
func ServeFile(
	c *Context,
	file io.Reader,
	size int64,
	write bool,
) {
	if size < 0 {
		panic("size is required")
	}

	// Look for a given range header.
	rng := c.R.Header.Get("Range")
	if rng == "" {

		// Ensure reader only returns given size.
		if lr, ok := file.(*io.LimitedReader); !ok {
			file = io.LimitReader(file, size)
		} else if lr.N > size {
			lr.N = size
		}

		// Get rw header.
		h := c.W.Header()

		// Set given file content-length.
		h.Set("Content-Length", strconv.FormatInt(size, 10))
		c.W.WriteHeader(http.StatusOK)

		if write {
			// Read the entire file contents into writer.
			if _, err := c.W.ReadFrom(file); err != nil {
				log.Errorf(c, "error reading whole: %v", err)
			}
		}

		return
	}

	var i int

	if i = strings.IndexByte(rng, '='); i < 0 {
		// Range must have a separating '=' to indicate start.
		Error(c, http.StatusBadRequest, "Bad Range Header")
		return
	}

	if rng[:i] != "bytes" {
		// We only support byte ranges in our implementation
		Error(c, http.StatusBadRequest, "Unsupported Range Unit")
		return
	}

	// Reslice past '='
	rng = rng[i+1:]

	if i = strings.IndexByte(rng, '-'); i < 0 {
		// Range requires beginning and end separated by '-'.
		Error(c, http.StatusBadRequest, "Bad Range Header")
		return
	}

	var (
		err error

		// Default start + end ranges
		start, end = int64(0), size - 1

		// Start + end range strings
		startRng, endRng string
	)

	if startRng = rng[:i]; len(startRng) > 0 {
		// Parse the start of the provided byte range.
		start, err = strconv.ParseInt(startRng, 10, 64)
		if err != nil {
			Error(c, http.StatusBadRequest, "Bad Range Header")
			return
		}

		if start < 0 {
			// This range starts *before* the file start, why did they send this lol
			c.W.Header().Set("Content-Range", "bytes *"+strconv.FormatInt(size, 10))
			Error(c, http.StatusRequestedRangeNotSatisfiable, "Unsatisfiable Range")
			return
		}
	} else {
		// No start supplied,
		// implying file start
		startRng = "0"
	}

	if endRng = rng[i+1:]; len(endRng) > 0 {
		// Parse the end of the provided byte range.
		end, err = strconv.ParseInt(endRng, 10, 64)
		if err != nil {
			Error(c, http.StatusBadRequest, "Bad Range Header")
			return
		}

		if end >= size {
			// According to the http spec if end >= size the server should return the rest of the file
			// https://www.rfc-editor.org/rfc/rfc9110#section-14.1.2-6
			endRng = strconv.FormatInt(end, 10)
			end = size - 1
		}
	} else {
		// No end supplied, implying file end
		endRng = strconv.FormatInt(end, 10)
	}

	if start >= end {
		// This range starts _after_ their range end, unsatisfiable and nonsense!
		c.W.Header().Set("Content-Range", "bytes *"+strconv.FormatInt(size, 10))
		Error(c, http.StatusRequestedRangeNotSatisfiable, "Unsatisfiable Range")
		return
	}

	// Determine new content length
	// after slicing to given range.
	length := end - start + 1

	// Get rw header.
	h := c.W.Header()

	// Write the necessary length and range headers.
	h.Set("Content-Range", "bytes "+startRng+"-"+endRng+"/"+strconv.FormatInt(size, 10))
	h.Set("Content-Length", strconv.FormatInt(length, 10))
	c.W.WriteHeader(http.StatusPartialContent)
	if !write {
		return
	}

	if rs, ok := file.(io.ReadSeeker); ok {
		// Source supports seeking (usually *os.File),
		// seek to the 'start' byte position in file.
		if _, err := rs.Seek(start, 0); err != nil {
			log.Errorf(c, "error seeking: %v", err)
			return
		}
	} else {
		// Compat for when no seek call is implemented,
		// dump the first 'start' many bytes into void.
		src, dst := io.LimitReader(file, start), discard
		if _, err := dst.ReadFrom(src); err != nil {
			log.Errorf(c, "error reading start: %v", err)
			return
		}
	}

	if end < size-1 {
		// Range end < file end, limit it.
		file = io.LimitReader(file, length)

		// Else, even if it is within range,
		// ensure we only write up to size.
	} else if lr, ok := file.(*io.LimitedReader); !ok {
		file = io.LimitReader(file, size)
	} else if lr.N > size {
		lr.N = size
	}

	// Read the "seeked" source file into writer.
	if _, err := c.W.ReadFrom(file); err != nil {
		log.Errorf(c, "error reading after seek: %v", err)
	}
}

// StaticFS serves the given http.FileSystem{} as a static directory (i.e. no index.html handling, no generated
// directory listings), with appropriate range handling and by-extension content-type matching. The given 'pathValue'
// is used as the router path parameter to access to determine the filesystem path to access. i.e. if you register
// this StaticFS() under http.ServeMux{} pattern "GET /{filepath...}", then provide "filepath" as the pathValue.
// 'notFound' allows setting a custom file-not-found handler to be used when request filepath cannot be opened.
func StaticFS(fs http.FileSystem, pathValue string, notFound HandlerFunc) HandlerFunc {
	if notFound == nil {
		notFound = func(c *Context) { Error(c, http.StatusNotFound, "file not found") }
	}
	return func(c *Context) {
		// Get path from router path params.
		filepath := c.PathValue(pathValue)
		if !strings.HasPrefix(filepath, "/") {
			filepath = "/" + filepath
		}

		// Attempt to open file at path.
		file, err := fs.Open(filepath)
		if err != nil {
			notFound(c)
			return
		}

		// Close on return.
		defer file.Close()

		// Stat file for more info.
		stat, err := file.Stat()
		if err != nil {
			notFound(c)
			return
		}

		// Only interested in regular.
		if !stat.Mode().IsRegular() {
			notFound(c)
			return
		}

		// Try to determine content-type by its file extension.
		contentType := mime.TypeByExtension(path.Ext(filepath))
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		// Get rw header.
		h := c.W.Header()

		// Indicate content-type and
		// that we accept range requests.
		h.Set("Accept-Ranges", "bytes")
		h.Set("Content-Type", contentType)

		// Serve
		// the file.
		ServeFile(c,
			file,
			stat.Size(),

			// Only write response
			// if not HEAD request.
			c.R.Method != "HEAD",
		)
	}
}

// discard is io.Discard casted to the interfaces it supports.
var discard = io.Discard.(interface {
	io.ReaderFrom
	io.Writer
})
