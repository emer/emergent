// Copyright (c) 2021, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decoder

import (
	"fmt"

	"github.com/emer/emergent/emer"
	"github.com/emer/etable/etensor"
	"github.com/goki/mat32"
)

// Sigmoid is a sigmoidal activation function decoder, which is the best choice
// for factorial, independent categories where any number of them might be active at a time.
// It learns using the delta rule for each output unit.
type Sigmoid struct {
	Lrate    float32                     `def:"0.1" desc:"learning rate"`
	Layers   []emer.Layer                `desc:"layers to decode"`
	NCats    int                         `desc:"number of different categories to decode"`
	Units    []SigmoidUnit               `desc:"unit values -- read this for decoded output"`
	NInputs  int                         `desc:"number of inputs -- total sizes of layer inputs"`
	Inputs   []float32                   `desc:"input values, copied from layers"`
	ValsTsrs map[string]*etensor.Float32 `view:"-" desc:"for holding layer values"`
	Weights  etensor.Float32             `desc:"synaptic weights: outer loop is units, inner loop is inputs"`
}

// SigmoidUnit has variables for Sigmoid decoder unit
type SigmoidUnit struct {
	Target float32 `desc:"target activation value -- typically 0 or 1 but can be within that range too"`
	Act    float32 `desc:"final activation = 1 / (1 + e^-Net) -- this is the decoded output"`
	Net    float32 `desc:"net input = sum x * w"`
}

// InitLayer initializes detector with number of categories and layers
func (sm *Sigmoid) InitLayer(ncats int, layers []emer.Layer) {
	sm.Layers = layers
	nin := 0
	for _, ly := range sm.Layers {
		nin += ly.Shape().Len()
	}
	sm.Init(ncats, nin)
}

// Init initializes detector with number of categories and number of inputs
func (sm *Sigmoid) Init(ncats, ninputs int) {
	sm.NInputs = ninputs
	sm.Lrate = 0.1 // seems pretty good
	sm.NCats = ncats
	sm.Units = make([]SigmoidUnit, ncats)
	sm.Inputs = make([]float32, sm.NInputs)
	sm.Weights.SetShape([]int{sm.NCats, sm.NInputs}, nil, []string{"Cats", "Inputs"})
	for i := range sm.Weights.Values {
		sm.Weights.Values[i] = 0.1
	}
}

// Decode decodes the given variable name from layers (forward pass).
// Decoded values are in Units[i].Act -- see also Output to get into a []float32
func (sm *Sigmoid) Decode(varNm string) {
	sm.Input(varNm)
	sm.Forward()
}

// Output returns the resulting Decoded output activation values into given slice
// which is automatically resized if not of sufficient size.
func (sm *Sigmoid) Output(acts *[]float32) {
	if cap(*acts) < sm.NCats {
		*acts = make([]float32, sm.NCats)
	} else if len(*acts) != sm.NCats {
		*acts = (*acts)[:sm.NCats]
	}
	for ui := range sm.Units {
		u := &sm.Units[ui]
		(*acts)[ui] = u.Act
	}
}

// Train trains the decoder with given target correct answers, as []float32 values.
// Returns SSE (sum squared error) of difference between targets and outputs.
// Also returns and prints an error if targets are not sufficient length for NCats.
func (sm *Sigmoid) Train(targs []float32) (float32, error) {
	if len(targs) < sm.NCats {
		err := fmt.Errorf("decoder.Sigmoid: number of targets < NCats: %d < %d", len(targs), sm.NCats)
		fmt.Println(err)
		return 0, err
	}
	for ui := range sm.Units {
		u := &sm.Units[ui]
		u.Target = targs[ui]
	}
	sse := sm.Back()
	return sse, nil
}

// ValsTsr gets value tensor of given name, creating if not yet made
func (sm *Sigmoid) ValsTsr(name string) *etensor.Float32 {
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
func (sm *Sigmoid) Input(varNm string) {
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
func (sm *Sigmoid) Forward() {
	for ui := range sm.Units {
		u := &sm.Units[ui]
		net := float32(0)
		off := ui * sm.NInputs
		for j, in := range sm.Inputs {
			net += sm.Weights.Values[off+j] * in
		}
		u.Net = net
		u.Act = 1.0 / (1.0 + mat32.FastExp(-u.Net))
	}
}

// Back compute the backward error propagation pass
// Returns SSE (sum squared error) of difference between targets and outputs.
func (sm *Sigmoid) Back() float32 {
	lr := sm.Lrate
	var sse float32
	for ui := range sm.Units {
		u := &sm.Units[ui]
		err := (u.Target - u.Act)
		sse += err * err
		del := lr * err
		off := ui * sm.NInputs
		for j, in := range sm.Inputs {
			sm.Weights.Values[off+j] += del * in
		}
	}
	return sse
}
