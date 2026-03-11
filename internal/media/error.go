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

package media

import (
	"context"
	"net"

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"codeberg.org/gruf/go-errors/v2"
)

var codecUnsupportedDetails = gtsmodel.NewMediaErrorDetails(
	gtsmodel.MediaErrorTypeCodec,
	gtsmodel.MediaErrorTypeCodec_Unsupported,
)

var codecDetails = gtsmodel.NewMediaErrorDetails(
	gtsmodel.MediaErrorTypeCodec,
	0,
)

// errWithDetails allows optionally wrapping an error,
// but largely propagating MediaErrorDetails via error return.
type errWithDetails struct {
	error
	details gtsmodel.MediaErrorDetails
}

// withDetails wraps an optional error with given MediaErrorDetails as error type.
func withDetails(err error, details gtsmodel.MediaErrorDetails) error {
	return &errWithDetails{err, details}
}

func (err *errWithDetails) Error() string {
	if err.error == nil {
		// if no error was set, instead
		// use stringified details given.
		return err.details.String()
	}
	return err.error.Error()
}

func (err *errWithDetails) Unwrap() error {
	return err.error
}

// isStubError returns whether determined gtsmodel.MediaErrorDetails
// was due to a "stubbing" type error, i.e. a sort of non-error.
func isStubError(details gtsmodel.MediaErrorDetails) bool {
	return details.Type() == gtsmodel.MediaErrorTypePolicy ||
		details.Details() == gtsmodel.MediaErrorTypeCodec_Unsupported
}

// toErrorDetails will convert given error to extracted MediaErrorDetails (if any).
func toErrorDetails(err error) gtsmodel.MediaErrorDetails {
	if err == nil {
		// No error was returned, no details.
		return gtsmodel.NewMediaErrorDetails(
			gtsmodel.MediaErrorTypeNone,
			0,
		)

	} else if withDetails := errors.AsV2[*errWithDetails](err); withDetails != nil {
		// Return stored err details.
		return withDetails.details

	} else if errors.IsV2(err, context.Canceled, context.DeadlineExceeded) {
		// Interrupt error due to context cancelled.
		return gtsmodel.NewMediaErrorDetails(
			gtsmodel.MediaErrorTypeInterrupt,
			0,
		)

	} else if details := extractNetworkErrorDetails(err); details != 0 {
		// Return determined
		// error details.
		return details
	}

	// Any other type was unclassified error.
	return gtsmodel.NewMediaErrorDetails(
		gtsmodel.MediaErrorTypeUnknown,
		0,
	)
}

// extractNetworkErrorDetails looks for and returns any network / http related details in error.
func extractNetworkErrorDetails(err error) gtsmodel.MediaErrorDetails {
	if code := gtserror.StatusCode(err); code > 0 {
		// An HTTP status code was set, indicating error
		// due to HTTP response, extract and set details.
		return gtsmodel.NewMediaErrorDetails(
			gtsmodel.MediaErrorTypeHTTP,
			uint16(code), // nolint:gosec
		)

	} else if netErr := errors.AsV2[interface{ Timeout() bool }](err); netErr != nil {
		var details uint16

		// All "net{,/http}" package errors implement
		// Timeout(), use this to set type and details.
		if netErr.Timeout() {
			details = gtsmodel.MediaErrorTypeNetwork_Timeout
		} else if _, isDNS := netErr.(*net.DNSError); isDNS {
			details = gtsmodel.MediaErrorTypeNetwork_DNS
		}

		return gtsmodel.NewMediaErrorDetails(
			gtsmodel.MediaErrorTypeNetwork,
			details,
		)
	}
	return 0
}
