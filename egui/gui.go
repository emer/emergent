// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

//go:generate core generate -add-types

import (
	"io/fs"
	"sync"

	"cogentcore.org/core/core"
	"cogentcore.org/core/enums"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/tree"
	_ "cogentcore.org/lab/gosl/slbool/slboolcore" // include to get gui views
	"cogentcore.org/lab/lab"
	"github.com/emer/emergent/v2/netview"
)

// GUI manages all standard elements of a simulation Graphical User Interface
type GUI struct {
	lab.Browser

	// CycleUpdateInterval is number of cycles between updates of cycle-level plots.
	CycleUpdateInterval int

	// Active is true if the GUI is configured and running
	Active bool `display:"-"`

	// NetViews are the created netviews.
	NetViews []*netview.NetView

	// SimForm displays the Sim object fields in the left panel.
	SimForm *core.Form `display:"-"`

	// Body is the entire content of the sim window.
	Body *core.Body `display:"-"`

	// OnStop is called when running is stopped through the GUI,
	// via the Stopped method. It should update the network view for example.
	OnStop func(mode, level enums.Enum)

	// isRunning is true if sim is running.
	isRunning bool

	// stopNow can be set via SetStopNow method under mutex protection
	// to signal the current sim to stop running.
	// It is not used directly in the looper-based control logic, which has
	// its own direct Stop function, but it is set there in case there are
	// other processes that are looking at this flag.
	stopNow bool

	runMu sync.Mutex
}

// UpdateWindow triggers an update on window body,
// to be called from within the normal event processing loop.
// See GoUpdateWindow for version to call from separate goroutine.
func (gui *GUI) UpdateWindow() {
	if gui.Toolbar != nil {
		gui.Toolbar.Restyle()
	}
	gui.SimForm.Update()
	gui.Splits.NeedsRender()
	// todo: could update other stuff but not really necessary
}

// GoUpdateWindow triggers an update on window body,
// for calling from a separate goroutine.
func (gui *GUI) GoUpdateWindow() {
	gui.Splits.Scene.AsyncLock()
	defer gui.Splits.Scene.AsyncUnlock()
	gui.UpdateWindow()
}

// StartRun should be called whenever a process starts running.
// It sets stopNow = false and isRunning = true under a mutex.
func (gui *GUI) StartRun() {
	gui.runMu.Lock()
	gui.stopNow = false
	gui.isRunning = true
	gui.runMu.Unlock()
}

// IsRunning returns the state of the isRunning flag, under a mutex.
func (gui *GUI) IsRunning() bool {
	gui.runMu.Lock()
	defer gui.runMu.Unlock()
	return gui.isRunning
}

// StopNow returns the state of the stopNow flag, under a mutex.
func (gui *GUI) StopNow() bool {
	gui.runMu.Lock()
	defer gui.runMu.Unlock()
	return gui.stopNow
}

// SetStopNow sets the stopNow flag to true, under a mutex.
func (gui *GUI) SetStopNow() {
	gui.runMu.Lock()
	gui.stopNow = true
	gui.runMu.Unlock()
}

// Stopped is called when a run method stops running,
// from a separate goroutine (do not call from main event loop).
// Turns off the isRunning flag, calls OnStop with the given arguments,
// and calls GoUpdateWindow to update window state.
func (gui *GUI) Stopped(mode, level enums.Enum) {
	gui.runMu.Lock()
	gui.isRunning = false
	gui.stopNow = true // in case anyone else is looking
	gui.runMu.Unlock()
	if gui.OnStop != nil {
		gui.OnStop(mode, level)
	}
	gui.GoUpdateWindow()
}

// NewGUIBody returns a new GUI, with an initialized Body by calling [gui.MakeBody].
func NewGUIBody(b tree.Node, sim any, fsroot fs.FS, appname, title, about string) *GUI {
	gu := &GUI{}
	gu.MakeBody(b, sim, fsroot, appname, title, about)
	return gu
}

// MakeBody initializes default Body with a top-level [core.Splits] containing
// a [core.Form] editor of the given sim object, and a filetree for the data filesystem
// rooted at fsroot, and with given app name, title, and about information.
// The first arg is an optional existing [core.Body] to make into: if nil then
// a new body is made first.
func (gui *GUI) MakeBody(b tree.Node, sim any, fsroot fs.FS, appname, title, about string) {
	core.NoSentenceCaseFor = append(core.NoSentenceCaseFor, "github.com/emer")
	if b == nil {
		gui.Body = core.NewBody(appname).SetTitle(title)
		b = gui.Body
		core.AppAbout = about
	} else {
		gui.Toolbar = core.NewToolbar(b)
	}
	split := core.NewSplits(b)
	split.Styler(func(s *styles.Style) {
		s.Min.Y.Em(40)
	})
	split.Name = "split"
	gui.Splits = split
	gui.SimForm = core.NewForm(split).SetStruct(sim)
	gui.SimForm.Name = "sim-form"
	if tb, ok := sim.(core.ToolbarMaker); ok {
		if gui.Body != nil {
			gui.Body.AddTopBar(func(bar *core.Frame) {
				gui.Toolbar = core.NewToolbar(bar)
				gui.Toolbar.Maker(gui.MakeToolbar)
				gui.Toolbar.Maker(tb.MakeToolbar)
			})
		} else {
			gui.Toolbar.Maker(gui.MakeToolbar)
			gui.Toolbar.Maker(tb.MakeToolbar)
		}
	}
	fform := core.NewFrame(split)
	fform.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Overflow.Set(styles.OverflowAuto)
		s.Grow.Set(1, 1)
	})
	gui.Files = lab.NewDataTree(fform)
	tabs := lab.NewTabs(split)
	gui.Tabs = tabs
	lab.Lab = tabs
	tabs.Name = "tabs"
	gui.FS = fsroot
	gui.DataRoot = "Root"
	gui.CycleUpdateInterval = 10
	gui.UpdateFiles()
	gui.Files.Tabber = tabs
	split.SetTiles(core.TileSplit, core.TileSpan)
	split.SetSplits(.2, .5, .8)
}

// AddNetView adds NetView in tab with given name
func (gui *GUI) AddNetView(tabName string) *netview.NetView {
	nv := lab.NewTab(gui.Tabs, tabName, func(tab *core.Frame) *netview.NetView {
		nv := netview.NewNetView(tab)
		nv.Var = "Act"
		// tb.OnFinal(events.Click, func(e events.Event) {
		// 	nv.Current()
		// 	nv.Update()
		// })
		gui.NetViews = append(gui.NetViews, nv)
		return nv
	})
	return nv
}

// NetView returns the first created netview, or nil if none.
func (gui *GUI) NetView() *netview.NetView {
	if len(gui.NetViews) == 0 {
		return nil
	}
	return gui.NetViews[0]
}

// FinalizeGUI wraps the end functionality of the GUI
func (gui *GUI) FinalizeGUI(closePrompt bool) {
	gui.Active = true
	if !closePrompt || gui.Body == nil {
		return
	}
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
