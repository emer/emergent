// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import "github.com/emer/emergent/etime"

// Loop represents one loop
type Loop struct {
	Scope   etime.ScopeKey `desc:"level of this loop"`
	OnStart Funcs          `desc:"functions to call at start of loop"`
	RunPre  Funcs          `desc:"functions to call inside each iteration, prior to looping at lower level"`
	RunPost Funcs          `desc:"functions to call inside each iteration, after looping at lower level"`
	Stop    BoolFuncs      `desc:"functions that cause the loop to stop"`
	OnEnd   Funcs          `desc:"functions to run at the end of the loop"`
}

func NewLoop(sc etime.ScopeKey) *Loop {
	return &Loop{Scope: sc}
}
