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
func (gui *GUI) AddLooperCtrl(evalLoops *looper.EvaluationModeLoops, stepper *looper.Stepper) {
	//Todo added stopper here
	gui.AddToolbarItem(ToolbarItem{Label: "Stop",
		Icon:    "stop",
		Tooltip: "Interrupts running.  running / stepping picks back up where it left off.",
		Active:  ActiveRunning,
		Func: func() {
			stepper.StopFlag = true
			//stepper.StopLevel = etime.Cycle
			fmt.Println("Stop time!")
			gui.StopNow = true
			gui.Stopped()
		},
	})

	gui.ToolBar.AddAction(gi.ActOpts{Label: stepper.Mode.String() + " Run", Icon: "play", Tooltip: "Run the " + stepper.Mode.String() + " process", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!gui.IsRunning)
	}}, gui.Win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		if !gui.IsRunning {
			gui.IsRunning = true
			gui.ToolBar.UpdateActions()
			go func() {
				//evalLoops.StepClear() // DO NOT SUBMIT Is this necessary? Also check comments below.
				stepper.StopFlag = false
				stepper.Run()
				gui.Stopped()
			}()
		}
	})

	//stepLevel := evalLoops.Step.Default
	stepN := make(map[string]int)
	steps := evalLoops.Order
	stringToEnumTime := make(map[string]etime.Times)
	for _, st := range steps {
		stepN[st.String()] = 1
		stringToEnumTime[st.String()] = st
	}

	gui.ToolBar.AddAction(gi.ActOpts{Label: "Step", Icon: "step-fwd", Tooltip: "Step the " + stepper.Mode.String() + " process according to the following step level and N", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!gui.IsRunning)
	}}, gui.Win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		if !gui.IsRunning {
			gui.IsRunning = true
			gui.ToolBar.UpdateActions()
			go func() {
				stepper.StopFlag = false
				stepper.Run()
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
	scb.SetCurVal(stepper.StopLevel.String())

	sb := gi.AddNewSpinBox(gui.ToolBar, "step-n")
	sb.Defaults()
	sb.Tooltip = "number of iterations per step"
	sb.SetProp("step", 1)
	sb.HasMin = true
	sb.Min = 1
	sb.Value = 1
	sb.SpinBoxSig.Connect(gui.ToolBar.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		stepN[stepper.StopLevel.String()] = int(data.(float32))
	})

	scb.ComboSig.Connect(gui.ToolBar.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		stepper.StopLevel = stringToEnumTime[data.(string)]
		sb.Value = float32(stepN[stepper.StopLevel.String()])
	})
}
