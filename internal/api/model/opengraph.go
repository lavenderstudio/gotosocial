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

package model

// OGMeta represents supported OpenGraph Meta tags
//
// see eg https://ogp.me/
type OGMeta struct {
	/* Vanilla og tags */

	Title       string // og:title
	Type        string // og:type
	Locale      string // og:locale
	URL         string // og:url
	SiteName    string // og:site_name
	Description string // og:description

	// Zero or more media entries of type image,
	// video, or audio (https://ogp.me/#array).
	Media []OGMedia

	/* Article tags. */

	ArticlePublisher     string // article:publisher
	ArticleAuthor        string // article:author
	ArticleModifiedTime  string // article:modified_time
	ArticlePublishedTime string // article:published_time

	/* Profile tags. */

	ProfileUsername string // profile:username

	/*
		Twitter card stuff
		https://developer.twitter.com/en/docs/twitter-for-websites/cards/overview/abouts-cards
	*/

	// Set to media URL for media posts.
	TwitterSummaryLargeImage string
	TwitterImageAlt          string
}

// OGMedia represents one OpenGraph media
// entry of type image, video, or audio.
type OGMedia struct {
	OGType   string // image/video/audio
	URL      string // og:${type}
	MIMEType string // og:${type}:type
	Width    string // og:${type}:width
	Height   string // og:${type}:height
	Alt      string // og:${type}:alt
}
