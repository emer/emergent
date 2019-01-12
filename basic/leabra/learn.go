// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

///////////////////////////////////////////////////////////////////////
//  learn.go contains the learning params and functions for leabra

// ActAvgPars has rate constants for averaging over activations at different time scales,
// to produce the running average activation values that then drive learning in the XCAL learning rules
type ActAvgPars struct {
  SsTau float32  `def:"2;4;7"  min:"1" desc:"time constant in cycles, which should be milliseconds typically (roughly, how long it takes for value to change significantly -- 1.4x the half-life), for continuously updating the super-short time-scale avg_ss value -- this is provides a pre-integration step before integrating into the avg_s short time scale -- it is particularly important for spiking -- in general 4 is the largest value without starting to impair learning, but a value of 7 can be combined with m_in_s = 0 with somewhat worse results"`
  STau float32   `def:"2" min:"1" desc:"time constant in cycles, which should be milliseconds typically (roughly, how long it takes for value to change significantly -- 1.4x the half-life), for continuously updating the short time-scale avg_s value from the super-short avg_ss value (cascade mode) -- avg_s represents the plus phase learning signal that reflects the most recent past information"`
  MTau float32  `def:"10" min:"1" desc:"time constant in cycles, which should be milliseconds typically (roughly, how long it takes for value to change significantly -- 1.4x the half-life), for continuously updating the medium time-scale avg_m value from the short avg_s value (cascade mode) -- avg_m represents the minus phase learning signal that reflects the expectation representation prior to experiencing the outcome (in addition to the outcome) -- the default value of 10 generally cannot be exceeded without impairing learning"`
  RuLrnM float32 `def:"0.1;0" min:"0" max:"1" desc:"how much of the medium term average activation to mix in with the short (plus phase) to compute the ru_avg_s_lrn variable that is used for the receiving unit's short-term average in learning -- for DELTA_FF_FB it should be 0 -- not needed -- for XCAL_CHL this is important to ensure that when unit turns off in plus phase (short time scale), enough medium-phase trace remains so that learning signal doesn't just go all the way to 0, at which point no learning would take place -- typically need faster time constant for updating s such that this trace of the m signal is lost -- can set ss_tau=7 and set this to 0 but learning is generally somewhat worse"`
  SuLrnM float32       `def:"0.5;0.1;0" min:"0" max:"1" desc:"how much of the medium term average activation to mix in with the short (plus phase) to compute the su_avg_s_lrn variable that is used for the sending unit's short-term average in learning -- for DELTA_FF_FB delta-rule based learning, this is typically .5 (half-and-half) but for XCAL_CHL it is typically the same as ru_lrn_m (.1)"`

  SsDt float32 `view:"-" expert:"+" desc:"rate = 1 / tau"`
  SDt float32 `view:"-" expert:"+" desc:"rate = 1 / tau"`
  MDt float32 `view:"-" expert:"+" desc:"rate = 1 / tau"`
  RuLrnS float32 `view:"-" expert:"+" desc:"1-ru_lrn_m"`
  SuLrnS float32 `view:"-" expert:"+" desc:"1-su_lrn_m"`
}

// AvgsFmAct computes averages based on current act, and overall dt integration factor
func (aa *ActAvgPars) AvgsFmAct(ruAct, dtInteg float32, avgSS, avgS, avgM, ruAvgSlrn, suAvgSlrn *float32) {
    avgSS += dtInteg * aa.SsDt * (ruAct - avgSS)
    avgS +=  dtInteg * aa.SDt * (avgSS - avgS)
    avgM +=  dtInteg * aa.MDt * (avgS - avgM)

    ruAvgSLrn = aa.RuLrnS * avgS + aa.RuLrnM * avgM
    suAvgSlrn = aa.RuLrnS * avgS + aa.SuLrnM * avgM
  }

func (aa *ActAvgPars) Update() {
  aa.SSDt = 1 / aa.SSTau
  aa.SDt = 1 / aa.STau
  aa.MDt = 1/ aa.MTau
  aa.RuLrnS = 1 - aa.RuLrnM
 aa.SuLrnS= 1 - aa.SuLrnM
}

func (aa *ActAvgPars) Defaults() {  
  aa.SSTau = 4.0
  aa.STau = 2.0
  aa.MTau = 10.0
  aa.RuLrnM = 0.1
  aa.SuLrnM = 0.1
  aa.Update()

}

