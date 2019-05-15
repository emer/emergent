// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// buildtst is a simple project that links in leabra libraries and ensures
// that changes in emergent don't prevent leabra projects from building.
package main

import (
	"github.com/emer/emergent/netview"
	"github.com/emer/emergent/params"
	"github.com/emer/etable/eplot"
	"github.com/emer/etable/etable"
	"github.com/emer/leabra/leabra"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/gimain"
	"github.com/goki/ki/kit"
)

// this is the stub main for gogi that calls our actual mainrun function, at end of file
func main() {
	gimain.Main(func() {
		mainrun()
	})
}

var ParamSets = params.Sets{
	{Name: "Base", Desc: "these are the best params", Sheets: params.Sheets{
		"Sim": &params.Sheet{
			{Sel: "Sim", Desc: "best params always finish in this time",
				Params: params.Params{
					"Sim.MaxEpcs": 50,
				}},
		},
		"Network": &params.Sheet{
			{Sel: "Prjn", Desc: "norm and momentum on works better, but wt bal is not better for smaller nets",
				Params: params.Params{
					"Prjn.Learn.Norm.On":     1,
					"Prjn.Learn.Momentum.On": 1,
					"Prjn.Learn.WtBal.On":    0,
				}},
			{Sel: "Layer", Desc: "using default 1.8 inhib for all of network -- can explore",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": 1.8,
				}},
			{Sel: "#Output", Desc: "output definitely needs lower inhib -- true for smaller layers in general",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": 1.4,
				}},
			{Sel: ".Back", Desc: "top-down back-projections MUST have lower relative weight scale, otherwise network hallucinates",
				Params: params.Params{
					"Prjn.WtScale.Rel": 0.2,
				}},
		},
	}},
	{Name: "DefaultInhib", Desc: "output uses default inhib instead of lower", Sheets: params.Sheets{
		"Network": &params.Sheet{
			{Sel: "#Output", Desc: "go back to default",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": 1.8,
				}},
		},
		"Sim": &params.Sheet{
			{Sel: "Sim", Desc: "takes longer -- generally doesn't finish..",
				Params: params.Params{
					"Sim.MaxEpcs": 100,
				}},
		},
	}},
}

// For build testing, add any elements here that need to be tested.
type Sim struct {
	Net     *leabra.Network  `view:"no-inline"`
	Lay     *leabra.Layer    `view:"no-inline"`
	Pats    *etable.Table    `view:"no-inline" desc:"the training patterns"`
	EpcLog  *etable.Table    `view:"no-inline" desc:"epoch-level log data"`
	Params  params.Sets      `view:"no-inline"`
	Time    leabra.Time      `desc:"leabra timing parameters and state"`
	NetView *netview.NetView `view:"-" desc:"the network viewer"`
	EpcPlot *eplot.Plot2D    `view:"-" desc:"the epoch plot"`
}

// this registers this Sim Type and gives it properties that e.g.,
// prompt for filename for save methods.
var KiT_Sim = kit.Types.AddType(&Sim{}, nil)

// TheSim is the overall state for this simulation
var TheSim Sim

// New creates new blank elements
func (ss *Sim) New() *gi.Window {
	ss.Net = &leabra.Network{}
	ss.Lay = &leabra.Layer{}
	ss.Pats = &etable.Table{}
	ss.EpcLog = &etable.Table{}
	ss.Params = ParamSets

	width := 1600
	height := 1200
	win := gi.NewWindow2D("ra25", "Leabra Random Associator", width, height, true)

	vp := win.WinViewport2D()
	updt := vp.UpdateStart()

	mfr := win.SetMainFrame()

	nv := netview.AddNewNetView(mfr, "Network")
	nv.Var = "Act"
	// nv.Params.ColorMap = "Jet" // default is ColdHot
	// which fares pretty well in terms of discussion here:
	// https://matplotlib.org/tutorials/colors/colormaps.html
	nv.SetNet(ss.Net)
	ss.NetView = nv

	plt := eplot.AddNewPlot2D(mfr, "EpcPlot")
	plt.SetTable(ss.EpcLog)
	ss.EpcPlot = plt

	vp.UpdateEndNoSig(updt)
	return win
}

func mainrun() {
	win := TheSim.New()
	win.StartEventLoop()
}
