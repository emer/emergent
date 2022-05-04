package looper

import (
	"strconv"
)

type LoopSegment struct {
	Name        string // Might be plus or minus for example
	Duration    int
	IsPlusPhase bool
	PhaseStart  NamedFuncs
	PhaseEnd    NamedFuncs
}

func (phase LoopSegment) String() string {
	s := phase.Name + ": "
	s = s + "(duration=" + strconv.Itoa(phase.Duration) + ") "
	if len(phase.PhaseStart) > 0 {
		s = s + "\tOnStart: " + phase.PhaseStart.String()
	}
	if len(phase.PhaseEnd) > 0 {
		s = s + "\tOnEnd: " + phase.PhaseEnd.String()
	}
	return s
}
