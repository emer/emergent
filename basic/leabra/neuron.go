// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

// leabra.Neuron holds all of the neuron (unit) level variables -- this is the most basic version with
// rate-code only and no optional features at all
type Neuron struct {
	Act  float32 `desc:"overall rate coded activation value -- what is sent to other neurons -- typically in range 0-1"`
	Ge   float32 `desc:"total excitatory synaptic conductance -- the net excitatory input to the neuron -- does *not* include Gbar.E"`
	Gi   float32 `desc:"total inhibitory synaptic conductance -- the net inhibitory input to the neuron -- does *not* include Gbar.I"`
	Inet float32 `desc:"net current produced by all channels -- drives update of Vm"`
	Vm   float32 `desc:"membrane potential -- integrates Inet current over time"`
	Targ float32 `desc:"target value: drives learning to produce this activation value"`
	Ext  float32 `desc:"external input: drives activation of unit from outside influences (e.g., sensory input)"`

	AvgSS     float32 `desc:"super-short time-scale activation average -- provides the lowest-level time integration -- for spiking this integrates over spikes before subsequent averaging, and it is also useful for rate-code to provide a longer time integral overall"`
	AvgS      float32 `desc:"short time-scale activation average -- tracks the most recent activation states (integrates over avg_ss values), and represents the plus phase for learning in XCAL algorithms"`
	AvgM      float32 `desc:"medium time-scale activation average -- integrates over avg_s values, and represents the minus phase for learning in XCAL algorithms"`
	AvgL      float32 `desc:"long time-scale average of medium-time scale (trial level) activation, used for the BCM-style floating threshold in XCAL"`
	AvgLLrn   float32 `desc:"how much to learn based on the long-term floating threshold (AvgL) for BCM-style Hebbian learning -- is modulated by level of AvgL itself (stronger Hebbian as average activation goes higher) and optionally the average amount of error experienced in the layer (to retain a common proportionality with the level of error-driven learning across layers)"`
	RuAvgSLrn float32 `desc:"short time-scale activation average that is actually used for learning, for recv unit -- typically includes a small contribution from AvgM in addition to mostly AvgS, as determined by ActAvgPars.RuLrnM -- important to ensure that when unit turns off in plus phase (short time scale), enough medium-phase trace remains so that learning signal doesn't just go all the way to 0, at which point no learning would take place"`
	SuAvgSLrn float32 `desc:"short time-scale activation average that is actually used for learning, for send unit -- typically includes a small contribution from AvgM in addition to mostly AvgS, as determined by ActAvgPars.SuLrnM -- important to ensure that when unit turns off in plus phase (short time scale), enough medium-phase trace remains so that learning signal doesn't just go all the way to 0, at which point no learning would take place"`

	ActM   float32 `desc:"records the traditional posterior-cortical minus phase activation, as activation after third quarter of current alpha cycle"`
	ActP   float32 `desc:"records the traditional posterior-cortical plus_phase activation, as activation at end of current alpha cycle"`
	ActDif float32 `desc:"ActP - ActM -- difference between plus and minus phase acts -- reflects the individual error gradient for this neuron in standard error-driven learning terms"`
	ActDel float32 `desc:"delta activation: change in Act from one cycle to next -- can be useful to track where changes are taking place"`
	ActAvg float32 `desc:"average activation (of final plus phase activation state) over long time intervals (time constant = DtPars.AvgTau -- typically 200) -- useful for finding hog units and seeing overall distribution of activation"`
	Noise  float32 `desc:"noise value added to unit (ActNoisePars determines distribution, and when / where it is added)"`

	ActSent float32 `desc:"last activation value sent (only send when diff is over threshold)"`
	GeRaw   float32 `desc:"raw excitatory conductance (net input) received from sending units (send delta's are added to this value)"`
}
