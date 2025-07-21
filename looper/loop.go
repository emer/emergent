// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import (
	"fmt"
	"strings"
)

// Loop contains one level of a multi-level iteration stack,
// with functions that can be called at the start and end
// of each iteration of the loop, and a Counter that increments
// for each iteration, terminating if >= Max, or IsDone returns true.
// Within each iteration, any sub-loop at the next level down
// in its [Stack] runs its full set of iterations.
// The control flow is:
//
//	for {
//		Events[Counter == AtCounter] // run events at counter
//		OnStart()
//		    Run Sub-Loop to completion
//		OnEnd()
//		Counter += Inc
//		if Counter >= Max || IsDone() {
//		    break
//		}
//	}
type Loop struct {

	// Counter increments every iteration through the loop, up to [Counter.Max].
	Counter Counter

	// Events occur when Counter.Cur is at their AtCounter.
	Events []*Event

	// OnStart functions are called at the beginning of each loop iteration.
	OnStart NamedFuncs

	// OnEnd functions are called at the end of each loop iteration.
	OnEnd NamedFuncs

	// IsDone functions are called after each loop iteration,
	// and if any return true, then the loop iteration is terminated.
	IsDone NamedFuncs

	// StepCount is the default step count for this loop level.
	StepCount int
}

// NewLoop returns a new loop with given Counter Max and increment.
func NewLoop(ctrMax, ctrIncr int) *Loop {
	lp := &Loop{}
	lp.Counter.SetMaxInc(ctrMax, ctrIncr)
	lp.StepCount = 1
	return lp
}

// AddEvent adds a new event at given counter. If an event already exists
// for that counter, the function is added to the list for that event.
func (lp *Loop) AddEvent(name string, atCtr int, fun func()) *Event {
	ev := lp.EventByCounter(atCtr)
	if ev == nil {
		ev = NewEvent(name, atCtr, fun)
		lp.Events = append(lp.Events, ev)
	} else {
		ev.OnEvent.Add(name, fun)
	}
	return ev
}

// EventByCounter returns event for given atCounter value, nil if not found.
func (lp *Loop) EventByCounter(atCtr int) *Event {
	for _, ev := range lp.Events {
		if ev.AtCounter == atCtr {
			return ev
		}
	}
	return nil
}

// EventByName returns event by name, nil if not found.
func (lp *Loop) EventByName(name string) *Event {
	for _, ev := range lp.Events {
		if ev.Name == name {
			return ev
		}
	}
	return nil
}

// SkipToMax sets the counter to its Max value for this level.
// for skipping over rest of loop.
func (lp *Loop) SkipToMax() {
	lp.Counter.SkipToMax()
}

// DocString returns an indented summary of this loop and those below it.
func (lp *Loop) DocString(st *Stack, level int) string {
	var sb strings.Builder
	ctrs := ""
	if lp.Counter.Inc > 1 {
		ctrs = fmt.Sprintf("[0 : %d : %d]:\n", lp.Counter.Max, lp.Counter.Inc)
	} else {
		ctrs = fmt.Sprintf("[0 : %d]:\n", lp.Counter.Max)
	}
	sb.WriteString(indent(level+1) + st.Order[level].String() + ctrs)
	if len(lp.Events) > 0 {
		sb.WriteString(indent(level+2) + "Events:\n")
		for _, ev := range lp.Events {
			sb.WriteString(indent(level+3) + ev.String() + "\n")
		}
	}
	if len(lp.OnStart) > 0 {
		sb.WriteString(indent(level+2) + "Start:  " + lp.OnStart.String() + "\n")
	}
	if level < len(st.Order)-1 {
		slp := st.Level(level + 1)
		sb.WriteString(slp.DocString(st, level+1))
	}
	if len(lp.OnEnd) > 0 {
		sb.WriteString(indent(level+2) + "End:    " + lp.OnEnd.String() + "\n")
	}
	if len(lp.IsDone) > 0 {
		sb.WriteString(indent(level+2) + "IsDone:  " + lp.IsDone.String() + "\n")
	}
	return sb.String()
}
