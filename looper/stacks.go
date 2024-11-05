// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

//go:generate core generate -add-types

import (
	"strings"

	"cogentcore.org/core/enums"
)

var (
	// If you want to debug the flow of time, set this to true.
	PrintControlFlow = false
)

// Stacks holds data relating to multiple stacks of loops,
// as well as the logic for stepping through it.
// It also holds helper methods for constructing the data.
// It's also a control object for stepping through Stacks of Loops.
// It holds data about how the flow is going.
type Stacks struct {

	// Stacks is the map of stacks by Mode.
	Stacks map[enums.Enum]*Stack

	// Mode has the current evaluation mode.
	Mode enums.Enum

	// following are internal run control state: see runLevel in run.go.
	isRunning          bool
	lastStartedCounter map[Scope]int
	internalStop       bool
}

// NewStacks returns a new initialized collection of Stacks.
func NewStacks() *Stacks {
	ss := &Stacks{}
	ss.Init()
	return ss
}

// Init initializes the state of the stacks, to be called on a newly created object.
func (ss *Stacks) Init() {
	ss.Stacks = map[enums.Enum]*Stack{}
	ss.lastStartedCounter = map[Scope]int{}
}

//////// Run API

// Run runs the stack of loops for given mode (Train, Test, etc).
// This resets any stepping settings for this stack and runs
// until completion or stopped externally.
func (ss *Stacks) Run(mode enums.Enum) {
	ss.Mode = mode
	ss.ClearStep(mode)
	ss.Cont()
}

// ResetAndRun calls ResetCountersByMode on this mode
// and then Run.  This ensures that the Stack is run from
// the start, regardless of what state it might have been in.
func (ss *Stacks) ResetAndRun(mode enums.Enum) {
	ss.ResetCountersByMode(mode)
	ss.Run(mode)
}

// Cont continues running based on current state of the stacks.
// This is common pathway for Step and Run, which set state and
// call Cont. Programatic calling of Step can continue with Cont.
func (ss *Stacks) Cont() {
	ss.isRunning = true
	ss.internalStop = false
	ss.runLevel(0) // 0 Means the top level loop
	ss.isRunning = false
}

// Step numSteps stopscales. Use this if you want to do exactly one trial
// or two epochs or 50 cycles or whatever
func (ss *Stacks) Step(mode enums.Enum, numSteps int, stopscale enums.Enum) {
	ss.Mode = mode
	st := ss.Stacks[ss.Mode]
	st.SetStep(numSteps, stopscale)
	ss.Cont()
}

// ClearStep clears stepping variables from given mode,
// so it will run to completion in a subsequent Cont().
// Called by Run
func (ss *Stacks) ClearStep(mode enums.Enum) {
	st := ss.Stacks[ss.Mode]
	st.ClearStep()
}

// Stop stops currently running stack of loops at given run time level
func (ss *Stacks) Stop(level enums.Enum) {
	st := ss.Stacks[ss.Mode]
	st.StopLevel = level
	st.StopCount = 0
	st.StopFlag = true
}

//////// Config API

// AddStack adds a new Stack for given mode
func (ss *Stacks) AddStack(mode enums.Enum) *Stack {
	stack := NewStack(mode)
	ss.Stacks[mode] = stack
	return stack
}

// Loop returns the Loop associated with given mode and timescale.
func (ss *Stacks) Loop(mode, time enums.Enum) *Loop {
	st := ss.Stacks[mode]
	if st == nil {
		return nil
	}
	return st.Loops[time]
}

// ModeStack returns the Stack for the current Mode
func (ss *Stacks) ModeStack() *Stack {
	return ss.Stacks[ss.Mode]
}

// AddEventAllModes adds a new event for all modes at given timescale.
func (ss *Stacks) AddEventAllModes(time enums.Enum, name string, atCtr int, fun func()) {
	for _, stack := range ss.Stacks {
		stack.Loops[time].AddEvent(name, atCtr, fun)
	}
}

//////// More detailed control API

// IsRunning is True if running.
func (ss *Stacks) IsRunning() bool {
	return ss.isRunning
}

// ResetCountersByMode resets counters for given mode.
func (ss *Stacks) ResetCountersByMode(mode enums.Enum) {
	for sk, _ := range ss.lastStartedCounter {
		skm, _ := sk.ModeTime()
		if skm == mode.Int64() {
			delete(ss.lastStartedCounter, sk)
		}
	}
	for m, stack := range ss.Stacks {
		if m == mode {
			for _, loop := range stack.Loops {
				loop.Counter.Cur = 0
			}
		}
	}
}

// ResetCounters resets the Cur on all loop Counters,
// and resets the Stacks's place in the loops.
func (ss *Stacks) ResetCounters() {
	ss.lastStartedCounter = map[Scope]int{}
	for _, stack := range ss.Stacks {
		for _, loop := range stack.Loops {
			loop.Counter.Cur = 0
		}
	}
}

// ResetCountersBelow resets the Cur on all loop Counters below given level
// (inclusive), and resets the Stacks's place in the loops.
func (ss *Stacks) ResetCountersBelow(mode enums.Enum, time enums.Enum) {
	for _, stack := range ss.Stacks {
		if stack.Mode != mode {
			continue
		}
		for lt, loop := range stack.Loops {
			if lt.Int64() > time.Int64() {
				continue
			}
			loop.Counter.Cur = 0
			sk := ToScope(mode, lt)
			delete(ss.lastStartedCounter, sk)
		}
	}
}

// DocString returns an indented summary of the loops and functions in the stack.
func (ss *Stacks) DocString() string {
	var sb strings.Builder
	for _, st := range ss.Stacks {
		sb.WriteString(st.DocString())
	}
	return sb.String()
}
