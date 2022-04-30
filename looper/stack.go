// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import (
	"github.com/emer/emergent/envlp"
	"github.com/emer/emergent/etime"
)

type EvaluationModeLoops struct {
	Loops map[etime.Times]*LoopStructure
	Order []etime.Times // This should be managed internally.
}

func (loops *EvaluationModeLoops) Init() *EvaluationModeLoops {
	loops.Loops = map[etime.Times]*LoopStructure{}
	return loops
}

func (loops *EvaluationModeLoops) AddTimeScales(times ...etime.Times) *EvaluationModeLoops {
	if loops.Loops == nil {
		loops.Loops = map[etime.Times]*LoopStructure{}
	}
	for _, time := range times {
		loops.Loops[time] = &LoopStructure{}
		loops.Order = append(loops.Order, time)
	}
	return loops
}

func (loops *EvaluationModeLoops) AddTime(time etime.Times, max int) *EvaluationModeLoops {
	loops.Loops[time] = &LoopStructure{Counter: &envlp.Ctr{Max: max}, IsDone: map[string]func() bool{}}
	loops.Order = append(loops.Order, time)
	return loops
}
