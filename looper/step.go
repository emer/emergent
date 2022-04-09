// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import "github.com/emer/emergent/etime"

// Step manages stepping state
type Step struct {
	Scope etime.ScopeKey `desc:"stepping level"`
	N     int            `desc:"number of times to iterate at StepScope level, no stepping if 0"`
	Cnt   int            `desc:"counter for number of times through loop"`
}

// Set sets the stepping scope and n -- 0 = no stepping
// resets counter.
func (st *Step) Set(scope etime.ScopeKey, n int) {
	st.Scope = scope
	st.N = n
	st.Cnt = 0
}

// Clear resets stepping (sets N = 0)
func (st *Step) Clear() {
	st.N = 0
}

// StopCheck checks if it is time to stop for this scope
// returns true if so
func (st *Step) StopCheck(scope etime.ScopeKey) bool {
	if st.N <= 0 {
		return false
	}
	if st.Scope != scope {
		return false
	}
	st.Cnt++
	if st.Cnt >= st.N {
		st.Cnt = 0
		return true
	}
	return false
}
