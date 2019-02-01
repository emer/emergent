// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import (
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/chewxy/math32"
	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/etensor"
	"github.com/emer/emergent/prjn"
	"github.com/goki/ki/indent"
)

// PrjnStru contains the basic structural information for specifying a projection of synaptic
// connections between two layers, and maintaining all the synaptic connection-level data.
// The exact same struct object is added to the Recv and Send layers, and it manages everything
// about the connectivity, and methods on the Prjn handle all the relevant computation.
type PrjnStru struct {
	Off         bool         `desc:"inactivate this projection -- allows for easy experimentation"`
	Class       string       `desc:"Class is for applying parameter styles, can be space separated multple tags"`
	Notes       string       `desc:"can record notes about this projection here"`
	Recv        emer.Layer   `desc:"receiving layer for this projection -- the emer.Layer interface can be converted to the specific Layer type you are using, e.g., rlay := prjn.Recv.(*leabra.Layer)"`
	Send        emer.Layer   `desc:"sending layer for this projection"`
	Pat         prjn.Pattern `desc:"pattern of connectivity"`
	RConN       []int32      `desc:"number of recv connections for each neuron in the receiving layer, as a flat list"`
	RConNAvgMax emer.AvgMax  `desc:"average and maximum number of recv connections in the receiving layer"`
	RConIdxSt   []int32      `desc:"starting index into ConIdx list for each neuron in receiving layer -- just a list incremented by ConN"`
	RConIdx     []int32      `desc:"index of other neuron on sending side of projection, ordered by the receiving layer's order of units as the outer loop (each start is in ConIdxSt), and then by the sending layer's units within that"`
	RSynIdx     []int32      `desc:"index of synaptic state values for each recv unit x connection, for the receiver projection which does not own the synapses, and instead indexes into sender-ordered list"`
	SConN       []int32      `desc:"number of sending connections for each neuron in the sending layer, as a flat list"`
	SConNAvgMax emer.AvgMax  `desc:"average and maximum number of sending connections in the sending layer"`
	SConIdxSt   []int32      `desc:"starting index into ConIdx list for each neuron in sending layer -- just a list incremented by ConN"`
	SConIdx     []int32      `desc:"index of other neuron on receiving side of projection, ordered by the sending layer's order of units as the outer loop (each start is in ConIdxSt), and then by the sending layer's units within that"`
}

// emer.Prjn interface

func (ps *PrjnStru) PrjnClass() string { return ps.Class }
func (ps *PrjnStru) PrjnName() string {
	return ps.Recv.LayName() + "Fm" + ps.Send.LayName()
}
func (ps *PrjnStru) RecvLay() emer.Layer   { return ps.Recv }
func (ps *PrjnStru) SendLay() emer.Layer   { return ps.Send }
func (ps *PrjnStru) Pattern() prjn.Pattern { return ps.Pat }

func (ps *PrjnStru) IsOff() bool {
	return ps.Off || ps.Recv.IsOff() || ps.Send.IsOff()
}

// Connect sets the connectivity between two layers and the pattern to use in interconnecting them
func (ps *PrjnStru) Connect(rlay, slay emer.Layer, pat prjn.Pattern) {
	ps.Recv = rlay
	ps.Send = slay
	ps.Pat = pat
}

// Validate tests for non-nil settings for the projection -- returns error
// message or nil if no problems (and logs them if logmsg = true)
func (ps *PrjnStru) Validate(logmsg bool) error {
	emsg := ""
	if ps.Pat == nil {
		emsg += "Pat is nil; "
	}
	if ps.Recv == nil {
		emsg += "Recv is nil; "
	}
	if ps.Send == nil {
		emsg += "Send is nil; "
	}
	if emsg != "" {
		err := errors.New(emsg)
		if logmsg {
			log.Println(emsg)
		}
		return err
	}
	return nil
}

