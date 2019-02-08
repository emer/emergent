// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import "github.com/goki/ki/kit"

// leabra.Time contains all the timing state and parameter information for running a model
type Time struct {
	Time      float32 `desc:"accumulated amount of time the network has been running, in simulation-time (not real world time), in seconds"`
	Cycle     int     `desc:"cycle counter: number of iterations of activation updating (settling) on the current alpha-cycle (100 msec / 10 Hz) trial -- this counts time sequentially through the entire trial, typically from 0 to 99 cycles"`
	CycleTot  int     `desc:"total cycle count -- this increments continuously from whenever it was last reset -- typically this is number of milliseconds in simulation time"`
	Quarter   int     `desc:"[0-3] current gamma-frequency (25 msec / 40 Hz) quarter of alpha-cycle (100 msec / 10 Hz) trial being processed.  Due to 0-based indexing, the first quarter is 0, second is 1, etc -- the plus phase final quarter is 3."`
	PlusPhase bool    `desc:"true if this is the plus phase (final quarter = 3) -- else minus phase"`

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
	if tm.CycPerQtr == 0 {
		tm.Defaults()
	}
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

//////////////////////////////////////////////////////////////////////////////////////
//  Quarters

// Quarters are the different alpha trial quarters, as a bitflag, for use in relevant timing
// parameters where quarters need to be specified
type Quarters int32

//go:generate stringer -type=Quarters

var KiT_Quarters = kit.Enums.AddEnum(QuartersN, true, nil)

func (ev Quarters) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *Quarters) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }

// The quarters
const (
	// Q1 is the first quarter, which, due to 0-based indexing, shows up as Quarter = 0 in timer
	Q1 Quarters = iota
	Q2
	Q3
	Q4
	QuartersN
)
