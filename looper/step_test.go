// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import (
	"fmt"
	"testing"

	"github.com/emer/emergent/v2/etime"
)

var printTest = false

func TestStep(t *testing.T) {
	trialCount := 0

	manager := NewManager()
	manager.AddStack(etime.Train).AddTime(etime.Run, 2).AddTime(etime.Epoch, 5).AddTime(etime.Trial, 4).AddTime(etime.Cycle, 3)
	manager.GetLoop(etime.Train, etime.Trial).OnStart.Add("Count Trials", func() { trialCount += 1 })
	manager.GetLoop(etime.Train, etime.Run).OnEnd.Add("Counters Test", func() {
		run := manager.Stacks[etime.Train].Loops[etime.Run].Counter.Cur
		epc := manager.Stacks[etime.Train].Loops[etime.Epoch].Counter.Cur
		if epc != 5 {
			t.Errorf("Run %d OnEnd epoch counter should be 5, not: %d", run, epc)
		}
	})

	run := manager.Stacks[etime.Train].Loops[etime.Run]
	epc := manager.Stacks[etime.Train].Loops[etime.Epoch]

	if printTest { // print version for human checking
		PrintControlFlow = true

		fmt.Println("#### Step Cyc 1:")
		manager.Step(etime.Train, 1, etime.Cycle)
		fmt.Println("#### Step Cyc 1:")
		manager.Step(etime.Train, 1, etime.Cycle)
		fmt.Println("#### Step Cyc 1:")
		manager.Step(etime.Train, 1, etime.Cycle)
		fmt.Println("#### Step Cyc 1:")
		manager.Step(etime.Train, 1, etime.Cycle)
		fmt.Println("#### Step Cyc 2:")
		manager.Step(etime.Train, 2, etime.Cycle)

		NoPrintBelow = etime.Trial

		fmt.Println("#### Step Run 1:")
		manager.Step(etime.Train, 1, etime.Run)
		if run.Counter.Cur != 1 {
			t.Errorf("Incorrect step run")
		}
		fmt.Println("#### Step Epoch 3:")
		manager.Step(etime.Train, 3, etime.Epoch)
		if run.Counter.Cur != 1 || epc.Counter.Cur != 3 {
			t.Errorf("Incorrect step epoch")
		}
		if trialCount != 32 { // 32 = 1*5*4+3*4
			t.Errorf("Cycles not counted correctly: %d != 32", trialCount)
		}

		NoPrintBelow = etime.AllTimes

		fmt.Println("#### Step Trial 2:")
		manager.Step(etime.Train, 2, etime.Trial)
		if trialCount != 34 { // 34 = 1*5*4+3*4+2
			t.Errorf("Cycles not counted correctly: %d != 34", trialCount)
		}
	} else {
		PrintControlFlow = false
		manager.Step(etime.Train, 1, etime.Cycle)
		manager.Step(etime.Train, 1, etime.Cycle)
		manager.Step(etime.Train, 1, etime.Cycle)
		manager.Step(etime.Train, 1, etime.Cycle)
		manager.Step(etime.Train, 2, etime.Cycle)
		manager.Step(etime.Train, 1, etime.Run)
		if run.Counter.Cur != 1 {
			t.Errorf("Incorrect step run")
		}
		manager.Step(etime.Train, 3, etime.Epoch)
		if run.Counter.Cur != 1 || epc.Counter.Cur != 3 {
			t.Errorf("Incorrect step epoch")
		}
		if trialCount != 32 { // 32 = 1*5*4+3*4
			t.Errorf("Cycles not counted correctly: %d != 32", trialCount)
		}
		manager.Step(etime.Train, 2, etime.Trial)
		if trialCount != 34 { // 34 = 1*5*4+3*4+2
			t.Errorf("Cycles not counted correctly: %d != 34", trialCount)
		}
	}
}

func TestStepIncr(t *testing.T) {
	trialCount := 0

	manager := NewManager()
	manager.AddStack(etime.Train).AddTime(etime.Run, 2).AddTime(etime.Epoch, 5).AddTimeIncr(etime.Trial, 10, 3).AddTime(etime.Cycle, 3)
	manager.GetLoop(etime.Train, etime.Trial).OnStart.Add("Count Trials", func() { trialCount += 1 })
	manager.GetLoop(etime.Train, etime.Run).OnEnd.Add("Counters Test", func() {
		run := manager.Stacks[etime.Train].Loops[etime.Run].Counter.Cur
		epc := manager.Stacks[etime.Train].Loops[etime.Epoch].Counter.Cur
		if epc != 5 {
			t.Errorf("Run %d OnEnd epoch counter should be 5, not: %d", run, epc)
		}
	})

	run := manager.Stacks[etime.Train].Loops[etime.Run]
	epc := manager.Stacks[etime.Train].Loops[etime.Epoch]

	if printTest { // print version for human checking
		PrintControlFlow = true

		fmt.Println("#### Step Cyc 1:")
		manager.Step(etime.Train, 1, etime.Cycle)
		fmt.Println("#### Step Cyc 1:")
		manager.Step(etime.Train, 1, etime.Cycle)
		fmt.Println("#### Step Cyc 1:")
		manager.Step(etime.Train, 1, etime.Cycle)
		fmt.Println("#### Step Cyc 1:")
		manager.Step(etime.Train, 1, etime.Cycle)
		fmt.Println("#### Step Cyc 2:")
		manager.Step(etime.Train, 2, etime.Cycle)

		NoPrintBelow = etime.Trial

		fmt.Println("#### Step Run 1:")
		manager.Step(etime.Train, 1, etime.Run)
		if run.Counter.Cur != 1 {
			t.Errorf("Incorrect step run")
		}
		fmt.Println("#### Step Epoch 3:")
		manager.Step(etime.Train, 3, etime.Epoch)
		if run.Counter.Cur != 1 || epc.Counter.Cur != 3 {
			t.Errorf("Incorrect step epoch")
		}
		if trialCount != 32 { // 32 = 1*5*4+3*4
			t.Errorf("Cycles not counted correctly: %d != 32", trialCount)
		}

		NoPrintBelow = etime.AllTimes

		fmt.Println("#### Step Trial 2:")
		manager.Step(etime.Train, 2, etime.Trial)
		if trialCount != 34 { // 34 = 1*5*4+3*4+2
			t.Errorf("Cycles not counted correctly: %d != 34", trialCount)
		}
	} else {
		PrintControlFlow = false
		manager.Step(etime.Train, 1, etime.Cycle)
		manager.Step(etime.Train, 1, etime.Cycle)
		manager.Step(etime.Train, 1, etime.Cycle)
		manager.Step(etime.Train, 1, etime.Cycle)
		manager.Step(etime.Train, 2, etime.Cycle)
		manager.Step(etime.Train, 1, etime.Run)
		if run.Counter.Cur != 1 {
			t.Errorf("Incorrect step run")
		}
		manager.Step(etime.Train, 3, etime.Epoch)
		if run.Counter.Cur != 1 || epc.Counter.Cur != 3 {
			t.Errorf("Incorrect step epoch")
		}
		if trialCount != 32 { // 32 = 1*5*4+3*4
			t.Errorf("Cycles not counted correctly: %d != 32", trialCount)
		}
		manager.Step(etime.Train, 2, etime.Trial)
		if trialCount != 34 { // 34 = 1*5*4+3*4+2
			t.Errorf("Cycles not counted correctly: %d != 34", trialCount)
		}
	}
}
