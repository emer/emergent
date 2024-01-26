// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ecmd

import (
	"flag"
	"fmt"
)

// Args provides maps for storing commandline args.
type Args struct {
	Ints    map[string]*Int
	Bools   map[string]*Bool
	Strings map[string]*String
	Floats  map[string]*Float

	// true when all args have been set to flag package
	Flagged bool `edit:"-"`
}

// Init must be called before use to create all the maps
func (ar *Args) Init() {
	ar.Ints = make(map[string]*Int)
	ar.Bools = make(map[string]*Bool)
	ar.Strings = make(map[string]*String)
	ar.Floats = make(map[string]*Float)
}

// AddInt adds a new Int arg
func (ar *Args) AddInt(name string, def int, desc string) {
	ar.Ints[name] = NewInt(name, def, desc)
}

// AddBool adds a new Bool arg
func (ar *Args) AddBool(name string, def bool, desc string) {
	ar.Bools[name] = NewBool(name, def, desc)
}

// AddString adds a new String arg
func (ar *Args) AddString(name string, def string, desc string) {
	ar.Strings[name] = NewString(name, def, desc)
}

// AddFloat adds a new Float arg
func (ar *Args) AddFloat(name string, def float64, desc string) {
	ar.Floats[name] = NewFloat(name, def, desc)
}

// Int returns int val by name
func (ar *Args) Int(name string) int {
	val, has := ar.Ints[name]
	if has {
		return val.Val
	}
	fmt.Printf("Arg named: %s not found in Args\n", name)
	return 0
}

// SetInt sets the default and current val
func (ar *Args) SetInt(name string, val int) {
	ar.Ints[name].Set(val)
}

// Bool returns bool val by name
func (ar *Args) Bool(name string) bool {
	val, has := ar.Bools[name]
	if has {
		return val.Val
	}
	fmt.Printf("Arg named: %s not found in Args\n", name)
	return false
}

// SetBool sets the default and current val
func (ar *Args) SetBool(name string, val bool) {
	ar.Bools[name].Set(val)
}

// String returns string val by name
func (ar *Args) String(name string) string {
	val, has := ar.Strings[name]
	if has {
		return val.Val
	}
	fmt.Printf("Arg named: %s not found in Args\n", name)
	return ""
}

// SetString sets the default and current val
func (ar *Args) SetString(name string, val string) {
	ar.Strings[name].Set(val)
}

// Float returns float val by name
func (ar *Args) Float(name string) float64 {
	val, has := ar.Floats[name]
	if has {
		return val.Val
	}
	fmt.Printf("Arg named: %s not found in Args\n", name)
	return 0
}

// SetFloat sets the default and current val
func (ar *Args) SetFloat(name string, val float64) {
	ar.Floats[name].Set(val)
}

// Flag sets all args to the system flag values, only if not already done.
func (ar *Args) Flag() {
	if ar.Flagged {
		return
	}
	for _, vl := range ar.Ints {
		if flag.Lookup(vl.Name) == nil {
			flag.IntVar(&vl.Val, vl.Name, vl.Def, vl.Desc)
		}
	}
	for _, vl := range ar.Bools {
		if flag.Lookup(vl.Name) == nil {
			flag.BoolVar(&vl.Val, vl.Name, vl.Def, vl.Desc)
		}
	}
	for _, vl := range ar.Strings {
		if flag.Lookup(vl.Name) == nil {
			flag.StringVar(&vl.Val, vl.Name, vl.Def, vl.Desc)
		}
	}
	for _, vl := range ar.Floats {
		if flag.Lookup(vl.Name) == nil {
			flag.Float64Var(&vl.Val, vl.Name, vl.Def, vl.Desc)
		}
	}
	ar.Flagged = true
}

// Parse parses command line args, setting values from command line
// Any errors will cause the program to exit with error message.
func (ar *Args) Parse() {
	ar.Flag()
	flag.Parse()
}

// Usage prints the set of command args.  It is called by the help bool arg.
func (ar *Args) Usage() {
	ar.Flag()
	flag.Usage()
}
