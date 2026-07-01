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
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"

	"codeberg.org/gruf/go-byteutil"
)

// Context ...
type Context struct {

	// Incoming
	// http.Request{}.
	R *http.Request

	// Path
	// params.
	P Params

	// Wrapped
	// http.ResponseWriter{}.
	W RecordResponseWriter

	// Per-context key
	// value storage.
	V KVs
}

// ToContext attempts to return existing httputil.Context{} instance
// by calling UnwrapContext() on the request's context, else wrapping
// the response writer and request in a new httputil.Context{} instance.
func ToContext(rw http.ResponseWriter, r *http.Request) *Context {
	if c := UnwrapContext(r.Context()); c != nil {
		return c
	}
	var c Context
	c.set(rw, r)
	c.P = make(Params, 0, 8)
	c.V = make(KVs, 8)
	return &c
}

// set sets up the receiving Context{} with given responsewriter and request.
func (ctx *Context) set(rw http.ResponseWriter, r *http.Request) {
	ctx.R = r.WithContext(&ctxvalue{Context: r.Context(), c: ctx})
	ctx.W = RecordResponseWriter{RW: rw}
}

// UnwrapContext attempts to return wrapped httputil.Context{}
// from the provided context.Context{}, if it has been set.
func UnwrapContext(ctx context.Context) *Context {
	switch ctx := ctx.(type) {
	case *Context:
		return ctx
	default:
		c, _ := ctx.Value(ctxkey{}).(*Context)
		return c
	}
}

// WithValue is functionally similar to context.WithValue(), but first
// attempts to call UnwrapContext() and set the key-value pair in the
// httputil.Context{}, else falling back to context.WithValue().
func WithValue(ctx context.Context, key, value any) context.Context {
	if c := UnwrapContext(ctx); c != nil {
		c.V.Set(key, value)
		return ctx
	}
	return context.WithValue(ctx, key, value)
}

// our wrapping types for storing our
// wrapping httputil.Context{} type within its
// http.Request{}'s context for later access.
//
// this also exposes the httputil.Context{}'s
// value map via context.Context{}.Value() func.
type ctxkey struct{}
type ctxvalue struct {
	context.Context
	c *Context
}

// Value: implements context.Context{}.Value().
func (ctx *ctxvalue) Value(key any) (value any) {
	_, ok := key.(ctxkey)
	if ok {
		return ctx.c
	}
	value, ok = ctx.c.V[key]
	if ok {
		return
	}
	return ctx.Context.Value(key)
}

// Deadline: implements context.Context{}.Deadline().
func (ctx *Context) Deadline() (deadline time.Time, ok bool) {
	return ctx.R.Context().Deadline()
}

// Done: implements context.Context{}.Done().
func (ctx *Context) Done() <-chan struct{} {
	return ctx.R.Context().Done()
}

// Err: implements context.Context{}.Err().
func (ctx *Context) Err() error {
	return ctx.R.Context().Err()
}

// Value: implements context.Context{}.Value().
func (ctx *Context) Value(key any) any {
	return ctx.R.Context().Value(key)
}

// PathValue is short-hand for calling ctx.P.Get(key).
func (ctx *Context) PathValue(key string) string {
	return ctx.P.Get(key)
}

// SetPathValue is short-hand for calling ctx.P.Set(key, value).
func (ctx *Context) SetPathValue(key, value string) {
	ctx.P.Set(key, value)
}

// ParseQuery attempts to parse http.Request{}.URL.RawQuery, caching
// in http.Request{}.Form. Returns cached value if already parsed.
func (ctx *Context) ParseQuery() (err error) {
	if ctx.R.Form != nil {
		return
	}
	ctx.R.Form, err = url.ParseQuery(ctx.R.URL.RawQuery)
	return
}

// Query ...
func (ctx *Context) Query(key string) string {
	v, _ := ctx.QueryGet(key)
	return v
}

// QueryGet ...
func (ctx *Context) QueryGet(key string) (string, bool) {
	vs := ctx.QueryArray(key)
	if len(vs) > 0 {
		return vs[0], true
	}
	return "", false
}

// QueryLookup ...
func (ctx *Context) QueryArray(key string) []string {
	_ = ctx.ParseQuery()
	return ctx.R.Form[key]
}

// ContentType returns the "content-type" header for this request,
// with all trailing parameters stripped. This is the same as calling
// ctx.GetMediaType() and only accepting the first parameter.
func (ctx *Context) ContentType() string {
	ct, _, _ := ctx.GetMediaType()
	return ct
}

