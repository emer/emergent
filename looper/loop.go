// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

type Loop struct {
	OnStart NamedFuncs
	// Either Main or the inner loop occurs between OnStart and OnEnd
	Main   NamedFuncs
	OnEnd  NamedFuncs
	IsDone map[string]func() bool `desc:"If true, end loop. Maintained as an unordered map because they should not have side effects."`

	Segments []LoopSegment `desc:"Only use Phases at the Theta Cycle timescale (200ms)."`

	Counter Ctr `desc:"Tracks time within the loop. Also tracks the maximum."`
}

func (loops *Loop) AddSegments(loopSegments ...LoopSegment) {
	for _, loopSegment := range loopSegments {
		loops.Segments = append(loops.Segments, loopSegment)
		loopSegment.OnStart = NamedFuncs{}
		loopSegment.OnEnd = NamedFuncs{}
	}
}
