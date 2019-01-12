// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import (
	"github.com/chewxy/math32"
	"github.com/emer/emergent/erand"
	"github.com/goki/ki/kit"
)

///////////////////////////////////////////////////////////////////////
//  act.go contains the activation params and functions for leabra

// leabra.Act contains all the activation computation params and functions for basic leabra
// this is then included in leabra.Layer to drive the computation
type Act struct {
	Act        ActPars       `desc:"X/X+1 rate code activation parameters"`
	OptThresh  OptThreshPars `desc:"optimization thresholds for faster processing"`
	Init       ActInitPars   `desc:"initial values for key network state variables -- initialized at start of trial with InitActs or DecayActs"`
	Dt         DtPars        `desc:"time and rate constants for temporal derivatives / updating of activation state"`
	Gbar       Chans         `desc:"[Defaults: 1, .2, 1, 1] maximal conductances levels for channels"`
	Erev       Chans         `desc:"[Defaults: 1, .3, .25, .1] reversal potentials for each channel"`
	Noise      ActNoisePars  `desc:"how, where, when, and how much noise to add to activations"`
	ClampRange emer.MinMax   `desc:"range of external input activation values allowed -- Max is .95 by default due to saturating nature of rate code activation function"`
	VmRange    emer.MinMax   `desc:"range for Vm membrane potential -- [0, 2.0] by default"`
	ErevSubThr Chans         `inactive:"+" view:"-" desc:"Erev - Act.Thr for each channel -- used in computing GeThrFmG among others"`
	ThrSubErev Chans         `inactive:"+" view:"-" desc:"Act.Thr - Erev for each channel -- used in computing GeThrFmG among others"`
}

func (ac *Act) Defaults() {
	ac.Gbar.SetAl(1.0, 0.2, 1.0, 1.0)
	ac.Erev.SetAl(1.0, 0.3, 0.25, 0.1)
	ac.ClampRange.Max = 0.95
	ac.VmRange.Max = 2.0
}

// Update must be called after any changes to parameteres
func (ac *Act) Update() {
	ac.ErevSubThr.SetFmOtherMinus(ac.Erev, ac.Act.Thr)
	ac.ThrSubErev.SetFmMinusOther(ac.Act.Thr, ac.Erev)
}

// VmFmG computes membrane potential Vm from conductances Ge and Gi.
// The Vm value is only used in pure rate-code computation within the sub-threshold regime
// because firing rate is a direct function of excitatory conductance Ge.
func (ac *Act) VmFmG(nrn *Neuron, thr int) {
	ge := nrn.Ge * ac.Gbar.E
	gi := nrn.Gi * ac.Gbar.I
	nrn.Inet = Compute_INet_impl(u, nrn.Vm, net_eff, gc_i, gc_k)
	nwVm := nrn.Vm + ac.Dt.Integ*ac.Dt.VmDt*nrn.Inet

	if ac.Noise.Type == VmNoise {
		nwVm += nrn.Noise
	}
	nrn.Vm = ac.VmRange.ClipVal(nwVm)
}

func (ac *Act) GeThrFmG(nrn *Neuron, thr int) {
	gcL := ac.Gbar.L
	return ((ac.Gbar.I*nrn.Gi*ac.ErevSubThr.I + gcL*ac.ErevSubThr.L) / ac.ThrSubErev.E)
}

// ActFmG computes rate-coded activation Act from conductances Ge and Gi
func (ac *Act) ActFmG(nrn *Neuron, thr int) {
	var nwAct float32
	if nrn.Act < ac.Act.VmActThr && nrn.Vm <= ac.Act.Thr {
		// note: this is quite important -- if you directly use the gelin
		// the whole time, then units are active right away -- need Vm dynamics to
		// drive subthreshold activation behavior
		nwAct = ac.Act.NoisyXX1(nrn.Vm - ac.Act.Thr)
	} else {
		geThr := ac.GeThrFmG(nrn, thr)
		nwAct = ac.Act.NoisyX11(nrn.Ge*ac.Gbar.E - geThr)
	}
	curAct := nrn.Act
	nwAct = curAct + ac.Dt.Integ*ac.Dt.VmDt*(nwAct-curAct)

	nrn.ActDel = nwAct - curAct
	if ac.Noise.Type == ActNoise {
		nwAct += nrn.Noise
	}
	nrn.Act = nwAct
}

