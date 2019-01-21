// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

// leabra.InhibParams contains all the inhibition computation params and functions for basic Leabra
// This is included in leabra.Layer to support computation.
// This also includes other misc layer-level params such as running-average activation in the layer
// which is used for netinput rescaling and potentially for adapting inhibition over time
type InhibParams struct {
	Layer     FFFBParams   `desc:"inhibition across the entire layer"`
	UnitGroup FFFBParams   `desc:"inhibition across groups of units, for layers with 4D shape"`
	ActAvg    ActAvgParams `desc:"running-average activation computation values -- for overall estimates of layer activation levels, used in netinput scaling"`
}

func (ip *InhibParams) Update() {
	ip.Layer.Update()
	ip.UnitGroup.Update()
	ip.ActAvg.Update()
}

func (ip *InhibParams) Defaults() {
	ip.Layer.Defaults()
	ip.UnitGroup.Defaults()
	ip.ActAvg.Defaults()
}

// FFFBParams parameterizes feedforward (FF) and feedback (FB) inhibition (FFFB)
// based on average (or maximum) netinput (FF) and activation (FB)
type FFFBParams struct {
	On       bool    `desc:"enable this level of inhibition"`
	Gi       float32 `min:"0" def:"1.8" desc:"[1.5-2.3 typical, can go lower or higher as needed] overall inhibition gain -- this is main paramter to adjust to change overall activation levels -- it scales both the the ff and fb factors uniformly"`
	FF       float32 `condshow:"On=true" min:"0" def:"1" desc:"overall inhibitory contribution from feedforward inhibition -- multiplies average netinput (i.e., synaptic drive into layer) -- this anticipates upcoming changes in excitation, but if set too high, it can make activity slow to emerge -- see also ff0 for a zero-point for this value"`
	FB       float32 `condshow:"On=true" min:"0" def:"1" desc:"overall inhibitory contribution from feedback inhibition -- multiplies average activation -- this reacts to layer activation levels and works more like a thermostat (turning up when the 'heat' in the layer is too high)"`
	FBTau    float32 `condshow:"On=true" min:"0" def:"1.4;3;5" desc:"time constant in cycles, which should be milliseconds typically (roughly, how long it takes for value to change significantly -- 1.4x the half-life) for integrating feedback inhibitory values -- prevents oscillations that otherwise occur -- the fast default of 1.4 should be used for most cases but sometimes a slower value (3 or higher) can be more robust, especially when inhibition is strong or inputs are more rapidly changing"`
	MaxVsAvg float32 `condshow:"On=true" def:"0;0.5;1" desc:"what proportion of the maximum vs. average netinput to use in the feedforward inhibition computation -- 0 = all average, 1 = all max, and values in between = proportional mix between average and max (ff_netin = avg + ff_max_vs_avg * (max - avg)) -- including more max can be beneficial especially in situations where the average can vary significantly but the activity should not -- max is more robust in many situations but less flexible and sensitive to the overall distribution -- max is better for cases more closely approximating single or strictly fixed winner-take-all behavior -- 0.5 is a good compromize in many cases and generally requires a reduction of .1 or slightly more (up to .3-.5) from the gi value for 0"`
	FF0      float32 `condshow:"On=true" def:"0.1" desc:"feedforward zero point for average netinput -- below this level, no FF inhibition is computed based on avg netinput, and this value is subtraced from the ff inhib contribution above this value -- the 0.1 default should be good for most cases (and helps FF_FB produce k-winner-take-all dynamics), but if average netinputs are lower than typical, you may need to lower it"`

	FBDt float32 // #READ_ONLY #EXPERT rate = 1 / tau
}

// FFInhib returns the feedforward inhibition value based on average and max excitatory conductance within
// relevant scope
func (fb *FFFBParams) FFInhib(avgGe, maxGe float32) float32 {
	ffNetin := avgGe + fb.MaxVsAvg*(maxGe-avgGe)
	var ffi float32
	if ffNetin > fb.FF0 {
		ffi = fb.FF * (ffNetin - fb.FF0)
	}
	return ffi
}

