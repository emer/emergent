// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

import (
	"github.com/emer/emergent/looper"
	"github.com/goki/gi/gi"
	"github.com/goki/ki/ki"
)

// AddLooperCtrl adds toolbar control for looper.Stack
// with Run, Step controls.
func (gui *GUI) AddLooperCtrl(st *looper.Stack) {
	gui.ToolBar.AddAction(gi.ActOpts{Label: st.Mode + " Run", Icon: "play", Tooltip: "Run the " + st.Mode + " process", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!gui.IsRunning)
	}}, gui.Win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		if !gui.IsRunning {
			gui.IsRunning = true
			gui.ToolBar.UpdateActions()
			go func() {
				st.StepClear()
				st.Run()
				gui.Stopped()
			}()
		}
	})

	stepLevel := st.Step.Default
	stepN := make(map[string]int)
	steps := st.Times()
	for _, st := range steps {
		stepN[st] = 1
	}

	gui.ToolBar.AddAction(gi.ActOpts{Label: "Step", Icon: "step-fwd", Tooltip: "Step the " + st.Mode + " process according to the following step level and N", UpdateFunc: func(act *gi.Action) {
		act.SetActiveStateUpdt(!gui.IsRunning)
	}}, gui.Win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		if !gui.IsRunning {
			gui.IsRunning = true
			gui.ToolBar.UpdateActions()
			go func() {
				st.SetStepTime(stepLevel, stepN[stepLevel])
				st.Run()
				gui.Stopped()
			}()
		}
	})

	scb := gi.AddNewComboBox(gui.ToolBar, "step")
	scb.ItemsFromStringList(steps, false, 30)
	scb.SetCurVal(stepLevel)

	sb := gi.AddNewSpinBox(gui.ToolBar, "step-n")
	sb.Defaults()
	sb.Tooltip = "number of iterations per step"
	sb.SetProp("step", 1)
	sb.HasMin = true
	sb.Min = 1
	sb.Value = 1
	sb.SpinBoxSig.Connect(gui.ToolBar.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		stepN[stepLevel] = int(data.(float32))
	})

	scb.ComboSig.Connect(gui.ToolBar.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		stepLevel = data.(string)
		sb.Value = float32(stepN[stepLevel])
	})
}