///////////////////////////////////////////////////////////////////////
//  ActPars

// ActPars are the activation parameters for leabra, using the GeLin (g_e linear) rate coded activation function
type ActPars struct {
	Thr          float32 `def:"0.5" desc:"threshold value Theta (Q) for firing output activation (.5 is more accurate value based on AdEx biological parameters and normalization"`
	Gain         float32 `def:"80;100;40;20" min:"0" desc:"gain (gamma) of the rate-coded activation functions -- 100 is default, 80 works better for larger models, and 20 is closer to the actual spiking behavior of the AdEx model -- use lower values for more graded signals, generally in lower input/sensory layers of the network"`
	NVar         float32 `def:"0.005;0.01" min:"0" desc:"variance of the Gaussian noise kernel for convolving with XX1 in NOISY_XX1 and NOISY_LINEAR -- determines the level of curvature of the activation function near the threshold -- increase for more graded responding there -- note that this is not actual stochastic noise, just constant convolved gaussian smoothness to the activation function"`
	AvgCorrect   float32 `desc:"correction factor (multiplier) for average activation level in this layer -- e.g., if using adaptation or stp, may be lower than usual -- taken into account in netinput scaling out of this layer"`
	VmActThr     float32 `def:"0.01" desc:"threshold on activation below which the direct vm - act.thr is used -- this should be low -- once it gets active should use net - g_e_thr ge-linear dynamics (gelin)"`
	SigMult      float32 `def:"0.33" expert:"+" desc:"multiplier on sigmoid used for computing values for net < thr"`
	SigMultPow   float32 `def:"0.8" expert:"+" desc:"power for computing sig_mult_eff as function of gain * nvar"`
	SigGain      float32 `def:"3" expert:"+" desc:"gain multipler on (net - thr) for sigmoid used for computing values for net < thr"`
	InterpRange  float32 `def:"0.01" expert:"+" desc:"interpolation range above zero to use interpolation"`
	GainCorRange float32 `def:"10" expert:"+" desc:"range in units of nvar over which to apply gain correction to compensate for convolution"`
	GainCor      float32 `def:"0.1 expert:"+" desc:"gain correction multiplier -- how much to correct gains"`

	SigGainNVar float32 `view:"-" desc:"sig_gain / nvar"`
	SigMultEff  float32 `view:"-" desc:"overall multiplier on sigmoidal component for values below threshold = sig_mult * pow(gain * nvar, sig_mult_pow)"`
	SigValAt0   float32 `view:"-" desc:"0.5 * sig_mult_eff -- used for interpolation portion"`
	InterpVal   float32 `view:"-" desc:"function value at interp_range - sig_val_at_0 -- for interpolation"`
}

// ActFmGe computes activation based on Ge excitatory conductance
func (ap *ActPars) ActFmGe(ge float32) float32 {
	// compute the activation function
}

// XX1 computes the basic x/(x+1) function
func (ap *ActPars) XX1(x float32) float32 { return x / (x + 1) }

// XX1GainCor computes x/(x+1) with gain correction within GainCorRange
// to compensate for convolution effects
func (ap *ActPars) XX1GainCor(x float32) float32 {
	gainCorFact := (ap.GainCorRange - (x / ap.NVar)) / ap.GainCorRange
	if gainCorFact < 0 {
		return ap.XX1(ap.Gain * x)
	}
	newGain := ap.Gain * (1 - ap.GainCor*ap.GainCorFact)
	return ap.XX1(newGain * x)
}

// NoisyXX1 computes the Noisy x/(x+1) function -- directly computes close approximation
// to x/(x+1) convolved with a gaussian noise function with variance nvar.
// No need for a lookup table -- very reasonable approximation for standard range of parameters
// (nvar = .01 or less -- higher values of nvar are less accurate with large gains,
// but ok for lower gains)
func (ap *ActPars) NoisyXX1(x float32) float32 {
	if x < 0 { // sigmoidal for < 0
		return ap.SigMultEff / (1 + math32.Exp(-(x * ap.SigGainNVar)))
	} else if x < ap.InterpRange {
		interp := 1 - ((ap.InterpRange - x) / ap.InterpRange)
		return ap.SigValAt0 + interp*ap.InterpVal
	} else {
		return ap.XX1GainCor(x)
	}
}

