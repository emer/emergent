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
	Stacks   map[string]*Stack `desc:"the collection of loop stacks -- key is typically etime.Mode"`
	StopFlag bool              `desc:"if true, running will stop at soonest opportunity"`
}

func NewSet() *Set {
	set := &Set{}
	set.InitMap()
	return set
}

func (set *Set) InitMap() {
	if set.Stacks == nil {
		set.Stacks = make(map[string]*Stack)
	}
}

func (set *Set) AddStack(st *Stack) {
	set.InitMap()
	set.Stacks[st.Mode] = st
	st.Set = set
}

// AddLevels adds given levels to all Stacks.
// For algorithms to add mechanism inner loops.
func (set *Set) AddLevels(times ...etime.Times) {
	for _, st := range set.Stacks {
		st.AddLevels(times...)
	}
}

// Stack returns Stack defined by given mode
func (set *Set) Stack(mode etime.Modes) *Stack {
	return set.Stacks[mode.String()]
}

// StackTry returns Stack defined by given mode, returning err if not found
func (set *Set) StackTry(mode etime.Modes) (*Stack, error) {
	return set.StackNameTry(mode.String())
}

// StackNameTry returns Stack based on name key, returning err if not found
func (set *Set) StackNameTry(name string) (*Stack, error) {
	st, ok := set.Stacks[name]
	if !ok {
		err := fmt.Errorf("Set StackNameTry: name: %s not found", name)
		log.Println(err)
		return nil, err
	}
	return st, nil
}

// InitAll runs Init on all Stacks
func (set *Set) InitAll() {
	for _, st := range set.Stacks {
		st.Init()
	}
}

// Init initializes Stack defined by given mode
func (set *Set) Init(mode etime.Modes) {
	set.InitName(mode.String())
}

// InitName initializes Stack of given name
func (set *Set) InitName(name string) (*Stack, error) {
	set.StopFlag = false
	st, err := set.StackNameTry(name)
	if err != nil {
		return st, err
	}
	st.Init()
	return st, err
}

// Run runs Stack defined by given mode
func (set *Set) Run(mode etime.Modes) {
	set.RunName(mode.String())
}

// RunName runs Stack of given name
func (set *Set) RunName(name string) (*Stack, error) {
	set.StopFlag = false
	st, err := set.StackNameTry(name)
	if err != nil {
		return st, err
	}
	st.Run()
	return st, err
}

// Step Steps Stack defined by given mode, at given step level,
// Stepping n times (n = 0 turns off stepping)
func (set *Set) Step(mode etime.Modes, step etime.Times, n int) (*Stack, error) {
	return set.StepName(mode.String(), etime.Scope(mode, step), n)
}

// StepName Steps Stack of given name, at given step level,
// Stepping n times (n = 0 turns off stepping)
func (set *Set) StepName(name string, step etime.ScopeKey, n int) (*Stack, error) {
	set.StopFlag = false
	st, err := set.StackNameTry(name)
	if err != nil {
		return st, err
	}
	st.SetStepScope(step, n)
	st.Run()
	return st, err
}
