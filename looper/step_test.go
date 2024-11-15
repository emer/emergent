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

func ExampleStacks() {
	stacks := NewStacks()
	stacks.AddStack(etime.Train, etime.Trial).
		AddTime(etime.Epoch, 3).
		AddTime(etime.Trial, 2)

	// add function closures:
	stacks.Loop(etime.Train, etime.Epoch).OnStart.Add("Epoch Start", func() { fmt.Println("Epoch Start") })
	stacks.Loop(etime.Train, etime.Epoch).OnEnd.Add("Epoch End", func() { fmt.Println("Epoch End") })
	stacks.Loop(etime.Train, etime.Trial).OnStart.Add("Trial Run", func() { fmt.Println("  Trial Run") })

	// add events:
	stacks.Loop(etime.Train, etime.Epoch).AddEvent("EpochTwoEvent", 2, func() { fmt.Println("Epoch==2") })
	stacks.Loop(etime.Train, etime.Trial).AddEvent("TrialOneEvent", 1, func() { fmt.Println("  Trial==1") })

	// fmt.Println(stacks.DocString())

	stacks.Run(etime.Train)

	// Output:
	// Epoch Start
	//   Trial Run
	//   Trial==1
	//   Trial Run
	// Epoch End
	// Epoch Start
	//   Trial Run
	//   Trial==1
	//   Trial Run
	// Epoch End
	// Epoch==2
	// Epoch Start
	//   Trial Run
	//   Trial==1
	//   Trial Run
	// Epoch End
}

func TestStep(t *testing.T) {
	trialCount := 0

	stacks := NewStacks()
	stacks.AddStack(etime.Train, etime.Trial).
		AddTime(etime.Run, 2).
		AddTime(etime.Epoch, 5).
		AddTime(etime.Trial, 4).
		AddTime(etime.Cycle, 3)
	stacks.Loop(etime.Train, etime.Trial).OnStart.Add("Count Trials", func() { trialCount += 1 })
	stacks.Loop(etime.Train, etime.Run).OnEnd.Add("Counters Test", func() {
		run := stacks.Stacks[etime.Train].Loops[etime.Run].Counter.Cur
		epc := stacks.Stacks[etime.Train].Loops[etime.Epoch].Counter.Cur
		if epc != 5 {
			t.Errorf("Run %d OnEnd epoch counter should be 5, not: %d", run, epc)
		}
	})

	run := stacks.Stacks[etime.Train].Loops[etime.Run]
	epc := stacks.Stacks[etime.Train].Loops[etime.Epoch]

	if printTest { // print version for human checking
		PrintControlFlow = true

		fmt.Println("#### Step Cyc 1:")
		stacks.Step(etime.Train, 1, etime.Cycle)
		fmt.Println("#### Step Cyc 1:")
		stacks.Step(etime.Train, 1, etime.Cycle)
		fmt.Println("#### Step Cyc 1:")
		stacks.Step(etime.Train, 1, etime.Cycle)
		fmt.Println("#### Step Cyc 1:")
		stacks.Step(etime.Train, 1, etime.Cycle)
		fmt.Println("#### Step Cyc 2:")
		stacks.Step(etime.Train, 2, etime.Cycle)

		fmt.Println("#### Step Run 1:")
		stacks.Step(etime.Train, 1, etime.Run)
		if run.Counter.Cur != 1 {
			t.Errorf("Incorrect step run")
		}
		fmt.Println("#### Step Epoch 3:")
		stacks.Step(etime.Train, 3, etime.Epoch)
		if run.Counter.Cur != 1 || epc.Counter.Cur != 3 {
			t.Errorf("Incorrect step epoch")
		}
		if trialCount != 32 { // 32 = 1*5*4+3*4
			t.Errorf("Cycles not counted correctly: %d != 32", trialCount)
		}

		fmt.Println("#### Step Trial 2:")
		stacks.Step(etime.Train, 2, etime.Trial)
		if trialCount != 34 { // 34 = 1*5*4+3*4+2
			t.Errorf("Cycles not counted correctly: %d != 34", trialCount)
		}
	} else {
		PrintControlFlow = false
		stacks.Step(etime.Train, 1, etime.Cycle)
		stacks.Step(etime.Train, 1, etime.Cycle)
		stacks.Step(etime.Train, 1, etime.Cycle)
		stacks.Step(etime.Train, 1, etime.Cycle)
		stacks.Step(etime.Train, 2, etime.Cycle)
		stacks.Step(etime.Train, 1, etime.Run)
		if run.Counter.Cur != 1 {
			t.Errorf("Incorrect step run")
		}
		stacks.Step(etime.Train, 3, etime.Epoch)
		if run.Counter.Cur != 1 || epc.Counter.Cur != 3 {
			t.Errorf("Incorrect step epoch")
		}
		if trialCount != 32 { // 32 = 1*5*4+3*4
			t.Errorf("Cycles not counted correctly: %d != 32", trialCount)
		}
		stacks.Step(etime.Train, 2, etime.Trial)
		if trialCount != 34 { // 34 = 1*5*4+3*4+2
			t.Errorf("Cycles not counted correctly: %d != 34", trialCount)
		}
	}
}

