// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import (
	"fmt"
	"io"

	"github.com/chewxy/math32"
	"github.com/emer/emergent/emer"
	"github.com/goki/ki/indent"
)

// leabra.Prjn is a basic Leabra projection with synaptic learning parameters
type Prjn struct {
	PrjnStru
	WtScale WtScaleParams  `desc:"weight scaling parameters: modulates overall strength of projection, using both absolute and relative factors"`
	Learn   LearnSynParams `desc:"synaptic-level learning parameters"`
	Syns    []Synapse      `desc:"synaptic state values, ordered by the sending layer units which owns them -- one-to-one with SConIdx array"`

	// misc state variables below:
	GScale float32         `desc:"scaling factor for integrating synaptic input conductances (G's) -- computed in TrialInit, incorporates running-average activity levels"`
	GInc   []float32       `desc:"local increment accumulator for synaptic conductance from sending units -- goes to either GeInc or GiInc on neuron depending on projection type -- this will be thread-safe"`
	WbRecv []WtBalRecvPrjn `desc:"weight balance state variables for this projection, one per recv neuron"`
}

// AsLeabra returns this prjn as a leabra.Prjn -- all derived prjns must redefine
// this to return the base Prjn type, so that the LeabraPrjn interface does not
// need to include accessors to all the basic stuff.
func (pj *Prjn) AsLeabra() *Prjn {
	return pj
}

func (pj *Prjn) Defaults() {
	pj.WtScale.Defaults()
	pj.Learn.Defaults()
	pj.GScale = 1
}

// UpdateParams updates all params given any changes that might have been made to individual values
func (pj *Prjn) UpdateParams() {
	pj.WtScale.Update()
	pj.Learn.Update()
}

// SetParams sets given parameters to this prjn, if the target type is Prjn
// calls UpdateParams to ensure derived parameters are all updated.
// If setMsg is true, then a message is printed to confirm each parameter that is set.
// it always prints a message if a parameter fails to be set.
func (pj *Prjn) SetParams(pars emer.Params, setMsg bool) bool {
	trg := pars.Target()
	if trg != "Prjn" {
		return false
	}
	pars.Set(pj, setMsg)
	pj.UpdateParams()
	return true
}

func (pj *Prjn) SynVarNames() []string {
	return SynapseVars
}

// SynVals returns values of given variable name on synapses for each synapse in the projection
// using the natural ordering of the synapses (sender based for Leabra)
func (pj *Prjn) SynVals(varnm string) []float32 {
	vl := make([]float32, len(pj.Syns))
	for si := range pj.Syns {
		sy := &pj.Syns[si]
		sv, ok := sy.VarByName(varnm)
		if ok {
			vl[si] = sv
		}
	}
	return vl
}

// SynVal returns value of given variable name on the synapse between given send, recv unit indexes -- returns error for access errors.
func (pj *Prjn) SynVal(varnm string, sidx, ridx int) (float32, error) {
	slay := pj.Send.(LeabraLayer).AsLeabra()
	rlay := pj.Recv.(LeabraLayer).AsLeabra()
	nr := len(rlay.Neurons)
	ns := len(slay.Neurons)
	if ridx >= nr {
		return 0, fmt.Errorf("Prjn.SynVal: recv unit index %v is > size of recv layer: %v", ridx, nr)
	}
	if sidx >= ns {
		return 0, fmt.Errorf("Prjn.SynVal: send unit index %v is > size of send layer: %v", sidx, ns)
	}
	nc := int(pj.RConN[ridx])
	st := int(pj.RConIdxSt[ridx])
	for ci := 0; ci < nc; ci++ {
		si := int(pj.RConIdx[st+ci])
		if si != sidx {
			continue
		}
		rsi := pj.RSynIdx[st+ci]
		sy := &pj.Syns[rsi]
		sv, ok := sy.VarByName(varnm)
		if ok {
			return sv, nil
		}
	}
	return 0, fmt.Errorf("Prjn.SynVal: recv unit index %v does not recv from send unit index %v, or variable name: %v not found in synapse", ridx, sidx, varnm)
}