// FBInhib computes feedback inhibition value as function of average activation
func (fb *FFFBParams) FBInhib(avgAct float32) float32 {
	fbi := fb.FB * avgAct
	return fbi
}

// FBUpdt updates feedback inhibition using time-integration rate constant
func (fb *FFFBParams) FBUpdt(fbi *float32, newFbi float32) {
	*fbi += fb.FBDt * (newFbi - *fbi)
}

func (fb *FFFBParams) Update() {
	fb.FBDt = 1 / fb.FBTau
}

func (fb *FFFBParams) Defaults() {
	fb.Gi = 1.8
	fb.FF = 1
	fb.FB = 1
	fb.FBTau = 1.4
	fb.MaxVsAvg = 0
	fb.FF0 = 0.1
	fb.Update()
}

///////////////////////////////////////////////////////////////////////
//  ActAvgParams

// ActAvgParams represents expected average activity levels in the layer.
// Used for computing running-average computation that is then used for netinput scaling.
// Also specifies time constant for updating average
// and for the target value for adapting inhibition in inhib_adapt.
type ActAvgParams struct {
	Init      float32 `min:"0" desc:"[typically 0.1 - 0.2] initial estimated average activity level in the layer (see also UseFirst option -- if that is off then it is used as a starting point for running average actual activity level, acts_m_avg and acts_p_avg) -- acts_p_avg is used primarily for automatic netinput scaling, to balance out layers that have different activity levels -- thus it is important that init be relatively accurate -- good idea to update from recorded acts_p_avg levels (see LayerAvgAct button, here and on network)"`
	Fixed     bool    `def:"false" desc:"if true, then the Init value is used as a constant for acts_p_avg_eff (the effective value used for netinput rescaling), instead of using the actual running average activation"`
	UseExtAct bool    `def:"false" desc:"if true, then use the activation level computed from the external inputs to this layer (avg of targ or ext unit vars) -- this will only be applied to layers with INPUT or TARGET / OUTPUT layer types, and falls back on the targ_init value if external inputs are not available or have a zero average -- implies fixed behavior"`
	UseFirst  bool    `condshow:"Fixed=false" def:"true" desc:"use the first actual average value to override targ_init value -- actual value is likely to be a better estimate than our guess"`
	Tau       float32 `condshow:"Fixed=false" def:"100" min:"1" desc:"time constant in trials for integrating time-average values at the layer level -- used for computing Pool.ActAvg.ActsMAvg, ActsPAvg"`
	Adjust    float32 `condshow:"Fixed=false" def:"1" desc:"adjustment multiplier on the computed acts_p_avg value that is used to compute acts_p_avg_eff, which is actually used for netinput rescaling -- if based on connectivity patterns or other factors the actual running-average value is resulting in netinputs that are too high or low, then this can be used to adjust the effective average activity value -- reducing the average activity with a factor < 1 will increase netinput scaling (stronger net inputs from layers that receive from this layer), and vice-versa for increasing (decreases net inputs)"`

	Dt float32 `inactive:"+" view:"-" desc:"rate = 1 / tau"`
}

// EffInit returns the initial value applied during InitWts for the AvgPAvgEff effective layer activity
func (aa *ActAvgParams) EffInit() float32 {
	if aa.Fixed {
		return aa.Init
	}
	return aa.Adjust * aa.Init
}

// AvgFmAct updates the running-average activation given average activity level in layer
func (aa *ActAvgParams) AvgFmAct(avg *float32, act float32) {
	if aa.UseFirst && *avg == aa.Init {
		*avg += 0.5 * (act - *avg)
	} else {
		*avg += aa.Dt * (act - *avg)
	}
}

// EffFmAvg updates the effective value from the running-average value
func (aa *ActAvgParams) EffFmAvg(eff *float32, avg float32) {
	if aa.Fixed {
		*eff = aa.Init
	} else {
		*eff = aa.Adjust * avg
	}
}

func (aa *ActAvgParams) Update() {
	aa.Dt = 1 / aa.Tau
}

func (aa *ActAvgParams) Defaults() {
	aa.Init = 0.15
	aa.Fixed = false
	aa.UseExtAct = false
	aa.UseFirst = true
	aa.Tau = 100
	aa.Adjust = 1
	aa.Update()
}
