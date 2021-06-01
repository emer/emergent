// Copyright (c) 2021, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decoder

import (
	"math"

	"github.com/emer/emergent/emer"
	"github.com/emer/etable/etensor"
	"github.com/goki/mat32"
)

// SoftMax is a softmax decoder
type SoftMax struct {
	Lrate    float32                     `desc:"learning rate"`
	Layers   []emer.Layer                `desc:"layers to decode"`
	NCats    int                         `desc:"number of different categories to decode"`
	Units    []Unit                      `desc:"unit values"`
	NInputs  int                         `desc:"number of inputs -- total sizes of layer inputs"`
	Inputs   []float32                   `desc:"input values, copied from layers"`
	Targ     int                         `desc:"current target index of correct category"`
	ValsTsrs map[string]*etensor.Float32 `view:"-" desc:"for holding layer values"`
	Weights  etensor.Float32             `desc:"synaptic weights: outer loop is units, inner loop is inputs"`
}

// Unit has variables for decoder unit
type Unit struct {
	Act      float32 `desc:"final activation = e^Ge / sum e^Ge"`
	Net      float32 `desc:"net input = sum x * w"`
	Exp      float32 `desc:"exp(Net)"`
	DActDNet float32 `desc:"derivative of activation with respect to net input"`
}

// Init initializes detector with number of categories and layers
func (sm *SoftMax) Init(ncats int, layers []emer.Layer) {
	sm.NCats = ncats
	sm.Units = make([]Unit, ncats)
	sm.Layers = layers
	sm.NInputs = 0
	for _, ly := range sm.Layers {
		sm.NInputs += ly.Shape().Len()
	}
	sm.Weights.SetShape([]int{sm.NCats, sm.NInputs}, nil, []string{"Cats", "Inputs"})
}

// Decode decodes the given variable name from layers (forward pass)
func (sm *SoftMax) Decode(varNm string) {
	sm.Input(varNm)
	sm.Forward()
}

// Train trains the decoder with given target correct answer (0..NCats-1)
func (sm *SoftMax) Train(targ int) {
	sm.Targ = targ
	sm.Back()
}

// ValsTsr gets value tensor of given name, creating if not yet made
func (sm *SoftMax) ValsTsr(name string) *etensor.Float32 {
	if sm.ValsTsrs == nil {
		sm.ValsTsrs = make(map[string]*etensor.Float32)
	}
	tsr, ok := sm.ValsTsrs[name]
	if !ok {
		tsr = &etensor.Float32{}
		sm.ValsTsrs[name] = tsr
	}
	return tsr
}

// Input grabs the input from given variable in layers
func (sm *SoftMax) Input(varNm string) {
	off := 0
	for _, ly := range sm.Layers {
		tsr := sm.ValsTsr(ly.Name())
		ly.UnitValsTensor(tsr, varNm)
		for j, v := range tsr.Values {
			sm.Inputs[off+j] = v
		}
		off += ly.Shape().Len()
	}
}

// Forward compute the forward pass from input
func (sm *SoftMax) Forward() {
	max := float32(-math.MaxFloat32)
	for ui := range sm.Units {
		u := &sm.Units[ui]
		net := float32(0)
		off := ui * sm.NInputs
		for j, in := range sm.Inputs {
			net += sm.Weights.Values[off+j] * in
		}
		u.Net = net
		if net > max {
			max = net
		}
	}
	sum := float32(0)
	for ui := range sm.Units {
		u := &sm.Units[ui]
		u.Net -= max
		u.Exp = mat32.FastExp(u.Net)
		sum += u.Exp
	}
	for ui := range sm.Units {
		u := &sm.Units[ui]
		u.Act /= sum
	}
}

// Back compute the backward error propagation pass
func (sm *SoftMax) Back() {
	// for ui := range sm.Units {
	// 	u := &sm.Units[ui]
	// 	for ui := range sm.Units {
	// 		u := &sm.Units[ui]
	// 	}
	// }
}
