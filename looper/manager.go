// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

//go:generate core generate -add-types

import (
	"fmt"
	"strconv"
	"strings"

	"cogentcore.org/core/base/indent"
	"github.com/emer/emergent/v2/etime"
)

var (
	// If you want to debug the flow of time, set this to true.
	PrintControlFlow = false

	// If PrintControlFlow = true, this cuts off printing at timescales
	// that are faster than this -- default is to print all.
	NoPrintBelow = etime.AllTimes
)

// Manager holds data relating to multiple stacks of loops,
// as well as the logic for stepping through it.
// It also holds helper methods for constructing the data.
// It's also a control object for stepping through Stacks of Loops.
// It holds data about how the flow is going.
type Manager struct {

	// map of stacks by Mode
	Stacks map[etime.Modes]*Stack

	// The current evaluation mode.
	Mode etime.Modes

	// Set to true while looping, false when done. Read only.
	isRunning bool

	// The Cur value of the Counter associated with the last started level, for each timescale.
	lastStartedCounter map[etime.ScopeKey]int
	internalStop       bool
}

// GetLoop returns the Loop associated with an evaluation mode and timescale.
func (man *Manager) GetLoop(modes etime.Modes, times etime.Times) *Loop {
	st := man.Stacks[modes]
	if st == nil {
		return nil
	}
	return st.Loops[times]
}

// NewManager returns a new initialized manager
func NewManager() *Manager {
	man := &Manager{}
	man.Init()
	return man
}

// Init initializes the state of the manager, to be called on a newly created object.
func (man *Manager) Init() {
	man.Stacks = map[etime.Modes]*Stack{}
	man.Mode = etime.Train
	man.lastStartedCounter = map[etime.ScopeKey]int{}
}

// AddStack adds a new Stack for given mode
func (man *Manager) AddStack(mode etime.Modes) *Stack {
	stack := &Stack{}
	stack.Init(mode)
	man.Stacks[mode] = stack
	return stack
}

// ModeStack returns the Stack for the current Mode
func (man *Manager) ModeStack() *Stack {
	return man.Stacks[man.Mode]
}

// AddEventAllModes adds Event(s) to all stacks at given time
func (man *Manager) AddEventAllModes(t etime.Times, event ...*Event) {
	for _, stack := range man.Stacks {
		stack.Loops[t].AddEvents(event...)
	}
}

// AddOnStartToAll adds given function taking mode and time args to OnStart in all stacks, loops
func (man *Manager) AddOnStartToAll(name string, fun func(mode etime.Modes, time etime.Times)) {
	for _, stack := range man.Stacks {
		stack.AddOnStartToAll(name, fun)
	}
}

// AddMainToAll adds given function taking mode and time args to Main in all stacks, loops
func (man *Manager) AddMainToAll(name string, fun func(mode etime.Modes, time etime.Times)) {
	for _, stack := range man.Stacks {
		stack.AddMainToAll(name, fun)
	}
}

// AddOnEndToAll adds given function taking mode and time args to OnEnd in all stacks, loops
func (man *Manager) AddOnEndToAll(name string, fun func(mode etime.Modes, time etime.Times)) {
	for _, stack := range man.Stacks {
		stack.AddOnEndToAll(name, fun)
	}
}

// DocString returns an indented summary of the loops and functions in the stack.
func (man *Manager) DocString() string {
	var sb strings.Builder

	// indentSize is number of spaces to indent for output
	var indentSize = 4

	for evalMode, st := range man.Stacks {
		sb.WriteString("Stack: " + evalMode.String() + "\n")
		for i, t := range st.Order {
			lp := st.Loops[t]
			sb.WriteString(indent.Spaces(i, indentSize) + evalMode.String() + ":" + t.String() + ":\n")
			sb.WriteString(indent.Spaces(i+1, indentSize) + "  Start:  " + lp.OnStart.String() + "\n")
			sb.WriteString(indent.Spaces(i+1, indentSize) + "  Main:  " + lp.Main.String() + "\n")
			if len(lp.IsDone) > 0 {
				s := ""
				for nm, _ := range lp.IsDone {
					s = s + nm + " "
				}
				sb.WriteString(indent.Spaces(i+1, indentSize) + "  Stop:  " + s + "\n")
			}
			sb.WriteString(indent.Spaces(i+1, indentSize) + "  End:   " + lp.OnEnd.String() + "\n")
			if len(lp.Events) > 0 {
				sb.WriteString(indent.Spaces(i+1, indentSize) + "  Events:\n")
				for _, ph := range lp.Events {
					sb.WriteString(indent.Spaces(i+2, indentSize) + ph.String() + "\n")
				}
			}
		}
	}
	return sb.String()
}

// All the rest is related to stepping

// IsRunning is True if running.
func (man *Manager) IsRunning() bool {
	return man.isRunning
}

// ResetCountersByMode resets counters for given mode.
func (man *Manager) ResetCountersByMode(mode etime.Modes) {
	for sk, _ := range man.lastStartedCounter {
		skm, _ := sk.ModeAndTime()
		if skm == mode {
			delete(man.lastStartedCounter, sk)
		}
	}
	for m, stack := range man.Stacks {
		if m == mode {
			for _, loop := range stack.Loops {
				loop.Counter.Cur = 0
			}
		}
	}
}

// ResetCounters resets the Cur on all loop Counters,
// and resets the Manager's place in the loops.
func (man *Manager) ResetCounters() {
	man.lastStartedCounter = map[etime.ScopeKey]int{}
	for _, stack := range man.Stacks {
		for _, loop := range stack.Loops {
			loop.Counter.Cur = 0
		}
	}
}

