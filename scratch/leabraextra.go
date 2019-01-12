// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scratch

import (
	"github.com/emer/emergent/emer"
	"github.com/goki/ki"
)


type Neuron struct {
  float         bias_wt;
  // #VIEW_HOT #CAT_Bias bias weight value -- the bias weight acts like a connection from a unit that is always active with a constant value of 1 -- reflects intrinsic excitability from a biological perspective
  float         bias_dwt;
  // #VIEW_HOT #CAT_Bias change in bias weight value as computed by a learning mechanism
  float      bias_fwt;       // #NO_SAVE #CAT_Learning bias weight: fast learning linear (underlying) weight value -- learns according to the lrate specified in the connection spec -- this is converted into the effective weight value, "wt", via sigmoidal contrast enhancement (wt_sig)
  float      bias_swt;       // #NO_SAVE #CAT_Learning bias weight: slow learning linear (underlying) weight value -- learns more slowly from weight changes than fast weights, and fwt decays down to swt over time
  float      ext_orig;       // #NO_SAVE #CAT_Activation original external input value (ext) -- need to save this in case ext gets transformed in various ways in the clamping process e.g., for ScalarValue layers

  float      spike;          // #CAT_Activation discrete spiking event, is 1.0 when the neuron spikes and 0.0 otherwise, and corresponds to act for spike activation -- for rate code equation (NXX1), spikes are triggered identically to spiking mode based on the vm membrane potential dynamics, even though act* is computed through the rate code equation
  float      spike_isi;      // #CAT_Activation time-averaged inter-spike-interval -- updated after every spike -- this is then used in computing the rate code relative to an expected maximum firing rate

v_m_eq       // #NO_SAVE #CAT_Activation equilibrium membrane potential -- this is NOT reset by spiking, so it reaches equilibrium values asymptotically -- it is used for rate code activation in sub-threshold range (whenever v_m_eq < act.thr) -- the gelin activation function does not otherwise provide useful dynamics in this subthreshold range

	margin       // #CAT_Activation relative status of this unit for the overall activation state / attractor, used for favoring units on the edges or margins of the attractor: -2 = below the low threshold -- solidly OFF, -1.0 = above the low threshold but below the midway threshold for the marginal units, +1.0 = above the midway threshold but still in the margin, +2.0 = above the high threshold and solidly within the main attractor state

	act_raw      // #CAT_Activation raw superficial-layer activation prior to mutliplication by deep_norm -- this may reflect layer 4 activation -- used in computing new deep_raw values
	deep_raw     // #NO_SAVE #CAT_Activation deep layer raw activation values -- these reflect the raw output from a microcolumn, in the form of layer 5b tufted neurons that project to the thalamus -- they integrate local thresholded input from superficial layer and top-down deep-layer input from other areas, to provide raw attentional and output signal from an area -- this signal drives deep_ctxt temporal integration (TI) for predictive learning, in addition to attention
	deep_raw_prv // #NO_SAVE #CAT_Activation previous value of the deep layer raw activation values -- used for temporal context learning
	deep_ctxt    // #NO_SAVE #CAT_Activation temporally-delayed local lateral integration of deep_raw signals sent via DeepCtxtConSpec connections to provide context for temporal integration (TI) learning -- added into net input of superficial neurons -- computed at start of new alpha trial (quarter after deep_raw_qtr)
	deep_mod     // #NO_SAVE #CAT_Activation current modulating value of deep layer 6 corticothalamic, regular spiking neurons that represents the net attentional filter applied to the superficial layers -- value is computed from deep_net received via SendDeepModConSpec projections from deep layer units, and directly multiplies the superficial activations (act)
	deep_lrn     // #NO_SAVE #CAT_Activation net influence of deep layer dynamics on learning rate for connections into this unit -- typically set to deep_mod prior to enforcing the mod_min floor value, so that baseline deep_mod=0 units get lowest background learning rate
	deep_mod_net // #NO_SAVE #CAT_Activation modulatory net input from deep layer activations (representing lamina 6 regular spiking, thalamocortical projecting neurons) via SendDeepModConSpec, drives deep mod of superficial neurons
	deep_raw_net // #NO_SAVE #CAT_Activation deep_raw net input from deep layer activations (representing lamina 5b intrinsic bursting neurons), typically for driver inputs into thalamic relay cells via SendDeepRawConSpec
	deep_raw_sent // #NO_SAVE #EXPERT #CAT_Activation last deep_raw activation value sent in computing deep_raw_net
	
	thal         // #NO_SAVE #CAT_Activation thalamic activation value, driven by a ThalSendUnitSpec or GpiInvUnitSpec -- used by deep params in LeabraUnitSpec and MSNConSpecs, and possibly other specs, to respond to thalamic inputs
	thal_gate    // #NO_SAVE #CAT_Activation discrete thalamic gating signal -- typically activates to 1 when thalamic pathway gates, and is 0 otherwise -- PFC and BG layers receive this signal to drive updating etc at the proper time -- other layers can use the LeabraNetwork times.thal_gate_cycle signal
	thal_cnt     // #NO_SAVE #CAT_Activation counter for thalamic activation value -- increments for active maintenance in PFCUnitSpec
	act_g        // #NO_SAVE #CAT_Activation records the activation state when gating occurs -- for PFC and BG units this is based on direct thal_gate signal, and for other units it is based on LeabraNetwork times.thal_gate signal, which is actvated when thalamic layers gate and are configured to update the global signal -- is either act_eq or act_nd depending on act_misc.rec_nd setting

		gc_kna_f     // #NO_SAVE #CAT_Activation fast time constant sodium-gated potassium channel activation -- drives adapatation -- fast time constant
	gc_kna_m     // #NO_SAVE #CAT_Activation medium time constant sodium-gated potassium channel activation -- drives adapatation -- medium time constant
	gc_kna_s     // #NO_SAVE #CAT_Activation slow time constant sodium-gated potassium channel activation -- drives adapatation -- slow time constant

	syn_tr       // #NO_SAVE #CAT_Activation presynaptic (sending) synapse value: total amount of transmitter ready to release = number of vesicles ready to release (syn_nr) x probability of release (syn_pr) (controlled by short-term-plasticity equations, stp) -- this multiplies activations to produce net sending activation (also affects act_eq, but not act_nd)
	syn_nr       // #NO_SAVE #CAT_Activation presynaptic (sending) synapse value: number of vesicles ready to release at next spike -- vesicles are depleated when released, resulting in short-term depression of net synaptic efficacy, and recover with both activity dependent and independent rate constants (controlled by short-term-plasticity equations, stp)
	syn_pr       // #NO_SAVE #CAT_Activation presynaptic (sending) synapse value: probability of vesicle release at next spike -- probability varies as a function of local calcium available to drive the release process -- this increases with recent synaptic activity (controlled by short-term-plasticity equations, stp)
	syn_kre      // #NO_SAVE #CAT_Activation presynaptic (sending) synapse value: dynamic time constant for rate of recovery of number of vesicles ready to release -- this dynamic time constant increases with each action potential, and decays back down over time, and makes the response to higher-frequency spike trains more linear (controlled by short-term-plasticity equations, stp)
	
	da_p         // #NO_SAVE #CAT_Activation positive-valence oriented dopamine value -- this typically exhibits phasic bursts (positive values) with unanticipated increases in reward outcomes / expectations, and phasic dips (negative values) with unanticipated decreases thereof.  This value can drive dopaminergic learning rules and activation changes in receiving neurons -- typically sent by VTAUnitSpec units -- see also da_n
	da_n         // #NO_SAVE #CAT_Activation negative-valence oriented dopamine value -- this typically exhibits phasic bursts (positive values) with unanticipated increases in negative outcomes / expectations, and phasic dips (negative values) with unanticipated decreases thereof.  This value can drive dopaminergic learning rules and activation changes in receiving neurons -- typically sent by VTAUnitSpec units with appropriate flags set -- see also da_p
	sev          // #NO_SAVE #CAT_Activation serotonin value -- driven by Dorsal Raphe Nucleus (DRNUnitSpec) or other sources -- generally thought to reflect longer time-averages of overall progress or lack thereof
	ach          // #NO_SAVE #CAT_Activation acetylcholine value -- driven by Tonically Active Neurons (TAN's) in the Striatum, or Basal Forebrain Cholenergic System (BFCS) potentially other sources -- effects depend strongly on types of receptors present
	shunt        // #NO_SAVE #CAT_Activation shunting value -- modulatory signal that shunts activity in other layers -- currently sent by PatchUnitSpec to MSNUnitSpecs in the BG to shunt dopamine and ach

	misc_1 // #NO_SAVE #CAT_Activation miscellaneous variable for special algorithms / subtypes that need it
	misc_2 // #NO_SAVE #CAT_Activation miscellaneous variable for special algorithms / subtypes that need it
	int   spk_t  // #NO_SAVE #CAT_Activation time in tot_cycle units when spiking last occurred (-1 for not yet)
	
}


