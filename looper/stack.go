// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import (
	"fmt"
	"strings"

	"cogentcore.org/core/enums"
)

// Stack contains a list of Loops to run, for a given Mode of processing,
// which distinguishes this stack, and is its key in the map of Stacks.
// The order of Loop stacks is determined by the Order list of loop levels.
type Stack struct {

	// Mode identifies the mode of processing this stack performs, e.g., Train or Test.
	Mode enums.Enum

	// Loops is the set of Loops for this Stack, keyed by the level enum value.
	// Order is determined by the Order list.
	Loops map[enums.Enum]*Loop

	// Order is the list and order of levels looped over by this stack of loops.
	// The order is from top to bottom, so longer timescales like Run should be at
	// the start and shorter level timescales like Trial should be at the end.
	Order []enums.Enum

	// OnInit are functions to run when Init is called, to restart processing,
	// which also resets the counters for this stack.
	OnInit NamedFuncs

	// StopNext will stop running at the end of the current StopLevel if set.
	StopNext bool

	// StopFlag will stop running ASAP if set.
	StopFlag bool

	// StopLevel sets the level to stop at the end of.
	// This is the current active Step level, which will be reset when done.
	StopLevel enums.Enum

	// StopCount determines how many iterations at StopLevel before actually stopping.
	// This is the current active Step control value.
	StopCount int

	// StepLevel is a saved copy of StopLevel for stepping.
	// This is what was set for last Step call (which sets StopLevel) or by GUI.
	StepLevel enums.Enum

	// StepCount is a saved copy of StopCount for stepping.
	// This is what was set for last Step call (which sets StopCount) or by GUI.
	StepCount int
}

// NewStack returns a new Stack for given mode and default step level.
func NewStack(mode, stepLevel enums.Enum) *Stack {
	st := &Stack{}
	st.newInit(mode, stepLevel)
	return st
}

// newInit initializes new data structures for a newly created object.
func (st *Stack) newInit(mode, stepLevel enums.Enum) {
	st.Mode = mode
	st.StepLevel = stepLevel
	st.StepCount = 1
	st.Loops = map[enums.Enum]*Loop{}
	st.Order = []enums.Enum{}
}

// Level returns the [Loop] at the given ordinal level in the Order list.
// Will panic if out of range.
func (st *Stack) Level(i int) *Loop {
	return st.Loops[st.Order[i]]
}

// AddLevel adds a new level to this Stack with a given counterMax number of iterations.
// The order in which this method is invoked is important,
// as it adds loops in order from top to bottom.
// Sets a default increment of 1 for the counter -- see AddLevelIncr for different increment.
func (st *Stack) AddLevel(level enums.Enum, counterMax int) *Stack {
	st.Loops[level] = NewLoop(counterMax, 1)
	st.Order = append(st.Order, level)
	return st
}

// AddOnStartToAll adds given function taking mode and level args to OnStart in all loops.
func (st *Stack) AddOnStartToAll(name string, fun func(mode, level enums.Enum)) {
	for tt, lp := range st.Loops {
		lp.OnStart.Add(name, func() {
			fun(st.Mode, tt)
		})
	}
}

// AddOnEndToAll adds given function taking mode and level args to OnEnd in all loops.
func (st *Stack) AddOnEndToAll(name string, fun func(mode, level enums.Enum)) {
	for tt, lp := range st.Loops {
		lp.OnEnd.Add(name, func() {
			fun(st.Mode, tt)
		})
	}
}

// AddLevelIncr adds a new level to this Stack with a given counterMax
// number of iterations, and increment per step.
// The order in which this method is invoked is important,
// as it adds loops in order from top to bottom.
func (st *Stack) AddLevelIncr(level enums.Enum, counterMax, counterIncr int) *Stack {
	st.Loops[level] = NewLoop(counterMax, counterIncr)
	st.Order = append(st.Order, level)
	return st
}

// LevelAbove returns the level above the given level in the stack
// returning false if this is the highest level,
// or given level does not exist in order.
func (st *Stack) LevelAbove(level enums.Enum) (enums.Enum, bool) {
	for i, tt := range st.Order {
		if tt == level && i > 0 {
			return st.Order[i-1], true
		}
	}
	return level, false
}

// LevelBelow returns the level below the given level in the stack
// returning false if this is the lowest level,
// or given level does not exist in order.
func (st *Stack) LevelBelow(level enums.Enum) (enums.Enum, bool) {
	for i, tt := range st.Order {
		if tt == level && i+1 < len(st.Order) {
			return st.Order[i+1], true
		}
	}
	return level, false
}

//////// Control

// SetStep sets stepping to given level and number of iterations.
// If numSteps == 0 then the default for the given stops
func (st *Stack) SetStep(numSteps int, stopLevel enums.Enum) {
	st.StopLevel = stopLevel
	lp := st.Loops[stopLevel]
	if numSteps > 0 {
		st.StopCount = numSteps
		lp.StepCount = numSteps
	} else {
		numSteps = lp.StepCount
	}
	st.StopCount = numSteps
	st.StepLevel = stopLevel
	st.StepCount = numSteps
	st.StopFlag = false
	st.StopNext = true
}

// ClearStep clears the active stepping control state: StopNext and StopFlag.
func (st *Stack) ClearStep() {
	st.StopNext = false
	st.StopFlag = false
}

// Counters returns a slice of the current counter values
// for this stack, in Order.
func (st *Stack) Counters() []int {
	ctrs := make([]int, len(st.Order))
	for i, tm := range st.Order {
		ctrs[i] = st.Loops[tm].Counter.Cur
	}
	return ctrs
}

// CountersString returns a string with loop level and counter values.
func (st *Stack) CountersString() string {
	ctrs := ""
	for _, tm := range st.Order {
		ctrs += fmt.Sprintf("%s: %d ", tm.String(), st.Loops[tm].Counter.Cur)
	}
	return ctrs
}

// DocString returns an indented summary of the loops and functions in the Stack.
func (st *Stack) DocString() string {
	var sb strings.Builder
	sb.WriteString("Stack " + st.Mode.String() + ":\n")
	sb.WriteString(st.Level(0).DocString(st, 0))
	return sb.String()
}