// BuildStru constructs the full connectivity among the layers as specified in this projection.
// Calls Validate and returns false if invalid.
// Pat.Connect is called to get the pattern of the connection.
// Then the connection indexes are configured according to that pattern.
func (ps *PrjnStru) BuildStru() bool {
	if ps.Off {
		return true
	}
	err := ps.Validate(true)
	if err != nil {
		return false
	}
	rsh := ps.Recv.LayShape()
	ssh := ps.Send.LayShape()
	recvn, sendn, cons := ps.Pat.Connect(rsh, ssh, ps.Recv == ps.Send)
	rlen := rsh.Len()
	slen := ssh.Len()
	tconr := ps.SetNIdxSt(&ps.RConN, &ps.RConNAvgMax, &ps.RConIdxSt, recvn)
	tcons := ps.SetNIdxSt(&ps.SConN, &ps.SConNAvgMax, &ps.SConIdxSt, sendn)
	if tconr != tcons {
		log.Printf("%v programmer error: total recv cons %v != total send cons %v\n", ps.String(), tconr, tcons)
	}
	ps.RConIdx = make([]int32, tconr)
	ps.RSynIdx = make([]int32, tconr)
	ps.SConIdx = make([]int32, tcons)

	sconN := make([]int32, slen) // temporary mem needed to tracks cur n of sending cons

	cbits := cons.Values
	for ri := 0; ri < rlen; ri++ {
		rbi := ri * slen     // recv bit index
		rtcn := ps.RConN[ri] // number of cons
		rst := ps.RConIdxSt[ri]
		rci := int32(0)
		for si := 0; si < slen; si++ {
			if !cbits.Index(rbi + si) { // no connection
				continue
			}
			sst := ps.SConIdxSt[si]
			if rci >= rtcn {
				log.Printf("%v programmer error: recv target total con number: %v exceeded at recv idx: %v, send idx: %v\n", ps.String(), rtcn, ri, si)
				break
			}
			ps.RConIdx[rst+rci] = int32(si)

			sci := sconN[si]
			stcn := ps.SConN[si]
			if sci >= stcn {
				log.Printf("%v programmer error: send target total con number: %v exceeded at recv idx: %v, send idx: %v\n", ps.String(), stcn, ri, si)
				break
			}
			ps.SConIdx[sst+sci] = int32(ri)
			ps.RSynIdx[rst+rci] = sst + sci
			(sconN[si])++
			rci++
		}
	}
	return true
}

// SetNIdxSt sets the *ConN and *ConIdxSt values given n tensor from Pat.
// Returns total number of connections for this direction.
func (ps *PrjnStru) SetNIdxSt(n *[]int32, avgmax *emer.AvgMax, idxst *[]int32, tn *etensor.Int32) int32 {
	ln := tn.Len()
	tnv := tn.Values
	*n = make([]int32, ln)
	*idxst = make([]int32, ln)
	idx := int32(0)
	avgmax.Init()
	for i := 0; i < ln; i++ {
		nv := tnv[i]
		(*n)[i] = nv
		(*idxst)[i] = idx
		idx += nv
		avgmax.UpdateVal(float32(nv), i)
	}
	avgmax.CalcAvg()
	return idx
}

// String satisfies fmt.Stringer for prjn
func (ps *PrjnStru) String() string {
	str := ""
	if ps.Recv == nil {
		str += "recv=nil; "
	} else {
		str += ps.Recv.LayName() + " <- "
	}
	if ps.Send == nil {
		str += "send=nil"
	} else {
		str += ps.Send.LayName()
	}
	if ps.Pat == nil {
		str += " Pat=nil"
	} else {
		str += " Pat=" + ps.Pat.Name()
	}
	return str
}

///////////////////////////////////////////////////////////////////////
//  WtBalRecvPrjn

