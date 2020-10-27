// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
The Stepper package allows you to set StepPoints in simulation code that will pause if some condition is satisfied.
While paused, the simulation waits for the top-level process (the user interface) to tell it to continue.
Once a continue notification is received, the simulation continues on its way, with all internal state
exactly as it was when the StopPoint was hit, without having to explicitly save anything.

There are two "running" states, Stepping and Running. The difference is that in the Running state, unless
there is a Stop request, the application will forego the possibly-complex checking for a pause (see StepPoint,
at the bottom of this file). StepPoint is written to make checking as quick as possible. Although the program
will not stop at StepPoints without interaction, it will pause if RunState is Paused. The main difference
between Paused and Stopped is that in the Paused state, the application waits for a state change, whereas in the
Stopped state, the Stepper exits, and no application state is preserved. After entering Stopped, the controlling
program (i.e., the user interface) should make sure that everything is properly reinitialized before running again.
*/

package stepper

import (
	"github.com/goki/ki/kit"
	"sync"
	"time"
)

type RunState int

const (
	Stopped  RunState = iota // execution is stopped. The Stepper is NOT waiting, so running again is basically a restart. The only way to go from Running or Stepping to Stopped is to explicitly call Stop(). Program state will not be preserved once the Stopped state is entered.
	Paused                   // execution is paused. The sim is waiting for further instructions, and can continue, or stop.
	Stepping                 // the application is running, but will pause if it hits a StepPoint that matches the current StepGrain.
	Running                  // the application is running, and will NOT pause at StepPoints. It will pause if a stop has been requested.
	RunStateN
)

var KiT_RunState = kit.Enums.AddEnum(RunStateN, kit.NotBitFlag, nil)

//go:generate stringer -type=RunState

// A StopConditionChecker is a callback to check whether an arbitrary condition has been matched.
// If a StopConditionChecker returns true, the program is suspended with a RunState of Paused,
// and will remain so until the RunState changes to Stepping, Running, or Stopped.
// As noted below for the PauseNotifier, the StopConditionChecker is called with the Stepper's lock held.
type StopConditionChecker func(sv interface{}, grain int) (matched bool)

// A PauseNotifier is a callback that will be invoked if the program enters the Paused state.
// It takes an arbitrary state variable, sv, which is set by RegisterPauseNotifier.
// NOTE! The PauseNotifier is called with the Stepper's lock held, so it must not call any Stepper methods
// that try to take the lock on entry, or a deadlock will result.
type PauseNotifier func(sv interface{})

// The Stepper struct contains all of the state info for stepping a program, enabling step points.
// where the running application can be suspended with no loss of state.
type Stepper struct {
	StateMut       sync.Mutex           `view:"-" desc:"mutex for RunState"`
	StateChange    *sync.Cond           `view:"-" desc:"state change condition variable"`
	RunState       RunState             `desc:"current run state"`
	StepGrain      int                  `desc:"granularity of one step. No enum type here so clients can define their own"`
	CondChecker    StopConditionChecker `view:"-" desc:"function to test for special stopping conditions"`
	CheckerState   interface{}          `view:"-" desc:"arbitrary state information for condition checker"`
	PauseNotifier  PauseNotifier        `view:"-" desc:"function to deal with any changes on client side when paused after stepping"`
	PNState        interface{}          `view:"-" desc:"arbitrary state information for pause notifier"`
	StepsPerClick  int                  `desc:"number of steps to execute before returning"`
	StepsRemaining int                  `desc:"number of steps yet to execute before returning"`
	WaitTimer      chan RunState        `desc:"watchdog timer channel"`
}

// New makes a new Stepper. Always call this to create a Stepper, so that initialization will be run correctly.
func New() *Stepper { return new(Stepper).Init() }

var oneTimeInit sync.Once // this ensures that global initialization only happens once

// Init puts everything into a good state before starting a run
// Init is called automatically by New, and should be called before running again after calling Stop (not Pause).
// Init should not be called explicitly when creating a new Stepper--the preferred way to initialize is to call New.
func (st *Stepper) Init() *Stepper {
	oneTimeInit.Do(func() {
		st.StateMut = sync.Mutex{}
		st.StateChange = sync.NewCond(&st.StateMut)
		st.RunState = Stopped
		st.StepGrain = 0 // probably an enum, but semantics are up to the client program
		st.StepsRemaining = 0
		st.StepsPerClick = 1
		st.WaitTimer = make(chan RunState, 1)
	})
	return st
}

