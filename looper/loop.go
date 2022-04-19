// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import (
	"github.com/emer/emergent/etime"
	"github.com/goki/ki/indent"
)

//THIS IS EXPERIMENTAL and UNFINISHED AND LIKELY WILL CHANGE
type LoopPart string

const (
	End  LoopPart = "END"
	Stop LoopPart = "STOP"
	Main LoopPart = "MAIN"
)

func (loop LoopPart) String(part etime.Times) string {
	return string(part) + ":" + string(loop)
}

// Loop represents one level of looping, with arbitrary functions
// called at 3 different points in the loop, corresponding to a
// do..while loop logic, with no initialization, which is necessary
// to ensure reentrant steppability.  In Go, the logic looks like this:
//
// for {
//    for { <subloops here> } // drills down levels for each subloop
//    Main()                  // Main is called after subloops -- increment counters!
//    if Stop() {
//        break
//    }
// }
// End()                      // Reset counters here so next pass starts over
//
type Loop struct {
	Stack *Stack         `desc:"stack that owns this loop"`
	Scope etime.ScopeKey `desc:"scope level of this loop"`
	Main  Funcs          `desc:"main functions to call inside each iteration, after looping at lower level for non-terminal levels -- any counters should be incremented here -- if there is an Env set for the Stack, then any counter in the Env at the corresponding Scope will automatically be incremented via Env:Incr or Env:Step functions added automatically"`
	Stop  BoolFuncs      `desc:"functions that cause the loop to stop -- if any return true, it stops"`
	End   Funcs          `desc:"functions to run at the end of the loop, after it has stopped.  counters etc should be reset here, so next iteration starts over afresh.  the Init function calls these to initialize before running."`
}

func NewLoop(sc etime.ScopeKey, st *Stack) *Loop {
	return &Loop{Scope: sc, Stack: st}
}

// StageString returns a string for given stage of loop, indented to level
func (lp *Loop) StageString(stage string, level int) string {
	return indent.Spaces(level, IndentSize) + string(lp.Scope) + ": " + stage
}
