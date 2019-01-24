// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import (
	"github.com/chewxy/math32"
	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/erand"
	"github.com/emer/emergent/etensor"
)

// leabra.LayerStru manages the structural elements of the layer, which are common
// to any Layer type
type LayerStru struct {
	Name      string        `desc:"Name of the layer -- this must be unique within the network, which has a map for quick lookup and layers are typically accessed directly by name"`
	Class     string        `desc:"Class is for applying parameter styles, can be space separated multple tags"`
	Off       bool          `desc:"inactivate this layer -- allows for easy experimentation"`
	Shape     etensor.Shape `desc:"shape of the layer -- can be 2D for basic layers and 4D for layers with sub-groups (hypercolumns) -- order is outer-to-inner (row major) so Y then X for 2D and for 4D: Y-X unit pools then Y-X units within pools"`
	Rel       emer.Rel      `desc:"Spatial relationship to other layer, determines positioning"`
	Pos       emer.Vec3i    `desc:"position of lower-left-hand corner of layer in 3D space, computed from Rel"`
	RecvPrjns PrjnList      `desc:"list of receiving projections into this layer from other layers"`
	SendPrjns PrjnList      `desc:"list of sending projections from this layer to other layers"`
}

// emer.Layer interface methods

func (ls *LayerStru) LayName() string            { return ls.Name }
func (ls *LayerStru) LayClass() string           { return ls.Class }
func (ls *LayerStru) IsOff() bool                { return ls.Off }
func (ls *LayerStru) LayShape() *etensor.Shape   { return &ls.Shape }
func (ls *LayerStru) LayPos() emer.Vec3i         { return ls.Pos }
func (ls *LayerStru) NRecvPrjns() int            { return len(ls.RecvPrjns) }
func (ls *LayerStru) RecvPrjn(idx int) emer.Prjn { return ls.RecvPrjns[idx] }
func (ls *LayerStru) NSendPrjns() int            { return len(ls.SendPrjns) }
func (ls *LayerStru) SendPrjn(idx int) emer.Prjn { return ls.SendPrjns[idx] }

// SetShape sets the layer shape and also uses default dim names
func (ls *LayerStru) SetShape(shape []int) {
	var dnms []string
	if len(shape) == 2 {
		dnms = []string{"X", "Y"}
	} else if len(shape) == 4 {
		dnms = []string{"GX", "GY", "X", "Y"} // group X,Y
	}
	ls.Shape.SetShape(shape, nil, dnms) // row major default
}

func (ls *LayerStru) RecvPrjnBySendName(sender string) (emer.Prjn, bool) {
	for _, pj := range ls.RecvPrjns {
		if pj.Send.LayName() == sender {
			return pj, true
		}
	}
	return nil, false
}

func (ls *LayerStru) SendPrjnByRecvName(recv string) (emer.Prjn, bool) {
	for _, pj := range ls.SendPrjns {
		if pj.Recv.LayName() == recv {
			return pj, true
		}
	}
	return nil, false
}

// NPools returns the number of unit sub-pools according to the shape parameters.
// Currently supported for a 4D shape, where the unit pools are the first 2 Y,X dims
// and then the units within the pools are the 2nd 2 Y,X dims
func (ls *LayerStru) NPools() int {
	if ls.Shape.NumDims() != 4 {
		return 0
	}
	sh := ls.Shape.Shape()
	return int(sh[0] * sh[1])
}

//////////////////////////////////////////////////////////////////////////////////////
//  Layer

// todo: need AvgMax Ge and Act for inhib
// todo: need overall good strategy for stats
// todo: need to pass Time around..

// leabra.Layer has parameters for running a basic rate-coded Leabra layer
type Layer struct {
	LayerStru
	Act     ActParams       `desc:"Activation parameters and methods for computing activations"`
	Inhib   InhibParams     `desc:"Inhibition parameters and methods for computing layer-level inhibition"`
	Learn   LearnNeurParams `desc:"Learning parameters and methods that operate at the neuron level"`
	Neurons []Neuron        `desc:"slice of neurons for this layer -- flat list of len = Shape.Len(). You must iterate over index and use pointer to modify values."`
	Pools   []Pool          `desc:"inhibition and other pooled, aggregate state variables -- flat list has at least of 1 for layer, and one for each unit group if shape supports that (4D).  You must iterate over index and use pointer to modify values."`
	CosDiff CosDiffStats    `desc:"cosine difference between ActM, ActP stats"`
}