// Watchdog timer for StateChange. Go Wait never times out, so this artificially injects a StateChange
// event to keep processes from getting stuck.
func (st *Stepper) WaitWithTimeout(cond *sync.Cond, secs int) {
	go func() {
		cond.Wait()
		st.WaitTimer <- st.RunState
	}()
	for {
		select {
		case <-st.WaitTimer:
			return
		case <-time.After(time.Duration(secs) * time.Second):
			cond.Broadcast()
		}
	}
}

// DontStop is a StopConditionChecker that does nothing, i.e., it will never trigger a pause
func DontStop(_ interface{}, _ int) bool {
	return false
}

// SetStepGrain sets the internal value of StepGrain to an uninterpreted int. Semantics are up to the client
func (st *Stepper) SetStepGrain(grain int) {
	st.StepGrain = grain
}

// Grain gets the current StepGrain
func (st *Stepper) Grain() int {
	return st.StepGrain
}

// Enter unconditionally enters the specified RunState. It broadcasts a StateChange, which should be picked
// up by a paused application.
func (st *Stepper) Enter(state RunState) {
	st.StateMut.Lock()
	defer st.StateMut.Unlock()
	st.RunState = state
	st.StateChange.Broadcast()
}

// RegisterStopChecker registers a StopConditionChecker callback. This is completely optional:
// it's fine to not have a StopConditionChecker at all.
func (st *Stepper) RegisterStopChecker(checker StopConditionChecker, cs interface{}) {
	st.CondChecker = checker
	st.CheckerState = cs
}

// RegisterPauseNotifier registers a PauseNotifier callback. A PauseNotifier is not required,
// but is recommended. As an alternative, the controlling code could poll Stepper state periodically.
func (st *Stepper) RegisterPauseNotifier(notifier PauseNotifier, pnState interface{}) {
	st.PauseNotifier = notifier
	st.PNState = pnState
}

// Stop sets RunState to Stopped. The running program will exit at the next StepPoint it hits.
func (st *Stepper) Stop() {
	st.Enter(Stopped)
}

// Pause sets RunState to Paused. The running program will actually pause at the next StepPoint call.
func (st *Stepper) Pause() {
	st.Enter(Paused)
}

// Active checks that the application is either Running or Stepping (neither Paused nor Stopped).
// This needs to use the State mutex because it checks two different fields.
func (st *Stepper) Active() bool {
	st.StateMut.Lock()
	defer st.StateMut.Unlock()
	return st.RunState == Running || st.RunState == Stepping
}

// StartStepping enters the Stepping run state. This should be called at the start of a run only.
func (st *Stepper) StartStepping(grain int, nSteps int) {
	st.StateMut.Lock()
	defer st.StateMut.Unlock()
	if nSteps > 0 {
		st.StepsPerClick = nSteps
		st.StepsRemaining = nSteps
	}
	st.StepGrain = grain
	st.RunState = Stepping
}

// SetNSteps sets the number of times to go through a StepPoint of the current granularity before actually pausing.
func (st *Stepper) SetNSteps(toTake int) {
	st.StateMut.Lock()
	defer st.StateMut.Unlock()
	st.StepsPerClick = toTake
	st.StepsRemaining = toTake
}

// StepPoint checks for possible pause or stop.
// If the application is:
// Running: keep going with no further examination of state.
// Stopped: return true, and the application should return (i.e., stop completely).
// Stepping: check to see if we should pause (if StepGrain matches, decrement StepsRemaining, stop if <= 0).
// Paused: wait for state change.
func (st *Stepper) StepPoint(grain int) (stop bool) {
	st.StateMut.Lock()
	defer st.StateMut.Unlock()
	if st.RunState == Stopped {
		return true
	}
	if st.RunState == Running {
		return false
	}
	if st.RunState != Paused && grain == st.StepGrain { // exact equality is the only test that really works well
		if st.PauseIfStepsComplete() {
			st.PauseNotifier(st.PNState)
		}
	}
	if st.CondChecker != nil {
		stopMatched := st.CondChecker(st.CheckerState, grain)
		if stopMatched {
			st.RunState = Paused
		}
	}
	for {
		switch st.RunState {
		case Stopped:
			return true
		case Running, Stepping:
			return false
		case Paused:
			st.WaitWithTimeout(st.StateChange, 10)
		}
	}
}

// PauseIfStepsComplete counts down StepsRemaining, and pauses if they go to zero.
func (st *Stepper) PauseIfStepsComplete() (pauseNow bool) {
	st.StateMut.Lock()
	defer st.StateMut.Unlock()
	st.StepsRemaining--
	if st.StepsRemaining <= 0 {
		st.RunState = Paused
		st.StepsRemaining = st.StepsPerClick
		return true
	} else {
		return false
	}
}
