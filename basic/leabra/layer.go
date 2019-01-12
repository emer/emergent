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

// leabra.Layer handles most of the computation for a layer
type Layer struct {
	Class string        // Class is for styling, can be space separated multple tags
	Shape etensor.Shape // shape of the layer	Rel   emer.Rel      // relationship to other layer, determines positioning
	Pos   emer.Vec3i    // position in 3D space, computed from Rel
	Act   ActPars       // Activation parameters
	Inh   InhPars       // Inhibition parameters
	// etc..  everything is just right here
	Units []Unit
}