func (ly *Layer) Defaults() {
	ly.Act.Defaults()
	ly.Inhib.Defaults()
	ly.Learn.Defaults()
}

// UpdateParams updates all params given any changes that might have been made to individual values
func (ly *Layer) UpdateParams() {
	ly.Act.Update()
	ly.Inhib.Update()
	ly.Learn.Update()
}

// Unit is emer.Layer interface method -- only possible with Neurons in place
func (ly *Layer) Unit(idx []int) (emer.Unit, bool) {
	fidx := ly.Shape.Offset(idx)
	if int(fidx) < len(ly.Neurons) {
		return &ly.Neurons[fidx], true
	}
	return nil, false
}

// Build constructs the layer state, including calling Build on the projections
// you MUST have properly configured the Inhib.Pool.On setting by this point
// to properly allocate Pools for the unit groups if necessary.
func (ly *Layer) Build() {
	nu := ly.Shape.Len()
	ly.Neurons = make([]Neuron, nu)
	np := 1
	if ly.Inhib.Pool.On {
		np += ly.NPools()
	}
	ly.Pools = make([]Pool, np)
	lpl := &ly.Pools[0]
	lpl.StIdx = 0
	lpl.EdIdx = nu
	if np > 1 {
		ly.BuildPools()
	}
	ly.RecvPrjns.Build()
}

// BuildPools initializes neuron start / end indexes for sub-group pools
func (ly *Layer) BuildPools() {
	if ly.Shape.NumDims() != 4 {
		return
	}
	sh := ly.Shape.Shape()
	spy := sh[0]
	spx := sh[1]
	lastOff := 0
	pi := 0
	for py := 0; py < spy; py++ {
		for px := 0; px < spx; px++ {
			idx := []int{py, px, 0, 0}
			off := ly.Shape.Offset(idx)
			if off == 0 {
				continue
			}
			pl := &ly.Pools[pi]
			pl.StIdx = lastOff
			pl.EdIdx = off
			pi++
			lastOff = off
		}
	}
}

// note: all basic computation can be performed on layer-level
// and prjn level

//////////////////////////////////////////////////////////////////////////////////////
//  Init methods

// InitWts initializes the weight values in the network, i.e., resetting learning
// Also calls InitActs
func (ly *Layer) InitWts() {
	for _, pj := range ly.SendPrjns {
		pj.InitWts()
	}
	for pi := range ly.Pools {
		pl := &ly.Pools[pi]
		pl.ActAvg.ActMAvg = ly.Inhib.ActAvg.Init
		pl.ActAvg.ActPAvg = ly.Inhib.ActAvg.Init
		pl.ActAvg.ActPAvgEff = ly.Inhib.ActAvg.EffInit()
	}
	ly.InitActAvg()
	ly.InitActs()
}

// InitActAvg initializes the running-average activation values that drive learning.
func (ly *Layer) InitActAvg() {
	for ni := range ly.Neurons {
		nrn := &ly.Neurons[ni]
		ly.Learn.InitActAvg(nrn)
	}
}

// InitActs fully initializes activation state -- only called automatically during InitWts
func (ly *Layer) InitActs() {
	for ni := range ly.Neurons {
		nrn := &ly.Neurons[ni]
		ly.Act.InitActs(nrn)
	}
}

// TrialInit handles all initialization at start of new input pattern, including computing
// netinput scaling from running average activation etc.
func (ly *Layer) TrialInit() {
	ly.AvgLFmAct()
	for pi := range ly.Pools {
		pl := &ly.Pools[pi]
		ly.Inhib.ActAvg.AvgFmAct(&pl.ActAvg.ActMAvg, pl.ActM.Avg)
		ly.Inhib.ActAvg.AvgFmAct(&pl.ActAvg.ActPAvg, pl.ActP.Avg)
		ly.Inhib.ActAvg.EffFmAvg(&pl.ActAvg.ActPAvgEff, pl.ActAvg.ActPAvg)
	}
	ly.GeScaleFmAvgAct()
	if ly.Act.Noise.Type != NoNoise && ly.Act.Noise.TrialFixed && ly.Act.Noise.Dist != erand.None {
		ly.GenNoise()
	}
	ly.DecayState(ly.Act.Init.Decay)
}

