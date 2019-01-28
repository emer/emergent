// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import (
	"fmt"
	"io"

	"github.com/chewxy/math32"
	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/erand"
	"github.com/emer/emergent/etensor"
	"github.com/goki/ki/bitflag"
	"github.com/goki/ki/indent"
	"github.com/goki/ki/ints"
	"github.com/goki/ki/kit"
)

// leabra.LayerStru manages the structural elements of the layer, which are common
// to any Layer type
type LayerStru struct {
	Name      string        `desc:"Name of the layer -- this must be unique within the network, which has a map for quick lookup and layers are typically accessed directly by name"`
	Class     string        `desc:"Class is for applying parameter styles, can be space separated multple tags"`
	Off       bool          `desc:"inactivate this layer -- allows for easy experimentation"`
	Shape     etensor.Shape `desc:"shape of the layer -- can be 2D for basic layers and 4D for layers with sub-groups (hypercolumns) -- order is outer-to-inner (row major) so Y then X for 2D and for 4D: Y-X unit pools then Y-X units within pools"`
	Type      LayerType     `desc:"type of layer -- Hidden, Input, Target, Compare"`
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

// LayerType is the type of the layer: Input, Hidden, Target, Compare
type LayerType int32

//go:generate stringer -type=LayerType

var KiT_LayerType = kit.Enums.AddEnum(LayerTypeN, false, nil)

func (ev LayerType) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *LayerType) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }

// The layer types
const (
	// Hidden is an internal representational layer that does not receive direct input / targets
	Hidden LayerType = iota

	// Input is a layer that receives direct external input in its Ext inputs
	Input

	// Target is a layer that receives direct external target inputs used for driving plus-phase learning
	Target

	// Compare is a layer that receives external comparison inputs, which drive statistics but
	// do NOT drive activation or learning directly
	Compare

	LayerTypeN
)

//////////////////////////////////////////////////////////////////////////////////////
//  Layer

// todo: need overall good strategy for stats

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
	ly.Inhib.Layer.On = true
	for _, pj := range ly.RecvPrjns {
		pj.Defaults()
	}
}

// UpdateParams updates all params given any changes that might have been made to individual values
// including those in the receiving projections of this layer
func (ly *Layer) UpdateParams() {
	ly.Act.Update()
	ly.Inhib.Update()
	ly.Learn.Update()
	for _, pj := range ly.RecvPrjns {
		pj.UpdateParams()
	}
}

// SetParams sets given parameters to this layer, if the target type is Layer
// calls UpdateParams to ensure derived parameters are all updated
func (ly *Layer) SetParams(pars emer.Params) bool {
	trg := pars.Target()
	if trg != "Layer" {
		return false
	}
	pars.Set(ly)
	ly.UpdateParams()
	return true
}

// StyleParam applies a given style to either this layer or the receiving projections in this layer
// depending on the style specification (.Class, #Name, Type) and target value of params.
// returns true if applied successfully.
func (ly *Layer) StyleParam(sty string, pars emer.Params) bool {
	if emer.StyleMatch(sty, ly.Name, ly.Class, "Layer") {
		if ly.SetParams(pars) {
			return true // done -- otherwise, might be for prjns
		}
	}
	set := false
	for _, pj := range ly.RecvPrjns {
		did := pj.StyleParam(sty, pars)
		if did {
			set = true
		}
	}
	return set
}

// StyleParams applies a given styles to either this layer or the receiving projections in this layer
// depending on the style specification (.Class, #Name, Type) and target value of params
func (ly *Layer) StyleParams(psty emer.ParamStyle) {
	for sty, pars := range psty {
		ly.StyleParam(sty, pars)
	}
}

// Unit is emer.Layer interface method to get given Unit
// only possible with Neurons in place
func (ly *Layer) Unit(idx []int) emer.Unit {
	fidx := ly.Shape.Offset(idx)
	if int(fidx) < len(ly.Neurons) {
		return &ly.Neurons[fidx]
	}
	return nil
}

