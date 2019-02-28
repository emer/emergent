// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import "github.com/emer/emergent/emer"

// Pool contains computed values for FFFB inhibition, and various other state values for layers
// and pools (unit groups) that can be subject to inhibition, including:
// * average / max stats on Ge and Act that drive inhibition
// * average activity overall that is used for normalizing netin (at layer level)
type Pool struct {
	StIdx, EdIdx int         `desc:"starting and ending (exlusive) indexes for the list of neurons in this pool"`
	Inhib        FFFBInhib   `desc:"FFFB inhibition computed values"`
	Ge           emer.AvgMax `desc:"average and max Ge excitatory conductance values, which drive FF inhibition"`
	Act          emer.AvgMax `desc:"average and max Act activation values, which drive FB inhibition"`
	ActM         emer.AvgMax `desc:"minus phase average and max Act activation values, for ActAvg updt"`
	ActP         emer.AvgMax `desc:"plus phase average and max Act activation values, for ActAvg updt"`
	ActAvg       ActAvg      `desc:"running-average activation levels used for netinput scaling and adaptive inhibition"`
}

func (pl *Pool) Init() {
	pl.Inhib.Init()
	pl.Ge.Init()
	pl.Act.Init()
}

// FFFBInhib contains values for computed FFFB inhibition
type FFFBInhib struct {
	FFi    float32 `desc:"computed feedforward inhibition"`
	FBi    float32 `desc:"computed feedback inhibition (total)"`
	Gi     float32 `desc:"overall value of the inhibition -- this is what is added into the unit Gi inhibition level (along with any synaptic unit-driven inhibition)"`
	GiOrig float32 `desc:"original value of the inhibition (before any  group effects set in)"`
	LayGi  float32 `desc:"for pools, this is the layer-level inhibition that is MAX'd with the pool-level inhibition to produce the net inhibition"`
}

func (fi *FFFBInhib) Init() {
	fi.FFi = 0
	fi.FBi = 0
	fi.Gi = 0
	fi.GiOrig = 0
	fi.LayGi = 0
}

// ActAvg are running-average activation levels used for netinput scaling and adaptive inhibition
type ActAvg struct {
	ActMAvg    float32 `desc:"running-average minus-phase activity -- used for adapting inhibition -- see ActAvgParams.Tau for time constant etc"`
	ActPAvg    float32 `desc:"running-average plus-phase activity -- used for netinput scaling -- see ActAvgParams.Tau for time constant etc"`
	ActPAvgEff float32 `desc:"ActPAvg * ActAvgParams.Adjust -- adjusted effective layer activity directly used in netinput scaling"`
}