// SetSynVal sets value of given variable name on the synapse between given send, recv unit indexes -- returns error for access errors.
func (pj *Prjn) SetSynVal(varnm string, sidx, ridx int, val float64) error {
	slay := pj.Send.(LeabraLayer).AsLeabra()
	rlay := pj.Recv.(LeabraLayer).AsLeabra()
	nr := len(rlay.Neurons)
	ns := len(slay.Neurons)
	if ridx >= nr {
		return fmt.Errorf("Prjn.SetSynVal: recv unit index %v is > size of recv layer: %v", ridx, nr)
	}
	if sidx >= ns {
		return fmt.Errorf("Prjn.SetSynVal: send unit index %v is > size of send layer: %v", sidx, ns)
	}
	nc := int(pj.RConN[ridx])
	st := int(pj.RConIdxSt[ridx])
	for ci := 0; ci < nc; ci++ {
		si := int(pj.RConIdx[st+ci])
		if si != sidx {
			continue
		}
		rsi := pj.RSynIdx[st+ci]
		sy := &pj.Syns[rsi]
		ok := sy.SetVarByName(varnm, val)
		if ok {
			if varnm == "Wt" {
				pj.Learn.LWtFmWt(sy)
			}
			return nil
		}
	}
	return fmt.Errorf("Prjn.SetSynVal: recv unit index %v does not recv from send unit index %v, or variable name: %v not found in synapse", ridx, sidx, varnm)
}

///////////////////////////////////////////////////////////////////////
//  Weights File

// WriteWtsJSON writes the weights from this projection from the receiver-side perspective
// in a JSON text format.  We build in the indentation logic to make it much faster and
// more efficient.
func (pj *Prjn) WriteWtsJSON(w io.Writer, depth int) {
	slay := pj.Send.(LeabraLayer).AsLeabra()
	rlay := pj.Recv.(LeabraLayer).AsLeabra()
	nr := len(rlay.Neurons)
	w.Write(indent.TabBytes(depth))
	w.Write([]byte("{\n"))
	depth++
	w.Write(indent.TabBytes(depth))
	w.Write([]byte(fmt.Sprintf("\"GScale\": %v\n", pj.GScale)))
	w.Write(indent.TabBytes(depth))
	w.Write([]byte(fmt.Sprintf("\"%v\": [\n", slay.Name)))
	depth++
	for ri := 0; ri < nr; ri++ {
		nc := int(pj.RConN[ri])
		st := int(pj.RConIdxSt[ri])
		w.Write(indent.TabBytes(depth))
		w.Write([]byte(fmt.Sprintf("\"%v\": {\n", ri)))
		depth++
		w.Write(indent.TabBytes(depth))
		w.Write([]byte(fmt.Sprintf("\"n\": %v,\n", nc)))
		w.Write(indent.TabBytes(depth))
		w.Write([]byte("\"Si\": ["))
		for ci := 0; ci < nc; ci++ {
			si := pj.RConIdx[st+ci]
			w.Write([]byte(fmt.Sprintf("%v ", si)))
		}
		w.Write([]byte("]\n"))
		w.Write(indent.TabBytes(depth))
		w.Write([]byte("\"Wt\": ["))
		for ci := 0; ci < nc; ci++ {
			rsi := pj.RSynIdx[st+ci]
			sy := &pj.Syns[rsi]
			w.Write([]byte(fmt.Sprintf("%v ", sy.Wt)))
		}
		w.Write([]byte("]\n"))
		depth--
		w.Write(indent.TabBytes(depth))
		if ri == nr-1 {
			w.Write([]byte("}\n"))
		} else {
			w.Write([]byte("},\n"))
		}
	}
	depth--
	w.Write(indent.TabBytes(depth))
	w.Write([]byte("]\n"))
	depth--
	w.Write(indent.TabBytes(depth))
	w.Write([]byte("}\n"))
}

// ReadWtsJSON reads the weights for this projection from the receiver-side perspective
// in a JSON text format.
func (pj *Prjn) ReadWtsJSON(r io.Reader) error {
	return nil
}

// Build constructs the full connectivity among the layers as specified in this projection.
// Calls PrjnStru.BuildStru and then allocates the synaptic values in Syns accordingly.
func (pj *Prjn) Build() error {
	if err := pj.BuildStru(); err != nil {
		return err
	}
	pj.Syns = make([]Synapse, len(pj.SConIdx))
	rsh := pj.Recv.LayShape()
	//	ssh := pj.Send.LayShape()
	rlen := rsh.Len()
	pj.GInc = make([]float32, rlen)
	pj.WbRecv = make([]WtBalRecvPrjn, rlen)
	return nil
}

