// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

// Funcs is a list of plain functions
type Funcs []func()

// Add adds a function to the list
func (fn *Funcs) Add(f func()) {
	*fn = append(*fn, f)
}

// Run runs the list of functions in order
func (fn *Funcs) Run() {
	for _, f := range *fn {
		f()
	}
}

///////////////////////////////////////////////////////

// BoolFuncs is a list of bool functions
type BoolFuncs []func() bool

// Add adds a function to the list
func (fn *BoolFuncs) Add(f func() bool) {
	*fn = append(*fn, f)
}

// Run runs the list of functions in order, returning true
// as soon as any of the functions return true, else false
func (fn *BoolFuncs) Run() bool {
	for _, f := range *fn {
		bv := f()
		if bv {
			return true
		}
	}
	return false
}
