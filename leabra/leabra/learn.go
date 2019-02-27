// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import (
	"github.com/chewxy/math32"
	"github.com/emer/emergent/erand"
)

///////////////////////////////////////////////////////////////////////
//  learn.go contains the learning params and functions for leabra

// leabra.LearnNeurParams manages learning-related parameters at the neuron-level.
// This is mainly the running average activations that drive learning
type LearnNeurParams struct {
	ActAvg  LrnActAvgParams `inline:"+" desc:"parameters for computing running average activations that drive learning"`
	AvgL    AvgLParams      `inline:"+" desc:"parameters for computing AvgL long-term running average"`
	CosDiff CosDiffParams   `inline:"+" desc:"parameters for computing cosine diff between minus and plus phase"`
}

func (ln *LearnNeurParams) Update() {
	ln.ActAvg.Update()
	ln.AvgL.Update()
	ln.CosDiff.Update()
}

func (ln *LearnNeurParams) Defaults() {
	ln.ActAvg.Defaults()
	ln.AvgL.Defaults()
	ln.CosDiff.Defaults()
}

// InitActAvg initializes the running-average activation values that drive learning.
// Called by InitWts (at start of learning).
func (ln *LearnNeurParams) InitActAvg(nrn *Neuron) {
	nrn.AvgSS = ln.ActAvg.Init
	nrn.AvgS = ln.ActAvg.Init
	nrn.AvgM = ln.ActAvg.Init
	nrn.AvgL = ln.AvgL.Init
	nrn.AvgSLrn = 0
}

// AvgsFmAct updates the running averages based on current activation.
// Computed after new activation for current cycle is updated.
func (ln *LearnNeurParams) AvgsFmAct(nrn *Neuron) {
	ln.ActAvg.AvgsFmAct(nrn.Act, &nrn.AvgSS, &nrn.AvgS, &nrn.AvgM, &nrn.AvgSLrn)
}

// AvgLFmAct computes long-term average activation value, and learning factor, from current AvgM.
// Called at start of new trial.
func (ln *LearnNeurParams) AvgLFmAvgM(nrn *Neuron) {
	ln.AvgL.AvgLFmAvgM(nrn.AvgM, &nrn.AvgL, &nrn.AvgLLrn)
}

///////////////////////////////////////////////////////////////////////
//  LearnSynParams

// todo: lrate sched

// leabra.LearnSynParams manages learning-related parameters at the synapse-level.
type LearnSynParams struct {
	Learn    bool            `desc:"enable learning for this projection"`
	Lrate    float32         `desc:"learning rate"`
	WtInit   erand.RndParams `inline:"+" desc:"initial random weight distribution"`
	XCal     XCalParams      `inline:"+" desc:"parameters for the XCal learning rule"`
	WtSig    WtSigParams     `inline:"+" desc:"parameters for the sigmoidal contrast weight enhancement"`
	Norm     DWtNormParams   `inline:"+" desc:"parameters for normalizing weight changes by abs max dwt"`
	Momentum MomentumParams  `inline:"+" desc:"parameters for momentum across weight changes"`
	WtBal    WtBalParams     `inline:"+" desc:"parameters for balancing strength of weight increases vs. decreases"`
}

func (ls *LearnSynParams) Update() {
	ls.XCal.Update()
	ls.WtSig.Update()
	ls.Norm.Update()
	ls.Momentum.Update()
	ls.WtBal.Update()
}

func (ls *LearnSynParams) Defaults() {
	ls.Learn = true
	ls.Lrate = 0.04
	ls.WtInit.Mean = 0.5
	ls.WtInit.Var = 0.25
	ls.WtInit.Dist = erand.Uniform
	ls.XCal.Defaults()
	ls.WtSig.Defaults()
	ls.Norm.Defaults()
	ls.Momentum.Defaults()
	ls.WtBal.Defaults()
}

// InitWts initializes weight values based on WtInit randomness parameters
// It also updates the linear weight value based on the sigmoidal weight value
func (ls *LearnSynParams) InitWts(syn *Synapse) {
	syn.Wt = float32(ls.WtInit.Gen(-1))
	syn.LWt = ls.WtSig.LinFmSigWt(syn.Wt)
	syn.DWt = 0
	syn.Norm = 0
	syn.Moment = 0
}

