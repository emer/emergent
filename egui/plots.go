// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

import (
	"fmt"
	"log"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/colors/gradient"
	"cogentcore.org/core/plot/plotcore"
	"cogentcore.org/core/tensor/tensorcore"
	"github.com/emer/emergent/v2/elog"
	"github.com/emer/emergent/v2/etime"
)

// AddPlots adds plots based on the unique tables we have,
// currently assumes they should always be plotted
func (gui *GUI) AddPlots(title string, lg *elog.Logs) {
	gui.Plots = make(map[etime.ScopeKey]*plotcore.PlotEditor)
	// for key, table := range Log.Tables {
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

		plt := gui.NewPlotTab(key, mode+" "+time+" Plot")
		plt.SetTable(lt.Table)
		plt.Options.FromMetaMap(lt.Meta)

		ConfigPlotFromLog(title, plt, lg, key)
	}
}

// AddMiscPlotTab adds a misc (non log-generated) plot with a new
// tab and plot of given name.
func (gui *GUI) AddMiscPlotTab(name string) *plotcore.PlotEditor {
	tab := gui.Tabs.NewTab(name)
	plt := plotcore.NewSubPlot(tab)
	gui.SetPlotByName(name, plt)
	return plt
}

func ConfigPlotFromLog(title string, plt *plotcore.PlotEditor, lg *elog.Logs, key etime.ScopeKey) {
	_, times := key.ModesAndTimes()
	time := times[0]
	lt := lg.Tables[key] // LogTable struct

	for _, item := range lg.Items {
		_, ok := item.Write[key]
		if !ok {
			continue
		}
		cp := plt.SetColumnOptions(item.Name, item.Plot, item.FixMin, item.Range.Min, item.FixMax, item.Range.Max)

		if item.Color != "" {
			cp.Color = errors.Log1(gradient.FromString(item.Color, nil))
		}
		cp.TensorIndex = item.TensorIndex
		cp.ErrColumn = item.ErrCol

		plt.Options.Title = title + " " + time + " Plot"
		plt.Options.XAxis = time
		if xaxis, has := lt.Meta["XAxis"]; has {
			plt.Options.XAxis = xaxis
		}
		if legend, has := lt.Meta["Legend"]; has {
			plt.Options.Legend = legend
		}
	}
	plt.ColumnsFromMetaMap(lt.Table.MetaData)
	plt.ColumnsFromMetaMap(lt.Meta)
}

// Plot returns plot for mode, time scope
func (gui *GUI) Plot(mode etime.Modes, time etime.Times) *plotcore.PlotEditor {
	return gui.PlotScope(etime.Scope(mode, time))
}

