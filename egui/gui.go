// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

import (
	"github.com/emer/emergent/etime"
	"github.com/emer/emergent/netview"
	"github.com/emer/etable/eplot"
	"github.com/emer/etable/etview"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/giv"
	"github.com/goki/ki/ki"
	"github.com/goki/mat32"
)

// GUI manages all standard elements of a simulation Graphical User Interface
type GUI struct {
	CycleUpdateInterval int  `desc:"how many cycles between updates of cycle-level plots"`
	Active              bool `view:"-" desc:"true if the GUI is configured and running"`
	IsRunning           bool `view:"-" desc:"true if sim is running"`
	StopNow             bool `view:"-" desc:"flag to stop running"`

	Plots map[etime.ScopeKey]*eplot.Plot2D `desc:"plots by scope"`
	Grids map[string]*etview.TensorGrid    `desc:"tensor grid views by name -- used e.g., for Rasters or ActRFs -- use Grid(name) to access"`

	ViewUpdt   *netview.ViewUpdt `view:"-" desc:"the view update for managing updates of netview"`
	NetData    *netview.NetData  `view:"-" desc:"net data for recording in nogui mode, if !nil"`
	ToolBar    *gi.ToolBar       `view:"-" desc:"the master toolbar"`
	StructView *giv.StructView   `view:"-" desc:"displays Sim fields on left"`
	TabView    *gi.TabView       `view:"-" desc:"tabs for different view elements: plots, rasters"`
	Win        *gi.Window        `view:"-" desc:"main GUI gui.Window"`
	ViewPort   *gi.Viewport2D    `view:"-" desc:"main viewport for Window"`
}

// UpdateWindow renders the viewport associated with the main window
func (gui *GUI) UpdateWindow() {
	gui.ViewPort.SetNeedsFullRender()
}

// Stopped is called when a run method stops running -- updates the IsRunning flag and toolbar
func (gui *GUI) Stopped() {
	gui.IsRunning = false
	if gui.Win == nil {
		return
	}
	if gui.ViewUpdt != nil {
		gui.UpdateNetViewWhenStopped()
	}
	if gui.ToolBar != nil {
		gui.ToolBar.UpdateActions()
	}
	gui.UpdateWindow()
}

// MakeWindow specifies default window settings that are largely used in all windwos
func (gui *GUI) MakeWindow(sim interface{}, appname, title, about string) {
	width := 1600
	height := 1200

	gi.SetAppName(appname)
	gi.SetAppAbout(about)

	gui.Win = gi.NewMainWindow(appname, title, width, height)

	gui.ViewPort = gui.Win.WinViewport2D()
	gui.ViewPort.UpdateStart()

	mfr := gui.Win.SetMainFrame()

	gui.ToolBar = gi.AddNewToolBar(mfr, "tbar")
	gui.ToolBar.SetStretchMaxWidth()

	split := gi.AddNewSplitView(mfr, "split")
	split.Dim = mat32.X
	split.SetStretchMax()

	gui.StructView = giv.AddNewStructView(split, "sv")
	gui.StructView.SetStruct(sim)

	gui.TabView = gi.AddNewTabView(split, "tv")

	split.SetSplits(.2, .8)
}

// AddNetView adds NetView in tab with given name
func (gui *GUI) AddNetView(tabName string) *netview.NetView {
	nv := gui.TabView.AddNewTab(netview.KiT_NetView, tabName).(*netview.NetView)
	nv.Var = "Act"
	return nv
}

// FinalizeGUI wraps the end functionality of the GUI
func (gui *GUI) FinalizeGUI(closePrompt bool) {
	vp := gui.Win.WinViewport2D()
	vp.UpdateEndNoSig(true)

	// main menu
	appnm := gi.AppName()
	mmen := gui.Win.MainMenu
	mmen.ConfigMenus([]string{appnm, "File", "Edit", "Window"})

	amen := gui.Win.MainMenu.ChildByName(appnm, 0).(*gi.Action)
	amen.Menu.AddAppMenu(gui.Win)

	emen := gui.Win.MainMenu.ChildByName("Edit", 1).(*gi.Action)
	emen.Menu.AddCopyCutPaste(gui.Win)

	if closePrompt {

		inQuitPrompt := false
		gi.SetQuitReqFunc(func() {
			if inQuitPrompt {
				return
			}
			inQuitPrompt = true
			gi.PromptDialog(vp, gi.DlgOpts{Title: "Really Quit?",
				Prompt: "Are you <i>sure</i> you want to quit and lose any unsaved params, weights, logs, etc?"}, gi.AddOk, gi.AddCancel,
				gui.Win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
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
				gui.Win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
					if sig == int64(gi.DialogAccepted) {
						gi.Quit()
					} else {
						inClosePrompt = false
					}
				})
		})
	}

	gui.Win.SetCloseCleanFunc(func(w *gi.Window) {
		go gi.Quit() // once main gui.Window is closed, quit
	})

	gui.Active = true
	gui.Win.MainMenuUpdated()
}
