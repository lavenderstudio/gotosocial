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
	"cmp"
	"context"
	"fmt"
	"os/signal"
	"reflect"
	"slices"
	"strings"
	"syscall"
	"time"

	"code.superseriousbusiness.org/gopkg/log"
	"code.superseriousbusiness.org/gopkg/xslices"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/db/bundb"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"github.com/spf13/pflag"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/schema"
)

func main() {
	var srcCfg, dstCfg config.DatabaseConfiguration
	srcCfg = config.Defaults.Database
	dstCfg = config.Defaults.Database

	// Register flags for source and dest database.
	srcCfg.RegisterFlags("src-db", pflag.CommandLine)
	dstCfg.RegisterFlags("dst-db", pflag.CommandLine)
	pflag.Parse()

	// Convert flags to a map-type
	// that our config expects.
	m := make(map[string]any)
	pflag.VisitAll(func(flag *pflag.Flag) {
		m[flag.Name] = flag.Value
	})

	// Unmarshal map to our config structs.
	err1 := srcCfg.UnmarshalMap("src-db", m)
	err2 := dstCfg.UnmarshalMap("dst-db", m)
	if err := cmp.Or(err1, err2); err != nil {
		panic(err)
	}

	// Setup basic logging.
	log.SetLevel(log.INFO)
	config.SetLogDbQueries(false)

	// Set empty storage backend.
	config.SetStorageBackend("")

	// Prepare runtime context that
	// gets cancelled on system signal.
	ctx, cncl := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	defer cncl()

	// Perform actual database transfer.
	err := transfer(ctx, srcCfg, dstCfg)
	if err != nil {
		panic(err)
	}
}

func transfer(ctx context.Context, srcCfg, dstCfg config.DatabaseConfiguration) error {
	// Open a new database connection with 'srcCfg'
	// configuration as the source transfer database.
	return do(ctx, srcCfg, func(srcDB *bun.DB) error {

		// Open connection to source DB.
		srcConn, err := srcDB.Conn(ctx)
		if err != nil {
			return err
		}

		// Open a new database connection with 'dstCfg'
		// updated configuration as destination database.
		return do(ctx, dstCfg, func(dstDB *bun.DB) error {
			const batchsz = 1000

			// Open connection to dest DB.
			dstConn, err := dstDB.Conn(ctx)
			if err != nil {
				return err
			}

			// Check if this is an SQLite database with WAL enabled,
			// if so we can perform regular merges to reduce WAL size.
			isWAL := dstDB.Dialect().Name() == dialect.SQLite &&
				strings.EqualFold(config.GetDatabaseSQLiteJournalMode(), "WAL")

			// Range list of tables registered under bun internally.
			for _, table := range srcDB.Dialect().Tables().All() {
				var total, transferred int64

				// Get a count of all rows in current
				// table for logging and confirmation.
				if err := srcConn.NewSelect().
					Table(table.Name).
					ColumnExpr("COUNT(1)").
					Scan(ctx, &total); err != nil {
					return err
				}

				l := log.WithField("table", table.Name)
				l.Infof("transferring %d rows", total)

				// Get table pager which determines
				// which columns we order / page on.
				pager := getTablePager(table)

				// Calculate correct table slice type
				// for arguments to bun query .Model().
				tableType := reflect.PointerTo(table.Type)
				sliceType := reflect.SliceOf(tableType)

				// Allocate a new pointer to required slice
				// type, and at the pointer location allocate
				// new slice that can store expected batches.
				slicePtr := reflect.New(sliceType)
				slicePtr.Elem().Set(reflect.MakeSlice(sliceType, 0, batchsz))

				// Get a slice of fields that have the nullzero
				// flag set w/ a default, or can be marshaled to null.
				// We do this outside the hot-loop below for speed.
				nullZeroWithDefault := slices.DeleteFunc(
					slices.Clone(table.Fields),
					func(field *schema.Field) bool {
						return !field.NotNull || !(field.NullZero || field.SQLDefault == "")
					})

				// Get time for query
				// speed calculations.
				before := time.Now()

				for {
					// Ordered select batch of models
					// from determined table, passing
					// primary key for paging by PK.
					if err := srcConn.NewSelect().
						Model(slicePtr.Interface()).
						Where(pager.WhereGreaterSQL, pager.WhereGreaterArgs()...).
						OrderExpr(pager.OrderBySQL, pager.Columns...).
						Limit(batchsz).
						Scan(ctx); err != nil {
						return fmt.Errorf("error selecting batch of %s: %w", sliceType, err)
					}

					// Access slice at ptr.
					slice := slicePtr.Elem()
					length := slice.Len()

					// When nothing is returned,
					// we've reached table end.
					if length == 0 {
						break
					}

					// Check whether slice contains
					// a zero value field with an SQL
					// set default value, i.e. will
					// trigger a bun slice insert bug:
					//
					// https://github.com/uptrace/bun/issues/1394
					var hasNullZeroWithDefault bool
				fieldloop:
					for _, field := range nullZeroWithDefault {
						for i := range length {
							// Get elem at index in slice.
							elem := slice.Index(i).Elem()

							// Check if element has zero value field,
							// if so set the flag and break from loop.
							//
							// NOTE: the logic here follows that of:
							//       func (q *InsertQuery) getFields(),
							//       particularly field.marshalsToDefaul().
							if (field.IsPtr && field.HasNilValue(elem)) ||
								field.HasZeroValue(elem) {
								hasNullZeroWithDefault = true
								break fieldloop
							}
						}
					}
					if !hasNullZeroWithDefault {
						// We can insert this entire
						// slice in a single db query.
						if _, err := dstConn.
							NewInsert().
							Model(slicePtr.Interface()).
							Exec(ctx); err != nil {
							return fmt.Errorf("error inserting batch of %s: %w", sliceType, err)
						}
					} else {
						// Insert values into dest db
						// one at a time, due to bug.
						for i := range length {
							elem := slice.Index(i)
							model := elem.Interface()
							if _, err := dstConn.
								NewInsert().
								Model(model).
								Exec(ctx); err != nil {
								return fmt.Errorf("error inserting %v of batch of %s: %w", log.Formatted(model), tableType, err)
							}
						}
					}

					// Set updated PK minimum value.
					last := slice.Index(length - 1)
					pager.UpdateMinimum(last.Elem())

					// Reset slice
					// for next iter.
					slice.SetLen(0)

					// Increment transfer count.
					transferred += int64(length)

					// Get updated time.
					now := time.Now()

					// Calculate rows / second tx speed.
					timeTaken := now.Sub(before).Seconds()
					rowsPerSec := float64(length) / float64(timeTaken)

					// Update for
					// next query
					before = now

					// Calculate percentage of all rows transferred so far.
					perc := (float64(transferred) / float64(total)) * 100

					l.Infof("[~%.2f%% done; ~%.0f rows/s] transferring",
						perc, rowsPerSec)
				}

				// Ensure we transferred
				// expected count of rows.
				if transferred != total {
					return fmt.Errorf("transferred %d %s rows but expected %d", transferred, table.Name, total)
				}

				l.Infof("finished transferring %d rows", transferred)

				if isWAL {
					// This is an SQLite database with WAL enabled, perform WAL merge.
					_, err := dstDB.ExecContext(ctx, "PRAGMA wal_checkpoint(RESTART);")
					if err != nil {
						return gtserror.Newf("error performing wal_checkpoint: %w", err)
					}
				}
			}

			return nil
		})
	})
}

