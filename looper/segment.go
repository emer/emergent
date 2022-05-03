package looper

import (
	"github.com/emer/emergent/envlp"
	"strconv"
)

type LoopSegment struct {
	Name        string // Might be plus or minus for example
	Duration    int
	IsPlusPhase bool
	PhaseStart  NamedFuncs
	PhaseEnd    NamedFuncs

	Counter *envlp.Ctr `desc:"Tracks time within the loop. Also tracks the maximum."`
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