// GetMediaType parses the "content-type" header value with mime.ParseMediaType(),
// caching result to prevent re-parsing, and masking empty content-type errors.
func (ctx *Context) GetMediaType() (string, map[string]string, error) {

	// mediatypekey is used to
	// store mediatype{} struct
	// within context val map.
	type mediatypekey struct{}

	// mediatype wraps return vals
	// of mime.ParseMediaType().
	type mediatype struct {
		Type   string
		Params map[string]string
	}

	// Check for existing stored result in context.
	switch v := ctx.V.Get(mediatypekey{}).(type) {
	case *mediatype:
		return v.Type, v.Params, nil
	case error:
		return "", nil, v
	}

	// Get content-type from request header.
	ct := ctx.R.Header.Get("Content-Type")

	// Parse media type from header value.
	t, p, err := mime.ParseMediaType(ct)
	switch {
	case err == nil:
		ctx.V.Set(mediatypekey{}, &mediatype{t, p})
		return t, p, nil
	case err.Error() == "mime: no media type":
		ctx.V.Set(mediatypekey{}, &mediatype{})
		return "", nil, nil
	default:
		ctx.V.Set(mediatypekey{}, err)
		return "", nil, err
	}
}

// ReadForm ...
func (ctx *Context) ReadForm(maxMemory int64) (*multipart.Form, url.Values, error) {

	// Inititially, parse query values.
	if err := ctx.ParseQuery(); err != nil {
		return nil, nil, err
	} else if ctx.R.Form == nil {
		return nil, nil, errors.New("nil form")
	}

	// If method without body,
	// query values *are* form.
	switch ctx.R.Method {
	case "POST", "PUT", "PATCH":
	default:
		return nil, ctx.R.Form, nil
	}

	// Return existing form values if already been parsed.
	if ctx.R.PostForm != nil || ctx.R.MultipartForm != nil {
		return ctx.R.MultipartForm, ctx.R.Form, nil
	}

	// Read content-type + params from hdr.
	ct, params, err := ctx.GetMediaType()
	if err != nil {
		return nil, nil, err
	}

	switch ct {
	case "application/x-www-form-urlencoded":
		// Limit request body to given maximum memory.
		body := io.LimitReader(ctx.R.Body, maxMemory)

		// Read req body into memory.
		b, err := io.ReadAll(body)
		if err != nil {
			return nil, nil, fmt.Errorf("error reading post form: %w", err)
		}

		// Parse the in-memory post form as URL encoded values.
		ctx.R.PostForm, err = url.ParseQuery(byteutil.B2S(b))
		if err != nil {
			return nil, nil, fmt.Errorf("error parsing post form: %w", err)
		}

		// Append post values to regular form.
		for k, vs := range ctx.R.PostForm {
			ctx.R.Form[k] = append(ctx.R.Form[k], vs...)
		}

	case "multipart/form-data":
		// Get boundary from media params.
		boundary, ok := params["boundary"]
		if !ok {
			return nil, nil, http.ErrMissingBoundary
		}

		// Wrap request bound in new multipart reader.
		mr := multipart.NewReader(ctx.R.Body, boundary)

		// Read the multipart form using wrapped reader.
		ctx.R.MultipartForm, err = mr.ReadForm(maxMemory)
		if err != nil {
			return nil, nil, fmt.Errorf("error reading multipart form: %w", err)
		}

		// Append the multipart values to regular form.
		for k, vs := range ctx.R.MultipartForm.Value {
			ctx.R.Form[k] = append(ctx.R.Form[k], vs...)
		}
	}

	return ctx.R.MultipartForm, ctx.R.Form, nil
}

// errorkey is used to
// store error slice ptr
// within context val map.
type errorkey struct{}

// Error appends error to stored
// error slice in Context{} value map.
func (ctx *Context) Error(err error) {
	ptr, _ := ctx.V[errorkey{}].(*[]error)
	if ptr == nil {
		ptr = new([]error)
		ctx.V[errorkey{}] = ptr
	}
	(*ptr) = append((*ptr), err)
}

// Errors returns any errors set via Error().
func (ctx *Context) Errors() []error {
	ptr, _ := ctx.V[errorkey{}].(*[]error)
	if ptr == nil {
		return nil
	}
	return (*ptr)
}
