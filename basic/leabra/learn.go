// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import "github.com/chewxy/math32"

///////////////////////////////////////////////////////////////////////
//  learn.go contains the learning params and functions for leabra

// ActAvgPars has rate constants for averaging over activations at different time scales,
// to produce the running average activation values that then drive learning in the XCAL learning rules
type ActAvgPars struct {
	SsTau  float32 `def:"2;4;7"  min:"1" desc:"time constant in cycles, which should be milliseconds typically (roughly, how long it takes for value to change significantly -- 1.4x the half-life), for continuously updating the super-short time-scale avg_ss value -- this is provides a pre-integration step before integrating into the avg_s short time scale -- it is particularly important for spiking -- in general 4 is the largest value without starting to impair learning, but a value of 7 can be combined with m_in_s = 0 with somewhat worse results"`
	STau   float32 `def:"2" min:"1" desc:"time constant in cycles, which should be milliseconds typically (roughly, how long it takes for value to change significantly -- 1.4x the half-life), for continuously updating the short time-scale avg_s value from the super-short avg_ss value (cascade mode) -- avg_s represents the plus phase learning signal that reflects the most recent past information"`
	MTau   float32 `def:"10" min:"1" desc:"time constant in cycles, which should be milliseconds typically (roughly, how long it takes for value to change significantly -- 1.4x the half-life), for continuously updating the medium time-scale avg_m value from the short avg_s value (cascade mode) -- avg_m represents the minus phase learning signal that reflects the expectation representation prior to experiencing the outcome (in addition to the outcome) -- the default value of 10 generally cannot be exceeded without impairing learning"`
	RuLrnM float32 `def:"0.1;0" min:"0" max:"1" desc:"how much of the medium term average activation to mix in with the short (plus phase) to compute the ru_avg_s_lrn variable that is used for the receiving unit's short-term average in learning -- for DELTA_FF_FB it should be 0 -- not needed -- for XCAL_CHL this is important to ensure that when unit turns off in plus phase (short time scale), enough medium-phase trace remains so that learning signal doesn't just go all the way to 0, at which point no learning would take place -- typically need faster time constant for updating s such that this trace of the m signal is lost -- can set ss_tau=7 and set this to 0 but learning is generally somewhat worse"`
	SuLrnM float32 `def:"0.5;0.1;0" min:"0" max:"1" desc:"how much of the medium term average activation to mix in with the short (plus phase) to compute the su_avg_s_lrn variable that is used for the sending unit's short-term average in learning -- for DELTA_FF_FB delta-rule based learning, this is typically .5 (half-and-half) but for XCAL_CHL it is typically the same as ru_lrn_m (.1)"`

	SsDt   float32 `view:"-" expert:"+" desc:"rate = 1 / tau"`
	SDt    float32 `view:"-" expert:"+" desc:"rate = 1 / tau"`
	MDt    float32 `view:"-" expert:"+" desc:"rate = 1 / tau"`
	RuLrnS float32 `view:"-" expert:"+" desc:"1-ru_lrn_m"`
	SuLrnS float32 `view:"-" expert:"+" desc:"1-su_lrn_m"`
}

// AvgsFmAct computes averages based on current act, and overall dt integration factor
func (aa *ActAvgPars) AvgsFmAct(ruAct, dtInteg float32, avgSS, avgS, avgM, ruAvgSlrn, suAvgSlrn *float32) {
	avgSS += dtInteg * aa.SsDt * (ruAct - avgSS)
	avgS += dtInteg * aa.SDt * (avgSS - avgS)
	avgM += dtInteg * aa.MDt * (avgS - avgM)

	ruAvgSLrn = aa.RuLrnS*avgS + aa.RuLrnM*avgM
	suAvgSlrn = aa.RuLrnS*avgS + aa.SuLrnM*avgM
}

func (aa *ActAvgPars) Update() {
	aa.SSDt = 1 / aa.SSTau
	aa.SDt = 1 / aa.STau
	aa.MDt = 1 / aa.MTau
	aa.RuLrnS = 1 - aa.RuLrnM
	aa.SuLrnS = 1 - aa.SuLrnM
}

