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

package relaypush

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"slices"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	gtsmodel "code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/processing/common"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
)

type Processor struct {
	c *common.Processor

	state     *state.State
	converter *typeutils.Converter
}

func New(
	common *common.Processor,
	state *state.State,
	converter *typeutils.Converter,
) Processor {
	return Processor{
		c:         common,
		state:     state,
		converter: converter,
	}
}

func (p *Processor) RelayPushesGet(ctx context.Context, accountID string) (*apimodel.PageableResponse, gtserror.WithCode) {
	// Get all relay pushes belonging to this account.
	relayPushes, err := p.state.DB.GetRelayPushesForAccountID(ctx, accountID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting relay pushes: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	l := len(relayPushes)
	if l == 0 {
		return paging.EmptyResponse(), nil
	}

	items := make([]any, 0, l)
	for _, relayPush := range relayPushes {
		item, errWithCode := p.c.APIRelayConnection(ctx, relayPush)
		if errWithCode != nil {
			return nil, errWithCode
		}
		items = append(items, item)
	}

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/api/v1/relay_pushes",
	}), nil
}

func (p *Processor) getRelayPushForAccount(
	ctx context.Context,
	id string,
	accountID string,
) (*gtsmodel.RelayPush, gtserror.WithCode) {
	// Get barebones push from the db.
	relayPush, err := p.state.DB.GetRelayPushByID(
		gtscontext.SetBarebones(ctx),
		id,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting relay push: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if relayPush == nil {
		// The relay push doesn't exist in the db.
		err := gtserror.New("relay push not found")
		return nil, gtserror.NewErrorNotFound(err)
	}

	if relayPush.AccountID != accountID {
		// This push belongs to someone else.
		err := gtserror.New("relay push does not belong to this account")
		return nil, gtserror.NewErrorNotFound(err)
	}

	return relayPush, nil
}

func (p *Processor) RelayPushGet(
	ctx context.Context,
	auth *apiutil.Auth,
	id string,
) (*apimodel.RelayConnection, gtserror.WithCode) {
	relayPush, errWithCode := p.getRelayPushForAccount(ctx,
		id, auth.Account.ID,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return p.c.APIRelayConnection(ctx, relayPush)
}

func (p *Processor) RelayPushCreate(
	ctx context.Context,
	auth *apiutil.Auth,
	relayActorURI *url.URL,
	public bool,
	unlisted bool,
	matchByDefault bool,
	ignoreSensitive bool,
	ignoreMedia bool,
	ignoreReplies bool,
) (*apimodel.RelayConnection, gtserror.WithCode) {
	// Fetch relay actor.
	relayActor, errWithCode := p.c.DereferenceRelayActorURI(
		ctx,
		auth.Account.Username,
		relayActorURI,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Populate flags.
	var flags gtsmodel.RelayFlags
	flags.SetPublic(public)
	flags.SetUnlisted(unlisted)
	flags.SetMatchByDefault(matchByDefault)
	flags.SetIgnoreSensitive(ignoreSensitive)
	flags.SetIgnoreMedia(ignoreMedia)
	flags.SetIgnoreReplies(ignoreReplies)

	// Instantiate the push.
	relayPush := &gtsmodel.RelayPush{
		ID:            id.NewULID(),
		AccountID:     auth.Account.ID,
		RelayActorURI: relayActor.URI,
		Flags:         flags,
	}

	// Store in the DB.
	if err := p.state.DB.PutRelayPush(ctx, relayPush); err != nil {
		err := gtserror.Newf("db error storing relay push: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Try to get our instance account to follow
	// the relay actor (most relays require this).
	if err := p.c.FollowRelayActor(ctx, relayActor); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Return API model.
	return p.c.APIRelayConnection(ctx, relayPush)
}

func (p *Processor) RelayPushUpdate(
	ctx context.Context,
	authed *apiutil.Auth,
	id string,
	public *bool,
	unlisted *bool,
	matchByDefault *bool,
	ignoreSensitive *bool,
	ignoreMedia *bool,
	ignoreReplies *bool,
) (*apimodel.RelayConnection, gtserror.WithCode) {
	// Get relay push from the db.
	relayPush, errWithCode := p.getRelayPushForAccount(ctx, id, authed.Account.ID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Update flags using provided fields.
	if public != nil {
		relayPush.Flags.SetPublic(*public)
	}
	if unlisted != nil {
		relayPush.Flags.SetUnlisted(*unlisted)
	}
	if matchByDefault != nil {
		relayPush.Flags.SetMatchByDefault(*matchByDefault)
	}
	if ignoreSensitive != nil {
		relayPush.Flags.SetIgnoreSensitive(*ignoreSensitive)
	}
	if ignoreMedia != nil {
		relayPush.Flags.SetIgnoreMedia(*ignoreMedia)
	}
	if ignoreReplies != nil {
		relayPush.Flags.SetIgnoreReplies(*ignoreReplies)
	}

	// Update flags field in the DB.
	if err := p.state.DB.UpdateRelayPush(ctx, relayPush, "flags"); err != nil {
		err := gtserror.Newf("db error updating relay push: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Return API model.
	return p.c.APIRelayConnection(ctx, relayPush)
}

func (p *Processor) RelayPushDelete(
	ctx context.Context,
	authed *apiutil.Auth,
	id string,
) (*apimodel.RelayConnection, gtserror.WithCode) {
	// Get relay push from the db.
	relayPush, errWithCode := p.getRelayPushForAccount(ctx,
		id, authed.Account.ID,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Prepare response already before deletion.
	resp, errWithCode := p.c.APIRelayConnection(ctx, relayPush)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Delete relay push.
	if err := p.state.DB.DeleteRelayPush(ctx, relayPush); err != nil {
		err := gtserror.Newf("db error deleting relay push: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return resp, nil
}

func (p *Processor) MatcherCreate(
	ctx context.Context,
	authed *apiutil.Auth,
	relayPushID string,
	keyword string,
	wholeWord bool,
	exclude bool,
) (*apimodel.RelayConnection, gtserror.WithCode) {
	// Get relay push from the db.
	relayPush, errWithCode := p.getRelayPushForAccount(ctx,
		relayPushID, authed.Account.ID,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Populate flags.
	var flags gtsmodel.RelayMatcherFlags
	flags.SetWholeWord(wholeWord)
	flags.SetExclude(exclude)

	// Instantiate matcher.
	matcher := &gtsmodel.RelayMatcher{
		ID:      id.NewULID(),
		RelayID: relayPushID,
		Flags:   flags,
		Keyword: keyword,
	}

	// Ensure matcher can be compiled.
	if err := matcher.Compile(); err != nil {
		err := gtserror.Newf("matcher could not be compiled: %w", err)
		return nil, gtserror.NewErrorUnprocessableEntity(err, err.Error())
	}

	// Store it.
	switch err := p.state.DB.PutRelayMatcher(ctx, matcher); {
	case err == nil:
		// no issue

	case errors.Is(err, db.ErrAlreadyExists):
		const text = "duplicate keyword"
		return nil, gtserror.NewWithCode(http.StatusConflict, text)

	default:
		err := gtserror.Newf("db error inserting matcher: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Update the relay push to include this matcher.
	relayPush.MatcherIDs = append(relayPush.MatcherIDs, matcher.ID)
	if err := p.state.DB.UpdateRelayPush(ctx,
		relayPush,
		"matchers",
	); err != nil {
		err := gtserror.Newf("db error updating relay push: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Return API model of parent relay connection.
	return p.c.APIRelayConnection(ctx, relayPush)
}

func (p *Processor) MatcherDelete(
	ctx context.Context,
	authed *apiutil.Auth,
	relayPushID string,
	matcherID string,
) (*apimodel.RelayConnection, gtserror.WithCode) {
	// Get push from the db; this call also
	// ensures we have permission to delete it.
	relayPush, errWithCode := p.getRelayPushForAccount(ctx,
		relayPushID, authed.Account.ID,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Make sure the matcher exists in the db.
	_, errWithCode = p.c.GetRelayMatcher(ctx, matcherID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Delete the matcher.
	if err := p.state.DB.DeleteRelayMatcher(ctx, matcherID); err != nil {
		err := gtserror.Newf("db error deleting matcher: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Update the relay push to
	// remove this matcher before returning.
	relayPush.MatcherIDs = slices.DeleteFunc(
		relayPush.MatcherIDs,
		func(mID string) bool {
			return mID == matcherID
		},
	)
	if err := p.state.DB.UpdateRelayPush(ctx,
		relayPush,
		"matchers",
	); err != nil {
		err := gtserror.Newf("db error updating relay push: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Return API model of parent relay connection.
	return p.c.APIRelayConnection(ctx, relayPush)
}

func (p *Processor) MatcherUpdate(
	ctx context.Context,
	authed *apiutil.Auth,
	relayPushID string,
	matcherID string,
	keyword *string,
	wholeWord *bool,
	exclude *bool,
) (*apimodel.RelayConnection, gtserror.WithCode) {
	// Get push from the db; this call also
	// ensures we have permission to update it.
	relayPush, errWithCode := p.getRelayPushForAccount(ctx,
		relayPushID, authed.Account.ID,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Get the matcher.
	matcher, errWithCode := p.c.GetRelayMatcher(ctx, matcherID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Update the matcher.
	errWithCode = p.c.UpdateRelayMatcher(ctx,
		matcher,
		keyword,
		wholeWord,
		exclude,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Return API model of parent relay connection.
	return p.c.APIRelayConnection(ctx, relayPush)
}
