// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

import (
	"cogentcore.org/core/abilities"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
	"github.com/emer/emergent/v2/etime"
	"github.com/emer/emergent/v2/looper"
)

// AddLooperCtrl adds toolbar control for looper.Stack
// with Run, Step controls.
func (gui *GUI) AddLooperCtrl(tb *core.Toolbar, loops *looper.Manager, modes []etime.Modes) {
	gui.AddToolbarItem(tb, ToolbarItem{Label: "Stop",
		Icon:    icons.Stop,
		Tooltip: "Interrupts running.  running / stepping picks back up where it left off.",
		Active:  ActiveRunning,
		Func: func() {
			loops.Stop(etime.Cycle)
			// fmt.Println("Stop time!")
			gui.StopNow = true
		},
	})

	for _, m := range modes {
		mode := m
		core.NewButton(tb).SetText(mode.String() + " Run").SetIcon(icons.PlayArrow).
			SetTooltip("Run the " + mode.String() + " process").
			StyleFirst(func(s *styles.Style) { s.SetEnabled(!gui.IsRunning) }).
			OnClick(func(e events.Event) {
				if !gui.IsRunning {
					gui.IsRunning = true
					tb.UpdateBar()
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
		core.NewButton(tb).SetText("Step").SetIcon(icons.SkipNext).
			SetTooltip("Step the " + mode.String() + " process according to the following step level and N").
			StyleFirst(func(s *styles.Style) {
				s.SetEnabled(!gui.IsRunning)
				s.SetAbilities(true, abilities.RepeatClickable)
			}).
			OnClick(func(e events.Event) {
				if !gui.IsRunning {
					gui.IsRunning = true
					tb.UpdateBar()
					go func() {
						stack := loops.Stacks[mode]
						loops.Step(mode, stepN[stack.StepLevel.String()], stack.StepLevel)
						gui.Stopped()
					}()
				}
			})

		scb := core.NewChooser(tb, "step")
		stepStrs := []string{}
		for _, s := range steps {
			stepStrs = append(stepStrs, s.String())
		}
		scb.SetStrings(stepStrs...)
		stack := loops.Stacks[mode]
		scb.SetCurrentValue(stack.StepLevel.String())

		sb := core.NewSpinner(tb, "step-n").SetTooltip("number of iterations per step").
			SetStep(1).SetMin(1).SetValue(1)
		sb.OnChange(func(e events.Event) {
			stepN[scb.CurrentItem.Value.(string)] = int(sb.Value)
		})

		scb.OnChange(func(e events.Event) {
			stack := loops.Stacks[mode]
			stack.StepLevel = stringToEnumTime[scb.CurrentItem.Value.(string)]
			sb.Value = float32(stepN[stack.StepLevel.String()])
		})
	}
}