// X11GainCorGain computes x/(x+1) with gain correction within GainCorRange
// to compensate for convolution effects -- using external gain factor
func (ap *ActPars) XX1GainCorGain(x, gain float32) {
	gainCorFact := (ap.GainCorRange - (x / ap.NVar)) / ap.GainCorRange
	if gainCorFact < 0 {
		return ap.XX1(gain * x)
	}
	newGain := gain * (1 - ap.GainCor*gainCorFact)
	return ap.XX1(newGain * x)
}

// NoisyXX1Gain computes the noisy x/(x+1) function -- directly computes close approximation
// to x/(x+1) convolved with a gaussian noise function with variance nvar.
// No need for a lookup table -- very reasonable approximation for standard range of parameters
// (nvar = .01 or less -- higher values of nvar are less accurate with large gains,
// but ok for lower gains).  Using external gain factor.
func (ap *ActPars) NoisyXX1Gain(x, gain float32) {
	if x < ap.InterpRange {
		sigMultEffArg := ap.SigMult * math32.Pow(gain*ap.NVar, ap.SigMultPow)
		sigValAt0Arg := 0.5 * sigMultEffArg

		if x < 0 { // sigmoidal for < 0
			return sigMultEffArg / (1 + math32.Exp(-(x * ap.SigGainNVar)))
		} else { // else x < interp_range
			interp := 1 - ((ap.InterpRange - x) / ap.InterpRange)
			return SigValAt0Arg + interp*ap.InterpVal
		}
	} else {
		return ap.XX1GainCorGain(x, gain)
	}
}

func (ap *ActPars) Update() {
	ap.SigGainNVar = ap.SigGain / ap.NVar
	ap.SigMultEff = ap.SigMult * math32.Pow(ap.Gain*ap.NVar, ap.SigMultPow)
	ap.SigValAt0 = 0.5 * ap.SigMultEff
	ap.InterpVal = ap.XX1GainCor(ap.InterpRange) - ap.SigValAt0
}

func (ap *ActPars) Defaults() {
	ap.Thr = 0.5
	ap.Gain = 100
	ap.NVar = 0.005
	ap.VmActThr = 0.01
	ap.AvgCorrect = 1.0
	ap.SigMult = 0.33
	ap.SigMultPow = 0.8
	ap.SigGain = 3.0
	ap.InterpRange = 0.01
	ap.GainCorRange = 10.0
	ap.GainCor = 0.1
	ap.UpdateParams()
}

//////////////////////////////////////////////////////////////////////////////////////
//  OptThreshPars

// OptThreshPars provides optimization thresholds for faster processing
type OptThreshPars struct {
	Send  float32 `def:"0.1" desc:"don't send activation when act <= send -- greatly speeds processing"`
	Delta float32 `def:"0.005" desc:"don't send activation changes until they exceed this threshold: only for when LeabraNetwork::send_delta is on!"`
}

func (ot *OptThreshPars) Defaults() {
	ot.Send = .1
	ot.Delta = 0.005
}

//////////////////////////////////////////////////////////////////////////////////////
//  ActInitPars

// ActInitPars are initial values for key network state variables.
// Initialized at start of trial with Init_Acts or DecayState.
type ActInitPars struct {
	Vm    float32 `def:"0.4" desc:"initial membrane potential -- see e_rev.l for the resting potential (typically .3) -- often works better to have a somewhat elevated initial membrane potential relative to that"`
	Act   float32 `def:"0" desc:"initial activation value -- typically 0"`
	Netin float32 `def:"0" desc:"baseline level of excitatory net input -- netin is initialized to this value, and it is added in as a constant background level of excitatory input -- captures all the other inputs not represented in the model, and intrinsic excitability, etc"`
}

func (ai *ActInitPars) Defaults() {
	ai.Vm = 0.4
	ai.Act = 0
	ai.Netin = 0
}

//////////////////////////////////////////////////////////////////////////////////////
//  DtPars