//////////////////////////////////////////////////////////////////////////////////////
//  Init methods

// InitWts initializes weight values according to Learn.WtInit params
func (pj *Prjn) InitWts() {
	for si := range pj.Syns {
		sy := &pj.Syns[si]
		pj.Learn.InitWts(sy)
	}
	for wi := range pj.WbRecv {
		wb := &pj.WbRecv[wi]
		wb.Init()
	}
	pj.LeabraPrj.InitGInc()
}

// InitWtSym initializes weight symmetry -- is given the reciprocal projection where
// the Send and Recv layers are reversed.
func (pj *Prjn) InitWtSym(rpjp LeabraPrjn) {
	rpj := rpjp.AsLeabra()
	slay := pj.Send.(LeabraLayer).AsLeabra()
	ns := len(slay.Neurons)
	for si := 0; si < ns; si++ {
		nc := int(pj.SConN[si])
		st := int(pj.SConIdxSt[si])
		for ci := 0; ci < nc; ci++ {
			sy := &pj.Syns[st+ci]
			ri := pj.SConIdx[st+ci]
			// now we need to find the reciprocal synapse on rpj!
			// look in ri for sending connections
			rsi := ri
			rsnc := int(rpj.SConN[rsi])
			rsst := int(rpj.SConIdxSt[rsi])
			for rci := 0; rci < rsnc; rci++ {
				rri := int(rpj.SConIdx[rsst+rci])
				if rri == si {
					rsy := &rpj.Syns[rsst+rci]
					rsy.Wt = sy.Wt
					rsy.LWt = sy.LWt
					// note: if we support SymFmTop then can have option to go other way
					// also for Scale support, copy scales
				}
			}
		}
	}
}

// IniteGInc initializes the per-projection GInc threadsafe increment -- not
// typically needed (called during InitWts only) but can be called when needed
func (pj *Prjn) InitGInc() {
	for ri := range pj.GInc {
		pj.GInc[ri] = 0
	}
}

//////////////////////////////////////////////////////////////////////////////////////
//  Act methods

// SendGDelta sends the delta-activation from sending neuron index si,
// to integrate synaptic conductances on receivers
func (pj *Prjn) SendGDelta(si int, delta float32) {
	scdel := delta * pj.GScale
	nc := pj.SConN[si]
	st := pj.SConIdxSt[si]
	syns := pj.Syns[st : st+nc]
	scons := pj.SConIdx[st : st+nc]
	for ci := range syns {
		ri := scons[ci]
		pj.GInc[ri] += scdel * syns[ci].Wt
	}
}

// RecvGInc increments the receiver's GeInc or GiInc from that of all the projections.
func (pj *Prjn) RecvGInc() {
	rlay := pj.Recv.(LeabraLayer).AsLeabra()
	if pj.Type == emer.Inhib {
		for ri := range rlay.Neurons {
			rn := &rlay.Neurons[ri]
			rn.GiInc += pj.GInc[ri]
			pj.GInc[ri] = 0
		}
	} else {
		for ri := range rlay.Neurons {
			rn := &rlay.Neurons[ri]
			rn.GeInc += pj.GInc[ri]
			pj.GInc[ri] = 0
		}
	}
}

//////////////////////////////////////////////////////////////////////////////////////
//  Learn methods

