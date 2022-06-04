// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import (
	"github.com/emer/emergent/estats"
	"github.com/emer/emergent/etime"
)

// Stack contains a list of Loops Ordered from top to bottom.
// For example, a Stack might be created like this:
//   mystack := manager.AddStack(etime.Train).AddTime(etime.Run, 2).AddTime(etime.Trial, 3)
//   myStack.Loops[etime.Run].OnStart.Add("NewRun", initRunFunc)
//   myStack.Loops[etime.Trial].OnStart.Add("PresentTrial", trialFunc)
// When run, myStack will behave like this:
// initRunFunc, trialFunc, trialFunc, trialFunc, initRunFunc, trialFunc, trialFunc, trialFunc
type Stack struct {
	Mode  etime.Modes           `desc:"evaluation mode for this stack"`
	Loops map[etime.Times]*Loop `desc:"An ordered map of Loops, from the outer loop at the start to the inner loop at the end."`
	Order []etime.Times         `desc:"The list and order of time scales looped over by this stack of loops,  ordered from top to bottom, so longer timescales like Run should be at the beginning and shorter timescales like Trial should be and the end."`
}

// Init makes sure data structures are initialized, and empties them if they are.
func (stack *Stack) Init(mode etime.Modes) {
	stack.Mode = mode
	stack.Loops = map[etime.Times]*Loop{}
	stack.Order = []etime.Times{}
}

// AddTime adds a new timescale to this Stack with a given number of iterations. The order in which this method is invoked is important, as it adds loops in order from top to bottom.
func (stack *Stack) AddTime(time etime.Times, max int) *Stack {
	stack.Loops[time] = &Loop{Counter: Ctr{Max: max}, IsDone: map[string]func() bool{}}
	stack.Order = append(stack.Order, time)
	return stack
}

// AddOnStartToAll adds given function taking mode and time args to OnStart in all loops
func (stack *Stack) AddOnStartToAll(name string, fun func(mode etime.Modes, time etime.Times)) {
	for tt, lp := range stack.Loops {
		curTime := tt
		lp.OnStart.Add(stack.Mode.String()+":"+curTime.String()+":"+name, func() {
			fun(stack.Mode, curTime)
		})
	}
}

// AddMainToAll adds given function taking mode and time args to Main in all loops
func (stack *Stack) AddMainToAll(name string, fun func(mode etime.Modes, time etime.Times)) {
	for tt, lp := range stack.Loops {
		curTime := tt
		lp.Main.Add(stack.Mode.String()+":"+curTime.String()+":"+name, func() {
			fun(stack.Mode, curTime)
		})
	}
}

// AddOnEndToAll adds given function taking mode and time args to OnEnd in all loops
func (stack *Stack) AddOnEndToAll(name string, fun func(mode etime.Modes, time etime.Times)) {
	for tt, lp := range stack.Loops {
		curTime := tt
		lp.OnEnd.Add(stack.Mode.String()+":"+curTime.String()+":"+name, func() {
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
