// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/abilities"
	"cogentcore.org/core/tree"
	"github.com/emer/emergent/v2/etime"
	"github.com/emer/emergent/v2/looper"
)

// AddLooperCtrl adds toolbar control for looper.Stack
// with Run, Step controls.
func (gui *GUI) AddLooperCtrl(p *tree.Plan, loops *looper.Manager, modes []etime.Modes) {
	gui.AddToolbarItem(p, ToolbarItem{Label: "Stop",
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
		tree.AddAt(p, mode.String()+"-run", func(w *core.Button) {
			tb := p.Parent.(*core.Toolbar)
			w.SetText(mode.String() + " Run").SetIcon(icons.PlayArrow).
				SetTooltip("Run the " + mode.String() + " process")
			w.FirstStyler(func(s *styles.Style) { s.SetEnabled(!gui.IsRunning) })
			w.OnClick(func(e events.Event) {
				if !gui.IsRunning {
					gui.IsRunning = true
					tb.Restyle()
					go func() {
						loops.Run(mode)
						gui.Stopped()
					}()
				}
			})
		})

		stepN := make(map[string]int)
		steps := loops.Stacks[mode].Order
		stringToEnumTime := make(map[string]etime.Times)
		for _, st := range steps {
			stepN[st.String()] = 1
			stringToEnumTime[st.String()] = st
		}

		tree.AddAt(p, mode.String()+"-step", func(w *core.Button) {
			tb := p.Parent.(*core.Toolbar)
			w.SetText("Step").SetIcon(icons.SkipNext).
				SetTooltip("Step the " + mode.String() + " process according to the following step level and N")
			w.FirstStyler(func(s *styles.Style) {
				s.SetEnabled(!gui.IsRunning)
				s.SetAbilities(true, abilities.RepeatClickable)
			})
			w.OnClick(func(e events.Event) {
				if !gui.IsRunning {
					gui.IsRunning = true
					tb.Restyle()
					go func() {
						stack := loops.Stacks[mode]
						loops.Step(mode, stepN[stack.StepLevel.String()], stack.StepLevel)
						gui.Stopped()
					}()
				}
			})
		})

		var chs *core.Chooser
		tree.AddAt(p, mode.String()+"-level", func(w *core.Chooser) {
			chs = w
			stepStrs := []string{}
			for _, s := range steps {
				stepStrs = append(stepStrs, s.String())
			}
			w.SetStrings(stepStrs...)
			stack := loops.Stacks[mode]
			w.SetCurrentValue(stack.StepLevel.String())
		})

		tree.AddAt(p, mode.String()+"-n", func(w *core.Spinner) {
			w.SetStep(1).SetMin(1).SetValue(1)
			w.SetTooltip("number of iterations per step").
				OnChange(func(e events.Event) {
					stepN[chs.CurrentItem.Value.(string)] = int(w.Value)
				})

			w.OnChange(func(e events.Event) {
				stack := loops.Stacks[mode]
				stack.StepLevel = stringToEnumTime[chs.CurrentItem.Value.(string)]
				w.Value = float32(stepN[stack.StepLevel.String()])
			})
		})
	}
}
