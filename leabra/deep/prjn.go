// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package deep

import (
	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/leabra/leabra"
)

// deep.Prjn is the DeepLeabra projection, based on basic rate-coded leabra.Prjn
type Prjn struct {
	leabra.Prjn
	DeepAttnGeInc []float32 `desc:"local increment accumulator for DeepModAttnGe excitatory conductance from sending units -- this will be thread-safe"`
}

func (pj *Prjn) Defaults() {
	leabra.Prjn.Defaults()
}

func (pj *Prjn) UpdateParams() {
	leabra.Prjn.UpdateParams()
}

func (pj *Prjn) SetParams(pars emer.Params, setMsg bool) bool {
	trg := pars.Target()
	if trg != "Prjn" {
		return false
	}
	pars.Set(pj, setMsg)
	pj.UpdateParams()
	return true
}

// SendDeepModGeDelta sends the delta-activation from sending neuron index si,
// to integrate excitatory conductance on receivers
func (pj *Prjn) SendGeDelta(si int, delta float32) {
	scdel := delta * pj.GeScale
	nc := pj.SConN[si]
	st := pj.SConIdxSt[si]
	syns := pj.Syns[st : st+nc]
	scons := pj.SConIdx[st : st+nc]
	for ci := range syns {
		ri := scons[ci]
		pj.GeInc[ri] += scdel * syns[ci].Wt
	}
}

// RecvGeInc increments the receiver's GeInc from that of all the projections
func (pj *Prjn) RecvGeInc() {
	rlay := pj.Recv.(*Layer)
	for ri := range rlay.Neurons {
		rn := &rlay.Neurons[ri]
		rn.GeInc += pj.GeInc[ri]
		pj.GeInc[ri] = 0
	}
}