class STATE_CLASS(LeabraActMiscSpec) : public STATE_CLASS(SpecMemberBase) {
  // ##INLINE ##NO_TOKENS ##CAT_Leabra miscellaneous activation computation parameters and specs
INHERITED(SpecMemberBase)
public:
  bool          rec_nd;         // record the act_nd non-depressed activation variable (instead of act_eq) for the act_q* quarter-trial and phase (act_m, act_p) activation state variables -- these are used primarily for statistics, or possibly for specialized learning mechanisms
  bool          avg_nd;         // use the act_nd non-depressed activation variable (instead of act_eq) for the time-average activation values (avg_ss, avg_s, avg_m, avg_l) used in the XCAL learning mechanism -- this is appropriate for action-potential driven learning dynamics, as compared to synaptic efficacy, when short term plasticity is present
  bool          dif_avg;        // compute act_dif as ru_avg_s_lrn - avg_m (difference of average values that actually drive learning) -- otherwise it is act_p - act_m (difference of final activation states in plus phase minus minus phase -- the typical error signal)
  float         net_gain;       // def:"1 #MIN_0 multiplier on total synaptic net input -- this multiplies the net_raw, but AFTER the net_raw variable is saved (upon which the netin_raw statistics are computed)

  bool          avg_trace;      // def:"false set act_avg unit variable to the exponentially decaying trace of activation -- used for TD (temporal differences) reinforcement learning for example -- lambda parameter determines how much of the prior trace carries over into the new trace 
  float         lambda;         // #CONDSHOW_ON_avg_trace determines how much of the prior trace carries over into the new trace (act_avg = lambda * act_avg + new_act)
  float         avg_tau;        // #CONDSHOW_OFF_avg_trace def:"200 #MIN_1 for integrating activation average (act_avg), time constant in trials (roughly, how long it takes for value to change significantly) -- used mostly for visualization and tracking "hog" units
  float         avg_init;        // def:"0.15 #MIN_0 initial activation average value -- used for act_avg, avg_s, avg_m, avg_l
  float         avg_dt;          // view:"-" expert:"+" rate = 1 / tau

  STATE_DECO_KEY("UnitSpec");
  STATE_TA_STD_CODE_SPEC(LeabraActMiscSpec);

  STATE_UAE(avg_dt = 1.0f / avg_tau;);

private:
  void        Initialize()      { Defaults_init(); }
  void        Defaults_init()   {
    rec_nd = true; avg_nd = true; dif_avg = false; net_gain = 1.0f;
    avg_trace = false; lambda = 0.0f; avg_tau = 200.0f;
    avg_init = 0.15f;

    avg_dt = 1.0f / avg_tau;
  }
};


