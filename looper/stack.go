// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import (
	"github.com/emer/emergent/v2/estats"
	"github.com/emer/emergent/v2/etime"
)

// Stack contains a list of Loops Ordered from top to bottom.
// For example, a Stack might be created like this:
//
//	mystack := manager.AddStack(etime.Train).AddTime(etime.Run, 2).AddTime(etime.Trial, 3)
//	myStack.Loops[etime.Run].OnStart.Add("NewRun", initRunFunc)
//	myStack.Loops[etime.Trial].OnStart.Add("PresentTrial", trialFunc)
//
// When run, myStack will behave like this:
// initRunFunc, trialFunc, trialFunc, trialFunc, initRunFunc, trialFunc, trialFunc, trialFunc
type Stack struct {

	// evaluation mode for this stack
	Mode etime.Modes

	// An ordered map of Loops, from the outer loop at the start to the inner loop at the end.
	Loops map[etime.Times]*Loop

	// The list and order of time scales looped over by this stack of loops,  ordered from top to bottom, so longer timescales like Run should be at the beginning and shorter timescales like Trial should be and the end.
	Order []etime.Times

	// If true, stop model at the end of the current StopLevel.
	StopNext bool

	// If true, stop model ASAP.
	StopFlag bool

	// Time level to stop at the end of.
	StopLevel etime.Times

	// How many iterations at StopLevel before actually stopping.
	StopIterations int

	// Saved Time level for stepping -- what was set for last step or by gui.
	StepLevel etime.Times

	// Saved number of steps for stepping -- what was set for last step or by gui.
	StepIterations int
}

// Init initializes new data structures for a newly created object
func (stack *Stack) Init(mode etime.Modes) {
	stack.Mode = mode
	stack.StepLevel = etime.Trial
	stack.StepIterations = 1
	stack.Loops = map[etime.Times]*Loop{}
	stack.Order = []etime.Times{}
}

// AddTime adds a new timescale to this Stack with a given ctrMax number of iterations.
// The order in which this method is invoked is important,
// as it adds loops in order from top to bottom.
// Sets a default increment of 1 for the counter -- see AddTimeIncr for different increment.
func (stack *Stack) AddTime(time etime.Times, ctrMax int) *Stack {
	stack.Loops[time] = &Loop{Counter: Ctr{Max: ctrMax, Inc: 1}, IsDone: map[string]func() bool{}}
	stack.Order = append(stack.Order, time)
	return stack
}

// AddTimeIncr adds a new timescale to this Stack with a given ctrMax number of iterations,
// and increment per step.
// The order in which this method is invoked is important,
// as it adds loops in order from top to bottom.
func (stack *Stack) AddTimeIncr(time etime.Times, ctrMax, ctrIncr int) *Stack {
	stack.Loops[time] = &Loop{Counter: Ctr{Max: ctrMax, Inc: ctrIncr}, IsDone: map[string]func() bool{}}
	stack.Order = append(stack.Order, time)
	return stack
}

// AddOnStartToAll adds given function taking mode and time args to OnStart in all loops
func (stack *Stack) AddOnStartToAll(name string, fun func(mode etime.Modes, time etime.Times)) {
	for tt, lp := range stack.Loops {
		curTime := tt
		lp.OnStart.Add(name, func() {
			fun(stack.Mode, curTime)
		})
	}
}

// AddMainToAll adds given function taking mode and time args to Main in all loops
func (stack *Stack) AddMainToAll(name string, fun func(mode etime.Modes, time etime.Times)) {
	for tt, lp := range stack.Loops {
		curTime := tt
		lp.Main.Add(name, func() {
			fun(stack.Mode, curTime)
		})
	}
}

// AddOnEndToAll adds given function taking mode and time args to OnEnd in all loops
func (stack *Stack) AddOnEndToAll(name string, fun func(mode etime.Modes, time etime.Times)) {
	for tt, lp := range stack.Loops {
		curTime := tt
		lp.OnEnd.Add(name, func() {
			fun(stack.Mode, curTime)
		})
	}
}

// TimeAbove returns the time above the given time in the stack
// returning etime.NoTime if this is the highest level,
// or given time does not exist in order.
func (stack *Stack) TimeAbove(time etime.Times) etime.Times {
	for i, tt := range stack.Order {
		if tt == time && i > 0 {
			return stack.Order[i-1]
		}
	}
	return etime.NoTime
}

// TimeBelow returns the time below the given time in the stack
// returning etime.NoTime if this is the lowest level,
// or given time does not exist in order.
func (stack *Stack) TimeBelow(time etime.Times) etime.Times {
	for i, tt := range stack.Order {
		if tt == time && i+1 < len(stack.Order) {
			return stack.Order[i+1]
		}
	}
	return etime.NoTime
}

// CtrsToStats sets the current counter values to estats Int values
// by their time names only (no eval Mode).  These values can then
// be read by elog LogItems to record the counters in logs.
// Typically, a TrialName string is also expected to be set,
// to describe the current trial (Step) contents in a useful way,
// and other relevant info (e.g., group / category info) can also be set.
func (stack *Stack) CtrsToStats(stats *estats.Stats) {
	for _, tm := range stack.Order {
		lp := stack.Loops[tm]
		stats.SetInt(tm.String(), lp.Counter.Cur)
	}
}

// SetStep sets stepping to given level and iterations
func (stack *Stack) SetStep(numSteps int, stopscale etime.Times) {
	stack.StopLevel = stopscale
	stack.StopIterations = numSteps
	stack.StepLevel = stopscale
	stack.StepIterations = numSteps
	stack.StopFlag = false
	stack.StopNext = true
}

// ClearStep clears stepping control state
func (stack *Stack) ClearStep() {
	stack.StopNext = false
	stack.StopFlag = false
}
