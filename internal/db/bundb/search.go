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

package bundb

import (
	"context"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type searchDB struct {
	db    *bun.DB
	state *state.State
}

// Query example (SQLite):
//
// TODO
func (s *searchDB) SearchForAccounts(
	ctx context.Context,
	params db.AccountSearchParams,
	page *paging.Page,
) (
	[]*gtsmodel.Account,
	error,
) {
	// Select account IDs.
	q := s.db.NewSelect()
	q = q.Table("accounts")
	q = q.Column("accounts.id")

	// Apply relationship params.
	switch params.Relationship {
	case db.RelationFollowers:
		q = q.Join("JOIN ? ON ? = ?", bun.Ident("follows"), bun.Ident("follows.target_account_id"), bun.Ident("accounts.id"))
		q = q.Where("? = ?", bun.Ident("follows.account_id"), params.RequesterID)

	case db.RelationFollowing:
		q = q.Join("JOIN ? ON ? = ?", bun.Ident("follows"), bun.Ident("follows.account_id"), bun.Ident("accounts.id"))
		q = q.Where("? = ?", bun.Ident("follows.target_account_id"), params.RequesterID)
	}

	// Apply location params.
	switch params.Location {
	case db.LocationRemote:
		q = q.Where("? IS NOT NULL", bun.Ident("accounts.domain"))

	case db.LocationLocal:
		q = q.Where("? IS NULL", bun.Ident("accounts.domain"))
	}

	// Apply account username/domain FTS match.
	q = db.AccountUsernameDomainFts.Match2(q,
		params.UsernameQuery,
		params.DomainQuery,
		params.UsernameOrDomainQuery,
	)

	// Apply account note FTS match.
	q = db.AccountNoteFts.Match1(q,
		params.NoteQuery,
	)

	// Set query parameters according to given page and scan.
	ids, err := selectPageOfIDs(ctx, q, "accounts.id", page)
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return nil, nil
	}

	return s.state.DB.GetAccountsByIDs(ctx, ids)
}

// Query example (SQLite):
//
// TODO
func (s *searchDB) SearchForStatuses(
	ctx context.Context,
	params db.StatusSearchParams,
	page *paging.Page,
) (
	[]*gtsmodel.Status,
	error,
) {
	if page == nil || page.Limit == 0 {
		panic("paging must be set")
	} else if params.RequesterID == "" {
		panic("requester must be set")
	}

	// Select status IDs.
	q := s.db.NewSelect()
	q = q.Table("statuses")
	q = q.Column("statuses.id")

	// Ignore status boosts.
	q = q.Where("? IS NULL",
		bun.Ident("statuses.boost_of_id"))

	// Status stubs should not be included.
	q = q.Where(db.BitNotSet("statuses.flags",
		gtsmodel.StatusFlagDeleted))

	var self bool
	switch params.AuthorID {

	// No author
	// specified.
	case "":

		// Apply relationship params.
		switch params.Relationship {
		case db.RelationFollowers:
			q = q.Join("JOIN ? ON ? = ?", bun.Ident("follows"), bun.Ident("follows.target_account_id"), bun.Ident("statuses.account_id"))
			q = q.Where("? = ?", bun.Ident("follows.account_id"), params.RequesterID)

		case db.RelationFollowing:
			q = q.Join("JOIN ? ON ? = ?", bun.Ident("follows"), bun.Ident("follows.account_id"), bun.Ident("statuses.accounts_id"))
			q = q.Where("? = ?", bun.Ident("follows.target_account_id"), params.RequesterID)
		}

		// Apply location params.
		switch params.Location {
		case db.LocationRemote:
			q = q.Where(db.BitNotSet("statuses.flags", gtsmodel.StatusFlagLocal))

		case db.LocationLocal:
			q = q.Where(db.BitIsSet("statuses.flags", gtsmodel.StatusFlagLocal))
		}

	// author == requester.
	case params.RequesterID:
		self = true
		fallthrough

	// Specified
	// author.
	default:

		// Return only statuses authored by given account ID.
		q = q.Where("? = ?", bun.Ident("statuses.account_id"),
			params.AuthorID)
	}

	if !self {
		// Show only public statuses of indexable authors,
		// (or any statuses by those the requester is following).
		//
		// There is further nuance required than this, but we leave
		// this up to the caller making use of the visibility filter.
		q = q.Join("JOIN ? ON ? = ?", bun.Ident("accounts"),
			bun.Ident("accounts.id"), bun.Ident("statuses.account_id"))
		q = db.WhereGroupQuery(q, " OR ",
			db.WhereGroup(" OR ",
				db.Where("? = ?", bun.Ident("statuses.account_id"), params.RequesterID),
				db.Where("EXISTS(SELECT 1 FROM ? WHERE (? = ?) AND (? = ?))",
					bun.Ident("follows"),
					bun.Ident("follows.account_id"), params.RequesterID,
					bun.Ident("follows.target_account_id"), bun.Ident("accounts.id"))),
			db.WhereGroup(" AND ",
				db.Where("? = ?", bun.Ident("statuses.visibility"), gtsmodel.VisibilityPublic),
				db.Where("? = TRUE", bun.Ident("accounts.indexable"))),
		)
	}

	// Apply media params.
	switch params.Media {
	case db.OnlyTrue:
		q = q.Where(db.WhereNotArrayIsNullOrEmpty(s.db, "statuses.attachments"))

	case db.OnlyFalse:
		q = q.Where(db.WhereArrayIsNullOrEmpty(s.db, "statuses.attachments"))
	}

	// Apply reply params.
	switch params.Reply {
	case db.OnlyTrue:
		q = q.Where("? IS NOT NULL", bun.Ident("statuses.in_reply_to_id"))

	case db.OnlyFalse:
		q = q.Where("? IS NULL", bun.Ident("statuses.in_reply_to_id"))
	}

	// Apply poll params.
	switch params.Poll {
	case db.OnlyTrue:
		q = q.Where("? IS NOT NULL", bun.Ident("statuses.poll_id"))

	case db.OnlyFalse:
		q = q.Where("? IS NULL", bun.Ident("statuses.poll_id"))
	}

	// Apply sensitive params.
	switch params.Sensitive {
	case db.OnlyTrue:
		q = q.Where(db.BitIsSet("statuses.flags", gtsmodel.StatusFlagSensitive))

	case db.OnlyFalse:
		q = q.Where(db.BitNotSet("statuses.flags", gtsmodel.StatusFlagSensitive))
	}

	// Apply status content/warning FTS match.
	q = db.StatusContentWarningFts.Match2(q,
		params.ContentQuery,
		params.ContentWarningQuery,
		params.ContentOrWarningQuery,
	)

	// Set query parameters according to given page and scan.
	ids, err := selectPageOfIDs(ctx, q, "statuses.id", page)
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return nil, nil
	}

	return s.state.DB.GetStatusesByIDs(ctx, ids)
}

// Query example (SQLite):
//
// TODO
func (s *searchDB) SearchForTags(
	ctx context.Context,
	nameQuery string,
	page *paging.Page,
) (
	[]*gtsmodel.Tag,
	error,
) {
	// Select tag IDs.
	q := s.db.NewSelect()
	q = q.Table("tags")
	q = q.Column("tags.id")

	// Apply tag name search FTS matching.
	q = db.TagNameFts.Match1(q, nameQuery)

	// Set query parameters according to page and scan.
	ids, err := selectPageOfIDs(ctx, q, "tags.id", page)
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return nil, nil
	}

	return s.state.DB.GetTags(ctx, ids)
}
