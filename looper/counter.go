// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

// Counter combines an integer with a maximum value. It supports time tracking within looper.
type Counter struct {

	// current counter value
	Cur int

	// maximum counter value -- only used if > 0
	Max int

	// increment per iteration
	Inc int
}

// Incr increments the counter by 1. Does not interact with Max.
func (ct *Counter) Incr() {
	ct.Cur += ct.Inc
}

// SkipToMax sets the counter to its Max value -- for skipping over rest of loop
func (ct *Counter) SkipToMax() {
	ct.Cur = ct.Max
}

// IsOverMax returns true if counter is at or over Max (only if Max > 0)
func (ct *Counter) IsOverMax() bool {
	return ct.Max > 0 && ct.Cur >= ct.Max
}

// Set sets the Cur value with return value indicating whether it is different
// from current Cur.
func (ct *Counter) Set(cur int) bool {
	if ct.Cur == cur {
		return false
	}
	ct.Cur = cur
	return true
}

// SetCurMax sets the Cur and Max values, as a convenience.
func (ct *Counter) SetCurMax(cur, max int) {
	ct.Cur = cur
	ct.Max = max
}

// SetCurMaxPlusN sets the Cur value and Max as Cur + N -- run N more beyond current.
func (ct *Counter) SetCurMaxPlusN(cur, n int) {
	ct.Cur = cur
	ct.Max = cur + n
}
