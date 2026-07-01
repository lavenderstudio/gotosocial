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

package typeutils_test

import (
	"testing"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/stretchr/testify/suite"
)

type OpenGraphTestSuite struct {
	suite.Suite
}

func (suite *OpenGraphTestSuite) TestWithAccountWithNote() {
	instance := &apimodel.InstanceV1{
		AccountDomain: "example.org",
		Languages:     []string{"en"},
		Thumbnail:     "https://example.org/instance-avatar.webp",
		ThumbnailType: "image/webp",
	}

	acct := &apimodel.Account{
		Acct:        "example_account",
		DisplayName: "example person!!",
		URL:         "https://example.org/@example_account",
		Note:        "<p>This is my profile, read it and weep!<br/>Weep then!</p>",
		Username:    "example_account",
		Avatar:      "https://example.org/avatar.jpg",
	}

	accountMeta := typeutils.OpenGraphAccount(instance, &apimodel.WebAccount{Account: acct})

	suite.EqualValues(apimodel.OGMeta{
		Title:       "example person!! (@example_account@example.org)",
		Type:        "profile",
		Locale:      "en",
		URL:         "https://example.org/@example_account",
		SiteName:    "example.org",
		Description: "This is my profile, read it and weep!\nWeep then!",
		Media: []apimodel.OGMedia{
			{
				OGType: "image",
				Alt:    "Avatar for example_account",
				URL:    "https://example.org/avatar.jpg",
			},
		},
		ArticlePublisher:     "",
		ArticleAuthor:        "",
		ArticleModifiedTime:  "",
		ArticlePublishedTime: "",
		ProfileUsername:      "example_account@example.org",
	}, *accountMeta)
}

func (suite *OpenGraphTestSuite) TestWithAccountNoNote() {
	instance := &apimodel.InstanceV1{
		AccountDomain: "example.org",
		Languages:     []string{"en"},
		Thumbnail:     "https://example.org/instance-avatar.webp",
		ThumbnailType: "image/webp",
	}

	acct := &apimodel.Account{
		Acct:        "example_account",
		DisplayName: "example person!!",
		URL:         "https://example.org/@example_account",
		Note:        "", // <- empty
		Username:    "example_account",
		Avatar:      "https://example.org/avatar.jpg",
	}

	accountMeta := typeutils.OpenGraphAccount(instance, &apimodel.WebAccount{Account: acct})

	suite.EqualValues(apimodel.OGMeta{
		Title:       "example person!! (@example_account@example.org)",
		Type:        "profile",
		Locale:      "en",
		URL:         "https://example.org/@example_account",
		SiteName:    "example.org",
		Description: "This GoToSocial user hasn't written a bio yet!",
		Media: []apimodel.OGMedia{
			{
				OGType: "image",
				Alt:    "Avatar for example_account",
				URL:    "https://example.org/avatar.jpg",
			},
		},
		ArticlePublisher:     "",
		ArticleAuthor:        "",
		ArticleModifiedTime:  "",
		ArticlePublishedTime: "",
		ProfileUsername:      "example_account@example.org",
	}, *accountMeta)
}

func (suite *OpenGraphTestSuite) TestWithStatus() {
	instance := &apimodel.InstanceV1{
		AccountDomain: "example.org",
		Languages:     []string{"en"},
		Thumbnail:     "https://example.org/instance-avatar.webp",
		ThumbnailType: "image/webp",
	}

	acct := &apimodel.Account{
		Acct:        "example_account",
		DisplayName: "example person!!",
		URL:         "https://example.org/@example_account",
		Username:    "example_account",
		Avatar:      "https://example.org/avatar.jpg",
	}

	apiStatus := &apimodel.Status{
		ID:        "12345",
		CreatedAt: "2025-01-18T00:00:00+00:00",
		EditedAt:  util.Ptr("2025-01-18T11:00:00+00:00"),
		URI:       "https://example.org/statuses/12345",
		URL:       "https://example.org/@example_account/12345",
		Content:   "<p><b>test status</b><p><p>here's another line</p>",
		Account:   acct,
	}

	status := &apimodel.WebStatus{
		Status:         apiStatus,
		SpoilerContent: "", // <- empty
		Account: &apimodel.WebAccount{
			Account:          acct,
			AvatarAttachment: nil,
			HeaderAttachment: nil,
			WebLayout:        gtsmodel.WebLayoutMicroblog.String(),
		},
	}

	statusMeta := typeutils.OpenGraphStatus(instance, status.Account, status)

	suite.EqualValues(apimodel.OGMeta{
		Title:                "example person!! (@example_account@example.org)",
		Type:                 "article",
		Locale:               "en",
		URL:                  "https://example.org/@example_account/12345",
		SiteName:             "example.org",
		Description:          "test status\n\nhere's another line",
		ArticlePublisher:     "https://example.org/@example_account",
		ArticleAuthor:        "https://example.org/@example_account",
		ArticleModifiedTime:  "2025-01-18T11:00:00+00:00",
		ArticlePublishedTime: "2025-01-18T00:00:00+00:00",
		ProfileUsername:      "example_account@example.org",
	}, *statusMeta)
}

