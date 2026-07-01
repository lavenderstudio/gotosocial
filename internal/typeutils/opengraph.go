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

package typeutils

import (
	"math"
	"regexp"
	"strconv"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/text"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

// OpenGraphBase returns an *ogMeta suitable for serving at the base root of an
// instance. It also serves as a foundation for building account / status ogMeta.
func OpenGraphBase(instance *apimodel.InstanceV1) *apimodel.OGMeta {

	// Take first
	// lang as locale.
	var locale string
	if len(instance.Languages) > 0 {
		locale = instance.Languages[0]
	}

	og := &apimodel.OGMeta{
		Title:       text.StripHTMLFromText(instance.Title) + " - GoToSocial",
		Type:        "website",
		Locale:      locale,
		URL:         instance.URI,
		SiteName:    instance.AccountDomain,
		Description: toDescription(instance.ShortDescription),
		Media: []apimodel.OGMedia{
			{
				OGType:   "image",
				URL:      instance.Thumbnail,
				Alt:      instance.ThumbnailDescription,
				MIMEType: instance.ThumbnailType,
			},
		},
	}

	return og
}

// OGAccount builds an ogMeta struct for the given account. It's suitable for serving at account profile pages.
func OpenGraphAccount(instance *apimodel.InstanceV1, acct *apimodel.WebAccount) *apimodel.OGMeta {

	// Set title to something like
	// "Display Name (@username@account.domain)"
	accountdomain := instance.AccountDomain
	title := AccountTitle(acct, accountdomain)

	// Take first
	// lang as locale.
	var locale string
	if len(instance.Languages) > 0 {
		locale = instance.Languages[0]
	}

	// Create description
	// from note (if set).
	var description string
	if acct.Note != "" {
		description = toDescription(acct.Note)
	} else {
		const emptyDesc = "This GoToSocial user hasn't written a bio yet!"
		description = emptyDesc
	}

	// Parse image from
	// account avatar (if set).
	media := []apimodel.OGMedia{ogImgForAcct(acct)}

	// ProfileUsername in format `someone@example.org`.
	profileUsername := acct.Username + "@" + accountdomain

	return &apimodel.OGMeta{
		Title:           title,
		Type:            "profile",
		Locale:          locale,
		URL:             acct.URL,
		SiteName:        accountdomain,
		Description:     truncate(description),
		Media:           media,
		ProfileUsername: profileUsername,
	}
}

// util funct to return OGImage using account.
func ogImgForAcct(account *apimodel.WebAccount) apimodel.OGMedia {
	ogMedia := apimodel.OGMedia{
		OGType: "image",
		URL:    account.Avatar,
		Alt:    "Avatar for " + account.Username,
	}

	if desc := account.AvatarDescription; desc != "" {
		ogMedia.Alt += ": " + desc
	}

	// Check if the account
	// has an avatar set
	// that's not the default.
	a := account.AvatarAttachment
	if a == nil {
		// Nothing
		// left to do.
		return ogMedia
	}

	// Set the MIME type.
	ogMedia.MIMEType = a.MIMEType

	// Check width + height.
	width := a.Meta.Original.Width
	height := a.Meta.Original.Height

	// Find longest side.
	longest := max(
		width,
		height,
	)

	// Max width or length should
	// be 400 or this will look
	// very silly in previews.
	const max = 400
	if longest > max {
		multiplier := float32(max) / float32(longest)
		width = int(math.Round(float64(multiplier * float32(width))))
		height = int(math.Round(float64(multiplier * float32(height))))
	}

	ogMedia.Width = strconv.Itoa(width)
	ogMedia.Height = strconv.Itoa(height)
	return ogMedia
}

// OGStatus builds an ogMeta struct for
// the given status by the given account.
// It's suitable for serving at thread pages.
func OpenGraphStatus(
	instance *apimodel.InstanceV1,
	acct *apimodel.WebAccount,
	status *apimodel.WebStatus,
) *apimodel.OGMeta {
	// Set title to something like
	// "Display Name (@username@account.domain)"
	accountdomain := instance.AccountDomain
	title := AccountTitle(acct, accountdomain)

	// Take first
	// lang as locale.
	var locale string
	if len(instance.Languages) > 0 {
		locale = instance.Languages[0]
	}

	// Derive description based on
	// sensitivity + media attachments.
	var (
		description string
		attachLen   = len(status.MediaAttachments)
		attachSet   = attachLen != 0
		cwSet       = status.SpoilerContent != ""
		contentSet  = status.Content != ""
	)

	switch {

	// If content warning is set then this
	// is a sensitive post by default and
	// we should not use the post content
	// at all in the description.
	case cwSet:
		content := toDescription(status.SpoilerContent)
		if attachSet {
			description = "Sensitive content [" + mediaCount(attachLen) + "]" + ": " + content
		} else {
			description = "Sensitive content: " + content
		}

	// There's no content warning set but
	// the status is marked sensitive and
	// it has text content. We can use the
	// status content in the description
	// but warn that it's sensitive.
	case status.Sensitive && contentSet:
		content := toDescription(status.Content)
		if attachSet {
			description = "Sensitive content [" + mediaCount(attachLen) + "]" + ": " + content
		} else {
			description = "Sensitive content: " + content
		}

	// There's no content warning set
	// and no text content set, but
	// there are sensitive attachments.
	case status.Sensitive && attachSet:
		description = "Sensitive media [" + mediaCount(attachLen) + "]"

	// Status isn't sensitive and there's
	// no content warning set, use the
	// post content in the description.
	case !status.Sensitive && contentSet:
		content := toDescription(status.Content)
		if attachSet {
			description = "[" + mediaCount(attachLen) + "] " + content
		} else {
			description = content
		}

	// Status isn't sensitive and there's
	// no content warning or content set.
	case !status.Sensitive && !contentSet:
		if attachSet {
			description = mediaCount(attachLen)
		} else {
			description = "Post by " + title
		}

	// Fall back to
	// account title.
	default:
		description = title
	}

	// Prepare status media.
	var (
		media                    []apimodel.OGMedia
		twitterSummaryLargeImage string
		twitterImageAlt          string
	)

	// Only append media to
	// preview if not sensitive.
	if !status.Sensitive {
		for _, a := range status.MediaAttachments {
			if a.Type == "unknown" {
				// Skip unknown.
				continue
			}

			// Start building entry
			// with common media tags.
			desc := util.PtrOrZero(a.Description)
			ogMedia := apimodel.OGMedia{
				URL:      *a.URL,
				MIMEType: a.MIMEType,
				Alt:      desc,
			}

			// Add further tags
			// depending on type.
			switch a.Type {

			case "image":
				ogMedia.OGType = "image"
				ogMedia.Width = strconv.Itoa(a.Meta.Original.Width)
				ogMedia.Height = strconv.Itoa(a.Meta.Original.Height)

				// If this image is the only piece of media,
				// set TwitterSummaryLargeImage to indicate
				// that a large image summary is preferred.
				if attachLen == 1 {
					twitterSummaryLargeImage = *a.URL
					twitterImageAlt = desc
				}

			case "audio":
				ogMedia.OGType = "audio"

			case "video", "gifv":
				ogMedia.OGType = "video"
				ogMedia.Width = strconv.Itoa(a.Meta.Original.Width)
				ogMedia.Height = strconv.Itoa(a.Meta.Original.Height)
			}

			// Add this to our gathered entries.
			media = append(media, ogMedia)

			// Include static/thumb for non-visual files
			// (eg., audios) if they have a preview url set.
			if a.Type != "image" && a.PreviewURL != nil {
				media = append(
					media,
					apimodel.OGMedia{
						OGType:   "image",
						URL:      *a.PreviewURL,
						MIMEType: a.PreviewMIMEType,
						Width:    strconv.Itoa(a.Meta.Small.Width),
						Height:   strconv.Itoa(a.Meta.Small.Height),
						Alt:      util.PtrOrZero(a.Description),
					},
				)
			}
		}
	}

	// ProfileUsername in format `someone@example.org`.
	profileUsername := acct.Username + "@" + accountdomain

	return &apimodel.OGMeta{
		Title:                    title,
		Type:                     "article",
		Locale:                   locale,
		URL:                      status.URL,
		SiteName:                 accountdomain,
		Description:              truncate(description),
		Media:                    media,
		ArticlePublisher:         status.Account.URL,
		ArticleAuthor:            status.Account.URL,
		ArticlePublishedTime:     status.CreatedAt,
		ArticleModifiedTime:      util.PtrOrValue(status.EditedAt, status.CreatedAt),
		ProfileUsername:          profileUsername,
		TwitterSummaryLargeImage: twitterSummaryLargeImage,
		TwitterImageAlt:          twitterImageAlt,
	}
}

// AccountTitle parses a page title
// from account and accountDomain.
func AccountTitle(
	account *apimodel.WebAccount,
	accountDomain string,
) string {
	var displayName string
	if account.DisplayName != "" {
		displayName = account.DisplayName
	} else {
		displayName = account.Username
	}
	nameString := "@" + account.Username + "@" + accountDomain
	return displayName + " (" + nameString + ")"
}

// Finds any links unnested
// by text.ParseHTMLToPlain.
var unnestedURLsRegexp = regexp.MustCompile(`(?U) <(?:http|https):\/\/.+\..+>`)

// toDescription converts given HTML string to
// an appropriate string to use as "description"
// content inside an opengraph <meta> tag.
func toDescription(html string) string {
	// Parse html string to plaintext.
	plain := text.ParseHTMLToPlain(html)

	// Remove any unnested URLs as they look ugly
	// when rendered inside an opengraph description.
	//
	// Eg., replace
	// 		`#boobs <https://example.org/tags/boobs>`
	// with just
	// 		`#boobs`.
	plain = unnestedURLsRegexp.ReplaceAllString(plain, "")

	// Truncate to 2000 chars,
	// anything longer than
	// that is a bloody essay.
	return truncate(plain)
}

// truncate trims string
// to maximum 2000 runes.
func truncate(s string) string {
	const truncateLen = 2000

	r := []rune(s)
	if len(r) < truncateLen {
		// No need
		// to trim.
		return s
	}

	return string(r[:truncateLen-3]) + "…"
}

// mediaCount returns a
// neat media count string.
func mediaCount(attachLen int) string {
	switch attachLen {
	case 0:
		return ""
	case 1:
		return "1 media attachment"
	default:
		return strconv.FormatInt(int64(attachLen), 10) + " media attachments"
	}
}