// WtBalRecvPrjn are state variables used in computing the WtBal weight balance function
// There is one of these for each Recv Neuron participating in the projection.
type WtBalRecvPrjn struct {
	Avg  float32 `desc:"average of effective weight values that exceed WtBal.AvgThr across given Recv Neuron's connections for given Prjn"`
	Fact float32 `desc:"overall weight balance factor that drives changes in WbInc vs. WbDec via a sigmoidal function -- this is the net strength of weight balance changes"`
	Inc  float32 `desc:"weight balance increment factor -- extra multiplier to add to weight increases to maintain overall weight balance"`
	Dec  float32 `desc:"weight balance decrement factor -- extra multiplier to add to weight decreases to maintain overall weight balance'`
}

func (wb *WtBalRecvPrjn) Init() {
	wb.Avg = 0
	wb.Fact = 0
	wb.Inc = 1
	wb.Dec = 1
}

///////////////////////////////////////////////////////////////////////
//  leabra.Prjn

// leabra.Prjn is a basic Leabra projection with synaptic learning parameters
type Prjn struct {
	PrjnStru
	WtScale WtScaleParams  `desc:"weight scaling parameters: modulates overall strength of projection, using both absolute and relative factors"`
	Learn   LearnSynParams `desc:"synaptic-level learning parameters"`
	Syns    []Synapse      `desc:"synaptic state values, ordered by the sending layer units which "owns" them -- one-to-one with SConIdx array"`

	// misc state variables below:
	GeScale float32         `desc:"scaling factor for integrating excitatory synaptic input conductance Ge -- computed in TrialInit, incorporates running-average activity levels"`
	WbRecv  []WtBalRecvPrjn `desc:"weight balance state variables for this projection, one per recv neuron"`
}

func (pj *Prjn) Defaults() {
	pj.WtScale.Defaults()
	pj.Learn.Defaults()
	pj.GeScale = 1
}

// UpdateParams updates all params given any changes that might have been made to individual values
func (pj *Prjn) UpdateParams() {
	pj.WtScale.Update()
	pj.Learn.Update()
}

// SetParams sets given parameters to this prjn, if the target type is Prjn
// calls UpdateParams to ensure derived parameters are all updated
func (pj *Prjn) SetParams(pars emer.Params) bool {
	trg := pars.Target()
	if trg != "Prjn" {
		return false
	}
	pars.Set(pj)
	pj.UpdateParams()
	return true
}

// StyleParam applies a given style to this projection
// depending on the style specification (.Class, #Name, Type) and target value of params
func (pj *Prjn) StyleParam(sty string, pars emer.Params) bool {
	if emer.StyleMatch(sty, pj.PrjnName(), pj.Class, "Prjn") {
		return pj.SetParams(pars)
	}
	return false
}

// StyleParams applies a given styles to either this prjn
// depending on the style specification (.Class, #Name, Type) and target value of params
func (pj *Prjn) StyleParams(psty emer.ParamStyle) {
	for sty, pars := range psty {
		pj.StyleParam(sty, pars)
	}
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

// SynVal returns value of given variable name on the synapse between given recv unit index
// and send unit index -- returns error for access errors.
func (pj *Prjn) SynVal(varnm string, ridx, sidx int) (float32, error) {
	slay := pj.Send.(*Layer)
	rlay := pj.Recv.(*Layer)
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

///////////////////////////////////////////////////////////////////////
//  Build, Save Weights

// Build constructs the full connectivity among the layers as specified in this projection.
// Calls PrjnStru.BuildStru and then allocates the synaptic values in Syns accordingly.
func (pj *Prjn) Build() bool {
	if !pj.BuildStru() {
		return false
	}
	pj.Syns = make([]Synapse, len(pj.SConIdx))
	rsh := pj.Recv.LayShape()
	//	ssh := pj.Send.LayShape()
	rlen := rsh.Len()
	pj.WbRecv = make([]WtBalRecvPrjn, rlen)
	return true
}

// WriteWtsJSON writes the weights from this projection from the receiver-side perspective
// in a JSON text format.  We build in the indentation logic to make it much faster and
// more efficient.
func (pj *Prjn) WriteWtsJSON(w io.Writer, depth int) {
	slay := pj.Send.(*Layer)
	rlay := pj.Recv.(*Layer)
	nr := len(rlay.Neurons)
	w.Write(indent.TabBytes(depth))
	w.Write([]byte("{\n"))
	depth++
	w.Write(indent.TabBytes(depth))
	w.Write([]byte(fmt.Sprintf("\"GeScale\": %v\n", pj.GeScale)))
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
}

// InitWtSym initializes weight symmetry -- is given the reciprocal projection where
// the Send and Recv layers are reversed.
func (pj *Prjn) InitWtSym(rpj *Prjn) {
	slay := pj.Send.(*Layer)
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
					// note: if we support SymFmTop then can have option to go other way
					// also for Scale support, copy scales
				}
			}
		}
	}
}