// LWtFmWt updates the linear weight value based on the current effective Wt value.
// effective weight is sigmoidally contrast-enhanced relative to the linear weight.
func (ls *LearnSynParams) LWtFmWt(syn *Synapse) {
	syn.LWt = ls.WtSig.LinFmSigWt(syn.Wt)
}

// WtFmLWt updates the effective weight value based on the current linear Wt value.
// effective weight is sigmoidally contrast-enhanced relative to the linear weight.
func (ls *LearnSynParams) WtFmLWt(syn *Synapse) {
	syn.Wt = ls.WtSig.SigFmLinWt(syn.LWt)
}

// CHLdWt returns the error-driven and bcm Hebbian weight change components for the
// temporally eXtended Contrastive Attractor Learning (XCAL), CHL version
func (ls *LearnSynParams) CHLdWt(suAvgSLrn, suAvgM, ruAvgSLrn, ruAvgM, ruAvgL float32) (err, bcm float32) {
	srs := suAvgSLrn * ruAvgSLrn
	srm := suAvgM * ruAvgM
	bcm = ls.XCal.DWt(srs, ruAvgL)
	err = ls.XCal.DWt(srs, srm)
	return
}

// WtFmDWt updates the synaptic weights from accumulated weight changes
// wbInc and wbDec are the weight balance factors, wt is the sigmoidal contrast-enhanced
// weight and lwt is the linear weight value
func (ls *LearnSynParams) WtFmDWt(wbInc, wbDec float32, dwt, wt, lwt *float32) {
	if *dwt == 0 {
		return
	}
	if ls.WtSig.SoftBound {
		if *dwt > 0 {
			*dwt *= wbInc * (1 - *lwt)
		} else {
			*dwt *= wbDec * *lwt
		}
	} else {
		if *dwt > 0 {
			*dwt *= wbInc
		} else {
			*dwt *= wbDec
		}
	}
	*lwt += *dwt
	if *lwt < 0 {
		*lwt = 0
	} else if *lwt > 1 {
		*lwt = 1
	}
	*wt = ls.WtSig.SigFmLinWt(*lwt) //  todo: scale *
	*dwt = 0
	//    if(adapt_scale.on) {
	//      adapt_scale.AdaptWtScale(scale, wt);
	//    }
}

// LrnActAvgParams has rate constants for averaging over activations at different time scales,
// to produce the running average activation values that then drive learning in the XCAL learning rules
type LrnActAvgParams struct {
	SSTau float32 `def:"2;4;7"  min:"1" desc:"time constant in cycles, which should be milliseconds typically (roughly, how long it takes for value to change significantly -- 1.4x the half-life), for continuously updating the super-short time-scale avg_ss value -- this is provides a pre-integration step before integrating into the avg_s short time scale -- it is particularly important for spiking -- in general 4 is the largest value without starting to impair learning, but a value of 7 can be combined with m_in_s = 0 with somewhat worse results"`
	STau  float32 `def:"2" min:"1" desc:"time constant in cycles, which should be milliseconds typically (roughly, how long it takes for value to change significantly -- 1.4x the half-life), for continuously updating the short time-scale avg_s value from the super-short avg_ss value (cascade mode) -- avg_s represents the plus phase learning signal that reflects the most recent past information"`
	MTau  float32 `def:"10" min:"1" desc:"time constant in cycles, which should be milliseconds typically (roughly, how long it takes for value to change significantly -- 1.4x the half-life), for continuously updating the medium time-scale avg_m value from the short avg_s value (cascade mode) -- avg_m represents the minus phase learning signal that reflects the expectation representation prior to experiencing the outcome (in addition to the outcome) -- the default value of 10 generally cannot be exceeded without impairing learning"`
	LrnM  float32 `def:"0.1;0" min:"0" max:"1" desc:"how much of the medium term average activation to mix in with the short (plus phase) to compute the Neuron AvgSLrn variable that is used for the unit's short-term average in learning. This is important to ensure that when unit turns off in plus phase (short time scale), enough medium-phase trace remains so that learning signal doesn't just go all the way to 0, at which point no learning would take place -- typically need faster time constant for updating S such that this trace of the M signal is lost -- can set SSTau=7 and set this to 0 but learning is generally somewhat worse"`
	Init  float32 `def:"0.15" min:"0" max:"1" desc:"initial value for average"`

	SSDt float32 `view:"-" inactive:"+" desc:"rate = 1 / tau"`
	SDt  float32 `view:"-" inactive:"+" desc:"rate = 1 / tau"`
	MDt  float32 `view:"-" inactive:"+" desc:"rate = 1 / tau"`
	LrnS float32 `view:"-" inactive:"+" desc:"1-LrnM"`
}

