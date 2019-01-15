// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import (
	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/etensor"
	"github.com/goki/ki"
)

// LayerActPars is a sample style sheet for activation parameters
var LayerActPars = ki.Props{
	".Hidden": ki.Props{ // class tag
		"Act.Gain": 600, // specific param
	},
}

// leabra.LayerStru manages the structural elements of the layer, which are common
// to any Layer type
type LayerStru struct {
	Name      string        `desc:"Name of the layer -- this must be unique within the network, which has a map for quick lookup and layers are typically accessed directly by name"`
	Class     string        `desc:"Class is for applying parameter styles, can be space separated multple tags"`
	Shape     etensor.Shape `desc:"shape of the layer -- can be 2D for basic layers and 4D for layers with sub-groups (hypercolumns)"`
	Rel       emer.Rel      `desc:"Spatial relationship to other layer, determines positioning"`
	Pos       emer.Vec3i    `desc:"position in 3D space, computed from Rel"`
	RecvPrjns PrjnList      `desc:"list of receiving projections into this layer from other layers"`
	SendPrjns PrjnList      `desc:"list of sending projections from this layer to other layers"`
}

// leabra.Layer has parameters for running a basic layer
type Layer struct {
	LayerStru
	Act Act `desc:"Activation parameters and methods for computing activations"`
	//	Inhib       Inhib       `desc:"Inhibition parameters and methods for computing layer-level inhibition"`
	LearnNeuron LearnNeuron `desc:"Learning parameters and methods that operate at the neuron level"`
	Neurons     []*Neuron   `desc:"slice of neurons for this layer"`
}
