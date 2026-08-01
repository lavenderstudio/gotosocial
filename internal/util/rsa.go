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

package util

import (
	"crypto/rand"
	"crypto/rsa"

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
)

// Generate RSA keys
// of this length.
const rsaKeyBits = 2048

// NewActorRSA generates a new public/private RSA key pair.
func NewActorRSA() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, rsaKeyBits)
	if err != nil {
		err := gtserror.Newf("error creating new rsa private key: %w", err)
		return nil, nil, err
	}
	return privKey, &privKey.PublicKey, nil
}