// AvgLFmAct updates AvgL long-term running average activation that drives BCM Hebbian learning
func (ly *Layer) AvgLFmAct() {
	for ni := range ly.Neurons {
		nrn := &ly.Neurons[ni]
		ly.Learn.AvgLFmAct(nrn)
		if ly.Learn.AvgL.ErrMod {
			nrn.AvgLLrn *= ly.CosDiff.ModAvgLLrn
		}
	}
}

// GeScaleFmAvgAct computes the scaling factor for Ge excitatory conductance input
// based on sending layer average activation.
// This attempts to automatically adjust for overall differences in raw activity coming into the units
// to achieve a general target of around .5 to 1 for the integrated Ge value.
func (ly *Layer) GeScaleFmAvgAct() {
	totRel := float32(0)
	for _, pj := range ly.RecvPrjns {
		if pj.IsOff() {
			continue
		}
		slay := pj.Send.(*Layer)
		slpl := slay.Pools[0]
		savg := slpl.ActAvg.ActPAvgEff // todo: avg_correct
		snu := len(slay.Neurons)
		ncon := pj.RConNAvgMax.Avg
		pj.GeScale = pj.WtScale.FullScale(savg, float32(snu), ncon)
		totRel += pj.WtScale.Rel
	}

	for _, pj := range ly.RecvPrjns {
		if pj.IsOff() {
			continue
		}
		if totRel > 0 {
			pj.GeScale /= totRel
		}
	}
}

func (ly *Layer) GenNoise() {
	for ni := range ly.Neurons {
		nrn := &ly.Neurons[ni]
		nrn.Noise = float32(ly.Act.Noise.Gen(-1))
	}
}

func (ly *Layer) DecayState(decay float32) {
	for ni := range ly.Neurons {
		nrn := &ly.Neurons[ni]
		ly.Act.DecayState(nrn, decay)
	}
}

//////////////////////////////////////////////////////////////////////////////////////
//  Cycle

// InitGeInc initializes GeInc Ge increment -- optional
func (ly *Layer) InitGeInc() {
	for ni := range ly.Neurons {
		nrn := &ly.Neurons[ni]
		nrn.GeInc = 0
	}
}

// SendGeDelta sends change in activation since last sent, if above thresholds
func (ly *Layer) SendGeDelta() {
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
					sp.SendGeDelta(ni, delta)
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
				sp.SendGeDelta(ni, delta)
			}
			nrn.ActSent = 0
		}
	}
}

// GeFmGeInc integrates new excitatory conductance from GeInc increments sent during last SendGeDelta
func (ly *Layer) GeFmGeInc() {
	for ni := range ly.Neurons {
		nrn := &ly.Neurons[ni]
		ly.Act.GeFmGeInc(nrn)
	}
}

// AvgMaxGe computes the average and max Ge stats, used in inhibition
func (ly *Layer) AvgMaxGe() {
	for pi := range ly.Pools {
		pl := &ly.Pools[pi]
		pl.Ge.Init()
		for ni := pl.StIdx; ni < pl.EdIdx; ni++ {
			nrn := &ly.Neurons[ni]
			pl.Ge.UpdateVal(nrn.Ge, ni)
		}
		pl.Ge.CalcAvg()
	}
}

// InhibiFmGeAct computes inhibition Gi from Ge and Act averages within relevant Pools
func (ly *Layer) InhibFmGeAct() {
	lpl := &ly.Pools[0]
	ly.Inhib.Layer.Inhib(lpl.Ge.Avg, lpl.Ge.Max, lpl.Act.Avg, &lpl.Inhib)
	np := len(ly.Pools)
	if np > 1 {
		for pi := 1; pi < np; pi++ {
			pl := &ly.Pools[pi]
			ly.Inhib.Pool.Inhib(pl.Ge.Avg, pl.Ge.Max, pl.Act.Avg, &pl.Inhib)
			pl.Inhib.Gi = math32.Max(pl.Inhib.Gi, lpl.Inhib.Gi)
			for ni := pl.StIdx; ni < pl.EdIdx; ni++ {
				nrn := &ly.Neurons[ni]
				ly.Inhib.Self.Inhib(&nrn.GiSelf, nrn.Act)
				nrn.Gi = pl.Inhib.Gi + nrn.GiSelf
			}
		}
	} else {
		for ni := lpl.StIdx; ni < lpl.EdIdx; ni++ {
			nrn := &ly.Neurons[ni]
			ly.Inhib.Self.Inhib(&nrn.GiSelf, nrn.Act)
			nrn.Gi = lpl.Inhib.Gi + nrn.GiSelf
		}
	}
}