// DWt computes the weight change (learning) -- on sending projections
func (pj *Prjn) DWt() {
	if !pj.Learn.Learn {
		return
	}
	slay := pj.Send.(LeabraLayer).AsLeabra()
	rlay := pj.Recv.(LeabraLayer).AsLeabra()
	for si := range slay.Neurons {
		sn := &slay.Neurons[si]
		if sn.AvgS < pj.Learn.XCal.LrnThr && sn.AvgM < pj.Learn.XCal.LrnThr {
			continue
		}
		nc := int(pj.SConN[si])
		st := int(pj.SConIdxSt[si])
		syns := pj.Syns[st : st+nc]
		scons := pj.SConIdx[st : st+nc]
		for ci := range syns {
			sy := &syns[ci]
			ri := scons[ci]
			rn := &rlay.Neurons[ri]
			err, bcm := pj.Learn.CHLdWt(sn.AvgSLrn, sn.AvgM, rn.AvgSLrn, rn.AvgM, rn.AvgL)

			bcm *= pj.Learn.XCal.LongLrate(rn.AvgLLrn)
			err *= pj.Learn.XCal.MLrn
			dwt := bcm + err
			norm := float32(1)
			if pj.Learn.Norm.On {
				norm = pj.Learn.Norm.NormFmAbsDWt(&sy.Norm, math32.Abs(dwt))
			}
			if pj.Learn.Momentum.On {
				dwt = norm * pj.Learn.Momentum.MomentFmDWt(&sy.Moment, dwt)
			} else {
				dwt *= norm
			}
			sy.DWt += pj.Learn.Lrate * dwt
		}
		// aggregate max DWtNorm over sending synapses
		if pj.Learn.Norm.On {
			maxNorm := float32(0)
			for ci := range syns {
				sy := &syns[ci]
				if sy.Norm > maxNorm {
					maxNorm = sy.Norm
				}
			}
			for ci := range syns {
				sy := &syns[ci]
				sy.Norm = maxNorm
			}
		}
	}
}

// WtFmDWt updates the synaptic weight values from delta-weight changes -- on sending projections
func (pj *Prjn) WtFmDWt() {
	if !pj.Learn.Learn {
		return
	}
	if pj.Learn.WtBal.On {
		for si := range pj.Syns {
			sy := &pj.Syns[si]
			ri := pj.SConIdx[si]
			wb := &pj.WbRecv[ri]
			pj.Learn.WtFmDWt(wb.Inc, wb.Dec, &sy.DWt, &sy.Wt, &sy.LWt)
		}
	} else {
		for si := range pj.Syns {
			sy := &pj.Syns[si]
			pj.Learn.WtFmDWt(1, 1, &sy.DWt, &sy.Wt, &sy.LWt)
		}
	}
}

// WtBalFmWt computes the Weight Balance factors based on average recv weights
func (pj *Prjn) WtBalFmWt() {
	if !pj.Learn.Learn || !pj.Learn.WtBal.On {
		return
	}

	rlay := pj.Recv.(LeabraLayer).AsLeabra()
	if rlay.Type == emer.Target {
		return
	}
	for ri := range rlay.Neurons {
		nc := int(pj.RConN[ri])
		if nc <= 1 {
			continue
		}
		rn := &rlay.Neurons[ri]
		if rn.HasFlag(NeurHasTarg) { // todo: ensure that Pulvinar has this set, or do something else
			continue
		}
		wb := &pj.WbRecv[ri]
		st := int(pj.RConIdxSt[ri])
		rsidxs := pj.RSynIdx[st : st+nc]
		sumWt := float32(0)
		sumN := 0
		for ci := range rsidxs {
			rsi := rsidxs[ci]
			sy := &pj.Syns[rsi]
			if sy.Wt >= pj.Learn.WtBal.AvgThr {
				sumWt += sy.Wt
				sumN++
			}
		}
		if sumN > 0 {
			sumWt /= float32(sumN)
		} else {
			sumWt = 0
		}
		wb.Avg = sumWt
		wb.Fact, wb.Inc, wb.Dec = pj.Learn.WtBal.WtBal(sumWt, rn.ActAvg)
	}
}

///////////////////////////////////////////////////////////////////////
//  WtBalRecvPrjn

// WtBalRecvPrjn are state variables used in computing the WtBal weight balance function
// There is one of these for each Recv Neuron participating in the projection.
type WtBalRecvPrjn struct {
	Avg  float32 `desc:"average of effective weight values that exceed WtBal.AvgThr across given Recv Neuron's connections for given Prjn"`
	Fact float32 `desc:"overall weight balance factor that drives changes in WbInc vs. WbDec via a sigmoidal function -- this is the net strength of weight balance changes"`
	Inc  float32 `desc:"weight balance increment factor -- extra multiplier to add to weight increases to maintain overall weight balance"`
	Dec  float32 `desc:"weight balance decrement factor -- extra multiplier to add to weight decreases to maintain overall weight balance"`
}

func (wb *WtBalRecvPrjn) Init() {
	wb.Avg = 0
	wb.Fact = 0
	wb.Inc = 1
	wb.Dec = 1
}
