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

package federatingdb

import (
	"context"
	"errors"
	"net/url"
	"slices"
	"time"

	"code.superseriousbusiness.org/activity/streams/vocab"
	"code.superseriousbusiness.org/gopkg/log"
	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/messages"
)

func (f *DB) Announce(ctx context.Context, announce vocab.ActivityStreamsAnnounce) error {
	log.DebugKV(ctx, "announce", Serialize{announce})

	activityContext := getActivityContext(ctx)
	if activityContext.internal {
		return nil // Already processed.
	}

	requesting := activityContext.requestingAcct
	receiving := activityContext.receivingAcct

	if requesting.IsMoving() {
		// A Moving account
		// can't do this.
		return nil
	}

	// Ensure requestingAccount is among
	// the Actors doing the Announce.
	//
	// We don't support Announce forwards.
	actorIRIs := ap.GetActorIRIs(announce)
	if !slices.ContainsFunc(actorIRIs, func(actorIRI *url.URL) bool {
		return actorIRI.String() == requesting.URI
	}) {
		// Just return nil (status 202) here and
		// not error, as it's not really an error
		// per se, just something we don't support.
		log.Debugf(ctx,
			"requestingAccount %s was not among Announce Actors, dropping Announce forward",
			requesting.URI,
		)
		return nil
	}

	// Check if the announce originates from an actor
	// we target with at least one relay subscription.
	relaySubscriptions, err := f.state.DB.GetRelaySubscriptionsByActorURI(ctx, requesting.URI)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return gtserror.Newf("db error getting relay subscriptions for actor URI %s: %w", requesting.URI, err)
	}

	if len(relaySubscriptions) != 0 {
		// We subscribe to this actor with at least one
		// relay subscription, which means it's a relay.
		//
		// We only accept delivery from relay actors to
		// our instance account's inbox, so check this.
		if !receiving.IsInstance() {
			log.Debugf(ctx, "dropping delivery from %s (relay actor delivering to non-instance-actor inbox)", requesting.URI)
			return nil
		}

		// Some relay software doesn't set published
		// prop on the Announce. If this is so, just set
		// time.Now() and let the typeconverter use that.
		published := ap.GetPublished(announce)
		if published.IsZero() {
			ap.SetPublished(announce, time.Now())
		}

		// Convert boost to gtsmodel.
		//
		// We don't store boosts wrappers from
		// relays so don't bother checking here.
		boost, _, err := f.converter.ASAnnounceToStatus(ctx, announce)
		if err != nil {
			return gtserror.Newf("error converting announce to boost: %w", err)
		}

		// From relay actors we don't care about
		// storing and generating notifications
		// for Announces of our *own* posts.
		uri := boost.BoostOfURI
		if uri.Host == config.GetHost() ||
			uri.Host == config.GetAccountDomain() {
			log.Debugf(ctx, "dropping delivery from %s (relay actor announcing one of our posts)", requesting.URI)
			return nil
		}

		// Ensure we actually follow this
		// relay actor with the instance account.
		following, err := f.state.DB.IsFollowing(ctx, receiving.ID, requesting.ID)
		if err != nil {
			return gtserror.Newf("db error checking follow of actor URI %s: %w", requesting.URI, err)
		}
		if !following {
			// No follow means we're not interested.
			log.Debugf(ctx, "dropping delivery from %s (not following this actor)", requesting.URI)
			return nil
		}

		// Allow processing of the
		// relay announce to continue.
	}

	// Convert boost to gtsmodel.
	boost, isNew, err := f.converter.ASAnnounceToStatus(ctx, announce)
	if err != nil {
		return gtserror.Newf("error converting announce to boost: %w", err)
	}

	if !isNew {
		// We've already seen
		// and stored this boost;
		// nothing else to do here.
		return nil
	}

	// This is a new boost. Process side effects asynchronously.
	f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
		APObjectType:   ap.ActivityAnnounce,
		APActivityType: ap.ActivityCreate,
		GTSModel:       boost,
		Receiving:      receiving,
		Requesting:     requesting,
	})

	return nil
}