// AvgLPars are parameters for computing the long-term floating average value, avg_l
// which is used for driving BCM-style hebbian learning in XCAL -- this form of learning
// increases contrast of weights and generally decreases overall activity of neuron,
// to prevent "hog" units -- it is computed as a running average of the (gain multiplied)
// medium-time-scale average activation at the end of the trial.
// Also computes an adaptive amount of BCM learning, avg_l_lrn, based on avg_l.
type AvgLPars struct {
  Init float32           `def:"0.4" min:"0" max:"1" desc:"initial avg_l value at start of training"`
  Gain float32           `def:"1.5;2;2.5;3;4;5" min:"0" desc:"gain multiplier on activation used in computing the running average avg_l value that is the key floating threshold in the BCM Hebbian learning rule -- when using the DELTA_FF_FB learning rule, it should generally be 2x what it was before with the old XCAL_CHL rule, i.e., default of 5 instead of 2.5 -- it is a good idea to experiment with this parameter a bit -- the default is on the high-side, so typically reducing a bit from initial default is a good direction"`
  float         min;            // def:"0.2 min:"0 miniumum avg_l value -- running average cannot go lower than this value even when it otherwise would due to inactivity -- this value is generally good and typically does not need to be changed
  float         tau;            // def:"10 min:"1 time constant for updating the running average avg_l -- avg_l moves toward gain*act with this time constant on every trial - longer time constants can also work fine, but the default of 10 allows for quicker reaction to beneficial weight changes
  float         lrn_max;        // def:"0.5 min:"0 maximum avg_l_lrn value, which is amount of learning driven by avg_l factor -- when avg_l is at its maximum value (i.e., gain, as act does not exceed 1), then avg_l_lrn will be at this maximum value -- by default, strong amounts of this homeostatic Hebbian form of learning can be used when the receiving unit is highly active -- this will then tend to bring down the average activity of units -- the default of 0.5, in combination with the err_mod flag, works well for most models -- use around 0.0004 for a single fixed value (with err_mod flag off)
  float         lrn_min;        // def:"0.0001;0.0004 min:"0 miniumum avg_l_lrn value (amount of learning driven by avg_l factor) -- if avg_l is at its minimum value, then avg_l_lrn will be at this minimum value -- neurons that are not overly active may not need to increase the contrast of their weights as much -- use around 0.0004 for a single fixed value (with err_mod flag off)
  float         lay_act_thr;    // def:"0.01 threshold of layer average activation on this trial, in order to update avg_l values -- setting to 0 disables this check
  
  float         dt;             // view:"-" expert:"+" rate = 1 / tau
  float         min_lay_avg;    // view:"-" expert:"+" lay_avg_trg / max_gain_mult
  float         lrn_fact;       // view:"-" expert:"+" (lrn_max - lrn_min) / (avg_l_max - min)

  INLINE void   UpdtAvgL(float& avg_l, const float act, float lay_avg) {
    avg_l += dt * (gain * act - avg_l);
    if(avg_l < min) avg_l = min;
  }
  // update long-term average value from given activation, using average-based update

  INLINE float  GetLrn(const float avg_l) {
    return lrn_min + lrn_fact * (avg_l - min);
  }
  // get the avg_l_lrn value for given avg_l value

  INLINE void UpdtVals() {
    dt = 1.0f / tau;
    lrn_fact = (lrn_max - lrn_min) / (gain - min);
  }
  // #IGNORE
  
  STATE_DECO_KEY("UnitSpec");
  STATE_TA_STD_CODE_SPEC(LeabraAvgLSpec);
  
  STATE_UAE( UpdtVals(); );
  
private:
  void        Initialize()      { Defaults_init(); }
  void        Defaults_init() {
    init = 0.4f;        gain = 2.5f;            min = 0.2f;
    tau = 10.0f;        lay_act_thr = 0.01f;
    lrn_max = 0.5f;     lrn_min = 0.0001f;
    UpdtVals();
  }
};


