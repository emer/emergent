// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import (
	"fmt"
	"log"
	"strings"
)

// NamedFunc lets you keep an ordered map of functions.
type NamedFunc struct {
	Name string
	Func func()
}

// NamedFunc is an ordered map of functions.
type NamedFuncs []NamedFunc

// Add adds a named function to a list.
func (funcs *NamedFuncs) Add(name string, fun func()) *NamedFuncs {
	*funcs = append(*funcs, NamedFunc{Name: name, Func: fun})
	return funcs
}

// String describes named functions.
func (funcs *NamedFuncs) String() string {
	s := ""
	if len(*funcs) > 0 {
		for _, f := range *funcs {
			s = s + f.Name + " "
		}
	}
	return s
}

// HasNameLike checks if there's an existing function that contains a substring.
// This could be helpful to ensure that you don't add duplicate logic to a list
// of functions. If you plan on using this, add a comment documenting which name
// is important, because the default assumption is that names are just documentation.
func (funcs *NamedFuncs) HasNameLike(nameSubstring string) bool {
	for _, nf := range *funcs {
		if strings.Contains(nf.Name, nameSubstring) {
			return true
		}
	}
	return false
}

// FindName finds index of function by name, returns not found err message if not found
func (funcs *NamedFuncs) FindName(nm string) (int, error) {
	for i, fn := range *funcs {
		if fn.Name == nm {
			return i, nil
		}
	}
	err := fmt.Errorf("looper.Funcs:FindName: named function %s not found", nm)
	return -1, err
}

// Prepend adds a function to the start of the list
func (funcs *NamedFuncs) Prepend(nm string, f func()) {
	funcs.InsertAt(0, nm, f)
}

// InsertAt inserts function at given index
func (funcs *NamedFuncs) InsertAt(i int, nm string, f func()) {
	sz := len(*funcs)
	*funcs = append(*funcs, NamedFunc{})
	if i < sz {
		copy((*funcs)[i+1:], (*funcs)[i:sz])
	}
	(*funcs)[i] = NamedFunc{Name: nm, Func: f}
}

// InsertBefore inserts function before other function of given name
func (funcs *NamedFuncs) InsertBefore(before, nm string, f func()) error {
	i, err := funcs.FindName(before)
	if err != nil {
		err = fmt.Errorf("InsertBefore of %s: %s", nm, err)
		log.Println(err)
		return err
	}
	funcs.InsertAt(i, nm, f)
	return nil
}

// InsertAfter inserts function after other function of given name
func (funcs *NamedFuncs) InsertAfter(after, nm string, f func()) error {
	i, err := funcs.FindName(after)
	if err != nil {
		err = fmt.Errorf("InsertAfter of %s: %s", nm, err)
		log.Println(err)
		return err
	}
	funcs.InsertAt(i+1, nm, f)
	return nil
}

// Replace replaces function with other function of given name
func (funcs *NamedFuncs) Replace(nm string, f func()) error {
	i, err := funcs.FindName(nm)
	if err != nil {
		err = fmt.Errorf("Replace of %s: %s", nm, err)
		log.Println(err)
		return err
	}
	(*funcs)[i] = NamedFunc{Name: nm, Func: f}
	return nil
}

// Delete deletes function of given name
func (funcs *NamedFuncs) Delete(nm string) error {
	i, err := funcs.FindName(nm)
	if err != nil {
		err = fmt.Errorf("Delete: %s", err)
		log.Println(err)
		return err
	}
	sz := len(*funcs)
	copy((*funcs)[i:], (*funcs)[i+1:])
	(*funcs)[sz-1] = NamedFunc{}
	(*funcs) = (*funcs)[:sz-1]
	return nil
}

///////////////////////////////////////////////////////////////////
// NamedFuncsBool

// NamedFuncsBool is like NamedFuncs, but for functions that return a bool.
type NamedFuncsBool map[string]func() bool

// Add adds a function by name.
func (funcs *NamedFuncsBool) Add(name string, f func() bool) {
	if funcs == nil {
		funcs = &NamedFuncsBool{}
	}
	(*funcs)[name] = f
}
