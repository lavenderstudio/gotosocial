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
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"code.superseriousbusiness.org/gopkg/buffers"
	"codeberg.org/gruf/go-fastpath/v2"
)

// Param represents a path
// parameter key-value pair.
type Param struct{ K, V string }

// Params provides helper methods
// for a slice of Param{}s.
type Params []Param

// Get returns value for param with key.
func (ps Params) Get(key string) string {
	for i := range ps {
		if ps[i].K == key {
			return ps[i].V
		}
	}
	return ""
}

// Set will set param with key, to value.
func (ps *Params) Set(key, value string) {
	for i := range *ps {
		if (*ps)[i].K == key {
			(*ps)[i].V = value
			return
		}
	}
	ps.add(key, value)
}

// add will append given key-value param to Params{}.
func (ps *Params) add(key, value string) {
	(*ps) = append((*ps), Param{K: key, V: value})
}

// Tree represents the
// root node{} in a tree.
type Tree struct{ node }

// node represents an entry in an HTTP router
// search tree. the tree is split such that each
// node represents a singular path segment, of
// type static, named parameter, or end wildcard.
//
// it differs from similarly designed data structures
// in that it structures nodes around entire path
// segments for implementation simplicity, links to
// child static nodes with a hashmap for scaling up
// to large numbers of paths, each tree handles all
// methods (unlike a separate tree per method) for
// simplicity in calculating HTTP 405 responses, and
// it is much more flexible in supporting named param,
// wildcard and static children from the same node.
//
// performance is ~2x faster than http.ServeMux{},
// but ~1/2 as slow as httprouter.Router{}, while
// being more flexible than both.
type node struct {

	// static path children.
	static map[string]*node

	// all other
	// path children.
	other []*node

	// current node type:
	// - ':' = named path parameter
	// - '*' = named wild card
	type_ byte

	// parameter
	// prefix
	// (if any).
	prefix string

	// param name.
	name string

	// handlers keyed
	// by req method.
	handlers handlers

	// full request path
	// to reach node, if
	// handlers are set.
	fullpath string
}

// Add attempts to register a new handler with node under given method and path.
func (n *node) Add(method, path string, handler HandlerFunc) {
	defer func(n *node) { n.sort() }(n)

	// Clean and normalize
	// input request path.
	path = cleanPath(path)

	// Split path into parts.
	parts := splitPath(path)
	if len(parts) <= 0 {
		panic("BCE")
	}

	// If path empty (i.e. root), skip loop.
	if len(parts) == 1 && parts[0] == "" {
		parts = nil
	}

	for i, part := range parts {
		// Next node.
		var nn *node

		// Look for next delimited wildcard or
		// named parameter byte in this path part.
		prefix, type_, part := nextDelimited(part)
		if part == "" {
			panic("empty path part / param name")
		}

		switch type_ {
		case '*':
			if i != len(parts)-1 {
				panic("wildcard must appear at path end")
			}
			fallthrough
		case ':':
			for _, child := range n.other {
				// Check for matches with child.
				namematch := (child.name == part)
				typematch := (child.type_ == type_)
				prefmatch := (child.prefix == prefix)

				// Ensure no partial matches under same prefix causing conflicts / confusion.
				if prefmatch && ((namematch && !typematch) || (!namematch && typematch)) {
					panic(fmt.Sprintf("param name conflict between %s and %s in %s", child.name, part, path))
				}

				// If fully matched, set child as next.
				if prefmatch && namematch && typematch {
					nn = child
					break
				}
			}

			if nn == nil {
				// Create new child from details and add to current.
				nn = &node{prefix: prefix, name: part, type_: type_}
				n.other = append(n.other, nn)
			}

		default:
			// Check for existing
			// static segment node.
			nn = n.static[part]
			if nn == nil {

				if n.static == nil {
					// Lazily alloc static segment map.
					n.static = make(map[string]*node)
				}

				// Add new node
				// to static map.
				nn = &node{}
				n.static[part] = nn
			}
		}

		// Iterate to
		// next node.
		n = nn
	}

	// We successfully created / fetched existing
	// nodes specified by input path pattern. Ensure
	// determined (cleaned) fullpath is as expected.
	if n.fullpath != "" && n.fullpath != path {
		panic(fmt.Sprintf("BUG: n.fullPath=%q fullPath=%q", n.fullpath, path))
	}

	// Attempt to add handler at this node
	// for the given method, catching conflict.
	if !n.handlers.Set(method, handler) {
		panic(fmt.Sprintf("handler already registed for '%s %s'", method, path))
	}

	// Ensure full node path
	// is set for debugging.
	n.fullpath = path
}