// PlotScope returns plot for given scope
func (gui *GUI) PlotScope(scope etime.ScopeKey) *plotcore.PlotEditor {
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

// PlotByName returns a misc plot by name (instead of scope key).
func (gui *GUI) PlotByName(name string) *plotcore.PlotEditor {
	return gui.PlotScope(etime.ScopeKey(name))
}

// SetPlot stores given plot in Plots map.
func (gui *GUI) SetPlot(scope etime.ScopeKey, plt *plotcore.PlotEditor) {
	if gui.Plots == nil {
		gui.Plots = make(map[etime.ScopeKey]*plotcore.PlotEditor)
	}
	gui.Plots[scope] = plt
}

// SetPlotByName stores given misc plot by name (instead of scope key) in Plots map.
func (gui *GUI) SetPlotByName(name string, plt *plotcore.PlotEditor) {
	gui.SetPlot(etime.ScopeKey(name), plt)
}

// UpdatePlot updates plot for given mode, time scope.
// This version should be called in the GUI event loop, e.g., for direct
// updating in a toolbar action.  Use [GoUpdatePlot] if being called from
// a separate goroutine, when the sim is running.
func (gui *GUI) UpdatePlot(mode etime.Modes, tm etime.Times) *plotcore.PlotEditor {
	plot := gui.Plot(mode, tm)
	if plot != nil {
		plot.UpdatePlot()
	}
	return plot
}

// GoUpdatePlot updates plot for given mode, time scope.
// This version is for use in a running simulation, in a separate goroutine.
// It will cause the GUI to hang if called from within the GUI event loop:
// use [UpdatePlot] for that case.
func (gui *GUI) GoUpdatePlot(mode etime.Modes, tm etime.Times) *plotcore.PlotEditor {
	plot := gui.Plot(mode, tm)
	if plot != nil {
		plot.GoUpdatePlot()
	}
	return plot
}

// UpdatePlotScope updates plot at given scope.
// This version should be called in the GUI event loop, e.g., for direct
// updating in a toolbar action.  Use [GoUpdatePlot] if being called from
// a separate goroutine, when the sim is running.
func (gui *GUI) UpdatePlotScope(scope etime.ScopeKey) *plotcore.PlotEditor {
	plot := gui.PlotScope(scope)
	if plot != nil {
		plot.UpdatePlot()
	}
	return plot
}

// GoUpdatePlotScope updates plot at given scope.
// This version is for use in a running simulation, in a separate goroutine.
// It will cause the GUI to hang if called from within the GUI event loop:
// use [UpdatePlotScope] for that case.
func (gui *GUI) GoUpdatePlotScope(scope etime.ScopeKey) *plotcore.PlotEditor {
	plot := gui.PlotScope(scope)
	if plot != nil {
		plot.GoUpdatePlot()
	}
	return plot
}

// UpdateCyclePlot updates cycle plot for given mode.
// only updates every CycleUpdateInterval.
// This version should be called in the GUI event loop, e.g., for direct
// updating in a toolbar action.  Use [GoUpdateCyclePlot] if being called from
// a separate goroutine, when the sim is running.
func (gui *GUI) UpdateCyclePlot(mode etime.Modes, cycle int) *plotcore.PlotEditor {
	plot := gui.Plot(mode, etime.Cycle)
	if plot == nil {
		return plot
	}
	if (gui.CycleUpdateInterval > 0) && (cycle%gui.CycleUpdateInterval == 0) {
		plot.UpdatePlot()
	}
	return plot
}

// GoUpdateCyclePlot updates cycle plot for given mode.
// only updates every CycleUpdateInterval.
// This version is for use in a running simulation, in a separate goroutine.
// It will cause the GUI to hang if called from within the GUI event loop:
// use [UpdateCyclePlot] for that case.
func (gui *GUI) GoUpdateCyclePlot(mode etime.Modes, cycle int) *plotcore.PlotEditor {
	plot := gui.Plot(mode, etime.Cycle)
	if plot == nil {
		return plot
	}
	if (gui.CycleUpdateInterval > 0) && (cycle%gui.CycleUpdateInterval == 0) {
		plot.GoUpdatePlot()
	}
	return plot
}

// NewPlotTab adds a new plot with given key for Plots lookup
// and using given tab label.  For ad-hoc plots, you can
// construct a ScopeKey from any two strings using etime.ScopeStr.
func (gui *GUI) NewPlotTab(key etime.ScopeKey, tabLabel string) *plotcore.PlotEditor {
	plt := plotcore.NewSubPlot(gui.Tabs.NewTab(tabLabel))
	gui.Plots[key] = plt
	return plt
}

// AddTableView adds a table view of given log,
// typically particularly useful for Debug logs.
func (gui *GUI) AddTableView(lg *elog.Logs, mode etime.Modes, time etime.Times) *tensorcore.Table {
	if gui.TableViews == nil {
		gui.TableViews = make(map[etime.ScopeKey]*tensorcore.Table)
	}

	key := etime.Scope(mode, time)
	lt, ok := lg.Tables[key]
	if !ok {
		log.Printf("ERROR: in egui.AddTableView, log: %s not found\n", key)
		return nil
	}

	tt := gui.Tabs.NewTab(mode.String() + " " + time.String() + " ")
	tv := tensorcore.NewTable(tt)
	gui.TableViews[key] = tv
	tv.SetReadOnly(true)
	tv.SetTable(lt.Table)
	return tv
}

// TableView returns TableView for mode, time scope
func (gui *GUI) TableView(mode etime.Modes, time etime.Times) *tensorcore.Table {
	if !gui.Active {
		return nil
	}
	scope := etime.Scope(mode, time)
	tv, ok := gui.TableViews[scope]
	if !ok {
		fmt.Printf("egui TableView not found for scope: %s\n", scope)
		return nil
	}
	return tv
}

// UpdateTableView updates TableView for given mode, time scope
func (gui *GUI) UpdateTableView(mode etime.Modes, time etime.Times) *tensorcore.Table {
	tv := gui.TableView(mode, time)
	if tv != nil {
		tv.AsyncUpdateTable()
	}
	return tv
}
