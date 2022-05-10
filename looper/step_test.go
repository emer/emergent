package looper

import (
	"fmt"
	"testing"

	"github.com/emer/emergent/etime"
)

func Step(stepper *Manager, stopLevel etime.Times, iters int) {
	stepper.StopLevel = stopLevel
	stepper.StepIterations = iters
	stepper.StopFlag = false
	stepper.StopNext = true
	stepper.Run()
}

func TestStep(t *testing.T) {
	trialCount := 0

	manager := Manager{}.Init()
	manager.Stacks[etime.Train] = &Stack{}
	manager.Stacks[etime.Train].Init().AddTime(etime.Run, 2).AddTime(etime.Epoch, 5).AddTime(etime.Trial, 4).AddTime(etime.Cycle, 3)
	manager.GetLoop(etime.Train, etime.Trial).OnStart.Add("Count Trials", func() { trialCount += 1 })

	manager.Init()

	run := manager.Stacks[etime.Train].Loops[etime.Run]
	epc := manager.Stacks[etime.Train].Loops[etime.Epoch]

	PrintControlFlow = true

	fmt.Println("#### Step Cyc 1:")
	Step(manager, etime.Cycle, 1)
	fmt.Println("#### Step Cyc 1:")
	Step(manager, etime.Cycle, 1)
	fmt.Println("#### Step Cyc 1:")
	Step(manager, etime.Cycle, 1)
	fmt.Println("#### Step Cyc 1:")
	Step(manager, etime.Cycle, 1)
	fmt.Println("#### Step Cyc 2:")
	Step(manager, etime.Cycle, 2)

	NoPrintBelow = etime.Trial

	fmt.Println("#### Step Run 1:")
	Step(manager, etime.Run, 1)
	if run.Counter.Cur != 1 {
		t.Errorf("Incorrect step run")
	}
	fmt.Println("#### Step Epoch 3:")
	Step(manager, etime.Epoch, 3)
	if run.Counter.Cur != 1 || epc.Counter.Cur != 3 {
		t.Errorf("Incorrect step epoch")
	}
	if trialCount != 32 { // 32 = 1*5*4+3*4
		t.Errorf("Cycles not counted correctly")
	}

	NoPrintBelow = etime.AllTimes

	fmt.Println("#### Step Trial 2:")
	Step(manager, etime.Trial, 2)
	if trialCount != 34 { // 34 = 1*5*4+3*4+2
		t.Errorf("Cycles not counted correctly")
	}
}
