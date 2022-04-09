// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import "github.com/emer/emergent/etime"

// Loop represents one loop
type Loop struct {
	Scope etime.ScopeKey `desc:"level of this loop"`
	Start Funcs          `desc:"functions to call at start of loop -- in general it is best to avoid start functions where possible, to ensure reliable stepping behavior, so it picks up where it left off before"`
	Pre   Funcs          `desc:"functions to call inside each iteration, prior to looping at lower level -- in general it is best to avoid pre functions where possible, to ensure reliable stepping behavior, so it picks up where it left off before"`
	Post  Funcs          `desc:"functions to call inside each iteration, after looping at lower level -- any counters should be incremented here"`
	Stop  BoolFuncs      `desc:"functions that cause the loop to stop -- if any return true, it stops"`
	End   Funcs          `desc:"functions to run at the end of the loop, after it has stopped"`
}

func NewLoop(sc etime.ScopeKey) *Loop {
	return &Loop{Scope: sc}
}
