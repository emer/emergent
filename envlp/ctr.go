// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package envlp

import (
	"github.com/emer/emergent/estats"
	"github.com/emer/emergent/etime"
)

// Ctr is a counter that counts increments at a given time scale.
// It keeps track of the previous value.
type Ctr struct {
	Cur   int            `desc:"current counter value"`
	Prv   int            `view:"-" desc:"previous counter value, prior to last Incr() call (init to -1)"`
	Max   int            `desc:"maximum counter value -- only used if > 0"`
	Scope etime.ScopeKey `view:"-" desc:"the scope of this counter"`
}

// Init initializes counter -- Cur = 0, Prv = -1
func (ct *Ctr) Init() {
	ct.Prv = -1
	ct.Cur = 0
}

// Incr increments the counter by 1.
func (ct *Ctr) Incr() {
	ct.Prv = ct.Cur
	ct.Cur++
}

// IsOverMax returns true if counter is at or over Max (only if Max > 0)
func (ct *Ctr) IsOverMax() bool {
	return ct.Max > 0 && ct.Cur >= ct.Max
}

// ResetIfOverMax resets the current counter value to 0,
// if counter is at or over Max (only if Max > 0).
// returns true if reset
func (ct *Ctr) ResetIfOverMax() bool {
	if ct.IsOverMax() {
		ct.Set(0)
		return true
	}
	return false
}

// Set sets the Cur value if different from Cur, while preserving previous value.
// Returns true if changed
func (ct *Ctr) Set(cur int) bool {
	if ct.Cur == cur {
		return false
	}
	ct.Prv = ct.Cur
	ct.Cur = cur
	return true
}

///////////////////////////////////////////////////////////////////////
// Ctrs

// Ctrs is a map of counters by scope, used to manage counters in the Env.
type Ctrs map[etime.ScopeKey]*Ctr

// NewCtrs returns a new Ctrs map based on times and given mode
func NewCtrs(mode string, times ...etime.Times) Ctrs {
	ord := make([]etime.ScopeKey, len(times))
	for i, t := range times {
		ord[i] = etime.ScopeStr(mode, t.String())
	}
	return NewCtrsScope(ord...)
}

// NewCtrsScope returns a new Ctrs map based on scopes
func NewCtrsScope(scopes ...etime.ScopeKey) Ctrs {
	ct := make(map[etime.ScopeKey]*Ctr, len(scopes))
	for _, sc := range scopes {
		ct[sc] = &Ctr{}
	}
	return ct
}

// Init does Init on all the counters
func (cs *Ctrs) Init() {
	for _, ct := range *cs {
		ct.Init()
	}
}

// CtrsToStats sets the current counter values to estats Int values
// by their time names only (no eval Mode).
func (cs *Ctrs) CtrsToStats(stats *estats.Stats) {
	for _, ct := range *cs {
		_, tm := ct.Scope.ModesAndTimes()
		stats.SetInt(tm[0], ct.Cur)
	}
}