func do(ctx context.Context, cfg config.DatabaseConfiguration, do func(db *bun.DB) error) error {
	var state state.State

	defer func() {
		if state.DB != nil {
			// Lastly, if database service was started,
			// ensure it gets closed now all else stopped.
			if err := state.DB.Close(); err != nil {
				log.Errorf(ctx, "error stopping database: %v", err)
			}
		}

		// Finally reached end of shutdown.
		log.Info(ctx, "done! exiting...")
	}()

	// Initialize caches
	state.Caches.Init()
	if err := state.Caches.Start(); err != nil {
		return fmt.Errorf("error starting caches: %w", err)
	}

	log.Info(ctx, "starting db service...")

	// Set the provided database connection details.
	config.Config(func(c *config.Configuration) {
		c.Database = cfg
	})

	// Open conn to database now caches started.
	db, err := bundb.NewBunDBService(ctx, &state)
	if err != nil {
		return fmt.Errorf("error creating dbservice: %w", err)
	}

	// Get underlying bun DB service.
	bundb := db.(*bundb.DBService).DB()

	// Perform the provided db function.
	if err := do(bundb); err != nil {
		return fmt.Errorf("error executing query: %w", err)
	}

	return nil
}

// tablePager wraps and encompasses a bunch of the logic for
// paging through a reflected bun model's table in a database.
type tablePager struct {

	// A slice of SQL column names as bun.Ident,
	// boxed as interfaces for bun variadic args.
	Columns []any

	// A slice of struct field memory indices
	// of fields contained in index, used in
	// later reflect value extraction below.
	Indices [][]int

	// A slice of reflect values, used as iterative
	// previous minimum in the WHERE ? > ? clauses
	// of successive SELECT queries. This matches
	// the struct field columns and indices above.
	Minimum []reflect.Value

	// WhereGreaterSQL contains "WHERE ? > ? AND ..."
	// query expression for bun select query calls,
	// prepared for determined table index columns.
	WhereGreaterSQL string

	// OrderBySQL contains "ORDER BY ? " query
	// expression for bun select query calls,
	// prepared for determined table index columns.
	OrderBySQL string

	// cached arguments slice used in conjuction
	// with WhereGreaterSQL SELECT queries.
	whereGreaterArgs []any
}

