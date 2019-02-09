// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package deep

import (
	"fmt"

	"github.com/chewxy/math32"
	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/leabra/leabra"
	"github.com/goki/ki/kit"
)

// deep.Layer is the DeepLeabra layer, based on basic rate-coded leabra.Layer
type Layer struct {
	leabra.Layer                 // access as .Layer
	DeepBurst    DeepBurstParams `desc:"parameters for computing DeepBurst from act, in Superficial layers (but also needed in Deep layers for deep self connections)"`
	DeepCtxt     DeepCtxtParams  `desc:"parameters for computing DeepCtxt in Deep layers, from BurstCtxt inputs from Super senders"`
	DeepTRC      DeepTRCParams   `desc:"parameters for computing TRC plus-phase (outcome) activations based on TRCBurstGe excitatory input from BurstTRC projections"`
	DeepAttn     DeepAttnParams  `desc:"parameters for computing DeepAttn and DeepLrn attentional modulation signals based on DeepAttn projection inputs integrated into AttnGe excitatory conductances"`
	DeepNeurs    []Neuron        `desc:"slice of extra deep.Neuron state for this layer -- flat list of len = Shape.Len(). You must iterate over index and use pointer to modify values."`
	DeepPools    []Pool          `desc:"extra layer and sub-pool (unit group) statistics used in DeepLeabra -- flat list has at least of 1 for layer, and one for each sub-pool (unit group) if shape supports that (4D).  You must iterate over index and use pointer to modify values."`
}

// AsLeabra returns this layer as a leabra.Layer -- all derived layers must redefine
// this to return the base Layer type, so that the LeabraLayer interface does not
// need to include accessors to all the basic stuff
func (ly *Layer) AsLeabra() *leabra.Layer {
	return &ly.Layer
}

// AsDeep returns this layer as a deep.Layer -- all derived layers must redefine
// this to return the deep Layer type, so that the DeepLayer interface does not
// need to include accessors to all the fields.
func (ly *Layer) AsDeep() *Layer {
	return ly
}

func (ly *Layer) Defaults() {
	ly.Layer.Defaults()
	ly.DeepBurst.Defaults()
	ly.DeepCtxt.Defaults()
	ly.DeepTRC.Defaults()
	ly.DeepAttn.Defaults()
}

// UpdateParams updates all params given any changes that might have been made to individual values
// including those in the receiving projections of this layer
func (ly *Layer) UpdateParams() {
	ly.Layer.UpdateParams()
	ly.DeepBurst.Update()
	ly.DeepCtxt.Update()
	ly.DeepTRC.Update()
	ly.DeepAttn.Update()
}

// SetParams sets given parameters to this layer, if the target type is Layer
// calls UpdateParams to ensure derived parameters are all updated.
// If setMsg is true, then a message is printed to confirm each parameter that is set.
// it always prints a message if a parameter fails to be set.
func (ly *Layer) SetParams(pars emer.Params, setMsg bool) bool {
	trg := pars.Target()
	if trg != "Layer" {
		return false
	}
	pars.Set(ly, setMsg)
	ly.UpdateParams()
	return true
}

// UnitVarNames returns a list of variable names available on the units in this layer
func (ly *Layer) UnitVarNames() []string {
	return AllNeuronVars
}

// UnitVals is emer.Layer interface method to return values of given variable
func (ly *Layer) UnitVals(varNm string) ([]float32, error) {
	vidx, err := leabra.NeuronVarByName(varNm)
	if err == nil {
		return ly.Layer.UnitVals(varNm)
	}
	vidx, err = NeuronVarByName(varNm)
	if err != nil {
		return nil, err
	}
	vs := make([]float32, len(ly.DeepNeurs))
	for i := range ly.DeepNeurs {
		dnr := &ly.DeepNeurs[i]
		vs[i] = dnr.VarByIndex(vidx)
	}
	return vs, nil
}

