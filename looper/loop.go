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
	Counter Ctr            `desc:"Tracks time within the loop. Also tracks the maximum. OnStart, Main, and OnEnd will be called Ctr.Max times, or until IsDone is satisfied, whichever ends first."`
	OnStart NamedFuncs     `desc:"OnStart is called at the beginning of each loop."`
	Main    NamedFuncs     `desc:"OnStart is called in the middle of each loop. In general, only use Main for the last Loop in a Stack. For example, actual Net updates might occur here."`
	OnEnd   NamedFuncs     `desc:"OnStart is called at the end of each loop."`
	IsDone  NamedFuncsBool `desc:"If true, end loop. Maintained as an unordered map because they should not have side effects."`
	Events  []*Event       `desc:"Events occur when Ctr.Cur gets to their AtCtr."`
}

// AddEvents to the list of events.
func (loops *Loop) AddEvents(events ...*Event) {
	for _, event := range events {
		loops.Events = append(loops.Events, event)
	}
}

// AddNewEvent to the list.
func (loops *Loop) AddNewEvent(name string, atCtr int, fun func()) *Event {
	ev := NewEvent(name, atCtr, fun)
	loops.Events = append(loops.Events, ev)
	return ev
}
