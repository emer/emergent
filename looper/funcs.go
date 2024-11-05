// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import (
	"fmt"
	"slices"

	"cogentcore.org/core/base/errors"
)

// NamedFunc is a function closure with a name.
// Function returns a bool which is needed for stopping condition
// but is otherwise not used.
type NamedFunc struct {
	Name string
	Func func() bool
}

// NamedFuncs is an ordered list of named functions.
type NamedFuncs []NamedFunc

// Add adds a named function (with no bool return value).
func (funcs *NamedFuncs) Add(name string, fun func()) *NamedFuncs {
	*funcs = append(*funcs, NamedFunc{Name: name, Func: func() bool { fun(); return true }})
	return funcs
}

// AddBool adds a named function with a bool return value, for IsDone case.
func (funcs *NamedFuncs) AddBool(name string, fun func() bool) *NamedFuncs {
	*funcs = append(*funcs, NamedFunc{Name: name, Func: fun})
	return funcs
}

// String prints the list of named functions.
func (funcs *NamedFuncs) String() string {
	s := ""
	for _, f := range *funcs {
		s = s + f.Name + " "
	}
	return s
}

// Run runs all of the functions, returning true if any of
// the functions returned true.
func (funcs NamedFuncs) Run() bool {
	ret := false
	for _, fn := range funcs {
		r := fn.Func()
		if r {
			ret = true
		}
	}
	return ret
}

// FuncIndex finds index of function by name.
// Returns not found err message if not found.
func (funcs *NamedFuncs) FuncIndex(name string) (int, error) {
	for i, fn := range *funcs {
		if fn.Name == name {
			return i, nil
		}
	}
	err := fmt.Errorf("looper.Funcs:FuncIndex: named function %s not found", name)
	return -1, err
}

// InsertAt inserts function at given index.
func (funcs *NamedFuncs) InsertAt(i int, name string, fun func() bool) {
	*funcs = slices.Insert(*funcs, i, NamedFunc{Name: name, Func: fun})
}

// Prepend adds a function to the start of the list.
func (funcs *NamedFuncs) Prepend(name string, fun func() bool) {
	funcs.InsertAt(0, name, fun)
}

// InsertBefore inserts function before other function of given name.
func (funcs *NamedFuncs) InsertBefore(before, name string, fun func() bool) error {
	i, err := funcs.FuncIndex(before)
	if errors.Log(err) != nil {
		return err
	}
	funcs.InsertAt(i, name, fun)
	return nil
}

// InsertAfter inserts function after other function of given name.
func (funcs *NamedFuncs) InsertAfter(after, name string, fun func() bool) error {
	i, err := funcs.FuncIndex(after)
	if errors.Log(err) != nil {
		return err
	}
	funcs.InsertAt(i+1, name, fun)
	return nil
}

// Replace replaces function with other function of given name.
func (funcs *NamedFuncs) Replace(name string, fun func() bool) error {
	i, err := funcs.FuncIndex(name)
	if errors.Log(err) != nil {
		return err
	}
	(*funcs)[i] = NamedFunc{Name: name, Func: fun}
	return nil
}

// Delete deletes function of given name.
func (funcs *NamedFuncs) Delete(name string) error {
	i, err := funcs.FuncIndex(name)
	if errors.Log(err) != nil {
		return err
	}
	*funcs = slices.Delete(*funcs, i, i+1)
	return nil
}
