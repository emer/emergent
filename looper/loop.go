// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

// Loop contains one level of a multi-level iteration scheme.
// It wraps around an inner loop recorded in a Stack, or around Main functions.
// It records how many times the loop should be repeated in the Counter.
// It records what happens at the beginning and end of each loop.
// For example, a loop with 1 start, 1 end, and a Counter with max=3 will do:
// Start, Inner, End, Start, Inner, End, Start, Inner, End
// Where the Inner loop is specified by a Stack or by Main,
// and Start and End are functions on the loop.
// See Stack for more details on how loops are combined.
type Loop struct {

	// Tracks time within the loop. Also tracks the maximum. OnStart, Main, and OnEnd will be called Counter.Max times, or until IsDone is satisfied, whichever ends first.
	Counter Counter

	// OnStart is called at the beginning of each loop.
	OnStart NamedFuncs

	// OnStart is called in the middle of each loop. In general, only use Main for the last Loop in a Stack. For example, actual Net updates might occur here.
	Main NamedFuncs

	// OnStart is called at the end of each loop.
	OnEnd NamedFuncs

	// If true, end loop. Maintained as an unordered map because they should not have side effects.
	IsDone NamedFuncsBool

	// Events occur when Counter.Cur gets to their AtCounter.
	Events []*Event
}

// AddEvents to the list of events.
func (lp *Loop) AddEvents(events ...*Event) {
	for _, event := range events {
		lp.Events = append(lp.Events, event)
	}
}

// AddNewEvent to the list.
func (lp *Loop) AddNewEvent(name string, atCtr int, fun func()) *Event {
	ev := NewEvent(name, atCtr, fun)
	lp.Events = append(lp.Events, ev)
	return ev
}

// EventByName returns event by name, nil if not found
func (lp *Loop) EventByName(name string) *Event {
	for _, ev := range lp.Events {
		if ev.Name == name {
			return ev
		}
	}
	return nil
}

// SkipToMax sets the counter to its Max value for this level.
// for skipping over rest of loop
func (lp *Loop) SkipToMax() {
	lp.Counter.SkipToMax()
}
