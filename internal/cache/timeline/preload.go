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

package timeline

import (
	"errors"
	"sync"
	"sync/atomic"

	"code.superseriousbusiness.org/gopkg/log"
)

// preloader provides a means of synchronising the
// initial fill, or "preload", of a timeline cache.
// it has 4 possible states in the atomic pointer:
// - state changing = &(interface{}(*sync.WaitGroup))
// - preloaded      = &(interface{}(nil))
// - needs preload  = &(interface{}(false)) | (*interface{})(nil)
type preloader struct{ p atomic.Pointer[any] }

// needs_preload returns whether loaded preload state
// pointer indicates that preloader requires preload.
func needs_preload(ptr *any) bool { // nolint:revive
	return ptr == nil || (*ptr) == false
}

// Check will return the current preload state,
// waiting if a preload is currently in progress.
func (p *preloader) Check() bool {
	for {

		// Get state ptr.
		ptr := p.p.Load()

		// Check if
		// requires preloading.
		if needs_preload(ptr) {
			return false
		}

		// Check for a preload currently in progress.
		if wg, _ := (*ptr).(*sync.WaitGroup); wg != nil {
			wg.Wait()
			continue
		}

		// Anything else
		// means success.
		return true
	}
}

// CheckPreload will safely check the preload state,
// and if needed call the provided function. if a
// preload is in progress, it will wait until complete.
func (p *preloader) CheckPreload(preload func() error) error {
	for {

		// Get state ptr.
		ptr := p.p.Load()

		// Check if
		// requires preloading.
		if needs_preload(ptr) {

			// Needs preloading, start it.
			ok, err := p.do(ptr, preload)
			if !ok {

				// Failed to acquire preload,
				// other thread beat us to it.
				continue
			}

			// We ran!
			return err
		}

		// Check for a preload currently in progress.
		if wg, _ := (*ptr).(*sync.WaitGroup); wg != nil {
			wg.Wait()
			continue
		}

		// Anything else means
		// already preloaded.
		return nil
	}
}

// errSentinel error type checked and returned in Clear().
var errSentinel = errors.New("timeline cleared")

// Clear will perform the given clear function
// within the protection of the preloader mechanism.
func (p *preloader) Clear(clear func()) {
	for {

		// Get state ptr.
		ptr := p.p.Load()

		// Check if requires preload,
		// if so then nothing to clear.
		if needs_preload(ptr) {
			return
		}

		// Check for a preload currently in progress.
		if wg, _ := (*ptr).(*sync.WaitGroup); wg != nil {
			wg.Wait()
			continue
		}

		// We always return an error, to indicate
		// to the preloader that "preload" failed
		// and the next call needs to do preload.
		ok, err := p.do(ptr, func() (err error) {
			err = errSentinel

			// Call clear function
			// to safely drop cache.
			clear()
			return
		})
		if !ok {

			// Failed to acquire start,
			// other thread beat us to it.
			continue
		}

		if err != errSentinel {
			log.Errorf(nil, "BUG: invalid preloader state: %#v", (*p.p.Load()))
		}

		// We ran!
		return
	}
}

// do will attempt to acquire do state of the preloader, on success calling 'preload'.
// this returns whether do was acquired, and if called returns 'preload' error. in the
// case that 'do' is called, returned error determines the next state that preloader
// will update itself to. (err == nil) => "preloaded", (err != nil) => "needs preload".
// NOTE: this is the only function that may unset an in-progress sync.WaitGroup value.
func (p *preloader) do(old *any, do func() error) (done bool, err error) {

	// Optimistically setup a
	// new waitgroup to set as
	// the preload waiter.
	var wg sync.WaitGroup
	wg.Add(1)

	// Wrap waitgroup in
	// 'any' for pointer.
	a := any(&wg)
	ptr := &a

	// Attempt CAS operation to claim start.
	done = p.p.CompareAndSwap(old, ptr)
	if !done {
		return false, nil
	}

	panic := true
	defer func() {
		// Release.
		wg.Done()

		var ok bool
		if err != nil || panic {
			// Preload failed,
			// drop waiter ptr.
			a := any(false)
			ok = p.p.CompareAndSwap(ptr, &a)
		} else {
			// Preload success, set success value.
			ok = p.p.CompareAndSwap(ptr, new(any))
		}

		if !ok {
			log.Errorf(nil, "BUG: invalid preloader state: %#v", (*p.p.Load()))
		}
	}()

	// Perform preload.
	err = do()
	panic = false
	return
}
