// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import (
	"reflect"

	"github.com/goki/ki/bitflag"
	"github.com/goki/ki/kit"
)

// leabra.Neuron holds all of the neuron (unit) level variables -- this is the most basic version with
// rate-code only and no optional features at all.
// All variables accessible via Unit interface must be float32 and start at the top, in contiguous order
type Neuron struct {
	Flags NeurFlags `desc:"bit flags for binary state variables"`
	Act   float32   `desc:"overall rate coded activation value -- what is sent to other neurons -- typically in range 0-1"`
	Ge    float32   `desc:"total excitatory synaptic conductance -- the net excitatory input to the neuron -- does *not* include Gbar.E"`
	Gi    float32   `desc:"total inhibitory synaptic conductance -- the net inhibitory input to the neuron -- does *not* include Gbar.I"`
	Inet  float32   `desc:"net current produced by all channels -- drives update of Vm"`
	Vm    float32   `desc:"membrane potential -- integrates Inet current over time"`

	Targ float32 `desc:"target value: drives learning to produce this activation value"`
	Ext  float32 `desc:"external input: drives activation of unit from outside influences (e.g., sensory input)"`

	AvgSS   float32 `desc:"super-short time-scale activation average -- provides the lowest-level time integration -- for spiking this integrates over spikes before subsequent averaging, and it is also useful for rate-code to provide a longer time integral overall"`
	AvgS    float32 `desc:"short time-scale activation average -- tracks the most recent activation states (integrates over avg_ss values), and represents the plus phase for learning in XCAL algorithms"`
	AvgM    float32 `desc:"medium time-scale activation average -- integrates over avg_s values, and represents the minus phase for learning in XCAL algorithms"`
	AvgL    float32 `desc:"long time-scale average of medium-time scale (trial level) activation, used for the BCM-style floating threshold in XCAL"`
	AvgLLrn float32 `desc:"how much to learn based on the long-term floating threshold (AvgL) for BCM-style Hebbian learning -- is modulated by level of AvgL itself (stronger Hebbian as average activation goes higher) and optionally the average amount of error experienced in the layer (to retain a common proportionality with the level of error-driven learning across layers)"`
	AvgSLrn float32 `desc:"short time-scale activation average that is actually used for learning -- typically includes a small contribution from AvgM in addition to mostly AvgS, as determined by ActAvgPars.LrnM -- important to ensure that when unit turns off in plus phase (short time scale), enough medium-phase trace remains so that learning signal doesn't just go all the way to 0, at which point no learning would take place"`

	ActM   float32 `desc:"records the traditional posterior-cortical minus phase activation, as activation after third quarter of current alpha cycle"`
	ActP   float32 `desc:"records the traditional posterior-cortical plus_phase activation, as activation at end of current alpha cycle"`
	ActDif float32 `desc:"ActP - ActM -- difference between plus and minus phase acts -- reflects the individual error gradient for this neuron in standard error-driven learning terms"`
	ActDel float32 `desc:"delta activation: change in Act from one cycle to next -- can be useful to track where changes are taking place"`
	ActAvg float32 `desc:"average activation (of final plus phase activation state) over long time intervals (time constant = DtPars.AvgTau -- typically 200) -- useful for finding hog units and seeing overall distribution of activation"`

	Noise  float32 `desc:"noise value added to unit (ActNoisePars determines distribution, and when / where it is added)"`
	GiSelf float32 `desc:"total amount of self-inhibition -- time-integrated to avoid oscillations"`

	ActSent float32 `desc:"last activation value sent (only send when diff is over threshold)"`
	GeRaw   float32 `desc:"raw excitatory conductance (net input) received from sending units (send delta's are added to this value)"`
	GeInc   float32 `desc:"delta increment in GeRaw sent using SendGeDelta"`
}

var NeuronVars = []string{"Act", "Ge", "Gi", "Inet", "Vm", "Targ", "Ext", "AvgSS", "AvgS", "AvgM", "AvgL", "AvgLLrn", "AvgSLrn", "ActM", "ActP", "ActDif", "ActDel", "ActAvg", "Noise", "GiSelf", "ActSent", "GeRaw", "GeInc"}

var NeuronVarsMap map[string]int

func init() {
	NeuronVarsMap = make(map[string]int, len(NeuronVars))
	for i, v := range NeuronVars {
		NeuronVarsMap[v] = i
	}
}

func (nrn *Neuron) VarNames() []string {
	return NeuronVars
}

func (nrn *Neuron) VarByName(varNm string) (float32, bool) {
	i, ok := NeuronVarsMap[varNm]
	if !ok {
		return 0, false
	}
	// todo: would be ideal to avoid having to use reflect here..
	v := reflect.ValueOf(*nrn)
	return v.Field(i + 1).Interface().(float32), true
}

func (nrn *Neuron) HasFlag(flag NeurFlags) bool {
	return bitflag.Has32(int32(nrn.Flags), int(flag))
}

func (nrn *Neuron) SetFlag(flag NeurFlags) {
	bitflag.Set32((*int32)(&nrn.Flags), int(flag))
}

func (nrn *Neuron) ClearFlag(flag NeurFlags) {
	bitflag.Clear32((*int32)(&nrn.Flags), int(flag))
}

func (nrn *Neuron) SetMask(mask int32) {
	bitflag.SetMask32((*int32)(&nrn.Flags), mask)
}

func (nrn *Neuron) ClearMask(mask int32) {
	bitflag.ClearMask32((*int32)(&nrn.Flags), mask)
}

// NeurFlags are bit-flags encoding relevant binary state for neurons
type NeurFlags int32

//go:generate stringer -type=NeurFlags

var KiT_NeurFlags = kit.Enums.AddEnum(NeurFlagsN, true, nil)

func (ev NeurFlags) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *NeurFlags) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }

// The neuron flags
const (
	// NeurOff flag indicates that this neuron has been turned off (i.e., lesioned)
	NeurOff NeurFlags = iota

	// NeurHasExt means the neuron has external input in its Ext field
	NeurHasExt

	// NeurHasTarg means the neuron has external target input in its Targ field
	NeurHasTarg

	// NeurHasCmpr means the neuron has external comparison input in its Targ field -- used for computing
	// comparison statistics but does not drive neural activity ever
	NeurHasCmpr

	NeurFlagsN
)

/*
more specialized flags in C++ emergent -- only add in specialized cases where needed, although
there could be conflicts potentially, so may want to just go ahead and add here..
  enum LeabraUnitFlags {        // #BITS extra flags on top of ext flags for leabra
    SUPER       = 0x00000100,   // superficial layer neocortical cell -- has deep.on role = SUPER
    DEEP        = 0x00000200,   // deep layer neocortical cell -- has deep.on role = DEEP
    TRC         = 0x00000400,   // thalamic relay cell (Pulvinar) cell -- has deep.on role = TRC

    D1R         = 0x00001000,   // has predominantly D1 receptors
    D2R         = 0x00002000,   // has predominantly D2 receptors
    ACQUISITION = 0x00004000,   // involved in Acquisition
    EXTINCTION  = 0x00008000,   // involved in Extinction
    APPETITIVE  = 0x00010000,   // appetitive (positive valence) coding
    AVERSIVE    = 0x00020000,   // aversive (negative valence) coding
    PATCH       = 0x00040000,   // patch-like structure (striosomes)
    MATRIX      = 0x00080000,   // matrix-like structure
    DORSAL      = 0x00100000,   // dorsal
    VENTRAL     = 0x00200000,   // ventral
  };

*/
