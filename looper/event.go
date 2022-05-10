package looper

import (
	"strconv"
)

// A Event represents a length of time within a loop, if behavior is expected to change in distinct phases.
type Event struct {
	Name    string     `desc:"Might be 'plus' or 'minus' for example."`
	AtCtr   int        `desc:"The time that this Event occurs."`
	OnEvent NamedFuncs `desc:"Callback function for the Event."`
}

// String describes the Event in human readable text.
func (event Event) String() string {
	s := event.Name + ": "
	s = s + "(at " + strconv.Itoa(event.AtCtr) + ") "
	if len(event.OnEvent) > 0 {
		s = s + "\tEvents: " + event.OnEvent.String()
	}
	return s
}
