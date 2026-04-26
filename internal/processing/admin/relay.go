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

package admin

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
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

func (p *Processor) RelaySubscriptionsGet(ctx context.Context) (*apimodel.PageableResponse, gtserror.WithCode) {
	// Get all relay subscriptions.
	relaySubscriptions, err := p.state.DB.GetRelaySubscriptions(ctx)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting relay subscriptions: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	l := len(relaySubscriptions)
	if l == 0 {
		return paging.EmptyResponse(), nil
	}

	items := make([]any, 0, l)
	for _, relaySubscription := range relaySubscriptions {
		item, errWithCode := p.c.APIRelayConnection(ctx, relaySubscription)
		if errWithCode != nil {
			return nil, errWithCode
		}
		items = append(items, item)
	}

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/api/v1/admin/relay_subscriptions",
	}), nil
}

func (p *Processor) getRelaySubscription(
	ctx context.Context,
	id string,
) (*gtsmodel.RelaySubscription, gtserror.WithCode) {
	// Get barebones subscription from the db.
	relaySubscription, err := p.state.DB.GetRelaySubscriptionByID(
		gtscontext.SetBarebones(ctx),
		id,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting relay subscription: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if relaySubscription == nil {
		// The relay subscription
		// doesn't exist in the db.
		err := gtserror.New("relay subscription not found")
		return nil, gtserror.NewErrorNotFound(err)
	}

	return relaySubscription, nil
}

func (p *Processor) RelaySubscriptionGet(
	ctx context.Context,
	id string,
) (*apimodel.RelayConnection, gtserror.WithCode) {
	relaySubscription, errWithCode := p.getRelaySubscription(ctx, id)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return p.c.APIRelayConnection(ctx, relaySubscription)
}

func (p *Processor) RelaySubscriptionCreate(
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

	// Instantiate the subscription.
	relaySubscription := &gtsmodel.RelaySubscription{
		ID:            id.NewULID(),
		AccountID:     auth.Account.ID,
		RelayActorURI: relayActor.URI,
		Flags:         flags,
	}

	// Store in the DB.
	if err := p.state.DB.PutRelaySubscription(ctx, relaySubscription); err != nil {
		err := gtserror.Newf("db error storing relay subscription: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Try to get our instance account to follow
	// the relay actor (most relays require this).
	if err := p.c.FollowRelayActor(ctx, relayActor); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Return API model.
	return p.c.APIRelayConnection(ctx, relaySubscription)
}

func (p *Processor) RelaySubscriptionUpdate(
	ctx context.Context,
	id string,
	public *bool,
	unlisted *bool,
	matchByDefault *bool,
	ignoreSensitive *bool,
	ignoreMedia *bool,
	ignoreReplies *bool,
) (*apimodel.RelayConnection, gtserror.WithCode) {
	// Get relay subscription from the db.
	relaySubscription, errWithCode := p.getRelaySubscription(ctx, id)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Update flags using provided fields.
	if public != nil {
		relaySubscription.Flags.SetPublic(*public)
	}
	if unlisted != nil {
		relaySubscription.Flags.SetUnlisted(*unlisted)
	}
	if matchByDefault != nil {
		relaySubscription.Flags.SetMatchByDefault(*matchByDefault)
	}
	if ignoreSensitive != nil {
		relaySubscription.Flags.SetIgnoreSensitive(*ignoreSensitive)
	}
	if ignoreMedia != nil {
		relaySubscription.Flags.SetIgnoreMedia(*ignoreMedia)
	}
	if ignoreReplies != nil {
		relaySubscription.Flags.SetIgnoreReplies(*ignoreReplies)
	}

	// Update flags field in the DB.
	if err := p.state.DB.UpdateRelaySubscription(ctx, relaySubscription, "flags"); err != nil {
		err := gtserror.Newf("db error updating relay subscription: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Return API model.
	return p.c.APIRelayConnection(ctx, relaySubscription)
}

func (p *Processor) RelaySubscriptionDelete(
	ctx context.Context,
	id string,
) (*apimodel.RelayConnection, gtserror.WithCode) {
	// Get relay subscription from the db.
	relaySubscription, errWithCode := p.getRelaySubscription(ctx, id)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Prepare response already before deletion.
	resp, errWithCode := p.c.APIRelayConnection(ctx, relaySubscription)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Delete relay subscription.
	if err := p.state.DB.DeleteRelaySubscription(ctx, relaySubscription); err != nil {
		err := gtserror.Newf("db error deleting relay subscription: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return resp, nil
}

func (p *Processor) RelaySubscriptionMatcherCreate(
	ctx context.Context,
	relayID string,
	keyword string,
	wholeWord bool,
	exclude bool,
) (*apimodel.RelayConnection, gtserror.WithCode) {
	// Get the (barebones) parent relay
	// subscription from the db first.
	relaySubscription, errWithCode := p.getRelaySubscription(ctx, relayID)
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
		RelayID: relayID,
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

	// Update the relay subscription to add this matcher.
	relaySubscription.MatcherIDs = append(relaySubscription.MatcherIDs, matcher.ID)
	if err := p.state.DB.UpdateRelaySubscription(ctx,
		relaySubscription,
		"matchers",
	); err != nil {
		err := gtserror.Newf("db error updating relay subscription: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Return API model of parent relay connection.
	return p.c.APIRelayConnection(ctx, relaySubscription)
}

func (p *Processor) RelaySubscriptionMatcherDelete(
	ctx context.Context,
	relaySubscriptionID string,
	matcherID string,
) (*apimodel.RelayConnection, gtserror.WithCode) {
	// Get the (barebones) parent relay
	// subscription from the db first.
	relaySubscription, errWithCode := p.getRelaySubscription(ctx, relaySubscriptionID)
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

	// Update the relay subscription to remove this matcher.
	relaySubscription.MatcherIDs = slices.DeleteFunc(
		relaySubscription.MatcherIDs,
		func(mID string) bool {
			return mID == matcherID
		},
	)
	if err := p.state.DB.UpdateRelaySubscription(ctx,
		relaySubscription,
		"matchers",
	); err != nil {
		err := gtserror.Newf("db error updating relay subscription: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Return API model of parent relay connection.
	return p.c.APIRelayConnection(ctx, relaySubscription)
}

func (p *Processor) RelaySubscriptionMatcherUpdate(
	ctx context.Context,
	relayID string,
	matcherID string,
	keyword *string,
	wholeWord *bool,
	exclude *bool,
) (*apimodel.RelayConnection, gtserror.WithCode) {
	// Get the (barebones) parent relay
	// subscription from the db first.
	relaySubscription, errWithCode := p.getRelaySubscription(ctx, relayID)
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
	return p.c.APIRelayConnection(ctx, relaySubscription)
}
