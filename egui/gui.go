// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

//go:generate core generate -add-types

import (
	"cogentcore.org/core/gi"
	"cogentcore.org/core/giv"
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
	ViewUpdt *netview.ViewUpdt `view:"-"`

	// net data for recording in nogui mode, if !nil
	NetData *netview.NetData `view:"-"`

	// displays Sim fields on left
	StructView *giv.StructView `view:"-"`

	// tabs for different view elements: plots, rasters
	Tabs *gi.Tabs `view:"-"`

	// Body is the content of the sim window
	Body *gi.Body `view:"-"`
}

// UpdateWindow triggers an update on window body,
// to be called from within the normal event processing loop.
// See GoUpdateWindow for version to call from separate goroutine.
func (gui *GUI) UpdateWindow() {
	updt := gui.Body.Scene.UpdateStart()
	defer gui.Body.Scene.UpdateEndRender(updt)

	tb := gui.Body.GetTopAppBar()
	if tb != nil {
		tb.UpdateBar()
	}
	// todo: could update other stuff but not really neccesary
}

// GoUpdateWindow triggers an update on window body,
// for calling from a separate goroutine.
func (gui *GUI) GoUpdateWindow() {
	updt := gui.Body.Scene.UpdateStartAsync()
	defer gui.Body.Scene.UpdateEndAsyncRender(updt)

	tb := gui.Body.GetTopAppBar()
	if tb != nil {
		tb.UpdateBar()
	}
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
	if gui.ViewUpdt != nil {
		gui.UpdateNetViewWhenStopped()
	}
	gui.GoUpdateWindow()
}

// MakeBody returns default window Body content
func (gui *GUI) MakeBody(sim any, appname, title, about string) {
	giv.NoSentenceCaseFor = append(giv.NoSentenceCaseFor, "github.com/emer")

	gui.Body = gi.NewBody(appname).SetTitle(title)
	// gui.Body.App().About = about
	split := gi.NewSplits(gui.Body, "split")
	gui.StructView = giv.NewStructView(split, "sv").SetStruct(sim)
	if tb, ok := sim.(gi.Toolbarer); ok {
		gui.Body.AddAppBar(tb.ConfigToolbar)
	}
	gui.Tabs = gi.NewTabs(split, "tv")
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
			gi.SetQuitReqFunc(func() {
				if inQuitPrompt {
					return
				}
				inQuitPrompt = true
				gi.PromptDialog(vp, gi.DlgOpts{Title: "Really Quit?",
					Prompt: "Are you <i>sure</i> you want to quit and lose any unsaved params, weights, logs, etc?"}, gi.AddOk, gi.AddCancel,
					gui.Win.This(), func(recv, send ki.Ki, sig int64, data any) {
						if sig == int64(gi.DialogAccepted) {
							gi.Quit()
						} else {
							inQuitPrompt = false
						}
					})
			})

			inClosePrompt := false
			gui.Win.SetCloseReqFunc(func(w *gi.Window) {
				if inClosePrompt {
					return
				}
				inClosePrompt = true
				gi.PromptDialog(vp, gi.DlgOpts{Title: "Really Close gui.Window?",
					Prompt: "Are you <i>sure</i> you want to close the gui.Window?  This will Quit the App as well, losing all unsaved params, weights, logs, etc"}, gi.AddOk, gi.AddCancel,
					gui.Win.This(), func(recv, send ki.Ki, sig int64, data any) {
						if sig == int64(gi.DialogAccepted) {
							gi.Quit()
						} else {
							inClosePrompt = false
						}
					})
			})
		*/
	}

	// gui.Win.SetCloseCleanFunc(func(w *gi.Window) {
	// 	go gi.Quit() // once main gui.Window is closed, quit
	// })

	gui.Active = true
}
