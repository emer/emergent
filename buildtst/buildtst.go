// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// buildtst is a simple project that links in leabra libraries and ensures
// that changes in emergent don't prevent leabra projects from building.
package main

import (
	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/netview"
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

// DefaultParams are the initial default parameters for this simulation
var DefaultParams = emer.ParamStyle{
	{"Prjn", emer.Params{
		"Prjn.Learn.Norm.On":     1,
		"Prjn.Learn.Momentum.On": 1,
		"Prjn.Learn.WtBal.On":    0,
	}},
	// "Layer": {
	// 	"Layer.Inhib.Layer.Gi": 1.8, // this is the default
	// },
	{"#Output", emer.Params{
		"Layer.Inhib.Layer.Gi": 1.4, // this turns out to be critical for small output layer
	}},
	{".Back", emer.Params{
		"Prjn.WtScale.Rel": 0.2, // this is generally quite important
	}},
}

// For build testing, add any elements here that need to be tested.
type Sim struct {
	Net     *leabra.Network  `view:"no-inline"`
	Lay     *leabra.Layer    `view:"no-inline"`
	Pats    *etable.Table    `view:"no-inline" desc:"the training patterns"`
	EpcLog  *etable.Table    `view:"no-inline" desc:"epoch-level log data"`
	Params  emer.ParamStyle  `view:"no-inline"`
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
	ss.Params = DefaultParams

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
