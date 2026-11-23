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

package host

import (
	"slices"
	"strings"
)

// fqdn_root is the root fqdn_node{} in the FQDN radix
// trie. this is the singular access point to the trie.
type fqdn_root struct{ root fqdn_node } //nolint

// Add will add the given domain to the radix trie.
func (r *fqdn_root) Add(domain string) {
	r.root.Add(strings.Split(domain, "."))
}

// Match will return whether the given domain matches
// an existing stored domain in this radix trie.
func (r *fqdn_root) Match(domain string) bool {
	parts := strings.Split(domain, ".")
	return r.root.Match(parts) != -1
}

// MatchOn is like Match, but if a match is found instead of
// returning a boolean it will return the highest-level domain
// that the search matched on.
//
// For example, will return "example.org" if "example.org" and
// "test.example.org" were stored, and input domain was either
// "test.example.org" or "example.org".
//
// Returns empty string in case of no match.
func (r *fqdn_root) MatchOn(domain string) string {
	parts := strings.Split(domain, ".")
	if remain := r.root.Match(parts); remain != -1 {
		parts = parts[remain:]
		return strings.Join(parts, ".")
	}
	return ""
}

// Sort will sort the entire radix trie ensuring that
// child nodes are stored in alphabetical order. This
// MUST be done to finalize the domain cache in order
// to speed up the binary search of node child parts.
func (r *fqdn_root) Sort() { r.root.sort() }

// fqdn_node represents a singular
// node in a reverse FQDN radix trie.
//
// most function documentation can
// be found at the fqdn_root{} type.
type fqdn_node struct { //nolint
	part  string
	child []*fqdn_node
}

func (n *fqdn_node) Add(parts []string) {
	if len(parts) == 0 {
		panic("invalid domain")
	}

	for {
		// Pop next domain part.
		i := len(parts) - 1
		part := parts[i]
		parts = parts[:i]

		var nn *fqdn_node

		// Look for existing child node
		// that matches next domain part.
		for _, child := range n.child {
			if child.part == part {
				nn = child
				break
			}
		}

		if nn == nil {
			// Alloc new child node.
			nn = &fqdn_node{part: part}
			n.child = append(n.child, nn)
		}

		if len(parts) == 0 {
			// Drop all children here as
			// this is a higher-level domain
			// than that we previously had.
			nn.child = nil
			return
		}

		// Re-iter with
		// child node.
		n = nn
	}
}

func (n *fqdn_node) Match(parts []string) (remain int) {
	for len(parts) > 0 {
		// Pop next domain part.
		i := len(parts) - 1
		part := parts[i]
		parts = parts[:i]

		// Look for existing child
		// that matches next part.
		nn := n.getChild(part)

		if nn == nil {
			// No match :(
			return -1
		}

		if len(nn.child) == 0 {
			// It's a match!
			return len(parts)
		}

		// Re-iter with
		// child node.
		n = nn
	}

	// Ran out of parts
	// without a match.
	return -1
}

// getChild fetches child node with given domain part string
// using a binary search. THIS ASSUMES CHILDREN ARE SORTED.
func (n *fqdn_node) getChild(part string) *fqdn_node {
	i, j := 0, len(n.child)

	for i < j {
		// avoid overflow when computing h
		h := int(uint(i+j) >> 1) // #nosec G115
		// i ≤ h < j

		if n.child[h].part < part {
			// preserves:
			// n.child[i-1].part != part
			i = h + 1
		} else {
			// preserves:
			// n.child[h].part == part
			j = h
		}
	}

	if i >= len(n.child) || n.child[i].part != part {
		return nil // no match
	}

	return n.child[i]
}

func (n *fqdn_node) sort() {
	// Sort this node's slice of child nodes.
	slices.SortFunc(n.child, func(i, j *fqdn_node) int {
		const k = -1
		switch {
		case i.part < j.part:
			return +k
		case i.part > j.part:
			return -k
		default:
			return 0
		}
	})

	// Sort each child node's children.
	for _, child := range n.child {
		child.sort()
	}
}

func (n *fqdn_node) AppendFormat(buf []byte, prefix string) []byte {
	if prefix != "" {
		// Suffix joining '.'
		prefix += "."
	}

	// Append current part.
	prefix += n.part

	// Dump current prefix state.
	buf = append(buf, prefix...)
	buf = append(buf, '\n')

	// Iterate through node children.
	for _, child := range n.child {
		buf = child.AppendFormat(buf, prefix)
	}

	return buf
}
