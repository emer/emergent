// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import (
	"fmt"

	"cogentcore.org/core/colors/colormap"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/abilities"
	"cogentcore.org/core/tree"
)

func (nv *NetView) MakeToolbar(p *tree.Plan) {
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(nv.Update).SetText("Init").SetIcon(icons.Update)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(nv.Current).SetIcon(icons.Update)
	})
	tree.Add(p, func(w *core.Button) {
		w.SetText("Options").SetIcon(icons.Settings).
			SetTooltip("set parameters that control display (font size etc)").
			OnClick(func(e events.Event) {
				d := core.NewBody().AddTitle(nv.Name + " Options")
				core.NewForm(d).SetStruct(&nv.Options).
					OnChange(func(e events.Event) {
						nv.GoUpdateView()
					})
				d.RunWindowDialog(nv)
			})
	})
	tree.Add(p, func(w *core.Separator) {})
	tree.Add(p, func(w *core.Button) {
		w.SetText("Weights").SetType(core.ButtonAction).SetMenu(func(m *core.Scene) {
			fb := core.NewFuncButton(m).SetFunc(nv.SaveWeights)
			fb.SetIcon(icons.Save)
			fb.Args[0].SetTag(`extension:".wts,.wts.gz"`)
			fb = core.NewFuncButton(m).SetFunc(nv.OpenWeights)
			fb.SetIcon(icons.Open)
			fb.Args[0].SetTag(`extension:".wts,.wts.gz"`)
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetText("Params").SetIcon(icons.Info).SetMenu(func(m *core.Scene) {
			core.NewFuncButton(m).SetFunc(nv.ShowNonDefaultParams).SetIcon(icons.Info)
			core.NewFuncButton(m).SetFunc(nv.ShowAllParams).SetIcon(icons.Info)
			core.NewFuncButton(m).SetFunc(nv.ShowKeyLayerParams).SetIcon(icons.Info)
			core.NewFuncButton(m).SetFunc(nv.ShowKeyPathParams).SetIcon(icons.Info)
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetText("Net Data").SetIcon(icons.Save).SetMenu(func(m *core.Scene) {
			core.NewFuncButton(m).SetFunc(nv.Data.SaveJSON).SetText("Save Net Data").SetIcon(icons.Save)
			core.NewFuncButton(m).SetFunc(nv.Data.OpenJSON).SetText("Open Net Data").SetIcon(icons.Open)
			core.NewSeparator(m)
			core.NewFuncButton(m).SetFunc(nv.PlotSelectedUnit).SetIcon(icons.Open)
		})
	})
	tree.Add(p, func(w *core.Separator) {})
	tree.Add(p, func(w *core.Switch) {
		w.SetText("Paths").SetChecked(nv.Options.Paths).
			SetTooltip("Toggles whether pathways between layers are shown or not").
			OnChange(func(e events.Event) {
				nv.Options.Paths = w.IsChecked()
				nv.UpdateView()
			})
	})
	ditp := "data parallel index -- for models running multiple input patterns in parallel, this selects which one is viewed"
	tree.Add(p, func(w *core.Text) {
		w.SetText("Di:").SetTooltip(ditp)
	})
	tree.Add(p, func(w *core.Spinner) {
		w.SetMin(0).SetStep(1).SetValue(float32(nv.Di)).SetTooltip(ditp)
		w.OnChange(func(e events.Event) {
			maxData := nv.Net.MaxParallelData()
			md := int(w.Value)
			if md < maxData && md >= 0 {
				nv.Di = md
			}
			w.SetValue(float32(nv.Di))
			nv.UpdateView()
		})
	})

	tree.Add(p, func(w *core.Separator) {})

	tree.Add(p, func(w *core.Switch) {
		w.SetText("Raster").SetChecked(nv.Options.Raster.On).
			SetTooltip("Toggles raster plot mode -- displays values on one axis (Z by default) and raster counter (time) along the other (X by default)").
			OnChange(func(e events.Event) {
				nv.Options.Raster.On = w.IsChecked()
				// nv.ReconfigMeshes()
				nv.UpdateView()
			})
	})
	tree.Add(p, func(w *core.Switch) {
		w.SetText("X").SetType(core.SwitchCheckbox).SetChecked(nv.Options.Raster.XAxis).
			SetTooltip("If checked, the raster (time) dimension is plotted along the X (horizontal) axis of the layers, otherwise it goes in the depth (Z) dimension").
			OnChange(func(e events.Event) {
				nv.Options.Raster.XAxis = w.IsChecked()
				nv.UpdateView()
			})
	})
	vp, ok := nv.VarOptions[nv.Var]
	if !ok {
		vp = &VarOptions{}
		vp.Defaults()
	}

	var minSpin, maxSpin *core.Spinner
	var minSwitch, maxSwitch *core.Switch

	tree.Add(p, func(w *core.Separator) {})
	tree.AddAt(p, "minSwitch", func(w *core.Switch) {
		minSwitch = w
		w.SetText("Min").SetType(core.SwitchCheckbox).SetChecked(vp.Range.FixMin).
			SetTooltip("Fix the minimum end of the displayed value range to value shown in next box.  Having both min and max fixed is recommended where possible for speed and consistent interpretability of the colors.").
			OnChange(func(e events.Event) {
				vp := nv.VarOptions[nv.Var]
				vp.Range.FixMin = w.IsChecked()
				minSpin.UpdateWidget().NeedsRender()
				nv.UpdateView()
			})
		w.Updater(func() {
			vp := nv.VarOptions[nv.Var]
			if vp != nil {
				w.SetChecked(vp.Range.FixMin)
			}
		})
	})
	tree.AddAt(p, "minSpin", func(w *core.Spinner) {
		minSpin = w
		w.SetValue(vp.Range.Min).
			OnChange(func(e events.Event) {
				vp := nv.VarOptions[nv.Var]
				vp.Range.SetMin(w.Value)
				vp.Range.FixMin = true
				minSwitch.UpdateWidget().NeedsRender()
				if vp.ZeroCtr && vp.Range.Min < 0 && vp.Range.FixMax {
					vp.Range.SetMax(-vp.Range.Min)
				}
				if vp.ZeroCtr {
					maxSpin.UpdateWidget().NeedsRender()
				}
				nv.UpdateView()
			})
		w.Updater(func() {
			vp := nv.VarOptions[nv.Var]
			if vp != nil {
				w.SetValue(vp.Range.Min)
			}
		})
	})

	tree.AddAt(p, "cmap", func(w *core.ColorMapButton) {
		nv.ColorMapButton = w
		w.MapName = string(nv.Options.ColorMap)
		w.SetTooltip("Color map for translating values into colors -- click to select alternative.")
		w.Styler(func(s *styles.Style) {
			s.Min.X.Em(10)
			s.Min.Y.Em(1.2)
			s.Grow.Set(0, 1)
		})
		w.OnChange(func(e events.Event) {
			cmap, ok := colormap.AvailableMaps[string(nv.ColorMapButton.MapName)]
			if ok {
				nv.ColorMap = cmap
			}
			nv.UpdateView()
		})
	})

	tree.AddAt(p, "maxSwitch", func(w *core.Switch) {
		maxSwitch = w
		w.SetText("Max").SetType(core.SwitchCheckbox).SetChecked(vp.Range.FixMax).
			SetTooltip("Fix the maximum end of the displayed value range to value shown in next box.  Having both min and max fixed is recommended where possible for speed and consistent interpretability of the colors.").
			OnChange(func(e events.Event) {
				vp := nv.VarOptions[nv.Var]
				vp.Range.FixMax = w.IsChecked()
				maxSpin.UpdateWidget().NeedsRender()
				nv.UpdateView()
			})
		w.Updater(func() {
			vp := nv.VarOptions[nv.Var]
			if vp != nil {
				w.SetChecked(vp.Range.FixMax)
			}
		})
	})

	tree.AddAt(p, "maxSpin", func(w *core.Spinner) {
		maxSpin = w
		w.SetValue(vp.Range.Max).OnChange(func(e events.Event) {
			vp := nv.VarOptions[nv.Var]
			vp.Range.SetMax(w.Value)
			vp.Range.FixMax = true
			maxSwitch.UpdateWidget().NeedsRender()
			if vp.ZeroCtr && vp.Range.Max > 0 && vp.Range.FixMin {
				vp.Range.SetMin(-vp.Range.Max)
			}
			if vp.ZeroCtr {
				minSpin.UpdateWidget().NeedsRender()
			}
			nv.UpdateView()
		})
		w.Updater(func() {
			vp := nv.VarOptions[nv.Var]
			if vp != nil {
				w.SetValue(vp.Range.Max)
			}
		})
	})

	tree.AddAt(p, "zeroCtrSwitch", func(w *core.Switch) {
		w.SetText("ZeroCtr").SetChecked(vp.ZeroCtr).
			SetTooltip("keep Min - Max centered around 0, and use negative heights for units -- else use full min-max range for height (no negative heights)").
			OnChange(func(e events.Event) {
				vp := nv.VarOptions[nv.Var]
				vp.ZeroCtr = w.IsChecked()
				nv.UpdateView()
			})
		w.Updater(func() {
			vp := nv.VarOptions[nv.Var]
			if vp != nil {
				w.SetChecked(vp.ZeroCtr)
			}
		})
	})
}

func (nv *NetView) MakeViewbar(p *tree.Plan) {
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.Update).SetTooltip("reset to default initial display").
			OnClick(func(e events.Event) {
				nv.SceneXYZ().SetCamera("default")
				nv.UpdateView()
			})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.ZoomIn).SetTooltip("zoom in")
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Zoom(-.05)
			nv.UpdateView()
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.ZoomOut).SetTooltip("zoom out")
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Zoom(.05)
			nv.UpdateView()
		})
	})
	tree.Add(p, func(w *core.Separator) {})
	tree.Add(p, func(w *core.Text) {
		w.SetText("Rot:").SetTooltip("rotate display")
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.KeyboardArrowLeft)
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Orbit(5, 0)
			nv.UpdateView()
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.KeyboardArrowUp)
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Orbit(0, 5)
			nv.UpdateView()
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.KeyboardArrowDown)
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Orbit(0, -5)
			nv.UpdateView()
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.KeyboardArrowRight)
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Orbit(-5, 0)
			nv.UpdateView()
		})
	})
	tree.Add(p, func(w *core.Separator) {})

	tree.Add(p, func(w *core.Text) {
		w.SetText("Pan:").SetTooltip("pan display")
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.KeyboardArrowLeft)
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Pan(-.2, 0)
			nv.UpdateView()
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.KeyboardArrowUp)
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Pan(0, .2)
			nv.UpdateView()
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.KeyboardArrowDown)
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Pan(0, -.2)
			nv.UpdateView()
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.KeyboardArrowRight)
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Pan(.2, 0)
			nv.UpdateView()
		})
	})
	tree.Add(p, func(w *core.Separator) {})

	tree.Add(p, func(w *core.Text) { w.SetText("Save:") })

	for i := 1; i <= 4; i++ {
		nm := fmt.Sprintf("%d", i)
		tree.AddAt(p, "saved-"+nm, func(w *core.Button) {
			w.SetText(nm).
				SetTooltip("first click (or + Shift) saves current view, second click restores to saved state")
			w.OnClick(func(e events.Event) {
				sc := nv.SceneXYZ()
				cam := nm
				if e.HasAllModifiers(e.Modifiers(), key.Shift) {
					sc.SaveCamera(cam)
				} else {
					err := sc.SetCamera(cam)
					if err != nil {
						sc.SaveCamera(cam)
					}
				}
				fmt.Printf("Camera %s: %v\n", cam, sc.Camera.GenGoSet(""))
				nv.UpdateView()
			})
		})
	}
	tree.Add(p, func(w *core.Separator) {})

	tree.Add(p, func(w *core.Text) {
		w.SetText("Time:").
			SetTooltip("states are recorded over time -- last N can be reviewed using these buttons")
	})

	tree.AddAt(p, "rec", func(w *core.Text) {
		w.SetText(fmt.Sprintf("  %4d  ", nv.RecNo)).
			SetTooltip("current view record: -1 means latest, 0 = earliest")
		w.Styler(func(s *styles.Style) {
			s.Min.X.Ch(5)
		})
		w.Updater(func() {
			w.SetText(fmt.Sprintf("  %4d  ", nv.RecNo))
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.FirstPage).SetTooltip("move to first record (start of history)")
		w.OnClick(func(e events.Event) {
			if nv.RecFullBkwd() {
				nv.UpdateView()
			}
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.FastRewind).SetTooltip("move earlier by N records (default 10)")
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			if nv.RecFastBkwd() {
				nv.UpdateView()
			}
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.SkipPrevious).SetTooltip("move earlier by 1")
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			if nv.RecBkwd() {
				nv.UpdateView()
			}
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.PlayArrow).SetTooltip("move to latest and always display latest (-1)")
		w.OnClick(func(e events.Event) {
			if nv.RecTrackLatest() {
				nv.UpdateView()
			}
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.SkipNext).SetTooltip("move later by 1")
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			if nv.RecFwd() {
				nv.UpdateView()
			}
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.FastForward).SetTooltip("move later by N (default 10)")
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			if nv.RecFastFwd() {
				nv.UpdateView()
			}
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.LastPage).SetTooltip("move to end (current time, tracking latest updates)")
		w.OnClick(func(e events.Event) {
			if nv.RecTrackLatest() {
				nv.UpdateView()
			}
		})
	})
}
