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
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"

	"code.superseriousbusiness.org/gopkg/log"
)

// Redirect is a wrapper around http.Redirect() for httputil.Context{}.
func Redirect(c *Context, code int, url string) { http.Redirect(&c.W, c.R, url, code) }

// Error is a wrapper around http.Error() for httputil.Context{}.
func Error(c *Context, code int, msg string) { http.Error(&c.W, msg, code) }

// NegotiateAccept wraps NegotiateFormat() to return error if no format can be determined.
func NegotiateAccept(c *Context, offers ...string) (format string, err error) {
	format, accepts := NegotiateFormat(c.R.Header, offers...)
	if format == "" {
		return "", fmt.Errorf("could not determine content-type from %v to use; offers %v", accepts, offers)
	}
	return
}

// JSON calls EncodeJSONResponse() using Context{}, with content-type = AppJSON,
// This function handles the case of JSON unmarshal errors and pools read buffers.
func JSON(c *Context, code int, data any) {
	EncodeJSONResponse(c, code, "application/json", data)
}

// JSON calls EncodeJSONResponse() using Context{}, with given content-type.
// This function handles the case of JSON unmarshal errors and pools read buffers.
func JSONType(c *Context, code int, contentType string, data any) {
	EncodeJSONResponse(c, code, contentType, data)
}

// XML calls EncodeJSONResponse() using Context{}, with content-type = AppXML,
// This function handles the case of XML unmarshal errors and pools read buffers.
func XML(c *Context, code int, data any) {
	EncodeXMLResponse(c, code, "application/xml", data)
}

// XML calls EncodeXMLResponse() using Context{}, with given content-type.
// This function handles the case of XML unmarshal errors and pools read buffers.
func XMLType(c *Context, code int, contentType string, data any) {
	EncodeXMLResponse(c, code, contentType, data)
}

// Data calls WriteResponseBytes() using Context{}, with given content-type.
func Data(c *Context, code int, contentType string, data []byte) {
	WriteResponseBytes(c, code, contentType, data)
}

// Data calls WriteResponseString() using Context{}, with given content-type.
func String(c *Context, code int, contentType string, data string) {
	WriteResponseString(c, code, contentType, data)
}

// WriteResponse buffered streams 'data' as HTTP response
// to ResponseWriter with given status code content-type.
func WriteResponse(
	c *Context,
	statusCode int,
	contentType string,
	data io.Reader,
	length int64,
) {
	if length < 0 {
		// The worst-case scenario, length is not known so we need to
		// read the entire thing into memory to know length & respond.
		writeResponseUnknownLength(c, statusCode, contentType, data)
		return
	}

	// Get rw header.
	h := c.W.Header()

	// The best-case scenario,
	// streamed content of known length.
	h.Set("Content-Type", contentType)
	h.Set("Content-Length", strconv.FormatInt(length, 10))
	c.W.WriteHeader(statusCode)

	// Write streamed response to client with data.
	if _, err := c.W.ReadFrom(data); err != nil {
		log.Errorf(c, "error streaming: %v", err)
	}
}

// WriteResponseBytes is functionally similar to
// WriteResponse except that it takes prepared bytes.
func WriteResponseBytes(
	c *Context,
	statusCode int,
	contentType string,
	data []byte,
) {
	h := c.W.Header()
	h.Set("Content-Type", contentType)
	h.Set("Content-Length", strconv.Itoa(len(data)))
	c.W.WriteHeader(statusCode)
	if _, err := c.W.Write(data); err != nil && err != io.EOF {
		log.Errorf(c, "error writing: %v", err)
	}
}

// WriteResponseBytes is functionally similar to
// WriteResponse except that it takes prepared string.
func WriteResponseString(
	c *Context,
	statusCode int,
	contentType string,
	data string,
) {
	h := c.W.Header()
	h.Set("Content-Type", contentType)
	h.Set("Content-Length", strconv.Itoa(len(data)))
	c.W.WriteHeader(statusCode)
	if _, err := c.W.WriteString(data); err != nil && err != io.EOF {
		log.Errorf(c, "error writing: %v", err)
	}
}