// getTablePager returns a new tablePager{} appropriate for
// SELECT queries that index through pages of given table schema.
func getTablePager(table *schema.Table) (index *tablePager) {
	index = new(tablePager)

	var pageOn []*schema.Field
	if len(table.PKs) == 1 {
		// If only a single primary-key
		// column exists, use this.
		pageOn = table.PKs
	} else if i := slices.IndexFunc(table.Fields,
		func(field *schema.Field) bool {
			return field.Name == "id"
		}); i >= 0 {
		// For tables that contain a non-primary-key
		// "id" column, also prefer indexing on this.
		pageOn = []*schema.Field{table.Fields[i]}
	} else if len(table.PKs) == 0 {
		// If no primary keys exist for table,
		// just use list of all valid SQL cols.
		pageOn = slices.Clone(table.Fields)
		pageOn = slices.DeleteFunc(pageOn,
			func(field *schema.Field) bool {
				return field.Name == ""
			})
	} else {
		// Else use all of
		// its primary keys.
		pageOn = table.PKs
	}

	// Prepare slices of expected lengths.
	index.Columns = make([]any, 0, len(pageOn))
	index.Indices = make([][]int, 0, len(pageOn))
	index.Minimum = make([]reflect.Value, 0, len(pageOn))

	for _, field := range pageOn {
		// Append struct field's column name ident to index columns.
		index.Columns = append(index.Columns, bun.Ident(field.Name))

		// Append struct field's index in memory to index.
		index.Indices = append(index.Indices, field.Index)

		// Set a zero value for the starting previous minimum
		// for this table index, for most kinds we can accept
		// the default, but for integers we start at -1.
		if field.StructField.Type.Kind() == reflect.Int {
			index.Minimum = append(index.Minimum, reflect.ValueOf(int(-1)))
		} else {
			index.Minimum = append(index.Minimum, reflect.New(field.StructField.Type).Elem())
		}
	}

	if len(pageOn) == 1 {
		// Handle the simplest case.
		index.WhereGreaterSQL = "? > ?"
		index.OrderBySQL = "? ASC"
	} else {
		// Prepare argument placeholder SQL used as core of queries.
		placeholderSQL := strings.Repeat("?, ", len(index.Columns))
		placeholderSQL = placeholderSQL[:len(placeholderSQL)-2]

		// Prepare "ORDER BY CONCAT(?) ASC" query string.
		orderBySQL := "CONCAT(" + placeholderSQL + ")"
		index.OrderBySQL = orderBySQL + " ASC"

		// Prepare the "WHERE CONCAT(?) > CONCAT(?)" query string.
		whereGreaterSQL := "CONCAT(" + placeholderSQL + ") > "
		whereGreaterSQL += "CONCAT(" + placeholderSQL + ")"
		index.WhereGreaterSQL = whereGreaterSQL
	}

	return
}

// WhereGreaterArgs returns the arguments to
// be used in conjunction with WhereGreaterSQL.
func (idx *tablePager) WhereGreaterArgs() []any {
	if len(idx.Columns) != len(idx.Minimum) {
		panic("must be same number of columns and primary key values")
	}
	clear(idx.whereGreaterArgs[0:cap(idx.whereGreaterArgs)])
	idx.whereGreaterArgs = xslices.GrowJust(idx.whereGreaterArgs[:0], len(idx.Columns)+len(idx.Minimum))
	for i := range idx.Columns {
		idx.whereGreaterArgs = append(idx.whereGreaterArgs, idx.Columns[i])
	}
	for i := range idx.Minimum {
		idx.whereGreaterArgs = append(idx.whereGreaterArgs, idx.Minimum[i].Interface())
	}
	return idx.whereGreaterArgs
}

// UpdateMinimum updates the current Minimum slice element values to latest
// values contained in structValue, as extracted according to Indices.
func (idx *tablePager) UpdateMinimum(structValue reflect.Value) {
	if len(idx.Indices) != len(idx.Minimum) {
		panic("must be same number of field indices and primary key values")
	}
	clear(idx.Minimum)
	for i, index := range idx.Indices {
		idx.Minimum[i] = structValue.FieldByIndex(index)
	}
}