// UnitVal returns value of given variable name on given unit,
// using shape-based dimensional index
func (ly *Layer) UnitVal(varNm string, idx []int) (float32, error) {
	_, err := leabra.NeuronVarByName(varNm)
	if err == nil {
		return ly.Layer.UnitVal(varNm, idx)
	}
	fidx := ly.Shape.Offset(idx)
	nn := len(ly.DeepNeurs)
	if fidx < 0 || fidx >= nn {
		return 0, fmt.Errorf("Layer UnitVal index: %v out of range, N = %v", fidx, nn)
	}
	dnr := &ly.DeepNeurs[fidx]
	return dnr.VarByName(varNm)
}

// UnitVal1D returns value of given variable name on given unit,
// using 1-dimensional index.
func (ly *Layer) UnitVal1D(varNm string, idx int) (float32, error) {
	_, err := leabra.NeuronVarByName(varNm)
	if err == nil {
		return ly.Layer.UnitVal1D(varNm, idx)
	}
	nn := len(ly.DeepNeurs)
	if idx < 0 || idx >= nn {
		return 0, fmt.Errorf("Layer UnitVal1D index: %v out of range, N = %v", idx, nn)
	}
	dnr := &ly.DeepNeurs[idx]
	return dnr.VarByName(varNm)
}

// Build constructs the layer state, including calling Build on the projections
// you MUST have properly configured the Inhib.Pool.On setting by this point
// to properly allocate Pools for the unit groups if necessary.
func (ly *Layer) Build() error {
	err := ly.Layer.Build()
	if err != nil {
		return err
	}
	ly.DeepNeurs = make([]Neuron, len(ly.Neurons))
	ly.DeepPools = make([]Pool, len(ly.Pools))
	return nil
}

//////////////////////////////////////////////////////////////////////////////////////
//  Init methods

func (ly *Layer) InitActs() {
	ly.Layer.InitActs()
	for ni := range ly.DeepNeurs {
		dnr := &ly.DeepNeurs[ni]
		dnr.ActNoAttn = 0
		dnr.DeepBurst = 0
		dnr.DeepBurstPrv = 0
		dnr.DeepCtxt = 0
		dnr.TRCBurstGe = 0
		dnr.DeepBurstSent = 0
		dnr.AttnGe = 0
		dnr.DeepAttn = 0
		dnr.DeepLrn = 0
	}
}

// GScaleFmAvgAct computes the scaling factor for synaptic input conductances G,
// based on sending layer average activation.
// This attempts to automatically adjust for overall differences in raw activity coming into the units
// to achieve a general target of around .5 to 1 for the integrated G values.
// DeepLeabra version separately normalizes the Deep projection types.
func (ly *Layer) GScaleFmAvgAct() {
	totGeRel := float32(0)
	totGiRel := float32(0)
	totTrcRel := float32(0)
	totAttnRel := float32(0)
	for _, p := range ly.RecvPrjns {
		if p.IsOff() {
			continue
		}
		pj := p.(leabra.LeabraPrjn).AsLeabra()
		slay := p.SendLay().(leabra.LeabraLayer).AsLeabra()
		slpl := slay.Pools[0]
		savg := slpl.ActAvg.ActPAvgEff
		snu := len(slay.Neurons)
		ncon := pj.RConNAvgMax.Avg
		pj.GScale = pj.WtScale.FullScale(savg, float32(snu), ncon)
		switch pj.Type {
		case emer.Inhib:
			totGiRel += pj.WtScale.Rel
		case BurstTRC:
			totTrcRel += pj.WtScale.Rel
		case DeepAttn:
			totAttnRel += pj.WtScale.Rel
		default:
			// note: BurstCtxt is added in here!
			totGeRel += pj.WtScale.Rel
		}
	}

	for _, p := range ly.RecvPrjns {
		if p.IsOff() {
			continue
		}
		pj := p.(leabra.LeabraPrjn).AsLeabra()
		switch pj.Type {
		case emer.Inhib:
			if totGiRel > 0 {
				pj.GScale /= totGiRel
			}
		case BurstTRC:
			if totTrcRel > 0 {
				pj.GScale /= totTrcRel
			}
		case DeepAttn:
			if totAttnRel > 0 {
				pj.GScale /= totAttnRel
			}
		default:
			if totGeRel > 0 {
				pj.GScale /= totGeRel
			}
		}
	}
}

