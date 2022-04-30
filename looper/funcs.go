// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

type namedFunc struct {
	Name string
	Func func()
}

type orderedMapFuncs []namedFunc

func (funcs *orderedMapFuncs) Add(name string, fun func()) *orderedMapFuncs {
	*funcs = append(*funcs, namedFunc{Name: name, Func: fun})
	return funcs
}

func (funcs orderedMapFuncs) String() string {
	s := ""
	if len(funcs) > 0 {
		for _, f := range funcs {
			s = s + f.Name + " "
		}
	}
	return s
}
