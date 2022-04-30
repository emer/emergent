package looper

import "github.com/emer/emergent/envlp"

type Phase struct {
	Name             string // Might be plus or minus for example
	Duration         int
	IsPlusPhase      bool
	OnMillisecondEnd orderedMapFuncs
	PhaseStart       orderedMapFuncs
	PhaseEnd         orderedMapFuncs

	Counter *envlp.Ctr `desc:"Tracks time within the loop. Also tracks the maximum."`
}