// DtPars are time and rate constants for temporal derivatives in Leabra (Vm, net input)
type DtPars struct {
	Integ  float32 `def:"1;0.5" min:"0" desc:"overall rate constant for numerical integration, for all equations at the unit level -- all time constants are specified in millisecond units, with one cycle = 1 msec -- if you instead want to make one cycle = 2 msec, you can do this globaly by setting this integ value to 2 (etc).  However, stability issues will likely arise if you go too high.  For improved numerical stability, you may even need to reduce this value to 0.5 or possibly even lower (typically however this is not necessary).  MUST also coordinate this with network.time_inc variable to ensure that global network.time reflects simulated time accurately"`
	VmTau  float32 `def:"2.81:10" min:"1" desc:"[3.3 std for rate code, 2.81 for spiking] membrane potential and rate-code activation time constant in cycles, which should be milliseconds typically (roughly, how long it takes for value to change significantly -- 1.4x the half-life) -- reflects the capacitance of the neuron in principle -- biological default for AeEx spiking model C = 281 pF = 2.81 normalized -- for rate-code activation, this also determines how fast to integrate computed activation values over time"`
	NetTau float32 `def:"1.4;3;5" min:"1" desc:"net input time constant in cycles, which should be milliseconds typically (roughly, how long it takes for value to change significantly -- 1.4x the half-life) -- this is important for damping oscillations -- generally reflects time constants associated with synaptic channels which are not modeled in the most abstract rate code models (set to 1 for detailed spiking models with more realistic synaptic currents) -- larger values (e.g., 3) can be important for models with higher netinputs that otherwise might be more prone to oscillation, and is default for GPiInvUnitSpec"`

	VmDt  float32 `view:"-" expert:"+" desc:"nominal rate = 1 / tau"`
	NetDt float32 `view:"-" expert:"+" desc:"rate = 1 / tau"`
}

func (dp *DtPars) Update() {
	dp.VmDt = 1 / dp.VmTau
	dp.NetDt = 1 / dp.NetTau
}

func (dp *DtPars) Defaults() {
	dp.Integ = 1
	dp.VmTau = 3.3
	dp.NetTau = 1.4
	ap.Update()

}

//////////////////////////////////////////////////////////////////////////////////////
//  Chans

// Chans are ion channels used in computing point-neuron activation function
type Chans struct {
	E float32 `desc:"excitatory sodium (Na) AMPA channels activated by synaptic glutamate"`
	L float32 `desc:"constant leak (potassium, K+) channels -- determines resting potential (typically higher than resting potential of K)"`
	I float32 `desc:"inhibitory chloride (Cl-) channels activated by synaptic GABA"`
	K float32 `desc:"gated / active potassium channels -- typicaly hyperpolarizing relative to leak / rest"`
}

// SetAll sets all the values
func (ch *Chans) SetAll(e, l, i, k float32) {
	ch.E, ch.L, ch.I, ch.K = e, l, i, k
}

// SetFmOtherMinus sets all the values from other Chans minus given value
func (ch *Chans) SetFmOtherMinus(oth Chans, minus float32) {
	ch.E, ch.L, ch.I, ch.K = oth.E-minus, oth.L-minus, oth.I-minus, oth.k-minus
}

// SetFmMinusOther sets all the values from given value minus other Chans
func (ch *Chans) SetFmMinusOther(minus float32, oth Chans) {
	ch.E, ch.L, ch.I, ch.K = minus-oth.E, minus-oth.L, minus-oth.I, minus-oth.k
}

//////////////////////////////////////////////////////////////////////////////////////
//  Noise

// ActNoiseType are different types / locations of random noise for activations
type ActNoiseType int

//go:generate stringer -type=ActNoiseType

var KiT_ActNoiseType = kit.Enums.AddEnum(ActNoiseTypeN, false, nil)

func (ev ActNoiseType) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *ActNoiseType) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }

// The activation noise types
const (
	// NoNoise means no noise added
	NoNoise ActNoiseType = iota

	// VmNoise means noise is added to the membrane potential.
	// IMPORTANT: this should NOT be used for rate-code (NXX1) activations,
	// because they do not depend directly on the vm -- this then has no effect
	VmNoise

	// GeNoise means noise is added to the excitatory conductance (Ge).
	// This should be used for rate coded activations (NXX1)
	GeNoise

	// ActNoise means noise is added to the final rate code activation
	ActNoise

	// GeMultNoise means that noise is multiplicative on the Ge excitatory conductance values
	GeMultNoise
)

// ActNoisePars contains parameters for activation-level noise
type ActNoisePars struct {
	erand.RndPars
	Type       ActNoiseType `desc:"where and how to add processing noise"`
	TrialFixed bool         `desc:"keep the same noise value over the entire trial -- prevents noise from being washed out and produces a stable effect that can be better used for learning -- this is strongly recommended for most learning situations"`
}

func (an *ActNoisePars) Defaults() {
	an.TrialFixed = true
}
