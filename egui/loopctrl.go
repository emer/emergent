// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

import (
	"github.com/emer/emergent/v2/etime"
	"github.com/emer/emergent/v2/looper"
	"goki.dev/gi/v2/gi"
	"goki.dev/ki/v2"
)

// AddLooperCtrl adds toolbar control for looper.Stack
// with Run, Step controls.
func (gui *GUI) AddLooperCtrl(loops *looper.Manager, modes []etime.Modes) {
	gui.AddToolbarItem(ToolbarItem{Label: "Stop",
		Icon:    "stop",
		Tooltip: "Interrupts running.  running / stepping picks back up where it left off.",
		Active:  ActiveRunning,
		Func: func() {
			loops.Stop(etime.Cycle)
			// fmt.Println("Stop time!")
			gui.StopNow = true
			gui.Stopped()
		},
	})

	for _, m := range modes {
		mode := m

		gui.ToolBar.AddAction(gi.ActOpts{Label: mode.String() + " Run", Icon: "play", Tooltip: "Run the " + mode.String() + " process", UpdateFunc: func(act *gi.Action) {
			act.SetActiveStateUpdt(!gui.IsRunning)
		}}, gui.Win.This(), func(recv, send ki.Ki, sig int64, data any) {
			if !gui.IsRunning {
				gui.IsRunning = true
				gui.ToolBar.UpdateActions()
				go func() {
					loops.Run(mode)
					gui.Stopped()
				}()
			}
		})

		//stepLevel := evalLoops.Step.Default
		stepN := make(map[string]int)
		steps := loops.Stacks[mode].Order
		stringToEnumTime := make(map[string]etime.Times)
		for _, st := range steps {
			stepN[st.String()] = 1
			stringToEnumTime[st.String()] = st
		}

		gui.ToolBar.AddAction(gi.ActOpts{Label: "Step", Icon: "step-fwd", Tooltip: "Step the " + mode.String() + " process according to the following step level and N", UpdateFunc: func(act *gi.Action) {
			act.SetActiveStateUpdt(!gui.IsRunning)
		}}, gui.Win.This(), func(recv, send ki.Ki, sig int64, data any) {
			if !gui.IsRunning {
				gui.IsRunning = true
				gui.ToolBar.UpdateActions()
				go func() {
					stack := loops.Stacks[mode]
					loops.Step(mode, stepN[stack.StepLevel.String()], stack.StepLevel)
					gui.Stopped()
				}()
			}
		})

		scb := gi.AddNewComboBox(gui.ToolBar, "step")
		stepStrs := []string{}
		for _, s := range steps {
			stepStrs = append(stepStrs, s.String())
		}
		scb.ItemsFromStringList(stepStrs, false, 30)
		stack := loops.Stacks[mode]
		scb.SetCurVal(stack.StepLevel.String())

		sb := gi.AddNewSpinBox(gui.ToolBar, "step-n")
		sb.Defaults()
		sb.Tooltip = "number of iterations per step"
		sb.SetProp("step", 1)
		sb.HasMin = true
		sb.Min = 1
		sb.Value = 1
		sb.SpinBoxSig.Connect(gui.ToolBar.This(), func(recv, send ki.Ki, sig int64, data any) {
			stepN[scb.CurVal.(string)] = int(data.(float32))
		})

		scb.ComboSig.Connect(gui.ToolBar.This(), func(recv, send ki.Ki, sig int64, data any) {
			stack := loops.Stacks[mode]
			stack.StepLevel = stringToEnumTime[scb.CurVal.(string)]
			sb.Value = float32(stepN[stack.StepLevel.String()])
		})
	}
}
