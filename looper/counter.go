// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

// Counter combines an integer with a maximum value.
// It supports iteration tracking within looper.
type Counter struct {

	// Cur is the current counter value.
	Cur int

	// Max is the maximum counter value.
	// Only used if > 0 ([Loop] requires an IsDone condition to stop).
	Max int

	// Inc is the increment per iteration.
	Inc int
}

// SetMaxIncr sets the given Max and Inc value for the counter.
func (ct *Counter) SetMaxInc(mx, inc int) {
	ct.Max = mx
	ct.Inc = inc
}

// Incr increments the counter by Inc. Does not interact with Max.
func (ct *Counter) Incr() {
	ct.Cur += ct.Inc
}

// SkipToMax sets the counter to its Max value,
// for skipping over rest of loop iterations.
func (ct *Counter) SkipToMax() {
	ct.Cur = ct.Max
}

// IsOverMax returns true if counter is at or over Max (only if Max > 0).
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
