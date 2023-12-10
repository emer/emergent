// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

import (
	"github.com/emer/emergent/v2/etime"
	"github.com/emer/emergent/v2/looper"
	"goki.dev/gi/v2/gi"
	"goki.dev/girl/states"
	"goki.dev/girl/styles"
	"goki.dev/goosi/events"
	"goki.dev/icons"
)

// AddLooperCtrl adds toolbar control for looper.Stack
// with Run, Step controls.
func (gui *GUI) AddLooperCtrl(tb *gi.Toolbar, loops *looper.Manager, modes []etime.Modes) {
	gui.AddToolbarItem(tb, ToolbarItem{Label: "Stop",
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
		gi.NewButton(tb).SetText(mode.String() + " Run").SetIcon(icons.PlayArrow).
			SetTooltip("Run the " + mode.String() + " process").
			Style(func(s *styles.Style) {
				s.State.SetFlag(gui.IsRunning, states.Disabled)
			}).
			OnClick(func(e events.Event) {
				if !gui.IsRunning {
					gui.IsRunning = true
					tb.ApplyStyleTree()
					tb.SetNeedsRender(true)
					go func() {
						loops.Run(mode)
						gui.StepDone()
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
		gi.NewButton(tb).SetText("Step").SetIcon(icons.SkipNext).
			SetTooltip("Step the " + mode.String() + " process according to the following step level and N").
			Style(func(s *styles.Style) {
				s.State.SetFlag(gui.IsRunning, states.Disabled)
			}).
			OnClick(func(e events.Event) {
				if !gui.IsRunning {
					gui.IsRunning = true
					tb.Update()
					go func() {
						stack := loops.Stacks[mode]
						loops.Step(mode, stepN[stack.StepLevel.String()], stack.StepLevel)
						gui.Stopped()
					}()
				}
			})

		scb := gi.NewChooser(tb, "step")
		stepStrs := []string{}
		for _, s := range steps {
			stepStrs = append(stepStrs, s.String())
		}
		scb.SetStrings(stepStrs, false, 30)
		stack := loops.Stacks[mode]
		scb.SetCurVal(stack.StepLevel.String())

		sb := gi.NewSpinner(tb, "step-n").SetTooltip("number of iterations per step").
			SetStep(1).SetMin(1).SetValue(1)
		sb.OnChange(func(e events.Event) {
			stepN[scb.CurVal.(string)] = int(sb.Value)
		})

		scb.OnChange(func(e events.Event) {
			stack := loops.Stacks[mode]
			stack.StepLevel = stringToEnumTime[scb.CurVal.(string)]
			sb.Value = float32(stepN[stack.StepLevel.String()])
		})
	}
}
