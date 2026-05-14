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
	"errors"
	"net/url"

	"code.superseriousbusiness.org/gopkg/xslices"
	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/transport"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

// getStatusDBOnly checks in the database for
// status with the given URI (or URL), without
// doing any external dereferencing.
func (d *Dereferencer) getStatusDBOnly(
	ctx context.Context,
	uriStr string,
) (*gtsmodel.Status, error) {
	// Request a barebones object:
	// status may be in the db but with
	// related models not yet dereffed.
	ctxBb := gtscontext.SetBarebones(ctx)

	// Search the database for existing by URI.
	status, err := d.state.DB.GetStatusByURI(ctxBb, uriStr)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("error checking database for status %s by uri: %w", uriStr, err)
		return nil, err
	}

	if status != nil {
		// Found it,
		// stop early.
		return status, nil
	}

	// Else, search database for existing by URL.
	status, err = d.state.DB.GetStatusByURL(ctxBb, uriStr)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("error checking database for status %s by url: %w", uriStr, err)
		return nil, err
	}

	// Return maybe status.
	return status, nil
}

// retrieveStatusable dereferences the given URI and
// processes the response into an ap.Statusable model.
//
// In case of HTTP redirects to a final URI that differs
// from the input URI, the input URI pointer will be changed
// to the final URI, and the database will be checked once
// more to see if the status was stored locally at that URI.
// If so, the stored status will be returned as alreadyStatus.
//
// Will return malformed if the final redirected URI is not
// either the AP ID/URI or the URL of the dereffed statusable.
func (d *Dereferencer) retrieveStatusable(
	ctx context.Context,
	tsport transport.Transport,
	uri *url.URL,
) (
	statusable ap.Statusable,
	alreadyStatus *gtsmodel.Status,
	err error,
) {
	// Save this for later comparison.
	initialURIStr := uri.String()

	// Dereference latest version of the status.
	rsp, err := tsport.Dereference(ctx, uri)
	if err != nil {
		err := gtserror.Newf("error dereferencing %s: %w", uri, err)
		return nil, nil, gtserror.SetUnretrievable(err)
	}

	// Attempt to resolve ActivityPub status from response.
	statusable, err = ap.ResolveStatusable(ctx, rsp.Body)

	// Tidy up now done.
	_ = rsp.Body.Close()

	if err != nil {
		// ResolveStatusable will set gtserror.WrongType
		// on the returned error, so we don't need to do it here.
		err := gtserror.Newf("error resolving statusable %s: %w", uri, err)
		return nil, nil, err
	}

	// Check whether input URI and final returned URI
	// have changed (i.e. we followed some redirects).
	//
	// NOTE: this URI check + database call is performed
	// AFTER reading and closing body, for performance.
	var (
		finalURI    = rsp.Request.URL
		finalURIStr = rsp.Request.URL.String()
		redirected  = finalURIStr != initialURIStr
	)

	if redirected {
		// Update passed-in URI
		// for benefit of the caller.
		uri = finalURI

		// Check whether we have this status
		// stored under *final* URI and return
		// it to the caller if so.
		var err error
		alreadyStatus, err = d.getStatusDBOnly(ctx, finalURIStr)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			err := gtserror.Newf("db error getting status after redirects: %w", err)
			return nil, nil, err
		}
	}

	// Ensure the final URI we fetched the status
	// from matches either (one of) the URL(s) or
	// the ID/URI of the dereferenced statusable.
	okURIs := append(
		ap.GetURL(statusable),      // status URL(s)
		ap.GetJSONLDId(statusable), // status URI
	)
	matches, err := util.URIMatches(finalURI, okURIs...)
	if err != nil {
		err := gtserror.Newf("error checking final dereferenced status uri %s: %w", finalURIStr, err)
		return nil, nil, err
	}

	if !matches {
		// There's not a match, so the remote is doing
		// something weird. Gather URI strings we would
		// have accepted into nice slice for logging.
		var okURIStrs []string
		okURIStrs = xslices.Gather(
			okURIStrs,
			okURIs,
			func(u *url.URL) string {
				return u.String()
			},
		)

		// Construct error to give a bit more information
		// in case there were one or more redirects.
		var err error
		if redirected {
			err = gtserror.Newf(
				"final http URI %s, after redirect(s) from initial URI %s, does not match dereferenced statusable id or url(s) %+v",
				finalURIStr, initialURIStr, okURIStrs,
			)
		} else {
			err = gtserror.Newf(
				"http URI %s does not match dereferenced statusable id or url(s) %+v",
				initialURIStr, okURIStrs,
			)
		}

		// Set malformed on the returned error.
		return nil, nil, gtserror.SetMalformed(err)
	}

	return
}

// convertStatusable converts the given statusable to its gts model equivalent.
// The requestUser param is used to ensure the status author is dereferenced.
// The URI is used to check that the status's AP ID/URI matches expectations.
func (d *Dereferencer) convertStatusable(
	ctx context.Context,
	requestUser string,
	uri *url.URL,
	statusable ap.Statusable,
) (*gtsmodel.Status, error) {
	// Get the attributed-to ID/URI in order to fetch account.
	attributedTo, err := ap.ExtractAttributedToURI(statusable)
	if err != nil {
		return nil, gtserror.New("attributedTo was empty")
	}

	// Ensure we have the author account of the status dereferenced
	// (and up-to-date); this is needed to convert to our GTS model.
	if _, _, err := d.getAccountByURI(ctx, requestUser, attributedTo, false); err != nil {

		// Note that we specifically DO NOT wrap the error, instead collapsing it as string.
		// Errors fetching an account do not necessarily relate to dereferencing the status.
		return nil, gtserror.Newf("failed to dereference status author %s: %v", uri, err)
	}

	// Convert AP model to our GTS model.
	status, err := d.converter.ASStatusToStatus(ctx, statusable)
	if err != nil {
		return nil, gtserror.Newf("error converting statusable to gts model for status %s: %w", uri, err)
	}

	// Ensure final status isn't attempting
	// to claim being authored by local user.
	if status.Account.IsLocal() {
		return nil, gtserror.Newf(
			"dereferenced status %s claiming to be local",
			status.URI,
		)
	}

	return status, nil
}
