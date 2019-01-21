// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import (
	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/etensor"
)

// leabra.LayerStru manages the structural elements of the layer, which are common
// to any Layer type
type LayerStru struct {
	Name      string        `desc:"Name of the layer -- this must be unique within the network, which has a map for quick lookup and layers are typically accessed directly by name"`
	Class     string        `desc:"Class is for applying parameter styles, can be space separated multple tags"`
	Shape     etensor.Shape `desc:"shape of the layer -- can be 2D for basic layers and 4D for layers with sub-groups (hypercolumns)"`
	Rel       emer.Rel      `desc:"Spatial relationship to other layer, determines positioning"`
	Pos       emer.Vec3i    `desc:"position of lower-left-hand corner of layer in 3D space, computed from Rel"`
	RecvPrjns PrjnList      `desc:"list of receiving projections into this layer from other layers"`
	SendPrjns PrjnList      `desc:"list of sending projections from this layer to other layers"`
}

// emer.Layer interface methods

func (ls *LayerStru) LayName() string            { return ls.Name }
func (ls *LayerStru) LayClass() string           { return ls.Class }
func (ls *LayerStru) LayShape() *etensor.Shape   { return &ls.Shape }
func (ls *LayerStru) LayPos() emer.Vec3i         { return ls.Pos }
func (ls *LayerStru) NRecvPrjns() int            { return len(ls.RecvPrjns) }
func (ls *LayerStru) RecvPrjn(idx int) emer.Prjn { return ls.RecvPrjns[idx] }
func (ls *LayerStru) NSendPrjns() int            { return len(ls.SendPrjns) }
func (ls *LayerStru) SendPrjn(idx int) emer.Prjn { return ls.SendPrjns[idx] }

// SetShape sets the layer shape and also uses default dim names
func (ls *LayerStru) SetShape(shape []int) {
	var dnms []string
	if len(shape) == 2 {
		dnms = []string{"X", "Y"}
	} else if len(shape) == 4 {
		dnms = []string{"GX", "GY", "X", "Y"} // group X,Y
	}
	ls.Shape.SetShape(shape, nil, dnms) // row major default
}

func (ls *LayerStru) RecvPrjnBySendName(sender string) (emer.Prjn, bool) {
	for _, pj := range ls.RecvPrjns {
		if pj.Send.LayName() == sender {
			return pj, true
		}
	}
	return nil, false
}

func (ls *LayerStru) SendPrjnByRecvName(recv string) (emer.Prjn, bool) {
	for _, pj := range ls.SendPrjns {
		if pj.Recv.LayName() == recv {
			return pj, true
		}
	}
	return nil, false
}

// NUnitGroups returns the number of unit groups according to the shape parameters
// currently supported for a 4D shape, where the unit groups are the first 2 X,Y dims
// and then the units within the group are the 2nd 2
func (ls *LayerStru) NUnitGroups() int {
	if ls.Shape.NumDims() != 4 {
		return 0
	}
	sh := ls.Shape.Shape()
	return int(sh[0] * sh[1])
}

//////////////////////////////////////////////////////////////////////////////////////
//  Layer

// todo: need AvgMax Ge and Act for inhib
// todo: need overall good strategy for stats
// todo: need to pass Time around..

// leabra.Layer has parameters for running a basic rate-coded Leabra layer
type Layer struct {
	LayerStru
	Act     ActParams       `desc:"Activation parameters and methods for computing activations"`
	Inhib   InhibParams     `desc:"Inhibition parameters and methods for computing layer-level inhibition"`
	Learn   LearnNeurParams `desc:"Learning parameters and methods that operate at the neuron level"`
	Neurons []*Neuron       `desc:"slice of neurons for this layer -- flat list of len = Shape.Len()"`
	Pools   []*Pool         `desc:"inhibition and other pooled, aggregate state variables -- flat list has at least of 1 for layer, and one for each unit group if shape supports that (4D)"`
}

// UpdateParams updates all params given any changes that might have been made to individual values
func (ls *Layer) UpdateParams() {
	ls.Act.Update()
	ls.Inhib.Update()
	ls.Learn.Update()
}

// Unit is emer.Layer interface method -- only possible with Neurons in place
func (ls *Layer) Unit(idx []int) (emer.Unit, bool) {
	fidx := ls.Shape.Offset(idx)
	if int(fidx) < len(ls.Neurons) {
		return ls.Neurons[fidx], true
	}
	return nil, false
}

// Build constructs the layer state, including calling Build on the projections
// you MUST have properly configured the Inhib.UnitGroup.On setting by this point
// to properly allocate Pools for the unit groups if necessary.
func (ls *Layer) Build() {
	nu := ls.Shape.Len()
	ls.Neurons = make([]*Neuron, nu)
	np := 1
	if ls.Inhib.UnitGroup.On {
		np += ls.NUnitGroups()
	}
	ls.Pools = make([]*Pool, np)
	ls.RecvPrjns.Build()
}

// note: all basic computation can be performed on layer-level
// and prjn level

//////////////////////////////////////////////////////////////////////////////////////
//  Init methods

func (ly *Layer) InitWts() {
	for _, pj := range ly.SendPrjns {
		pj.InitWts()
	}
	for _, pl := range ly.Pools {
		pl.ActAvg.ActMAvg = ly.Inhib.ActAvg.Init
		pl.ActAvg.ActPAvg = ly.Inhib.ActAvg.Init
		pl.ActAvg.ActPAvgEff = ly.Inhib.ActAvg.EffInit()
	}
}

func (ly *Layer) InitActs() {
}

// TrialInit handles all initialization at start of new input pattern, including computing
// netinput scaling from running average activation etc.
func (ly *Layer) TrialInit() {
	for _, pl := range ly.Pools {
		ly.Inhib.ActAvg.AvgFmAct(&pl.ActAvg.ActMAvg, pl.ActM.Avg)
		ly.Inhib.ActAvg.AvgFmAct(&pl.ActAvg.ActPAvg, pl.ActP.Avg)
		ly.Inhib.ActAvg.EffFmAvg(&pl.ActAvg.ActPAvgEff, pl.ActAvg.ActPAvg)
	}
}

//////////////////////////////////////////////////////////////////////////////////////
//  Act methods

func (ly *Layer) InitGeRaw() {
}

func (ly *Layer) SendGeDelta() {
}

func (ly *Layer) GeFmGeRaw() {
}

func (ly *Layer) AvgMaxGe() {
}

func (ly *Layer) InhibFm() {
}

// todo: decide about thr param!

func (ly *Layer) ActFmG() {
	for _, nrn := range ly.Neurons {
		ly.Act.VmFmG(nrn, 0)
		ly.Act.ActFmG(nrn, 0)
	}
}