// Find looks for an HandlerFunc under given method and path, otherwise returning a list of alternate supported
// methods for matching path. Named parameters / wildcard values are stored in given parameter slice pointer.
func (n *node) Find(method, path string, params *Params) (HandlerFunc, string) {
	if params == nil {
		panic("nil params")
	}

	// Clean request path.
	path = cleanPath(path)
	if path == "" {

		// The only thing an empty path
		// can match is the root handler.
		h := n.handlers.Get(method)
		if h != nil {
			return h, ""
		}

		// Else, return supported methods.
		return nil, n.handlers.Methods()
	}

	// Split path into parts.
	parts := splitPath(path)
	if len(parts) <= 0 {
		panic("BCE")
	}

	type frame struct {
		// current node
		// being checked.
		node *node

		// whether static segment
		// part has been checked.
		static bool

		// last search index
		// in child slice.
		child int
	}

	// stack of node iteration frames,
	// permitting back-tracking from
	// current search node on failed match.
	frames := make([]frame, len(parts))
	if len(frames) != len(parts) {
		panic("BCE")
	}

	// Set first node.
	frames[0].node = n

	// last_match stores the last-found
	// match, used to provide a list of
	// alternate methods for same path.
	var last_match *node

	// set_match sets the last_match
	// variable, if not been already.
	set_match := func(n *node) {
		if last_match != nil {
			return
		}
		last_match = n
	}

walk:
	for i := 0; //
	i < len(parts); {
		// Get path part.
		part := parts[i]

		// Get walk frame
		// and set cur node.
		cur := &frames[i]
		cur.node = n

		// If not yet checked for static
		// path segment with name, do so.
		if !cur.static {
			cur.static = true

			// Check static path map.
			nn := n.static[part]
			if nn != nil {

				// Check if reached end
				// of split path segments.
				if i == len(parts)-1 {

					// Set match for later
					// alt method reference.
					set_match(nn)

					// Get handler for method.
					h := nn.handlers.Get(method)
					if h != nil {
						return h, ""
					}

					// None found, attempt
					// to rewind for further
					// matching, or return.
					goto rewind_or_return
				}

				// Iter.
				i++

				// Jump to
				// child.
				n = nn
				continue walk
			}
		}

		// Pick-up from last iteration through this
		// node's wildcard / named parameter children.
		for cur.child < len(n.other) {

			// Set child as next-node
			// and immediately increment
			// index to prevent re-match.
			nn := n.other[cur.child]
			cur.child++

			// Check for prefix.
			if nn.prefix != "" {

				// If segment prefix doesn't match
				// child's, then it can't be a match.
				if !strings.HasPrefix(part, nn.prefix) {
					continue
				}

				// Strip prefix from segment.
				part = part[len(nn.prefix):]
			}

			// Handle node type.
			switch nn.type_ {

			case '*':
				// Set match for later
				// alt method reference.
				set_match(nn)

				// Get handler for method.
				h := nn.handlers.Get(method)
				if h != nil {

					// Get all path parts after
					// wildcard by trimming all
					// previous segments (+ "/").
					for j := 0; j < i; j++ {
						partlen := len(parts[j])
						path = path[partlen+1:]
					}

					// Set wildcard param value.
					params.add(nn.name, path)
					return h, ""
				}

				// None found, attempt
				// to rewind for further
				// matching, or return.
				goto rewind_or_return

			case ':':
				// Set path param value.
				params.add(nn.name, part)

				// Check if reached end
				// of split path segments.
				if i == len(parts)-1 {

					// Set match for later
					// alt method reference.
					set_match(nn)

					// Get handler for method.
					h := nn.handlers.Get(method)
					if h != nil {
						return h, ""
					}

					// None found, attempt
					// to rewind for further
					// matching, or return.
					goto rewind_or_return
				}

				// Iter.
				i++

				// Jump to
				// child.
				n = nn
				continue walk
			}
		}

	rewind_or_return:
		if i > 0 {
			if cur.node.type_ == ':' {
				// Pop last added parameter from list.
				(*params) = (*params)[:len(*params)-1]
			}

			// Reset cur frame.
			frames[i] = frame{}

			// Decr
			// part.
			i--

			// Jump to prev.
			n = frames[i].node
			continue
		}

		// Reset provided params.
		(*params) = (*params)[:0]

		var methods string
		if last_match != nil {
			// A last match was set but no handler
			// was found, get known methods for 405.
			methods = last_match.handlers.Methods()
		}

		// Return with any
		// alternate methods.
		return nil, methods
	}

	panic("unreachable")
}

