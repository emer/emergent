// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ecmd

// Int represents a int valued arg
type Int struct {

	// name of arg -- must be unique
	Name string

	// description of arg
	Desc string

	// value as parsed
	Val int

	// default initial value
	Def int
}

// NewInt returns a new Int arg
func NewInt(name string, def int, desc string) *Int {
	return &Int{Name: name, Desc: desc, Def: def}
}

// Set sets default and current val
func (vl *Int) Set(val int) {
	vl.Val = val
	vl.Def = val
}

// Bool represents a bool valued arg
type Bool struct {

	// name of arg -- must be unique
	Name string

	// description of arg
	Desc string

	// value as parsed
	Val bool

	// default initial value
	Def bool
}

// NewBool returns a new Bool arg
func NewBool(name string, def bool, desc string) *Bool {
	return &Bool{Name: name, Desc: desc, Val: def, Def: def}
}

// Set sets default and current val
func (vl *Bool) Set(val bool) {
	vl.Val = val
	vl.Def = val
}

// String represents a string valued arg
type String struct {

	// name of arg -- must be unique
	Name string

	// description of arg
	Desc string

	// value as parsed
	Val string

	// default initial value
	Def string
}

// NewString returns a new String arg
func NewString(name string, def string, desc string) *String {
	return &String{Name: name, Desc: desc, Val: def, Def: def}
}

// Set sets default and current val
func (vl *String) Set(val string) {
	vl.Val = val
	vl.Def = val
}

// Float represents a float64 valued arg
type Float struct {

	// name of arg -- must be unique
	Name string

	// description of arg
	Desc string

	// value as parsed
	Val float64

	// default initial value
	Def float64
}

// NewFloat returns a new Float arg
func NewFloat(name string, def float64, desc string) *Float {
	return &Float{Name: name, Desc: desc, Val: def, Def: def}
}

// Set sets default and current val
func (vl *Float) Set(val float64) {
	vl.Val = val
	vl.Def = val
}
