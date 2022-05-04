// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import (
	"github.com/emer/emergent/etime"
)

type LoopStack struct {
	Loops map[etime.Times]*Loop
	Order []etime.Times // This should be managed internally.
}

func (loops *LoopStack) Init() *LoopStack {
	loops.Loops = map[etime.Times]*Loop{}
	return loops
}

func (loops *LoopStack) AddTimeScales(times ...etime.Times) *LoopStack {
	if loops.Loops == nil {
		loops.Loops = map[etime.Times]*Loop{}
	}
	for _, time := range times {
		loops.Loops[time] = &Loop{}
		loops.Order = append(loops.Order, time)
	}
	return loops
}

func (loops *LoopStack) AddTime(time etime.Times, max int) *LoopStack {
	loops.Loops[time] = &Loop{Counter: Ctr{Max: max}, IsDone: map[string]func() bool{}}
	loops.Order = append(loops.Order, time)
	return loops
}
