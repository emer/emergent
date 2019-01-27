// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

// leabra.Time contains all the timing state and parameter information for running a model
type Time struct {
	Time      float32 `desc:"accumulated amount of time the network has been running, in simulation-time (not real world time), in seconds"`
	Cycle     int     `desc:"cycle counter: number of iterations of activation updating (settling) on the current alpha-cycle (100 msec / 10 Hz) trial -- this counts time sequentially through the entire trial, typically from 0 to 99 cycles"`
	CycleTot  int     `desc:"total cycle count -- this increments continuously from whenever it was last reset -- typically this is number of milliseconds in simulation time"`
	Quarter   int     `desc:"[0-3] current gamma-frequency (25 msec / 40 Hz) quarter of alpha-cycle (100 msec / 10 Hz) trial being processed"`
	PlusPhase bool    `desc:"true if this is the plus phase (final quarter) -- else minus phase"`

	TimePerCyc float32 `def:"0.001" desc:"amount of time to increment per cycle"`
	CycPerQtr  int     `def:"25" desc:"number of cycles per quarter to run -- 25 = standard 100 msec alpha-cycle"`
}

// NewTime returns a new Time struct with default parameters
func NewTime() *Time {
	tm := &Time{}
	tm.Defaults()
	return tm
}

// Defaults sets default values
func (tm *Time) Defaults() {
	tm.TimePerCyc = 0.001
	tm.CycPerQtr = 25
}

// Reset resets the counters all back to zero
func (tm *Time) Reset() {
	tm.Time = 0
	tm.Cycle = 0
	tm.CycleTot = 0
	tm.Quarter = 0
	tm.PlusPhase = false
}

// TrialStart starts a new alpha-trial (set of 4 quarters)
func (tm *Time) TrialStart() {
	tm.Cycle = 0
	tm.Quarter = 0
}

// CycleInc increments at the cycle level
func (tm *Time) CycleInc() {
	tm.Cycle++
	tm.CycleTot++
	tm.Time += tm.TimePerCyc
}

// QuarterInc increments at the quarter level, updating Quarter and PlusPhase
func (tm *Time) QuarterInc() {
	tm.Quarter++
	if tm.Quarter == 3 {
		tm.PlusPhase = true
	} else {
		tm.PlusPhase = false
	}
}
