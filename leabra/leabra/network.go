// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import (
	"github.com/emer/emergent/emer"
	"github.com/goki/ki"
	"github.com/goki/ki/kit"
)

// leabra.Network has parameters for running a basic rate-coded Leabra network
type Network struct {
	NetworkStru
	WtBalInterval int `def:"10" desc:"how frequently to update the weight balance average weight factor -- relatively expensive"`
	WtBalCtr      int `inactive:"+" desc:"counter for how long it has been since last WtBal"`
}

var KiT_Network = kit.Types.AddType(&Network{}, NetworkProps)

var NetworkProps = ki.Props{
	"ToolBar": ki.PropSlice{
		// {"Open", ki.Props{
		// 	"label": "Open",
		// 	"icon":  "file-open",
		// 	"desc":  "Open a json-formatted Ki tree structure",
		// 	"Args": ki.PropSlice{
		// 		{"File Name", ki.Props{
		// 			"default-field": "Filename",
		// 			"ext":           ".json",
		// 		}},
		// 	},
		// }},
		{"SaveWtsJSON", ki.Props{
			"label": "Save Wts...",
			"icon":  "file-save",
			"desc":  "Save json-formatted weights",
			"Args": ki.PropSlice{
				{"Weights File Name", ki.Props{
					//						"default-field": "ColorFilename",
					"ext": ".wts",
				}},
			},
		}},
		{"EditLayer", ki.Props{
			"label":       "Edit Layer...",
			"icon":        "edit",
			"desc":        "edit given layer",
			"show-return": true,
			"Args": ki.PropSlice{
				{"Layer Name", ki.Props{}},
			},
		}},
	},
}

// NewLayer returns new layer of proper type
func (nt *Network) NewLayer() emer.Layer {
	return &Layer{}
}

// NewPrjn returns new prjn of proper type
func (nt *Network) NewPrjn() emer.Prjn {
	return &Prjn{}
}

// EditLayer is gui method for accessing layers
func (nt *Network) EditLayer(name string) *Layer {
	ly, err := nt.LayerByNameTry(name)
	if err != nil {
		return nil
	}
	return ly.(*Layer)
}

// Defaults sets all the default parameters for all layers and projections
func (nt *Network) Defaults() {
	nt.WtBalInterval = 10
	nt.WtBalCtr = 0
	for li, ly := range nt.Layers {
		ly.Defaults()
		ly.SetIndex(li)
	}
}

// UpdateParams updates all the derived parameters if any have changed, for all layers
// and projections
func (nt *Network) UpdateParams() {
	for _, ly := range nt.Layers {
		ly.UpdateParams()
	}
}

// Layer returns the leabra.Layer version of the layer
func (nt *Network) Layer(idx int) *Layer {
	return nt.Layers[idx].(*Layer)
}

//////////////////////////////////////////////////////////////////////////////////////
//  Init methods

// InitWts initializes synaptic weights and all other associated long-term state variables
// including running-average state values (e.g., layer running average activations etc)
func (nt *Network) InitWts() {
	nt.WtBalCtr = 0
	for _, ly := range nt.Layers {
		if ly.IsOff() {
			continue
		}
		ly.(LeabraLayer).InitWts()
	}
	// separate pass to enforce symmetry
	for _, ly := range nt.Layers {
		if ly.IsOff() {
			continue
		}
		ly.(LeabraLayer).InitWtSym()
	}
}

// InitActs fully initializes activation state -- not automatically called
func (nt *Network) InitActs() {
	for _, ly := range nt.Layers {
		if ly.IsOff() {
			continue
		}
		ly.(LeabraLayer).InitActs()
	}
}

// InitExt initializes external input state -- call prior to applying external inputs to layers
func (nt *Network) InitExt() {
	for _, ly := range nt.Layers {
		if ly.IsOff() {
			continue
		}
		ly.(LeabraLayer).InitExt()
	}
}

