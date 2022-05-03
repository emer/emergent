// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import "strings"

type NamedFunc struct {
	Name string
	Func func()
}

type NamedFuncs []NamedFunc

func (funcs *NamedFuncs) Add(name string, fun func()) *NamedFuncs {
	*funcs = append(*funcs, NamedFunc{Name: name, Func: fun})
	return funcs
}

func (funcs NamedFuncs) String() string {
	s := ""
	if len(funcs) > 0 {
		for _, f := range funcs {
			s = s + f.Name + " "
		}
	}
	return s
}

func (funcs NamedFuncs) HasNameLike(nameSubstring string) bool {
	for _, nf := range funcs {
		if strings.Contains(nf.Name, nameSubstring) {
			return true
		}
	}
	return false
}