func (ly *Layer) DecayState(decay float32) {
	ly.Layer.DecayState(decay)
	for ni := range ly.DeepNeurs {
		dnr := &ly.DeepNeurs[ni]
		dnr.ActNoAttn -= decay * (dnr.ActNoAttn - ly.Act.Init.Act)
		dnr.DeepBurstSent = 0
	}
}

//////////////////////////////////////////////////////////////////////////////////////
//  Cycle

// SendGDelta sends change in activation since last sent, if above thresholds.
// Deep version sends either to standard Ge or AttnGe for DeepAttn projections.
func (ly *Layer) SendGDelta(ltime *leabra.Time) {
	for ni := range ly.Neurons {
		nrn := &ly.Neurons[ni]
		if nrn.Act > ly.Act.OptThresh.Send {
			delta := nrn.Act - nrn.ActSent
			if math32.Abs(delta) > ly.Act.OptThresh.Delta {
				for si := range ly.SendPrjns {
					sp := ly.SendPrjns[si]
					if sp.IsOff() {
						continue
					}
					pj := sp.(DeepPrjn)
					ptyp := pj.PrjType()
					if ptyp == BurstCtxt || ptyp == BurstTRC {
						continue
					}
					if ptyp == DeepAttn {
						if ly.DeepAttn.On {
							pj.SendAttnGeDelta(ni, delta)
						}
					} else {
						pj.SendGDelta(ni, delta)
					}
				}
				nrn.ActSent = nrn.Act
			}
		} else if nrn.ActSent > ly.Act.OptThresh.Send {
			delta := -nrn.ActSent // un-send the last above-threshold activation to get back to 0
			for si := range ly.SendPrjns {
				sp := ly.SendPrjns[si]
				if sp.IsOff() {
					continue
				}
				pj := sp.(DeepPrjn)
				ptyp := pj.PrjType()
				if ptyp == BurstCtxt || ptyp == BurstTRC {
					continue
				}
				if ptyp == DeepAttn {
					if ly.DeepAttn.On {
						pj.SendAttnGeDelta(ni, delta)
					}
				} else {
					pj.SendGDelta(ni, delta)
				}
			}
			nrn.ActSent = 0
		}
	}
}

// GFmInc integrates new synaptic conductances from increments sent during last SendGDelta.
func (ly *Layer) GFmInc(ltime *leabra.Time) {
	if ly.Type == TRC && ly.DeepBurst.IsBurstQtr(ltime.Quarter) {
		// note: TRCBurstGe is sent at *end* of previous cycle, after DeepBurst act is computed
		lpl := &ly.DeepPools[0]
		if lpl.TRCBurstGe.Max > 0.1 { // have some actual input
			for ni := range ly.Neurons {
				nrn := &ly.Neurons[ni]
				dnr := &ly.DeepNeurs[ni]
				ly.Act.GRawFmInc(nrn) // key to integrate and reset Inc's
				geRaw := ly.DeepTRC.BurstGe(dnr.TRCBurstGe)
				ly.Act.Dt.GFmRaw(geRaw, &nrn.Ge) // Ge driven exclusively from Burst
			}
			return
		}
	}
	ly.Layer.GFmInc(ltime) // regular
	if ly.DeepAttn.On {
		for _, p := range ly.RecvPrjns {
			if p.IsOff() {
				continue
			}
			pj := p.(DeepPrjn)
			ptyp := pj.PrjType()
			if ptyp != DeepAttn {
				continue
			}
			pj.RecvAttnGeInc()
		}
	}
}

