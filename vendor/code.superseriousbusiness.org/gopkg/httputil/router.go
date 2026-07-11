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
	"net/http"
	"path"
	"slices"
	"strings"
)

// HandlerFunc ...
type HandlerFunc = func(c *Context)

// Middleware is an interface that provides a middleware handler
// with control over the control-flow of its provided HandlerFunc.
type Middleware interface{ Compile(HandlerFunc) HandlerFunc }

// MiddlewareFunc is a function signature that implements Middleware{}.
type MiddlewareFunc func(h HandlerFunc) HandlerFunc

func (m MiddlewareFunc) Compile(h HandlerFunc) HandlerFunc {
	if m == nil || h == nil {
		panic("nil func")
	}
	return m(h)
}

// FlatMiddleware is an alternative Middleware{} interface
// that may optionally be implemented, which specifies that
// the passed middleware needs no control over the handler
// control-flow, and can instead be executed in-line.
//
// i.e. handler1(); handler2()
//
// as oppposed to
//
// handler1(func() { handler2() })
type FlatMiddleware interface {
	Middleware
	Handler() HandlerFunc
}

// FlatMiddlewareFunc is a function
// signature that implements FlatMiddleware{}.
type FlatMiddlewareFunc HandlerFunc

func (m FlatMiddlewareFunc) Compile(h HandlerFunc) HandlerFunc {
	if m == nil || h == nil {
		panic("nil func")
	}
	return func(c *Context) {
		m(c)
		h(c)
	}
}

func (m FlatMiddlewareFunc) Handler() HandlerFunc {
	return HandlerFunc(m)
}

// Router ...
type Router struct {

	// Handler to use when no
	// handler is found for request.
	NotFound HandlerFunc

	// Handler to use when no
	// handler is found for request,
	// but matches exist for other
	// methods on same request path.
	NoMethod HandlerFunc

	// Handler to use when an OPTIONS
	// request is received without an
	// appropriate handler set for path.
	Options HandlerFunc

	t Tree // root node in search tree
	m []Middleware
}

// SetNotFound sets the NotFound handler,
// wrapping in any base router middleware.
func (r *Router) SetNotFound(h HandlerFunc) {
	if h == nil {
		r.NotFound = nil
		return
	}
	r.NotFound = Compile(r.m, h)
}

// SetNoMethod sets the NoMethod handler,
// wrapping in any base router middleware.
func (r *Router) SetNoMethod(h HandlerFunc) {
	if h == nil {
		r.NoMethod = nil
		return
	}
	r.NoMethod = Compile(r.m, h)
}

// SetOptions sets the Options handler,
// wrapping in any base router middleware.
func (r *Router) SetOptions(h HandlerFunc) {
	if h == nil {
		r.Options = nil
		return
	}
	r.Options = Compile(r.m, h)
}

// Find looks for an HandlerFunc under given method and path, otherwise returning a list of alternate supported
// methods for matching path. Named parameters / wildcard values are stored in given parameter slice pointer.
func (r *Router) Find(method string, path string, params *Params) (HandlerFunc, string) {
	return r.t.Find(method, path, params)
}

// ServeHTTP: implements http.Handler.
func (rr *Router) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	// Alloc context.
	c := new(Context)
	c.P = make(Params, 0, 2)
	c.V = make(KVs, 8)

	// Setup
	// context.
	c.set(rw, r)

	// Get request router
	// search parameters.
	method := r.Method
	path := r.URL.Path

	// Look for a handler function under given params.
	handler, methods := rr.t.Find(method, path, &c.P)
	if handler != nil {

		// Pass to
		// handler.
		handler(c)
		return
	}

	if r.Method == "OPTIONS" {
		// Respond with methods available.
		c.W.Header().Set("Allow", methods)
		if rr.Options != nil {
			rr.Options(c)
			return
		} else {
			c.W.WriteHeader(200)
			return
		}
	}

	if len(methods) > 0 {
		// No handler was found, but alt methods
		// are supported, respond with HTTP 405.
		c.W.Header().Set("Allow", methods)
		if rr.NoMethod != nil {
			rr.NoMethod(c)
			return
		} else {
			Error(c, 405, "Method Not Allowed")
			return
		}
	}

	// No handler was found.
	if rr.NotFound != nil {
		rr.NotFound(c)
		return
	} else {
		Error(c, 404, "Not Found")
		return
	}
}

