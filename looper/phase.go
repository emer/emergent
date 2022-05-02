package looper

import (
	"github.com/emer/emergent/envlp"
	"strconv"
)

type Phase struct {
	Name             string // Might be plus or minus for example
	Duration         int
	IsPlusPhase      bool
	OnMillisecondEnd orderedMapFuncs
	PhaseStart       orderedMapFuncs
	PhaseEnd         orderedMapFuncs

	Counter *envlp.Ctr `desc:"Tracks time within the loop. Also tracks the maximum."`
}

func (phase Phase) String() string {
	s := phase.Name + ": "
	s = s + "(duration=" + strconv.Itoa(phase.Duration) + ") "
	if len(phase.PhaseStart) > 0 {
		s = s + "\tOnStart: " + phase.PhaseStart.String()
	}
	if len(phase.OnMillisecondEnd) > 0 {
		s = s + "\tOnMsEnd: " + phase.OnMillisecondEnd.String()
	}
	if len(phase.PhaseEnd) > 0 {
		s = s + "\tOnEnd: " + phase.PhaseEnd.String()
	}
	return s
}
