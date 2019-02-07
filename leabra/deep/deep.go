// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package deep

///////////////////////////////////////////////////////////////////////
//  deep.go contains the DeepLeabra parameters and functions

// DeepBurstParams are parameters determining how the DeepBurst activation is computed
// from the superficial layer activation values.
type DeepBurstParams struct {
	On         bool    `desc:"Enable the computation of DeepBurst from superficial activation state"`
	FmActNoMod bool    `desc:"Use the ActNoMod activation state to compute DeepBurst activation (otherwise Act) -- if DeepMod modulation is applied to Act, then this creates a positive feedback loop that can be problematic, so using the non-modulated activation value can be better."`
	ThrRel     float32 `view:"if On" max:"1" def:"0.1,0.2,0.5" desc:"Relative component of threshold on superficial activation value, below which it does not drive DeepBurst (and above which, DeepBurst = Act).  This is the distance between the average and maximum activation values within layer (e.g., 0 = average, 1 = max).  Overall effective threshold is MAX of relative and absolute thresholds."`
	ThrAbs     float32 `view:"if On" min:"0" max:"1" def:"0.1,0.2,0.5" desc:"Absolute component of threshold on superficial activation value, blow which it does not drive DeepBurst (and above which, DeepBurst = Act).  Overall effective threshold is MAX of relative and absolute thresholds."`
}

func (db *DeepBurstParams) Update() {
}

func (db *DeepBurstParams) Defaults() {
	db.On = true
	db.FmActNoMod = true
	db.ThrRel = 0.1
	db.ThrAbs = 0.1
}

// DeepAttnParams are parameters determining how the DeepAttn and DeepLrn attentional modulation
// is computed from the DeepAttnGe inputs
type DeepAttnParams struct {
	On  bool    `desc:"Enable the computation of DeepAttn, DeepLrn from DeepAttnGe (otherwise, DeepAttn and DeepLrn = 1"`
	Min float32 `view:"if On" min:"0" max:"1" def:"0.8" desc:"Minimum DeepAttn value, which can provide a non-zero baseline for attentional modulation (typical attentional modulation levels are roughly 30% or so)"`
	Thr float32 `view:"if On" min:"0" desc:"Threshold on DeepAttnGe before deep attn is applied -- if not receiving even this amount of overall input from deep layer senders, then just set DeepAttn and DeepLrn to 1 for all neurons, as there isn't enough of a signal to go on yet"`

	Range float32 `view:"-" inactive:"+" desc:"1 - Min -- range for the DeepAttnGe to modulate value of DeepAttn, between Min and 1"`
}

func (db *DeepAttnParams) Update() {
	db.Range = 1 - db.Min
}

func (db *DeepAttnParams) Defaults() {
	db.On = true
	db.Min = 0.8
	db.Thr = 0.1
	db.Update()
}

// DeepCtxtParams are parameters determining how the DeepCtxt temporal context state is computed
type DeepAttnParams struct {
	FmPrv float32 `min:"0" max:"1" desc:"Amount of prior deep context to retain when updating new deep context: (1-FmPrv) will be used for the amount of new context to add.  This provides a built-in level of hysteresis / longer-term memory of prior information -- can also achieve this kind of functionality, with more learning dynamics, using a deep context projection from the deep layer to itself."`
	FmNew float32 `view:"-" inactive:"+" desc:"1 - ctxt_prv -- new context amount"`
}

func (db *DeepCtxtParams) Update() {
	db.FmNew = 1 - db.FmPrv
}

func (db *DeepCtxtParams) Defaults() {
	db.FmPrv = 0
	db.Update()
}

// TRCParams provides parameters for controlling behavior of thalamic relay cells
type TRCParams struct {
	POnlyM   bool    `desc:"TRC plus-phase (clamping) for TRC units only occurs if the minus phase max activation for given unit group is above .1"`
	Binarize bool    `desc:"Apply threshold to DeepBurstGe input for computing plus-phase activations -- above BinThr, then Act = BinOn, below = BinOff.  Typically used for one-to-one trc prjns with fixed wts = 1, so threshold is in terms of sending activation"`
	BinThr   float32 `view:"if Binarize" desc:"Threshold for binarizing -- typically used for one-to-one trc prjns with fixed wts = 1, so threshold is in terms of sending activation"`
	BinOn    float32 `def:"0.3" view:"if Binarize" desc:"Effective value for units above threshold -- lower value around 0.3 or so seems best"`
	BinOff   float32 `def:"0" view:"if Binarize" desc:"Effective value for units below threshold -- typically 0"`
}

func (tp *TRCParams) Update() {
}

func (tp *TRCParams) Defaults() {
	tp.POnlyM = false
	tp.Binarize = false
	tp.BinThr = 0.4
	tp.BinOn = 0.3
	tp.BinOff = 0
}
