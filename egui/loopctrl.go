// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

import (
	"fmt"
	"github.com/emer/emergent/etime"
	"github.com/emer/emergent/looper"
	"github.com/goki/gi/gi"
	"github.com/goki/ki/ki"
)

// AddLooperCtrl adds toolbar control for looper.Stack
// with Run, Step controls.
func (gui *GUI) AddLooperCtrl(loops *looper.Manager, modes []etime.Modes) {
	gui.AddToolbarItem(ToolbarItem{Label: "Stop",
		Icon:    "stop",
		Tooltip: "Interrupts running.  running / stepping picks back up where it left off.",
		Active:  ActiveRunning,
		Func: func() {
			loops.StopFlag = true
			loops.StopLevel = etime.Cycle
			fmt.Println("Stop time!")
			gui.StopNow = true
			gui.Stopped()
		},
	})

	for _, m := range modes {
		mode := m

		gui.ToolBar.AddAction(gi.ActOpts{Label: mode.String() + " Run", Icon: "play", Tooltip: "Run the " + mode.String() + " process", UpdateFunc: func(act *gi.Action) {
			act.SetActiveStateUpdt(!gui.IsRunning)
		}}, gui.Win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
			if !gui.IsRunning {
				gui.IsRunning = true
				gui.ToolBar.UpdateActions()
				go func() {
					loops.StopFlag = false
					loops.Mode = mode
					loops.Run()
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

		lastSelectedScbTimeScale := etime.Run

		gui.ToolBar.AddAction(gi.ActOpts{Label: "Step", Icon: "step-fwd", Tooltip: "Step the " + mode.String() + " process according to the following step level and N", UpdateFunc: func(act *gi.Action) {
			act.SetActiveStateUpdt(!gui.IsRunning)
		}}, gui.Win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
			if !gui.IsRunning {
				gui.IsRunning = true
				gui.ToolBar.UpdateActions()
				go func() {
					loops.Mode = mode
					loops.Step(stepN[loops.StopLevel.String()], lastSelectedScbTimeScale)
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
		scb.SetCurVal(loops.StopLevel.String())

		sb := gi.AddNewSpinBox(gui.ToolBar, "step-n")
		sb.Defaults()
		sb.Tooltip = "number of iterations per step"
		sb.SetProp("step", 1)
		sb.HasMin = true
		sb.Min = 1
		sb.Value = 1
		sb.SpinBoxSig.Connect(gui.ToolBar.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
			stepN[scb.CurVal.(string)] = int(data.(float32))
		})

		scb.ComboSig.Connect(gui.ToolBar.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
			lastSelectedScbTimeScale = stringToEnumTime[scb.CurVal.(string)]
			sb.Value = float32(stepN[loops.StopLevel.String()])
		})
	}
}
