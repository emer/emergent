// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env

import "github.com/emer/emergent/v2/etime"

// Counter is a counter that counts increments at a given time scale.
// It keeps track of when it has been incremented or not, and
// retains the previous value.
type Counter struct {

	// current counter value
	Cur int

	// previous counter value, prior to last Incr() call (init to -1)
	Prv int `display:"-"`

	// did this change on the last Step() call or not?
	Chg bool `display:"-"`

	// where relevant, this is a fixed maximum counter value, above which the counter will reset back to 0 -- only used if > 0
	Max int

	// the unit of time scale represented by this counter (just FYI)
	Scale etime.Times `display:"-"`
}

// Init initializes counter -- Cur = 0, Prv = -1
func (ct *Counter) Init() {
	ct.Prv = -1
	ct.Cur = 0
	ct.Chg = false
}

// Same resets Chg = false -- good idea to call this on all counters at start of Step
// or can put in an else statement, but that is more error-prone.
func (ct *Counter) Same() {
	ct.Chg = false
}

// Incr increments the counter by 1.  If Max > 0 then if Incr >= Max
// the counter is reset to 0 and true is returned.  Otherwise false.
func (ct *Counter) Incr() bool {
	ct.Chg = true
	ct.Prv = ct.Cur
	ct.Cur++
	if ct.Max > 0 && ct.Cur >= ct.Max {
		ct.Cur = 0
		return true
	}
	return false
}

// Set sets the Cur value if different from Cur, while preserving previous value
// and setting Chg appropriately.  Returns true if changed.
// does NOT check Cur vs. Max.
func (ct *Counter) Set(cur int) bool {
	if ct.Cur == cur {
		ct.Chg = false
		return false
	}
	ct.Chg = true
	ct.Prv = ct.Cur
	ct.Cur = cur
	return true
}

// Query returns the current, previous and changed values for this counter
func (ct *Counter) Query() (cur, prv int, chg bool) {
	return ct.Cur, ct.Prv, ct.Chg
}
