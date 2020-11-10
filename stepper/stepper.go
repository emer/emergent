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
	"sync"
	"time"

	"github.com/goki/ki/kit"
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

// A StopCheckFn is a callback to check whether an arbitrary condition has been matched.
// If a StopCheckFn returns true, the program is suspended with a RunState of Paused,
// and will remain so until the RunState changes to Stepping, Running, or Stopped.
// As noted below for the PauseNotifyFn, the StopCheckFn is called with the Stepper's lock held.
type StopCheckFn func(grain int) (matched bool)

// A PauseNotifyFn is a callback that will be invoked if the program enters the Paused state.
// NOTE! The PauseNotifyFn is called with the Stepper's lock held, so it must not call any Stepper methods
// that try to take the lock on entry, or a deadlock will result.
type PauseNotifyFn func()

// The Stepper struct contains all of the state info for stepping a program, enabling step points.
// where the running application can be suspended with no loss of state.
type Stepper struct {
	RunState      RunState      `desc:"current run state"`
	StepGrain     int           `desc:"granularity of one step. No enum type here so clients can define their own"`
	StepsPer      int           `desc:"number of steps to execute before returning"`
	PauseNotifyFn PauseNotifyFn `view:"-" desc:"function to deal with any changes on client side when paused after stepping"`
	StopCheckFn   StopCheckFn   `view:"-" desc:"function to test for special stopping conditions"`
	stateMut      sync.Mutex    `view:"-" desc:"mutex for RunState"`
	stateChange   *sync.Cond    `view:"-" desc:"state change condition variable"`
	stepsLeft     int           `view:"-" desc:"number of steps yet to execute before returning"`
	waitTimer     chan RunState `desc:"watchdog timer channel"`
	initOnce      sync.Once     `view:"-" desc:"this ensures that global initialization only happens once"`
}

// New makes a new Stepper. Always call this to create a Stepper, so that initialization will be run correctly.
func New() *Stepper { return new(Stepper).Init() }

// Init puts everything into a good state before starting a run
// Init is called automatically by New, and should be called before running again after calling Stop (not Pause).
// Init should not be called explicitly when creating a new Stepper--the preferred way to initialize is to call New.
func (st *Stepper) Init() *Stepper {
	st.initOnce.Do(func() {
		st.stateMut = sync.Mutex{}
		st.stateChange = sync.NewCond(&st.stateMut)
		st.RunState = Stopped
		st.StepGrain = 0 // probably an enum, but semantics are up to the client program
		st.stepsLeft = 0
		st.StepsPer = 1
		st.waitTimer = make(chan RunState, 1)
	})
	return st
}

// Reset StepsPer and StepGrain parameters
func (st *Stepper) ResetParams(nSteps int, grain int) {
	st.StepsPer = nSteps
	st.stepsLeft = nSteps
	st.StepGrain = grain
}

// Enter unconditionally enters the specified RunState. It broadcasts a stateChange, which should be picked
// up by a paused application.
func (st *Stepper) Enter(state RunState) {
	st.stateMut.Lock()
	defer st.stateMut.Unlock()
	st.RunState = state
	st.stateChange.Broadcast()
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
	st.stateMut.Lock()
	defer st.stateMut.Unlock()
	return st.RunState == Running || st.RunState == Stepping
}

// Start enters the Stepping run state. This should be called at the start of a run only.
func (st *Stepper) Start(grain int, nSteps int) {
	st.stateMut.Lock()
	defer st.stateMut.Unlock()
	if nSteps > 0 {
		st.StepsPer = nSteps
		st.stepsLeft = nSteps
	}
	st.StepGrain = grain
	st.RunState = Stepping
}

// StepPoint checks for possible pause or stop.
// If the application is:
// Running: keep going with no further examination of state.
// Stopped: return true, and the application should return (i.e., stop completely).
// Stepping: check to see if we should pause (if StepGrain matches, decrement stepsLeft, stop if <= 0).
// Paused: wait for state change.
func (st *Stepper) StepPoint(grain int) (stop bool) {
	st.stateMut.Lock()
	defer st.stateMut.Unlock()
	if st.RunState == Stopped {
		return true
	}
	if st.RunState == Running {
		return false
	}
	if st.RunState != Paused && grain == st.StepGrain { // exact equality is the only test that really works well
		if st.pauseIfStepsComplete() {
			st.PauseNotifyFn()
		}
	}
	if st.StopCheckFn != nil {
		stopMatched := st.StopCheckFn(grain)
		if stopMatched {
			st.RunState = Paused
			st.PauseNotifyFn()
		}
	}
	for {
		switch st.RunState {
		case Stopped:
			return true
		case Running, Stepping:
			return false
		case Paused:
			st.waitWithTimeout(st.stateChange, 10)
		}
	}
}

// PauseIfStepsComplete counts down stepsLeft, and pauses if they go to zero.
func (st *Stepper) pauseIfStepsComplete() (pauseNow bool) {
	st.stepsLeft--
	if st.stepsLeft <= 0 {
		st.RunState = Paused
		st.stepsLeft = st.StepsPer
		return true
	} else {
		return false
	}
}

// Watchdog timer for stateChange. Go Wait never times out, so this artificially injects a stateChange
// event to keep processes from getting stuck.
func (st *Stepper) waitWithTimeout(cond *sync.Cond, secs int) {
	go func() {
		cond.Wait()
		st.waitTimer <- st.RunState
	}()
	for {
		select {
		case <-st.waitTimer:
			return
		case <-time.After(time.Duration(secs) * time.Second):
			cond.Broadcast()
		}
	}
}
