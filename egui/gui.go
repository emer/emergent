// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

//go:generate core generate -add-types

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
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
	SimForm *core.Form `display:"-"`

	// tabs for different view elements: plots, rasters
	Tabs *core.Tabs `display:"-"`

	// Body is the content of the sim window
	Body *core.Body `display:"-"`

	//	Toolbar is the overall sim toolbar
	Toolbar *core.Toolbar `display:"-"`
}

// UpdateWindow triggers an update on window body,
// to be called from within the normal event processing loop.
// See GoUpdateWindow for version to call from separate goroutine.
func (gui *GUI) UpdateWindow() {
	if gui.Toolbar != nil {
		gui.Toolbar.Restyle()
	}
	gui.Body.Scene.NeedsRender()
	// todo: could update other stuff but not really necessary
}

// GoUpdateWindow triggers an update on window body,
// for calling from a separate goroutine.
func (gui *GUI) GoUpdateWindow() {
	gui.Body.Scene.AsyncLock()
	defer gui.Body.Scene.AsyncUnlock()
	gui.UpdateWindow()
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
	gui.SimForm = core.NewForm(split).SetStruct(sim)
	gui.SimForm.Name = "sim-form"
	if tb, ok := sim.(core.ToolbarMaker); ok {
		gui.Body.AddTopBar(func(bar *core.Frame) {
			gui.Toolbar = core.NewToolbar(bar)
			gui.Toolbar.Maker(tb.MakeToolbar)
		})
	}
	gui.Tabs = core.NewTabs(split)
	gui.Tabs.Name = "tabs"
	split.SetSplits(.2, .8)
}

// AddNetView adds NetView in tab with given name
func (gui *GUI) AddNetView(tabName string) *netview.NetView {
	nvt, tb := gui.Tabs.NewTab(tabName)
	nv := netview.NewNetView(nvt)
	nv.Var = "Act"
	tb.OnFinal(events.Click, func(e events.Event) {
		nv.Current()
		nv.Update()
	})
	return nv
}

// FinalizeGUI wraps the end functionality of the GUI
func (gui *GUI) FinalizeGUI(closePrompt bool) {
	if closePrompt {
		gui.Body.AddCloseDialog(func(d *core.Body) bool {
			d.SetTitle("Close?")
			core.NewText(d).SetType(core.TextSupporting).SetText("Are you sure you want to close?")
			d.AddBottomBar(func(bar *core.Frame) {
				d.AddOK(bar).SetText("Close").OnClick(func(e events.Event) {
					gui.Body.Close()
				})
			})
			return true
		})
	}
	gui.Active = true
}
