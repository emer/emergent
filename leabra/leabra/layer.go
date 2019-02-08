// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import (
	"errors"
	"fmt"
	"io"

	"github.com/chewxy/math32"
	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/erand"
	"github.com/emer/emergent/etensor"
	"github.com/goki/ki/bitflag"
	"github.com/goki/ki/indent"
	"github.com/goki/ki/ints"
)

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

// AsLeabra returns this layer as a leabra.Layer -- all derived layers must redefine
// this to return the base Layer type, so that the LeabraLayer interface does not
// need to include accessors to all the basic stuff
func (ly *Layer) AsLeabra() *Layer {
	return ly
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
	return NeuronVars
}

// UnitVals is emer.Layer interface method to return values of given variable
func (ly *Layer) UnitVals(varNm string) ([]float32, error) {
	vidx, err := NeuronVarByName(varNm)
	if err != nil {
		return nil, err
	}
	vs := make([]float32, len(ly.Neurons))
	for i := range ly.Neurons {
		nrn := &ly.Neurons[i]
		vs[i] = nrn.VarByIndex(vidx)
	}
	return vs, nil
}

// UnitVal returns value of given variable name on given unit,
// using shape-based dimensional index
func (ly *Layer) UnitVal(varNm string, idx []int) (float32, error) {
	fidx := ly.Shape.Offset(idx)
	nn := len(ly.Neurons)
	if fidx < 0 || fidx >= nn {
		return 0, fmt.Errorf("Layer UnitVal index: %v out of range, N = %v", fidx, nn)
	}
	nrn := &ly.Neurons[fidx]
	return nrn.VarByName(varNm)
}

// UnitVal1D returns value of given variable name on given unit,
// using 1-dimensional index.
func (ly *Layer) UnitVal1D(varNm string, idx int) (float32, error) {
	nn := len(ly.Neurons)
	if idx < 0 || idx >= nn {
		return 0, fmt.Errorf("Layer UnitVal1D index: %v out of range, N = %v", idx, nn)
	}
	nrn := &ly.Neurons[idx]
	return nrn.VarByName(varNm)
}

//////////////////////////////////////////////////////////////////////////////////////
//  Build

