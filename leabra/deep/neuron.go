// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package deep

import (
	"fmt"
	"reflect"

	"github.com/emer/emergent/leabra/leabra"
)

// deep.Neuron holds the extra neuron (unit) level variables for DeepLeabra computation.
// DeepLeabra includes both attentional and predictive learning functions of the deep layers
// and thalamocortical circuitry.
// These are maintained in a separate parallel slice from the leabra.Neuron variables.
type Neuron struct {
	ActNoAttn     float32 `desc:"non-attention modulated activation of the superficial-layer neurons -- i.e., the activation prior to any modulation by the DeepAttn modulatory signal.  Using this as a driver of DeepBurst when there is DeepAttn modulation of superficial-layer activations prevents a positive-feedback loop that can be problematic."`
	DeepBurst     float32 `desc:"Deep layer bursting activation values, representing activity of layer 5b intrinsic bursting (5IB) neurons, which project into the thalamus (TRC) and other deep layers locally.  Somewhat confusingly, this is computed on the Superficial layer neurons, as a thresholded function of the unit activation.  DeepBurst is only updated during the bursting quarter(s) (typically the 4th quarter) of the alpha cycle, and it is sent via BurstCtxt projections to Deep layers (representing activation of layer 6 CT corticothalamic neurons) to drive DeepCtxt value there, and via BurstTRC projections to TRC layers to drive the plus-phase outcome activation (e.g., in Pulvinar) for predictive learning."`
	DeepBurstPrv  float32 `desc:"DeepBurst from the previous alpha trial -- this is typically used for learning in the BurstCtxt projection."`
	DeepCtxt      float32 `desc:"Temporally-delayed local integration of DeepBurst signals sent via BurstCtxt projection into separate Deep layer neurons, which represent the activation of layer 6 CT corticothalamic neurons.  DeepCtxt is updated at end of a DeepBurst quarter, and thus takes effect during subsequent quarter(s) until updated again."`
	TRCBurstGe    float32 `desc:"Total excitatory conductance received from DeepBurst activations into TRC neurons, continuously updated during the bursting quarter(s).  This drives plus-phase, outcome activation of TRC neurons."`
	DeepBurstSent float32 `desc:"Last DeepBurst activation value sent, for computing TRCBurstGe using efficient delta mechanism."`
	AttnGe        float32 `desc:"Total excitatory conductance received from from deep layer activations (representing layer 6 regular spiking CT corticothalamic neurons) via DeepAttn projections.  This is setn continuously all the time from deep layers using standard delta-based Ge computation, and drives both DeepAttn and DeepLrn values."`
	DeepAttn      float32 `desc:"Current attention modulatory value in Super neurons, based on inputs from deep layer 6 CT corticothalamic, regular spiking neurons that represents the net attentional filter applied to the superficial layers.  This value directly multiplies the superficial layer activations (Act) (ActNoAttn represents value prior to this multiplication).  Value is computed from AttnGe received via DeepAttn projections from Deep layers."`
	DeepLrn       float32 `desc:"Version of DeepAttn that modulates learning rates instead of activations -- learning is assumed to be more strongly affected than activation, so this value, computed from DeepAttnGe, typically has a lower range than DeepAttn."`
}

var NeuronVars = []string{"ActNoAttn", "DeepBurst", "DeepBurstPrv", "DeepCtxt", "TRCBurstGe", "DeepBurstSent", "AttnGet", "DeepAttn", "DeepLrn"}

var NeuronVarsMap map[string]int

var AllNeuronVars []string

func init() {
	NeuronVarsMap = make(map[string]int, len(NeuronVars))
	for i, v := range NeuronVars {
		NeuronVarsMap[v] = i
	}
	ln := len(leabra.NeuronVars)
	AllNeuronVars = make([]string, len(NeuronVars)+ln)
	copy(AllNeuronVars, leabra.NeuronVars)
	copy(AllNeuronVars[ln:], NeuronVars)
}

func (nrn *Neuron) VarNames() []string {
	return NeuronVars
}

// NeuronVarByName returns the index of the variable in the Neuron, or error
func NeuronVarByName(varNm string) (int, error) {
	i, ok := NeuronVarsMap[varNm]
	if !ok {
		return 0, fmt.Errorf("Neuron VarByName: variable name: %v not valid", varNm)
	}
	return i, nil
}

// VarByIndex returns variable using index (0 = first variable in NeuronVars list)
func (nrn *Neuron) VarByIndex(idx int) float32 {
	// todo: would be ideal to avoid having to use reflect here..
	v := reflect.ValueOf(*nrn)
	return v.Field(idx + 0).Interface().(float32)
}

// VarByName returns variable by name, or error
func (nrn *Neuron) VarByName(varNm string) (float32, error) {
	i, err := NeuronVarByName(varNm)
	if err != nil {
		return 0, err
	}
	return nrn.VarByIndex(i), nil
}
