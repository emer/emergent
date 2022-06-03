// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env

import (
	"fmt"

	"github.com/emer/emergent/estats"
)

// Ctrs contains an ordered slice of timescales,
// and a lookup map of counters by timescale
// used to manage counters in the Env.
type Ctrs struct {
	Order []TimeScales        `desc:"ordered list of the counter timescales, from outer-most (highest) to inner-most (lowest)"`
	Ctrs  map[TimeScales]*Ctr `desc:"map of the counters by timescale"`
}

// SetTimes initializes Ctrs for given mode
// and list of times ordered from highest to lowest
func (cs *Ctrs) SetTimes(mode string, times ...TimeScales) {
	cs.Order = times
	cs.Ctrs = make(map[TimeScales]*Ctr, len(times))
	for _, tm := range times {
		cs.Ctrs[tm] = &Ctr{Scale: tm}
	}
}

// ByTime returns counter by timescale key -- nil if not found
func (cs *Ctrs) ByScope(tm TimeScales) *Ctr {
	return cs.Ctrs[tm]
}

// ByTimeTry returns counter by timescale key -- returns nil, error if not found
func (cs *Ctrs) ByTimeTry(tm TimeScales) (*Ctr, error) {
	ct, ok := cs.Ctrs[tm]
	if ok {
		return ct, nil
	}
	err := fmt.Errorf("env.Ctrs: scope not found: %s", tm.String())
	return nil, err
}

// Init does Init on all the counters
func (cs *Ctrs) Init() {
	for _, ct := range cs.Ctrs {
		ct.Init()
	}
}

// CtrsToStats sets the current counter values to estats Int values
// by their time names only (no eval Mode).
func (cs *Ctrs) CtrsToStats(mode string, stats *estats.Stats) {
	for _, ct := range cs.Ctrs {
		tm := ct.Scale.String()
		stats.SetInt(mode+":"+tm, ct.Cur)
	}
}
