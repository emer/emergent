// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

//go:generate core generate -add-types

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	_ "cogentcore.org/core/goal/gosl/slbool/slboolcore" // include to get gui views
	"cogentcore.org/core/styles"
	"cogentcore.org/core/tensor/databrowser"
	"github.com/emer/emergent/v2/netview"
)

// GUI manages all standard elements of a simulation Graphical User Interface
type GUI struct {
	databrowser.Browser

	// how many cycles between updates of cycle-level plots
	CycleUpdateInterval int

	// true if the GUI is configured and running
	Active bool `display:"-"`

	// true if sim is running
	IsRunning bool `display:"-"`

	// flag to stop running
	StopNow bool `display:"-"`

	// the view update for managing updates of netview
	ViewUpdate *netview.ViewUpdate `display:"-"`

	// displays Sim fields on left
	SimForm *core.Form `display:"-"`

	// Body is the content of the sim window
	Body *core.Body `display:"-"`
}

// UpdateWindow triggers an update on window body,
// to be called from within the normal event processing loop.
// See GoUpdateWindow for version to call from separate goroutine.
func (gui *GUI) UpdateWindow() {
	if gui.Toolbar != nil {
		gui.Toolbar.Restyle()
	}
	gui.SimForm.Update()
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
	gui.Splits = split
	gui.SimForm = core.NewForm(split).SetStruct(sim)
	gui.SimForm.Name = "sim-form"
	if tb, ok := sim.(core.ToolbarMaker); ok {
		gui.Body.AddTopBar(func(bar *core.Frame) {
			gui.Toolbar = core.NewToolbar(bar)
			gui.Toolbar.Maker(gui.MakeToolbar)
			gui.Toolbar.Maker(tb.MakeToolbar)
		})
	}
	fform := core.NewFrame(split)
	fform.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Overflow.Set(styles.OverflowAuto)
		s.Grow.Set(1, 1)
	})
	gui.Files = databrowser.NewDataTree(fform)
	tabs := databrowser.NewTabs(split)
	gui.Tabs = tabs
	tabs.Name = "tabs"
	gui.Files.Tabber = tabs
	split.SetTiles(core.TileSplit, core.TileSpan)
	split.SetSplits(.2, .5, .8)
}

// AddNetView adds NetView in tab with given name
func (gui *GUI) AddNetView(tabName string) *netview.NetView {
	nv := databrowser.NewTab(gui.Tabs, tabName, func(tab *core.Frame) *netview.NetView {
		nv := netview.NewNetView(tab)
		nv.Var = "Act"
		// tb.OnFinal(events.Click, func(e events.Event) {
		// 	nv.Current()
		// 	nv.Update()
		// })
		return nv
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
