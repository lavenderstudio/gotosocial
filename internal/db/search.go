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

package db

import (
	"context"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

type RelationType int

const (
	RelationAny RelationType = iota
	RelationFollowing
	RelationFollowers
)

type LocationType int

const (
	LocationAny LocationType = iota
	LocationRemote
	LocationLocal
	LocationNone
)

type OnlyBool int

const (
	AnyValue OnlyBool = iota
	OnlyFalse
	OnlyTrue
)

type CommonSearchParams struct {
	// Account ID of user
	// making the request.
	RequesterID string

	// Relation of returned
	// results to requester.
	Relationship RelationType

	// Location of results.
	Location LocationType
}

type AccountSearchParams struct {
	CommonSearchParams

	// Account username OR
	// domain FTS query string.
	UsernameOrDomainQuery string

	// Account username
	// FTS query string.
	UsernameQuery string

	// Account domain
	// FTS query string.
	DomainQuery string

	// Accounte note
	// FTS query string.
	NoteQuery string
}

type StatusSearchParams struct {
	CommonSearchParams

	// Only return statuses
	// authored by account.
	AuthorID string

	// Only return statuses
	// that are replies.
	Reply OnlyBool

	// Only return statuses
	// that are sensitive.
	Sensitive OnlyBool

	// Only return statuses
	// that have attachments.
	Media OnlyBool

	// Only return statuses
	// that contain polls.
	Poll OnlyBool

	// Status content OR
	// warning FTS query string.
	ContentOrWarningQuery string

	// Status content
	// FTS query string.
	ContentQuery string

	// Status content warning
	// FTS query string.
	ContentWarningQuery string
}

type Search interface {
	// SearchForAccounts performs an
	// account search using given args,
	// query strings perform FTS lookup.
	SearchForAccounts(
		ctx context.Context,

		// search parameters.
		params AccountSearchParams,

		// paging params.
		page *paging.Page,
	) (
		[]*gtsmodel.Account,
		error,
	)

	// SearchForStatuses performs a
	// status search using given args,
	// query strings perform FTS lookup.
	SearchForStatuses(
		ctx context.Context,

		// search parameters.
		params StatusSearchParams,

		// paging params.
		page *paging.Page,
	) (
		[]*gtsmodel.Status,
		error,
	)

	// SearchForTags performs a
	// tag search using given args,
	// query strings perform FTS lookup.
	SearchForTags(
		ctx context.Context,

		// tag name
		// query string.
		nameQuery string,

		// paging params.
		page *paging.Page,
	) (
		[]*gtsmodel.Tag,
		error,
	)
}
