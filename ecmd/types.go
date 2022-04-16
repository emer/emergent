// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ecmd

// Int represents a int valued arg
type Int struct {
	Name string `desc:"name of arg -- must be unique"`
	Desc string `desc:"description of arg"`
	Val  int    `desc:"value as parsed"`
	Def  int    `desc:"default initial value"`
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
	Name string `desc:"name of arg -- must be unique"`
	Desc string `desc:"description of arg"`
	Val  bool   `desc:"value as parsed"`
	Def  bool   `desc:"default initial value"`
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
	Name string `desc:"name of arg -- must be unique"`
	Desc string `desc:"description of arg"`
	Val  string `desc:"value as parsed"`
	Def  string `desc:"default initial value"`
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
	Name string  `desc:"name of arg -- must be unique"`
	Desc string  `desc:"description of arg"`
	Val  float64 `desc:"value as parsed"`
	Def  float64 `desc:"default initial value"`
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
