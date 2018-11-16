// Copyright (c) 2018, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
package leabra provides the basic reference leabra implementation, for rate-coded
activations and standard error-driven learning.  Other packages provide spiking
or deep leabra etc.
*/
package leabra

import (
	"github.com/emer/emergent/emer"
	"github.com/goki/ki"
)

// todo: establish good naming conventions for everything
// * Pars = parameters -- shorter than Params and just as clear? maybe not?
// * XFmY = general form of computation function??

// ActPars are the activation parameters for leabra
type ActPars struct {
	Gain float32
	NVar float32
	// etc
}

// ActFmGe computes activation based on Ge excitatory conductance
func (ap *ActPars) ActFmGe(ge float32) float32 {
	// compute the activation function
}

// etc -- all the standard things from unit and layer specs..

// todo: styling needs to be able to handle field paths for applying parameter values
// from ki.Props

// LayerActPars is a sample style sheet for activation parameters
var LayerActPars = ki.Props{
	".Hidden": ki.Props{ // class tag
		"Act.Gain": 600, // specific param
	},
}

// leabra.Layer handles most of the computation for a layer
type Layer struct {
	ki.Node            // Layers are ki.Node objects: have names, parent / child tree structure and properties
	Class   string     // Class is for styling, can be space separated multple tags
	Geom    emer.Shape // shape of the layer -- maybe want the emer.Shape type or else just []int
	Rel     emer.Rel   // relationship to other layer, determines positioning
	Pos     emer.Vec3i // position in 3D space, computed from Rel
	Act     ActPars    // Activation parameters
	Inh     InhPars    // Inhibition parameters
	// etc..  everything is just right here
	Units []Unit
}

// ActFmGe computes the activation from the Ge excitatory conductance for all units
func (ly *Layer) ActFmGe() {
	for _, un := range ly.Units {
		un.Act = ly.Act.ActFmGe(un.Ge)
	}
}

// leabra.Unit is just a data type holding values -- all functions defined on Layer
type Unit struct {
	Act float32
	Ge  float32
	Gi  float32
	// etc
}