func (aa *ActAvgPars) Defaults() {
	aa.SSTau = 4.0
	aa.STau = 2.0
	aa.MTau = 10.0
	aa.RuLrnM = 0.1
	aa.SuLrnM = 0.1
	aa.Update()

}

// AvgLPars are parameters for computing the long-term floating average value, AvgL
// which is used for driving BCM-style hebbian learning in XCAL -- this form of learning
// increases contrast of weights and generally decreases overall activity of neuron,
// to prevent "hog" units -- it is computed as a running average of the (gain multiplied)
// medium-time-scale average activation at the end of the trial.
// Also computes an adaptive amount of BCM learning, AvgLLrn, based on AvgL.
type AvgLPars struct {
	Init   float32 `def:"0.4" min:"0" max:"1" desc:"initial AvgL value at start of training"`
	Gain   float32 `def:"1.5;2;2.5;3;4;5" min:"0" desc:"gain multiplier on activation used in computing the running average AvgL value that is the key floating threshold in the BCM Hebbian learning rule -- when using the DELTA_FF_FB learning rule, it should generally be 2x what it was before with the old XCAL_CHL rule, i.e., default of 5 instead of 2.5 -- it is a good idea to experiment with this parameter a bit -- the default is on the high-side, so typically reducing a bit from initial default is a good direction"`
	Min    float32 `def:"0.2" min:"0" desc:"miniumum AvgL value -- running average cannot go lower than this value even when it otherwise would due to inactivity -- default value is generally good and typically does not need to be changed"`
	Tau    float32 `def:"10" min:"1" desc:"time constant for updating the running average AvgL -- AvgL moves toward gain*act with this time constant on every trial - longer time constants can also work fine, but the default of 10 allows for quicker reaction to beneficial weight changes"`
	LrnMax float32 `def:"0.5" min:"0" desc:"maximum AvgLLrn value, which is amount of learning driven by AvgL factor -- when AvgL is at its maximum value (i.e., gain, as act does not exceed 1), then AvgLLrn will be at this maximum value -- by default, strong amounts of this homeostatic Hebbian form of learning can be used when the receiving unit is highly active -- this will then tend to bring down the average activity of units -- the default of 0.5, in combination with the err_mod flag, works well for most models -- use around 0.0004 for a single fixed value (with err_mod flag off)"`
	LrnMin float32 `def:"0.0001;0.0004" min:"0" desc:"miniumum AvgLLrn value (amount of learning driven by AvgL factor) -- if AvgL is at its minimum value, then AvgLLrn will be at this minimum value -- neurons that are not overly active may not need to increase the contrast of their weights as much -- use around 0.0004 for a single fixed value (with err_mod flag off)"`
	ErrMod bool    `def:"true" desc:"modulate amount learning by normalized level of error within layer"`
	ModMin float32 `def:"0.01" condshow:"ErrMod=true" desc:"minimum modulation value for ErrMod-- ensures a minimum amount of self-organizing learning even for network / layers that have a very small level of error signal"`

	Dt      float32 `view:"-" expert:"+" desc:"rate = 1 / tau"`
	LrnFact float32 `view:"-" expert:"+" desc:"(LrnMax - LrnMin) / (Gain - Min)"`
}

// AvgLFmAct computes long-term average activation value, and learning factor, from given activation
func (al *AvgLPars) AvgLFmAct(avgl, act float32) (float32, float32) {
	avgl += al.Dt * (al.Gain*act - avgl)
	if avgl < al.Min {
		avgl = min
	}
	lrn := al.LrnFact * (avgl - al.Min)
	return avgl, lrn
}

// ErrModFmLayErr computes AvgLLrn multiplier from layer cosine diff avg statistic
func (al *AvgLPars) ErrModFmLayErr(layCosDiffAvg float32) float32 {
	mod := float32(1)
	if !al.ErrMod {
		return mod
	}
	mod *= math32.Max(lay, CosDiffAvg, al.ModMin)
}

func (al *AvgLPars) Update() {
	al.Dt = 1 / al.Tau
	al.LrnFact = (lrn_max - lrn_min) / (gain - min)
}