// AvgsFmAct computes averages based on current act
func (aa *LrnActAvgParams) AvgsFmAct(ruAct float32, avgSS, avgS, avgM, avgSLrn *float32) {
	*avgSS += aa.SSDt * (ruAct - *avgSS)
	*avgS += aa.SDt * (*avgSS - *avgS)
	*avgM += aa.MDt * (*avgS - *avgM)

	*avgSLrn = aa.LrnS**avgS + aa.LrnM**avgM
}

func (aa *LrnActAvgParams) Update() {
	aa.SSDt = 1 / aa.SSTau
	aa.SDt = 1 / aa.STau
	aa.MDt = 1 / aa.MTau
	aa.LrnS = 1 - aa.LrnM
}

func (aa *LrnActAvgParams) Defaults() {
	aa.SSTau = 2.0
	aa.STau = 2.0
	aa.MTau = 10.0
	aa.LrnM = 0.1
	aa.Init = 0.15
	aa.Update()

}

// AvgLParams are parameters for computing the long-term floating average value, AvgL
// which is used for driving BCM-style hebbian learning in XCAL -- this form of learning
// increases contrast of weights and generally decreases overall activity of neuron,
// to prevent "hog" units -- it is computed as a running average of the (gain multiplied)
// medium-time-scale average activation at the end of the trial.
// Also computes an adaptive amount of BCM learning, AvgLLrn, based on AvgL.
type AvgLParams struct {
	Init   float32 `def:"0.4" min:"0" max:"1" desc:"initial AvgL value at start of training"`
	Gain   float32 `def:"1.5;2;2.5;3;4;5" min:"0" desc:"gain multiplier on activation used in computing the running average AvgL value that is the key floating threshold in the BCM Hebbian learning rule -- when using the DELTA_FF_FB learning rule, it should generally be 2x what it was before with the old XCAL_CHL rule, i.e., default of 5 instead of 2.5 -- it is a good idea to experiment with this parameter a bit -- the default is on the high-side, so typically reducing a bit from initial default is a good direction"`
	Min    float32 `def:"0.2" min:"0" desc:"miniumum AvgL value -- running average cannot go lower than this value even when it otherwise would due to inactivity -- default value is generally good and typically does not need to be changed"`
	Tau    float32 `def:"10" min:"1" desc:"time constant for updating the running average AvgL -- AvgL moves toward gain*act with this time constant on every trial - longer time constants can also work fine, but the default of 10 allows for quicker reaction to beneficial weight changes"`
	LrnMax float32 `def:"0.5" min:"0" desc:"maximum AvgLLrn value, which is amount of learning driven by AvgL factor -- when AvgL is at its maximum value (i.e., gain, as act does not exceed 1), then AvgLLrn will be at this maximum value -- by default, strong amounts of this homeostatic Hebbian form of learning can be used when the receiving unit is highly active -- this will then tend to bring down the average activity of units -- the default of 0.5, in combination with the err_mod flag, works well for most models -- use around 0.0004 for a single fixed value (with err_mod flag off)"`
	LrnMin float32 `def:"0.0001;0.0004" min:"0" desc:"miniumum AvgLLrn value (amount of learning driven by AvgL factor) -- if AvgL is at its minimum value, then AvgLLrn will be at this minimum value -- neurons that are not overly active may not need to increase the contrast of their weights as much -- use around 0.0004 for a single fixed value (with err_mod flag off)"`
	ErrMod bool    `def:"true" desc:"modulate amount learning by normalized level of error within layer"`
	ModMin float32 `def:"0.01" view:"if ErrMod=true" desc:"minimum modulation value for ErrMod-- ensures a minimum amount of self-organizing learning even for network / layers that have a very small level of error signal"`

	Dt      float32 `view:"-" inactive:"+" desc:"rate = 1 / tau"`
	LrnFact float32 `view:"-" inactive:"+" desc:"(LrnMax - LrnMin) / (Gain - Min)"`
}

