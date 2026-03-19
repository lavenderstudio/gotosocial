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

package cleaner

import (
	"context"
	"time"
	"unsafe"

	"code.superseriousbusiness.org/gopkg/log"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/storage"
)

const selectLimit = 50

type Cleaner struct{ state *state.State }

func New(state *state.State) *Cleaner {
	c := new(Cleaner)
	c.state = state
	return c
}

// Emoji returns the emoji set of cleaner utilities.
func (c *Cleaner) Emoji() *Emoji {
	if unsafe.Sizeof(Emoji{}) != unsafe.Sizeof(Cleaner{}) ||
		unsafe.Offsetof(Emoji{}.Cleaner) != 0 {
		panic(gtserror.New("compile time unsafe pointer assertion"))
	}
	return (*Emoji)(unsafe.Pointer(c))
}

// Media returns the media set of cleaner utilities.
func (c *Cleaner) Media() *Media {
	if unsafe.Sizeof(Media{}) != unsafe.Sizeof(Cleaner{}) ||
		unsafe.Offsetof(Media{}.Cleaner) != 0 {
		panic(gtserror.New("compile time unsafe pointer assertion"))
	}
	return (*Media)(unsafe.Pointer(c))
}

// Status returns the status set of cleaner utilities.
func (c *Cleaner) Status() *Status {
	if unsafe.Sizeof(Status{}) != unsafe.Sizeof(Cleaner{}) ||
		unsafe.Offsetof(Status{}.Cleaner) != 0 {
		panic(gtserror.New("compile time unsafe pointer assertion"))
	}
	return (*Status)(unsafe.Pointer(c))
}

// haveFiles returns whether all of the provided files exist within current storage.
func (c *Cleaner) haveFiles(ctx context.Context, files ...string) (bool, error) {
	for _, path := range files {
		if path == "" {
			// File not stored.
			return false, nil
		}

		// Check whether each file exists in storage.
		have, err := c.state.Storage.Has(ctx, path)
		if err != nil {
			return false, gtserror.Newf("error checking storage for %s: %w", path, err)
		}

		if !have {
			// Missing file(s).
			return false, nil
		}
	}
	return true, nil
}

// removeFiles removes the provided files, returning the number of them returned.
func (c *Cleaner) removeFiles(ctx context.Context, files ...string) {
	if gtscontext.DryRun(ctx) {
		// Dry run,
		// do nothing.
		return
	}

	for _, path := range files {
		if path == "" {
			// not stored.
			continue
		}

		// Remove each provided storage path.
		log.Debugf(ctx, "removing file: %s", path)
		err := c.state.Storage.Delete(ctx, path)
		if err != nil && !storage.IsNotFound(err) {
			log.Errorf(ctx, "error removing %s: %v", path, err)
			continue
		}
	}
}

// ScheduleJobs schedules cleaning
// jobs using configured parameters.
//
// Returns an error if `MediaCleanupFrom`
// is not a valid format (hh:mm:ss).
func (c *Cleaner) ScheduleJobs() error {
	const hourMinute = "15:04"

	var (
		now            = time.Now()
		cleanupEvery   = config.GetMediaCleanupEvery()
		cleanupFromStr = config.GetMediaCleanupFrom()
	)

	// Parse cleanupFromStr as hh:mm.
	// Resulting time will be on 1 Jan year zero.
	//
	// TODO: make ConfigCleanupFrom() a wrapped time.Time{}
	// type to move parsing as a stage into the config package.
	cleanupFrom, err := time.Parse(hourMinute, cleanupFromStr)
	if err != nil {
		return gtserror.Newf(
			"error parsing '%s' in time format 'hh:mm': %w",
			cleanupFromStr, err,
		)
	}

	// Determine first cleanup date from now, that's in future.
	firstCleanupAt := firstAt(now, cleanupFrom, cleanupEvery)

	log.Infof(nil,
		"scheduling media clean to run every %s, starting from %s; next clean will run at %s",
		cleanupEvery, cleanupFromStr, firstCleanupAt,
	)

	// Schedule media cleaning at parsed schedule.
	if !c.state.Workers.Scheduler.AddRecurring(
		"@mediacleanup",
		firstCleanupAt,
		cleanupEvery,
		c.cleanMedia,
	) {
		panic("failed to schedule @mediacleanup")
	}

	// Schedule status cleaning at fixed schedule.
	if !c.state.Workers.Scheduler.AddRecurring(
		"@statuscleanup",
		firstAt(now, time.Time{}, 24*time.Hour),
		24*time.Hour,
		c.cleanStatuses,
	) {
		panic("failed to schedule @statuscleanup")
	}

	return nil
}

func (c *Cleaner) cleanMedia(ctx context.Context, start time.Time) {
	log.Info(ctx, "starting")
	c.Media().All(ctx, config.GetMediaRemoteCacheDays())
	c.Emoji().All(ctx, config.GetMediaRemoteCacheDays())
	log.Infof(ctx, "finished after %s", time.Since(start))
}

func (c *Cleaner) cleanStatuses(ctx context.Context, start time.Time) {
	log.Info(ctx, "starting")
	c.Status().All(ctx, 7)
	log.Infof(ctx, "finished after %s", time.Since(start))
}

func firstAt(now, atHourMin time.Time, every time.Duration) time.Time {
	firstAt := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		atHourMin.Hour(),
		atHourMin.Minute(),
		0,
		0,
		now.Location(),
	)
	for firstAt.Before(now) {
		firstAt = firstAt.Add(every)
	}
	return firstAt
}