func (al *AvgLPars) Defaults() {
	al.Init = 0.4
	al.Gain = 2.5
	al.Min = 0.2
	al.Tau = 10
	al.LrnMax = 0.5
	al.LrnMin = 0.0001
	al.ErrMod = true
	al.ModMin = 0.01
	al.Update()
}

//////////////////////////////////////////////////////////////////////////////////////
//  XCalPars

// XCalPars are parameters for temporally eXtended Contrastive Attractor Learning function (XCAL)
// which is the standard learning equation for leabra .
type XCalPars struct {
	MLrn    float32 `def:"1" min:"0" desc:"multiplier on learning based on the medium-term floating average threshold which produces error-driven learning -- this is typically 1 when error-driven learning is being used, and 0 when pure Hebbian learning is used. The long-term floating average threshold is provided by the receiving unit"`
	SetLLrn bool    `def:"false" desc:"if true, set a fixed AvgLLrn weighting factor that determines how much of the long-term floating average threshold (i.e., BCM, Hebbian) component of learning is used -- this is useful for setting a fully Hebbian learning connection, e.g., by setting MLrn = 0 and LLrn = 1. If false, then the receiving unit's AvgLLrn factor is used, which dynamically modulates the amount of the long-term component as a function of how active overall it is"`
	LLrn    float32 `condshow:"SetLLrn=true" desc:"fixed l_lrn weighting factor that determines how much of the long-term floating average threshold (i.e., BCM, Hebbian) component of learning is used -- this is useful for setting a fully Hebbian learning connection, e.g., by setting MLrn = 0 and LLrn = 1."`
	DRev    float32 `def:"0.1" min:"0" max:"0.99" desc:"proportional point within LTD range where magnitude reverses to go back down to zero at zero -- err-driven svm component does better with smaller values, and BCM-like mvl component does better with larger values -- 0.1 is a compromise"`
	DThr    float32 `def:"0.0001;0.01" min:"0" desc:"minimum LTD threshold value below which no weight change occurs -- this is now *relative* to the threshold"`
	LrnThr  float   `def:"0.01" desc:"xcal learning threshold -- don't learn when sending unit activation is below this value in both phases -- due to the nature of the learning function being 0 when the sr coproduct is 0, it should not affect learning in any substantial way -- nonstandard learning algorithms that have different properties should ignore it"`

	DRevRatio float32 `inactive:"+" view:"-" desc:"-(1-DRev)/DRev -- multiplication factor in learning rule -- builds in the minus sign!"`
}

// XCAL function for weight change -- the "check mark" function -- no DGain, no ThrPMin
func (xc *XCalPars) XCalDwt(srval, thrP float32) float32 {
	var dwt float32
	if srval < xc.DThr {
		rval = 0
	} else if srval > thrP*xc.DRev {
		rval = (srval - thrP)
	} else {
		rval = srval * xc.DrevRatio
	}
	return rval
}

func (xc *XCalPars) Update() {
	if xc.DRev > 0 {
		xc.DRevRatio = -(1 - xc.DRev) / xc.DRev
	} else {
		xc.DRevRatio = -1
	}
}

func (xc *XCalPars) Defaults() {
	xc.MLrn = 1
	xc.SetLLrn = false
	xc.LLrn = 1
	xc.DRev = 0.1
	xc.DThr = 0.0001
	xc.LrnThr = 0.01
	xc.Update()
}

//////////////////////////////////////////////////////////////////////////////////////
//  WtSigPars

// WtSigPars are sigmoidal weight contrast enhancement function parameters
type WtSigPars struct {
	Gain      float32 `def:"1;6" min:"0" desc:"gain (contrast, sharpness) of the weight contrast function (1 = linear)"`
	Off       float32 `def:"1" min:"0" desc:"offset of the function (1=centered at .5, >1=higher, <1=lower) -- 1 is standard for XCAL"`
	SoftBound bool    `def:"true" desc:"apply exponential soft bounding to the weight changes"`
}

// SigFun is the sigmoid function for value w in 0-1 range, with gain and offset params
func SigFun(w, gain, off float32) float32 {
	if w <= 0 {
		return 0
	}
	if w >= 1 {
		return 1
	}
	return (1 / (1 + math32.Pow((off*(1-w))/w, gain)))
}

