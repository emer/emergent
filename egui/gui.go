package egui

import (
	"github.com/Astera-org/models/library/elog"
	"github.com/emer/emergent/netview"
	"github.com/emer/etable/eplot"
	"github.com/emer/etable/etensor"
	"github.com/emer/etable/etview"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/giv"
	"github.com/goki/ki/ki"
	"github.com/goki/mat32"
)

// GUI manages all standard elements of a simulation Graphical User Interface
type GUI struct {
	NetViewText         string `desc:"text to display at bottom of the NetView -- has relevant network state"`
	CycleUpdateInterval int    `desc:"how many cycles between updates of cycle-level plots"`
	Active              bool   `view:"-" desc:"true if the GUI is configured and running"`
	IsRunning           bool   `view:"-" desc:"true if sim is running"`
	StopNow             bool   `view:"-" desc:"flag to stop running"`

	Plots       map[elog.ScopeKey]*eplot.Plot2D `desc:"plots by scope"`
	RasterGrids map[string]*etview.TensorGrid   `desc:"spike raster grid views, by layer name"`

	NetView    *netview.NetView `view:"-" desc:"the network viewer"`
	ToolBar    *gi.ToolBar      `view:"-" desc:"the master toolbar"`
	StructView *giv.StructView  `view:"-" desc:"displays Sim fields on left"`
	TabView    *gi.TabView      `view:"-" desc:"tabs for different view elements: plots, rasters"`
	Win        *gi.Window       `view:"-" desc:"main GUI gui.Window"`
	ViewPort   *gi.Viewport2D   `view:"-" desc:"main viewport for Window"`
}

// UpdateWindow renders the viewport associated with the main window
func (gui *GUI) UpdateWindow() {
	gui.ViewPort.SetNeedsFullRender()

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

	gui.NetView = gui.TabView.AddNewTab(netview.KiT_NetView, "NetView").(*netview.NetView)
	gui.NetView.Var = "Act"

	split.SetSplits(.2, .8)

}

// AddToolbarItem adds a toolbar item but also checks when it be active in the UI
func (gui *GUI) AddToolbarItem(item ToolbarItem) {
	switch item.Active {
	case ActiveStopped:
		gui.ToolBar.AddAction(gi.ActOpts{Label: item.Label, Icon: item.Icon, Tooltip: item.Tooltip, UpdateFunc: func(act *gi.Action) {
			act.SetActiveStateUpdt(!gui.IsRunning)
		}}, gui.Win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
			item.Func()
		})
	case ActiveRunning:
		gui.ToolBar.AddAction(gi.ActOpts{Label: item.Label, Icon: item.Icon, Tooltip: item.Tooltip, UpdateFunc: func(act *gi.Action) {
			act.SetActiveStateUpdt(gui.IsRunning)
		}}, gui.Win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
			item.Func()
		})
	case ActiveAlways:
		gui.ToolBar.AddAction(gi.ActOpts{Label: item.Label, Icon: item.Icon, Tooltip: item.Tooltip}, gui.Win.This(),
			func(recv, send ki.Ki, sig int64, data interface{}) {
				item.Func()
			})
	}
}

// AddPlots adds plots based on the unique tables we have, currently assumes they should always be plotted
func (gui *GUI) AddPlots(title string, lg *elog.Logs) {
	gui.Plots = make(map[elog.ScopeKey]*eplot.Plot2D)
	//for key, table := range Log.Tables {
	for _, key := range lg.TableOrder {
		modes, times := key.ModesAndTimes()
		time := times[0]
		mode := modes[0]
		lt := lg.Tables[key] // LogTable struct
		if doplot, has := lt.Meta["Plot"]; has {
			if doplot == "false" {
				continue
			}
		}

		plt := gui.TabView.AddNewTab(eplot.KiT_Plot2D, mode+" "+time+" Plot").(*eplot.Plot2D)
		gui.Plots[key] = plt
		plt.SetTable(lt.Table)

		for _, item := range lg.Items {
			_, ok := item.Write[key]
			if !ok {
				continue
			}
			plt.SetColParams(item.Name, item.Plot.ToBool(), item.FixMin.ToBool(), item.Range.Min, item.FixMax.ToBool(), item.Range.Max)

			plt.Params.Title = title + " " + time + " Plot"
			plt.Params.XAxisCol = time
			if xaxis, has := lt.Meta["XAxisCol"]; has {
				plt.Params.XAxisCol = xaxis
			}
			if legend, has := lt.Meta["LegendCol"]; has {
				plt.Params.LegendCol = legend
			}
		}
	}
}

// RasterGrid gets spike raster grid of given name, creating if not yet made
func (gui *GUI) RasterGrid(name string) *etview.TensorGrid {
	if gui.RasterGrids == nil {
		gui.RasterGrids = make(map[string]*etview.TensorGrid)
	}
	tsr, ok := gui.RasterGrids[name]
	if !ok {
		tsr = &etview.TensorGrid{}
		gui.RasterGrids[name] = tsr
	}
	return tsr
}

// ConfigRasterGrid configures the raster grid
func (gui *GUI) ConfigRasterGrid(tg *etview.TensorGrid, sr *etensor.Float32) {
	tg.SetStretchMax()
	sr.SetMetaData("grid-fill", "1")
	tg.SetTensor(sr)
}

// Plot returns plot for mode, time scope
func (gui *GUI) Plot(mode elog.EvalModes, time elog.Times) *eplot.Plot2D {
	return gui.PlotScope(elog.Scope(mode, time))
}

// PlotScope returns plot for given scope
func (gui *GUI) PlotScope(scope elog.ScopeKey) *eplot.Plot2D {
	if !gui.Active {
		return nil
	}
	plot, ok := gui.Plots[scope]
	if !ok {
		// fmt.Printf("egui Plot not found for scope: %s\n", scope)
		return nil
	}
	return plot
}

// UpdatePlot updates plot for given mode, time scope
func (gui *GUI) UpdatePlot(mode elog.EvalModes, time elog.Times) *eplot.Plot2D {
	plot := gui.Plot(mode, time)
	if plot != nil {
		plot.GoUpdate()
	}
	return plot
}

// UpdatePlotScope updates plot at given scope
func (gui *GUI) UpdatePlotScope(scope elog.ScopeKey) *eplot.Plot2D {
	plot := gui.PlotScope(scope)
	if plot != nil {
		plot.GoUpdate()
	}
	return plot
}

// UpdateCyclePlot updates cycle plot for given mode.
// only updates every CycleUpdateInterval
func (gui *GUI) UpdateCyclePlot(mode elog.EvalModes, cycle int) *eplot.Plot2D {
	plot := gui.Plot(mode, elog.Cycle)
	if plot == nil {
		return plot
	}
	if (gui.CycleUpdateInterval > 0) && (cycle%gui.CycleUpdateInterval == 0) {
		plot.GoUpdate()
	}
	return plot
}

// UpdateNetView updates the gui visualization of the network
// set the NetViewText field prior to updating
func (gui *GUI) UpdateNetView() {
	if gui.NetView != nil && gui.NetView.IsVisible() {
		gui.NetView.Record(gui.NetViewText)
		// note: essential to use Go version of update when called from another goroutine
		gui.NetView.GoUpdate() // note: using counters is significantly slower..
	}
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
