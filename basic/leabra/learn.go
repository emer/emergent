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

// leabra.LearnNeuron manages learning-related parameters at the neuron-level.
// This is mainly the running average activations that drive learning
type LearnNeuron struct {
	ActAvg ActAvgPars `inline:"+" desc:"parameters for computing running average activations that drive learning"`
	AvgL   AvgLPars   `inline:"+" desc:"parameters for computing AvgL long-term running average"`
}

// ActAvgInit initializes average activation values.
// Called at start of learning.
func (ln *LearnNeuron) ActAvgInit(nrn *Neuron) {
	nrn.AvgSS = ln.ActAvg.Init
	nrn.AvgS = ln.ActAvg.Init
	nrn.AvgM = ln.ActAvg.Init
	nrn.AvgL = ln.AvgL.Init
	nrn.AvgSLrn = 0
}

// AvgsFmAct updates the running averages based on current activation.
// Computed after new activation for current cycle is updated.
func (ln *LearnNeuron) AvgsFmAct(nrn *Neuron) {
	aa.ActAvg.AvgsFmAct(nrn.Act, &nrn.AvgSS, &nrn.AvgS, &nrn.AvgM, &nrn.AvgSLrn)
}

// AvgLFmAct computes long-term average activation value, and learning factor, from given activation.
// Called at start of new trial.
func (ln *LearnNeuron) AvgLFmAct(nrn *Neuron) {
	aa.AvgL.AvgLFmAct(nrn.Act, &nrn.AvgL, &nrn.AvgLLrn)
	// todo: layer-level err mod needs to be added in
}

///////////////////////////////////////////////////////////////////////
//  LearnSyn

// leabra.LearnSyn manages learning-related parameters at the synapse-level.
type LearnSyn struct {
	WtInit   erand.RndPars `inline:"+" desc:"initial random weight distribution"`
	XCal     XCalPars      `inline:"+" desc:"parameters for the XCal learning rule"`
	WtSig    WtSigPars     `inline:"+" desc:"parameters for the sigmoidal contrast weight enhancement"`
	DWtNorm  DWtNormPars   `inline:"+" desc:"parameters for normalizing weight changes by abs max dwt"`
	Momentum MomentumPars  `inline:"+" desc:"parameters for momentum across weight changes"`
	WtBal    WtBalPars     `inline:"+" desc:"parameters for balancing strength of weight increases vs. decreases"`
}

func (ls *LearnSyn) Defaults() {
	ls.WtInit.Mean = 0.5
	ls.WtInit.Var = 0.25
	ls.WtInit.Dist = erand.Uniform
}

func (ls *LearnSyn) InitWts(syn *Synapse) {
	//    Init_Weights_symflag(net, thr_no);
	//    LEABRA_CON_STATE* cg = (LEABRA_CON_STATE*)pcg;
	//
	//    cg->err_dwt_max = 0.0f;    cg->bcm_dwt_max = 0.0f; cg->dwt_max = 0.0f;
	//    cg->wb_inc = 1.0f;         cg->wb_dec = 1.0f;
	//

	//    float* wts = cg->OwnCnVar(WT);
	//    float* dwts = cg->OwnCnVar(DWT);
	//    float* scales = cg->OwnCnVar(SCALE);
	//    // NOTE: it is ESSENTIAL that Init_Weights ONLY does wt, dwt, and scale -- all other vars
	//    // MUST be initialized in post -- projections with topo weights ONLY do these specific
	//    // variables but no others..
	//
	//    int eff_thr_no = net->HasNetFlag(NETWORK_STATE::INIT_WTS_1_THREAD) ? 0 : thr_no;
	//
	//    const int sz = cg->size;
	//    for(int i=0; i<sz; i++) {
	//      scales[i] = 1.0f;         // default -- must be set in prjn spec if different
	//    }
	//
	//    for(int i=0; i<sz; i++) {
	//      if(rnd.type != STATE_CLASS(Random)::NONE) {
	//        C_Init_Weight_Rnd(wts[i], eff_thr_no);
	//      }
	//      C_Init_dWt(dwts[i]);
	//    }
}