// SigFun61 is the sigmoid function for value w in 0-1 range, with default gain = 6, offset = 1 params
func SigFun61(w float32) float32 {
	if w <= 0 {
		return 0
	}
	if w >= 1 {
		return 1
	}
	pw := (1 - w) / w
	return (1 / (1 + pw*pw*pw*pw*pw*pw))
}

// SigInvFun is the inverse of the sigmoid function
func SigInvFun(w, gain, off float32) float32 {
	if w <= 0 {
		return 0
	}
	if w >= 1 {
		return 1
	}
	return 1 / (1 + math32.Pow((1-w)/w, 1/gain)/off)
}

// SigInvFun61 is the inverse of the sigmoid function, with default gain = 6, offset = 1 params
func SigInvFun61(w float32) float32 {
	if w <= 0 {
		return 0
	}
	if w >= 1 {
		return 1
	}
	return 1 / (1 + math32.Pow((1-w)/w, 1/6))
}

// SigFmLinWt returns sigmoidal contrast-enhanced weight from linear weight
func (ws *WtSigPars) SigFmLinWt(lw float32) float32 {
	if ws.Gain == 1 && ws.Off == 1 {
		return lw
	}
	if ws.Gain == 6 && ws.Off == 1 {
		return SigFun61(lw)
	}
	return SigFun(lw, ws.Gain, ws.Off)
}

// LinFmSigWt returns linear weight from sigmoidal contrast-enhanced weight
func (ws *WtSigPars) LinFmSigWt(sw float32) float32 {
	if ws.Gain == 1 && ws.Off == 1 {
		return sw
	}
	if ws.Gain == 6 && ws.Off == 1 {
		return SigInvFun61(sw)
	}
	return SigInvFun(sw, ws.Gain, ws.Off)
}

func (ws *WtSigPars) Defaults() {
	ws.Gain = 6
	ws.Off = 1
	ws.SoftBound = true
}

//////////////////////////////////////////////////////////////////////////////////////
//  DwtNormPars

// DwtNormPars are weight change (dwt) normalization parameters, using MAX(ABS(dwt)) aggregated over
// Sending connections in a given projection for a given unit.
// Slowly decays and instantly resets to any current max(abs)
// Serves as an estimate of the variance in the weight changes, assuming zero net mean overall.
type DwtNormPars struct {
	On       bool    `def:"true" desc:"whether to use dwt normalization, only on error-driven dwt component, based on projection-level max_avg value -- slowly decays and instantly resets to any current max"`
	DecayTau float32 `condshow:"On=true" min:"1" def:"1000;10000" desc:"time constant for decay of dwnorm factor -- generally should be long-ish, between 1000-10000 -- integration rate factor is 1/tau"`
	NormMin  float32 `condshow:"On=true" min:"0" def:"0.001" desc:"minimum effective value of the normalization factor -- provides a lower bound to how much normalization can be applied"`
	LrComp   float32 `condshow:"On=true" min:"0" def:"0.15" desc:"overall learning rate multiplier to compensate for changes due to use of normalization -- allows for a common master learning rate to be used between different conditions -- 0.1 for synapse-level, maybe higher for other levels"`
	Stats    bool    `condshow:"On=true" def:"false" desc:"record the avg, max values of err, bcm hebbian, and overall dwt change per con group and per projection"`

	DecayDt  float32 `inactive:"+" view:"-" desc:"rate constant of decay = 1 / decay_tau"`
	DecayDtC float32 `inactive:"+" view:"-" desc:"complement rate constant of decay = 1 - (1 / decay_tau)"`
}

// DwtNormPars updates the dwnorm running max_abs, slowly decaying value
// jumps up to max(abs_dwt) and slowly decays
// returns the effective normalization factor, as a multiplier, including lrate comp
func (dn *DwtNormPars) NormFmAbsDwt(dwnorm, absDwt float32) float32 {
	dwnorm = math32.Max(dn.DecayDtC*dwnorm, absDwt)
	if dwnorm == 0 {
		return 1
	}
	norm := math32.Max(dwnorm, dn.NormMin)
	return dn.LrComp / norm
}

func (dn *DwtNormPars) Update() {
	dn.DecayDt = 1 / dn.DecayTau
	dn.DecayDtC = 1 - dn.DecayDt
}