// EncodeJSONResponse encodes 'data' as JSON HTTP response
// to ResponseWriter with given status code, content-type.
func EncodeJSONResponse(
	c *Context,
	statusCode int,
	contentType string,
	data any,
) {
	// Acquire buffer.
	buf := buf1k.Get()

	// Wrap buffer in JSON encoder.
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)

	// Encode JSON data into byte buffer.
	if err := enc.Encode(data); err == nil {

		// Drop new-line added by encoder.
		if buf.B[len(buf.B)-1] == '\n' {
			buf.B = buf.B[:len(buf.B)-1]
		}

		// Respond with the now-known
		// size byte slice within buf.
		WriteResponseBytes(c,
			statusCode,
			contentType,
			buf.B,
		)

	} else {
		// This will always be a JSON error, we
		// can't really add any more useful context.
		log.Error(c, err)

		// Any error returned here is unrecoverable.
		http.Error(&c.W, "Internal Server Error",
			http.StatusInternalServerError)
	}

	// Release buffer.
	buf1k.Put(buf)
}

// EncodeJSONResponse encodes 'data' as XML HTTP response
// to ResponseWriter with given status code, content-type.
func EncodeXMLResponse(
	c *Context,
	statusCode int,
	contentType string,
	data any,
) {
	// Acquire buffer.
	buf := buf1k.Get()

	// Write XML header string to buf.
	buf.B = append(buf.B, xml.Header...)

	// Wrap buffer in XML encoder.
	enc := xml.NewEncoder(buf)
	enc.Indent("", "  ")

	// Encode JSON data into byte buffer.
	if err := enc.Encode(data); err == nil {

		// Respond with the now-known
		// size byte slice within buf.
		WriteResponseBytes(c,
			statusCode,
			contentType,
			buf.B,
		)

	} else {
		// This will always be an XML error, we
		// can't really add any more useful context.
		log.Error(c, err)

		// Any error returned here is unrecoverable.
		http.Error(&c.W, "Internal Server Error",
			http.StatusInternalServerError)
	}

	// Release buffer.
	buf1k.Put(buf)
}

// EncodeCSVResponse encodes 'records' as CSV HTTP response
// to ResponseWriter with given status code, using CSV content-type.
func EncodeCSVResponse(
	c *Context,
	statusCode int,
	records [][]string,
) {
	// Acquire buffer.
	buf := buf1k.Get()

	// Wrap buffer in CSV writer.
	csvWriter := csv.NewWriter(buf)

	// Write all the records to the buffer.
	if err := csvWriter.WriteAll(records); err == nil {

		// Respond with the now-known
		// size byte slice within buf.
		WriteResponseBytes(c,
			statusCode,
			"text/csv",
			buf.B,
		)

	} else {
		// This will always be an csv error, we
		// can't really add any more useful context.
		log.Error(c, err)

		// Any error returned here is unrecoverable.
		http.Error(&c.W, "Internal Server Error",
			http.StatusInternalServerError)
	}

	// Release buffers.
	buf1k.Put(buf)
}

// writeResponseUnknownLength handles reading data of unknown legnth
// efficiently into memory, and passing on to WriteResponseBytes().
func writeResponseUnknownLength(
	c *Context,
	statusCode int,
	contentType string,
	data io.Reader,
) {
	// Acquire buffer.
	buf := buf16k.Get()

	// Read content into buffer.
	_, err := buf.ReadFrom(data)

	if err == nil {
		// Respond with the now-known
		// size byte slice within buf.
		WriteResponseBytes(c,
			statusCode,
			contentType,
			buf.B,
		)

	} else {
		// This will always be a reader error (non EOF),
		// but that doesn't mean the writer is closed!
		log.Errorf(c, "error reading: %v", err)
		http.Error(&c.W, "Internal Server Error",
			http.StatusInternalServerError)
	}

	// Release buffer.
	buf1k.Put(buf)
}

// RenderHTML executes named template from given
// base, passing given data. The response is stored
// in a buffer and later passed to WriteResponseBytes()
// with content-type text/html.
func RenderHTML(
	c *Context,
	statusCode int,
	base *template.Template,
	name string,
	data any,
) {

	// Acquire buffer.
	buf := buf1k.Get()

	// Execute template with name and data into byte buffer.
	if err := base.ExecuteTemplate(buf, name, data); err == nil {

		// Respond with the now-known
		// size byte slice within buf.
		WriteResponseBytes(c,
			statusCode,
			"text/html; charset=utf-8",
			buf.B,
		)

	} else {
		// Any error returned here is unrecoverable.
		log.Error(c, "error executing: %v", err)
		http.Error(&c.W, "Internal Server Error",
			http.StatusInternalServerError)
	}

	// Release buffer.
	buf1k.Put(buf)
}