// AvgMaxGe computes the average and max Ge stats, used in inhibition
// Deep version also computes AttnGe stats
func (ly *Layer) AvgMaxGe(ltime *leabra.Time) {
	ly.Layer.AvgMaxGe(ltime)
	ly.LeabraLay.(DeepLayer).AvgMaxAttnGe(ltime)
}

// AvgMaxAttnGe computes the average and max AttnGe stats
func (ly *Layer) AvgMaxAttnGe(ltime *leabra.Time) {
	for pi := range ly.DeepPools {
		pl := &ly.Pools[pi]
		dpl := &ly.DeepPools[pi]
		dpl.AttnGe.Init()
		for ni := pl.StIdx; ni < pl.EdIdx; ni++ {
			dnr := &ly.DeepNeurs[ni]
			dpl.AttnGe.UpdateVal(dnr.AttnGe, ni)
		}
		dpl.AttnGe.CalcAvg()
	}
}

// ActFmG computes rate-code activation from Ge, Gi, Gl conductances
// and updates learning running-average activations from that Act
func (ly *Layer) ActFmG(ltime *leabra.Time) {
	ly.Layer.ActFmG(ltime)
	ly.LeabraLay.(DeepLayer).DeepAttnFmG(ltime)
}

// DeepAttnFmG computes DeepAttn and DeepLrn from AttnGe input,
// and then applies the DeepAttn modulation to the Act activation value.
func (ly *Layer) DeepAttnFmG(ltime *leabra.Time) {
	lpl := &ly.DeepPools[0]
	attnMax := lpl.AttnGe.Max
	for ni := range ly.DeepNeurs {
		dnr := &ly.DeepNeurs[ni]
		nrn := &ly.Neurons[ni]
		switch {
		case !ly.DeepAttn.On:
			dnr.DeepAttn = 1
			dnr.DeepLrn = 1
		case ly.Type == Deep:
			dnr.DeepAttn = nrn.Act // record Deep activation = DeepAttn signal coming from deep layers
			dnr.DeepLrn = 1
		case ly.Type == TRC:
			dnr.DeepAttn = 1
			dnr.DeepLrn = 1
		default:
			if attnMax < ly.DeepAttn.Thr {
				dnr.DeepAttn = 1
				dnr.DeepLrn = 1
			} else {
				dnr.DeepLrn = dnr.AttnGe / attnMax
				dnr.DeepAttn = ly.DeepAttn.DeepAttnFmG(dnr.DeepLrn)
			}
		}
		dnr.ActNoAttn = nrn.Act
		nrn.Act *= dnr.DeepAttn
	}
}

// AvgMaxAct computes the average and max Act stats, used in inhibition
// Deep version also computes AvgMaxActNoAttn
func (ly *Layer) AvgMaxAct(ltime *leabra.Time) {
	ly.Layer.AvgMaxAct(ltime)
	ly.LeabraLay.(DeepLayer).AvgMaxActNoAttn(ltime)
}

// AvgMaxActNoAttn computes the average and max ActNoAttn stats
func (ly *Layer) AvgMaxActNoAttn(ltime *leabra.Time) {
	for pi := range ly.DeepPools {
		pl := &ly.Pools[pi]
		dpl := &ly.DeepPools[pi]
		dpl.ActNoAttn.Init()
		for ni := pl.StIdx; ni < pl.EdIdx; ni++ {
			dnr := &ly.DeepNeurs[ni]
			dpl.ActNoAttn.UpdateVal(dnr.ActNoAttn, ni)
		}
		dpl.ActNoAttn.CalcAvg()
	}
}

//////////////////////////////////////////////////////////////////////////////////////
//  DeepBurst -- computed every cycle at end of standard Cycle in DeepBurst quarter

