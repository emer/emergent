// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import (
	"github.com/emer/emergent/envlp"
)

type LoopStructure struct {
	OnStart orderedMapFuncs
	// Either Main or the inner loop occurs between OnStart and OnEnd
	Main   orderedMapFuncs
	OnEnd  orderedMapFuncs
	IsDone map[string]func() bool `desc:"If true, end loop. Maintained as an unordered map because they should not have side effects."`

	Phases []Phase `desc:"Only use Phases at the Theta Cycle timescale (200ms)."`
	// TODO Add an axon.time here but move it to etimes

	Counter *envlp.Ctr `desc:"Tracks time within the loop. Also tracks the maximum."`
}

func (loops *LoopStructure) AddPhases(phases ...Phase) {
	for _, phase := range phases {
		loops.Phases = append(loops.Phases, phase)
		phase.OnMillisecondEnd = orderedMapFuncs{}
		phase.PhaseStart = orderedMapFuncs{}
		phase.PhaseEnd = orderedMapFuncs{}
	}
}
