// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import (
	"fmt"
	"testing"

	"github.com/emer/emergent/etime"
)

func TestStack(t *testing.T) {
	set := NewSet()
	trn := NewStack(etime.Train.String(), etime.Run, etime.Epoch, etime.Trial)
	set.AddStack(trn)
	trn.Step.LoopTrace = true
	// trn.Step.FuncTrace = true

	run := 0
	lp := trn.Loop(etime.Run)
	lp.Main.Add("TestRun:Main", func() {
		run++
	})
	lp.Stop.Add("TestRun:Stop", func() bool {
		return run >= 2
	})
	lp.End.Add("TestRun:End", func() {
		run = 0
	})

	epoch := 0
	lp = trn.Loop(etime.Epoch)
	lp.Main.Add("TestEpoch:Main", func() {
		epoch++
	})
	lp.Stop.Add("TestEpoch:Stop", func() bool {
		if epoch >= 3 {
			return true
		}
		return false
	})
	lp.End.Add("TestEpoch:End", func() {
		epoch = 0
	})

	trial := 0
	lp = trn.Loop(etime.Trial)
	lp.Main.Add("TestTrial:Main", func() {
		trial++
	})
	lp.Stop.Add("TestTrial:Stop", func() bool {
		if trial >= 3 {
			return true
		}
		return false
	})
	lp.End.Add("TestTrial:End", func() {
		trial = 0
	})

	fmt.Println(trn.DocString())
	fmt.Println("##########################")

	set.Run(etime.Train)

	fmt.Printf("\n##############\nInit\n")
	set.Init(etime.Train)

	// stepping
	fmt.Printf("\n##############\nStep Trial 1\n")
	set.Step(etime.Train, etime.Trial, 1)
	// stepping
	fmt.Printf("\n##############\nStep Trial 1\n")
	set.Run(etime.Train)
	fmt.Printf("\n##############\nStep Trial 1\n")
	set.Run(etime.Train)
	fmt.Printf("\n##############\nStep Trial 1\n")
	set.Run(etime.Train)

	// stepping
	fmt.Printf("\n##############\nInit\n")
	set.Init(etime.Train)

	fmt.Printf("\n##############\nStep Trial 2\n")
	set.Step(etime.Train, etime.Trial, 2)
	// stepping
	fmt.Printf("\n##############\nStep Trial 2\n")
	set.Run(etime.Train)
	fmt.Printf("\n##############\nStep Trial 2\n")
	set.Run(etime.Train)
	fmt.Printf("\n##############\nStep Trial 2\n")
	set.Run(etime.Train)

	// stepping
	fmt.Printf("\n##############\nInit\n")
	set.Init(etime.Train)

	fmt.Printf("\n##############\nStep Epoch 1\n")
	set.Step(etime.Train, etime.Epoch, 1)
	// stepping
	fmt.Printf("\n##############\nStep Epoch 1\n")
	set.Run(etime.Train)
	fmt.Printf("\n##############\nStep Epoch 1\n")
	set.Run(etime.Train)
	fmt.Printf("\n##############\nStep Epoch 1\n")
	set.Run(etime.Train)
}
