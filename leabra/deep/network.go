// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package deep

import (
	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/leabra/leabra"
	"github.com/goki/ki/kit"
)

// deep.Network has parameters for running a DeepLeabra network
type Network struct {
	leabra.Network
}

var KiT_Network = kit.Types.AddType(&Network{}, NetworkProps)

var NetworkProps = leabra.NetworkProps

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
	nt.Network.Defaults()
}

// UpdateParams updates all the derived parameters if any have changed, for all layers
// and projections
func (nt *Network) UpdateParams() {
	nt.Network.UpdateParams()
}

// Layer returns the deep.Layer version of the layer
func (nt *Network) Layer(idx int) *Layer {
	return nt.Layers[idx].(*Layer)
}

//////////////////////////////////////////////////////////////////////////////////////
//  Init methods

//////////////////////////////////////////////////////////////////////////////////////
//  Act methods

// Cycle runs one cycle of activation updating
// Deep version adds call to update DeepBurst at end
func (nt *Network) Cycle(ltime *leabra.Time) {
	nt.Network.Cycle(ltime)
	nt.DeepBurst(ltime)
}

// DeepBurst is called at end of Cycle, computes DeepBurst and sends it to other layers
func (nt *Network) DeepBurst(ltime *leabra.Time) {
	nt.ThrLayFun(func(ly leabra.LeabraLayer) { ly.(DeepLayer).DeepBurstFmAct(ltime) }, "DeepBurstFmAct")
	nt.ThrLayFun(func(ly leabra.LeabraLayer) { ly.(DeepLayer).SendTRCBurstGeDelta(ltime) }, "SendTRCBurstGeDelta")
	nt.ThrLayFun(func(ly leabra.LeabraLayer) { ly.(DeepLayer).TRCBurstGeFmInc(ltime) }, "TRCBurstGeFmInc")
	nt.ThrLayFun(func(ly leabra.LeabraLayer) { ly.(DeepLayer).AvgMaxTRCBurstGe(ltime) }, "AvgMaxTRCBurstGe")
}

// QuarterFinal does updating after end of a quarter
func (nt *Network) QuarterFinal(ltime *leabra.Time) {
	nt.Network.QuarterFinal(ltime)
	nt.DeepCtxt(ltime)
}

// DeepCtxt sends DeepBurst to Deep layers and integrates DeepCtxt on Deep layers
func (nt *Network) DeepCtxt(ltime *leabra.Time) {
	nt.ThrLayFun(func(ly leabra.LeabraLayer) { ly.(DeepLayer).SendDeepCtxtGe(ltime) }, "SendDeepCtxtGe")
	nt.ThrLayFun(func(ly leabra.LeabraLayer) { ly.(DeepLayer).DeepCtxtFmGe(ltime) }, "DeepCtxtFmGe")
}
