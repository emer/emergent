// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

import (
	"github.com/emer/emergent/elog"
	"github.com/emer/etable/eplot"
)

// AddPlots adds plots based on the unique tables we have, currently assumes they should always be plotted
func (gui *GUI) AddPlots(title string, lg *elog.Logs) {
	gui.Plots = make(map[elog.ScopeKey]*eplot.Plot2D)
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

// SetPlot stores given plot in Plots map
func (gui *GUI) SetPlot(scope elog.ScopeKey, plt *eplot.Plot2D) {
	if gui.Plots == nil {
		gui.Plots = make(map[elog.ScopeKey]*eplot.Plot2D)
	}
	gui.Plots[scope] = plt
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
