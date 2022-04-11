// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package envlp

import (
	"fmt"

	"github.com/emer/emergent/estats"
	"github.com/emer/emergent/etime"
)

// Ctrs contains an ordered slice of scopes,
// and a lookup map of counters by scope,
// used to manage counters in the Env.
type Ctrs struct {
	Order []etime.ScopeKey        `desc:"ordered list of the counter scopes, from outer-most (highest) to inner-most (lowest)"`
	Ctrs  map[etime.ScopeKey]*Ctr `desc:"map of the counters by scope"`
}

// SetTimes initializes Ctrs for given mode
// and list of times ordered from highest to lowest
func (cs *Ctrs) SetTimes(mode string, times ...etime.Times) {
	ord := make([]etime.ScopeKey, len(times))
	for i, t := range times {
		ord[i] = etime.ScopeStr(mode, t.String())
	}
	cs.SetScopes(ord...)
}

// SetScopes initializes  returns a new Ctrs based on scopes
// ordered from highest to lowest
func (cs *Ctrs) SetScopes(scopes ...etime.ScopeKey) {
	cs.Order = etime.CloneScopeSlice(scopes)
	cs.Ctrs = make(map[etime.ScopeKey]*Ctr, len(scopes))
	for _, sc := range scopes {
		cs.Ctrs[sc] = &Ctr{}
	}
}

// ByScope returns counter by scope key -- nil if not found
func (cs *Ctrs) ByScope(scope etime.ScopeKey) *Ctr {
	return cs.Ctrs[scope]
}

// ByScopeTry returns counter by scope key -- returns nil, error if not found
func (cs *Ctrs) ByScopeTry(scope etime.ScopeKey) (*Ctr, error) {
	ct, ok := cs.Ctrs[scope]
	if ok {
		return ct, nil
	}
	err := fmt.Errorf("envlp.Ctrs: scope not found: %s", scope)
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
func (cs *Ctrs) CtrsToStats(stats *estats.Stats) {
	for _, ct := range cs.Ctrs {
		_, tm := ct.Scope.ModesAndTimes()
		stats.SetInt(tm[0], ct.Cur)
	}
}