func TestStepIncr(t *testing.T) {
	trialCount := 0

	stacks := NewStacks()
	stacks.AddStack(etime.Train, etime.Trial).
		AddTime(etime.Run, 2).
		AddTime(etime.Epoch, 5).
		AddTimeIncr(etime.Trial, 10, 3).
		AddTime(etime.Cycle, 3)
	stacks.Loop(etime.Train, etime.Trial).OnStart.Add("Count Trials", func() { trialCount += 1 })
	stacks.Loop(etime.Train, etime.Run).OnEnd.Add("Counters Test", func() {
		run := stacks.Stacks[etime.Train].Loops[etime.Run].Counter.Cur
		epc := stacks.Stacks[etime.Train].Loops[etime.Epoch].Counter.Cur
		if epc != 5 {
			t.Errorf("Run %d OnEnd epoch counter should be 5, not: %d", run, epc)
		}
	})

	run := stacks.Stacks[etime.Train].Loops[etime.Run]
	epc := stacks.Stacks[etime.Train].Loops[etime.Epoch]

	if printTest { // print version for human checking
		PrintControlFlow = true

		fmt.Println("#### Step Cyc 1:")
		stacks.Step(etime.Train, 1, etime.Cycle)
		fmt.Println("#### Step Cyc 1:")
		stacks.Step(etime.Train, 1, etime.Cycle)
		fmt.Println("#### Step Cyc 1:")
		stacks.Step(etime.Train, 1, etime.Cycle)
		fmt.Println("#### Step Cyc 1:")
		stacks.Step(etime.Train, 1, etime.Cycle)
		fmt.Println("#### Step Cyc 2:")
		stacks.Step(etime.Train, 2, etime.Cycle)

		fmt.Println("#### Step Run 1:")
		stacks.Step(etime.Train, 1, etime.Run)
		if run.Counter.Cur != 1 {
			t.Errorf("Incorrect step run")
		}
		fmt.Println("#### Step Epoch 3:")
		stacks.Step(etime.Train, 3, etime.Epoch)
		if run.Counter.Cur != 1 || epc.Counter.Cur != 3 {
			t.Errorf("Incorrect step epoch")
		}
		if trialCount != 32 { // 32 = 1*5*4+3*4
			t.Errorf("Cycles not counted correctly: %d != 32", trialCount)
		}

		fmt.Println("#### Step Trial 2:")
		stacks.Step(etime.Train, 2, etime.Trial)
		if trialCount != 34 { // 34 = 1*5*4+3*4+2
			t.Errorf("Cycles not counted correctly: %d != 34", trialCount)
		}
	} else {
		PrintControlFlow = false
		stacks.Step(etime.Train, 1, etime.Cycle)
		stacks.Step(etime.Train, 1, etime.Cycle)
		stacks.Step(etime.Train, 1, etime.Cycle)
		stacks.Step(etime.Train, 1, etime.Cycle)
		stacks.Step(etime.Train, 2, etime.Cycle)
		stacks.Step(etime.Train, 1, etime.Run)
		if run.Counter.Cur != 1 {
			t.Errorf("Incorrect step run")
		}
		stacks.Step(etime.Train, 3, etime.Epoch)
		if run.Counter.Cur != 1 || epc.Counter.Cur != 3 {
			t.Errorf("Incorrect step epoch")
		}
		if trialCount != 32 { // 32 = 1*5*4+3*4
			t.Errorf("Cycles not counted correctly: %d != 32", trialCount)
		}
		stacks.Step(etime.Train, 2, etime.Trial)
		if trialCount != 34 { // 34 = 1*5*4+3*4+2
			t.Errorf("Cycles not counted correctly: %d != 34", trialCount)
		}
	}
}
