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

package common

import (
	"context"
	"errors"
	"fmt"

	"code.superseriousbusiness.org/gopkg/log"
	"code.superseriousbusiness.org/gotosocial/internal/ap"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	gtsmodel "code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/messages"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/uris"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

// APIRelationship returns the API model relationship
// between account and target account, or an error.
func (p *Processor) APIRelationship(
	ctx context.Context,
	account *gtsmodel.Account,
	targetAccountID string,
) (*apimodel.Relationship, gtserror.WithCode) {
	if account == nil {
		const text = "unauthorized"
		err := errors.New(text)
		return nil, gtserror.NewErrorUnauthorized(err)
	}

	r, err := p.state.DB.GetRelationship(ctx, account.ID, targetAccountID)
	if err != nil {
		err := gtserror.Newf("db error getting relationship: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	relationship, err := p.converter.RelationshipToAPIRelationship(ctx, r)
	if err != nil {
		err := gtserror.Newf("db error getting relationship: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return relationship, nil
}

// FollowRequestsGet fetches a list of the accounts
// that have follow requests targeting account.
//
// The pagePath param should be set to the API path that's being
// used to call this function, eg "/api/v1/follow_requests".
func (p *Processor) FollowRequestsGet(
	ctx context.Context,
	account *gtsmodel.Account,
	page *paging.Page,
	pagePath string,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	// Fetch follow requests targeting the given account.
	followReqs, err := p.state.DB.GetAccountFollowRequests(ctx, account.ID, page)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Check for empty response.
	count := len(followReqs)
	if count == 0 {
		return paging.EmptyResponse(), nil
	}

	// Get the lowest and highest
	// ID values, used for paging.
	lo := followReqs[count-1].ID
	hi := followReqs[0].ID

	// Func to fetch follow source at index.
	getIdx := func(i int) *gtsmodel.Account {
		return followReqs[i].Account
	}

	// Get a filtered slice of public API account models.
	items := p.GetVisibleAPIAccountsPaged(ctx,
		account,
		getIdx,
		count,
	)

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  pagePath,
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
	}), nil
}

// FollowRequestAccept handles the accepting of
// a follow request from targetAcctID to account.
func (p *Processor) FollowRequestAccept(
	ctx context.Context,
	account *gtsmodel.Account,
	followerID string,
) (*apimodel.Relationship, gtserror.WithCode) {
	follow, err := p.state.DB.AcceptFollowRequest(ctx, followerID, account.ID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(err)
	}

	if follow.Account != nil {
		// Only enqueue work in the case we have a request creating account stored.
		// NOTE: due to how AcceptFollowRequest works, the inverse shouldn't be possible.
		p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
			APObjectType:   ap.ActivityFollow,
			APActivityType: ap.ActivityAccept,
			GTSModel:       follow,
			Origin:         follow.Account,
			Target:         follow.TargetAccount,
		})
	}

	return p.APIRelationship(ctx, account, followerID)
}

// FollowRequestReject handles the rejection of
// a follow request from followerID to account.
func (p *Processor) FollowRequestReject(
	ctx context.Context,
	account *gtsmodel.Account,
	followerID string,
) (*apimodel.Relationship, gtserror.WithCode) {
	followReq, err := p.state.DB.GetFollowRequest(ctx, followerID, account.ID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(err)
	}

	err = p.state.DB.RejectFollowRequest(ctx, followerID, account.ID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(err)
	}

	if followReq.Account != nil {
		// Only enqueue work in the case we have a request creating account stored.
		// NOTE: due to how GetFollowRequest works, the inverse shouldn't be possible.
		p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
			APObjectType:   ap.ActivityFollow,
			APActivityType: ap.ActivityReject,
			GTSModel:       followReq,
			Origin:         followReq.Account,
			Target:         followReq.TargetAccount,
		})
	}

	return p.APIRelationship(ctx, account, followerID)
}

// FollowersGet fetches a list of given account's followers,
// filtered for visibility on requester account (can be nil for no filter).
//
// The pagePath param should be set to the API path that's being
// used to call this function, eg "/api/v1/accounts/[id]/followers".
func (p *Processor) FollowersGet(
	ctx context.Context,
	account *gtsmodel.Account,
	requester *gtsmodel.Account,
	page *paging.Page,
	pagePath string,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	// Fetch follows that target account.
	follows, err := p.state.DB.GetAccountFollowers(ctx, account.ID, page)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("db error getting followers: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Check for empty response.
	count := len(follows)
	if count == 0 {
		return paging.EmptyResponse(), nil
	}

	// Get the lowest and highest
	// ID values, used for paging.
	lo := follows[count-1].ID
	hi := follows[0].ID

	// Func to fetch follow source at index.
	getIdx := func(i int) *gtsmodel.Account {
		return follows[i].Account
	}

	// Get a filtered slice of public API account models.
	items := p.GetVisibleAPIAccountsPaged(ctx,
		requester,
		getIdx,
		len(follows),
	)

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  pagePath,
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
	}), nil
}

