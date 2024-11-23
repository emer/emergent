// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import (
	"fmt"
	"testing"

	"github.com/emer/emergent/v2/looper/levels"
)

var printTest = false

func ExampleStacks() {
	stacks := NewStacks()
	stacks.AddStack(levels.Train, levels.Trial).
		AddLevel(levels.Epoch, 3).
		AddLevel(levels.Trial, 2)

	// add function closures:
	stacks.Loop(levels.Train, levels.Epoch).OnStart.Add("Epoch Start", func() { fmt.Println("Epoch Start") })
	stacks.Loop(levels.Train, levels.Epoch).OnEnd.Add("Epoch End", func() { fmt.Println("Epoch End") })
	stacks.Loop(levels.Train, levels.Trial).OnStart.Add("Trial Run", func() { fmt.Println("  Trial Run") })

	// add events:
	stacks.Loop(levels.Train, levels.Epoch).AddEvent("EpochTwoEvent", 2, func() { fmt.Println("Epoch==2") })
	stacks.Loop(levels.Train, levels.Trial).AddEvent("TrialOneEvent", 1, func() { fmt.Println("  Trial==1") })

	// fmt.Println(stacks.DocString())

	stacks.Run(levels.Train)

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
	stacks.AddStack(levels.Train, levels.Trial).
		AddLevel(levels.Run, 2).
		AddLevel(levels.Epoch, 5).
		AddLevel(levels.Trial, 4).
		AddLevel(levels.Cycle, 3)
	stacks.Loop(levels.Train, levels.Trial).OnStart.Add("Count Trials", func() { trialCount += 1 })
	stacks.Loop(levels.Train, levels.Run).OnEnd.Add("Counters Test", func() {
		run := stacks.Stacks[levels.Train].Loops[levels.Run].Counter.Cur
		epc := stacks.Stacks[levels.Train].Loops[levels.Epoch].Counter.Cur
		if epc != 5 {
			t.Errorf("Run %d OnEnd epoch counter should be 5, not: %d", run, epc)
		}
	})

	run := stacks.Stacks[levels.Train].Loops[levels.Run]
	epc := stacks.Stacks[levels.Train].Loops[levels.Epoch]

	if printTest { // print version for human checking
		PrintControlFlow = true

		fmt.Println("#### Step Cyc 1:")
		stacks.Step(levels.Train, 1, levels.Cycle)
		fmt.Println("#### Step Cyc 1:")
		stacks.Step(levels.Train, 1, levels.Cycle)
		fmt.Println("#### Step Cyc 1:")
		stacks.Step(levels.Train, 1, levels.Cycle)
		fmt.Println("#### Step Cyc 1:")
		stacks.Step(levels.Train, 1, levels.Cycle)
		fmt.Println("#### Step Cyc 2:")
		stacks.Step(levels.Train, 2, levels.Cycle)

		fmt.Println("#### Step Run 1:")
		stacks.Step(levels.Train, 1, levels.Run)
		if run.Counter.Cur != 1 {
			t.Errorf("Incorrect step run")
		}
		fmt.Println("#### Step Epoch 3:")
		stacks.Step(levels.Train, 3, levels.Epoch)
		if run.Counter.Cur != 1 || epc.Counter.Cur != 3 {
			t.Errorf("Incorrect step epoch")
		}
		if trialCount != 32 { // 32 = 1*5*4+3*4
			t.Errorf("Cycles not counted correctly: %d != 32", trialCount)
		}

		fmt.Println("#### Step Trial 2:")
		stacks.Step(levels.Train, 2, levels.Trial)
		if trialCount != 34 { // 34 = 1*5*4+3*4+2
			t.Errorf("Cycles not counted correctly: %d != 34", trialCount)
		}
	} else {
		PrintControlFlow = false
		stop := stacks.Step(levels.Train, 1, levels.Cycle)
		if stop != levels.Cycle {
			t.Errorf("stop != Cycle: %s", stop)
		}
		stop = stacks.Step(levels.Train, 1, levels.Cycle)
		if stop != levels.Cycle {
			t.Errorf("stop != Cycle: %s", stop)
		}
		stop = stacks.Step(levels.Train, 1, levels.Cycle)
		if stop != levels.Cycle {
			t.Errorf("stop != Cycle: %s", stop)
		}
		stop = stacks.Step(levels.Train, 1, levels.Cycle)
		if stop != levels.Cycle {
			t.Errorf("stop != Cycle: %s", stop)
		}
		stop = stacks.Step(levels.Train, 2, levels.Cycle)
		if stop != levels.Cycle {
			t.Errorf("stop != Cycle: %s", stop)
		}
		stop = stacks.Step(levels.Train, 1, levels.Run)
		if run.Counter.Cur != 1 {
			t.Errorf("Incorrect step run")
		}
		if stop != levels.Run {
			t.Errorf("stop != Run: %s", stop)
		}
		stop = stacks.Step(levels.Train, 3, levels.Epoch)
		if run.Counter.Cur != 1 || epc.Counter.Cur != 3 {
			t.Errorf("Incorrect step epoch")
		}
		if stop != levels.Epoch {
			t.Errorf("stop != Epoch: %s", stop)
		}
		if trialCount != 32 { // 32 = 1*5*4+3*4
			t.Errorf("Cycles not counted correctly: %d != 32", trialCount)
		}
		stop = stacks.Step(levels.Train, 2, levels.Trial)
		if trialCount != 34 { // 34 = 1*5*4+3*4+2
			t.Errorf("Cycles not counted correctly: %d != 34", trialCount)
		}
		if stop != levels.Trial {
			t.Errorf("stop != Trial: %s", stop)
		}
	}
}

