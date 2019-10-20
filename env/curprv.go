// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env

// CurPrv is basic state management for current and previous values
type CurPrv struct {
	Cur float32 `desc:"current value"`
	Prv float32 `desc:"previous value"`
}

// Update updates the new current value, copying Cur to Prv
func (cv *CurPrv) Update(cur float32) {
	cv.Prv = cv.Cur
	cv.Cur = cur
}

// Diff returns the difference between current and previous values
func (cv *CurPrv) Diff() float32 {
	return cv.Cur - cv.Prv
}
