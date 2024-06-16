// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

//go:generate core generate -add-types

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/plot/plotcore"
	"cogentcore.org/core/tensor/tensorcore"
	_ "cogentcore.org/core/vgpu/gosl/slbool/slboolcore" // include to get gui views
	"github.com/emer/emergent/v2/etime"
	"github.com/emer/emergent/v2/netview"
)

// GUI manages all standard elements of a simulation Graphical User Interface
type GUI struct {

	// how many cycles between updates of cycle-level plots
	CycleUpdateInterval int

	// true if the GUI is configured and running
	Active bool `display:"-"`

	// true if sim is running
	IsRunning bool `display:"-"`

	// flag to stop running
	StopNow bool `display:"-"`

	// plots by scope
	Plots map[etime.ScopeKey]*plotcore.PlotEditor

	// plots by scope
	TableViews map[etime.ScopeKey]*tensorcore.Table

	// tensor grid views by name -- used e.g., for Rasters or ActRFs -- use Grid(name) to access
	Grids map[string]*tensorcore.TensorGrid

	// the view update for managing updates of netview
	ViewUpdate *netview.ViewUpdate `display:"-"`

	// net data for recording in nogui mode, if !nil
	NetData *netview.NetData `display:"-"`

	// displays Sim fields on left
	StructView *core.Form `display:"-"`

	// tabs for different view elements: plots, rasters
	Tabs *core.Tabs `display:"-"`

	// Body is the content of the sim window
	Body *core.Body `display:"-"`
}

// UpdateWindow triggers an update on window body,
// to be called from within the normal event processing loop.
// See GoUpdateWindow for version to call from separate goroutine.
func (gui *GUI) UpdateWindow() {
	tb := gui.Body.GetTopAppBar()
	if tb != nil {
		tb.Restyle()
	}
	gui.Body.Scene.NeedsRender()
	// todo: could update other stuff but not really neccesary
}

// GoUpdateWindow triggers an update on window body,
// for calling from a separate goroutine.
func (gui *GUI) GoUpdateWindow() {
	gui.Body.Scene.AsyncLock()
	defer gui.Body.Scene.AsyncUnlock()

	tb := gui.Body.GetTopAppBar()
	if tb != nil {
		tb.Restyle()
	}
	gui.Body.Scene.NeedsRender()
	// todo: could update other stuff but not really neccesary
}

// Stopped is called when a run method stops running,
// from a separate goroutine (do not call from main event loop).
// Updates the IsRunning flag and toolbar.
func (gui *GUI) Stopped() {
	gui.IsRunning = false
	if gui.Body == nil {
		return
	}
	if gui.ViewUpdate != nil {
		gui.UpdateNetViewWhenStopped()
	}
	gui.GoUpdateWindow()
}

// MakeBody returns default window Body content
func (gui *GUI) MakeBody(sim any, appname, title, about string) {
	core.NoSentenceCaseFor = append(core.NoSentenceCaseFor, "github.com/emer")

	gui.Body = core.NewBody(appname).SetTitle(title)
	// gui.Body.App().About = about
	split := core.NewSplits(gui.Body)
	split.Name = "split"
	gui.StructView = core.NewForm(split).SetStruct(sim)
	gui.StructView.Name = "sv"
	if tb, ok := sim.(core.ToolbarMaker); ok {
		gui.Body.AddAppBar(tb.MakeToolbar)
	}
	gui.Tabs = core.NewTabs(split)
	gui.Tabs.Name = "tv"
	split.SetSplits(.2, .8)
}

// AddNetView adds NetView in tab with given name
func (gui *GUI) AddNetView(tabName string) *netview.NetView {
	nvt := gui.Tabs.NewTab(tabName)
	nv := netview.NewNetView(nvt)
	nv.Var = "Act"
	nv.UpdateTree() // need children
	return nv
}

// FinalizeGUI wraps the end functionality of the GUI
func (gui *GUI) FinalizeGUI(closePrompt bool) {
	if closePrompt {
		/*
			inQuitPrompt := false
			core.SetQuitReqFunc(func() {
				if inQuitPrompt {
					return
				}
				inQuitPrompt = true
				core.PromptDialog(vp, core.DlgOpts{Title: "Really Quit?",
					Prompt: "Are you <i>sure</i> you want to quit and lose any unsaved params, weights, logs, etc?"}, core.AddOk, core.AddCancel,
					gui.Win.This(), func(recv, send tree.Node, sig int64, data any) {
						if sig == int64(core.DialogAccepted) {
							core.Quit()
						} else {
							inQuitPrompt = false
						}
					})
			})

			inClosePrompt := false
			gui.Win.SetCloseReqFunc(func(w *core.Window) {
				if inClosePrompt {
					return
				}
				inClosePrompt = true
				core.PromptDialog(vp, core.DlgOpts{Title: "Really Close gui.Window?",
					Prompt: "Are you <i>sure</i> you want to close the gui.Window?  This will Quit the App as well, losing all unsaved params, weights, logs, etc"}, core.AddOk, core.AddCancel,
					gui.Win.This(), func(recv, send tree.Node, sig int64, data any) {
						if sig == int64(core.DialogAccepted) {
							core.Quit()
						} else {
							inClosePrompt = false
						}
					})
			})
		*/
	}

	// gui.Win.SetCloseCleanFunc(func(w *core.Window) {
	// 	go core.Quit() // once main gui.Window is closed, quit
	// })

	gui.Active = true
}
