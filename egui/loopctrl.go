// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

import (
	"cmp"
	"slices"
	"strings"

	"cogentcore.org/core/core"
	"cogentcore.org/core/enums"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/abilities"
	"cogentcore.org/core/tree"
	"github.com/emer/emergent/v2/etime"
	"github.com/emer/emergent/v2/looper"
)

// AddLooperCtrl adds toolbar control for looper.Stacks with Init, Run, Step controls,
// with selector for which stack is being controlled.
// A prefix can optionally be provided if multiple loops are used.
func (gui *GUI) AddLooperCtrl(p *tree.Plan, loops *looper.Stacks, prefix ...string) {
	pfx := ""
	lblpfx := ""
	if len(prefix) == 1 {
		pfx = strings.ToLower(prefix[0]) + "-"
		lblpfx = prefix[0] + " "
	}
	modes := make([]enums.Enum, len(loops.Stacks))
	var stepChoose *core.Chooser
	var stepNSpin *core.Spinner
	i := 0
	for m := range loops.Stacks {
		modes[i] = m
		i++
	}
	slices.SortFunc(modes, func(a, b enums.Enum) int {
		return cmp.Compare(a.Int64(), b.Int64())
	})
	curMode := modes[0]
	curStep := loops.Stacks[curMode].StepLevel

	updateSteps := func() {
		st := loops.Stacks[curMode]
		stepStrs := make([]string, len(st.Order))
		cur := ""
		for i, s := range st.Order {
			sv := s.String()
			stepStrs[i] = sv
			if s.Int64() == curStep.Int64() {
				cur = sv
			}
		}
		stepChoose.SetStrings(stepStrs...)
		stepChoose.SetCurrentValue(cur)
	}

	if len(modes) > 1 {
		tree.AddAt(p, pfx+"loop-mode", func(w *core.Switches) {
			w.SetType(core.SwitchSegmentedButton)
			w.Mutex = true
			w.SetEnums(modes...)
			w.SelectValue(curMode)
			w.FinalStyler(func(s *styles.Style) {
				s.Grow.Set(0, 0)
			})
			w.OnChange(func(e events.Event) {
				sel := w.SelectedItem()
				if sel == nil || sel.Value == nil {
					return
				}
				curMode = sel.Value.(enums.Enum)
				st := loops.Stacks[curMode]
				if st != nil {
					curStep = st.StepLevel
				}
				updateSteps()
				stepChoose.Update()
				stepN := st.Loops[curStep].StepCount
				stepNSpin.SetValue(float32(stepN))
				stepNSpin.Update()
			})
		})
	}

	gui.AddToolbarItem(p, ToolbarItem{Label: lblpfx + "Init",
		Icon:    icons.Update,
		Tooltip: "Initializes running and state for current mode.",
		Active:  ActiveStopped,
		Func: func() {
			loops.InitMode(curMode)
		},
	})

	gui.AddToolbarItem(p, ToolbarItem{Label: lblpfx + "Stop",
		Icon:    icons.Stop,
		Tooltip: "Interrupts current running. Will pick back up where it left off.",
		Active:  ActiveRunning,
		Func: func() {
			loops.Stop(etime.Cycle)
			// fmt.Println("Stop time!")
			gui.StopNow = true
		},
	})

	tree.AddAt(p, pfx+"loop-run", func(w *core.Button) {
		tb := gui.Toolbar
		w.SetText("Run").SetIcon(icons.PlayArrow).
			SetTooltip("Run the current mode, picking up from where it left off last time (Init to restart)")
		w.FirstStyler(func(s *styles.Style) { s.SetEnabled(!gui.IsRunning) })
		w.OnClick(func(e events.Event) {
			if !gui.IsRunning {
				gui.IsRunning = true
				tb.Restyle()
				go func() {
					loops.Run(curMode)
					gui.Stopped()
				}()
			}
		})
	})

	tree.AddAt(p, pfx+"loop-step", func(w *core.Button) {
		tb := gui.Toolbar
		w.SetText("Step").SetIcon(icons.SkipNext).
			SetTooltip("Step the current mode, according to the following step level and N")
		w.FirstStyler(func(s *styles.Style) {
			s.SetEnabled(!gui.IsRunning)
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			if !gui.IsRunning {
				gui.IsRunning = true
				tb.Restyle()
				go func() {
					st := loops.Stacks[curMode]
					nst := int(stepNSpin.Value)
					loops.Step(curMode, nst, st.StepLevel)
					gui.Stopped()
				}()
			}
		})
	})

	tree.AddAt(p, pfx+"step-level", func(w *core.Chooser) {
		stepChoose = w
		updateSteps()
		w.SetCurrentValue(curStep.String())
		w.OnChange(func(e events.Event) {
			st := loops.Stacks[curMode]
			if w.CurrentItem.Value == nil {
				return
			}
			cs := w.CurrentItem.Value.(string)
			for _, l := range st.Order {
				if l.String() == cs {
					st.StepLevel = l
					stepNSpin.Value = float32(st.Loops[l].StepCount)
					stepNSpin.Update()
					break
				}
			}
		})
	})

	tree.AddAt(p, pfx+"step-n", func(w *core.Spinner) {
		stepNSpin = w
		w.SetStep(1).SetMin(1).SetValue(1)
		w.SetTooltip("number of iterations per step")
		w.OnChange(func(e events.Event) {
			st := loops.Stacks[curMode]
			if st != nil {
				st.StepCount = int(w.Value)
				st.Loops[st.StepLevel].StepCount = st.StepCount
			}
		})
	})

}
