package looper

import (
	"strconv"
)

type LoopSegment struct {
	Name     string // Might be "plus" or "minus" for example
	Duration int
	OnStart  NamedFuncs
	OnEnd    NamedFuncs
}

func (loopSegment LoopSegment) String() string {
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
