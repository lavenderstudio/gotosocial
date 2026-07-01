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
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"

	"code.superseriousbusiness.org/gopkg/buffers"
	"codeberg.org/gruf/go-byteutil"
)

// StatusHijacked is a faux HTTP status code
// that can be set on RecordResponseWriter to
// indicate that http.Hijacker{} was called.
const StatusHijacked = -99

// RecordResponseWriter wraps an http.ResponseWriter{}
// to record the written status code and number of bytes.
type RecordResponseWriter struct {
	RW http.ResponseWriter

	// StatusCode contains the
	// recorded http.ResponseWriter{}
	// status code if WriteHeader has
	// been called.
	StatusCode int

	// Written contains the number
	// of bytes made to in cumulative
	// Write() / ReadFrom() calls.
	Written int64
}

// Unwrap returns the underlying http.ResponseWriter{}.
func (rr *RecordResponseWriter) Unwrap() http.ResponseWriter {
	return rr.RW
}

// Header: implements http.ResponseWriter{}.
func (rr *RecordResponseWriter) Header() http.Header {
	return rr.RW.Header()
}

// WriteHeader: implements http.ResponseWriter{}.
func (rr *RecordResponseWriter) WriteHeader(statusCode int) {
	rr.RW.WriteHeader(statusCode)
	rr.StatusCode = statusCode
}

// Write: implements http.ResponseWriter{}.
func (rr *RecordResponseWriter) Write(b []byte) (int, error) {
	if rr.StatusCode == 0 {
		// This matches default
		// http.ResponseWriter{}
		// behaviour if not yet
		// already been set.
		rr.StatusCode = 200
	}
	n, err := rr.RW.Write(b)
	rr.Written += int64(n)
	return n, err
}

// WriteString: implements io.StringWriter{}.
func (rr *RecordResponseWriter) WriteString(s string) (int, error) {
	return rr.Write(byteutil.S2B(s))
}

// ReadFrom: implements io.ReaderFrom{}.
func (rr *RecordResponseWriter) ReadFrom(r io.Reader) (n int64, err error) {
	if rr.StatusCode == 0 {
		// This matches default
		// http.ResponseWriter{}
		// behaviour if not yet
		// already been set.
		rr.StatusCode = 200
	}
	if rf, ok := rr.RW.(io.ReaderFrom); ok {
		// this should almost always be
		// true, unless some middleware
		// replaces http.ResponseWriter{}.
		n, err = rf.ReadFrom(r)
		rr.Written += n
	} else {
		// Acquire buffer.
		buf := buf16k.Get()

		// Perform copy with our fetched buffer.
		n, err = buffers.CopyBuffer(rr.RW, r, buf)
		rr.Written += n

		// Release buffer.
		buf16k.Put(buf)
	}
	return
}

// Hijack: implements http.Hijacker{}, returning error if underlying does not support it.
func (rr *RecordResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if rr.StatusCode != 0 {
		return nil, nil, errors.New("response already written")
	}
	h, ok := rr.RW.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	rr.StatusCode = StatusHijacked
	return h.Hijack()
}

// Flush: implements http.Flusher{}.
func (rr *RecordResponseWriter) Flush() {
	if f, ok := rr.RW.(http.Flusher); ok {
		f.Flush()
	}
}
