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