class STATE_CLASS(SpikeFunSpec) : public STATE_CLASS(SpecMemberBase) {
  // ##INLINE ##NO_TOKENS ##CAT_Leabra spiking activation function specs -- conductance is computed postsynaptically using an alpha function based on spike pulses sent presynaptically -- for clamped layers, spiking probability is proportional to external input controlled by the clamp_type and clamp_max_p values -- soft clamping may still be a better option though
INHERITED(SpecMemberBase)
public:
  float         rise;            // def:"0 #MIN_0 exponential rise time (in cycles) of the synaptic conductance according to the alpha function 1/(decay - rise) [e^(-t/decay) - e^(-t/rise)] -- set to 0 to only include decay time (1/decay e^(-t/decay)), which is highly optimized (doesn't use window -- just uses recursive exp decay) and thus the default!
  float         decay;           // def:"5 #MIN_0 exponential decay time (in cycles) of the synaptic conductance according to the alpha function 1/(decay - rise) [e^(-t/decay) - e^(-t/rise)] -- set to 0 to implement a delta function (not very useful)
  float         g_gain;          // def:"9 #MIN_0 multiplier for the spike-generated conductances when using alpha function which is normalized by area under the curve -- needed to recalibrate the alpha-function currents relative to rate code net input which is overall larger -- in general making this the same as the decay constant works well, effectively neutralizing the area normalization (results in consistent peak current, but differential integrated current over time as a function of rise and decay)
  int           window;          // def:"3 #MIN_0 #MAX_10 spike integration window -- when rise==0, this window is used to smooth out the spike impulses similar to a rise time -- each net contributes over the window in proportion to 1/window -- for rise > 0, this is used for computing the alpha function -- should be long enough to incorporate the bulk of the alpha function, but the longer the window, the greater the computational cost (max of 10 imposed by fixed buffer required in LeabraUnitState_cpp structure)
  float         act_max_hz;      // def:"180 #MIN_1 for translating spiking interval (rate) into rate-code activation equivalent (and vice-versa, for clamped layers), what is the maximum firing rate associated with a maximum activation value (max act is typically 1.0 -- depends on act_range)
  float         int_tau;         // def:"5 #MIN_1 time constant for integrating the spiking interval in estimating spiking rate

  float         gg_decay;        // view:"-" #NO_SAVE g_gain/decay
  float         gg_decay_sq;     // view:"-" #NO_SAVE g_gain/decay^2
  float         gg_decay_rise;   // view:"-" #NO_SAVE g_gain/(decay-rise)
  float         oneo_decay;      // view:"-" #NO_SAVE 1.0/decay
  float         oneo_rise;       // view:"-" #NO_SAVE 1.0/rise
  float         int_dt;          // view:"-" expert:"+" rate = 1 / tau

  INLINE float  ComputeAlpha(float t) {
    if(decay == 0.0f) return (t == 0.0f) ? g_gain : 0.0f; // delta function
    // todo: replace with exp_fast -- and benchmark!
    // if(rise == 0.0f) return gg_decay * expf(-t * oneo_decay);         // exponential
    // if(rise == decay) return t * gg_decay_sq * expf(-t * oneo_decay); // symmetric alpha
    // return gg_decay_rise * (expf(-t * oneo_decay) - expf(-t * oneo_rise)); // full alpha
    if(rise == 0.0f) return gg_decay * STATE_CLASS(taMath_float)::exp_fast(-t * oneo_decay);         // exponential
    if(rise == decay) return t * gg_decay_sq * STATE_CLASS(taMath_float)::exp_fast(-t * oneo_decay); // symmetric alpha
    return gg_decay_rise * (STATE_CLASS(taMath_float)::exp_fast(-t * oneo_decay) - STATE_CLASS(taMath_float)::exp_fast(-t * oneo_rise)); // full alpha
  }

  INLINE int    ActToInterval(const float time_inc, const float integ, const float act)
  { return (int) (1.0f / (time_inc * integ * act * act_max_hz)); }
  // #CAT_Activation compute spiking interval based on network time_inc, dt.integ, and unit act -- note that network time_inc is usually .001 = 1 msec per cycle -- this depends on that being accurately set

  INLINE float  ActFromInterval(float spike_isi, const float time_inc, const float integ) {
    if(spike_isi == 0.0f) {
      return 0.0f;              // rate is 0
    }
    float max_hz_int = 1.0f / (time_inc * integ * act_max_hz); // interval at max hz..
    return max_hz_int / spike_isi; // normalized
  }
  // #CAT_Activation compute rate-code activation from estimated spiking interval

  INLINE void   UpdateSpikeInterval(float& spike_isi, float cur_int) {
    if(spike_isi == 0.0f) {
      spike_isi = cur_int;      // use it
    }
    else if(cur_int < 0.8f * spike_isi) {
      spike_isi = cur_int;      // if significantly less than we take that
    }
    else {                                         // integrate on slower
      spike_isi += int_dt * (cur_int - spike_isi); // running avg updt
    }
  }
  // #CAT_Activation update running-average spike interval estimate

  INLINE void   UpdateRates() {
    if(window <= 0) window = 1;
    if(decay > 0.0f) {
      gg_decay = g_gain / decay;
      gg_decay_sq = g_gain / (decay * decay);
      if(decay != rise)
        gg_decay_rise = g_gain / (decay - rise);

      oneo_decay = 1.0f / decay;
      if(rise > 0.0f)
        oneo_rise = 1.0f / rise;
      else
        oneo_rise = 1.0f;
    }
    int_dt = 1.0f / int_tau;
  }
  // #IGNORE update derive rates
  
  
  STATE_DECO_KEY("UnitSpec");
  STATE_TA_STD_CODE_SPEC(SpikeFunSpec);

  STATE_UAE( UpdateRates(); );
  
private:
  void        Initialize()    { Defaults_init(); }
  void        Defaults_init() {
    g_gain = 9.0f; rise = 0.0f; decay = 5.0f; window = 3;
    act_max_hz = 180.0f;  int_tau = 5.0f;
    UpdateRates();
  }
};


class STATE_CLASS(SpikeMiscSpec) : public STATE_CLASS(SpecMemberBase) {
  // ##INLINE ##NO_TOKENS ##CAT_Leabra extra misc spiking parameters 
INHERITED(SpecMemberBase)
public:
  enum ClampType {                // how to generate spikes during hard clamp conditions
    POISSON,                        // generate spikes according to Poisson distribution with probability = clamp_max_p * u->ext
    UNIFORM,                        // generate spikes according to Uniform distribution with probability = clamp_max_p * u->ext
    REGULAR,                        // generate spikes every 1 / (clamp_max_p * u->ext) cycles -- this works the best, at least in smaller networks, due to the lack of additional noise, and the synchrony of the inputs for driving synchrony elsewhere
    CLAMPED,                        // just use the straight clamped activation value -- do not do any further modifications
  };

