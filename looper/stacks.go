// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

//go:generate core generate -add-types

import (
	"cmp"
	"slices"
	"strings"

	"cogentcore.org/core/enums"
	"golang.org/x/exp/maps"
)

var (
	// If you want to debug the flow of processing, set this to true.
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
	ls := &Stacks{}
	ls.newInit()
	return ls
}

// newInit initializes the state of the stacks, to be called on a newly created object.
func (ls *Stacks) newInit() {
	ls.Stacks = map[enums.Enum]*Stack{}
	ls.lastStartedCounter = map[Scope]int{}
}

//////// Run API

// Run runs the stack of loops for given mode (Train, Test, etc).
// This resets any stepping settings for this stack and runs
// until completion or stopped externally.
// Returns the level that was running when it stopped.
func (ls *Stacks) Run(mode enums.Enum) enums.Enum {
	ls.Mode = mode
	ls.ClearStep(mode)
	return ls.Cont()
}

// ResetAndRun calls ResetCountersByMode on this mode
// and then Run.  This ensures that the Stack is run from
// the start, regardless of what state it might have been in.
// Returns the level that was running when it stopped.
func (ls *Stacks) ResetAndRun(mode enums.Enum) enums.Enum {
	ls.ResetCountersByMode(mode)
	return ls.Run(mode)
}

// Cont continues running based on current state of the stacks.
// This is common pathway for Step and Run, which set state and
// call Cont. Programatic calling of Step can continue with Cont.
// Returns the level that was running when it stopped.
func (ls *Stacks) Cont() enums.Enum {
	ls.isRunning = true
	ls.internalStop = false
	_, stop := ls.runLevel(0) // 0 Means the top level loop
	ls.isRunning = false
	return stop
}

// Step numSteps at given stopLevel. Use this if you want to do exactly one trial
// or two epochs or 50 cycles or whatever. If numSteps <= 0 then the default
// number of steps for given step level is used.
// Returns the level that was running when it stopped.
func (ls *Stacks) Step(mode enums.Enum, numSteps int, stopLevel enums.Enum) enums.Enum {
	ls.Mode = mode
	st := ls.Stacks[ls.Mode]
	st.SetStep(numSteps, stopLevel)
	return ls.Cont()
}

// ClearStep clears stepping variables from given mode,
// so it will run to completion in a subsequent Cont().
// Called by Run.
func (ls *Stacks) ClearStep(mode enums.Enum) {
	st := ls.Stacks[ls.Mode]
	st.ClearStep()
}

// Stop stops currently running stack of loops at given run level.
func (ls *Stacks) Stop(level enums.Enum) {
	st := ls.Stacks[ls.Mode]
	st.StopLevel = level
	st.StopCount = 0
	st.StopFlag = true
}

//////// Config API

// AddStack adds a new Stack for given mode and default step level.
func (ls *Stacks) AddStack(mode, stepLevel enums.Enum) *Stack {
	st := NewStack(mode, stepLevel)
	ls.Stacks[mode] = st
	return st
}

// Loop returns the Loop associated with given mode and loop level.
func (ls *Stacks) Loop(mode, level enums.Enum) *Loop {
	st := ls.Stacks[mode]
	if st == nil {
		return nil
	}
	return st.Loops[level]
}

// ModeStack returns the Stack for the current Mode
func (ls *Stacks) ModeStack() *Stack {
	return ls.Stacks[ls.Mode]
}

// AddEventAllModes adds a new event for all modes at given loop level.
func (ls *Stacks) AddEventAllModes(level enums.Enum, name string, atCtr int, fun func()) {
	for _, st := range ls.Stacks {
		st.Loops[level].AddEvent(name, atCtr, fun)
	}
}

// AddOnStartToAll adds given function taking mode and level args to OnStart in all stacks, loops
func (ls *Stacks) AddOnStartToAll(name string, fun func(mode, level enums.Enum)) {
	for _, st := range ls.Stacks {
		st.AddOnStartToAll(name, fun)
	}
}

// AddOnEndToAll adds given function taking mode and level args to OnEnd in all stacks, loops
func (ls *Stacks) AddOnEndToAll(name string, fun func(mode, level enums.Enum)) {
	for _, st := range ls.Stacks {
		st.AddOnEndToAll(name, fun)
	}
}

// AddOnStartToLoop adds given function taking mode arg to OnStart in all stacks for given loop.
func (ls *Stacks) AddOnStartToLoop(level enums.Enum, name string, fun func(mode enums.Enum)) {
	for m, st := range ls.Stacks {
		st.Loops[level].OnStart.Add(name, func() { fun(m) })
	}
}

// AddOnEndToLoop adds given function taking mode arg to OnEnd in all stacks for given loop.
func (ls *Stacks) AddOnEndToLoop(level enums.Enum, name string, fun func(mode enums.Enum)) {
	for m, st := range ls.Stacks {
		st.Loops[level].OnEnd.Add(name, func() { fun(m) })
	}
}

// Modes returns a sorted list of stack modes, for iterating in Mode enum value order.
func (ls *Stacks) Modes() []enums.Enum {
	mds := maps.Keys(ls.Stacks)
	slices.SortFunc(mds, func(a, b enums.Enum) int {
		return cmp.Compare(a.Int64(), b.Int64())
	})
	return mds
}

//////// More detailed control API

// IsRunning is True if running.
func (ls *Stacks) IsRunning() bool {
	return ls.isRunning
}

// InitMode initializes [Stack] of given mode,
// resetting counters and calling the OnInit functions.
func (ls *Stacks) InitMode(mode enums.Enum) {
	ls.ResetCountersByMode(mode)
	st := ls.Stacks[mode]
	st.OnInit.Run()
}

// ResetCountersByMode resets counters for given mode.
func (ls *Stacks) ResetCountersByMode(mode enums.Enum) {
	for sk, _ := range ls.lastStartedCounter {
		skm, _ := sk.ModeLevel()
		if skm == mode.Int64() {
			delete(ls.lastStartedCounter, sk)
		}
	}
	for m, st := range ls.Stacks {
		if m == mode {
			for _, loop := range st.Loops {
				loop.Counter.Cur = 0
			}
		}
	}
}

// Init initializes all stacks. See [Stacks.InitMode] for more info.
func (ls *Stacks) Init() {
	ls.lastStartedCounter = map[Scope]int{}
	for _, st := range ls.Stacks {
		ls.InitMode(st.Mode)
	}
}

// ResetCounters resets the Cur on all loop Counters,
// and resets the Stacks's place in the loops.
func (ls *Stacks) ResetCounters() {
	ls.lastStartedCounter = map[Scope]int{}
	for _, st := range ls.Stacks {
		for _, loop := range st.Loops {
			loop.Counter.Cur = 0
		}
	}
}

// ResetCountersBelow resets the Cur on all loop Counters below given level
// (inclusive), and resets the Stacks's place in the loops.
func (ls *Stacks) ResetCountersBelow(mode enums.Enum, level enums.Enum) {
	for _, st := range ls.Stacks {
		if st.Mode != mode {
			continue
		}
		for lt, loop := range st.Loops {
			if lt.Int64() > level.Int64() {
				continue
			}
			loop.Counter.Cur = 0
			sk := ToScope(mode, lt)
			delete(ls.lastStartedCounter, sk)
		}
	}
}

// DocString returns an indented summary of the loops and functions in the stack.
func (ls *Stacks) DocString() string {
	var sb strings.Builder
	for _, st := range ls.Stacks {
		sb.WriteString(st.DocString())
	}
	return sb.String()
}
