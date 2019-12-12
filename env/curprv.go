// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env

// CurPrvF32 is basic state management for current and previous values, float32 values
type CurPrvF32 struct {
	Cur float32 `desc:"current value"`
	Prv float32 `desc:"previous value"`
}

// Set sets the new current value, copying Cur to Prv
func (cv *CurPrvF32) Set(cur float32) {
	cv.Prv = cv.Cur
	cv.Cur = cur
}

// Incr increments Cur by 1
func (cv *CurPrvF32) Incr() {
	cv.Prv = cv.Cur
	cv.Cur += 1.0
}

// Diff returns the difference between current and previous values
func (cv *CurPrvF32) Diff() float32 {
	return cv.Cur - cv.Prv
}

//////////////////////////////
// Int

// CurPrvInt is basic state management for current and previous values, int values
type CurPrvInt struct {
	Cur int `desc:"current value"`
	Prv int `desc:"previous value"`
}

// Set sets the new current value, copying Cur to Prv
func (cv *CurPrvInt) Set(cur int) {
	cv.Prv = cv.Cur
	cv.Cur = cur
}

// Incr increments Cur by 1
func (cv *CurPrvInt) Incr() {
	cv.Prv = cv.Cur
	cv.Cur++
}

// Diff returns the difference between current and previous values
func (cv *CurPrvInt) Diff() int {
	return cv.Cur - cv.Prv
}

//////////////////////////////
// String

// CurPrvString is basic state management for current and previous values, string values
type CurPrvString struct {
	Cur string `desc:"current value"`
	Prv string `desc:"previous value"`
}

// Set sets the new current value, copying Cur to Prv
func (cv *CurPrvString) Set(cur string) {
	cv.Prv = cv.Cur
	cv.Cur = cur
}