  bool          ex;             // def:"false turn on exponential excitatory current that drives v_m rapidly upward for spiking as it gets past its nominal firing threshold (act.thr) -- nicely captures the Hodgkin Huxley dynamics of Na and K channels -- uses Brette & Gurstner 2005 AdEx formulation -- this mechanism has an unfortunate interaction with the continuous inhibitory currents generated by the standard FF_FB inhibitory function, which cause this mechanism to desensitize and fail to spike
  float         exp_slope;        // #CONDSHOW_ON_ex def:"0.02 slope in v_m (2 mV = .02 in normalized units) for extra exponential excitatory current that drives v_m rapidly upward for spiking as it gets past its nominal firing threshold (act.thr) -- nicely captures the Hodgkin Huxley dynamics of Na and K channels -- uses Brette & Gurstner 2005 AdEx formulation -- a value of 0 disables this mechanism
  float         spk_thr;        // #CONDSHOW_ON_ex def:"1.2 membrane potential threshold for actually triggering a spike when using the exponential mechanism -- the nominal threshold in act.thr enters into the exponential mechanism, but this value is actually used for spike thresholding 
  float         vm_r;                // def:"0;0.15;0.3 #AKA_v_m_r post-spiking membrane potential to reset to, produces refractory effect if lower than vm_init -- 0.30 is apropriate biologically-based value for AdEx (Brette & Gurstner, 2005) parameters
  int           t_r;                // def:"3 post-spiking explicit refractory period, in cycles -- prevents v_m updating for this number of cycles post firing
  float         clamp_max_p;        // def:"0.12 #MIN_0 #MAX_1 maximum probability of spike rate firing for hard-clamped external inputs -- multiply ext value times this to get overall probability of firing a spike -- distribution is determined by clamp_type
  ClampType     clamp_type;        // how to generate spikes when layer is hard clamped -- in many cases soft clamping may work better

  float         eff_spk_thr;    // #HIDDEN view:"-" effective spiking threshold -- depends on whether exponential mechanism is being used (= act.thr if not ex, else spk_thr)

  STATE_DECO_KEY("UnitSpec");
  STATE_TA_STD_CODE_SPEC(SpikeMiscSpec);

private:
  void        Initialize()      { Defaults_init(); }
  void        Defaults_init() {
    ex = false;
    exp_slope = 0.02f;
    spk_thr = 1.2f;
    clamp_max_p = 0.12f;
    clamp_type = REGULAR;
    vm_r = 0.30f;
    t_r = 3;

    eff_spk_thr = 0.5f;           // ex = off
  }
};

type DtPars struct {
  VmCyc int `min:"1" desc:"number of steps to integrate membrane potential vm -- each cycle is a midpoint method integration step"`
  FastCyc int        // #AKA_vm_eq_cyc def:"0 number of cycles at start of a trial to run units in a fast integration mode -- the rate-code activations have no effective time constant and change immediately to the new computed value (vm_time is ignored) and vm is computed as an equilibirium potential given current inputs: set to 1 to quickly activate soft-clamped input layers (primary use); set to 100 to always use this computation
}

class STATE_CLASS(ShortPlastSpec) : public STATE_CLASS(SpecMemberBase) {
  // ##INLINE ##NO_TOKENS ##CAT_Leabra short-term plasticity specifications -- different algorithms are available to update the syn_tr amount of neurotransmitter available to release, which multiplies computed firing rate or spiking (but not act_nd) to produce a net sending activation efficacy in the act and act_eq variables
INHERITED(SpecMemberBase)
public:
  enum STPAlgorithm {           // which algorithm to use for STP
     CYCLES,                    // uses standard equations summarized in Hennig, 2013 (eq 6) to capture both facilitation and depression dynamics as a function of presynaptic firing -- models interactions between number of vesicles available to release, and probability of release, and a time-varying recovery rate -- rate code uses generated spike var to drive this
     TRIAL_BINARY,              // units continously above thresh for n_trials will depress to 0 synaptic transmitter and individually recover at rec_prob to full strength on trial-by-trial basis (update at start of trial)
  };
 
  bool          on;             // synaptic depression is in effect: multiplies normal activation computed by current activation function in effect
  STPAlgorithm  algorithm;      // #CONDSHOW_ON_on which algorithm to use for computing short term synaptic plasticity, syn_tr (and other related syn_ vars depending on algo)
  float         f_r_ratio;      // #CONDSHOW_ON_on&&algorithm:CYCLES def:"0.01:3 ratio of facilitating (t_fac) to depression recovery (t_rec) time constants -- influences overall nature of response balance (ratio = 1 is balanced, > 1 is facilitating, < 1 is depressing).  Wang et al 2006 found: ~2.5 for strongly facilitating PFC neurons (E1), ~0.02 for strongly depressing PFC and visual cortex (E2), and ~1.0 for balanced PFC (E3)
  float         rec_tau;        // #CONDSHOW_ON_on&&algorithm:CYCLES def:"100:1000 min:"1 [200 std] time constant in cycles, which should be milliseconds typically (roughly, how long it takes for value to change significantly -- 1.4x the half-life) for the constant form of the recovery of number of available vesicles to release at each action potential -- one factor influencing how strong and long-lasting depression is: nr += (1-nr)/rec_tau.  Wang et al 2006 found: ~200ms for strongly depressing in visual cortex and facilitating PFC (E1), 600ms for depressing PFC (E2), and between 200-600 for balanced (E3)
  float         p0;             // #CONDSHOW_ON_on&&algorithm:CYCLES def:"0.1:0.4 [0.2 std] baseline probability of release -- lower values around .1 produce more strongly facilitating dynamics, while .4 makes depression dominant -- interacts with f_r_ratio time constants as well.  Tuning advice: keeping all other params at their default values, and focusing on depressing dynamics, this value relative to p0_norm = 0.2 can give different degrees of depression: 0.2 = strong depression, 0.15 = weaker, and 0.1 = very weak depression dynamics
  float         p0_norm;        // #CONDSHOW_ON_on&&algorithm:CYCLES def:"0.1:1 [0.2 std] baseline probability of release to use for normalizing the overall net synaptic transmitter release (syn_tr) -- for depressing synapses, this should be = p0, but for facilitating, it make sense to normalize at a somewhat higher level, so that the syn_tr starts out lower and rises to a max -- it maxes out at 1.0 so you don't want to lose dynamic range
  float         kre_tau;        // #CONDSHOW_ON_on&&algorithm:CYCLES def:"100 time constant in cycles, which should be milliseconds typically (roughly, how long it takes for value to change significantly -- 1.4x the half-life) on dynamic enhancement of time constant of recovery due to activation -- recovery time constant increases as a function of activity, helping to linearize response (reduce level of depression) at higher frequencies -- supported by multiple sources of biological data (Hennig, 2013)
  float         kre;            // #CONDSHOW_ON_on&&algorithm:CYCLES def:"0.002;0 how much the dynamic enhancement of recovery time constant increases for each action potential -- determines how strong this dynamic component is -- set to 0 to turn off this extra adaptation
  float         fac_tau;        // #CONDSHOW_ON_on&&algorithm:CYCLES view:"-" #SHOW auto computed from f_r_ratio and rec_tau: time constant in cycles, which should be milliseconds typically (roughly, how long it takes for value to change significantly -- 1.4x the half-life) for the dynamics of facilitation of release probability: pr += (p0 - pr) / fac_tau. Wang et al 2006 found: 6ms for visual cortex, 10-20ms strongly depressing PFC (E2), ~500ms for strongly facilitating (E1), and between 200-600 for balanced (E3)
  float         fac;            // #CONDSHOW_ON_on&&algorithm:CYCLES def:"0.2:0.5 min:"0 strength of facilitation effect -- how much each action potential facilitates the probability of release toward a maximum of one: pr += fac (1-pr) -- typically right around 0.3 in Wang et al, 2006

