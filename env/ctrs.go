// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env

import (
	"fmt"

	"github.com/emer/emergent/v2/estats"
	"github.com/emer/emergent/v2/etime"
)

// Counters contains an ordered slice of timescales,
// and a lookup map of counters by timescale
// used to manage counters in the Env.
type Counters struct {

	// ordered list of the counter timescales, from outer-most (highest) to inner-most (lowest)
	Order []etime.Times

	// map of the counters by timescale
	Counters map[etime.Times]*Counter
}

// SetTimes initializes Counters for given mode
// and list of times ordered from highest to lowest
func (cs *Counters) SetTimes(mode string, times ...etime.Times) {
	cs.Order = times
	cs.Counters = make(map[etime.Times]*Counter, len(times))
	for _, tm := range times {
		cs.Counters[tm] = &Counter{Scale: tm}
	}
}

// ByTime returns counter by timescale key -- nil if not found
func (cs *Counters) ByScope(tm etime.Times) *Counter {
	return cs.Counters[tm]
}

// ByTimeTry returns counter by timescale key -- returns nil, error if not found
func (cs *Counters) ByTimeTry(tm etime.Times) (*Counter, error) {
	ct, ok := cs.Counters[tm]
	if ok {
		return ct, nil
	}
	err := fmt.Errorf("env.Counters: scope not found: %s", tm.String())
	return nil, err
}

// Init does Init on all the counters
func (cs *Counters) Init() {
	for _, ct := range cs.Counters {
		ct.Init()
	}
}

// CountersToStats sets the current counter values to estats Int values
// by their time names only (no eval Mode).
func (cs *Counters) CountersToStats(mode string, stats *estats.Stats) {
	for _, ct := range cs.Counters {
		tm := ct.Scale.String()
		stats.SetInt(mode+":"+tm, ct.Cur)
	}
}
