// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env

import "github.com/emer/emergent/etime"

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

// OverMax counter is over Max (only if Max > 0)
func (ct *Ctr) OverMax() bool {
	return ct.Max > 0 && ct.Cur >= ct.Max
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

////////////////////////////////
// Ctrs

// Ctrs is a map of counters by scope -- easiest way to manage counters
type Ctrs map[time.ScopeKey]*Ctr

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
	ct := make(map[time.ScopeKey]*Ctr, len(scopes))
	for _, sc := range scopes {
		ct[sc] = &Ctr{}
	}
	return ct
}