  float         thresh;         // #CONDSHOW_ON_on&&algorithm:TRIAL_BINARY def:"0.5 the levels of activation in q3 over which a unit is subject to synaptic NT depletion
  int           n_trials;       // #CONDSHOW_ON_on&&algorithm:TRIAL_BINARY number of continious trials above threshold after which syn_tr transmitter available goes to 0
  float         rec_prob;       // #CONDSHOW_ON_on&&algorithm:TRIAL_BINARY min:"0 #MAX_1  depleted units recover independently with recovery probability on a trial by trial basis 

  float         rec_dt;         // #CONDSHOW_ON_on&&algorithm:CYCLES view:"-" expert:"+" rate constant for recovery = 1 / rec_tau
  float         fac_dt;         // #CONDSHOW_ON_on&&algorithm:CYCLES view:"-" expert:"+" rate constant for facilitation =  1 / fac_tau
  float         kre_dt;         // #CONDSHOW_ON_on&&algorithm:CYCLES view:"-" expert:"+" rate constant for recovery enhancement = 1 / kre_tau
  float         oneo_p0_norm;   // #CONDSHOW_ON_on&&algorithm:CYCLES view:"-" expert:"+" 1 / p0_norm
  

  INLINE float dNR(float dt_integ, float syn_kre, float syn_nr, float syn_pr, float spike) {
    return (dt_integ * rec_dt + syn_kre) * (1.0f - syn_nr) - syn_pr * syn_nr * spike;
  }
  
  INLINE float dPR(float dt_integ, float syn_pr, float spike) {
    return dt_integ * fac_dt * (p0 - syn_pr) + fac * (1.0f - syn_pr) * spike;
  }

  INLINE float dKRE(float dt_integ, float syn_kre, float spike) {
    return -dt_integ * kre_dt * syn_kre + kre * (1.0f - syn_kre) * spike;
  }

  INLINE float TR(float syn_nr, float syn_pr) {
    float syn_tr = oneo_p0_norm * (syn_nr * syn_pr); // normalize pr by p0_norm
    if(syn_tr > 1.0f) syn_tr = 1.0f;                  // max out at 1.0
    return syn_tr;
  }
  
  STATE_DECO_KEY("UnitSpec");
  STATE_TA_STD_CODE_SPEC(ShortPlastSpec);
  
  STATE_UAE
    ( fac_tau = f_r_ratio * rec_tau;  rec_dt = 1.0f / rec_tau;
      fac_dt = 1.0f / fac_tau;  kre_dt = 1.0f / kre_tau;
      oneo_p0_norm = 1.0f / p0_norm; );
  
private:
  void        Initialize() {
    on = false;  algorithm = CYCLES;  f_r_ratio = 0.02f;  kre = 0.002f;
    Defaults_init();
  }
  
  void        Defaults_init() {
    p0 = 0.2f;
    p0_norm = 0.2f;
    rec_tau = 200.0f;
    fac = 0.3f;
    kre_tau = 100.0f;

    fac_tau = f_r_ratio * rec_tau;
    rec_dt = 1.0f / rec_tau;
    fac_dt = 1.0f / fac_tau;
    kre_dt = 1.0f / kre_tau;
    oneo_p0_norm = 1.0f / p0_norm;
  
    thresh = 0.5f;
    n_trials = 1;
    rec_prob = 0.1f;
  }    
};

class STATE_CLASS(SynDelaySpec) : public STATE_CLASS(SpecMemberBase) {
  // ##INLINE ##NO_TOKENS ##CAT_Leabra synaptic delay -- activation sent to other units is delayed by a given number of cycles
INHERITED(SpecMemberBase)
public:
  bool          on;                // is synaptic delay active?
  int           delay;             // #CONDSHOW_ON_on min:"0 number of cycles to delay for

  STATE_DECO_KEY("UnitSpec");
    STATE_TA_STD_CODE_SPEC(SynDelaySpec);
private:
  void        Initialize()      { on = false; delay = 4; Defaults_init(); }
  void        Defaults_init()   { }; // note: does NOT do any init -- these vals are not really subject to defaults in the usual way, so don't mess with them
};