// DeepBurstFmAct updates DeepBurst layer 5 IB bursting value from current Act (superficial activation)
// Subject to thresholding.
func (ly *Layer) DeepBurstFmAct(ltime *leabra.Time) {
	if !ly.DeepBurst.On || !ly.DeepBurst.IsBurstQtr(ltime.Quarter) {
		return
	}
	lpl := &ly.DeepPools[0]
	actMax := lpl.ActNoAttn.Max
	actAvg := lpl.ActNoAttn.Avg
	thr := actAvg + ly.DeepBurst.ThrRel*(actMax-actAvg)
	thr = math32.Max(thr, ly.DeepBurst.ThrAbs)
	for ni := range ly.DeepNeurs {
		dnr := &ly.DeepNeurs[ni]
		burst := float32(0)
		if dnr.ActNoAttn > thr {
			burst = dnr.ActNoAttn
		}
		dnr.DeepBurst = burst
	}
}

// SendTRCBurstGeDelta sends change in DeepBurst activation since last sent, over BurstTRC
// projections.
func (ly *Layer) SendTRCBurstGeDelta(ltime *leabra.Time) {
	if !ly.DeepBurst.On || !ly.DeepBurst.IsBurstQtr(ltime.Quarter) {
		return
	}
	for ni := range ly.DeepNeurs {
		dnr := &ly.DeepNeurs[ni]
		if dnr.DeepBurst > ly.Act.OptThresh.Send {
			delta := dnr.DeepBurst - dnr.DeepBurstSent
			if math32.Abs(delta) > ly.Act.OptThresh.Delta {
				for si := range ly.SendPrjns {
					sp := ly.SendPrjns[si]
					if sp.IsOff() {
						continue
					}
					pj := sp.(DeepPrjn)
					ptyp := pj.PrjType()
					if ptyp != BurstTRC {
						continue
					}
					pj.SendTRCBurstGeDelta(ni, delta)
				}
				dnr.DeepBurstSent = dnr.DeepBurst
			}
		} else if dnr.DeepBurstSent > ly.Act.OptThresh.Send {
			delta := -dnr.DeepBurstSent // un-send the last above-threshold activation to get back to 0
			for si := range ly.SendPrjns {
				sp := ly.SendPrjns[si]
				if sp.IsOff() {
					continue
				}
				pj := sp.(DeepPrjn)
				ptyp := pj.PrjType()
				if ptyp != BurstTRC {
					continue
				}
				pj.SendTRCBurstGeDelta(ni, delta)
			}
			dnr.DeepBurstSent = 0
		}
	}
}

// TRCBurstGeFmInc computes the TRCBurstGe input from sent values
func (ly *Layer) TRCBurstGeFmInc(ltime *leabra.Time) {
	if !ly.DeepBurst.On || !ly.DeepBurst.IsBurstQtr(ltime.Quarter) {
		return
	}
	for _, p := range ly.RecvPrjns {
		if p.IsOff() {
			continue
		}
		pj := p.(DeepPrjn)
		ptyp := pj.PrjType()
		if ptyp != BurstTRC {
			continue
		}
		pj.RecvTRCBurstGeInc()
	}
	// note: full integration of Inc happens next cycle..
}

// AvgMaxTRCBurstGe computes the average and max TRCBurstGe stats
func (ly *Layer) AvgMaxTRCBurstGe(ltime *leabra.Time) {
	for pi := range ly.DeepPools {
		pl := &ly.Pools[pi]
		dpl := &ly.DeepPools[pi]
		dpl.TRCBurstGe.Init()
		for ni := pl.StIdx; ni < pl.EdIdx; ni++ {
			dnr := &ly.DeepNeurs[ni]
			dpl.TRCBurstGe.UpdateVal(dnr.TRCBurstGe, ni)
		}
		dpl.TRCBurstGe.CalcAvg()
	}
}

//////////////////////////////////////////////////////////////////////////////////////
//  DeepCtxt -- once after DeepBurst quarter