// FollowingGet fetches a list of accounts followed by given account,
// filtered for visibility on requester account (can be nil for no filter).
//
// The pagePath param should be set to the API path that's being
// used to call this function, eg "/api/v1/accounts/[id]/following".
func (p *Processor) FollowingGet(
	ctx context.Context,
	account *gtsmodel.Account,
	requester *gtsmodel.Account,
	page *paging.Page,
	pagePath string,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	// Fetch follows owned by account.
	follows, err := p.state.DB.GetAccountFollows(ctx, account.ID, page)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("db error getting followers: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Check for empty response.
	count := len(follows)
	if count == 0 {
		return paging.EmptyResponse(), nil
	}

	// Get the lowest and highest
	// ID values, used for paging.
	lo := follows[count-1].ID
	hi := follows[0].ID

	// Func to fetch follow source at index.
	getIdx := func(i int) *gtsmodel.Account {
		return follows[i].TargetAccount
	}

	// Get a filtered slice of public API account models.
	items := p.GetVisibleAPIAccountsPaged(ctx,
		requester,
		getIdx,
		len(follows),
	)

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  pagePath,
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
	}), nil
}

// Unfollow removes any follows and follow requests
// from the db where account targets targetAcc.
//
// If `sideEffects` is true, then federation side effects
// (Undo Follow) will also be queued in the client worker.
func (p *Processor) Unfollow(
	ctx context.Context,
	account *gtsmodel.Account,
	targetAcc *gtsmodel.Account,
	sideEffects bool,
) error {
	// Get follow from account to target account.
	follow, err := p.state.DB.GetFollow(ctx, account.ID, targetAcc.ID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return gtserror.Newf(
			"db error getting follow from %s targeting %s: %w",
			account.ID, targetAcc.ID, err,
		)
	}

	if follow != nil {
		// Delete known follow from database with ID.
		err := p.state.DB.DeleteFollowByID(ctx, follow.ID)

		// If err == db.ErrNoEntries here then it indicates
		// a race condition with another unfollow for the same
		// account->target, but we should still process side
		// effects to be on the safe side.
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return gtserror.Newf(
				"db error deleting follow from %s targeting %s: %w",
				account.ID, targetAcc.ID, err,
			)
		}

		if sideEffects {
			// Queue unfollow side effects.
			p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
				APObjectType:   ap.ActivityFollow,
				APActivityType: ap.ActivityUndo,
				GTSModel:       follow,
				Origin:         account,
				Target:         targetAcc,
			})
		}
	}

	// Get follow request from requesting account to target account.
	followReq, err := p.state.DB.GetFollowRequest(ctx, account.ID, targetAcc.ID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return gtserror.Newf(
			"error getting follow request from %s targeting %s: %w",
			account.ID, targetAcc.ID, err,
		)
	}

	if followReq != nil {
		// Delete known follow request from database with ID.
		err := p.state.DB.DeleteFollowRequestByID(ctx, followReq.ID)

		// If err == db.ErrNoEntries here then it indicates
		// a race condition with another unfollow for the same
		// account->target, but we should still process side
		// effects to be on the safe side.
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return gtserror.Newf(
				"db error deleting follow request from %s targeting %s: %w",
				account.ID, targetAcc.ID, err,
			)
		}

		if sideEffects {
			// Queue unfollow-req side effects.
			p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
				APObjectType:   ap.ActivityFollow,
				APActivityType: ap.ActivityUndo,
				// Dummy out a follow to undo,
				// based on the follow request.
				GTSModel: &gtsmodel.Follow{
					AccountID:       account.ID,
					Account:         account,
					TargetAccountID: targetAcc.ID,
					TargetAccount:   targetAcc,
					URI:             followReq.URI,
				},
				Origin: account,
				Target: targetAcc,
			})
		}
	}

	return nil
}

// RemoveFromFollowers removes targetAcc from
// account's followers collection (if present).
func (p *Processor) RemoveFromFollowers(
	ctx context.Context,
	account *gtsmodel.Account,
	targetAcc *gtsmodel.Account,
) (*apimodel.Relationship, gtserror.WithCode) {
	// Check if a follow exists from
	// targetAccountID -> account.
	follow, err := p.state.DB.GetFollow(
		gtscontext.SetBarebones(ctx),
		targetAcc.ID,
		account.ID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("db error checking existing follow: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if follow != nil {
		// Follow exists, remove it.
		if err := p.state.DB.DeleteFollow(ctx,
			follow.AccountID,
			follow.TargetAccountID,
		); err != nil {
			err = gtserror.Newf("db error removing follow: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		// Handle side effects async.
		p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
			APObjectType:   ap.ActivityFollow,
			APActivityType: ap.ActivityReject,
			GTSModel:       follow,
			Origin:         account,
			Target:         targetAcc,
		})
	}

	// Return the (changed) relationship state.
	return p.APIRelationship(ctx, account, targetAcc.ID)
}

// BlocksGet gets a page of blocks owned by account.
//
// If "stub" is true, then returned accounts will
// be rendered as blocked accounts with minimal info.
func (p *Processor) BlocksGet(
	ctx context.Context,
	account *gtsmodel.Account,
	stub bool,
	page *paging.Page,
	path string,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	blocks, err := p.state.DB.GetAccountBlocking(ctx,
		account.ID,
		page,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Check for empty response.
	count := len(blocks)
	if len(blocks) == 0 {
		return util.EmptyPageableResponse(), nil
	}

	// Get the lowest and highest
	// ID values, used for paging.
	lo := blocks[count-1].ID
	hi := blocks[0].ID

	items := make([]interface{}, 0, count)
	for _, block := range blocks {
		var blockTarget *apimodel.Account
		var err error
		if stub {
			blockTarget, err = p.converter.AccountToAPIAccountBlocked(ctx, block.TargetAccount)
		} else {
			blockTarget, err = p.converter.AccountToAPIAccountPublic(ctx, block.TargetAccount)
		}
		if err != nil {
			log.Errorf(ctx, "error converting account to api: %v", err)
			continue
		}

		// Append target to return items.
		items = append(items, blockTarget)
	}

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  path,
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
	}), nil
}

