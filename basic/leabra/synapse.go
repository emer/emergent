// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

// leabra.Synapse holds state for the synaptic connection between neurons
type Synapse struct {
	Wt      float32 `desc:"synaptic weight value -- sigmoid contrast-enhanced"`
	LWt     float32 `desc:"linear (underlying) weight value -- learns according to the lrate specified in the connection spec -- this is converted into the effective weight value, Wt, via sigmoidal contrast enhancement (see WtSigPars)"`
	DWt     float32 `desc:"change in synaptic weight, from learning"`
	DWtNorm float32 `desc:"dwt normalization factor -- reset to max of abs value of dwt, decays slowly down over time -- serves as an estimate of variance in weight changes over time"`
	Moment  float32 `desc:"momentum -- time-integrated dwt changes, to accumulate a consistent direction of weight change and cancel out dithering contradictory changes"`
	WbInc   float32 `desc:"rate of weight increase from adaptive weight balance -- computed receiver based and so needs to be stored in the connection to optimize speed"`
	WbDec   float32 `desc:"rate of weight decrease from adaptive weight balance -- computed receiver based and so needs to be stored in the connection to optimize speed"`
}

// leabra.SynPrjn is a projection of synapses
type SynPrjn struct {
	Net         float32 // #NO_SAVE #CAT_Activation netinput to this con_group: only computed for special statistics such as RelNetin
	NetRaw      float32 // #NO_SAVE #CAT_Activation raw summed netinput to this con_group -- only used for NETIN_PER_PRJN
	WbAvg       float32 // #NO_SAVE #CAT_Learning average of effective weight values that exceed wt_bal.avg_thr across this con state -- used for weight balance
	WbFact      float32 // #NO_SAVE #CAT_Learning overall weight balance factor that drives changes in wb_inc vs. wb_dec via as sigmoidal function -- this is the net strength of weigth balance changes
	WbInc       float32 // #NO_SAVE #CAT_Learning weight balance increment factor -- extra multiplier to add to weight increases to maintain overall weight balance
	WbDec       float32 // #NO_SAVE #CAT_Learning weight balance decrement factor -- extra multiplier to add to weight decreases to maintain overall weight balance
	DWtNormCons float32 // #NO_SAVE #GUI_READ_ONLY #SHOW #CAT_Learning connection group level dwt_norm normalization factor -- slowly decaying max(abs(dwt)) across single unit's projection -- only updated if conspec dwt_norm.level is CON*
}
