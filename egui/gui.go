// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

//go:generate core generate -add-types

import (
	"embed"
	"fmt"
	"net/http"
	"strings"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/labels"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/htmlcore"
	"cogentcore.org/core/styles/abilities"
	"cogentcore.org/core/system"
	"cogentcore.org/core/tree"
	_ "cogentcore.org/lab/gosl/slbool/slboolcore" // include to get gui views
	"github.com/emer/emergent/v2/etime"
	"github.com/emer/emergent/v2/netview"
	"github.com/emer/etensor/plot/plotcore"
	"github.com/emer/etensor/tensor/tensorcore"
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

	// Readme is the sim ReadMe frame
	Readme *core.Frame `display:"-"`
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
func (gui *GUI) MakeBody(sim any, appname, title, about string, readme ...embed.FS) {
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

	if len(readme) > 0 {
		gui.addReadme(readme[0], split)
	} else {
		split.SetSplits(.2, .8)
	}
}

func (gui *GUI) addReadme(readmefs embed.FS, split *core.Splits) {
	gui.Readme = core.NewFrame(split)
	gui.Readme.Name = "readme"

	split.SetSplits(.2, .5, .3)

	ctx := htmlcore.NewContext()

	ctx.GetURL = func(rawURL string) (*http.Response, error) {
		return htmlcore.GetURLFromFS(readmefs, rawURL)
	}

	ctx.AddWikilinkHandler(gui.readmeWikilink("sim"))

	ctx.OpenURL = gui.readmeOpenURL

	readme, err := readmefs.ReadFile("README.md")

	if errors.Log(err) == nil {
		htmlcore.ReadMDString(ctx, gui.Readme, string(readme))
	}
}

func (gui *GUI) readmeWikilink(prefix string) htmlcore.WikilinkHandler {
	return func(text string) (url string, label string) {
		if !strings.HasPrefix(text, prefix+":") {
			return "", ""
		}
		text = strings.TrimPrefix(text, prefix+":")
		url = prefix + "://" + text
		fmt.Println("text: ", text)
		return url, text
	}
}

// Parses URL, highlights linked button or opens URL
func (gui *GUI) readmeOpenURL(url string) {
	focusSet := false

	if strings.HasPrefix(url, "sim://") {
		fmt.Println("open url: ", url)
		text := strings.TrimPrefix(url, "sim://") 

		var pathPrefix string = ""
		hasPath := false 
		if strings.Contains(text, "/"){
			pathPrefix, text, hasPath = strings.Cut(text, "/")
		}
		fmt.Println("pathPrefix:", pathPrefix," text:", text, "hasPath: ", hasPath)

		gui.Body.Scene.WidgetWalkDown(func(cw core.Widget, cwb *core.WidgetBase) bool {
			if focusSet {
				return tree.Break
			}
			if !hasPath && !cwb.IsVisible() {
				return tree.Break
			}
			if hasPath && !strings.Contains(cw.AsTree().Path(), pathPrefix){
				return tree.Continue
			}
			label := labels.ToLabel(cw)
			if !strings.EqualFold(label, text) {
				return tree.Continue
			}
			if cwb.AbilityIs(abilities.Focusable) {
				cwb.SetFocus()
				focusSet = true
				return tree.Break
			} 
			next := core.AsWidget(tree.Next(cwb)) 
			if next.AbilityIs(abilities.Focusable) {
				next.SetFocus()
				focusSet = true
				return tree.Break
			}
			return tree.Continue
		})
	}
	if !focusSet { 
		system.TheApp.OpenURL(url)
	}
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