func (p *Processor) getBlockTarget(
	ctx context.Context,
	account *gtsmodel.Account,
	targetAccountID string,
) (*gtsmodel.Account, *gtsmodel.Block, gtserror.WithCode) {
	// Account should not block or unblock itself.
	if account.ID == targetAccountID {
		err := gtserror.Newf("account %s cannot block or unblock itself", account.ID)
		return nil, nil, gtserror.NewErrorNotAcceptable(err, err.Error())
	}

	// Get block target.
	targetAccount, err := p.state.DB.GetAccountByID(ctx, targetAccountID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting account: %w", err)
		return nil, nil, gtserror.NewErrorInternalError(err)
	}

	if targetAccount == nil {
		const text = "target account not found"
		return nil, nil, gtserror.NewErrorNotFound(
			errors.New(text),
			text,
		)
	}

	// Check if already blocked.
	block, err := p.state.DB.GetBlock(ctx, account.ID, targetAccountID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error checking existing block: %w", err)
		return nil, nil, gtserror.NewErrorInternalError(err)
	}

	return targetAccount, block, nil
}

// BlockCreate handles the creation of a block from
// account to targetAccountID, either remote or local.
func (p *Processor) BlockCreate(
	ctx context.Context,
	account *gtsmodel.Account,
	targetAccountID string,
) (*apimodel.Relationship, gtserror.WithCode) {
	target, block, errWithCode := p.getBlockTarget(ctx, account, targetAccountID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if block != nil {
		// Block already exists, nothing to do.
		return p.APIRelationship(ctx, account, targetAccountID)
	}

	// Create and store a new block.
	blockID := id.NewULID()
	blockURI := uris.GenerateURIForBlock(
		account.PathPrefix(),
		account.Username,
		blockID,
	)
	block = &gtsmodel.Block{
		ID:              blockID,
		URI:             blockURI,
		AccountID:       account.ID,
		Account:         account,
		TargetAccountID: targetAccountID,
		TargetAccount:   target,
	}

	if err := p.state.DB.PutBlock(ctx, block); err != nil {
		err = fmt.Errorf("BlockCreate: error creating block in db: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Ensure each account unfollows the other.
	// We only care about processing unfollow side
	// effects from requesting account -> target
	// account, since requesting account is ours,
	// and target account might not be.
	if err := p.Unfollow(ctx,
		account,
		target,
		true, // sideEffects
	); err != nil {
		err = fmt.Errorf("BlockCreate: error unfollowing: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Ensure unfollowed in other direction,
	// don't process unfollow side effects.
	if err := p.Unfollow(ctx,
		target,
		account,
		false, // sideEffects
	); err != nil {
		err = fmt.Errorf("BlockCreate: error unfollowing: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Process block side effects (federation etc).
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		APObjectType:   ap.ActivityBlock,
		APActivityType: ap.ActivityCreate,
		GTSModel:       block,
		Origin:         account,
		Target:         target,
	})

	return p.APIRelationship(ctx, account, targetAccountID)
}

// BlockRemove handles the removal of a block from
// account to targetAccountID, either remote or local.
func (p *Processor) BlockRemove(
	ctx context.Context,
	account *gtsmodel.Account,
	targetAccountID string,
) (*apimodel.Relationship, gtserror.WithCode) {
	target, block, errWithCode := p.getBlockTarget(ctx, account, targetAccountID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if block == nil {
		// Already not blocked, nothing to do.
		return p.APIRelationship(ctx, account, targetAccountID)
	}

	// We got a block, remove it from the db.
	if err := p.state.DB.DeleteBlockByID(ctx, block.ID); err != nil {
		err := fmt.Errorf("BlockRemove: error removing block from db: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Populate account fields for convenience.
	block.Account = account
	block.TargetAccount = target

	// Process block removal side effects (federation etc).
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		APObjectType:   ap.ActivityBlock,
		APActivityType: ap.ActivityUndo,
		GTSModel:       block,
		Origin:         account,
		Target:         target,
	})

	return p.APIRelationship(ctx, account, targetAccountID)
}
