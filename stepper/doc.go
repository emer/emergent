// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package `stepper` allows you to set StepPoints in simulation code that will pause if some condition is satisfied.
While paused, the simulation waits for the top-level process (the user interface) to tell it to continue.
Once a continue notification is received, the simulation continues on its way, with all internal state
exactly as it was when the StepPoint was hit, without having to explicitly save anything.

There are two "running" states, Stepping and Running. The difference is that in the Running state, unless
there is a Stop request, the application will forego the possibly-complex checking for a pause. The actual StepPoint
function is written to make checking as quick as possible. Although the program
will not stop at StepPoints without interaction, it will pause if RunState is Paused. The main difference
between Paused and Stopped is that in the Paused state, the application waits for a state change, whereas in the
Stopped state, the Stepper exits, and no application state is preserved. After entering Stopped, the controlling
program (i.e., the user interface) should make sure that everything is properly reinitialized before running again.

Usage Basics

The Stepper struct includes an integer field, "StepGrain", which controls whether it will actually pause. The StepPoint
function checks that its argument is equal to the current StepGrain, and if so, calls the PauseNotifyFn callback
with whatever state information was set up when the PauseNotifyFn was registered. It then enters a loop, waiting
for a change in the RunState field of the Stepper. If it sees that RunState has become Stopped, it return true,
and the caller (i.e., the simulation) should exit the current run. If it sees
that RunState has changed to either Running or Stepping, it returns false, and the caller should continue.

Before running, the caller should call Stepper.New() to get a fresh, initialized Stepper, and must also call
RegisterPauseNotifyFn to designate a callback to be invoked when StepPoint is called with a value of grain that
matches the value of StepGrain set in the struct. Internally, StepGrain is just an integer, but the intention is that
callers will define a StepGrain enumerated type to designate meaningful points at which to pause. Note that the
current implementation makes no use of the actual values of StepGrain, i.e., the fact that one value is greater than
another has no effect. This might change in a future version.

In addition to the required PauseNotifyFn callback, there is an optional StopCheckFn callback.
This callback can check any state information that it likes, and if it returns true, the PauseNotifyFn will be invoked.

Whether or not a StopCheckFn has been set, if RunState is Stepping and the grain argument matches StepGrain,
StepPoint decrements the value of StepsPer. If stepsLeft goes to zero, PauseNotifyFn is called, and the Stepper goes into
a wait loop, waiting for RunState to be something other than Paused. If RunState becomes Stopped, StepPoint exits with
a value of true, indicating that the caller should end the current run. If the new state is either Running or Stepping,
StepPoint returns false, indicating that the caller should continue.

*/
package stepper