//////////////////////////////////////////////////////////////////////////////////////
//  Act methods

// SendGeDelta sends the delta-activation from sending neuron index si,
// to integrate excitatory conductance on receivers
func (pj *Prjn) SendGeDelta(si int, delta float32) {
	scdel := delta * pj.GeScale
	nc := int(pj.SConN[si])
	st := int(pj.SConIdxSt[si])
	rlay := pj.Recv.(*Layer)
	for ci := 0; ci < nc; ci++ {
		sy := &pj.Syns[st+ci]
		ri := pj.SConIdx[st+ci]
		rn := &rlay.Neurons[ri]
		rn.GeInc += scdel * sy.Wt // todo: will need atomic for thread-safety
	}
}

//////////////////////////////////////////////////////////////////////////////////////
//  Learn methods

// DWt computes the weight change (learning) -- on sending projections
func (pj *Prjn) DWt() {
	if !pj.Learn.Learn {
		return
	}
	slay := pj.Send.(*Layer)
	rlay := pj.Recv.(*Layer)
	ns := len(slay.Neurons)
	for si := 0; si < ns; si++ {
		sn := &slay.Neurons[si]
		if sn.AvgS < pj.Learn.XCal.LrnThr && sn.AvgM < pj.Learn.XCal.LrnThr {
			continue
		}
		nc := int(pj.SConN[si])
		st := int(pj.SConIdxSt[si])
		for ci := 0; ci < nc; ci++ {
			sy := &pj.Syns[st+ci]
			ri := pj.SConIdx[st+ci]
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
			for ci := 0; ci < nc; ci++ {
				sy := &pj.Syns[st+ci]
				if sy.Norm > maxNorm {
					maxNorm = sy.Norm
				}
			}
			for ci := 0; ci < nc; ci++ {
				sy := &pj.Syns[st+ci]
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
	slay := pj.Send.(*Layer)
	ns := len(slay.Neurons)
	for si := 0; si < ns; si++ {
		nc := int(pj.SConN[si])
		st := int(pj.SConIdxSt[si])
		for ci := 0; ci < nc; ci++ {
			sy := &pj.Syns[st+ci]
			ri := pj.SConIdx[st+ci]
			wb := &pj.WbRecv[ri]
			pj.Learn.WtFmDWt(wb.Inc, wb.Dec, &sy.DWt, &sy.Wt, &sy.LWt)
		}
	}

	// todo: compare vs. putting WbInc / Dec into the synapses!
	// for si := range pj.Syns {
	// 	sy := &pj.Syns[si]
	// 	pj.Learn.WtFmDWt(sy.DWt, sy.WbInc, sy.WbDec, &sy.Wt, &sy.LWt)
	// }
}

// WtBalFmWt computes the Weight Balance factors based on average recv weights
func (pj *Prjn) WtBalFmWt() {
	if !pj.Learn.Learn || !pj.Learn.WtBal.On {
		return
	}

	rlay := pj.Recv.(*Layer)
	if rlay.Type == Target {
		return
	}
	nr := len(rlay.Neurons)
	for ri := 0; ri < nr; ri++ {
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
		sumWt := float32(0)
		sumN := 0
		for ci := 0; ci < nc; ci++ {
			rsi := pj.RSynIdx[st+ci]
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
