// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env

// Counter maintains a current and previous counter value,
// and a Max value with methods to manage.
type Counter struct {

	// Cur is the current counter value.
	Cur int

	// Prev previous counter value, prior to last Incr() call (init to -1)
	Prev int `display:"-"`

	// Changed reports if it changed on the last Step() call or not.
	Changed bool `display:"-"`

	// Max is the maximum counter value, above which the counter will reset back to 0.
	// Only used if > 0.
	Max int
}

// Init initializes counter: Cur = 0, Prev = -1
func (ct *Counter) Init() {
	ct.Prev = -1
	ct.Cur = 0
	ct.Changed = false
}

// Same resets Changed = false -- good idea to call this on all counters at start of Step
// or can put in an else statement, but that is more error-prone.
func (ct *Counter) Same() {
	ct.Changed = false
}

// Incr increments the counter by 1. If Max > 0 then if Incr >= Max
// the counter is reset to 0 and true is returned. Otherwise false.
func (ct *Counter) Incr() bool {
	ct.Changed = true
	ct.Prev = ct.Cur
	ct.Cur++
	if ct.Max > 0 && ct.Cur >= ct.Max {
		ct.Cur = 0
		return true
	}
	return false
}

// Set sets the Cur value if different from Cur, while preserving previous value
// and setting Changed appropriately.  Returns true if changed.
// does NOT check Cur vs. Max.
func (ct *Counter) Set(cur int) bool {
	if ct.Cur == cur {
		ct.Changed = false
		return false
	}
	ct.Changed = true
	ct.Prev = ct.Cur
	ct.Cur = cur
	return true
}

// Query returns the current, previous and changed values for this counter
func (ct *Counter) Query() (cur, prev int, chg bool) {
	return ct.Cur, ct.Prev, ct.Changed
}