class STATE_CLASS(LeabraAvgLModSpec) : public STATE_CLASS(SpecMemberBase) {
  // ##INLINE ##NO_TOKENS ##CAT_Leabra parameters for modulating the learning rate for BCM-style hebbian learning in XCAL by other factors in the network
INHERITED(SpecMemberBase)
public:
  enum ModMode {                     // #BITS how to modulate BCM hebbian learning
    NO_MOD = 0x00,                   // #NO_BIT don't use any modulation of BCM hebbian learning
    LAY_ERR_MOD = 0x01,              // decrease in proportion to layer-level cos_diff_avg_lrn, to make hebbian term roughly proportional to amount of error driven learning signal across layers -- cos_diff_avg computes the running average of the cos diff value between act_m and act_p (no diff is 1, max diff is 0), and cos_diff_avg_lrn = 1 - cos_diff_avg (and 0 for non-HIDDEN layers) -- see LeabraLayerSpec cos_diff.avg_tau rate constant for integrating cos_diff_avg value -- this adjusts amount of BCM hebbian per layer -- generally useful
    NET_ERR_MOD = 0x02,             // decrease in proportion to (1 - avg_cos_err) -- as network performs better (avg_cos_err goes closer to 1), less hebbian learning is applied -- has not generally been useful but could possibly be
  };
  
  ModMode       mod;            // def:"LAY_ERR_MOD whether and how to modulate amount of BCM hebbian learning as function of other variables
  float         lay_mod_min;    // def:"0.01 #CONDSHOW_ON_mod:LAY_ERR_MOD minimum mod_avg_l_lrn modulation value for LAY_ERR_MOD -- ensures a minimum amount of self-organizing learning even for network / layers that have a very small level of error signal
  float         net_mod_min;    // def:"0.5 #CONDSHOW_ON_mod:NET_ERR_MOD minimum mod_avg_l_lrn modulation value for NET_ERR_MOD -- ensures a minimum amount of self-organizing learning even when error is low
  float         net_err_start;  // def:"0.5 #CONDSHOW_ON_mod:NET_ERR_MOD for NET_ERR_MOD, do not start decreasing the amount of hebbian learning until the network avg_cos_err has gone above this value (higher is better, 1 = perfect) -- beneficial to wait until learning as progressed well before starting to back off on hebbian -- modulation effect goes from 1 to mod_min within the range of avg_cos_err between err_start and err_end
  float         net_err_end;    // def:"1 #CONDSHOW_ON_mod:NET_ERR_MOD for NET_ERR_MOD, upper level of network avg_cos_err (higher is better, 1 = perfect) where the modulatory factor will reach mod_min (and stay there) -- can set less than 1 if model doesn't fully converge to get full range of modulation

  INLINE float  GetMod(float lay_cos_diff_avg_lrn, float net_avg_cos_err) {
    float mod_avg_l_lrn = 1.0f;
    if(mod & LAY_ERR_MOD) {
      mod_avg_l_lrn *= fmaxf(lay_cos_diff_avg_lrn, lay_mod_min);
    }
    if(mod & NET_ERR_MOD) {
      if(net_avg_cos_err > net_err_start) {
        if(net_avg_cos_err >= net_err_end) {
          mod_avg_l_lrn *= net_mod_min;
        }
        else {
          float emod = ((net_avg_cos_err - net_err_start) / (net_err_end - net_err_start)); // 0..1
          emod = fminf(1.0f, emod);
          mod_avg_l_lrn *= (1.0f - emod * (1.0f - net_mod_min));
        }
      }
    }
    return mod_avg_l_lrn;
  }
  // get the mod_avg_l_lrn modulation factor -- called in LayerSpec Compute_CosDiff, where cos_diff_avg_lrn is computed -- sets layer mod_avg_l_lrn variable

  INLINE void   UpdtVals() {
    if(net_err_start <= 0.0f || net_err_start >= 1.0f) net_err_start = 0.5f; // cant be zero or 1
    if(net_err_end < net_err_start) net_err_end = 1.0f;
  }
  // #IGNORE update the learn factor
  
  STATE_DECO_KEY("UnitSpec");
  STATE_TA_STD_CODE_SPEC(LeabraAvgLModSpec);
  STATE_UAE( UpdtVals(); );
  
private:
  void        Initialize()      { Defaults_init(); }
  void        Defaults_init() {
    mod = NO_MOD;  lay_mod_min = 0.01f;
    net_mod_min = 0.5f; net_err_start = 0.5f;  net_err_end = 1.0f;
  }

};