// AvgLFmAvgM computes long-term average activation value, and learning factor, from given
// medium-scale running average activation avgM
func (al *AvgLParams) AvgLFmAvgM(avgM float32, avgL, lrn *float32) {
	*avgL += al.Dt * (al.Gain*avgM - *avgL)
	if *avgL < al.Min {
		*avgL = al.Min
	}
	*lrn = al.LrnFact * (*avgL - al.Min)
}

// ErrModFmLayErr computes AvgLLrn multiplier from layer cosine diff avg statistic
func (al *AvgLParams) ErrModFmLayErr(layCosDiffAvg float32) float32 {
	mod := float32(1)
	if !al.ErrMod {
		return mod
	}
	mod *= math32.Max(layCosDiffAvg, al.ModMin)
	return mod
}

func (al *AvgLParams) Update() {
	al.Dt = 1 / al.Tau
	al.LrnFact = (al.LrnMax - al.LrnMin) / (al.Gain - al.Min)
}

func (al *AvgLParams) Defaults() {
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
//  CosDiffParams

// CosDiffParams specify how to integrate cosine of difference between plus and minus phase activations
// Used to modulate amount of hebbian learning, and overall learning rate.
type CosDiffParams struct {
	Tau float32 `def:"100" min:"1" desc:"time constant in trials (roughly how long significant change takes, 1.4 x half-life) for computing running average CosDiff value for the layer, CosDiffAvg = cosine difference between ActM and ActP -- this is an important statistic for how much phase-based difference there is between phases in this layer -- it is used in standard X_COS_DIFF modulation of l_mix in LeabraConSpec, and for modulating learning rate as a function of predictability in the DeepLeabra predictive auto-encoder learning -- running average variance also computed with this: cos_diff_var"`
	//   bool          lrate_mod; // modulate learning rate in this layer as a function of the cos_diff on this trial relative to running average cos_diff values (see avg_tau) -- lrate_mod = cos_diff_lrate_mult * (cos_diff / cos_diff_avg) -- if this layer is less predictable than previous trials, we don't learn as much
	//   float         lrmod_z_thr; // #DEF_-1.5 #CONDSHOW_ON_lrate_mod&&!lrmod_fm_trc threshold for setting learning rate modulation to zero, as function of z-normalized cos_diff value on this trial -- normalization computed using incrementally computed average and variance values -- this essentially has the network ignoring trials where the diff was significantly below average -- replaces the manual unlearnable trial mechanism
	//   bool          set_net_unlrn;  // #CONDSHOW_ON_lrate_mod&&!lrmod_fm_trc set the network-level unlearnable_trial flag based on our learning rate modulation factor -- only makes sense for one layer to do this

	Dt  float32 `inactive:"+" view:"-" desc:"rate constant = 1 / Tau"`
	DtC float32 `inactive:"+" view:"-" desc:"complement of rate constant = 1 - Dt"`
}

func (cd *CosDiffParams) Update() {
	cd.Dt = 1 / cd.Tau
	cd.DtC = 1 - cd.Dt
}

func (cd *CosDiffParams) Defaults() {
	cd.Tau = 100
	cd.Update()
}

// AvgVarFmCos updates the average and variance from current cosine diff value
func (cd *CosDiffParams) AvgVarFmCos(avg, vr *float32, cos float32) {
	if *avg == 0 { // first time -- set
		*avg = cos
		*vr = 0
	} else {
		del := cos - *avg
		incr := cd.Dt * del
		*avg += incr
		// following is magic exponentially-weighted incremental variance formula
		// derived by Finch, 2009: Incremental calculation of weighted mean and variance
		if *vr == 0 {
			*vr = 2 * cd.DtC * del * incr
		} else {
			*vr = cd.DtC * (*vr + del*incr)
		}
	}
}

// LrateMod computes learning rate modulation based on cos diff vals
// func (cd *CosDiffParams) LrateMod(cos, avg, vr float32) float32 {
// 	if vr <= 0 {
// 		return 1
// 	}
// 	zval := (cos - avg) / math32.Sqrt(vr) // stdev = sqrt of var
// 	// z-normal value is starting point for learning rate factor
// 	//    if zval < lrmod_z_thr {
// 	// 	return 0
// 	// }
// 	return 1
// }

//////////////////////////////////////////////////////////////////////////////////////
//  CosDiffStats

// CosDiffStats holds cosine-difference statistics at the layer level
type CosDiffStats struct {
	Cos        float32 `desc:"cosine (normalized dot product) activation difference between ActP and ActM on this trial for this layer -- computed by CosDiffFmActs at end of QuarterFinal for quarter = 3"`
	Avg        float32 `desc:"running average of cosine (normalized dot product) difference between ActP and ActM -- computed with CosDiff.Tau time constant in QuarterFinal, and used for modulating BCM Hebbian learning (see AvgLrn) and overall learning rate"`
	Var        float32 `desc:"running variance of cosine (normalized dot product) difference between ActP and ActM -- computed with CosDiff.Tau time constant in QuarterFinal, used for modulating overall learning rate"`
	AvgLrn     float32 `desc:"1 - Avg and 0 for non-Hidden layers"`
	ModAvgLLrn float32 `desc:"1 - AvgLrn and 0 for non-Hidden layers -- this is the value of Avg used for AvgLParams ErrMod modulation of the AvgLLrn factor if enabled"`
}

func (cd *CosDiffStats) Init() {
	cd.Cos = 0
	cd.Avg = 0
	cd.Var = 0
	cd.AvgLrn = 0
	cd.ModAvgLLrn = 0
}

//////////////////////////////////////////////////////////////////////////////////////
//  XCalParams

// XCalParams are parameters for temporally eXtended Contrastive Attractor Learning function (XCAL)
// which is the standard learning equation for leabra .
type XCalParams struct {
	MLrn    float32 `def:"1" min:"0" desc:"multiplier on learning based on the medium-term floating average threshold which produces error-driven learning -- this is typically 1 when error-driven learning is being used, and 0 when pure Hebbian learning is used. The long-term floating average threshold is provided by the receiving unit"`
	SetLLrn bool    `def:"false" desc:"if true, set a fixed AvgLLrn weighting factor that determines how much of the long-term floating average threshold (i.e., BCM, Hebbian) component of learning is used -- this is useful for setting a fully Hebbian learning connection, e.g., by setting MLrn = 0 and LLrn = 1. If false, then the receiving unit's AvgLLrn factor is used, which dynamically modulates the amount of the long-term component as a function of how active overall it is"`
	LLrn    float32 `view:"if SetLLrn=true" desc:"fixed l_lrn weighting factor that determines how much of the long-term floating average threshold (i.e., BCM, Hebbian) component of learning is used -- this is useful for setting a fully Hebbian learning connection, e.g., by setting MLrn = 0 and LLrn = 1."`
	DRev    float32 `def:"0.1" min:"0" max:"0.99" desc:"proportional point within LTD range where magnitude reverses to go back down to zero at zero -- err-driven svm component does better with smaller values, and BCM-like mvl component does better with larger values -- 0.1 is a compromise"`
	DThr    float32 `def:"0.0001;0.01" min:"0" desc:"minimum LTD threshold value below which no weight change occurs -- this is now *relative* to the threshold"`
	LrnThr  float32 `def:"0.01" desc:"xcal learning threshold -- don't learn when sending unit activation is below this value in both phases -- due to the nature of the learning function being 0 when the sr coproduct is 0, it should not affect learning in any substantial way -- nonstandard learning algorithms that have different properties should ignore it"`

	DRevRatio float32 `inactive:"+" view:"-" desc:"-(1-DRev)/DRev -- multiplication factor in learning rule -- builds in the minus sign!"`
}

func (xc *XCalParams) Update() {
	if xc.DRev > 0 {
		xc.DRevRatio = -(1 - xc.DRev) / xc.DRev
	} else {
		xc.DRevRatio = -1
	}
}

func (xc *XCalParams) Defaults() {
	xc.MLrn = 1
	xc.SetLLrn = false
	xc.LLrn = 1
	xc.DRev = 0.1
	xc.DThr = 0.0001
	xc.LrnThr = 0.01
	xc.Update()
}

// DWt is the XCAL function for weight change -- the "check mark" function -- no DGain, no ThrPMin
func (xc *XCalParams) DWt(srval, thrP float32) float32 {
	var dwt float32
	if srval < xc.DThr {
		dwt = 0
	} else if srval > thrP*xc.DRev {
		dwt = (srval - thrP)
	} else {
		dwt = srval * xc.DRevRatio
	}
	return dwt
}

// LongLrate returns the learning rate for long-term floating average component (BCM)
func (xc *XCalParams) LongLrate(avgLLrn float32) float32 {
	if xc.SetLLrn {
		return xc.LLrn
	}
	return avgLLrn
}

//////////////////////////////////////////////////////////////////////////////////////
//  WtSigParams

// WtSigParams are sigmoidal weight contrast enhancement function parameters
type WtSigParams struct {
	Gain      float32 `def:"1;6" min:"0" desc:"gain (contrast, sharpness) of the weight contrast function (1 = linear)"`
	Off       float32 `def:"1" min:"0" desc:"offset of the function (1=centered at .5, >1=higher, <1=lower) -- 1 is standard for XCAL"`
	SoftBound bool    `def:"true" desc:"apply exponential soft bounding to the weight changes"`
}

func (ws *WtSigParams) Update() {
}

func (ws *WtSigParams) Defaults() {
	ws.Gain = 6
	ws.Off = 1
	ws.SoftBound = true
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
	return 1.0 / (1.0 + math32.Pow((1.0-w)/w, 1/gain)/off)
}

// SigInvFun61 is the inverse of the sigmoid function, with default gain = 6, offset = 1 params
func SigInvFun61(w float32) float32 {
	if w <= 0 {
		return 0
	}
	if w >= 1 {
		return 1
	}
	rval := 1.0 / (1.0 + math32.Pow((1.0-w)/w, 1.0/6.0))
	return rval
}

// SigFmLinWt returns sigmoidal contrast-enhanced weight from linear weight
func (ws *WtSigParams) SigFmLinWt(lw float32) float32 {
	if ws.Gain == 1 && ws.Off == 1 {
		return lw
	}
	if ws.Gain == 6 && ws.Off == 1 {
		return SigFun61(lw)
	}
	return SigFun(lw, ws.Gain, ws.Off)
}

// LinFmSigWt returns linear weight from sigmoidal contrast-enhanced weight
func (ws *WtSigParams) LinFmSigWt(sw float32) float32 {
	if ws.Gain == 1 && ws.Off == 1 {
		return sw
	}
	if ws.Gain == 6 && ws.Off == 1 {
		return SigInvFun61(sw)
	}
	return SigInvFun(sw, ws.Gain, ws.Off)
}

//////////////////////////////////////////////////////////////////////////////////////
//  DWtNormParams

// DWtNormParams are weight change (dwt) normalization parameters, using MAX(ABS(dwt)) aggregated over
// Sending connections in a given projection for a given unit.
// Slowly decays and instantly resets to any current max(abs)
// Serves as an estimate of the variance in the weight changes, assuming zero net mean overall.
type DWtNormParams struct {
	On       bool    `def:"true" desc:"whether to use dwt normalization, only on error-driven dwt component, based on projection-level max_avg value -- slowly decays and instantly resets to any current max"`
	DecayTau float32 `view:"if On=true" min:"1" def:"1000;10000" desc:"time constant for decay of dwnorm factor -- generally should be long-ish, between 1000-10000 -- integration rate factor is 1/tau"`
	NormMin  float32 `view:"if On=true" min:"0" def:"0.001" desc:"minimum effective value of the normalization factor -- provides a lower bound to how much normalization can be applied"`
	LrComp   float32 `view:"if On=true" min:"0" def:"0.15" desc:"overall learning rate multiplier to compensate for changes due to use of normalization -- allows for a common master learning rate to be used between different conditions -- 0.1 for synapse-level, maybe higher for other levels"`
	Stats    bool    `view:"if On=true" def:"false" desc:"record the avg, max values of err, bcm hebbian, and overall dwt change per con group and per projection"`

	DecayDt  float32 `inactive:"+" view:"-" desc:"rate constant of decay = 1 / decay_tau"`
	DecayDtC float32 `inactive:"+" view:"-" desc:"complement rate constant of decay = 1 - (1 / decay_tau)"`
}

// DWtNormParams updates the dwnorm running max_abs, slowly decaying value
// jumps up to max(abs_dwt) and slowly decays
// returns the effective normalization factor, as a multiplier, including lrate comp
func (dn *DWtNormParams) NormFmAbsDWt(norm *float32, absDwt float32) float32 {
	*norm = math32.Max(dn.DecayDtC**norm, absDwt)
	if *norm == 0 {
		return 1
	}
	return dn.LrComp / math32.Max(*norm, dn.NormMin)
}

func (dn *DWtNormParams) Update() {
	dn.DecayDt = 1 / dn.DecayTau
	dn.DecayDtC = 1 - dn.DecayDt
}

func (dn *DWtNormParams) Defaults() {
	dn.On = true
	dn.DecayTau = 1000
	dn.LrComp = 0.15
	dn.NormMin = 0.001
	dn.Stats = false
	dn.Update()
}

//////////////////////////////////////////////////////////////////////////////////////
//  MomentumParams

// MomentumParams implements standard simple momentum -- accentuates consistent directions of weight change and
// cancels out dithering -- biologically captures slower timecourse of longer-term plasticity mechanisms.
type MomentumParams struct {
	On     bool    `def:"true" desc:"whether to use standard simple momentum"`
	MTau   float32 `view:"if On=true" min:"1" def:"10" desc:"time constant factor for integration of momentum -- 1/tau is dt (e.g., .1), and 1-1/tau (e.g., .95 or .9) is traditional momentum time-integration factor"`
	LrComp float32 `view:"if On=true" min:"0" def:"0.1" desc:"overall learning rate multiplier to compensate for changes due to JUST momentum without normalization -- allows for a common master learning rate to be used between different conditions -- generally should use .1 to compensate for just momentum itself"`

	MDt  float32 `inactive:"+" view:"-" desc:"rate constant of momentum integration = 1 / m_tau"`
	MDtC float32 `inactive:"+" view:"-" desc:"complement rate constant of momentum integration = 1 - (1 / m_tau)"`
}

// MomentFmDWt updates synaptic moment variable based on dwt weight change value
// and returns new momentum factor * LrComp
func (mp *MomentumParams) MomentFmDWt(moment *float32, dwt float32) float32 {
	*moment = mp.MDtC**moment + dwt
	return mp.LrComp * *moment
}

func (mp *MomentumParams) Update() {
	mp.MDt = 1 / mp.MTau
	mp.MDtC = 1 - mp.MDt
}

func (mp *MomentumParams) Defaults() {
	mp.On = true
	mp.MTau = 10
	mp.LrComp = 0.1
	mp.Update()
}

//////////////////////////////////////////////////////////////////////////////////////
//  WtBalParams

// WtBalParams are weight balance soft renormalization params:
// maintains overall weight balance by progressively penalizing weight increases as a function of
// how strong the weights are overall (subject to thresholding) and long time-averaged activation.
// Plugs into soft bounding function.
type WtBalParams struct {
	On      bool    `desc:"perform weight balance soft normalization?  if so, maintains overall weight balance across units by progressively penalizing weight increases as a function of amount of averaged receiver weight above a high threshold (hi_thr) and long time-average activation above an act_thr -- this is generally very beneficial for larger models where hog units are a problem, but not as much for smaller models where the additional constraints are not beneficial -- uses a sigmoidal function: WbInc = 1 / (1 + HiGain*(WbAvg - HiThr) + ActGain * (nrn.ActAvg - ActThr)))"`
	AvgThr  float32 `view:"if On=true" def:"0.25" desc:"threshold on weight value for inclusion into the weight average that is then subject to the further HiThr threshold for then driving a change in weight balance -- this AvgThr allows only stronger weights to contribute so that weakening of lower weights does not dilute sensitivity to number and strength of strong weights"`
	HiThr   float32 `view:"if On=true" def:"0.4" desc:"high threshold on weight average (subject to AvgThr) before it drives changes in weight increase vs. decrease factors"`
	HiGain  float32 `view:"if On=true" def:"4" desc:"gain multiplier applied to above-HiThr thresholded weight averages -- higher values turn weight increases down more rapidly as the weights become more imbalanced"`
	LoThr   float32 `view:"if On=true" def:"0.4" desc:"low threshold on weight average (subject to AvgThr) before it drives changes in weight increase vs. decrease factors"`
	LoGain  float32 `view:"if On=true" def:"6;0" desc:"gain multiplier applied to below-lo_thr thresholded weight averages -- higher values turn weight increases up more rapidly as the weights become more imbalanced -- generally beneficial but sometimes not -- worth experimenting with either 6 or 0"`
	ActThr  float32 `view:"if On=true" def:"0.25" desc:"threshold for long time-average activation (act_avg) contribution to weight balance -- based on act_avg relative to act_thr -- same statistic that we use to measure hogging with default .3 threshold"`
	ActGain float32 `view:"if On=true" def:"0;2" desc:"gain multiplier applied to above-threshold weight averages -- higher values turn weight increases down more rapidly as the weights become more imbalanced -- see act_thr for equation"`
}

func (wb *WtBalParams) Update() {
}

func (wb *WtBalParams) Defaults() {
	wb.On = false
	wb.AvgThr = 0.25
	wb.HiThr = 0.4
	wb.HiGain = 4
	wb.LoThr = 0.4
	wb.LoGain = 6
	wb.ActThr = 0.25
	wb.ActGain = 0
}

// WtBal computes weight balance factors for increase and decrease based on extent
// to which weights and average act exceed thresholds
func (wb *WtBalParams) WtBal(wbAvg, actAvg float32) (fact, inc, dec float32) {
	inc = 1
	dec = 1
	if wbAvg < wb.LoThr {
		if wbAvg < wb.AvgThr {
			wbAvg = wb.AvgThr // prevent extreme low if everyone below thr
		}
		fact = wb.LoGain * (wb.LoThr - wbAvg)
		dec = 1 / (1 + fact)
		inc = 2 - dec
	} else if wbAvg > wb.HiThr {
		fact += wb.HiGain * (wbAvg - wb.HiThr)
		if actAvg > wb.ActThr {
			fact += wb.ActGain * (actAvg - wb.ActThr)
		}
		inc = 1 / (1 + fact) // gets sigmoidally small toward 0 as fact gets larger -- is quick acting but saturates -- apply pressure earlier..
		dec = 2 - inc        // as inc goes down, dec goes up..  sum to 2
	}
	return fact, inc, dec
}

/*
  /////////////////////////////////////
  // CtLeabraXCAL code

  INLINE void   GetLrates(LEABRA_CON_STATE* cg, LEABRA_NETWORK_STATE* net, int thr_no,
                          float& clrate, bool& deep_on, float& bg_lrate, float& fg_lrate)  {
    LEABRA_LAYER_STATE* rlay = cg->GetRecvLayer(net);
    clrate = cur_lrate * rlay->lrate_mod;
    deep_on = deep.on;
    if(deep_on) {
      if(!rlay->deep_lrate_mod)
        deep_on = false;          // only applicable to deep_norm active layers
    }
    if(deep_on) {
      bg_lrate = deep.bg_lrate;
      fg_lrate = deep.fg_lrate;
    }
  }
  // #IGNORE get the current learning rates including layer-specific and potential deep modulations

  // todo: should go back and explore this at some point:
  // if(xcal.one_thr) {
  //   float eff_thr = ru_avg_l_lrn * ru_avg_l + (1.0f - ru_avg_l_lrn) * srm;
  //   eff_thr = fminf(eff_thr, 1.0f);
  //   dwt += clrate * xcal.dWtFun(srs, eff_thr);
  // }
  // also: fminf(ru_avg_l,1.0f) for threshold as an option..


*/
