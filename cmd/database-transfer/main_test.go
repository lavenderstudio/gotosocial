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

package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strings"
	"testing"
	"time"

	"code.superseriousbusiness.org/gopkg/log"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/google/go-cmp/cmp"
	"github.com/uptrace/bun"
)

// get docker implementation binary to use.
var _docker = func() string {
	path, err := exec.LookPath("docker")
	if err == nil {
		return path
	}
	path, err = exec.LookPath("podman")
	if err == nil {
		return path
	}
	return ""
}()

func init() {
	// Setup basic logging.
	log.SetLevel(log.ERROR)
	config.SetLogDbQueries(false)

	// Set empty storage backend.
	config.SetStorageBackend("")
}

func TestDatabaseTransfer(t *testing.T) {
	for _, testcase := range []struct {
		Name    string
		SrcInit func(*testing.T) (config.DatabaseConfiguration, func(), error)
		DstInit func(*testing.T) (config.DatabaseConfiguration, func(), error)
	}{
		{Name: "sqlite_to_sqlite", SrcInit: initSQLite, DstInit: initSQLite},
		{Name: "sqlite_to_postgres", SrcInit: initSQLite, DstInit: initPostgreSQL},
		{Name: "postgres_to_sqlite", SrcInit: initPostgreSQL, DstInit: initSQLite},
		{Name: "postgres_to_postgres", SrcInit: initPostgreSQL, DstInit: initPostgreSQL},
	} {
		t.Run(testcase.Name, func(t *testing.T) {
			// Initialize source database config.
			srcCfg, srcCncl, err := testcase.SrcInit(t)
			if err != nil {
				t.Fatalf("error initializing source database: %v", err)
			}

			// Ensure source
			// database removed.
			defer srcCncl()

			// Initialize destination database config.
			dstCfg, dstCncl, err := testcase.DstInit(t)
			if err != nil {
				t.Fatalf("error initializing destination database: %v", err)
			}

			// Ensure destination
			// database is removed.
			defer dstCncl()

			// Load test data into source database.
			err = loadTestData(t.Context(), srcCfg)
			if err != nil {
				t.Error(err)
			}

			// Test database transfer from source to destination.
			err = testDatabaseTransfer(t.Context(), srcCfg, dstCfg)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func initSQLite(t *testing.T) (config.DatabaseConfiguration, func(), error) {
	dbpath := path.Join(t.TempDir(), fmt.Sprintf("sqlite-%d.db", rand.Int64()))
	return config.DatabaseConfiguration{
			Type:    "sqlite",
			Address: dbpath,
		},
		func() { os.Remove(dbpath) },
		nil
}

func initPostgreSQL(t *testing.T) (config.DatabaseConfiguration, func(), error) {
	const addr = "127.0.0.1"
	const user = "postgres"
	const pass = "postgres"
	const name = "postgres"

	if _docker == "" {
		t.Skip("skipping postgres test, no docker binary found")
	}

	// Postgres can be slow to stop so start on a unique port,
	// (starting at 1024 as that's the unprivileged port range).
	port := 1024 + uint16(rand.Int32N(int32(^uint16(0)-1024)))

	// Prepare command to start postgres container.
	cmd := exec.CommandContext(t.Context(), _docker,
		"run",
		"--detach",
		"--env", fmt.Sprintf("POSTGRES_DB=%s", name),
		"--env", fmt.Sprintf("POSTGRES_USER=%s", user),
		"--env", fmt.Sprintf("POSTGRES_PASSWORD=%s", pass),
		"--env", "POSTGRES_HOST_AUTH_METHOD=trust",
		"--env", "PGHOST=0.0.0.0",
		"--env", fmt.Sprintf("PGPORT=%d", port),
		"--publish", fmt.Sprintf("%s:%d:%d", addr, port, port),
		"docker.io/postgres:latest",
	)

	// Run cmd and catch output.
	output, err := cmd.Output()
	if err != nil {
		return config.DatabaseConfiguration{}, nil, gtserror.Newf("error starting postgres container: %w", err)
	}

	// Trim any leading / end space
	// from output, this should be CID.
	cid := strings.TrimSpace(string(output))

	// Give time for it to start.
	t.Logf("container_id=%s", cid)
	time.Sleep(time.Second)

	kill := func() {
		// Attempt to kill container with this ID.
		cmd := exec.Command(_docker, "kill", cid)
		if err := cmd.Run(); err != nil {
			t.Fatalf("error killing container %s: %v", cid, err)
		}
	}

	return config.DatabaseConfiguration{
			Type:    "postgres",
			Address: addr,
			Postgres: config.PostgresConfiguration{
				Port:     port,
				User:     user,
				Password: pass,
				Database: name,
			},
		},
		kill,
		nil
}

// testDatabaseTransfer performs a test transfer, and later validation, from source to destination database configuration.
func testDatabaseTransfer(ctx context.Context, srcCfg, dstCfg config.DatabaseConfiguration) error {

	// Perform actual database transfer.
	err := transfer(ctx, srcCfg, dstCfg)
	if err != nil {
		return gtserror.Newf("error during database transfer: %w", err)
	}

	// Gather source data from database.
	srcData, err := gatherData(ctx, srcCfg)
	if err != nil {
		return gtserror.Newf("error gathering source data: %w", err)
	}

	// Gather destination data from database.
	dstData, err := gatherData(ctx, dstCfg)
	if err != nil {
		return gtserror.Newf("error gathering destination data: %w", err)
	}

	// Compare both data maps to ensure they are equal.
	if diff := cmp.Diff(srcData, dstData); diff != "" {
		return gtserror.Newf("data differs: %s", diff)
	}

	return nil
}

// loadTestData loads all our testrig models into the database defined by configuration.
func loadTestData(ctx context.Context, cfg config.DatabaseConfiguration) error {
	return do(ctx, cfg, func(db *bun.DB) error {
		if err := insertModels(ctx, db, testrig.NewTestTokens()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestApplications()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestBlocks()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestReports()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestRules()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestDomainBlocks()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestDomainLimits()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestInstances()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestUsers()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestAccounts()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestAccountSettings()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestAttachments()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestStatuses()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestStatusPins()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestEmojis()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestEmojiCategories()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestFederationErrors()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestStatusToEmojis()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestTags()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestStatusToTags()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestMentions()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestFaves()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestFollows()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestLists()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestListEntries()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestNotifications()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestTombstones()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestBookmarks()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestAccountNotes()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestMarkers()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestThreads()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestPolls()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestPollVotes()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestFilters()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestFilterKeywords()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestFilterStatuses()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestUserMutes()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestWebPushSubscriptions()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestInteractionRequests()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestStatusEdits()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, testrig.NewTestScheduledStatuses()); err != nil {
			return err
		}

		if err := insertModels(ctx, db, map[string]*gtsmodel.InstanceSettings{
			"": testrig.NewTestInstanceSettings(),
		}); err != nil {
			return err
		}

		return nil
	})
}

// placeHolderTime is reflected time value at time of
// testing that we replace all model time fields with
// the value of. this ensures we don't end-up with unique
// per-database time column values created on insert.
var placeHolderTime = reflect.ValueOf(time.Now())

// insertModels inserts all of given models in map into given database, inserting placeholder values where
// necessary to account for columns that in some cases can get non-deterministic database-set defaults.
func insertModels[Type any](ctx context.Context, db *bun.DB, models map[string]Type) error {
	for _, v := range models {
		rvalue := reflect.ValueOf(v)
		for rvalue.Kind() == reflect.Pointer {
			rvalue = rvalue.Elem()
		}
		if rvalue.Kind() == reflect.Struct {
			for i := 0; i < rvalue.NumField(); i++ {
				field := rvalue.Field(i)
				if field.Type() == reflect.TypeOf(time.Time{}) {
					field.Set(placeHolderTime)
				}
			}
		}
		if _, err := db.NewInsert().
			Model(v).
			Exec(ctx); err != nil {
			return gtserror.Newf("error inserting model %T: %w", v, err)
		}
	}
	return nil
}

// gatherData ...
func gatherData(ctx context.Context, cfg config.DatabaseConfiguration) (map[reflect.Type]any, error) {
	m := make(map[reflect.Type]any)

	if err := do(ctx, cfg, func(db *bun.DB) error {
		for _, table := range db.Dialect().Tables().All() {

			// Calculate correct table slice type
			// for arguments to bun query .Model().
			tableType := reflect.PointerTo(table.Type)
			sliceType := reflect.SliceOf(tableType)

			// Allocate a new pointer to required slice
			// type, and at the pointer location allocate
			// new slice that can store expected models.
			slicePtr := reflect.New(sliceType)
			slicePtr.Elem().Set(reflect.MakeSlice(sliceType, 0, 0))

			// Prepare a "pager" instance so we can use
			// it for generated OrderBy SQL expression.
			pager := getTablePager(table)

			// Scan all values of type
			// into prepared slice ptr.
			if err := db.NewSelect().
				Model(slicePtr.Interface()).
				OrderExpr(pager.OrderBySQL, pager.Columns...).
				Scan(ctx); err != nil {
				return gtserror.Newf("error scanning model %s: %w", sliceType, err)
			}

			// Store selected slice data under type ptr.
			m[table.Type] = slicePtr.Elem().Interface()
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return m, nil
}
