package looper

import (
	"strconv"
)

// A Span represents a length of time within a loop, if behavior is expected to change in distinct phases.
type Span struct {
	Name     string     `desc:"Might be 'plus' or 'minus' for example."`
	Duration int        `desc:"The length of this Span."`
	OnStart  NamedFuncs `desc:"Called at the start of the Span."`
	OnEnd    NamedFuncs `desc:"Called at the end of the Span."`
}

// String describes the Span in human readable text.
func (loopSegment Span) String() string {
	s := loopSegment.Name + ": "
	s = s + "(duration=" + strconv.Itoa(loopSegment.Duration) + ") "
	if len(loopSegment.OnStart) > 0 {
		s = s + "\tOnStart: " + loopSegment.OnStart.String()
	}
	if len(loopSegment.OnEnd) > 0 {
		s = s + "\tOnEnd: " + loopSegment.OnEnd.String()
	}
	return s
}
