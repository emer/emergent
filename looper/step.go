// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import (
	"fmt"
	"github.com/emer/emergent/etime"
	"strconv"
)

// If you want to debug the flow of time, set this to true.
var printControlFlow = false

// Stepper is a control object for stepping through a looper.DataManager filled with Stacks of Loops. It doesn't directly contain any data about what the flow of nested loops should be, only data about how the flow is going.
type Stepper struct {
	StopFlag       bool         `desc:"If true, stop model ASAP."`
	StopNext       bool         `desc:"If true, stop model at the end of the current StopLevel."`
	StopLevel      etime.Times  `desc:"Time level to stop at the end of."`
	StepIterations int          `desc:"How many steps to do."`
	Loops          *DataManager `desc:"The information about loops."`
	Mode           etime.Modes  `desc:"The current evaluation mode."`
	isRunning      bool         `desc:"Set to true while looping, false when done. Read only."`

	// For internal use
	lastStartedCtr map[etime.ScopeKey]int `desc:"The Cur value of the Ctr associated with the last started level, for each timescale."`
	internalStop   bool
}

// IsRunning is True if running.
func (stepper Stepper) IsRunning() bool {
	return stepper.isRunning
}

// Init sets some default values.
func (stepper *Stepper) Init(loopman *DataManager) {
	stepper.Loops = loopman
	stepper.StopLevel = etime.Run
	stepper.Mode = etime.Train
	stepper.lastStartedCtr = map[etime.ScopeKey]int{}
}

// Run runs the loops contained within the stepper. If you want it to stop before the full end of the loop, set variables on the Stepper.
func (stepper *Stepper) Run() {
	stepper.isRunning = true

	// Reset internal variables
	stepper.internalStop = false

	// 0 Means the top level loop, probably Run
	stepper.runLevel(0)

	stepper.isRunning = false
}

// runLevel implements nested for loops recursively. It is set up so that it can be stopped and resumed at any point.
func (stepper *Stepper) runLevel(currentLevel int) bool {
	st := stepper.Loops.Stacks[stepper.Mode]
	if currentLevel >= len(st.Order) {
		return true // Stack overflow, expected at bottom of stack.
	}
	time := st.Order[currentLevel]
	loop := st.Loops[time]
	ctr := &loop.Counter

	for ctr.Cur < ctr.Max || ctr.Max < 0 { // Loop forever for negative maxes
		stopAtLevel := st.Order[currentLevel] == stepper.StopLevel // Based on conversion of etime.Times to int
		if stepper.StopFlag && stopAtLevel {
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
				stepper.StopFlag = true
			}
		}

		// Don't ever Start the same iteration of the same level twice.
		lastCtr, ok := stepper.lastStartedCtr[etime.Scope(stepper.Mode, time)]
		if !ok || ctr.Cur > lastCtr {
			stepper.lastStartedCtr[etime.Scope(stepper.Mode, time)] = ctr.Cur
			if printControlFlow && time > etime.Trial {
				fmt.Println(time.String() + ":Start:" + strconv.Itoa(ctr.Cur))
			}
			for _, fun := range loop.OnStart {
				fun.Func()
			}
		} else if printControlFlow && time > etime.Trial {
			fmt.Println("Skipping start: " + time.String() + ":" + strconv.Itoa(ctr.Cur))
		}

		// Recursion!
		stepper.eventLogic(loop)
		runComplete := stepper.runLevel(currentLevel + 1)

		if runComplete {
			for _, fun := range loop.Main {
				fun.Func()
			}
			if printControlFlow && time > etime.Trial {
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
func (stepper *Stepper) eventLogic(loop *Loop) {
	ctr := &loop.Counter
	for _, phase := range loop.Events {
		if ctr.Cur == phase.OccurTime {
			for _, function := range phase.OnOccur {
				function.Func()
			}
		}
	}
}
