// portions of the below content negotiation are modified from:
// https://github.com/gin-gonic/gin
//
// The MIT License (MIT)
//
// Copyright (c) 2014-present Manuel Martínez-Almeida
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package httputil

import (
	"net/http"
	"slices"
	"strings"
)

// NegotiateFormat takes the http.Header from an incoming request, and a
// slice of Offers, and performs content negotiation for the given headers
// with the given content-type offers. It will return a string representation
// of the first suitable content-type, or an empty string. It additionally
// returns the parsed accept types provided in the header.
//
// For example, if http.Header{} has Accept headers of value:
// [application/json, text/html], and the provided offers are of value
// [application/json, application/xml], then the returned string will be
// 'application/json', which indicates the content-type that should be returned.
//
// If there are no Accept headers in the request, then the first offer will be returned,
// under the assumption that it's better to serve *something* than error out completely.
//
// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Content_negotiation#server-driven_content_negotiation
func NegotiateFormat(hdr http.Header, offers ...string) (format string, accepts []string) {
	offers = slices.DeleteFunc(offers, func(offer string) bool { return offer == "" })
	if len(offers) == 0 {
		panic("you must provide at least one non-empty offer")
	}
	for _, value := range hdr.Values("Accept") {
		accepts = append(accepts, parseAccept(value)...)
	}
	if len(accepts) == 0 {
		return offers[0], nil
	}
	for _, accept := range accepts {
		for _, offer := range offers {
			i := 0

			// Loop through accept and offer variables checking
			// for match, accounting for wildcard '*' variables.
			//
			// Noting that a wildcard in the first portion up-to slash
			// is only permitted as part of '*/*', which accepts all.
			for ; i < len(accept) && i < len(offer); i++ {
				if accept[i] == '*' || offer[i] == '*' {
					return offer, accepts
				}

				// If they don't match then
				// this offer failed, continue.
				if accept[i] != offer[i] {
					break
				}
			}

			if i == len(accept) {
				// We found a match!
				return offer, accepts
			}
		}
	}
	return "", accepts
}

// According to RFC 2616 and RFC 2396, non-ASCII characters are not allowed in headers,
// therefore we can just iterate over the string without handling as utf8 rune types.
func parseAccept(in string) (out []string) {
	for len(in) > 0 {
		var part string

		// Split input string on commas.
		i := strings.IndexByte(in, ',')
		if i >= 0 {
			part = in[:i]
			in = in[i+1:]
		} else {
			part = in
			in = ""
		}

		// Trim anything after semicolon.
		j := strings.IndexByte(part, ';')
		if j >= 0 {
			part = part[:j]
		}

		// Trim any space and append.
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return
}
