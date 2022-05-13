package looper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/emer/emergent/etime"
	"github.com/goki/ki/indent"
)

var (
	// If you want to debug the flow of time, set this to true.
	PrintControlFlow = false

	// If PrintControlFlow = true, this cuts off printing at timescales that are faster than this -- default is to print all.
	NoPrintBelow = etime.AllTimes
)

// Manager holds data relating to multiple stacks of loops, as well as the logic for stepping through it. It also holds helper methods for constructing the data.
// It's also a control object for stepping through Stacks of Loops. It holds data about how the flow is going.
type Manager struct {
	Stacks map[etime.Modes]*Stack

	StopFlag       bool        `desc:"If true, stop model ASAP."`
	StopNext       bool        `desc:"If true, stop model at the end of the current StopLevel."`
	StopLevel      etime.Times `desc:"Time level to stop at the end of."`
	StepIterations int         `desc:"How many steps to do."`
	Mode           etime.Modes `desc:"The current evaluation mode."`
	isRunning      bool        `desc:"Set to true while looping, false when done. Read only."`

	// For internal use
	lastStartedCtr map[etime.ScopeKey]int `desc:"The Cur value of the Ctr associated with the last started level, for each timescale."`
	internalStop   bool
}

// GetLoop returns the Loop associated with an evaluation mode and timescale.
func (loopman *Manager) GetLoop(modes etime.Modes, times etime.Times) *Loop {
	return loopman.Stacks[modes].Loops[times]
}

// Init initializes variables on the Manager.
func (loopman Manager) Init() *Manager {
	loopman.Stacks = map[etime.Modes]*Stack{}
	loopman.StopLevel = etime.Run
	loopman.Mode = etime.Train
	loopman.lastStartedCtr = map[etime.ScopeKey]int{}
	loopman.ResetCounters()
	return &loopman
}

// AddStack adds a new Stack for given mode
func (loopman Manager) AddStack(mode etime.Modes) *Stack {
	stack := &Stack{}
	loopman.Stacks[etime.Train] = stack
	stack.Init()
	return stack
}

// ApplyAcrossAllModesAndTimes applies a function across all evaluation modes and timescales within the Manager. The function might call GetLoop(curMode, curTime) and modify it.
func (loopman *Manager) ApplyAcrossAllModesAndTimes(fun func(etime.Modes, etime.Times)) {
	for _, m := range []etime.Modes{etime.Train, etime.Test} {
		curMode := m // For closures.
		for _, t := range []etime.Times{etime.Trial, etime.Epoch} {
			curTime := t
			fun(curMode, curTime)
		}
	}
}

// AddEventAllModes adds a Event to the stack for all modes.
func (loopman *Manager) AddEventAllModes(t etime.Times, event Event) {
	// Note that phase is copied
	for mode, _ := range loopman.Stacks {
		stack := loopman.Stacks[mode]
		stack.Loops[t].AddEvents(event)
	}
}

// DocString returns an indented summary of the loops and functions in the stack.
func (loopman Manager) DocString() string {
	var sb strings.Builder

	// indentSize is number of spaces to indent for output
	var indentSize = 4

	for evalMode, st := range loopman.Stacks {
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
				sb.WriteString(indent.Spaces(i+1, indentSize) + "  Phases:\n")
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
func (stepper Manager) IsRunning() bool {
	return stepper.isRunning
}

// ResetCountersByMode is like ResetCounters, but only for one mode.
func (stepper *Manager) ResetCountersByMode(modes etime.Modes) {
	for sk, _ := range stepper.lastStartedCtr {
		skm, _ := sk.ModeAndTime()
		if skm == modes {
			stepper.lastStartedCtr[sk] = 0
		}
	}
	for m, stack := range stepper.Stacks {
		if m == modes {
			for _, loop := range stack.Loops {
				loop.Counter.Cur = 0
			}
		}
	}
}

// ResetCounters resets the Cur on all loop Counters, and resets the Stepper's place in the loops.
func (stepper *Manager) ResetCounters() {
	for m, _ := range stepper.Stacks {
		stepper.ResetCountersByMode(m)
	}
}

// Step numSteps stopscales. Use this if you want to do exactly one trial or two epochs or 50 cycles or something.
func (stepper *Manager) Step(numSteps int, stopscale etime.Times) {
	stepper.StopLevel = stopscale
	stepper.StepIterations = numSteps
	stepper.StopFlag = false
	stepper.StopNext = true
	stepper.Run()
}

// Run runs the loops contained within the stepper. If you want it to stop before the full end of the loop, set variables on the Stepper.
func (stepper *Manager) Run() {
	stepper.isRunning = true

	// Reset internal variables
	stepper.internalStop = false

	// 0 Means the top level loop, probably Run
	stepper.runLevel(0)

	stepper.isRunning = false
}

// runLevel implements nested for loops recursively. It is set up so that it can be stopped and resumed at any point.
func (stepper *Manager) runLevel(currentLevel int) bool {
	st := stepper.Stacks[stepper.Mode]
	if currentLevel >= len(st.Order) {
		return true // Stack overflow, expected at bottom of stack.
	}
	time := st.Order[currentLevel]
	loop := st.Loops[time]
	ctr := &loop.Counter

	for ctr.Cur < ctr.Max || ctr.Max < 0 { // Loop forever for negative maxes
		stopAtLevelOrLarger := st.Order[currentLevel] >= stepper.StopLevel // Based on conversion of etime.Times to int
		if stepper.StopFlag && stopAtLevelOrLarger {
			stepper.internalStop = true
		}
		if stepper.internalStop {
			// This should occur before ctr incrementing and before functions.
			stepper.StopFlag = false
			return false // Don't continue above, e.g. Stop functions
		}
		if stepper.StopNext && st.Order[currentLevel] == stepper.StopLevel {
			stepper.StepIterations -= 1
			if stepper.StepIterations <= 0 {
				stepper.StopNext = false
				stepper.StopFlag = true // Stop at the top of the next StopLevel
			}
		}

		// Don't ever Start the same iteration of the same level twice.
		lastCtr, ok := stepper.lastStartedCtr[etime.Scope(stepper.Mode, time)]
		if !ok || ctr.Cur > lastCtr {
			stepper.lastStartedCtr[etime.Scope(stepper.Mode, time)] = ctr.Cur
			if PrintControlFlow && time >= NoPrintBelow {
				fmt.Println(time.String() + ":Start:" + strconv.Itoa(ctr.Cur))
			}
			// Events occur at the very start.
			stepper.eventLogic(loop)
			for _, fun := range loop.OnStart {
				fun.Func()
			}
		} else if PrintControlFlow && time >= NoPrintBelow {
			fmt.Println("Skipping start: " + time.String() + ":" + strconv.Itoa(ctr.Cur))
		}

		// Recursion!
		runComplete := stepper.runLevel(currentLevel + 1)

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
			for name, fun := range loop.IsDone {
				if fun() {
					_ = name      // For debugging
					goto exitLoop // Exit multiple for-loops without flag variable.
				}
			}
			ctr.Cur = ctr.Cur + 1 // Increment
		}
	}

exitLoop:
	// Only get to this point if this loop is done.
	if !stepper.internalStop {
		ctr.Cur = 0
		stepper.lastStartedCtr[etime.Scope(stepper.Mode, time)] = -1
	}
	return true
}

// eventLogic handles events that occur at specific timesteps.
func (stepper *Manager) eventLogic(loop *Loop) {
	ctr := &loop.Counter
	for _, phase := range loop.Events {
		if ctr.Cur == phase.AtCtr {
			for _, function := range phase.OnEvent {
				function.Func()
			}
		}
	}
}
