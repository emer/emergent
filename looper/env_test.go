// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import (
	"fmt"
	"testing"

	"github.com/emer/emergent/envlp"
	"github.com/emer/emergent/estats"
	"github.com/emer/emergent/etime"
	"github.com/emer/etable/etensor"
)

type TestEnv struct {
	Nm    string     `desc:"name of this environment"`
	Dsc   string     `desc:"description of this environment"`
	EMode string     `desc:"eval mode for this env"`
	Ctrs  envlp.Ctrs `desc:"counters for this environment"`
}

func (ev *TestEnv) Name() string          { return ev.Nm }
func (ev *TestEnv) Desc() string          { return ev.Dsc }
func (ev *TestEnv) Mode() string          { return ev.EMode }
func (ev *TestEnv) Counters() *envlp.Ctrs { return &ev.Ctrs }
func (ev *TestEnv) Counter(time etime.Times) *envlp.Ctr {
	return ev.Ctrs.ByScope(etime.ScopeStr(ev.EMode, time.String()))
}
func (ev *TestEnv) String() string { return "" }
func (ev *TestEnv) CtrsToStats(stats *estats.Stats) {
	ev.Ctrs.CtrsToStats(stats)
}

func (ev *TestEnv) Config(mode string) {
	ev.EMode = mode
	ev.Ctrs.SetTimes(mode, etime.Run, etime.Epoch, etime.Trial)
}

func (ev *TestEnv) Validate() error {
	return nil
}

func (ev *TestEnv) Init() {
	run := 0
	rc := ev.Counter(etime.Run)
	if rc != nil {
		run = rc.Cur
	}
	ev.Ctrs.Init()
	if rc != nil {
		rc.Set(run)
	}
}

func (ev *TestEnv) Step() {
	tc := ev.Counter(etime.Trial)
	tc.Incr()
}

func (ev *TestEnv) State(element string) etensor.Tensor {
	return nil
}

func (ev *TestEnv) Action(element string, input etensor.Tensor) {
	// nop
}

// Compile-time check that implements Env interface
var _ envlp.Env = (*TestEnv)(nil)

func TestEnvStack(t *testing.T) {
	ev := &TestEnv{}
	ev.Config(etime.Train.String())

	ev.Counter(etime.Run).Max = 2
	ev.Counter(etime.Epoch).Max = 3
	ev.Counter(etime.Trial).Max = 3

	ev.Init()

	set := NewSet()
	trn := NewStackEnv(ev)
	set.AddStack(trn)
	trn.Step.LoopTrace = true
	// trn.Step.FuncTrace = true

	fmt.Println(trn.DocString())
	fmt.Println("##########################")

	set.Run(etime.Train, etime.Run)

	// stepping
	fmt.Printf("\n##############\nStep Trial 1\n")
	ev.Init()
	ev.Counter(etime.Run).Init()
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
	ev.Init()
	ev.Counter(etime.Run).Init()
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
	ev.Init()
	ev.Counter(etime.Run).Init()
	set.Step(etime.Train, etime.Run, etime.Epoch, 1)
	// stepping
	fmt.Printf("\n##############\nStep Epoch 1\n")
	set.Run(etime.Train, etime.Run)
	fmt.Printf("\n##############\nStep Epoch 1\n")
	set.Run(etime.Train, etime.Run)
	fmt.Printf("\n##############\nStep Epoch 1\n")
	set.Run(etime.Train, etime.Run)
}
