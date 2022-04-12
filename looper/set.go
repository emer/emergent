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
	set.Stacks[st.Scope()] = st
	st.Set = set
}

// AddLevels adds given levels to all Stacks.
// For algorithms to add mechanism inner loops.
func (set *Set) AddLevels(times ...etime.Times) {
	for _, st := range set.Stacks {
		st.AddLevels(times...)
	}
}

// Stack returns Stack defined by given top-level scope
func (set *Set) Stack(mode etime.Modes, time etime.Times) *Stack {
	return set.StackScope(etime.Scope(mode, time))
}

// StackScope returns Stack defined by given top-level scope
func (set *Set) StackScope(scope etime.ScopeKey) *Stack {
	return set.Stacks[scope]
}

// StackTry returns Stack defined by given top-level scope
func (set *Set) StackTry(mode etime.Modes, time etime.Times) (*Stack, error) {
	return set.StackScopeTry(etime.Scope(mode, time))
}

// StackScopeTry returns Stack defined by given top-level scope
func (set *Set) StackScopeTry(scope etime.ScopeKey) (*Stack, error) {
	st, ok := set.Stacks[scope]
	if !ok {
		err := fmt.Errorf("Set StackScope: scope: %s not found", scope)
		log.Println(err)
		return nil, err
	}
	return st, nil
}

// Run Runs Stack defined by given top-level scope
func (set *Set) Run(mode etime.Modes, time etime.Times) {
	set.RunScope(etime.Scope(mode, time))
}

// RunScope Runs Stack defined by given top-level scope
func (set *Set) RunScope(scope etime.ScopeKey) (*Stack, error) {
	set.StopFlag = false
	st, err := set.StackScopeTry(scope)
	if err != nil {
		return st, err
	}
	st.Run()
	return st, err
}

// Step Steps Stack defined by given top-level scope, at given step level,
// Stepping n times (n = 0 turns off stepping)
func (set *Set) Step(mode etime.Modes, time etime.Times, step etime.Times, n int) (*Stack, error) {
	return set.StepScope(etime.Scope(mode, time), etime.Scope(mode, step), n)
}

// StepScope Steps Stack defined by given top-level scope, at given step level,
// Stepping n times (n = 0 turns off stepping)
func (set *Set) StepScope(scope, step etime.ScopeKey, n int) (*Stack, error) {
	set.StopFlag = false
	st, err := set.StackScopeTry(scope)
	if err != nil {
		return st, err
	}
	st.SetStepScope(step, n)
	st.Run()
	return st, err
}