// SendDeepCtxtGe sends full DeepBurst activation over BurstCtxt projections to integrate
// DeepCtxtGe excitatory conductance on deep layers.
// This must be called at the end of the DeepBurst quarter for this layer.
func (ly *Layer) SendDeepCtxtGe(ltime *leabra.Time) {
	if !ly.DeepBurst.On || !ly.DeepBurst.IsBurstQtr(ltime.Quarter) {
		return
	}
	for ni := range ly.DeepNeurs {
		dnr := &ly.DeepNeurs[ni]
		if dnr.DeepBurst > ly.Act.OptThresh.Send {
			for si := range ly.SendPrjns {
				sp := ly.SendPrjns[si]
				if sp.IsOff() {
					continue
				}
				pj := sp.(DeepPrjn)
				ptyp := pj.PrjType()
				if ptyp != BurstCtxt {
					continue
				}
				pj.SendDeepCtxtGe(ni, dnr.DeepBurst)
			}
		}
	}
}

// DeepCtxtFmGe integrates new DeepCtxtGe excitatory conductance from projections, and computes
// overall DeepCtxt value, only on Deep layers.
// This must be called at the end of the DeepBurst quarter for this layer, after SendDeepCtxtGe.
func (ly *Layer) DeepCtxtFmGe(ltime *leabra.Time) {
	if ly.Type != Deep || !ly.DeepBurst.IsBurstQtr(ltime.Quarter) {
		return
	}
	for ni := range ly.DeepNeurs {
		dnr := &ly.DeepNeurs[ni]
		dnr.DeepCtxtGe = 0
	}
	for _, p := range ly.RecvPrjns {
		if p.IsOff() {
			continue
		}
		pj := p.(DeepPrjn)
		ptyp := pj.PrjType()
		if ptyp != BurstCtxt {
			continue
		}
		pj.RecvDeepCtxtGeInc()
	}
	for ni := range ly.DeepNeurs {
		dnr := &ly.DeepNeurs[ni]
		dnr.DeepCtxt = ly.DeepCtxt.DeepCtxtFmGe(dnr.DeepCtxtGe, dnr.DeepCtxt)
	}
}

// QuarterFinal does updating after end of a quarter
func (ly *Layer) QuarterFinal(ltime *leabra.Time) {
	ly.Layer.QuarterFinal(ltime)
	ly.LeabraLay.(DeepLayer).DeepBurstPrv(ltime)
}

// DeepBurstPrv saves DeepBurst as DeepBurstPrv
func (ly *Layer) DeepBurstPrv(ltime *leabra.Time) {
	if !ly.DeepBurst.On || !ly.DeepBurst.NextIsBurstQtr(ltime.Quarter) {
		return
	}
	for ni := range ly.DeepNeurs {
		dnr := &ly.DeepNeurs[ni]
		dnr.DeepBurstPrv = dnr.DeepBurst
	}
}

//////////////////////////////////////////////////////////////////////////////////////
//  LayerType

// DeepLeabra extensions to the emer.LayerType types

//go:generate stringer -type=LayerType

var KiT_LayerType = kit.Enums.AddEnum(LayerTypeN, false, nil)

// The DeepLeabra layer types
const (
	// Super are superficial-layer neurons, which also compute DeepBurst activation as a
	// thresholded version of superficial activation, and send that to both TRC (for plus
	// phase outcome) and Deep layers (for DeepCtxt temporal context).
	Super emer.LayerType = emer.LayerTypeN + iota

	// Deep are deep-layer neurons, reflecting activation of layer 6 regular spiking
	// CT corticothalamic neurons, which drive both attention in Super (via DeepAttn
	// projections) and  predictions in TRC (Pulvinar) via standard projections.
	Deep

	// TRC are thalamic relay cell neurons, typically in the Pulvinar, which alternately reflect
	// predictions driven by Deep layer projections, and actual outcomes driven by BurstTRC
	// projections from corresponding Super layer neurons that provide strong driving inputs to
	// TRC neurons.
	TRC

	LayerTypeN
)
