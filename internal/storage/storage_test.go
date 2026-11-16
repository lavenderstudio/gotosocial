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

//go:build !nos3

package storage_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestS3PresignedURLKeyPrefix(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead || r.URL.Path != "/user-uploads/" {
			t.Errorf("unexpected S3 request: %s %s", r.Method, r.URL)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var original config.Configuration
	config.Config(func(cfg *config.Configuration) { original = *cfg })
	t.Cleanup(func() {
		config.Config(func(cfg *config.Configuration) { *cfg = original })
	})

	for _, prefix := range []string{"", "social.systems/"} {
		t.Run("prefix="+prefix, func(t *testing.T) {
			config.Config(func(cfg *config.Configuration) {
				cfg.StorageS3Endpoint = strings.TrimPrefix(server.URL, "http://")
				cfg.StorageS3AccessKey = "test-access-key"
				cfg.StorageS3SecretKey = "test-secret-key"
				cfg.StorageS3UseSSL = false
				cfg.StorageS3BucketName = "user-uploads"
				cfg.StorageS3BucketLookup = "path"
				cfg.StorageS3Region = "us-east-1"
				cfg.StorageS3KeyPrefix = prefix
				cfg.StorageS3Proxy = false
				cfg.StorageS3RedirectURL = ""
			})

			driver, err := storage.NewS3Storage()
			require.NoError(t, err)
			t.Cleanup(func() { driver.PresignedCache.Stop() })

			const key = "01884/attachment/original/image.jpg"
			signed := driver.URL(context.Background(), key)
			require.NotNil(t, signed)
			require.Equal(t, "/user-uploads/"+prefix+key, signed.Path)
			require.NotEmpty(t, signed.Query().Get("X-Amz-Signature"))
			require.Equal(t, signed, driver.URL(context.Background(), key))
		})
	}
}
