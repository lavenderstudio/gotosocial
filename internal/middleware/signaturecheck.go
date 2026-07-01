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

package middleware

import (
	"context"
	"net/http"
	"net/url"

	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gopkg/log"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"

	"code.superseriousbusiness.org/httpsig"
)

const (
	sigHeader  = string(httpsig.Signature)
	authHeader = string(httpsig.Authorization)

	// untyped error returned by httpsig when no signature is present
	noSigError = "neither \"" + sigHeader + "\" nor \"" + authHeader + "\" have signature parameters"
)

// ExtractSignature returns a middleware for extracting and validating HTTP signatures,
// NOTE: validate, not authenticate. It extracts signature details into the context and checks for
// blocked URIs. authentication happens in: ./internal/federation.Federator{}.AuthenticateFederatedRequest().
//
// The middleware first checks whether an incoming http request has been
// http-signed with a well-formed signature. If so, it will check if the
// domain that signed the request is permitted to access the server, using
// the provided uriBlocked function. If the domain is blocked, the middleware
// will abort the request chain with http code 403 forbidden. If it is not
// blocked, the handler will set the key verifier and the signature in the
// context for use down the line.
//
// In case of an error, the request will be aborted with http code 500.
func ExtractSignature(uriBlocked func(context.Context, *url.URL) (bool, error)) httputil.MiddlewareFunc {
	return func(h httputil.HandlerFunc) httputil.HandlerFunc {
		if h == nil {
			panic("nil func")
		}

		return func(c *httputil.Context) {
			// Create the signature verifier from the request;
			// this will error if the request wasn't signed.
			verifier, err := httpsig.NewVerifier(c.R)
			if err != nil {

				// Only actually *abort* the request with 401
				// if a signature was present but malformed.
				// Otherwise proceed with an unsigned request;
				// it's up to other functions to reject this.
				if err.Error() != noSigError {
					log.Warnf(c, "invalid signature scheme: %v", err)
					c.W.WriteHeader(http.StatusUnauthorized)
					return
				}

				// Pass on
				// to next.
				h(c)
				return
			}

			// The request was signed! The key ID should be given
			// in the signature so that we know where to fetch it
			// from the remote server. This will be something like:
			// https://example.org/users/some_remote_user#main-key
			pubKeyIDStr := verifier.KeyId()

			// Key can sometimes be nil, according to url parse
			// func: 'Trying to parse a hostname and path without
			// a scheme is invalid but may not necessarily return
			// an error, due to parsing ambiguities'. Catch this.
			pubKeyID, err := url.Parse(pubKeyIDStr)
			if err != nil || pubKeyID == nil {
				log.Warnf(c, "invalid pubkey id url: %s", pubKeyIDStr)
				c.W.WriteHeader(http.StatusUnauthorized)
				return
			}

			// If the domain is blocked we want to bail as fast as
			// possible without the request proceeding further.
			blocked, err := uriBlocked(c, pubKeyID)
			if err != nil {
				log.Errorf(c, "db error checking domain block %s: %s", pubKeyID.Host, err)
				c.W.WriteHeader(http.StatusInternalServerError)
				return
			}

			if blocked {
				log.Infof(c, "domain %s is blocked", pubKeyID.Host)
				c.W.WriteHeader(http.StatusForbidden)
				return
			}

			// Assume signature was set on Signature header,
			// but fall back to Authorization header if necessary.
			signature := c.R.Header.Get(sigHeader)
			if signature == "" {
				signature = c.R.Header.Get(authHeader)
			}

			// Set relevant values in the request context
			// for later signature checking down the line.
			//
			// Note since we're passing in an httputil.Context{}
			// here, the value will get set in httputil.Context{}.V,
			// which is why we can ignore the return value here.
			_ = gtscontext.SetHTTPSignatureVerifier(c, verifier)
			_ = gtscontext.SetHTTPSignature(c, signature)
			_ = gtscontext.SetHTTPSignaturePubKeyID(c, pubKeyID)

			// Pass on
			// to next.
			h(c)
		}
	}
}