func (dn *DwtNormPars) Defaults() {
	dn.On = true
	dn.DecayTau = 1000
	dn.LrComp = 0.15
	dn.NormMin = 0.001
	dn.Stats = false
	UpdtVals()
}

//////////////////////////////////////////////////////////////////////////////////////
//  MomentumPars

// MomentumPars implements standard simple momentum -- accentuates consistent directions of weight change and
// cancels out dithering -- biologically captures slower timecourse of longer-term plasticity mechanisms.
type MomentumPars struct {
	On     bool    `def:"true" desc:"whether to use standard simple momentum"`
	MTau   float32 `condshow:"On=true" min:"1" def:"10" desc:"time constant factor for integration of momentum -- 1/tau is dt (e.g., .1), and 1-1/tau (e.g., .95 or .9) is traditional momentum time-integration factor"`
	LrComp float32 `condshow:"On=true" min:"0" def:"0.1" desc:"overall learning rate multiplier to compensate for changes due to JUST momentum without normalization -- allows for a common master learning rate to be used between different conditions -- generally should use .1 to compensate for just momentum itself"`

	MDt  float32 `inactive:"+" view:"-" desc:"rate constant of momentum integration = 1 / m_tau"`
	MDtC float32 `inactive:"+" view:"-" desc:"complement rate constant of momentum integration = 1 - (1 / m_tau)"`
}

// MomentFmDt compute momentum from weight change value
func (mp *MomentumPars) MomentFmDwt(moment, dwt float32) float32 {
	moment = mp.MDtC*moment + dwt
	return moment
}

func (mp *MomentumPars) Update() {
	mp.MDt = 1 / mp.MTau
	mp.MDtC = 1 - mp.MDt
}

func (mp *MomentumPars) Defaults() {
	mp.On = true
	mp.MTau = 10
	mp.LrComp = 0.1
	mp.Update()
}

//////////////////////////////////////////////////////////////////////////////////////
//  WtBalPars

// WtBalPars are weight balance soft renormalization params:
// maintains overall weight balance by progressively penalizing weight increases as a function of
// how strong the weights are overall (subject to thresholding) and long time-averaged activation.
// Plugs into soft bounding function.
type WtBalPars struct {
	On      bool    `desc:"perform weight balance soft normalization?  if so, maintains overall weight balance across units by progressively penalizing weight increases as a function of amount of averaged weight above a high threshold (hi_thr) and long time-average activation above an act_thr -- this is generally very beneficial for larger models where hog units are a problem, but not as much for smaller models where the additional constraints are not beneficial -- uses a sigmoidal function: wb_inc = 1 / (1 + hi_gain*(wb_avg - hi_thr) + act_gain * (act_avg - act_thr)))"`
	AvgThr  float32 `condshow:"On=true" def:"0.25" desc:"threshold on weight value for inclusion into the weight average that is then subject to the further hi_thr threshold for then driving a change in weight balance -- this avg_thr allows only stronger weights to contribute so that weakening of lower weights does not dilute sensitivity to number and strength of strong weights"`
	HiThr   float32 `condshow:"On=true" def:"0.4" desc:"high threshold on weight average (subject to avg_thr) before it drives changes in weight increase vs. decrease factors"`
	HiGain  float32 `condshow:"On=true"def:"4" desc:"gain multiplier applied to above-hi_thr thresholded weight averages -- higher values turn weight increases down more rapidly as the weights become more imbalanced"`
	LoThr   float32 `condshow:"On=true" def:"0.4" desc:"low threshold on weight average (subject to avg_thr) before it drives changes in weight increase vs. decrease factors"`
	LoGain  float32 `condshow:"On=true" def:"6;0" desc:"gain multiplier applied to below-lo_thr thresholded weight averages -- higher values turn weight increases up more rapidly as the weights become more imbalanced -- generally beneficial but sometimes not -- worth experimenting with either 6 or 0"`
	ActThr  float32 `condshow:"On=true" def:"0.25" desc:"threshold for long time-average activation (act_avg) contribution to weight balance -- based on act_avg relative to act_thr -- same statistic that we use to measure hogging with default .3 threshold"`
	ActGain float32 `condshow:"On=true" def:"0;2" desc:"gain multiplier applied to above-threshold weight averages -- higher values turn weight increases down more rapidly as the weights become more imbalanced -- see act_thr for equation"`
	NoTarg  bool    `condshow:"On=true" def:"true" desc:"exclude receiving projections into TARGET layers where units are clamped and also TRC (Pulvinar) thalamic neurons -- typically for clamped layers you do not want to be applying extra constraints such as this weight balancing dynamic -- the BCM hebbian learning is also automatically turned off for such layers as well"`
}