func (r *Router) HEAD(pattern string, h HandlerFunc)   { r.Handle("HEAD", pattern, h) }
func (r *Router) GET(pattern string, h HandlerFunc)    { r.Handle("GET", pattern, h) }
func (r *Router) POST(pattern string, h HandlerFunc)   { r.Handle("POST", pattern, h) }
func (r *Router) PUT(pattern string, h HandlerFunc)    { r.Handle("PUT", pattern, h) }
func (r *Router) PATCH(pattern string, h HandlerFunc)  { r.Handle("PATCH", pattern, h) }
func (r *Router) DELETE(pattern string, h HandlerFunc) { r.Handle("DELETE", pattern, h) }

// Handle registers a new handler under the given HTTP method and routing path pattern.
func (r *Router) Handle(method, pattern string, h HandlerFunc) {
	r.handle(method, pattern, slices.Clone(r.m), h)
}

// handle registers a new handler under the given HTTP method and routing path pattern, with
// middleware. this is used internally as .h() func passed to subsequent RouteGroup{} instances.
func (r *Router) handle(method, pattern string, m []Middleware, h HandlerFunc) {
	if !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}
	r.t.Add(method, pattern, Compile(m, h))
}

// Static creates a new http.Dir(path) and passes the result to Router{}.StaticFS().
func (r *Router) Static(pattern, path string) {
	r.StaticFS(pattern, http.Dir(path))
}

// StaticFS registers a new httputil.StaticFS(fs) at given pattern prefix.
func (r *Router) StaticFS(pattern string, fs http.FileSystem) {
	if r == nil {
		panic("nil router")
	}
	staticfs := &StaticFS{FileSystem: fs, NotFound: func(c *Context) {
		if r.NotFound != nil {
			r.NotFound(c)
			return
		} else {
			Error(c, 404, "Not Found")
			return
		}
	}}
	h := func(c *Context) { staticfs.ServeFile(c, c.PathValue("filepath")) }
	pattern = path.Join(pattern, "*filepath")
	r.Handle("HEAD", pattern, h)
	r.Handle("GET", pattern, h)
}

// Use appends middleware to this router instance.
// This will only effect handlers added after this point.
func (r *Router) Use(m ...Middleware) {
	r.m = append(r.m, m...)
}

// Group registers a new sub-RouteGroup{} under given prefix.
func (r *Router) Group(prefix string) *RouteGroup {

	// Return new group
	// that uses prefix.
	return &RouteGroup{
		r: r,
		h: func(method, pattern string, m []Middleware, h HandlerFunc) {
			r.handle(method, path.Join(prefix, pattern), safeAppend(r.m, m), h)
		},
	}
}

// RouteGroup ...
type RouteGroup struct {
	r *Router
	h func(method, pattern string, m []Middleware, h HandlerFunc)
	m []Middleware
}

// check ensures that RouteGroup was
// correctly initialized from Router{}.
func (g *RouteGroup) check() {
	if g.r == nil || g.h == nil {
		panic("not initialized from router")
	}
}

func (g *RouteGroup) HEAD(pattern string, h HandlerFunc)   { g.Handle("HEAD", pattern, h) }
func (g *RouteGroup) GET(pattern string, h HandlerFunc)    { g.Handle("GET", pattern, h) }
func (g *RouteGroup) POST(pattern string, h HandlerFunc)   { g.Handle("POST", pattern, h) }
func (g *RouteGroup) PUT(pattern string, h HandlerFunc)    { g.Handle("PUT", pattern, h) }
func (g *RouteGroup) PATCH(pattern string, h HandlerFunc)  { g.Handle("PATCH", pattern, h) }
func (g *RouteGroup) DELETE(pattern string, h HandlerFunc) { g.Handle("DELETE", pattern, h) }

