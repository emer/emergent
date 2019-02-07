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
	leabra.Network.Defaults()
}

// UpdateParams updates all the derived parameters if any have changed, for all layers
// and projections
func (nt *Network) UpdateParams() {
	leabra.Network.UpdateParams()
}

// Layer returns the deep.Layer version of the layer
func (nt *Network) Layer(idx int) *Layer {
	return nt.Layers[idx].(*Layer)
}

//////////////////////////////////////////////////////////////////////////////////////
//  Init methods

//////////////////////////////////////////////////////////////////////////////////////
//  Act methods

// Cycle runs one cycle of activation updating:
// * Sends Ge increments from sending to receiving layers
// * Average and Max Ge stats
// * Inhibition based on Ge stats and Act Stats (computed at end of Cycle)
// * Activation from Ge, Gi, and Gl
// * Average and Max Act stats
func (nt *Network) Cycle() {
	nt.SendGeDelta() // also does integ
	nt.AvgMaxGe()
	nt.InhibFmGeAct()
	nt.ActFmG()
	nt.AvgMaxAct()
}