// ResetCountersBelow resets the Cur on all loop Counters below given level
// (inclusive), and resets the Manager's place in the loops.
func (man *Manager) ResetCountersBelow(mode etime.Modes, time etime.Times) {
	for _, stack := range man.Stacks {
		if stack.Mode != mode {
			continue
		}
		for lt, loop := range stack.Loops {
			if lt > time {
				continue
			}
			loop.Counter.Cur = 0
			sk := etime.Scope(mode, lt)
			delete(man.lastStartedCounter, sk)
		}
	}
}

// Step numSteps stopscales. Use this if you want to do exactly one trial
// or two epochs or 50 cycles or whatever
func (man *Manager) Step(mode etime.Modes, numSteps int, stopscale etime.Times) {
	man.Mode = mode
	st := man.Stacks[man.Mode]
	st.SetStep(numSteps, stopscale)
	man.Cont()
}

// ClearStep clears stepping variables from given mode,
// so it will run to completion in a subsequent Cont().
// Called by Run
func (man *Manager) ClearStep(mode etime.Modes) {
	st := man.Stacks[man.Mode]
	st.ClearStep()
}

// Run runs the stack of loops for given mode (Train, Test, etc).
// This resets any stepping settings for this stack and runs
// until completion or stopped externally.
func (man *Manager) Run(mode etime.Modes) {
	man.Mode = mode
	man.ClearStep(mode)
	man.Cont()
}

// ResetAndRun calls ResetCountersByMode on this mode
// and then Run.  This ensures that the Stack is run from
// the start, regardless of what state it might have been in.
func (man *Manager) ResetAndRun(mode etime.Modes) {
	man.ResetCountersByMode(mode)
	man.Run(mode)
}

// Cont continues running based on current state of the manager.
// This is common pathway for Step and Run, which set state and
// call Cont.  Programatic calling of Step can continue with Cont.
func (man *Manager) Cont() {
	man.isRunning = true
	man.internalStop = false
	man.runLevel(0) // 0 Means the top level loop
	man.isRunning = false
}

// Stop stops currently running stack of loops at given run time level
func (man *Manager) Stop(level etime.Times) {
	st := man.Stacks[man.Mode]
	st.StopLevel = level
	st.StopIterations = 0
	st.StopFlag = true
}

// runLevel implements nested for loops recursively.
// It is set up so that it can be stopped and resumed at any point.
func (man *Manager) runLevel(currentLevel int) bool {
	st := man.Stacks[man.Mode]
	if currentLevel >= len(st.Order) {
		return true // Stack overflow, expected at bottom of stack.
	}
	time := st.Order[currentLevel]
	loop := st.Loops[time]
	ctr := &loop.Counter

	for ctr.Cur < ctr.Max || ctr.Max < 0 { // Loop forever for negative maxes
		stopAtLevelOrLarger := st.Order[currentLevel] >= st.StopLevel // Based on conversion of etime.Times to int
		if st.StopFlag && stopAtLevelOrLarger {
			man.internalStop = true
		}
		if man.internalStop {
			// This should occur before ctr incrementing and before functions.
			st.StopFlag = false
			return false // Don't continue above, e.g. Stop functions
		}
		if st.StopNext && st.Order[currentLevel] == st.StopLevel {
			st.StopIterations -= 1
			if st.StopIterations <= 0 {
				st.StopNext = false
				st.StopFlag = true // Stop at the top of the next StopLevel
			}
		}

		// Don't ever Start the same iteration of the same level twice.
		lastCounter, ok := man.lastStartedCounter[etime.Scope(man.Mode, time)]
		if !ok || ctr.Cur > lastCounter {
			man.lastStartedCounter[etime.Scope(man.Mode, time)] = ctr.Cur
			if PrintControlFlow && time >= NoPrintBelow {
				fmt.Println(time.String() + ":Start:" + strconv.Itoa(ctr.Cur))
			}
			// Events occur at the very start.
			man.eventLogic(loop)
			for _, fun := range loop.OnStart {
				fun.Func()
			}
		} else if PrintControlFlow && time >= NoPrintBelow {
			fmt.Println("Skipping start: " + time.String() + ":" + strconv.Itoa(ctr.Cur))
		}

		// Recursion!
		runComplete := man.runLevel(currentLevel + 1)

		if runComplete {
			for _, fun := range loop.Main {
				fun.Func()
			}
			if PrintControlFlow && time >= NoPrintBelow {
				fmt.Println(time.String() + ":End:  " + strconv.Itoa(ctr.Cur))
			}
			for _, fun := range loop.OnEnd {
				fun.Func()
			}

			// Increment
			ctr.Incr()
			// Reset the counter at the next level. Do this here so that the counter number is visible during loop.OnEnd.
			if currentLevel+1 < len(st.Order) {
				st.Loops[st.Order[currentLevel+1]].Counter.Cur = 0
				man.lastStartedCounter[etime.Scope(man.Mode, st.Order[currentLevel+1])] = -1
			}

			for name, fun := range loop.IsDone {
				if fun() {
					if PrintControlFlow {
						fmt.Println("Stopping early with: " + name + " condition")
					}
					goto exitLoop // Exit IsDone and Counter for-loops without flag variable.
				}
			}
		}
	}

exitLoop:
	// Only get to this point if this loop is done.
	return true
}

// eventLogic handles events that occur at specific timesteps.
func (man *Manager) eventLogic(loop *Loop) {
	ctr := &loop.Counter
	for _, phase := range loop.Events {
		if ctr.Cur == phase.AtCounter {
			for _, function := range phase.OnEvent {
				function.Func()
			}
		}
	}
}