class STATE_CLASS(DeepSpec) : public STATE_CLASS(SpecMemberBase) {
  // ##INLINE ##NO_TOKENS ##CAT_Leabra specs for DeepLeabra deep neocortical layer dynamics, which capture attentional, thalamic auto-encoder, and temporal integration mechanisms 
INHERITED(SpecMemberBase)
public:
  enum DeepRole {          // what role do these neurons play in the deep layer network dynamics -- determines what function the deep_net plays
    SUPER,                 // superficial layer cortical neurons -- generate a deep_raw activation based on thresholds, and receive a deep_net signal from ongoing deep layer neuron activations via SendDeepModConSpec connections, which they then turn into a deep_mod activation value that multiplies activations in this layer
    DEEP,                  // deep layer cortical neurons -- receive deep_net inputs via DeepCtxtConSpec from superficial layer neurons to drive deep_ctxt values that are added into net input along with other inputs to support predictive learning of thalamic relay cell layer activations, which these should project to and receive from -- they should also receive direct top-down projections from other deep layer neurons for top-down attentional signals
    TRC,                   // thalamic relay cell neurons -- receive a SendDeepRawConSpec topographic projection from lower-layer superficial neurons (sending their deep_raw values into deep_net) which is all that is used for net input in the plus phase -- minus phase activation is driven from projections from deep layer predictive learning neurons
  };

  bool       on;         // enable the DeepLeabra mechanisms, including temporal integration via deep_ctxt context connections, thalamic-based auto-encoder driven by deep_raw projections, and attentional modulation by deep_mod
  DeepRole   role;       // #CONDSHOW_ON_on what role do these neurons play in overall deep layer network dynamics -- determines what function the deep_net plays, among other things
  float      raw_thr_rel;    // #CONDSHOW_ON_on #MAX_1 def:"0.1;0.2;0.5 #AKA_thr_rel relative threshold on act_raw value (distance between average and maximum act_raw values within layer, e.g., 0 = average, 1 = max) for deep_raw neurons to fire -- neurons below this level have deep_raw = 0 -- above this level, deep_raw = act_raw
  float      raw_thr_abs;    // #CONDSHOW_ON_on min:"0 #MAX_1 def:"0.1;0.2;0.5 #AKA_thr_abs absolute threshold on act_raw value for deep_raw neurons to fire -- see thr_rel for relative threshold and activation value -- effective threshold is MAX of relative and absolute thresholds
  float      mod_min;     // #CONDSHOW_ON_on&&role:SUPER min:"0 #MAX_1 minimum deep_mod value -- provides a non-zero baseline for deep-layer modulation
  float      mod_thr;     // #CONDSHOW_ON_on&&role:SUPER min:"0 threshold on deep_mod_net before deep mod is applied -- if not receiving even this amount of overall input from deep_mod sender, then do not use the deep_mod_net to drive deep_mod and deep_lrn values -- only for SUPER units -- based on LAYER level maximum for base LeabraLayerSpec, PVLV classes are based on actual deep_mod_net for each unit
  float      ctxt_prv;    // #CONDSHOW_ON_on&&role:DEEP min:"0 #MAX_1 amount of prior deep context to retain when updating deep context net input -- (1-ctxt_prv) will be used for the amount of new context to add -- provides a built-in level of hysteresis / longer-term memory of prior informaiton -- can also achieve this kind of functionality, with more learning dynamics, using a deep ti context projection from the deep layer itself!
  int        tick_updt;   // #CONDSHOW_ON_on&&role:DEEP if this value is >= 0, then only perform normal deep context updating when network.tick is this value -- otherwise use the else_prv value instead of ctxt_prv to determine how much of the previous context to retain (typically set this to a high value near 1 to retain information from the tick_updt time period) -- this simulates a simple form of gating-like behavior in the updating of deep context information
  float      else_prv;    // #CONDSHOW_OFF_tick_updt:-1||!on when tick_updt is being used, this is the amount of prior deep context to retain on all other ticks aside from tick_updt when updating deep context net input -- (1-else_prv) will be used for the amount of new context to add -- ctxt_prv is still used on the time of tick_updt in case that is non-zero

  float      mod_range;  // view:"-" expert:"+" 1 - mod_min -- range for the netinput to modulate value of deep_mod, between min and 1 value
  float      ctxt_new;   // view:"-" expert:"+" 1 - ctxt_prv -- new context amount
  float      else_new;   // view:"-" expert:"+" 1 - else_prv -- new context amount
  
  INLINE bool   IsSuper()
  { return on && role == SUPER; }
  // are we SUPER?
  INLINE bool   IsDeep()
  { return on && role == DEEP; }
  // are we DEEP?
  INLINE bool   IsTRC()
  { return on && role == TRC; }
  // are we thalamic relay cell units?

  INLINE bool   ApplyDeepMod()
  { return on && role == SUPER; }
  // should deep modulation be applied to these units?

  INLINE bool   SendDeepMod()
  { return on && role == DEEP; }
  // should we send our activation into deep_net of other (superficial) units via SendDeepModConSpec connections?

  INLINE bool   ApplyDeepCtxt()
  { return on && role == DEEP; }
  // should we apply deep context netinput?  only for deep guys

  STATE_DECO_KEY("UnitSpec");
  STATE_TA_STD_CODE_SPEC(DeepSpec);
  
  STATE_UAE(mod_range = 1.0f - mod_min;  ctxt_new = 1.0f - ctxt_prv;
                    else_new = 1.0f - else_prv; );

private:
  void        Initialize()      {  on = false; role = SUPER; Defaults_init(); }
  void        Defaults_init() {
    raw_thr_rel = 0.1f;  raw_thr_abs = 0.1f;  mod_min = 0.8f;  mod_thr = 0.1f;
    ctxt_prv = 0.0f;  tick_updt = -1;  else_prv = 0.9f;
  
    mod_range = 1.0f - mod_min;
    ctxt_new = 1.0f - ctxt_prv;
    else_new = 1.0f - else_prv;
  }
    
};