func TestStepIncr(t *testing.T) {
	trialCount := 0

	stacks := NewStacks()
	stacks.AddStack(levels.Train, levels.Trial).
		AddLevel(levels.Run, 2).
		AddLevel(levels.Epoch, 5).
		AddLevelIncr(levels.Trial, 10, 3).
		AddLevel(levels.Cycle, 3)
	stacks.Loop(levels.Train, levels.Trial).OnStart.Add("Count Trials", func() { trialCount += 1 })
	stacks.Loop(levels.Train, levels.Run).OnEnd.Add("Counters Test", func() {
		run := stacks.Stacks[levels.Train].Loops[levels.Run].Counter.Cur
		epc := stacks.Stacks[levels.Train].Loops[levels.Epoch].Counter.Cur
		if epc != 5 {
			t.Errorf("Run %d OnEnd epoch counter should be 5, not: %d", run, epc)
		}
	})

	run := stacks.Stacks[levels.Train].Loops[levels.Run]
	epc := stacks.Stacks[levels.Train].Loops[levels.Epoch]

	if printTest { // print version for human checking
		PrintControlFlow = true

		fmt.Println("#### Step Cyc 1:")
		stacks.Step(levels.Train, 1, levels.Cycle)
		fmt.Println("#### Step Cyc 1:")
		stacks.Step(levels.Train, 1, levels.Cycle)
		fmt.Println("#### Step Cyc 1:")
		stacks.Step(levels.Train, 1, levels.Cycle)
		fmt.Println("#### Step Cyc 1:")
		stacks.Step(levels.Train, 1, levels.Cycle)
		fmt.Println("#### Step Cyc 2:")
		stacks.Step(levels.Train, 2, levels.Cycle)

		fmt.Println("#### Step Run 1:")
		stacks.Step(levels.Train, 1, levels.Run)
		if run.Counter.Cur != 1 {
			t.Errorf("Incorrect step run")
		}
		fmt.Println("#### Step Epoch 3:")
		stacks.Step(levels.Train, 3, levels.Epoch)
		if run.Counter.Cur != 1 || epc.Counter.Cur != 3 {
			t.Errorf("Incorrect step epoch")
		}
		if trialCount != 32 { // 32 = 1*5*4+3*4
			t.Errorf("Cycles not counted correctly: %d != 32", trialCount)
		}

		fmt.Println("#### Step Trial 2:")
		stacks.Step(levels.Train, 2, levels.Trial)
		if trialCount != 34 { // 34 = 1*5*4+3*4+2
			t.Errorf("Cycles not counted correctly: %d != 34", trialCount)
		}
	} else {
		PrintControlFlow = false
		stacks.Step(levels.Train, 1, levels.Cycle)
		stacks.Step(levels.Train, 1, levels.Cycle)
		stacks.Step(levels.Train, 1, levels.Cycle)
		stacks.Step(levels.Train, 1, levels.Cycle)
		stacks.Step(levels.Train, 2, levels.Cycle)
		stacks.Step(levels.Train, 1, levels.Run)
		if run.Counter.Cur != 1 {
			t.Errorf("Incorrect step run")
		}
		stacks.Step(levels.Train, 3, levels.Epoch)
		if run.Counter.Cur != 1 || epc.Counter.Cur != 3 {
			t.Errorf("Incorrect step epoch")
		}
		if trialCount != 32 { // 32 = 1*5*4+3*4
			t.Errorf("Cycles not counted correctly: %d != 32", trialCount)
		}
		stacks.Step(levels.Train, 2, levels.Trial)
		if trialCount != 34 { // 34 = 1*5*4+3*4+2
			t.Errorf("Cycles not counted correctly: %d != 34", trialCount)
		}
	}
}
