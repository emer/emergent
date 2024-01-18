// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// distsplot plots histograms of random distributions
package main

//go:generate goki generate -add-types

import (
	"strconv"

	"cogentcore.org/core/events"
	"cogentcore.org/core/gi"
	"cogentcore.org/core/giv"
	"cogentcore.org/core/icons"
	"github.com/emer/emergent/v2/erand"
	"github.com/emer/etable/v2/eplot"
	"github.com/emer/etable/v2/etable"
	"github.com/emer/etable/v2/etensor"
	_ "github.com/emer/etable/v2/etview" // include to get gui views
	"github.com/emer/etable/v2/histogram"
	"github.com/emer/etable/v2/minmax"
)

func main() {
	TheSim.Config()
	TheSim.ConfigGui()
}

// LogPrec is precision for saving float values in logs
const LogPrec = 4

// Sim holds the params, table, etc
type Sim struct {

	// random params
	Dist erand.RndParams

	// number of samples
	NSamp int

	// number of bins in the histogram
	NBins int

	// range for histogram
	Range minmax.F64

	// table for raw data
	Table *etable.Table `view:"no-inline"`

	// histogram of data
	Hist *etable.Table `view:"no-inline"`

	// the plot
	Plot *eplot.Plot2D `view:"-"`
}

// TheSim is the overall state for this simulation
var TheSim Sim

// Config configures all the elements using the standard functions
func (ss *Sim) Config() {
	ss.Dist.Defaults()
	ss.Dist.Dist = erand.Gaussian
	ss.Dist.Mean = 0.5
	ss.Dist.Var = 0.15
	ss.NSamp = 1000000
	ss.NBins = 100
	ss.Range.Set(0, 1)
	ss.Update()
	ss.Table = &etable.Table{}
	ss.Hist = &etable.Table{}
	ss.ConfigTable(ss.Table)
	ss.Run()
}

// Update updates computed values
func (ss *Sim) Update() {
}

// Run generates the data and plots a histogram of results
func (ss *Sim) Run() {
	ss.Update()
	dt := ss.Table

	dt.SetNumRows(ss.NSamp)
	for vi := 0; vi < ss.NSamp; vi++ {
		vl := ss.Dist.Gen(-1)
		dt.SetCellFloat("Val", vi, float64(vl))
	}

	histogram.F64Table(ss.Hist, dt.Cols[0].(*etensor.Float64).Values, ss.NBins, ss.Range.Min, ss.Range.Max)
	if ss.Plot != nil {
		ss.Plot.UpdatePlot()
	}
}

func (ss *Sim) ConfigTable(dt *etable.Table) {
	dt.SetMetaData("name", "Data")
	dt.SetMetaData("read-only", "true")
	dt.SetMetaData("precision", strconv.Itoa(LogPrec))

	sch := etable.Schema{
		{"Val", etensor.FLOAT64, nil, nil},
	}
	dt.SetFromSchema(sch, 0)
}

func (ss *Sim) ConfigPlot(plt *eplot.Plot2D, dt *etable.Table) *eplot.Plot2D {
	plt.Params.Title = "Rand Dist Histogram"
	plt.Params.XAxisCol = "Value"
	plt.Params.Type = eplot.Bar
	plt.Params.XAxisRot = 45
	plt.SetTable(dt)
	// order of params: on, fixMin, min, fixMax, max
	plt.SetColParams("Value", eplot.Off, eplot.FloatMin, 0, eplot.FloatMax, 0)
	plt.SetColParams("Count", eplot.On, eplot.FixMin, 0, eplot.FloatMax, 0)
	return plt
}

// ConfigGui configures the GoGi gui interface for this simulation,
func (ss *Sim) ConfigGui() *gi.Body {
	b := gi.NewAppBody("distplot")
	b.App().About = `This plots histograms of random distributions. See <a href="https://github.com/emer/emergent/v2">emergent on GitHub</a>.</p>`

	split := gi.NewSplits(b, "split")

	sv := giv.NewStructView(split, "sv")
	sv.SetStruct(ss)

	tv := gi.NewTabs(split, "tv")

	pt := tv.NewTab("Histogram")
	plt := eplot.NewPlot2D(pt)
	ss.Plot = ss.ConfigPlot(plt, ss.Hist)

	split.SetSplits(.3, .7)

	b.AddAppBar(func(tb *gi.Toolbar) {
		gi.NewButton(tb).SetText("Run").SetIcon(icons.Update).
			SetTooltip("Generate data and plot histogram.").
			OnClick(func(e events.Event) {
				ss.Run()
			})
		gi.NewButton(tb).SetText("README").SetIcon(icons.FileMarkdown).
			SetTooltip("Opens your browser on the README file that contains instructions for how to run this model.").
			OnClick(func(e events.Event) {
				gi.OpenURL("https://github.com/emer/emergent/v2/blob/master/erand/distplot/README.md")
			})
	})
	b.NewWindow().Run().Wait()
	return b
}