/*
  INLINE void Init_Weights_scale(CON_STATE* rcg, NETWORK_STATE* net, int thr_no, float init_wt_val) override {
    Init_Weights_symflag(net, thr_no);

    // this is called *receiver based*!!!

    int eff_thr_no = net->HasNetFlag(NETWORK_STATE::INIT_WTS_1_THREAD) ? 0 : thr_no;

    const int sz = rcg->size;
    for(int i=0; i<sz; i++) {
      if(rnd.type != STATE_CLASS(Random)::NONE) {
        C_Init_Weight_Rnd(rcg->PtrCn(i, SCALE, net), eff_thr_no);
      }
      rcg->PtrCn(i, WT, net) = init_wt_val;
      C_Init_dWt(rcg->PtrCn(i, DWT, net));
    }
  }

  INLINE void  ApplySymmetry_s(CON_STATE* cg, NETWORK_STATE* net, int thr_no) override {
    if(!wt_limits.sym) return;
    UNIT_STATE* su = cg->ThrOwnUnState(net, thr_no);
    const int sz = cg->size;
    for(int i=0; i<sz;i++) {
      int con_idx = -1;
      CON_STATE* rscg = net->FindRecipSendCon(con_idx, cg->UnState(i,net), su);
      if(rscg && con_idx >= 0) {
        CON_SPEC_CPP* rscs = rscg->GetConSpec(net);
        if(rscs && rscs->wt_limits.sym) {
          if(wt_limits.sym_fm_top) {
            cg->OwnCn(i, WT) = rscg->OwnCn(con_idx, WT);
            cg->OwnCn(i, SCALE) = rscg->OwnCn(con_idx, SCALE); // only diff: sync scales!
          }
          else {
            rscg->OwnCn(con_idx, WT) = cg->OwnCn(i, WT);
            rscg->OwnCn(con_idx, SCALE) = cg->OwnCn(i, SCALE);
          }
        }
      }
    }
  }

  INLINE void Init_Weights_post(CON_STATE* pcg, NETWORK_STATE* net, int thr_no) override {
    LEABRA_CON_STATE* cg = (LEABRA_CON_STATE*)pcg;
    cg->Init_ConState();

    float* wts = cg->OwnCnVar(WT);
    float* swts = cg->OwnCnVar(SWT);
    float* fwts = cg->OwnCnVar(FWT);
    float* scales = cg->OwnCnVar(SCALE);
    float* dwnorms = cg->OwnCnVar(DWNORM);
    float* moments = cg->OwnCnVar(MOMENT);
    float* wbincs = cg->OwnCnVar(WB_INC);
    float* wbdecs = cg->OwnCnVar(WB_DEC);
    for(int i=0; i<cg->size; i++) {
      fwts[i] = LinFmSigWt(wts[i]); // swt, fwt are linear underlying weight values
      dwnorms[i] = 0.0f;
      moments[i] = 0.0f;
      swts[i] = fwts[i];
      wts[i] *= scales[i];
      wbincs[i] = wbdecs[i] = 1.0f;

      LEABRA_CON_STATE* rcg = cg->UnCons(i, net);
      rcg->Init_ConState();    // recv based otherwise doesn't get initialized!
    }
  }

  INLINE void  LoadWeightVal(float wtval, CON_STATE* cg, int cidx, NETWORK_STATE* net) override {
    cg->Cn(cidx, WT, net) = wtval;
    float linwt = LinFmSigWt(wtval / cg->Cn(cidx, SCALE, net));
    cg->Cn(cidx, SWT, net) = linwt;
    cg->Cn(cidx, FWT, net) = linwt;
  }

  INLINE void SetConScale(float scale, CON_STATE* cg, int cidx, NETWORK_STATE* net, int thr_no) override {
    cg->Cn(cidx, SCALE, net) = scale;
  }

  INLINE virtual void  Trial_Init_Specs(LEABRA_NETWORK_STATE* net) {
    if(learn) {
      if(wt_bal.on) {
        net->net_misc.wt_bal = true;
      }
      if(dwt_norm.RecvConsAgg()) {
        net->net_misc.recv_con_dwnorm = true;
      }
    }
  }
  // #CAT_Learning initialize specs and specs update network flags -- e.g., set current learning rate based on schedule given epoch (or error value)

  INLINE void  RenormScales(CON_STATE* cg, NETWORK_STATE* net, int thr_no, bool mult_norm,
                            float avg_wt) override {
    const int sz = cg->size;
    if(sz < 2) return;
    float avg = 0.0f;
    for(int i=0; i<sz; i++) {
      avg += cg->Cn(i, SCALE, net);
    }
    avg /= (float)sz;
    if(mult_norm) {
      float adj = avg_wt / avg;
      for(int i=0; i<sz; i++) {
        cg->Cn(i, SCALE, net) *= adj;
      }
    }
    else {
      float adj = avg_wt - avg;
      for(int i=0; i<sz; i++) {
        cg->Cn(i, SCALE, net) += adj;
      }
    }
  }


  ///////////////////////////////////////////////////////////////
  //    Activation: Netinput -- only NetinDelta is supported

  INLINE virtual bool  DoesStdNetin() { return true; }
  // #IGNORE does this connection send standard netinput? if so, it will be included in the CUDA send netin computation -- otherwise a separate function is required
  INLINE virtual bool  DoesStdDwt() { return true; }
  // #IGNORE does this connection compute a standard XCAL dWt function? if so, it will be included in the CUDA Compute_dWt computation -- otherwise a separate function is required
  INLINE virtual bool  IsMarkerCon() { return false; }
  // #IGNORE is this a marker con (MarkerConSpec) -- optimized check for higher speed
  INLINE virtual bool  IsDeepCtxtCon() { return false; }
  // #IGNORE is this a deep context connection (DeepCtxtConSpec) -- optimized check for higher speed
  INLINE virtual bool  IsDeepRawCon() { return false; }
  // #IGNORE is this a send deep_raw connection (SendDeepRawConSpec) -- optimized check for higher speed
  INLINE virtual bool  IsDeepModCon() { return false; }
  // #IGNORE is this a send deep_mod connection (SendDeepModConSpec) -- optimized check for higher speed

  INLINE void   C_Send_NetinDelta(const float wt, float* send_netin_vec,
                                  const int ru_idx, const float su_act_delta_eff)
  { send_netin_vec[ru_idx] += wt * su_act_delta_eff; }
  // #IGNORE
#ifdef TA_VEC_USE
  INLINE void   Send_NetinDelta_vec(LEABRA_CON_STATE* cg, const float su_act_delta_eff,
                                    float* send_netin_vec, const float* wts) {
    VECF sa(su_act_delta_eff);
    const int sz = cg->size;
    const int parsz = cg->vec_chunked_size;
    int i;
    for(i=0; i<parsz; i += TA_VEC_SIZE) {
      VECF wt;  wt.load(wts+i);
      VECF dp = wt * sa;
      VECF rnet;
      float* stnet = send_netin_vec + cg->UnIdx(i);
      rnet.load(stnet);
      rnet += dp;
      rnet.store(stnet);
    }

    // remainder of non-vector chunkable ones
    for(; i<sz; i++) {
      send_netin_vec[cg->UnIdx(i)] += wts[i] * su_act_delta_eff;
    }
  }
  // #IGNORE vectorized version
#endif
  INLINE void   Send_NetinDelta_impl(LEABRA_CON_STATE* cg, LEABRA_NETWORK_STATE* net,
                                     int thr_no, const float su_act_delta, const float* wts)  {
    LEABRA_PRJN_STATE* prjn = cg->GetPrjnState(net);
    const float su_act_delta_eff = prjn->scale_eff * su_act_delta;
    if(net->NetinPerPrjn()) {
      float* send_netin_vec = net->ThrSendNetinTmpPerPrjn(thr_no, cg->other_idx);
#ifdef TA_VEC_USE
      Send_NetinDelta_vec(cg, su_act_delta_eff, send_netin_vec, wts);
#else
      CON_STATE_LOOP(cg, C_Send_NetinDelta(wts[i], send_netin_vec,
                                           cg->UnIdx(i), su_act_delta_eff));
#endif
    }
    else {
      float* send_netin_vec = net->ThrSendNetinTmp(thr_no);
#ifdef TA_VEC_USE
      Send_NetinDelta_vec(cg, su_act_delta_eff, send_netin_vec, wts);
#else
      CON_STATE_LOOP(cg, C_Send_NetinDelta(wts[i], send_netin_vec,
                                           cg->UnIdx(i), su_act_delta_eff));
#endif
    }
  }

  // #IGNORE implementation that uses specified weights -- typically only diff in different subclasses is the weight variables used
  INLINE virtual void   Send_NetinDelta(LEABRA_CON_STATE* cg, LEABRA_NETWORK_STATE* net,
                                        int thr_no, const float su_act_delta) {
    // note: _impl is used b/c subclasses replace WT var with another variable
    Send_NetinDelta_impl(cg, net, thr_no, su_act_delta, cg->OwnCnVar(WT));
  }
  // #IGNORE #CAT_Activation sender-based delta-activation net input for con group (send net input to receivers) -- always goes into tmp matrix (thr_no >= 0!) and is then integrated into net through Compute_NetinInteg function on units

  // recv-based also needed for some statistics, but is NOT used for main compute code -- uses act_eq for sender act as well
  INLINE float  C_Compute_Netin(const float wt, const float su_act)
  { return wt * su_act; }
  // #IGNORE NOTE: doesn't work with spiking -- need a separate function to use act_eq for that case -- using act_eq does NOT work with scalarval etc
  INLINE float  Compute_Netin(CON_STATE* rcg, NETWORK_STATE* net, int thr_no) override  {
    LEABRA_CON_STATE* cg = (LEABRA_CON_STATE*)rcg;
    LEABRA_PRJN_STATE* prjn = cg->GetPrjnState(net);
    // this is slow b/c going through the PtrCn
    float rval=0.0f;
    CON_STATE_LOOP(cg, rval += C_Compute_Netin(cg->PtrCn(i,WT,net),
                                               cg->UnState(i,net)->act));
    return prjn->scale_eff * rval;
  }
  // #IGNORE

  ///////////////////////////////////////////////////////////////
  //    Learning

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

  INLINE void  C_Compute_dWt_CtLeabraXCAL_Expt
  (float& err, float& bcm, float ru_ru_avg_s_lrn, float ru_su_avg_s_lrn, float ru_avg_m,
   float su_su_avg_s_lrn, float su_ru_avg_s_lrn, float su_avg_m,
   float ru_avg_l, float wt_lin)
  {
    float srs = su_su_avg_s_lrn * ru_ru_avg_s_lrn;
    float srm = su_avg_m * ru_avg_m;

    switch(rule.bcmrule) {
    case STATE_CLASS(LeabraLearnSpec)::SRS:
      bcm = xcal.dWtFun(srs, ru_avg_l);
      break;
    case STATE_CLASS(LeabraLearnSpec)::RS:
      bcm = su_su_avg_s_lrn * xcal.dWtFun(ru_ru_avg_s_lrn, ru_avg_l);
      break;
    case STATE_CLASS(LeabraLearnSpec)::RS_SIN:
      bcm = xcal.dWtFun(su_su_avg_s_lrn * ru_ru_avg_s_lrn, su_su_avg_s_lrn * ru_avg_l);
      break;
    case STATE_CLASS(LeabraLearnSpec)::CPCA:
      bcm = ru_ru_avg_s_lrn * ((rule.cp_gain * su_su_avg_s_lrn) - wt_lin);
      break;
    }

    switch(rule.errule) {
    case STATE_CLASS(LeabraLearnSpec)::ERR_DELTA_FF_FB:
      if(feedback) {
        err = ru_su_avg_s_lrn * (su_ru_avg_s_lrn - su_avg_m);
      }
      else {
        err = su_su_avg_s_lrn * (ru_ru_avg_s_lrn - ru_avg_m);
      }
      break;
    case STATE_CLASS(LeabraLearnSpec)::XCAL:
      err = xcal.dWtFun(srs, srm);
      break;
    case STATE_CLASS(LeabraLearnSpec)::DELTA:
      err = su_su_avg_s_lrn * (ru_ru_avg_s_lrn - ru_avg_m);
      break;
    case STATE_CLASS(LeabraLearnSpec)::XCAL_DELTA:
      err = su_su_avg_s_lrn * xcal.dWtFun(ru_ru_avg_s_lrn, ru_avg_m);
      break;
    case STATE_CLASS(LeabraLearnSpec)::XCAL_DELTA_SIN:
      err = xcal.dWtFun(su_su_avg_s_lrn * ru_ru_avg_s_lrn, su_su_avg_s_lrn * ru_avg_m);
      break;
    case STATE_CLASS(LeabraLearnSpec)::REV_DELTA:
      err = ru_su_avg_s_lrn * (su_ru_avg_s_lrn - su_avg_m);
      break;
    case STATE_CLASS(LeabraLearnSpec)::REV_XCAL_DELTA:
      err = ru_su_avg_s_lrn * xcal.dWtFun(su_ru_avg_s_lrn, su_avg_m);
      break;
    case STATE_CLASS(LeabraLearnSpec)::REV_XCAL_DELTA_SIN:
      err = xcal.dWtFun(ru_su_avg_s_lrn * su_ru_avg_s_lrn, ru_su_avg_s_lrn * su_avg_m);
      break;
    case STATE_CLASS(LeabraLearnSpec)::CHL:
      err = srs - srm;
      break;
    }
  }
  // #IGNORE compute temporally eXtended Contrastive Attractor Learning (XCAL), experimental version, returning new dwt


  INLINE void  C_Compute_dWt_CtLeabraXCAL_CHL
  (float& err, float& bcm, float ru_ru_avg_s_lrn, float ru_avg_m, float su_su_avg_s_lrn,
   float su_avg_m, float ru_avg_l)
  {
    float srs = su_su_avg_s_lrn * ru_ru_avg_s_lrn;
    float srm = su_avg_m * ru_avg_m;

    bcm = xcal.dWtFun(srs, ru_avg_l);
    err = xcal.dWtFun(srs, srm);
  }
  // #IGNORE compute temporally eXtended Contrastive Attractor Learning (XCAL), CHL version, returning new dwt

  INLINE void  C_Compute_dWt_CtLeabraXCAL_Delta
  (float& err, float& bcm, float ru_ru_avg_s_lrn, float ru_su_avg_s_lrn, float ru_avg_m,
   float su_su_avg_s_lrn, float su_ru_avg_s_lrn, float su_avg_m,
   float ru_avg_l)
  {
    bcm = xcal.dWtFun(su_su_avg_s_lrn * ru_ru_avg_s_lrn, su_su_avg_s_lrn * ru_avg_l);
    if(feedback) {
      err = ru_su_avg_s_lrn * (su_ru_avg_s_lrn - su_avg_m);
    }
    else {
      err = su_su_avg_s_lrn * (ru_ru_avg_s_lrn - ru_avg_m);
    }
  }
  // #IGNORE compute temporally eXtended Contrastive Attractor Learning (XCAL), DELTA_FF_FB version, returning new dwt

  INLINE float  C_Compute_dWt_CtLeabraXCAL_MarginSign(float ru_margin, float su_su_avg_s_lrn) {
    return margin.sign_lrn * margin.SignDwt(ru_margin) * su_su_avg_s_lrn;
  }
  // #IGNORE margin sign_dwt -- only if margin.sign_dwt


  INLINE void  C_Compute_dWt_CtLeabraXCAL
  (float& err, float& bcm, float ru_ru_avg_s_lrn, float ru_su_avg_s_lrn, float ru_avg_m,
   float su_su_avg_s_lrn, float su_ru_avg_s_lrn, float su_avg_m, float ru_avg_l, float wt_lin)
  {
    switch(rule.rule) {
    case STATE_CLASS(LeabraLearnSpec)::DELTA_FF_FB:
      C_Compute_dWt_CtLeabraXCAL_Delta
        (err, bcm, ru_ru_avg_s_lrn, ru_su_avg_s_lrn, ru_avg_m, su_su_avg_s_lrn,
         su_ru_avg_s_lrn, su_avg_m, ru_avg_l);
      break;
    case STATE_CLASS(LeabraLearnSpec)::XCAL_CHL:
      C_Compute_dWt_CtLeabraXCAL_CHL
        (err, bcm, ru_ru_avg_s_lrn, ru_avg_m, su_su_avg_s_lrn, su_avg_m, ru_avg_l);
      break;
    case STATE_CLASS(LeabraLearnSpec)::EXPT:
      C_Compute_dWt_CtLeabraXCAL_Expt
        (err, bcm, ru_ru_avg_s_lrn, ru_su_avg_s_lrn, ru_avg_m, su_su_avg_s_lrn,
         su_ru_avg_s_lrn, su_avg_m, ru_avg_l, wt_lin);
      break;
    }
  }
  // #IGNORE compute temporally eXtended Contrastive Attractor Learning (XCAL)

  INLINE void   Compute_dWt(CON_STATE* scg, NETWORK_STATE* rnet, int thr_no) override  {
    LEABRA_NETWORK_STATE* net = (LEABRA_NETWORK_STATE*)rnet;
    if(!learn || (use_unlearnable && net->unlearnable_trial)) return;
    LEABRA_CON_STATE* cg = (LEABRA_CON_STATE*)scg;
    LEABRA_UNIT_STATE* su = cg->ThrOwnUnState(net, thr_no);
    if(su->avg_s < xcal.lrn_thr && su->avg_m < xcal.lrn_thr) return;
    // no need to learn!

    float clrate, bg_lrate, fg_lrate;
    bool deep_on;
    GetLrates(cg, net, thr_no, clrate, deep_on, bg_lrate, fg_lrate);

    const float su_su_avg_s_lrn = su->su_avg_s_lrn;
    const float su_ru_avg_s_lrn = su->ru_avg_s_lrn;
    const float su_avg_s = su->avg_s;
    const float su_avg_m = su->avg_m;
    const int sz = cg->size;

    LEABRA_PRJN_STATE* prjn = cg->GetPrjnState(net);
    if(momentum.on) {
      clrate *= momentum.lr_comp;
    }

    float err_dwt_max = 0.0f;
    float bcm_dwt_max = 0.0f;
    float dwt_max = 0.0f;
    float err_dwt_avg = 0.0f;
    float bcm_dwt_avg = 0.0f;
    float dwt_avg = 0.0f;

    float* dwts = cg->OwnCnVar(DWT);
    float* fwts = cg->OwnCnVar(FWT);
    float* dwnorms = cg->OwnCnVar(DWNORM);
    float* moments = cg->OwnCnVar(MOMENT);

    for(int i=0; i<sz; i++) {
      LEABRA_UNIT_STATE* ru = cg->UnState(i, net);
      if(ru->lesioned()) continue;
      float lrate_eff = clrate;
      if(deep_on) {
        lrate_eff *= (bg_lrate + fg_lrate * ru->deep_lrn);
      }
      if(margin.lrate_mod) {
        lrate_eff *= margin.MarginLrate(ru->margin);
      }
      float l_lrn_eff = xcal.LongLrate(ru->avg_l_lrn);
      float err, bcm;
      C_Compute_dWt_CtLeabraXCAL
        (err, bcm, ru->ru_avg_s_lrn, ru->su_avg_s_lrn, ru->avg_m,
         su_su_avg_s_lrn, su_ru_avg_s_lrn, su_avg_m, ru->avg_l, fwts[i]);

      if(margin.sign_dwt) {
        bcm += C_Compute_dWt_CtLeabraXCAL_MarginSign(ru->margin, su_su_avg_s_lrn);
      }

      bcm *= l_lrn_eff;
      err *= xcal.m_lrn;

      float abserr = fabsf(err);
      if(dwt_norm.stats) {
        float absbcm = fabsf(bcm);
        err_dwt_max = fmaxf(abserr, err_dwt_max);
        bcm_dwt_max = fmaxf(absbcm, bcm_dwt_max);
        err_dwt_avg += abserr;
        bcm_dwt_avg += absbcm;
      }

      float new_dwt = bcm + err;
      float norm = 1.0f;
      if(dwt_norm.on) {
        norm = dwt_norm.ComputeNorm(dwnorms[i], fabsf(new_dwt)); // always update
      }

      if(momentum.on) {
        // apparently quite important for norm to be applied to post-momentum dwt
        new_dwt = norm * momentum.ComputeMoment(moments[i], new_dwt);
      }
      else {
        new_dwt *= norm;
      }
      dwts[i] += lrate_eff * new_dwt;

      if(dwt_norm.stats) {
        float absdwt = fabsf(new_dwt);
        dwt_max = fmaxf(absdwt, dwt_max);
        dwt_avg += absdwt;
      }
    }

    if(dwt_norm.stats) {
      cg->err_dwt_max = err_dwt_max;
      cg->bcm_dwt_max = bcm_dwt_max;
      cg->dwt_max = dwt_max;

      if(sz > 0) {
        float nrm = 1.0f / (float)sz;
        cg->err_dwt_avg = err_dwt_avg * nrm;
        cg->bcm_dwt_avg = bcm_dwt_avg * nrm;
        cg->dwt_avg = dwt_avg * nrm;
      }
    }

    if(dwt_norm.SendConsAgg()) {
      DwtNorm_SendCons(cg, net, thr_no);
    }
  }


  INLINE void   C_Compute_Weights_CtLeabraXCAL
    (float& wt, float dwt, float& fwt, float& swt, float& scale,
     const float wb_inc, const float wb_dec, int thr_no)
  {
    if(dwt == 0.0f) return;
    if(wt_sig.soft_bound) {
      if(dwt > 0.0f)    dwt *= wb_inc * (1.0f - fwt);
      else              dwt *= wb_dec * fwt;
    }
    else {
      if(dwt > 0.0f)    dwt *= wb_inc;
      else              dwt *= wb_dec;
    }
    fwt += dwt;
    C_ApplyLimits(fwt);
    // swt = fwt;  // leave swt as pristine original weight value -- saves time
    // and is useful for visualization!
    wt = scale * SigFmLinWt(fwt);
    // dwt = 0.0f;

    if(adapt_scale.on) {
      adapt_scale.AdaptWtScale(scale, wt);
    }
  }
  // #IGNORE overall compute weights for CtLeabraXCAL learning rule -- no slow wts

  INLINE void   C_Compute_Weights_CtLeabraXCAL_slow
    (float& wt, float dwt, float& fwt, float& swt, float& scale,
     const float wb_inc, const float wb_dec, int thr_no)
  {
    if(wt_sig.soft_bound) {
      if(dwt > 0.0f)    dwt *= wb_inc * (1.0f - fwt);
      else              dwt *= wb_dec * fwt;
    }
    else {
      if(dwt > 0.0f)    dwt *= wb_inc;
      else              dwt *= wb_dec;
    }
    fwt += dwt;
    float eff_wt = slow_wts.swt_pct * swt + slow_wts.fwt_pct * fwt;
    float nwt = scale * SigFmLinWt(eff_wt);
    wt += slow_wts.wt_dt * (nwt - wt);
    swt += slow_wts.slow_dt * (fwt - swt);
    // dwt = 0.0f;

    if(adapt_scale.on) {
      adapt_scale.AdaptWtScale(scale, wt);
    }
  }
  // #IGNORE overall compute weights for CtLeabraXCAL learning rule -- slow wts


  INLINE float C_Compute_Weights_dwtshare
  (bool dwt_sh, float* dwts, const int i, const int neigh, const int sz) {
    if(dwt_sh) {
      float dwt = 0.0f;
      for(int ni = -neigh; ni <= neigh; ni++) {
        int j = i + ni;
        if(j < 0)         j += sz;
        else if(j >= sz)  j -= sz;
        dwt += dwts[j];
      }
      return dwt;
    }
    else {
      return dwts[i];
    }
  }
  // #IGNORE do dwt sharing or just dwt, depending on dwt_sh

  INLINE void   Compute_Weights(CON_STATE* scg, NETWORK_STATE* net, int thr_no) override {
    if(!learn) return;
    LEABRA_CON_STATE* cg = (LEABRA_CON_STATE*)scg;
    float* wts = cg->OwnCnVar(WT);      float* dwts = cg->OwnCnVar(DWT);
    float* fwts = cg->OwnCnVar(FWT);    float* swts = cg->OwnCnVar(SWT);
    float* scales = cg->OwnCnVar(SCALE);
    const int sz = cg->size;

    int neigh = dwt_share.neigh;
    bool dwt_sh = (dwt_share.on && sz > 2 * neigh &&
                   (dwt_share.p_share == 1.0f || Random::BoolProb(dwt_share.p_share, thr_no)));

    if(wt_bal.on) {
      // note: MUST get these from ru -- diff for each con -- can't copy to sender!
      // storing in synapses is about 2x faster and essentially no overhead vs. no wtbal
      float* wbincs = cg->OwnCnVar(WB_INC);
      float* wbdecs = cg->OwnCnVar(WB_DEC);

      if(slow_wts.on) {
        for(int i=0; i<sz; i++) {
          float dwt = C_Compute_Weights_dwtshare(dwt_sh, dwts, i, neigh, sz);
          C_Compute_Weights_CtLeabraXCAL_slow
            (wts[i], dwt, fwts[i], swts[i], scales[i], wbincs[i], wbdecs[i], thr_no);
        }
      }
      else {
        for(int i=0; i<sz; i++) {
          float dwt = C_Compute_Weights_dwtshare(dwt_sh, dwts, i, neigh, sz);
          C_Compute_Weights_CtLeabraXCAL
            (wts[i], dwt, fwts[i], swts[i], scales[i], wbincs[i], wbdecs[i], thr_no);
        }
      }
    }
    else {
      if(slow_wts.on) {
        for(int i=0; i<sz; i++) {
          float dwt = C_Compute_Weights_dwtshare(dwt_sh, dwts, i, neigh, sz);
          C_Compute_Weights_CtLeabraXCAL_slow
            (wts[i], dwt, fwts[i], swts[i], scales[i], 1.0f, 1.0f, thr_no);
        }
      }
      else {
        for(int i=0; i<sz; i++) {
          float dwt = C_Compute_Weights_dwtshare(dwt_sh, dwts, i, neigh, sz);
          C_Compute_Weights_CtLeabraXCAL
            (wts[i], dwt, fwts[i], swts[i], scales[i], 1.0f, 1.0f, thr_no);
        }
      }
    }
    // reset dwts after updating -- dwtshare requires doing this after the fact
    for(int i=0; i<sz; i++) {
      dwts[i] = 0.0f;
    }
  }

  INLINE virtual void DwtNorm_SendCons(LEABRA_CON_STATE* cg, LEABRA_NETWORK_STATE* net,
                                       int thr_no) {
    float* dwnorms = cg->OwnCnVar(DWNORM);
    const int sz = cg->size;
    float max_dwnorm = 0.0f;
    for(int i=0; i<sz; i++) {
      max_dwnorm = fmaxf(max_dwnorm, dwnorms[i]); // simple max
    }
    cg->cons_dwnorm = max_dwnorm;
    for(int i=0; i<sz; i++) {
      dwnorms[i] = max_dwnorm;
    }
  }
  // #IGNORE compute dwt_norm sender-based con group level dwnorm factor

  INLINE virtual void   Compute_WtBal_DwtNormRecv(LEABRA_CON_STATE* cg, LEABRA_NETWORK_STATE* net, int thr_no) {
    bool do_wb = wt_bal.on;
    bool do_norm = dwt_norm.RecvConsAgg();
    if(!learn || cg->size < 1 || !(do_wb || do_norm)) return;
    LEABRA_UNIT_STATE* ru = cg->ThrOwnUnState(net, thr_no);
    if(wt_bal.no_targ &&
       (ru->HasUnitFlag(LEABRA_UNIT_STATE::TRC) || ru->HasExtFlag(LEABRA_UNIT_STATE::TARG))) {
      do_wb = false;
      if(!do_norm) return;      // no need
    }

    float sum_wt = 0.0f;
    int sum_n = 0;
    float max_dwnorm = 0.0f;

    const int sz = cg->size;
    for(int i=0; i<sz; i++) {
      if(do_wb) {
        float wt = cg->PtrCn(i,WT,net);
        if(wt >= wt_bal.avg_thr) {
          sum_wt += wt;
          sum_n++;
        }
      }
      if(do_norm) {
        float dwnorm = cg->PtrCn(i,DWNORM,net);
        max_dwnorm = fmaxf(max_dwnorm, dwnorm);
      }
    }

    if(do_norm) {
      cg->cons_dwnorm = max_dwnorm;
    }

    if(do_wb) {
      if(sum_n > 0)
        sum_wt /= (float)sum_n;
      else
        sum_wt = 0.0f;
      cg->wb_avg = sum_wt;
      wt_bal.WtBal(sum_wt, ru->act_avg, cg->wb_fact, cg->wb_inc, cg->wb_dec);
    }
    // note: these are specific to recv unit and cannot be copied to sender!
    // BUT can copy to synapses:

    for(int i=0; i<sz; i++) {
      if(do_wb) {
        cg->PtrCn(i,WB_INC,net) = cg->wb_inc;
        cg->PtrCn(i,WB_DEC,net) = cg->wb_dec;
      }
      if(do_norm) {
        cg->PtrCn(i,DWNORM,net) = max_dwnorm;
      }
    }
  }
  // #IGNORE compute weight balance factors and / or DwtNorm at a recv level

*/

