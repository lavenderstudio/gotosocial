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

package dereferencing

import (
	"context"
	"net/url"

	"code.superseriousbusiness.org/gopkg/log"
	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/transport"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

// GetRelayedStatus dereferences and returns a status with
// the given URI only if it is permitted to be relayed by
// matching with at least one active relay subscription.
//
// If the status matches, then it will be inserted into
// the db and the rest of its thread will be dereferenced,
// calling newThreadEntryCallback for each parent or child.
// The status will then be returned for further processing.
//
// If the status doesn't match against at least one relay
// subscription, but there is no error, then nil, nil will
// be returned, and the status can be considered dropped.
//
// If the status matches against a relay subscription, but
// permissivity checks show that it is pending approval,
// then nil, nil will be returned, and the status is dropped.
//
// All necessary dereferencing will be done using the passed
// instance service account, as that is the actor that
// receives messages from subscribed relay actors.
func (d *Dereferencer) GetRelayedStatus(
	ctx context.Context,
	instanceAcct *gtsmodel.Account,
	relayAcct *gtsmodel.Account,
	uri *url.URL,
	newThreadEntryCallback func(context.Context, *gtsmodel.Status) error,
) (*gtsmodel.Status, error) {
	// Check whether this status URI is a blocked domain / subdomain.
	if blocked, err := d.state.DB.IsDomainBlocked(ctx, uri.Host); err != nil {
		return nil, gtserror.Newf("error checking blocked domain: %w", err)
	} else if blocked {
		err := gtserror.Newf("%s is blocked", uri.Host)
		return nil, gtserror.SetUnretrievable(err)
	}

	// Stringify URI once.
	uriStr := uri.String()

	// Create logger.
	l := log.
		WithContext(ctx).
		WithField("uri", uriStr)

	// Acquire per-URI deref lock.
	unlock := d.state.FedLocks.Lock(uriStr)
	unlock = util.DoOnce(unlock)
	defer unlock()

	// Search the database for existing status.
	if status, err := d.getStatusDBOnly(ctx, uriStr); err != nil {
		return nil, err
	} else if status != nil {
		// If we already have the status,
		// someone sent it to us already
		// or we dereferenced it in some
		// other way. Nothing to do here.
		return status, nil
	}

	// We don't have the relayed status
	// stored locally, go dereference it.
	status, statusable, err := d.dereferenceRelayableStatus(
		ctx, l,
		instanceAcct,
		relayAcct,
		uri,
	)
	if status == nil {
		// Status is not relevant / not
		// relayable for whatever reason,
		// or there was an error dereffing.
		return nil, err
	}

	// Status is relevant so we'll probably store it.
	// Generate new status ID from the provided creation date.
	status.ID = id.NewULIDFromTime(status.CreatedAt)

	// Check if this is a permitted status we should accept.
	// Function also sets "PendingApproval" bool as necessary.
	permit, err := d.isPermittedStatus(ctx,
		instanceAcct.Username,
		nil,    // existing status (this is new, so nil)
		status, // the new status
		true,   // isNew = true
	)
	if err != nil {
		err := gtserror.Newf("error checking permissibility for status %s: %w", uri, err)
		return nil, err
	}

	if !permit {
		// Return a checkable error type that can be ignored.
		err := gtserror.Newf("dropping unpermitted status: %s", uri)
		return nil, gtserror.SetNotPermitted(err)
	}

	if status.Flags.PendingApproval() {
		// If the status is pending approval we don't really
		// care about it, we'll get it from somewhere else at
		// some point perhaps. We only care about statuses
		// that are already approved or don't require approval.
		l.Debug("status pending approval, not interested")
		return nil, nil
	}

	// All checks passed, enrich status
	// peripheral stuff (attachments etc).
	if _, err := d.handleStatusPeripherals(ctx,
		instanceAcct.Username,
		uri,
		&gtsmodel.Status{URI: uriStr},
		status,
	); err != nil {
		err := gtserror.Newf("error handling peripheral dereferencing: %w", err)
		return nil, err
	}

	// Store the enriched status.
	if err := d.state.DB.PutStatus(ctx, status); err != nil {
		err := gtserror.Newf("error inserting new status %s: %w", uri, err)
		return nil, err
	}

	// Unlock per-URI
	// deref lock.
	unlock()

	// Deref parents + children.
	d.dereferenceThread(ctx,
		instanceAcct.Username,
		uri,
		status,
		statusable,
		true, // isNew = true
		newThreadEntryCallback,
	)

	return status, nil
}

func (d *Dereferencer) GetRelayedAnnounce(
	ctx context.Context,
	instanceAcct *gtsmodel.Account,
	relayAcct *gtsmodel.Account,
	boostWrapper *gtsmodel.Status,
	newThreadEntryCallback func(context.Context, *gtsmodel.Status) error,
) (*gtsmodel.Status, error) {
	// Check whether boosted status URI
	// is a blocked domain / subdomain.
	uri := boostWrapper.BoostOfURI
	if blocked, err := d.state.DB.IsDomainBlocked(ctx, uri.Host); err != nil {
		return nil, gtserror.Newf("error checking blocked domain: %w", err)
	} else if blocked {
		err := gtserror.Newf("%s is blocked", uri.Host)
		return nil, gtserror.SetUnretrievable(err)
	}

	// Create logger.
	uriStr := boostWrapper.BoostOfURIStr
	l := log.
		WithContext(ctx).
		WithField("uri", uriStr)

	// Acquire per-URI deref lock.
	unlock := d.state.FedLocks.Lock(uriStr)
	unlock = util.DoOnce(unlock)
	defer unlock()

	// Search the database for existing status.
	if status, err := d.getStatusDBOnly(ctx, uriStr); err != nil {
		return nil, err
	} else if status != nil {
		// If we already have the status,
		// someone sent it to us already
		// or we dereferenced it in some
		// other way. Nothing to do here.
		return status, nil
	}

	// We don't have the relayed status
	// stored locally, go dereference it.
	status, statusable, err := d.dereferenceRelayableStatus(
		ctx, l,
		instanceAcct,
		relayAcct,
		uri,
	)
	if status == nil {
		// Status is not relevant / not
		// relayable for whatever reason,
		// or there was an error dereffing.
		return nil, err
	}

	// Status is relevant so we'll probably store it.
	// Generate new status ID from the provided creation date.
	status.ID = id.NewULIDFromTime(status.CreatedAt)

	// Check if this is a permitted status we should accept.
	// Function also sets "PendingApproval" bool as necessary.
	permit, err := d.isPermittedStatus(ctx,
		instanceAcct.Username,
		nil,    // existing status (this is new, so nil)
		status, // the new status
		true,   // isNew = true
	)
	if err != nil {
		err := gtserror.Newf("error checking permissibility for status %s: %w", uri, err)
		return nil, err
	}

	if !permit {
		// Return a checkable error type that can be ignored.
		err := gtserror.Newf("dropping unpermitted status: %s", uri)
		return nil, gtserror.SetNotPermitted(err)
	}

	if status.Flags.PendingApproval() {
		// If the boosted status is pending approval we don't
		// really care about it, we'll get it from somewhere
		// else at some point perhaps. We only care about statuses
		// that are already approved or don't require approval.
		l.Debug("status pending approval, not interested")
		return nil, nil
	}

	// Check if the boost itself is permitted, if it's
	// not then we shouldn't store the relayed status.
	boostWrapper.BoostOf = status
	boostPermit, err := d.isPermittedStatus(ctx,
		instanceAcct.Username,
		nil,          // existing status (this is new, so nil)
		boostWrapper, // the new boost
		true,         // isNew = true
	)
	if err != nil {
		err := gtserror.Newf("error checking permissibility for boost of %s: %w", uri, err)
		return nil, err
	}

	if !boostPermit {
		// Return a checkable error type that can be ignored.
		err := gtserror.Newf("dropping unpermitted boost: %s", uri)
		return nil, gtserror.SetNotPermitted(err)
	}

	if status.Flags.PendingApproval() {
		// If the boost wrapper is pending
		// approval we can't be arsed with it.
		l.Debug("boost pending approval, not interested")
		return nil, nil
	}

	// All checks passed, enrich status
	// peripheral stuff (attachments etc).
	if _, err := d.handleStatusPeripherals(ctx,
		instanceAcct.Username,
		uri,
		&gtsmodel.Status{URI: uriStr},
		status,
	); err != nil {
		err := gtserror.Newf("error handling peripheral dereferencing: %w", err)
		return nil, err
	}

	// Store the enriched status.
	// Note: we don't bother storing the boost.
	if err := d.state.DB.PutStatus(ctx, status); err != nil {
		err := gtserror.Newf("error inserting new status %s: %w", uri, err)
		return nil, err
	}

	// Unlock per-URI
	// deref lock.
	unlock()

	// Deref parents + children.
	d.dereferenceThread(ctx,
		instanceAcct.Username,
		uri,
		status,
		statusable,
		true, // isNew = true
		newThreadEntryCallback,
	)

	return status, nil
}

func (d *Dereferencer) dereferenceRelayableStatus(
	ctx context.Context,
	l log.Entry,
	instanceAcct *gtsmodel.Account,
	relayAcct *gtsmodel.Account,
	uri *url.URL,
) (*gtsmodel.Status, ap.Statusable, error) {
	// Don't have the status locally, so we need to deref.
	//
	// Create transport on behalf of our instance account,
	// which is the account we receive relay traffic with.
	tsport, err := d.transportController.NewTransport(
		instanceAcct.PublicKeyURI,
		instanceAcct.PrivateKey,
	)
	if err != nil {
		return nil, nil, gtserror.Newf("couldn't create transport: %w", err)
	}

	// Dereference statusable from remote, checking
	// if we already had this status stored under a
	// different URI (ie., final URI after redirects).
	statusable, alreadyStatus, err := d.retrieveStatusable(ctx, tsport, uri)
	if err != nil {
		return nil, nil, err
	}

	// If alreadyStatus was returned, it means
	// we already had this status stored in the
	// db under its final reachable URI and not
	// the URI passed in. No need to go further.
	if alreadyStatus != nil {
		return alreadyStatus, statusable, nil
	}

	// We didn't have the status yet + we dereffed it.
	//
	// Convert the statusable to GtS format so we
	// can do relevance checks *before* storing it.
	//
	// This will also fetch the status author account.
	status, err := d.convertStatusable(ctx,
		instanceAcct.Username, uri, statusable,
	)
	if err != nil {
		return nil, statusable, err
	}

	// We're only interested in relayed statuses that
	// are public or unlisted, as followers-only statuses
	// will be definition by sent to followers anyway.
	vis := status.Visibility
	if !(vis == gtsmodel.VisibilityPublic ||
		vis == gtsmodel.VisibilityUnlocked) {
		l.Debug("status neither public nor unlisted")
		return nil, statusable, nil
	}

	// For relevance checks we have to dereference the
	// parent status already (if there is one) and check
	// if the author is the same as for this status.
	var inReplyToAccountURI string
	if status.InReplyToURI != "" {
		var err error
		inReplyToAccountURI, err = d.retrieveInReplyToAccountURI(ctx, tsport, status)
		if err != nil {
			return nil, statusable, err
		}
	}

	// Check if this status matches at least one relay
	// subscription. If it doesn't, we shouldn't relay it.
	sub, err := d.relayFilter.MatchedBySubscription(ctx,
		relayAcct,
		status,
		inReplyToAccountURI, // may be empty string
	)
	if err != nil {
		err := gtserror.Newf(
			"error checking if status %s is matched by a relay subscription: %w",
			uri, err,
		)
		return nil, statusable, err
	}

	if sub == nil {
		l.Debug("status not matched by any subscriptions")
		return nil, statusable, nil
	}

	// Looks relayable!
	return status, statusable, nil
}

// retrieveInReplyToAccountURI dereferences the parent
// of the given status to retrieve the URI of the account.
func (d *Dereferencer) retrieveInReplyToAccountURI(
	ctx context.Context,
	tsport transport.Transport,
	status *gtsmodel.Status,
) (string, error) {
	if status.InReplyTo != nil {
		// Had parent stored already.
		return status.InReplyTo.AccountURI, nil
	}

	// We'll have to dereference.
	inReplyToURI := status.InReplyToURI
	parentURI, err := url.Parse(inReplyToURI)
	if err != nil {
		err := gtserror.Newf("error parsing parent URI: %w", err)
		return "", err
	}

	// Make sure it's not replying to something on our
	// domain that may have been deleted from the db.
	if parentURI.Host == config.GetHost() ||
		parentURI.Host == config.GetAccountDomain() {
		return "", nil
	}

	// Check whether the parent status URI is a blocked
	// domain, we don't wanna be making requests to baddies.
	blocked, err := d.state.DB.IsDomainBlocked(ctx, parentURI.Host)
	if err != nil {
		err := gtserror.Newf("db error checking blocked domain: %w", err)
		return "", err
	}

	if blocked {
		// Just return without
		// dereffing, kiss my ass.
		return "", nil
	}

	// We don't have the parent stored, try to fetch but
	// *don't* store it, we only want to check for now.
	parentStatusable, parentStatus, err := d.retrieveStatusable(ctx, tsport, parentURI)
	if err != nil {
		err := gtserror.Newf("error retrieving %s: %w", inReplyToURI, err)
		return "", err
	}

	if parentStatus != nil {
		// We had the parent status
		// stored at a different URI!
		return parentStatus.AccountURI, nil
	}

	attributedTo := ap.GetAttributedTo(parentStatusable)
	if len(attributedTo) == 0 {
		err := gtserror.Newf("parent %s had no attributedTo", inReplyToURI)
		return "", err
	}

	return attributedTo[0].String(), nil
}
