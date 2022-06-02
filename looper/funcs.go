// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import "strings"

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

// HasNameLike is a helper function to check if there's an existing function that contains a substring. This could be helpful to ensure that you don't add duplicate logic to a list of functions. If you plan on using this, add a comment documenting which name is important, because the default assumption is that names are just documentation.
func (funcs *NamedFuncs) HasNameLike(nameSubstring string) bool {
	for _, nf := range *funcs {
		if strings.Contains(nf.Name, nameSubstring) {
			return true
		}
	}
	return false
}

// NamedFuncsBool is like NamedFuncs, but for functions that return a bool.
type NamedFuncsBool map[string]func() bool

// Add adds a function by name.
func (funcs *NamedFuncsBool) Add(name string, f func() bool) {
	if funcs == nil {
		funcs = &NamedFuncsBool{}
	}
	(*funcs)[name] = f
}