// AppendFormat appends new-line separated list of
// ${method} + ${path} entries within receiving node.
func (n *node) AppendFormat(buf []byte) []byte {
	if n.fullpath != "" {
		for i := range n.handlers.handlers {
			method := n.handlers.handlers[i].method
			buf = append(buf, method...)
			buf = append(buf, ' ')
			buf = append(buf, n.fullpath...)
			buf = append(buf, '\n')
		}
	}
	for _, child := range n.static {
		buf = child.AppendFormat(buf)
	}
	for _, child := range n.other {
		buf = child.AppendFormat(buf)
	}
	return buf
}

// sort ensures that node and all its children
// are sorted in correct matching order for Find().
func (n *node) sort() {
	for _, child := range n.static {
		child.sort()
	}
	for _, child := range n.other {
		child.sort()
	}
	slices.SortFunc(n.other, func(c1, c2 *node) int {
		l1 := c1.LongestEdge()
		l2 := c2.LongestEdge()
		switch {
		// Prefix length is the primary property
		// around which children are sorted. Those
		// with longest prefixes prioritised higher.
		case c1.prefix != "" && c2.prefix == "":
			return -1
		case c2.prefix != "" && c1.prefix == "":
			return +1
		case c1.prefix != "" && c2.prefix != "":
			return -strings.Compare(c1.prefix, c2.prefix)

		// Next highest priority is
		// length of its longest edge.
		case l1 < l2:
			return -1
		case l2 > l1:
			return +1

		// Finally, individual named
		// parameter takes precedence
		// over a wildcard parameter.
		case c1.type_ == ':' && c2.type_ != ':':
			return -1
		case c2.type_ == ':' && c1.type_ != ':':
			return +1

		// All else
		// equal.
		default:
			return 0
		}
	})
}

// Longest returns the longest edge running
// from current node, to its furthest child.
func (n *node) LongestEdge() int {
	var longest int
	for _, child := range n.static {
		if l := child.LongestEdge(); l > longest {
			longest = l
		}
	}
	for _, child := range n.other {
		if l := child.LongestEdge(); l > longest {
			longest = l
		}
	}
	return longest + 1
}

// String returns new-line separated list of
// ${method} + ${path} entries in receiving node.
func (n *node) String() string {
	return string(n.AppendFormat(nil))
}

// handlers stores a list of HandlerFuncs
// keyed by method string. this is stored
// as a list instead of hashmap as there will
// only ever be a maximum of 9 members.
type handlers struct {
	handlers []struct {
		method  string
		handler HandlerFunc
	}

	// cached methods
	// result string.
	methods string
}

// Get returns the stored handler associated with method.
func (hs handlers) Get(method string) HandlerFunc {
	for i := range hs.handlers {
		if (hs.handlers)[i].method == method {
			return hs.handlers[i].handler
		}
	}
	return nil
}