class STATE_CLASS(TRCSpec) : public STATE_CLASS(SpecMemberBase) {
  // ##INLINE ##NO_TOKENS ##CAT_Leabra specs for DeepLeabra thalamic relay cells -- engaged only for deep.on and deep.role == TRC
INHERITED(SpecMemberBase)
public:
  bool       p_only_m;          // TRC plus-phase (clamping) for TRC units only occurs if the minus phase max activation for given unit group is above .1
  bool       thal_gate;         // apply thalamic gating to TRC activations -- multiply netin by current thal parameter
  bool       binarize;          // apply threshold to deep_raw_net -- above gets bin_on, below gets bin_off -- typically used for one-to-one trc prjns with fixed wts = 1, so threshold is in terms of sending activation
  float      bin_thr;           // #CONDSHOW_ON_binarize threshold for binarizing -- typically used for one-to-one trc prjns with fixed wts = 1, so threshold is in terms of sending activation
  float      bin_on;            // #CONDSHOW_ON_binarize def:"0.3 effective netin for units above threshold -- lower value around 0.3 or so seems best
  float      bin_off;           // #CONDSHOW_ON_binarize def:"0 effective netin for units below threshold -- typically 0

  INLINE float  TRCClampNet(float deep_raw_net)
  { if(binarize)  return (deep_raw_net >= bin_thr) ? bin_on : bin_off;
    else          return deep_raw_net; }
  // compute TRC plus-phase clamp netinput
  
  STATE_DECO_KEY("UnitSpec");
  STATE_TA_STD_CODE_SPEC(TRCSpec);
  
  // STATE_UAE();

private:
  void        Initialize()
  { thal_gate = false; Defaults_init(); }

  void        Defaults_init() {
    p_only_m = false;
    binarize = false; bin_thr = 0.4f; bin_on = 0.3f; bin_off = 0.0f;
  }
};


class STATE_CLASS(DaModSpec) : public STATE_CLASS(SpecMemberBase) {
  // ##INLINE ##NO_TOKENS ##CAT_Leabra specs for effects of da-based modulation: plus-phase = learning effects
INHERITED(SpecMemberBase)
public:
  bool          on;               // whether to add dopamine factor to net input
  bool          mod_gain;         // modulate gain instead of net input
  float         minus;            // #CONDSHOW_ON_on how much to multiply da_p in the minus phase to add to netinput -- use negative values for NoGo/indirect pathway/D2 type neurons
  float         plus;             // #CONDSHOW_ON_on #AKA_gain how much to multiply da_p in the plus phase to add to netinput -- use negative values for NoGo/indirect pathway/D2 type neurons
  float         da_neg_gain;      // #CONDSHOW_ON_on&&mod_gain for negative dopamine, how much to change the default gain value as a function of dopamine: gain_eff = gain * (1 + da * da_neg_gain) -- da is multiplied by minus or plus depending on phase
  float         da_pos_gain;      // #CONDSHOW_ON_on&&mod_gain for positive dopamine, how much to change the default gain value as a function of dopamine: gain_eff = gain * (1 + da * da_pos_gain) -- da is multiplied by minus or plus depending on phase

  INLINE bool   DoDaModNetin() { return on && !mod_gain; }
  // are we doing netin modulation
  INLINE bool   DoDaModGain() { return on && mod_gain; }
  // are we doing gain modulation

  INLINE float  DaModGain(float da, float gain, bool plus_phase) {
    float da_eff = da;
    if(plus_phase)
      da_eff *= plus;
    else
      da_eff *= minus;
    if(da < 0.0f) {
      return gain * (1.0f + da_eff * da_neg_gain);
    }
    else {
      return gain * (1.0f + da_eff * da_pos_gain);
    }
  }
  // get da-modulated gain value
  
  STATE_DECO_KEY("UnitSpec");
  STATE_TA_STD_CODE_SPEC(DaModSpec);
private:
  void        Initialize()
  { on = false;  mod_gain = false;  minus = 0.0f;  plus = 0.01f;
    da_neg_gain = 0.1f; da_pos_gain = 0.1f; Defaults_init(); }
  void        Defaults_init() { };
};

class STATE_CLASS(KNaAdaptSpec) : public STATE_CLASS(SpecMemberBase) {
  // ##INLINE ##NO_TOKENS ##CAT_Leabra sodium-gated potassium channel adaptation mechanism -- evidence supports at least 3 different time constants: M-type (fast), Slick (medium), and Slack (slow)
INHERITED(SpecMemberBase)
public:
  bool          on;             // apply K-Na adaptation overall?
  float         rate_rise;      // #CONDSHOW_ON_on def:"0.8 extra multiplier for rate-coded activations on rise factors -- adjust to match discrete spiking
  bool          f_on;           // #CONDSHOW_ON_on use fast time-scale adaptation
  float         f_rise;         // #CONDSHOW_ON_on&&f_on def:"0.05 rise rate of fast time-scale adaptation as function of Na concentration -- directly multiplies -- 1/rise = tau for rise rate
  float         f_max;          // #CONDSHOW_ON_on&&f_on def:"0.1 maximum potential conductance of fast K channels -- divide nA biological value by 10 for the normalized units here
  float         f_tau;          // #CONDSHOW_ON_on&&f_on def:"50 time constant in cycles for decay of fast time-scale adaptation, which should be milliseconds typically (roughly, how long it takes for value to change significantly -- 1.4x the half-life)
  bool          m_on;           // #CONDSHOW_ON_on use medium time-scale adaptation
  float         m_rise;         // #CONDSHOW_ON_on&&m_on def:"0.02 rise rate of medium time-scale adaptation as function of Na concentration -- directly multiplies -- 1/rise = tau for rise rate
  float         m_max;          // #CONDSHOW_ON_on&&m_on def:"0.1 maximum potential conductance of medium K channels -- divide nA biological value by 10 for the normalized units here
  float         m_tau;          // #CONDSHOW_ON_on&&m_on def:"200 time constant in cycles for medium time-scale adaptation, which should be milliseconds typically (roughly, how long it takes for value to change significantly -- 1.4x the half-life)
  bool          s_on;           // #CONDSHOW_ON_on use slow time-scale adaptation
  float         s_rise;         // #CONDSHOW_ON_on&&s_on def:"0.001 rise rate of slow time-scale adaptation as function of Na concentration -- directly multiplies -- 1/rise = tau for rise rate
  float         s_max;          // #CONDSHOW_ON_on&&s_on def:"1 maximum potential conductance of slow K channels -- divide nA biological value by 10 for the normalized units here
  float         s_tau;          // #CONDSHOW_ON_on&&s_on def:"1000 time constant in cycles for slow time-scale adaptation, which should be milliseconds typically (roughly, how long it takes for value to change significantly -- 1.4x the half-life)