// TrialInit handles all initialization at start of new input pattern, including computing
// netinput scaling from running average activation etc.
func (nt *Network) TrialInit() {
	for _, ly := range nt.Layers {
		if ly.IsOff() {
			continue
		}
		ly.(LeabraLayer).TrialInit()
	}
}

//////////////////////////////////////////////////////////////////////////////////////
//  Act methods

// Cycle runs one cycle of activation updating:
// * Sends Ge increments from sending to receiving layers
// * Average and Max Ge stats
// * Inhibition based on Ge stats and Act Stats (computed at end of Cycle)
// * Activation from Ge, Gi, and Gl
// * Average and Max Act stats
// This basic version doesn't use the time info, but more specialized types do, and we
// want to keep a consistent API for end-user code.
func (nt *Network) Cycle(ltime *Time) {
	nt.SendGDelta(ltime) // also does integ
	nt.AvgMaxGe(ltime)
	nt.InhibFmGeAct(ltime)
	nt.ActFmG(ltime)
	nt.AvgMaxAct(ltime)
}

// SendGeDelta sends change in activation since last sent, if above thresholds
// and integrates sent deltas into GeRaw and time-integrated Ge values
func (nt *Network) SendGDelta(ltime *Time) {
	nt.ThrLayFun(func(ly LeabraLayer) { ly.SendGDelta(ltime) }, "SendGDelta")
	nt.ThrLayFun(func(ly LeabraLayer) { ly.GFmInc(ltime) }, "GFmInc   ")
}

// AvgMaxGe computes the average and max Ge stats, used in inhibition
func (nt *Network) AvgMaxGe(ltime *Time) {
	nt.ThrLayFun(func(ly LeabraLayer) { ly.AvgMaxGe(ltime) }, "AvgMaxGe")
}

// InhibiFmGeAct computes inhibition Gi from Ge and Act stats within relevant Pools
func (nt *Network) InhibFmGeAct(ltime *Time) {
	nt.ThrLayFun(func(ly LeabraLayer) { ly.InhibFmGeAct(ltime) }, "InhibFmGeAct")
}

// ActFmG computes rate-code activation from Ge, Gi, Gl conductances
func (nt *Network) ActFmG(ltime *Time) {
	nt.ThrLayFun(func(ly LeabraLayer) { ly.ActFmG(ltime) }, "ActFmG   ")
}

// AvgMaxGe computes the average and max Ge stats, used in inhibition
func (nt *Network) AvgMaxAct(ltime *Time) {
	nt.ThrLayFun(func(ly LeabraLayer) { ly.AvgMaxAct(ltime) }, "AvgMaxAct")
}

// QuarterFinal does updating after end of a quarter
func (nt *Network) QuarterFinal(ltime *Time) {
	nt.ThrLayFun(func(ly LeabraLayer) { ly.QuarterFinal(ltime) }, "QuarterFinal")
}

//////////////////////////////////////////////////////////////////////////////////////
//  Learn methods

// DWt computes the weight change (learning) based on current running-average activation values
func (nt *Network) DWt() {
	nt.ThrLayFun(func(ly LeabraLayer) { ly.DWt() }, "DWt     ")
}

// WtFmDWt updates the weights from delta-weight changes.
// Also calls WtBalFmWt every WtBalInterval times
func (nt *Network) WtFmDWt() {
	nt.ThrLayFun(func(ly LeabraLayer) { ly.WtFmDWt() }, "WtFmDWt")
	nt.WtBalCtr++
	if nt.WtBalCtr >= nt.WtBalInterval {
		nt.WtBalCtr = 0
		nt.WtBalFmWt()
	}
}

// WtBalFmWt updates the weight balance factors based on average recv weights
func (nt *Network) WtBalFmWt() {
	nt.ThrLayFun(func(ly LeabraLayer) { ly.WtBalFmWt() }, "WtBalFmWt")
}