// Set attempts to store handler under given method, returning false on conflict.
func (hs *handlers) Set(method string, handler HandlerFunc) bool {
	method = strings.ToUpper(method)

	// Check for existing handler.
	for i := range hs.handlers {
		if hs.handlers[i].method == method {
			return false
		}
	}

	// Append new handler to slice under method.
	hs.handlers = append(hs.handlers, struct {
		method  string
		handler HandlerFunc
	}{
		method:  method,
		handler: handler,
	})

	// On addition of new method, recalculate
	// HTTP 405 / OPTIONS methods result string.
	methods := make([]string, len(hs.handlers))
	if len(methods) != len(hs.handlers) {
		panic("BCE")
	}
	for i := range hs.handlers {
		methods[i] = hs.handlers[i].method
	}
	hs.methods = strings.Join(methods, ",")
	return true
}

// Methods returns a list of methods registered under handlers.
func (hs handlers) Methods() string {
	return hs.methods
}

// splitPath splits the given path by slashes.
func splitPath(path string) []string {
	idxs := make([]int, 0, 64)
	for s := path; s != ""; {
		i := strings.IndexByte(s, '/')
		if i == -1 {
			break
		}
		idxs = append(idxs, i)
		s = s[i+1:]
	}
	if len(idxs) == 0 {
		return []string{path}
	}
	a := make([]string, len(idxs)+1)
	if len(a) != len(idxs)+1 {
		panic("BCE")
	}
	for i, j := range idxs {
		a[i] = path[:j]
		path = path[j+1:]
	}
	a[len(a)-1] = path
	return a
}

// cleanPath cleans the given path removing any leading or
// trailing slashes, and cleaning according to path.Clean().
func cleanPath(path string) string {
	if path == "" {
		return ""
	}

	// Acquire mem buffer.
	buf := buf256.Get()

	// Use the buffer byte slice with
	// a path builder for path cleaning.
	pb := fastpath.Builder{B: buf.B}
	pb.SetAbsolute(true)
	pb.Append(path)

	// Ensure we re-pool updated
	// path buffer if it was grown.
	if cap(pb.B) > cap(buf.B) {
		buf.B = pb.B
	}

	// Check if string has
	// actually changed, if not
	// we can release buffer now.
	if path == pb.String() {
		buf256.Put(buf)

		// Trim leading
		// path slashes.
		for len(path) > 0 &&
			path[0] == '/' {
			path = path[1:]
		}

		// Trim trailing
		// path slashes.
		for len(path) > 0 &&
			path[len(path)-1] == '/' {
			path = path[:len(path)-1]
		}

	} else {

		// Trim leading
		// path slashes.
		for len(pb.B) > 0 &&
			pb.B[0] == '/' {
			pb.B = pb.B[1:]
		}

		// Trim trailing
		// path slashes.
		for len(pb.B) > 0 &&
			pb.B[len(pb.B)-1] == '/' {
			pb.B = pb.B[:len(pb.B)-1]
		}

		// Get string path.
		path = string(pb.B)

		// Release buffer.
		buf256.Put(buf)
	}

	return path
}

// nextDelimited returns the string up-to, non-delimited byte, and
// string after non-delimited byte. For both ':' and '*' bytes, where
// they can be delimited by having consecutive instance of them.
func nextDelimited(in string) (string, byte, string) {
	var delim byte
	buf := buf256.Get()
	for i, r := range in {
		switch r {
		case ':':
			if delim == ':' {
				delim = 0
				buf.B = append(buf.B, ':')
			} else {
				delim = ':'
			}
		case '*':
			if delim == '*' {
				delim = 0
				buf.B = append(buf.B, '*')
			} else {
				delim = '*'
			}
		default:
			if delim != 0 {
				buf256.Put(buf)
				return in[:i-1], delim, in[i:]
			}
			delim = 0
			buf.B = utf8.AppendRune(buf.B, r)
		}
	}
	in = string(buf.B)
	buf256.Put(buf)
	return "", 0, in
}

// global buffer pool of 256 byte buffers.
var buf256 = buffers.Pool(256)