// BuildSubPools initializes neuron start / end indexes for sub-group pools
func (ly *Layer) BuildSubPools() {
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

// BuildPools builds the inhibitory pools structures -- nu = number of units in layer
func (ly *Layer) BuildPools(nu int) error {
	np := 1
	if ly.Inhib.Pool.On {
		np += ly.NPools()
	}
	ly.Pools = make([]Pool, np)
	lpl := &ly.Pools[0]
	lpl.StIdx = 0
	lpl.EdIdx = nu
	if np > 1 {
		ly.BuildSubPools()
	}
	return nil
}

// BuildPrjns builds the projections, recv-side
func (ly *Layer) BuildPrjns() error {
	emsg := ""
	for _, pj := range ly.RecvPrjns {
		if pj.IsOff() {
			continue
		}
		err := pj.Build()
		if err != nil {
			emsg += err.Error() + "\n"
		}
	}
	if emsg != "" {
		return errors.New(emsg)
	}
	return nil
}

// Build constructs the layer state, including calling Build on the projections
// you MUST have properly configured the Inhib.Pool.On setting by this point
// to properly allocate Pools for the unit groups if necessary.
func (ly *Layer) Build() error {
	nu := ly.Shape.Len()
	if nu == 0 {
		return fmt.Errorf("Build Layer %v: no units specified in Shape", ly.Name)
	}
	ly.Neurons = make([]Neuron, nu)
	err := ly.BuildPools(nu)
	if err != nil {
		return err
	}
	err = ly.BuildPrjns()
	return err
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
	// todo: save average activity state
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

// ReadWtsJSON reads the weights from this layer from the receiver-side perspective
// in a JSON text format.
func (ly *Layer) ReadWtsJSON(r io.Reader) error {
	return nil
}

// note: all basic computation can be performed on layer-level and prjn level

//////////////////////////////////////////////////////////////////////////////////////
//  Init methods

// InitWts initializes the weight values in the network, i.e., resetting learning
// Also calls InitActs
func (ly *Layer) InitWts() {
	for _, p := range ly.SendPrjns {
		if p.IsOff() {
			continue
		}
		p.(LeabraPrjn).InitWts()
	}
	for pi := range ly.Pools {
		pl := &ly.Pools[pi]
		pl.ActAvg.ActMAvg = ly.Inhib.ActAvg.Init
		pl.ActAvg.ActPAvg = ly.Inhib.ActAvg.Init
		pl.ActAvg.ActPAvgEff = ly.Inhib.ActAvg.EffInit()
	}
	ly.LeabraLay.InitActAvg()
	ly.LeabraLay.InitActs()
	ly.CosDiff.Init()
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

// InitWtsSym initializes the weight symmetry -- higher layers copy weights from lower layers
func (ly *Layer) InitWtSym() {
	for _, p := range ly.SendPrjns {
		if p.IsOff() {
			continue
		}
		// key ordering constraint on which way weights are copied
		if p.RecvLay().LayIndex() < p.SendLay().LayIndex() {
			continue
		}
		rpj, has := ly.RecipToSendPrjn(p)
		if !has {
			continue
		}
		p.(LeabraPrjn).InitWtSym(rpj.(LeabraPrjn))
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

// ApplyExt applies external input in the form of an etensor.Float32.
// If the layer is a Target or Compare layer type, then it goes in Targ
// otherwise it goes in Ext
func (ly *Layer) ApplyExt(ext *etensor.Float32) {
	// todo: compare shape?
	clrmsk := bitflag.Mask32(int(NeurHasExt), int(NeurHasTarg), int(NeurHasCmpr))
	setmsk := int32(0)
	toTarg := false
	if ly.Type == emer.Target {
		setmsk = bitflag.Mask32(int(NeurHasTarg))
		toTarg = true
	} else if ly.Type == emer.Compare {
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
	ly.LeabraLay.AvgLFmAvgM()
	for pi := range ly.Pools {
		pl := &ly.Pools[pi]
		ly.Inhib.ActAvg.AvgFmAct(&pl.ActAvg.ActMAvg, pl.ActM.Avg)
		ly.Inhib.ActAvg.AvgFmAct(&pl.ActAvg.ActPAvg, pl.ActP.Avg)
		ly.Inhib.ActAvg.EffFmAvg(&pl.ActAvg.ActPAvgEff, pl.ActAvg.ActPAvg)
	}
	ly.LeabraLay.GScaleFmAvgAct()
	if ly.Act.Noise.Type != NoNoise && ly.Act.Noise.TrialFixed && ly.Act.Noise.Dist != erand.None {
		ly.LeabraLay.GenNoise()
	}
	ly.LeabraLay.DecayState(ly.Act.Init.Decay)
	if ly.Act.Clamp.Hard && ly.Type == emer.Input {
		ly.LeabraLay.HardClamp()
	}
}

// AvgLFmAvgM updates AvgL long-term running average activation that drives BCM Hebbian learning
func (ly *Layer) AvgLFmAvgM() {
	for ni := range ly.Neurons {
		nrn := &ly.Neurons[ni]
		ly.Learn.AvgLFmAvgM(nrn)
		if ly.Learn.AvgL.ErrMod {
			nrn.AvgLLrn *= ly.CosDiff.ModAvgLLrn
		}
	}
}

// GScaleFmAvgAct computes the scaling factor for synaptic input conductances G,
// based on sending layer average activation.
// This attempts to automatically adjust for overall differences in raw activity coming into the units
// to achieve a general target of around .5 to 1 for the integrated Ge value.
func (ly *Layer) GScaleFmAvgAct() {
	totGeRel := float32(0)
	totGiRel := float32(0)
	for _, p := range ly.RecvPrjns {
		if p.IsOff() {
			continue
		}
		pj := p.(LeabraPrjn).AsLeabra()
		slay := p.SendLay().(LeabraLayer).AsLeabra()
		slpl := slay.Pools[0]
		savg := slpl.ActAvg.ActPAvgEff
		snu := len(slay.Neurons)
		ncon := pj.RConNAvgMax.Avg
		pj.GScale = pj.WtScale.FullScale(savg, float32(snu), ncon)
		if pj.Type == emer.Inhib {
			totGiRel += pj.WtScale.Rel
		} else {
			totGeRel += pj.WtScale.Rel
		}
	}

	for _, p := range ly.RecvPrjns {
		if p.IsOff() {
			continue
		}
		pj := p.(LeabraPrjn).AsLeabra()
		if pj.Type == emer.Inhib {
			if totGiRel > 0 {
				pj.GScale /= totGiRel
			}
		} else {
			if totGeRel > 0 {
				pj.GScale /= totGeRel
			}
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
	for pi := range ly.Pools { // decaying average act is essential for inhib
		pl := &ly.Pools[pi]
		pl.Act.Max -= decay * pl.Act.Max
		pl.Act.Avg -= decay * pl.Act.Avg
		pl.Inhib.FFi -= decay * pl.Inhib.FFi
		pl.Inhib.FBi -= decay * pl.Inhib.FBi
		pl.Inhib.Gi -= decay * pl.Inhib.Gi
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

// InitGInc initializes GeInc and GiIn increment -- optional
func (ly *Layer) InitGInc() {
	for ni := range ly.Neurons {
		nrn := &ly.Neurons[ni]
		nrn.GeInc = 0
		nrn.GiInc = 0
	}
	for _, p := range ly.RecvPrjns {
		if p.IsOff() {
			continue
		}
		p.(LeabraPrjn).InitGInc()
	}
}

// SendGDelta sends change in activation since last sent, to increment recv
// synaptic conductances G, if above thresholds
func (ly *Layer) SendGDelta() {
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
					sp.(LeabraPrjn).SendGDelta(ni, delta)
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
				sp.(LeabraPrjn).SendGDelta(ni, delta)
			}
			nrn.ActSent = 0
		}
	}
}

// GFmInc integrates new synaptic conductances from increments sent during last SendGDelta.
func (ly *Layer) GFmInc() {
	for _, p := range ly.RecvPrjns {
		if p.IsOff() {
			continue
		}
		p.(LeabraPrjn).RecvGInc()
	}
	for ni := range ly.Neurons {
		nrn := &ly.Neurons[ni]
		ly.Act.GeGiFmInc(nrn)
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
				nrn.Gi = pl.Inhib.Gi + nrn.GiSelf + nrn.GiSyn
			}
		}
	} else {
		for ni := lpl.StIdx; ni < lpl.EdIdx; ni++ {
			nrn := &ly.Neurons[ni]
			ly.Inhib.Self.Inhib(&nrn.GiSelf, nrn.Act)
			nrn.Gi = lpl.Inhib.Gi + nrn.GiSelf + nrn.GiSyn
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
	if time.Quarter == 3 {
		ly.LeabraLay.CosDiffFmActs()
	}
}

// CosDiffFmActs computes the cosine difference in activation state between minus and plus phases.
// this is also used for modulating the amount of BCM hebbian learning
func (ly *Layer) CosDiffFmActs() {
	lpl := &ly.Pools[0]
	avgM := lpl.ActM.Avg
	avgP := lpl.ActP.Avg
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

	if ly.Type != emer.Hidden {
		ly.CosDiff.AvgLrn = 0 // no BCM for non-hidden layers
		ly.CosDiff.ModAvgLLrn = 0
	} else {
		ly.CosDiff.AvgLrn = 1 - ly.CosDiff.Avg
		ly.CosDiff.ModAvgLLrn = ly.Learn.AvgL.ErrModFmLayErr(ly.CosDiff.AvgLrn)
	}
}

// DWt computes the weight change (learning) -- calls DWt method on sending projections
func (ly *Layer) DWt() {
	for _, p := range ly.SendPrjns {
		if p.IsOff() {
			continue
		}
		p.(LeabraPrjn).DWt()
	}
}

// WtFmDWt updates the weights from delta-weight changes -- on the sending projections
func (ly *Layer) WtFmDWt() {
	for _, p := range ly.SendPrjns {
		if p.IsOff() {
			continue
		}
		p.(LeabraPrjn).WtFmDWt()
	}
}

// WtBalFmWt computes the Weight Balance factors based on average recv weights
func (ly *Layer) WtBalFmWt() {
	for _, p := range ly.RecvPrjns {
		if p.IsOff() {
			continue
		}
		p.(LeabraPrjn).WtBalFmWt()
	}
}

//////////////////////////////////////////////////////////////////////////////////////
//  Stats

// SSE returns the sum-squared-error and avg-squared-error
// over the layer, in terms of ActP - ActM (valideven on non-target layers FWIW).
// Uses the given tolerance per-unit to count an error at all
// (e.g., .5 = activity just has to be on the right side of .5).
func (ly *Layer) SSE(tol float32) (sum, avg float32) {
	nn := len(ly.Neurons)
	if nn == 0 {
		return 0, 0
	}
	sse := float32(0)
	for ni := range ly.Neurons {
		nrn := &ly.Neurons[ni]
		d := nrn.ActP - nrn.ActM
		if math32.Abs(d) < tol {
			continue
		}
		sse += d * d
	}
	return sse, sse / float32(nn)
}
