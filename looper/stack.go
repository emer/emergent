// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import (
	"strings"

	"cogentcore.org/core/enums"
	"github.com/emer/emergent/v2/estats"
	"github.com/emer/emergent/v2/etime"
)

// Stack contains a list of Loops to run, for a given Mode of processing.
// The order of Loop stacks is determined by the Order list.
type Stack struct {

	// Mode identifies the mode of processing this stack performs, e.g., Train or Test.
	Mode enums.Enum

	// Loops is the set of Loops for this Stack, keyed by the timescale.
	// Order is determined by the Order list.
	Loops map[enums.Enum]*Loop

	// Order is the list and order of time scales looped over by this stack of loops.
	// The ordered is from top to bottom, so longer timescales like Run should be at
	// the beginning and shorter timescales like Trial should be and the end.
	Order []enums.Enum

	// If true, stop model at the end of the current StopLevel.
	StopNext bool

	// StopFlag will stop running ASAP if set.
	StopFlag bool

	// StopLevel sets the Time level to stop at the end of.
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

// NewStack returns a new Stack for given mode.
func NewStack(mode enums.Enum) *Stack {
	stack := &Stack{}
	stack.Init(mode)
	return stack
}

// Init initializes new data structures for a newly created object
func (st *Stack) Init(mode enums.Enum) {
	st.Mode = mode
	st.StepLevel = etime.Trial
	st.StepCount = 1
	st.Loops = map[enums.Enum]*Loop{}
	st.Order = []enums.Enum{}
}

// Level returns the [Loop] at the given ordinal level in the Order list.
// Will panic if out of range.
func (st *Stack) Level(i int) *Loop {
	return st.Loops[st.Order[i]]
}

// AddTime adds a new timescale to this Stack with a given ctrMax number of iterations.
// The order in which this method is invoked is important,
// as it adds loops in order from top to bottom.
// Sets a default increment of 1 for the counter -- see AddTimeIncr for different increment.
func (st *Stack) AddTime(time enums.Enum, ctrMax int) *Stack {
	st.Loops[time] = NewLoop(ctrMax, 1)
	st.Order = append(st.Order, time)
	return st
}

// AddTimeIncr adds a new timescale to this Stack with a given ctrMax number of iterations,
// and increment per step.
// The order in which this method is invoked is important,
// as it adds loops in order from top to bottom.
func (st *Stack) AddTimeIncr(time enums.Enum, ctrMax, ctrIncr int) *Stack {
	st.Loops[time] = NewLoop(ctrMax, ctrIncr)
	st.Order = append(st.Order, time)
	return st
}

// TimeAbove returns the time above the given time in the stack
// returning etime.NoTime if this is the highest level,
// or given time does not exist in order.
func (st *Stack) TimeAbove(time enums.Enum) enums.Enum {
	for i, tt := range st.Order {
		if tt == time && i > 0 {
			return st.Order[i-1]
		}
	}
	return etime.NoTime
}

// TimeBelow returns the time below the given time in the stack
// returning etime.NoTime if this is the lowest level,
// or given time does not exist in order.
func (st *Stack) TimeBelow(time enums.Enum) enums.Enum {
	for i, tt := range st.Order {
		if tt == time && i+1 < len(st.Order) {
			return st.Order[i+1]
		}
	}
	return etime.NoTime
}

// SetStep sets stepping to given level and number of iterations.
func (st *Stack) SetStep(numSteps int, stopscale enums.Enum) {
	st.StopLevel = stopscale
	st.StopCount = numSteps
	st.StepLevel = stopscale
	st.StepCount = numSteps
	st.StopFlag = false
	st.StopNext = true
}

// ClearStep clears the active stepping control state: StopNext and StopFlag.
func (st *Stack) ClearStep() {
	st.StopNext = false
	st.StopFlag = false
}

// CountersToStats sets the current counter values to estats Int values
// by their time names only (no eval Mode). These values can then
// be read by elog LogItems to record the counters in logs.
// Typically, a TrialName string is also expected to be set,
// to describe the current trial (Step) contents in a useful way,
// and other relevant info (e.g., group / category info) can also be set.
func (st *Stack) CountersToStats(stats *estats.Stats) {
	for _, tm := range st.Order {
		lp := st.Loops[tm]
		stats.SetInt(tm.String(), lp.Counter.Cur)
	}
}

// DocString returns an indented summary of the loops and functions in the Stack.
func (st *Stack) DocString() string {
	var sb strings.Builder
	sb.WriteString("Stack " + st.Mode.String() + ":\n")
	sb.WriteString(st.Level(0).DocString(st, 0))
	return sb.String()
}
