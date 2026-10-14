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

//go:build !moderncsqlite3

package sqlite

import (
	"unicode"
	"unicode/utf8"

	"code.superseriousbusiness.org/gopkg/buffers"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"codeberg.org/gruf/go-byteutil"

	"github.com/clipperhouse/uax29/iterators"
	"github.com/clipperhouse/uax29/words"

	"github.com/ncruces/go-sqlite3/ext/fts5"
	"golang.org/x/text/cases"
	"golang.org/x/text/transform"
)

// buffer size
// in below pool.
const bufsz = 2048

// bufpool is a global memory pool
// of tokenizer byte buffers of bufsz.
var bufpool = buffers.Pool(bufsz)

// Uax29Tokenizer is an HTML-aware (tags, at least) string
// tokenizer for use with sqlite's FTS5 extension based on
// word splitting according to Unicode Standard Annex #29.
//
// it strips any HTML tags from
// input text for DOCUMENT queries.
type Uax29Tokenizer struct{}

// This function is expected to tokenize text. For each token in the input string, the supplied callback must be invoked.
func (t Uax29Tokenizer) Tokenize(flag fts5.TokenizeFlag, text string, _ string, xToken fts5.TokenCallback) error {
	if xToken == nil {
		// nil check performed outside loop.
		return gtserror.New("nil callback")
	}

	// Check if tokenize operation is for
	// an INSERT|UPDATE|DELETE query, or
	// if just for incoming MATCH query.
	//
	// For MATCH queries we don't need
	// to bother stripping HTML tags.
	if flag&fts5.TOKENIZE_DOCUMENT != 0 {
		text = zeroHTML(text)
	}

	// Acquire transform
	// destination buffer.
	buf := bufpool.Get()
	defer bufpool.Put(buf)

	// Prepare a copy of our globally
	// configured string word segmenter.
	s := segmenter
	s.SetText(text)

	// Split words.
	for s.Next() {

		// Get token start / end positions.
		start, end := s.Start(), s.End()

		// Get token bytes.
		token := s.Text()
		btoken := byteutil.S2B(token)

		// Perform any text normalization on the segment token.
		buf.B, _, _ = transform.Append(normalize, buf.B[:0], btoken)

		// Pass the normalized token to FTS5 callback.
		err := xToken(0, buf.String(), start, end)
		if err != nil {
			return err
		}
	}

	return s.Err()
}

// text case fold normalizer.
var normalize = cases.Fold()

// Unicode® Standard Annex #29 word splitter.
// https://www.unicode.org/reports/tr29
//
// setup to strip useless single-char tokens and maintain hashtags.
//
// TODO: by either forking this package, or creating our own package
// using their internal unicode range table generators we can gain
// more granular control over token segmenting and gain some small
// performance optimizations when it comes to filtering.
var segmenter = func() iterators.StringSegmenter {
	s := words.StringSegmenter{&iterators.StringSegmenter{}}

	// Set joiners, this also sets
	// internal word split func.
	s.Joiners(&words.Joiners{

		// Permit hashtags.
		Leading: []rune{'#'},
	})

	// Set custom filter to remove
	// useless single-char tokens.
	s.Filter(func(b []byte) bool {

		// Decode first rune in token.
		r, sz := utf8.DecodeRune(b)
		if sz == len(b) {

			// For single characters,
			// only allow letter / num.
			return unicode.IsLetter(r) ||
				unicode.IsNumber(r)
		}

		// Permit.
		return true
	})

	// Return VALUE of contained
	// pointer with our config.
	return (*s.StringSegmenter)
}()

// zeroHTML will zero data contained within enclosed
// less-than-greater-than braces, i.e. <>, from input.
// this drops the HTML while maintaining text indices.
func zeroHTML(text string) string {

	// Acquire text buffer.
	var buf *byteutil.Buffer
	if len(text) > bufsz {

		// Too large, allocate new.
		buf = new(byteutil.Buffer)
		buf.B = make([]byte, len(text))
	} else {

		// Acquire from pool.
		buf = bufpool.Get()
	}

	// Fill buffer with text.
	buf.B = buf.B[0:len(text)]
	_ = copy(buf.B, text)

	// Our more complicated text is stored as HTML,
	// so drop all text (but maintaining indices for
	// the xToken() callback) within any HTML tags.
	for i, tagIdx := 0, -1; i < len(buf.B); {

		// Read next rune in source buffer text.
		r, sz := utf8.DecodeRune(buf.B[i:])
		switch r {
		case '<':
			tagIdx = i
		case '>':
			if tagIdx >= 0 {
				// Drop between tag start + end.
				clear(buf.B[tagIdx : i+sz])
				tagIdx = -1
			}
		}

		// Increment.
		i += sz
	}

	// Replace text with
	// our cleaned copy.
	text = string(buf.B)

	// Free buffer.
	bufpool.Put(buf)

	return text
}
