// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import (
	"strconv"
)

// A Event has function(s) that can be called at a particular point
// in the loop, when the counter is AtCtr value.
type Event struct {

	// Might be 'plus' or 'minus' for example.
	Name string `desc:"Might be 'plus' or 'minus' for example."`

	// The counter value upon which this Event occurs.
	AtCtr int `desc:"The counter value upon which this Event occurs."`

	// Callback function for the Event.
	OnEvent NamedFuncs `desc:"Callback function for the Event."`
}

// String describes the Event in human readable text.
func (event *Event) String() string {
	s := event.Name + ": "
	s = s + "(at " + strconv.Itoa(event.AtCtr) + ") "
	if len(event.OnEvent) > 0 {
		s = s + "\tEvents: " + event.OnEvent.String()
	}
	return s
}

// NewEvent returns a new event with given name, function, at given counter
func NewEvent(name string, atCtr int, fun func()) *Event {
	ev := &Event{Name: name, AtCtr: atCtr}
	ev.OnEvent.Add(name, fun)
	return ev
}
