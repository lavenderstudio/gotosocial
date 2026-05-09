package sched

import (
	"time"
)

var (
	// zerotime is zero
	// time.Time (unix epoch).
	zerotime time.Time
)

// Timing provides scheduling for a Job, determining the next time
// for given current time that execution is required. Please note that
// calls to .Next() may alter the results of the next call, and should
// only be called by the Scheduler.
type Timing interface {
	Next(time.Time) time.Time
}

// immediately is a default Timing implementation
// that fires once immediately, then never again.
type immediately struct{ done bool }

func (t *immediately) Next(now time.Time) time.Time {
	if t.done {
		return zerotime
	}
	t.done = true
	return now
}

// Once implements Timing to
// provide a run-once Job execution.
type Once time.Time

func (o *Once) Next(time.Time) time.Time {
	ret := *(*time.Time)(o)
	*o = Once(zerotime) // reset
	return ret
}

// Periodic implements Timing to
// provide a recurring Job execution.
type Periodic time.Duration

func (p Periodic) Next(now time.Time) time.Time {
	return now.Add(time.Duration(p))
}

// PeriodicAt implements Timing to provide a
// recurring Job execution starting at 'Once' time.
type PeriodicAt struct {
	Once   Once
	Period Periodic
}

func (p *PeriodicAt) Next(now time.Time) time.Time {
	if next := p.Once.Next(now); !next.IsZero() {
		return next
	}
	return p.Period.Next(now)
}

// TimingWrap allows combining two
// different Timing implementations.
type TimingWrap struct {
	Outer Timing
	Inner Timing

	// determined next times
	outerNext time.Time
	innerNext time.Time
}

func (t *TimingWrap) Next(now time.Time) time.Time {
	if t.outerNext.IsZero() {
		// Regenerate outermost next run time
		t.outerNext = t.Outer.Next(now)
	}

	if t.innerNext.IsZero() {
		// Regenerate innermost next run time
		t.innerNext = t.Inner.Next(now)
	}

	// If outer comes before inner, return outer
	if t.outerNext != zerotime &&
		t.outerNext.Before(t.innerNext) {
		next := t.outerNext
		t.outerNext = zerotime
		return next
	}

	// Else, return inner
	next := t.innerNext
	t.innerNext = zerotime
	return next
}
