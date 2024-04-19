// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

//go:generate core generate -add-types

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/views"
	"github.com/emer/emergent/v2/etime"
	"github.com/emer/emergent/v2/netview"
	"github.com/emer/etable/v2/eplot"
	"github.com/emer/etable/v2/etview"
	_ "github.com/emer/gosl/v2/slboolview" // include to get gui views
)

// GUI manages all standard elements of a simulation Graphical User Interface
type GUI struct {

	// how many cycles between updates of cycle-level plots
	CycleUpdateInterval int

	// true if the GUI is configured and running
	Active bool `view:"-"`

	// true if sim is running
	IsRunning bool `view:"-"`

	// flag to stop running
	StopNow bool `view:"-"`

	// plots by scope
	Plots map[etime.ScopeKey]*eplot.Plot2D

	// plots by scope
	TableViews map[etime.ScopeKey]*etview.TableView

	// tensor grid views by name -- used e.g., for Rasters or ActRFs -- use Grid(name) to access
	Grids map[string]*etview.TensorGrid

	// the view update for managing updates of netview
	ViewUpdate *netview.ViewUpdate `view:"-"`

	// net data for recording in nogui mode, if !nil
	NetData *netview.NetData `view:"-"`

	// displays Sim fields on left
	StructView *views.StructView `view:"-"`

	// tabs for different view elements: plots, rasters
	Tabs *core.Tabs `view:"-"`

	// Body is the content of the sim window
	Body *core.Body `view:"-"`
}

// UpdateWindow triggers an update on window body,
// to be called from within the normal event processing loop.
// See GoUpdateWindow for version to call from separate goroutine.
func (gui *GUI) UpdateWindow() {
	tb := gui.Body.GetTopAppBar()
	if tb != nil {
		tb.ApplyStyleUpdate()
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
		tb.ApplyStyleUpdate()
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
	views.NoSentenceCaseFor = append(views.NoSentenceCaseFor, "github.com/emer")

	gui.Body = core.NewBody(appname).SetTitle(title)
	// gui.Body.App().About = about
	split := core.NewSplits(gui.Body, "split")
	gui.StructView = views.NewStructView(split, "sv").SetStruct(sim)
	if tb, ok := sim.(core.Toolbarer); ok {
		gui.Body.AddAppBar(tb.ConfigToolbar)
	}
	gui.Tabs = core.NewTabs(split, "tv")
	split.SetSplits(.2, .8)
}

// AddNetView adds NetView in tab with given name
func (gui *GUI) AddNetView(tabName string) *netview.NetView {
	nvt := gui.Tabs.NewTab(tabName)
	nv := netview.NewNetView(nvt)
	nv.Var = "Act"
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