func (suite *OpenGraphTestSuite) TestWithStatusWithImage() {
	instance := &apimodel.InstanceV1{
		AccountDomain: "example.org",
		Languages:     []string{"en"},
		Thumbnail:     "https://example.org/instance-avatar.webp",
		ThumbnailType: "image/webp",
	}

	acct := &apimodel.Account{
		Acct:        "example_account",
		DisplayName: "example person!!",
		URL:         "https://example.org/@example_account",
		Username:    "example_account",
		Avatar:      "https://example.org/avatar.jpg",
	}

	imageAttachment := &apimodel.Attachment{
		ID:         "00IMAGE00",
		Type:       "image",
		URL:        util.Ptr("https://example.org/@example_account/12345/example.png"),
		TextURL:    util.Ptr("https://example.org/@example_account/12345/example.png"),
		PreviewURL: util.Ptr("https://example.org/@example_account/12345/small/example.png"),
		Meta: &apimodel.MediaMeta{
			Original: apimodel.MediaDimensions{
				Width:  1920,
				Height: 1080,
			},
			Small: apimodel.MediaDimensions{
				Width:  320,
				Height: 240,
			},
		},
		Description: util.Ptr("an example image"),
		Blurhash:    util.Ptr("LKE3VIw}0KD%a2o{M|t7NFWps:t7"), // <- from testmodels
	}

	anotherImageAttachment := &apimodel.Attachment{
		ID:         "00IMAGE11",
		Type:       "image",
		URL:        util.Ptr("https://example.org/@example_account/12345/example2.png"),
		TextURL:    util.Ptr("https://example.org/@example_account/12345/example2.png"),
		PreviewURL: util.Ptr("https://example.org/@example_account/12345/small/example2.png"),
		Meta: &apimodel.MediaMeta{
			Original: apimodel.MediaDimensions{
				Width:  1000,
				Height: 1000,
			},
			Small: apimodel.MediaDimensions{
				Width:  200,
				Height: 200,
			},
		},
		Description: util.Ptr("another example image"),
	}

	apiStatus := &apimodel.Status{
		ID:               "12345",
		CreatedAt:        "2025-01-18T00:00:00+00:00",
		EditedAt:         util.Ptr("2025-01-18T11:00:00+00:00"),
		URL:              "https://example.org/@example_account/12345",
		Content:          "<p>test status <span class=\"h-card\"><a href=\"https://example.org/c/mutual_aid\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>mutual_aid@example.org</span></a></span> <a href=\"https://example.org/tags/MutualAidRequest\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>MutualAidRequest</span></a></p>",
		Account:          acct,
		MediaAttachments: []*apimodel.Attachment{imageAttachment, anotherImageAttachment},
	}

	webAttachment := &apimodel.WebAttachment{
		Attachment:       imageAttachment,
		Sensitive:        false,
		MIMEType:         "image/png",
		PreviewMIMEType:  "image/png",
		ParentStatusLink: "https://example.org/@example_account/12345",
	}

	anotherWebAttachment := &apimodel.WebAttachment{
		Attachment:       anotherImageAttachment,
		Sensitive:        false,
		MIMEType:         "image/png",
		PreviewMIMEType:  "image/png",
		ParentStatusLink: "https://example.org/@example_account/12345",
	}

	status := &apimodel.WebStatus{
		Status:           apiStatus,
		MediaAttachments: []*apimodel.WebAttachment{webAttachment, anotherWebAttachment},
		Account:          &apimodel.WebAccount{Account: acct},
	}

	statusMeta := typeutils.OpenGraphStatus(instance, status.Account, status)

	suite.EqualValues(apimodel.OGMeta{
		Title:       "example person!! (@example_account@example.org)",
		Type:        "article",
		Locale:      "en",
		URL:         "https://example.org/@example_account/12345",
		SiteName:    "example.org",
		Description: "[2 media attachments] test status @mutual_aid@example.org #MutualAidRequest",
		Media: []apimodel.OGMedia{
			{
				OGType:   "image",
				Alt:      "an example image",
				URL:      "https://example.org/@example_account/12345/example.png",
				MIMEType: "image/png",
				Width:    "1920",
				Height:   "1080",
			},
			{
				OGType:   "image",
				Alt:      "another example image",
				URL:      "https://example.org/@example_account/12345/example2.png",
				MIMEType: "image/png",
				Width:    "1000",
				Height:   "1000",
			},
		},
		ArticlePublisher:     "https://example.org/@example_account",
		ArticleAuthor:        "https://example.org/@example_account",
		ArticleModifiedTime:  "2025-01-18T11:00:00+00:00",
		ArticlePublishedTime: "2025-01-18T00:00:00+00:00",
		ProfileUsername:      "example_account@example.org",
	}, *statusMeta)
}

func TestOpenGraphTestSuite(t *testing.T) {
	suite.Run(t, &OpenGraphTestSuite{})
}
