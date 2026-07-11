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

package binding

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"

	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gopkg/httputil/binding/internal"
	"code.superseriousbusiness.org/gopkg/xencoding"
)

const (
	MIMEJSON              = "application/json"
	MIMEXML               = "application/xml"
	MIMEXML2              = "text/xml"
	MIMEPOSTForm          = "application/x-www-form-urlencoded"
	MIMEMultipartPOSTForm = "multipart/form-data"
)

// ShouldBind attempts to bind request data according
// to HTTP method and header content-type into dst.
func ShouldBind(c *httputil.Context, dst any, maxMemory int64) error {

	// GET can only bind with
	// query form parameters.
	if c.R.Method == "GET" {
		return BindForm(c, dst, maxMemory)
	}

	// Handle binding based on content-type.
	switch ct, _, _ := c.GetMediaType(); ct {
	case MIMEJSON:
		return BindJSON(c, dst, maxMemory)
	case MIMEXML, MIMEXML2:
		return BindXML(c, dst, maxMemory)
	case MIMEMultipartPOSTForm:
		return BindFormMultipart(c, dst, maxMemory)
	case MIMEPOSTForm, "":
		return BindForm(c, dst, maxMemory)
	default:
		return fmt.Errorf("unexpected content-type: %s", ct)
	}
}

// BindForm attempts to bind form data from request into dst.
func BindForm(c *httputil.Context, dst any, maxMemory int64) error {
	_, form, err := c.ReadForm(maxMemoryOrDefault(maxMemory))
	if err != nil {
		return err
	}
	return internal.MapForm(dst, form)
}

// BindFormMultipart attempts to bind multipart form data from request into dst.
func BindFormMultipart(c *httputil.Context, dst any, maxMemory int64) error {
	_, _, err := c.ReadForm(maxMemoryOrDefault(maxMemory))
	if err != nil {
		return err
	}
	return internal.MapFormMultipart(dst, c.R)
}

// BindJSON attempts to bind JSON request body into dst.
func BindJSON(c *httputil.Context, dst any, maxMemory int64) error {
	body := io.LimitReader(c.R.Body, maxMemoryOrDefault(maxMemory))
	err := xencoding.Decode(body, json.NewDecoder, dst)
	_ = c.R.Body.Close()
	return err
}

// BindXML attempts to bind XML request body into dst.
func BindXML(c *httputil.Context, dst any, maxMemory int64) error {
	body := io.LimitReader(c.R.Body, maxMemoryOrDefault(maxMemory))
	err := xencoding.Decode(body, xml.NewDecoder, dst)
	_ = c.R.Body.Close()
	return err
}

func maxMemoryOrDefault(maxMemory int64) int64 {
	if maxMemory < 1 {
		maxMemory = 10 * 1024 * 1024 // 10MiB
	}
	return maxMemory
}
