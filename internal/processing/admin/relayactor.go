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
	"fmt"
	"net/http"
	"slices"

	"code.superseriousbusiness.org/gopkg/log"
	"code.superseriousbusiness.org/gotosocial/internal/ap"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/messages"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/text"
	"code.superseriousbusiness.org/gotosocial/internal/uris"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"code.superseriousbusiness.org/gotosocial/internal/validate"
)

func (p *Processor) apiRelayActor(
	ctx context.Context,
	relayActor *gtsmodel.RelayActor,
) (*apimodel.RelayActor, gtserror.WithCode) {
	v, err := p.converter.RelayActorToAPIRelayActor(ctx, relayActor)
	if err != nil {
		err := gtserror.NewfAt(3, "error converting relay actor: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}
	return v, nil
}

func (p *Processor) getRelayActor(
	ctx context.Context,
	relayActorID string,
) (*gtsmodel.RelayActor, gtserror.WithCode) {
	// Get barebones actor from the db.
	relayActor, err := p.state.DB.GetRelayActorByID(
		gtscontext.SetBarebones(ctx),
		relayActorID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting relay actor: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if relayActor == nil {
		// The relay actor
		// doesn't exist in the db.
		err := gtserror.New("relay actor not found")
		return nil, gtserror.NewErrorNotFound(err)
	}

	return relayActor, nil
}

func (p *Processor) RelayActorCreate(
	ctx context.Context,
	auth *apiutil.Auth,
	form *apimodel.RelayActorCreateRequest,
) (*apimodel.RelayActor, gtserror.WithCode) {
	// Prefix input username with "relay."
	// to differentiate it from a normal user.
	username := uris.RelayUsernamePrefix + form.Username

	// Check username available.
	ok, err := p.state.DB.IsUsernameAvailable(ctx, username)
	if err != nil {
		err := gtserror.Newf("db error checking username availability: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}
	if !ok {
		err := fmt.Errorf("username %s is not available", form.Username)
		return nil, gtserror.NewErrorConflict(err, err.Error())
	}

	// Account URIs and keys and stuff.
	uris := uris.GenerateActorURIs(uris.RelaysPath, form.Username)
	privKey, pubKey, err := util.NewActorRSA()
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Instantiate the relay actor account.
	account := &gtsmodel.Account{
		ID:                           id.NewRandomULID(),
		Username:                     username,
		Locked:                       form.Locked,
		Discoverable:                 form.Discoverable,
		URI:                          uris.ActorURI,
		URL:                          uris.ActorURL,
		InboxURI:                     uris.InboxURI,
		OutboxURI:                    uris.OutboxURI,
		FollowingURI:                 uris.FollowingURI,
		FollowersURI:                 uris.FollowersURI,
		FeaturedCollectionURI:        uris.FeaturedCollectionURI,
		ActorType:                    gtsmodel.AccountActorTypeService,
		PrivateKey:                   privKey,
		PublicKey:                    pubKey,
		PublicKeyURI:                 uris.PublicKeyURI,
		HidesToPublicFromUnauthedWeb: util.Ptr(true),
		HidesCcPublicFromUnauthedWeb: util.Ptr(true),
	}

	// Normalize display
	// name from form.
	if form.DisplayName != nil {
		displayName := *form.DisplayName
		if err := validate.DisplayName(displayName); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}

		// HTML tags not allowed in display name.
		account.DisplayName = text.StripHTMLFromText(displayName)
	}

	// Normalize note
	// from form.
	if form.Note != nil {
		noteRaw := *form.Note
		if err := validate.Note(noteRaw); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
		account.NoteRaw = noteRaw
	}

	// Normalize fields
	// from form.
	if form.FieldsAttributes != nil {
		if err := p.c.UpdateFields(
			account,
			*form.FieldsAttributes,
		); err != nil {
			return nil, err
		}
	}

	// Process display name, note, fields,
	// and any concomitant emoji changes.
	//
	// Allow text/markdown annotation.
	p.c.ProcessAccountText(ctx,
		account,
		"text/markdown",
	)

	// Set avatar if provided.
	if form.Avatar != nil && form.Avatar.Size != 0 {
		if form.AvatarDescription != nil {
			desc := text.StripHTMLFromText(*form.AvatarDescription)
			form.AvatarDescription = &desc
		}

		avatarInfo, errWithCode := p.c.UpdateAvatar(ctx,
			account,
			form.Avatar,
			form.AvatarDescription,
		)
		if errWithCode != nil {
			return nil, errWithCode
		}
		account.AvatarMediaAttachmentID = avatarInfo.ID
		account.AvatarMediaAttachment = avatarInfo
	}

	// Set header if provided.
	if form.Header != nil && form.Header.Size != 0 {
		if form.HeaderDescription != nil {
			desc := text.StripHTMLFromText(*form.HeaderDescription)
			form.HeaderDescription = util.Ptr(desc)
		}

		headerInfo, errWithCode := p.c.UpdateHeader(ctx,
			account,
			form.Header,
			form.HeaderDescription,
		)
		if errWithCode != nil {
			return nil, errWithCode
		}
		account.HeaderMediaAttachmentID = headerInfo.ID
		account.HeaderMediaAttachment = headerInfo
	}

	// Store the account.
	switch err := p.state.DB.PutAccount(ctx, account); {
	case err == nil:
		// No problem.
	case errors.Is(err, db.ErrAlreadyExists):
		// Username conflict? We already checked
		// for this above but maybe a race.
		const text = "conflict in database when inserting account: is username already in use?"
		return nil, gtserror.NewErrorConflict(err, text)
	default:
		err := gtserror.Newf("db error inserting account: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Stub out stats for this actor.
	if err := p.state.DB.StubAccountStats(ctx, account); err != nil {
		err := gtserror.Newf("db error stubbing account stats: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Populate flags.
	var flags gtsmodel.RelayFlags
	flags.SetPublic(util.PtrOrValue(form.Public, true))            // default true
	flags.SetUnlisted(util.PtrOrZero(form.Unlisted))               // default false
	flags.SetMatchByDefault(util.PtrOrZero(form.MatchByDefault))   // default false
	flags.SetIgnoreSensitive(util.PtrOrZero(form.IgnoreSensitive)) // default false
	flags.SetIgnoreMedia(util.PtrOrZero(form.IgnoreMedia))         // default false
	flags.SetIgnoreReplies(util.PtrOrZero(form.IgnoreReplies))     // default false

	// Populate relay actor itself,
	// linking back to the account we
	// just created using the AP URI.
	relayActor := &gtsmodel.RelayActor{
		ID:                 id.NewULID(),
		CreatedByAccountID: auth.Account.ID,
		URI:                account.URI,
		ActorAccount:       account,
		Flags:              flags,
	}

	// Store the relay actor.
	if err := p.state.DB.PutRelayActor(ctx, relayActor); err != nil {
		// If there was an error, (try to) remove
		// the account we just inserted, otherwise
		// the username will never be freed up.
		if err := p.state.DB.DeleteAccount(ctx, account.ID); err != nil {
			log.Errorf(ctx, "error deleting account after failed PutRelayActor: %v", err)
		}

		// Now handle the error.
		if errors.Is(err, db.ErrAlreadyExists) {
			const text = "conflict in database when inserting relay actor: is URI already in use?"
			return nil, gtserror.NewErrorConflict(err, text)
		} else { //nolint
			err := gtserror.Newf("db error inserting account: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	return p.apiRelayActor(ctx, relayActor)
}

func (p *Processor) RelayActorsGet(ctx context.Context) ([]*apimodel.RelayActor, gtserror.WithCode) {
	relayActors, err := p.state.DB.GetRelayActors(ctx)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting relay actors: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	l := len(relayActors)
	if l == 0 {
		return make([]*apimodel.RelayActor, 0), nil
	}

	items := make([]*apimodel.RelayActor, 0, l)
	for _, actor := range relayActors {
		item, errWithCode := p.apiRelayActor(ctx, actor)
		if errWithCode != nil {
			return nil, errWithCode
		}
		items = append(items, item)
	}

	return items, nil
}

func (p *Processor) RelayActorGet(
	ctx context.Context,
	relayActorID string,
) (*apimodel.RelayActor, gtserror.WithCode) {
	relayActor, errWithCode := p.getRelayActor(ctx, relayActorID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return p.apiRelayActor(ctx, relayActor)
}

func (p *Processor) RelayActorUpdate(
	ctx context.Context,
	relayActorID string,
	form *apimodel.RelayActorUpdateRequest,
) (*apimodel.RelayActor, gtserror.WithCode) {
	relayActor, errWithCode := p.getRelayActor(ctx, relayActorID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Populate the relay actor.
	if err := p.state.DB.PopulateRelayActor(ctx, relayActor); err != nil {
		err := gtserror.Newf("db error populating relay actor: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Set flags,
	// if changed.
	flagsChanged := false
	if form.Public != nil {
		relayActor.Flags.SetPublic(*form.Public)
		flagsChanged = true
	}
	if form.Unlisted != nil {
		relayActor.Flags.SetUnlisted(*form.Unlisted)
		flagsChanged = true
	}
	if form.MatchByDefault != nil {
		relayActor.Flags.SetMatchByDefault(*form.MatchByDefault)
		flagsChanged = true
	}
	if form.IgnoreSensitive != nil {
		relayActor.Flags.SetIgnoreSensitive(*form.IgnoreSensitive)
		flagsChanged = true
	}
	if form.IgnoreMedia != nil {
		relayActor.Flags.SetIgnoreMedia(*form.IgnoreMedia)
		flagsChanged = true
	}
	if form.IgnoreReplies != nil {
		relayActor.Flags.SetIgnoreReplies(*form.IgnoreReplies)
		flagsChanged = true
	}

	// Update flags,
	// if changed.
	if flagsChanged {
		if err := p.state.DB.UpdateRelayActor(ctx,
			relayActor,
			"flags",
		); err != nil {
			err := gtserror.Newf("db error updating relay actor: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	// Use the common processor to update
	// the relay actor account, passing
	// only the fields we care about.
	if errWithCode := p.c.UpdateAccount(ctx,
		relayActor.ActorAccount,
		&apimodel.UpdateAccountRequest{
			Discoverable:         form.Discoverable,
			DisplayName:          form.DisplayName,
			Note:                 form.Note,
			Avatar:               form.Avatar,
			AvatarDescription:    form.AvatarDescription,
			Header:               form.Header,
			HeaderDescription:    form.HeaderDescription,
			Locked:               form.Locked,
			FieldsAttributes:     form.FieldsAttributes,
			JSONFieldsAttributes: form.JSONFieldsAttributes,
		},
		"text/markdown",
	); errWithCode != nil {
		return nil, errWithCode
	}

	return p.apiRelayActor(ctx, relayActor)
}

func (p *Processor) RelayActorDelete(
	ctx context.Context,
	auth *apiutil.Auth,
	relayActorID string,
) (*apimodel.RelayActor, gtserror.WithCode) {
	relayActor, errWithCode := p.getRelayActor(ctx, relayActorID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Prepare response before deletion.
	resp, errWithCode := p.apiRelayActor(ctx, relayActor)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Delete the relay actor entry.
	if err := p.state.DB.DeleteRelayActor(ctx, relayActor); err != nil {
		err := gtserror.Newf("db error deleting relay actor: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Pass a message through the client workers
	// to delete (stubbify) the relay actor account.
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		APActivityType: ap.ActivityDelete,
		APObjectType:   ap.ActorService,
		GTSModel:       relayActor.ActorAccount,
		Origin:         auth.Account,
		Target:         relayActor.ActorAccount,
	})

	return resp, nil
}

func (p *Processor) RelayActorMatcherCreate(
	ctx context.Context,
	relayActorID string,
	keyword string,
	wholeWord bool,
	exclude bool,
) (*apimodel.RelayActor, gtserror.WithCode) {
	// Get the (barebones) parent relay
	// actor from the db first.
	relayActor, errWithCode := p.getRelayActor(ctx, relayActorID)
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
		RelayID: relayActorID,
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

	// Update the relay actor to add this matcher.
	relayActor.MatcherIDs = append(relayActor.MatcherIDs, matcher.ID)
	if err := p.state.DB.UpdateRelayActor(ctx,
		relayActor,
		"matchers",
	); err != nil {
		err := gtserror.Newf("db error updating relay actor: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Return API model of parent relay actor.
	return p.apiRelayActor(ctx, relayActor)
}

func (p *Processor) RelayActorMatcherDelete(
	ctx context.Context,
	relayActorID string,
	matcherID string,
) (*apimodel.RelayActor, gtserror.WithCode) {
	// Get the (barebones) parent relay
	// actor from the db first.
	relayActor, errWithCode := p.getRelayActor(ctx, relayActorID)
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

	// Update the relay actor to remove this matcher.
	relayActor.MatcherIDs = slices.DeleteFunc(
		relayActor.MatcherIDs,
		func(mID string) bool {
			return mID == matcherID
		},
	)
	if err := p.state.DB.UpdateRelayActor(ctx,
		relayActor,
		"matchers",
	); err != nil {
		err := gtserror.Newf("db error updating relay actor: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Return API model of parent relay actor.
	return p.apiRelayActor(ctx, relayActor)
}

func (p *Processor) RelayActorMatcherUpdate(
	ctx context.Context,
	relayActorID string,
	matcherID string,
	keyword *string,
	wholeWord *bool,
	exclude *bool,
) (*apimodel.RelayActor, gtserror.WithCode) {
	// Get the (barebones) parent relay
	// actor from the db first.
	relayActor, errWithCode := p.getRelayActor(ctx, relayActorID)
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

	// Return API model of parent relay actor.
	return p.apiRelayActor(ctx, relayActor)
}

// Get barebones relay actor account
// from the db using relayActorID.
func (p *Processor) getRelayActorAccount(
	ctx context.Context,
	relayActorID string,
) (*gtsmodel.Account, gtserror.WithCode) {
	// Get the (barebones) parent relay
	// actor from the db first.
	relayActor, errWithCode := p.getRelayActor(ctx, relayActorID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Get the (barebones) relay actor account.
	account, err := p.state.DB.GetAccountByURI(
		gtscontext.SetBarebones(ctx),
		relayActor.URI,
	)
	if err != nil {
		err := gtserror.Newf("db error getting account: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return account, nil
}

// Get page of follow requests targeting relayActorID's account.
func (p *Processor) RelayActorFollowRequestsGet(
	ctx context.Context,
	relayActorID string,
	page *paging.Page,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	// Get the (barebones) relay actor account.
	account, errWithCode := p.getRelayActorAccount(ctx, relayActorID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Get follow requests for the relay actor account.
	path := "/api/v1/admin/relay_actors/" + relayActorID + "/follow_requests"
	return p.c.FollowRequestsGet(ctx, account, page, path)
}

// Accept a follow request from targetAcctID to relayActorID's account.
func (p *Processor) RelayActorFollowRequestAccept(
	ctx context.Context,
	relayActorID string,
	targetAcctID string,
) (*apimodel.Relationship, gtserror.WithCode) {
	// Get the (barebones) relay actor account.
	account, errWithCode := p.getRelayActorAccount(ctx, relayActorID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return p.c.FollowRequestAccept(ctx, account, targetAcctID)
}

// Reject a follow request from targetAcctID to relayActorID's account.
func (p *Processor) RelayActorFollowRequestReject(
	ctx context.Context,
	relayActorID string,
	targetAcctID string,
) (*apimodel.Relationship, gtserror.WithCode) {
	// Get the (barebones) relay actor account.
	account, errWithCode := p.getRelayActorAccount(ctx, relayActorID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return p.c.FollowRequestReject(ctx, account, targetAcctID)
}

// Get page of accounts that follow relayActorID's account.
func (p *Processor) RelayActorFollowersGet(
	ctx context.Context,
	relayActorID string,
	page *paging.Page,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	// Get the (barebones) relay actor account.
	account, errWithCode := p.getRelayActorAccount(ctx, relayActorID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Get followers, passing a nil requester account;
	// we can't manage relay actor followers properly
	// if we filter them (say, if they block the admin).
	path := "/api/v1/admin/relay_actors/" + relayActorID + "/followers"
	return p.c.FollowersGet(ctx, account, nil, page, path)
}

// Remove a follow from targetAcctID to relayActorID's account.
func (p *Processor) RelayActorFollowerRemove(
	ctx context.Context,
	relayActorID string,
	targetAcctID string,
) (*apimodel.Relationship, gtserror.WithCode) {
	// Get the (barebones) relay actor account.
	account, errWithCode := p.getRelayActorAccount(ctx, relayActorID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Get target account from the db.
	targetAccount, err := p.state.DB.GetAccountByID(ctx, targetAcctID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("db error getting account: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if targetAccount == nil {
		err := gtserror.New("account not found")
		return nil, gtserror.NewErrorNotFound(err)
	}

	return p.c.RemoveFromFollowers(ctx, account, targetAccount)
}

// Get page of blocks owned by relayActorID's account.
func (p *Processor) RelayActorBlocksGet(
	ctx context.Context,
	relayActorID string,
	page *paging.Page,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	// Get the (barebones) relay actor account.
	account, errWithCode := p.getRelayActorAccount(ctx, relayActorID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Return blocks without stubbing as we
	// want to be able to manage blocks properly.
	path := "/api/v1/admin/relay_actors/" + relayActorID + "/blocks"
	return p.c.BlocksGet(ctx,
		account,
		false, // stub
		page,
		path,
	)
}

// Block targetAcctID from relayActorID's account.
func (p *Processor) RelayActorBlock(
	ctx context.Context,
	relayActorID string,
	targetAcctID string,
) (*apimodel.Relationship, gtserror.WithCode) {
	// Get the (barebones) relay actor account.
	account, errWithCode := p.getRelayActorAccount(ctx, relayActorID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return p.c.BlockCreate(ctx, account, targetAcctID)
}

// Unblock targetAcctID from relayActorID's account.
func (p *Processor) RelayActorUnblock(
	ctx context.Context,
	relayActorID string,
	targetAcctID string,
) (*apimodel.Relationship, gtserror.WithCode) {
	// Get the (barebones) relay actor account.
	account, errWithCode := p.getRelayActorAccount(ctx, relayActorID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return p.c.BlockRemove(ctx, account, targetAcctID)
}