// Handle registers a new handler under the given HTTP method and routing path pattern.
func (g *RouteGroup) Handle(method, pattern string, h HandlerFunc) {
	g.check()
	g.h(method, pattern, g.m, h)
}

// Static creates a new http.Dir(path) and passes the result to RouteGroup{}.StaticFS().
func (g *RouteGroup) Static(pattern, path string) {
	g.StaticFS(pattern, http.Dir(path))
}

// StaticFS registers a new httputil.StaticFS(fs) at given pattern prefix.
func (g *RouteGroup) StaticFS(pattern string, fs http.FileSystem) {
	g.check()
	r := g.r
	staticfs := &StaticFS{FileSystem: fs, NotFound: func(c *Context) {
		if r.NotFound != nil {
			r.NotFound(c)
			return
		} else {
			Error(c, 404, "Not Found")
			return
		}
	}}
	h := func(c *Context) { staticfs.ServeFile(c, c.PathValue("filepath")) }
	pattern = path.Join(pattern, "*filepath")
	g.Handle("HEAD", pattern, h)
	g.Handle("GET", pattern, h)
}

// Use appends middleware to this route group instance.
// This will only effect handlers added after this point.
func (g *RouteGroup) Use(m ...Middleware) {
	g.m = append(g.m, m...)
}

// Group registers a new sub-RouteGroup{} under given prefix.
func (g *RouteGroup) Group(prefix string) *RouteGroup {
	g.check()

	// Return new group
	// that uses prefix.
	return &RouteGroup{
		r: g.r,
		h: func(method, pattern string, m []Middleware, h HandlerFunc) {
			g.h(method, path.Join(prefix, pattern), safeAppend(g.m, m), h)
		},
	}
}

// Compile returns a prepared http.HandlerFunc function compiled
// from preliminary middleware handlers and the final handler func.
func Compile(m []Middleware, h HandlerFunc) HandlerFunc {
	var fs []FlatMiddleware
	slices.Reverse(m)
	for _, m := range m {
		if nilMiddleware(m) {
			continue
		}
		if f, ok := m.(FlatMiddleware); ok {
			fs = append(fs, f)
			continue
		}
		if len(fs) > 0 {
			slices.Reverse(fs)
			h = flatten(fs).Compile(h)
			clear(fs[0:cap(fs)])
			fs = fs[:0]
		}
		h = m.Compile(h)
	}
	if len(fs) > 0 {
		slices.Reverse(fs)
		h = flatten(fs).Compile(h)
	}
	return h
}

func safeAppend[T any](s1, s2 []T) []T {
	s3 := make([]T, 0, len(s1)+len(s2))
	s3 = append(s3, s1...)
	s3 = append(s3, s2...)
	return s3
}

// nilMiddleware returns whether middleware
// is a known nil (as in inoperable) value.
func nilMiddleware(m Middleware) bool {
	switch m := m.(type) {
	case nil:
		return true
	case MiddlewareFunc:
		return (m == nil)
	case FlatMiddlewareFunc:
		return (m == nil)
	default:
		return false
	}
}

// flatten will compress slice of flattenable
// middlware into a single middleware function.
func flatten(fs []FlatMiddleware) Middleware {
	if len(fs) == 1 {
		return fs[0]
	}
	ms := make([]HandlerFunc, len(fs))
	for i, f := range fs {
		ms[i] = f.Handler()
		if ms[i] == nil {
			panic("nil func")
		}
	}
	return MiddlewareFunc(func(h func(*Context)) func(*Context) {
		if h == nil {
			panic("nil func")
		}
		return func(c *Context) {
			for _, m := range ms {
				m(c)
			}
			h(c)
		}
	})
}
