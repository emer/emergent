package looper

import (
	"github.com/emer/emergent/etime"
	"testing"
)

func Step(stepper *Stepper, stopLevel etime.Times, iters int) {
	stepper.StopLevel = stopLevel
	stepper.StepIterations = iters
	stepper.StopFlag = false
	stepper.StopNext = true
	stepper.Run()
}

func TestStep(t *testing.T) {
	trialCount := 0

	manager := DataManager{}.Init()
	manager.Stacks[etime.Train] = &Stack{}
	manager.Stacks[etime.Train].Init().AddTime(etime.Run, 2).AddTime(etime.Epoch, 5).AddTime(etime.Trial, 4).AddTime(etime.Cycle, 200)
	manager.GetLoop(etime.Train, etime.Trial).OnStart.Add("Count Trials", func() { trialCount += 1 })

	manager.Steps.Init(manager)

	run := manager.Stacks[etime.Train].Loops[etime.Run]
	epc := manager.Stacks[etime.Train].Loops[etime.Epoch]

	Step(&manager.Steps, etime.Run, 1)
	if run.Counter.Cur != 1 {
		t.Errorf("Incorrect step run")
	}
	Step(&manager.Steps, etime.Epoch, 3)
	if run.Counter.Cur != 1 || epc.Counter.Cur != 3 {
		t.Errorf("Incorrect step epoch")
	}
	if trialCount != 32 { // 32 = 1*5*4+3*4
		t.Errorf("Cycles not counted correctly")
	}
	Step(&manager.Steps, etime.Trial, 2)
	if trialCount != 34 { // 34 = 1*5*4+3*4+2
		t.Errorf("Cycles not counted correctly")
	}
}
