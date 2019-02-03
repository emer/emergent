// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package timer provides a simple wall-clock duration timer based on standard
// time.  Accumulates total and average over multiple Start / Stop intervals.
package timer

import "time"

// Time manages the timer accumulated time and count
type Time struct {
	St    time.Time     `desc:"the most recent starting time"`
	Total time.Duration `desc:"the total accumulated time"`
	N     int           `desc:"the number of start/stops"`
}

// Reset resets the overall accumulated Total and N counters
func (t *Time) Reset() {
	t.Total = 0
	t.N = 0
}

// Start starts the timer
func (t *Time) Start() {
	t.St = time.Now()
}

// Stop stops the timer and accumulates the latest start - stop interval, and also returns it
func (t *Time) Stop() time.Duration {
	iv := time.Now().Sub(t.St)
	t.Total += iv
	t.N++
	return iv
}

// Avg returns the average start / stop interval (assumes each was measuring the same thing).
func (t *Time) Avg() time.Duration {
	if t.N == 0 {
		return 0
	}
	return t.Total / time.Duration(t.N)
}

// AvgSecs returns the average start / stop interval (assumes each was measuring the same thing)
// as a float64 of seconds
func (t *Time) AvgSecs() float64 {
	if t.N == 0 {
		return 0
	}
	return float64(t.Total) / (float64(t.N) * float64(time.Second))
}

// TotalSecs returns the total start / stop intervals as a float64 of seconds.
func (t *Time) TotalSecs() float64 {
	return float64(t.Total) / float64(time.Second)
}