// WtBal computes weight balance factors for increase and decrease based on extent
// to which weights and average act exceed thresholds
func (wb *WtBalPars) WtBal(wbAvg, actAvg float32) (wbFact, wbInc, wbDec float32) {
	wbInc = 1
	wbDec = 1
	if wbAvg < wb.LoThr {
		if wbAvg < wb.AvgThr {
			wbAvg = wb.AvgThr // prevent extreme low if everyone below thr
		}
		wbFact = wb.LoGain * (wb.LoThr - wbAvg)
		wbDec = 1 / (1 + wbFact)
		wbInc = 2 - wbDec
	} else if wbAvg > wb.HiThr {
		wbFact += wb.HiGain * (wbAvg - wb.HiThr)
		if actAvg > wb.ActThr {
			wbFact += wb.ActGain * (actAvg - wb.ActThr)
		}
		wbInc = 1 / (1 + wbFact) // gets sigmoidally small toward 0 as wbFact gets larger -- is quick acting but saturates -- apply pressure earlier..
		wbDec = 2 - wbInc        // as wb_inc goes down, wb_dec goes up..  sum to 2
	}
	return wbFact, wbInc, wbDec
}

func (wb *WtBalPars) Defaults() {
	wb.On = true
	wb.NoTarg = true
	wb.AvgThr = 0.25
	wb.HiThr = 0.4
	wb.HiGain = 4
	wb.LoThr = 0.4
	wb.LoGain = 6
	wb.ActThr = 0.25
	wb.ActGain = 0
}

/*

class STATE_CLASS(AdaptWtScaleSpec) : public STATE_CLASS(SpecMemberBase) {
  // ##INLINE ##NO_TOKENS ##CAT_Leabra parameters to adapt the scale multiplier on weights, as a function of weight value
INHERITED(SpecMemberBase)
public:
  bool          on;             // turn on weight scale adaptation as function of weight values
        tau;            // #CONDSHOW_ON_on def:"5000 time constant as a function of weight updates (trials) that weight scale adapts on -- should be fairly slow in general
        lo_thr;         // #CONDSHOW_ON_on def:"0.25 low threshold:  normalized contrast-enhanced effective weights (wt/scale, 0-1 range) below this value cause scale to move downward toward lo_scale value
        hi_thr;         // #CONDSHOW_ON_on def:"0.75 high threshold: normalized contrast-enhanced effective weights (wt/scale, 0-1 range) above this value cause scale to move upward toward hi_scale value
        lo_scale;       // #CONDSHOW_ON_on min:"0.01 def:"0.01 lowest value of scale
        hi_scale;       // #CONDSHOW_ON_on def:"2 highest value of scale

        dt;             // #READ_ONLY #EXPERT rate = 1 / tau

  INLINE void   AdaptWtScale(float& scale, const float wt) {
    const float nrm_wt = wt / scale;
    if(nrm_wt < lo_thr) {
      scale += dt * (lo_scale - scale);
    }
    else if(nrm_wt > hi_thr) {
      scale += dt * (hi_scale - scale);
    }
  }
  // adapt weight scale

  STATE_DECO_KEY("ConSpec");
  STATE_TA_STD_CODE_SPEC(AdaptWtScaleSpec);
  STATE_UAE( dt = 1.0f / tau; );
private:
  void  Initialize()     {   on = false;  Defaults_init(); }
  void  Defaults_init() {
    tau = 5000.0f;  lo_thr = 0.25f;  hi_thr = 0.75f;  lo_scale = 0.01f;  hi_scale = 2.0f;
    dt = 1.0f / tau;
  }
};

*/
