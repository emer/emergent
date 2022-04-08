// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import (
	"fmt"
	"log"

	"github.com/emer/emergent/etime"
)

// Set contains a set of interconnected loop Stacks (e.g., Train, Test, etc)
type Set struct {
	Stacks   map[etime.ScopeKey]*Stack `desc:"the collection of loop stacks"`
	StopFlag bool                      `desc:"if true, running will stop at soonest opportunity"`
}

func NewSet() *Set {
	set := &Set{}
	set.Stacks = make(map[etime.ScopeKey]*Stack)
	return set
}

func (set *Set) AddStack(st *Stack) {
	set.Stacks[st.Scope] = st
}

func (set *Set) Run(mode etime.EvalModes, time etime.Times) {
	set.RunScope(etime.Scope(mode, time))
}

func (set *Set) RunScope(scope etime.ScopeKey) error {
	set.StopFlag = false
	st, ok := set.Stacks[scope]
	if !ok {
		err := fmt.Errorf("RunScope: scope: %s not found", scope)
		log.Println(err)
		return err
	}
	st.Run(set)
	return nil
}