// UnitVals is emer.Layer interface method to return values of given variable
func (ly *Layer) UnitVals(varNm string) []float32 {
	vs := make([]float32, len(ly.Neurons))
	for i := range ly.Neurons {
		nrn := &ly.Neurons[i]
		vl, _ := nrn.VarByName(varNm)
		vs[i] = vl
	}
	return vs
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

// WriteWtsJSON writes the weights from this layer from the receiver-side perspective
// in a JSON text format.  We build in the indentation logic to make it much faster and
// more efficient.
func (ly *Layer) WriteWtsJSON(w io.Writer, depth int) {
	w.Write(indent.TabBytes(depth))
	w.Write([]byte("{\n"))
	depth++
	w.Write(indent.TabBytes(depth))
	w.Write([]byte(fmt.Sprintf("\"%v\": [\n", ly.Name)))
	depth++
	for _, pj := range ly.RecvPrjns {
		if pj.IsOff() {
			continue
		}
		pj.WriteWtsJSON(w, depth)
	}
	depth--
	w.Write(indent.TabBytes(depth))
	w.Write([]byte("],\n"))
	depth--
	w.Write(indent.TabBytes(depth))
	w.Write([]byte("},\n"))
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

// InitExt initializes external input state -- called prior to apply ext
func (ly *Layer) InitExt() {
	msk := bitflag.Mask32(int(NeurHasExt), int(NeurHasTarg), int(NeurHasCmpr))
	for ni := range ly.Neurons {
		nrn := &ly.Neurons[ni]
		nrn.Ext = 0
		nrn.Targ = 0
		nrn.ClearMask(msk)
	}
}

// ApplyExt applies external input in the form of an arrow tensor.Float32
// If the layer is a Target or Compare layer type, then it goes in Targ
// otherwise it goes in Ext
func (ly *Layer) ApplyExt(ext *etensor.Float32) {
	// todo: compare shape?
	clrmsk := bitflag.Mask32(int(NeurHasExt), int(NeurHasTarg), int(NeurHasCmpr))
	setmsk := int32(0)
	toTarg := false
	if ly.Type == Target {
		setmsk = bitflag.Mask32(int(NeurHasTarg))
		toTarg = true
	} else if ly.Type == Compare {
		setmsk = bitflag.Mask32(int(NeurHasCmpr))
		toTarg = true
	} else {
		setmsk = bitflag.Mask32(int(NeurHasExt))
	}
	ev := ext.Values
	mx := ints.MinInt(len(ev), len(ly.Neurons))
	for i := 0; i < mx; i++ {
		nrn := &ly.Neurons[i]
		vl := ev[i]
		if toTarg {
			nrn.Targ = vl
		} else {
			nrn.Ext = vl
		}
		nrn.ClearMask(clrmsk)
		nrn.SetMask(setmsk)
	}
}

// TrialInit handles all initialization at start of new input pattern, including computing
// netinput scaling from running average activation etc.
// should already have presented the external input to the network at this point.
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
	if ly.Act.Clamp.Hard && ly.Type == Input {
		ly.HardClamp()
	}
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

// GenNoise generates random noise for all neurons
func (ly *Layer) GenNoise() {
	for ni := range ly.Neurons {
		nrn := &ly.Neurons[ni]
		nrn.Noise = float32(ly.Act.Noise.Gen(-1))
	}
}

// DecayState decays activation state by given proportion (default is on ly.Act.Init.Decay)
func (ly *Layer) DecayState(decay float32) {
	for ni := range ly.Neurons {
		nrn := &ly.Neurons[ni]
		ly.Act.DecayState(nrn, decay)
	}
}

// HardClamp hard-clamps the activations in the layer -- called during TrialInit for hard-clamped Input layers
func (ly *Layer) HardClamp() {
	for ni := range ly.Neurons {
		nrn := &ly.Neurons[ni]
		ly.Act.HardClamp(nrn)
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
		if time.Quarter == 2 { // end of minus phase
			nrn.ActM = nrn.Act
			if nrn.HasFlag(NeurHasTarg) { // will be clamped in plus phase
				nrn.Ext = nrn.Targ
				nrn.SetFlag(NeurHasExt)
			}
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

	if ly.Type != Hidden {
		ly.CosDiff.AvgLrn = 0 // no BCM for non-hidden layers
		ly.CosDiff.ModAvgLLrn = 0
	} else {
		ly.CosDiff.AvgLrn = 1 - ly.CosDiff.Avg
		ly.CosDiff.ModAvgLLrn = ly.Learn.AvgL.ErrModFmLayErr(ly.CosDiff.AvgLrn)
	}
}

// DWt computes the weight change (learning) -- calls DWt method on sending projections
func (ly *Layer) DWt() {
	for _, pj := range ly.SendPrjns {
		pj.DWt()
	}
}

// WtFmDWt updates the weights from delta-weight changes -- on the sending projections
func (ly *Layer) WtFmDWt() {
	for _, pj := range ly.SendPrjns {
		pj.WtFmDWt()
	}
}
