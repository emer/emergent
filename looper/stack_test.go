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
	trn := NewStack(etime.Train, etime.Run, etime.Epoch, etime.Trial)
	set.AddStack(trn)

	run := 0
	lp := trn.Loop(etime.Run)
	lp.Main.Add(func() {
		run++
		fmt.Printf("Run Main: %d\n", run)
	})
	lp.Stop.Add(func() bool {
		if run >= 2 {
			fmt.Printf("Run Stop: %d\n", run)
			return true
		}
		return false
	})
	lp.End.Add(func() {
		run = 0
		fmt.Printf("Run End: %d\n", run)
	})

	epoch := 0
	lp = trn.Loop(etime.Epoch)
	lp.Main.Add(func() {
		epoch++
		fmt.Printf("\tEpoch Main: %d\n", epoch)
	})
	lp.Stop.Add(func() bool {
		if epoch >= 3 {
			fmt.Printf("\tEpoch Stop: %d\n", epoch)
			return true
		}
		return false
	})
	lp.End.Add(func() {
		epoch = 0
		fmt.Printf("\tEpoch End: %d\n", epoch)
	})

	trial := 0
	lp = trn.Loop(etime.Trial)
	lp.Main.Add(func() {
		trial++
		fmt.Printf("\t\tTrial Main: %d\n", trial)
	})
	lp.Stop.Add(func() bool {
		if trial >= 3 {
			fmt.Printf("\t\tTrial Stop: %d\n", trial)
			return true
		}
		return false
	})
	lp.End.Add(func() {
		trial = 0
		fmt.Printf("\t\tTrial End: %d\n", trial)
	})

	set.Run(etime.Train, etime.Run)

	// stepping
	fmt.Printf("\n##############\nStep Trial 1\n")
	run = 0
	epoch = 0
	trial = 0
	set.Step(etime.Train, etime.Run, etime.Trial, 1)
	// stepping
	fmt.Printf("\n##############\nStep Trial 1\n")
	set.Run(etime.Train, etime.Run)
	fmt.Printf("\n##############\nStep Trial 1\n")
	set.Run(etime.Train, etime.Run)
	fmt.Printf("\n##############\nStep Trial 1\n")
	set.Run(etime.Train, etime.Run)

	// stepping
	fmt.Printf("\n##############\nStep Trial 2\n")
	run = 0
	epoch = 0
	trial = 0
	set.Step(etime.Train, etime.Run, etime.Trial, 2)
	// stepping
	fmt.Printf("\n##############\nStep Trial 2\n")
	set.Run(etime.Train, etime.Run)
	fmt.Printf("\n##############\nStep Trial 2\n")
	set.Run(etime.Train, etime.Run)
	fmt.Printf("\n##############\nStep Trial 2\n")
	set.Run(etime.Train, etime.Run)

	// stepping
	fmt.Printf("\n##############\nStep Epoch 1\n")
	run = 0
	epoch = 0
	trial = 0
	set.Step(etime.Train, etime.Run, etime.Epoch, 1)
	// stepping
	fmt.Printf("\n##############\nStep Epoch 1\n")
	set.Run(etime.Train, etime.Run)
	fmt.Printf("\n##############\nStep Epoch 1\n")
	set.Run(etime.Train, etime.Run)
	fmt.Printf("\n##############\nStep Epoch 1\n")
	set.Run(etime.Train, etime.Run)
}
