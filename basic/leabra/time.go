// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

// leabra.Time contains all the timing state and parameter information for running a model
type Time struct {
	Time      float32 `desc:"accumulated amount of time the network has been running, in simulation-time (not real world time), in seconds"`
	Cycle     int     `desc:"cycle counter: number of iterations of activation updating (settling) on the current alpha-cycle (100 msec / 10 Hz) trial -- this counts time sequentially through the entire trial, typically from 0 to 99 cycles"`
	CycleTot  int     `desc:"total cycle count -- this increments continuously from whenever it was last reset -- typically this is number of milliseconds in simulation time"`
	Quarter   int     `desc:"current gamma-frequency (25 msec / 40 Hz) quarter of alpha-cycle (100 msec / 10 Hz) trial being processed"`
	PlusPhase bool    `desc:"true if this is the plus phase (final quarter) -- else minus phase"`
	CycPerQtr int     `def:"25" desc:"number of cycles per quarter to run -- 25 = standard 100 msec alpha-cycle"`
}