  float         f_dt;           // view:"-" expert:"+" rate = 1 / tau
  float         m_dt;           // view:"-" expert:"+" rate = 1 / tau
  float         s_dt;           // view:"-" expert:"+" rate = 1 / tau

  INLINE void  Compute_dKNa_spike_impl
    (bool con, bool spike, float& gc_kna, float rise, float gmax, float decay_dt)
  { if(!con )           gc_kna = 0.0f;
    else if(spike)      gc_kna += rise * (gmax - gc_kna);
    else                gc_kna -= decay_dt * gc_kna; }
  // compute the change in K channel conductance gc_kna for spiking and channel params

  INLINE void  Compute_dKNa_spike
    (bool spike, float& gc_kna_f, float& gc_kna_m, float& gc_kna_s)
  {
    Compute_dKNa_spike_impl(on && f_on, spike, gc_kna_f, f_rise, f_max, f_dt);
    Compute_dKNa_spike_impl(on && m_on, spike, gc_kna_m, m_rise, m_max, m_dt);
    Compute_dKNa_spike_impl(on && s_on, spike, gc_kna_s, s_rise, s_max, s_dt);
  }
  // update K channel conductances per params for discrete spiking

  INLINE void  Compute_dKNa_rate_impl
    (bool con, float act, float& gc_kna, float rise, float gmax, float decay_dt)
  { if(!con )   gc_kna = 0.0f;
    else        gc_kna += act * rate_rise * rise * (gmax - gc_kna) - decay_dt * gc_kna; }
  // compute the change in K channel conductance gc_kna for given activation and channel params

  INLINE void  Compute_dKNa_rate
    (float act, float& gc_kna_f, float& gc_kna_m, float& gc_kna_s) {
    Compute_dKNa_rate_impl(on && f_on, act, gc_kna_f, f_rise, f_max, f_dt);
    Compute_dKNa_rate_impl(on && m_on, act, gc_kna_m, m_rise, m_max, m_dt);
    Compute_dKNa_rate_impl(on && s_on, act, gc_kna_s, s_rise, s_max, s_dt);
  }
  // update K channel conductances per params for rate-code activation

  INLINE void   UpdtDts()
  { f_dt = 1.0f / f_tau; m_dt = 1.0f / m_tau; s_dt = 1.0f / s_tau; }
  
  STATE_DECO_KEY("UnitSpec");
  STATE_TA_STD_CODE_SPEC(KNaAdaptSpec);
  
  STATE_UAE( UpdtDts(); );
  
private:
  void        Initialize()      { on = false; Defaults_init(); }
  void        Defaults_init() {
    rate_rise = 0.8f; 
    f_on = true; f_tau = 50.0f;   f_rise = .05f;  f_max = .1f;
    m_on = true; m_tau = 200.0f;  m_rise = .02f;  m_max = .1f;
    s_on = true; s_tau = 1000.0f; s_rise = .001f; s_max = 1.0f;
    UpdtDts();
  }
};


class STATE_CLASS(KNaAdaptMiscSpec) : public STATE_CLASS(SpecMemberBase) {
  // ##INLINE ##NO_TOKENS ##CAT_Leabra extra params associated with sodium-gated potassium channel adaptation mechanism
INHERITED(SpecMemberBase)
public:
  bool          clamp;          // def:"true apply adaptation even to clamped layers -- only happens if kna_adapt.on is true
  bool          invert_nd;      // def:"true invert the adaptation effect for the act_nd (non-depressed) value that is typically used for learning-drivng averages (avg_ss, _s, _m) -- only happens if kna_adapt.on is true
  float         max_gc;         // #CONDSHOW_ON_clamp||invert_nd def:"0.2 for clamp or invert_nd, maximum k_na conductance that we expect to get (prior to multiplying by g_bar.k) -- apply a proportional reduction in clamped activation and/or enhancement of act_nd based on current k_na conductance -- default is appropriate for default kna_adapt params
  float         max_adapt;      // #CONDSHOW_ON_clamp||invert_nd has opposite effects for clamp and invert_nd (and only operative when kna_adapt.on in addition): for clamp on clamped layers, this is the maximum amount of adaptation to apply to clamped activations when conductance is at max_gc -- biologically, values around .5 correspond generally to strong adaptation in primary visual cortex (V1) -- for invert_nd, this is the maximum amount of adaptation to invert, which is key for allowing learning to operate successfully despite the depression of activations due to adaptation -- values around .2 to .4 are good for g_bar.k = .2, depending on how strongly inputs are depressed -- need to experiment to find the best value for a given config
  bool          no_targ;        // def:"true automatically exclude units in TARGET layers and also TRC (Pulvinar) thalamic neurons from adaptation effects -- typically such layers should not be subject to these effects, so this makes it easier to not have to manually set those override params

  INLINE float Compute_Clamped(float clamp_act, float gc_kna_f, float gc_kna_m, float gc_kna_s) {
    float gc_kna = gc_kna_f + gc_kna_m + gc_kna_s;
    float pct_gc = fminf(gc_kna / max_gc, 1.0f);
    return clamp_act * (1.0f - pct_gc * max_adapt);
  }
  // apply adaptation directly to a clamped activation value, reducing in proportion to amount of k_na current

  INLINE float Compute_ActNd(float act, float gc_kna_f, float gc_kna_m, float gc_kna_s) {
    float gc_kna = gc_kna_f + gc_kna_m + gc_kna_s;
    float pct_gc = fminf(gc_kna / max_gc, 1.0f);
    return act * (1.0f + pct_gc * max_adapt);
  }
  // apply inverse of adaptation to activation value, increasing in proportion to amount of k_na current

  STATE_DECO_KEY("UnitSpec");
  STATE_TA_STD_CODE_SPEC(KNaAdaptMiscSpec);
  
  // STATE_UAE( UpdtDts(); );
  
private:
  void        Initialize()      { Defaults_init(); }
  void        Defaults_init() {
    clamp = true;  invert_nd = true;  max_gc = .2f;  max_adapt = 0.3f;  no_targ = true;
  }
};


