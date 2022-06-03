// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import (
	"fmt"
	"testing"

	"github.com/emer/emergent/etime"
)

func TestStep(t *testing.T) {
	trialCount := 0

	manager := NewManager()
	manager.AddStack(etime.Train).AddTime(etime.Run, 2).AddTime(etime.Epoch, 5).AddTime(etime.Trial, 4).AddTime(etime.Cycle, 3)
	manager.GetLoop(etime.Train, etime.Trial).OnStart.Add("Count Trials", func() { trialCount += 1 })

	run := manager.Stacks[etime.Train].Loops[etime.Run]
	epc := manager.Stacks[etime.Train].Loops[etime.Epoch]

	if false { // print version for human checking
		PrintControlFlow = true

		fmt.Println("#### Step Cyc 1:")
		manager.Step(1, etime.Cycle)
		fmt.Println("#### Step Cyc 1:")
		manager.Step(1, etime.Cycle)
		fmt.Println("#### Step Cyc 1:")
		manager.Step(1, etime.Cycle)
		fmt.Println("#### Step Cyc 1:")
		manager.Step(1, etime.Cycle)
		fmt.Println("#### Step Cyc 2:")
		manager.Step(2, etime.Cycle)

		NoPrintBelow = etime.Trial

		fmt.Println("#### Step Run 1:")
		manager.Step(1, etime.Run)
		if run.Counter.Cur != 1 {
			t.Errorf("Incorrect step run")
		}
		fmt.Println("#### Step Epoch 3:")
		manager.Step(3, etime.Epoch)
		if run.Counter.Cur != 1 || epc.Counter.Cur != 3 {
			t.Errorf("Incorrect step epoch")
		}
		if trialCount != 32 { // 32 = 1*5*4+3*4
			t.Errorf("Cycles not counted correctly")
		}

		NoPrintBelow = etime.AllTimes

		fmt.Println("#### Step Trial 2:")
		manager.Step(2, etime.Trial)
		if trialCount != 34 { // 34 = 1*5*4+3*4+2
			t.Errorf("Cycles not counted correctly")
		}
	} else {
		manager.Step(1, etime.Cycle)
		manager.Step(1, etime.Cycle)
		manager.Step(1, etime.Cycle)
		manager.Step(1, etime.Cycle)
		manager.Step(2, etime.Cycle)
		manager.Step(1, etime.Run)
		if run.Counter.Cur != 1 {
			t.Errorf("Incorrect step run")
		}
		manager.Step(3, etime.Epoch)
		if run.Counter.Cur != 1 || epc.Counter.Cur != 3 {
			t.Errorf("Incorrect step epoch")
		}
		if trialCount != 32 { // 32 = 1*5*4+3*4
			t.Errorf("Cycles not counted correctly")
		}
		manager.Step(2, etime.Trial)
		if trialCount != 34 { // 34 = 1*5*4+3*4+2
			t.Errorf("Cycles not counted correctly")
		}
	}
}
