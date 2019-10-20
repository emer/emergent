// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env

// CurPrvF32 is basic state management for current and previous values, float32 values
type CurPrvF32 struct {
	Cur float32 `desc:"current value"`
	Prv float32 `desc:"previous value"`
}

// Update updates the new current value, copying Cur to Prv
func (cv *CurPrvF32) Update(cur float32) {
	cv.Prv = cv.Cur
	cv.Cur = cur
}

// Diff returns the difference between current and previous values
func (cv *CurPrvF32) Diff() float32 {
	return cv.Cur - cv.Prv
}

// CurPrvInt is basic state management for current and previous values, int values
type CurPrvInt struct {
	Cur int `desc:"current value"`
	Prv int `desc:"previous value"`
}

// Update updates the new current value, copying Cur to Prv
func (cv *CurPrvInt) Update(cur int) {
	cv.Prv = cv.Cur
	cv.Cur = cur
}

// Diff returns the difference between current and previous values
func (cv *CurPrvInt) Diff() int {
	return cv.Cur - cv.Prv
}