// ActAvgPars has rate constants for averaging over activations at different time scales,
// to produce the running average activation values that then drive learning in the XCAL learning rules
type ActAvgPars struct {
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
func (aa *ActAvgPars) AvgsFmAct(ruAct float32, avgSS, avgS, avgM, AvgSlrn *float32) {
	*avgSS += aa.SSDt * (ruAct - *avgSS)
	*avgS += aa.SDt * (*avgSS - *avgS)
	*avgM += aa.MDt * (*avgS - *avgM)

	AvgSLrn = aa.LrnS**avgS + aa.LrnM**avgM
}

func (aa *ActAvgPars) Update() {
	aa.SSDt = 1 / aa.SSTau
	aa.SDt = 1 / aa.STau
	aa.MDt = 1 / aa.MTau
	aa.LrnS = 1 - aa.LrnM
}

func (aa *ActAvgPars) Defaults() {
	aa.SSTau = 4.0
	aa.STau = 2.0
	aa.MTau = 10.0
	aa.LrnM = 0.1
	aa.Init = 0.15
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

	Dt      float32 `view:"-" inactive:"+" desc:"rate = 1 / tau"`
	LrnFact float32 `view:"-" inactive:"+" desc:"(LrnMax - LrnMin) / (Gain - Min)"`
}

// AvgLFmAct computes long-term average activation value, and learning factor, from given activation
func (al *AvgLPars) AvgLFmAct(act float32, avgl, lrn *float32) {
	avgl += al.Dt * (al.Gain*act - avgl)
	if avgl < al.Min {
		avgl = min
	}
	lrn = al.LrnFact * (avgl - al.Min)
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
//  DWtNormPars

// DWtNormPars are weight change (dwt) normalization parameters, using MAX(ABS(dwt)) aggregated over
// Sending connections in a given projection for a given unit.
// Slowly decays and instantly resets to any current max(abs)
// Serves as an estimate of the variance in the weight changes, assuming zero net mean overall.
type DWtNormPars struct {
	On       bool    `def:"true" desc:"whether to use dwt normalization, only on error-driven dwt component, based on projection-level max_avg value -- slowly decays and instantly resets to any current max"`
	DecayTau float32 `condshow:"On=true" min:"1" def:"1000;10000" desc:"time constant for decay of dwnorm factor -- generally should be long-ish, between 1000-10000 -- integration rate factor is 1/tau"`
	NormMin  float32 `condshow:"On=true" min:"0" def:"0.001" desc:"minimum effective value of the normalization factor -- provides a lower bound to how much normalization can be applied"`
	LrComp   float32 `condshow:"On=true" min:"0" def:"0.15" desc:"overall learning rate multiplier to compensate for changes due to use of normalization -- allows for a common master learning rate to be used between different conditions -- 0.1 for synapse-level, maybe higher for other levels"`
	Stats    bool    `condshow:"On=true" def:"false" desc:"record the avg, max values of err, bcm hebbian, and overall dwt change per con group and per projection"`

	DecayDt  float32 `inactive:"+" view:"-" desc:"rate constant of decay = 1 / decay_tau"`
	DecayDtC float32 `inactive:"+" view:"-" desc:"complement rate constant of decay = 1 - (1 / decay_tau)"`
}

// DWtNormPars updates the dwnorm running max_abs, slowly decaying value
// jumps up to max(abs_dwt) and slowly decays
// returns the effective normalization factor, as a multiplier, including lrate comp
func (dn *DWtNormPars) NormFmAbsDWt(dwnorm, absDwt float32) float32 {
	dwnorm = math32.Max(dn.DecayDtC*dwnorm, absDwt)
	if dwnorm == 0 {
		return 1
	}
	norm := math32.Max(dwnorm, dn.NormMin)
	return dn.LrComp / norm
}

func (dn *DWtNormPars) Update() {
	dn.DecayDt = 1 / dn.DecayTau
	dn.DecayDtC = 1 - dn.DecayDt
}

func (dn *DWtNormPars) Defaults() {
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
func (mp *MomentumPars) MomentFmDWt(moment, dwt float32) float32 {
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