// ActFmG computes rate-code activation from Ge, Gi, Gl conductances
// and updates learning running-average activations from that Act
func (ly *Layer) ActFmG() {
	for ni := range ly.Neurons {
		nrn := &ly.Neurons[ni]
		ly.Act.VmFmG(nrn)
		ly.Act.ActFmG(nrn)
		ly.Learn.AvgsFmAct(nrn)
	}
}

// AvgMaxAct computes the average and max Act stats, used in inhibition
func (ly *Layer) AvgMaxAct() {
	for pi := range ly.Pools {
		pl := &ly.Pools[pi]
		pl.Act.Init()
		for ni := pl.StIdx; ni < pl.EdIdx; ni++ {
			nrn := &ly.Neurons[ni]
			pl.Act.UpdateVal(nrn.Act, ni)
		}
		pl.Act.CalcAvg()
	}
}

//////////////////////////////////////////////////////////////////////////////////////
//  Quarter

// QuarterFinal does updating after end of a quarter
func (ly *Layer) QuarterFinal(time *Time) {
	for pi := range ly.Pools {
		pl := &ly.Pools[pi]
		if time.Quarter == 2 {
			pl.ActM = pl.Act
		} else if time.Quarter == 3 {
			pl.ActP = pl.Act
		}
	}
	for ni := range ly.Neurons {
		nrn := &ly.Neurons[ni]
		if time.Quarter == 2 {
			nrn.ActM = nrn.Act
		} else if time.Quarter == 3 {
			nrn.ActP = nrn.Act
			nrn.ActDif = nrn.ActP - nrn.ActM
			nrn.ActAvg += ly.Act.Dt.AvgDt * (nrn.Act - nrn.ActAvg)
		}
	}
	ly.CosDiffFmActs()
}

// CosDiffFmActs computes the cosine difference in activation state between minus and plus phases.
// this is also used for modulating the amount of BCM hebbian learning
func (ly *Layer) CosDiffFmActs() {
	lpl := &ly.Pools[0]
	avgM := lpl.ActM.Avg
	avgP := lpl.ActM.Avg
	cosv := float32(0)
	ssm := float32(0)
	ssp := float32(0)
	for ni := range ly.Neurons {
		nrn := &ly.Neurons[ni]
		ap := nrn.ActP - avgP // zero mean
		am := nrn.ActM - avgM
		cosv += ap * am
		ssm += am * am
		ssp += ap * ap
	}

	dist := math32.Sqrt(ssm * ssp)
	if dist != 0 {
		cosv /= dist
	}
	ly.CosDiff.Cos = cosv

	ly.Learn.CosDiff.AvgVarFmCos(&ly.CosDiff.Avg, &ly.CosDiff.Var, ly.CosDiff.Cos)
	//  lay->lrate_mod = lay_lrate;
	//   if(cos_diff.lrate_mod && !cos_diff.lrmod_fm_trc) {
	//     lay->lrate_mod *= cos_diff.CosDiffLrateMod(lay->cos_diff, lay->cos_diff_avg,
	//                                                lay->cos_diff_var);
	//     if(cos_diff.set_net_unlrn && lay->lrate_mod == 0.0f) {
	//       net->unlearnable_trial = true;
	//     }
	//   }

	// todo: need layer type!
	//   if((lay->layer_type != LAYER_STATE::HIDDEN) || us->deep.IsTRC()) {
	//     lay->cos_diff_avg_lrn = 0.0f; // no bcm for TARGET layers; irrelevant for INPUT
	//     lay->mod_avg_l_lrn = 0.0f;
	//   } else {
	ly.CosDiff.AvgLrn = 1 - ly.CosDiff.Avg
	ly.CosDiff.ModAvgLLrn = ly.Learn.AvgL.ErrModFmLayErr(ly.CosDiff.AvgLrn)

	//	lay.AvgCosDiff.Increment(ly.CosDiff)
}
