// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

// leabra.Neuron holds all of the neuron (unit) level variables -- this is the basic version with no
// depression, adaptation, etc
type Neuron struct {
	Act float32 `desc:"overall rate coded activation value -- what is sent to other neurons -- typically in range 0-1"`
	Ge  float32 `desc:"total excitatory synaptic conductance -- the net excitatory input to the neuron"`
	Gi  float32 `desc:"total inhibitory synaptic conductance -- the net inhibitory input to the neuron"`

	Inet float32 `desc:"net current produced by all channels -- drives update of Vm"`
	Vm   float32 `desc:"membrane potential -- integrates Inet current over time"`

	Targ float32 `desc:"target value: drives learning to produce this activation value"`
	Ext  float32 `desc:"external input: drives activation of unit from outside influences (e.g., sensory input)"`

	act_eq       // #VIEW_HOT #CAT_Activation rate-code equivalent activity value (time-averaged spikes or just act for rate code equation, NXX1) -- this includes any short-term plasticity in synaptic efficacy (e.g., depression or enhancement -- see LeabraUnitSpec::stp parameters)
	act_nd       // #CAT_Activation non-depressed rate-code equivalent activity value (act_eq) -- this is the rate code prior to any short-term plasticity effects (e.g., depression or enhancement -- see LeabraUnitSpec::stp parameters) -- this reflects the rate of actual action potentials fired by the neuron, but not the net effect of these AP's on postsynaptic receiving neurons
	act_q0       // #CAT_Activation records the activation state at the very start of the current alpha-cycle (100 msec / 10 Hz) trial, prior to any trial-level decay -- is either act_eq or act_nd depending on act_misc.rec_nd setting -- needed for leabra TI context weight learning in the LeabraTICtxtConspec connection -- this is equivalent to old p_act_p variable -- the activation in the previous plus phase
	act_q1       // #CAT_Activation records the activation state after the first gamma-frequency (25 msec / 40Hz) quarter of the current alpha-cycle (100 msec / 10 Hz) trial -- is either act_eq or act_nd depending on act_misc.rec_nd setting
	act_q2       // #CAT_Activation records the activation state after the second gamma-frequency (25 msec / 40Hz) quarter (first half) of the current alpha-cycle (100 msec / 10 Hz) trial -- is either act_eq or act_nd depending on act_misc.rec_nd setting
	act_q3       // #CAT_Activation records the activation state after the third gamma-frequency (25 msec / 40Hz) quarter of the current alpha-cycle (100 msec / 10 Hz) trial -- is either act_eq or act_nd depending on act_misc.rec_nd setting
	act_q4       // #CAT_Activation records the activation state after the fourth gamma-frequency (25 msec / 40Hz) quarter (end) of the current alpha-cycle (100 msec / 10 Hz) trial -- is either act_eq or act_nd depending on act_misc.rec_nd setting
	act_m        // #VIEW_HOT #CAT_Activation records the traditional posterior-cortical minus phase activation, as act_q3 activation after third quarter of current alpha cycle -- is either act_eq or act_nd depending on act_misc.rec_nd setting
	act_p        // #VIEW_HOT #CAT_Activation records the traditional posterior-cortical plus_phase activation, as act_q4 activation at end of current alpha cycle -- is either act_eq or act_nd depending on act_misc.rec_nd setting
	act_dif      // #VIEW_HOT #CAT_Activation act_p - act_m -- difference between plus and minus phase acts, -- reflects the individual error gradient for this neuron in standard error-driven learning terms
	net_prv_q    // #CAT_Activation net input from the previous quarter -- this is used for delayed inhibition as specified in del_inhib on layer spec
	net_prv_trl  // #CAT_Activation net input from the previous trial -- this is used for delayed inhibition as specified in del_inhib on layer spec
	da           // #NO_SAVE #NO_SAVE #CAT_Activation delta activation: change in act_nd from one cycle to next -- can be useful to track where changes are taking place -- only updated when gui active
	avg_ss       // #CAT_Learning super-short time-scale activation average -- provides the lowest-level time integration -- for spiking this integrates over spikes before subsequent averaging, and it is also useful for rate-code to provide a longer time integral overall
	avg_s        // #CAT_Learning short time-scale activation average -- tracks the most recent activation states (integrates over avg_ss values), and represents the plus phase for learning in XCAL algorithms
	ru_avg_s_lrn // #CAT_Learning short time-scale activation average that is actually used for learning -- typically includes a small contribution from avg_m in addition to mostly avg_s, as determined by UnitSpec act_avg.ru_lrn_m -- important to ensure that when unit turns off in plus phase (short time scale), enough medium-phase trace remains so that learning signal doesn't just go all the way to 0, at which point no learning would take place
	su_avg_s_lrn // #CAT_Learning short time-scale activation average that is actually used for learning -- typically includes a small contribution from avg_m in addition to mostly avg_s, as determined by UnitSpec act_avg.ru_lrn_m -- important to ensure that when unit turns off in plus phase (short time scale), enough medium-phase trace remains so that learning signal doesn't just go all the way to 0, at which point no learning would take place
	avg_m        // #CAT_Learning medium time-scale activation average -- integrates over avg_s values, and represents the minus phase for learning in XCAL algorithms
	avg_l        // #CAT_Learning long time-scale average of medium-time scale (trial level) activation, used for the BCM-style floating threshold in XCAL
	avg_l_lrn    // #CAT_Learning how much to learn based on the long-term floating threshold (avg_l) for BCM-style Hebbian learning -- is modulated level of avg_l itself (stronger hebbian as average activation goes higher) and optionally the average amount of error experienced in the layer (to retain a common proportionality with the level of error-driven learning across layers)
	act_avg      // #VIEW_HOT #CAT_Activation average activation (of final plus phase activation state) over long time intervals (time constant = act_misc.avg_tau -- typically 200) -- useful for finding hog units and seeing overall distribution of activation -- if act_misc.avg_trace is active, then it is instead an exponentially decaying trace -- used in TD reinforcement learning
	gc_i         // #NO_SAVE #CAT_Activation total inhibitory conductance -- does NOT include the g_bar.i
	gi_syn       // #NO_SAVE #CAT_Activation aggregated synaptic inhibition (from inhib connections) -- time integral of gi_raw -- this is added with layer-level inhibition (fffb) to get the full inhibition in gc.i
	gi_self      // #NO_SAVE #CAT_Activation self inhibitory current -- requires temporal integration dynamics and thus its own variable
	gi_ex        // #NO_SAVE #CAT_Activation extra inhibitory current, e.g., from previous trial or phase -- only updated when gui active
	noise        // #NO_SAVE #CAT_Activation noise value added to unit (noise_type on unit spec determines where it is added) -- this can be used in learning in some cases

	act_sent // #NO_SAVE #EXPERT #CAT_Activation last activation value sent (only send when diff is over threshold)
	net_raw  // #NO_SAVE #EXPERT #CAT_Activation raw net input received from sending units (send delta's are added to this value)
	gi_raw   // #NO_SAVE #EXPERT #CAT_Activation raw inhib net input received from sending units (increments the deltas in send_delta)
}
