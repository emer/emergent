package looper

import (
	"strconv"
)

// A Event represents a length of time within a loop, if behavior is expected to change in distinct phases.
type Event struct {
	Name      string     `desc:"Might be 'plus' or 'minus' for example."`
	OccurTime int        `desc:"The length of this Event."`
	OnOccur   NamedFuncs `desc:"Called at the start of the Event."`
}

// String describes the Event in human readable text.
func (event Event) String() string {
	s := event.Name + ": "
	s = s + "(at " + strconv.Itoa(event.OccurTime) + ") "
	if len(event.OnOccur) > 0 {
		s = s + "\tEvents: " + event.OnOccur.String()
	}
	return s
}
