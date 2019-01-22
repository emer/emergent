// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import (
	"errors"
	"log"

	"github.com/apache/arrow/go/arrow/tensor"
	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/prjn"
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

func (ps *PrjnStru) PrjnClass() string     { return ps.Class }
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
			(sconN[si])++
		}
	}
	return true
}

// SetNIdxSt sets the *ConN and *ConIdxSt values given n tensor from Pat.
// Returns total number of connections for this direction.
func (ps *PrjnStru) SetNIdxSt(n *[]int32, avgmax *emer.AvgMax, idxst *[]int32, tn *tensor.Int32) int32 {
	ln := tn.Len()
	tnv := tn.Int32Values()
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
//  leabra.Prjn

// leabra.Prjn is a basic Leabra projection with synaptic learning parameters
type Prjn struct {
	PrjnStru
	WtScale WtScaleParams  `desc:"weight scaling parameters: modulates overall strength of projection, using both absolute and relative factors"`
	Learn   LearnSynParams `desc:"synaptic-level learning parameters"`
	Syns    []Synapse      `desc:"synaptic state values, ordered by the sending layer units which "owns" them -- one-to-one with SConIdx array"`
	GeScale float32        `desc:"scaling factor for integrating excitatory synaptic input conductance Ge -- computed in TrialInit, incorporates running-average activity levels"`
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

// Build constructs the full connectivity among the layers as specified in this projection.
// Calls PrjnStru.BuildStru and then allocates the synaptic values in Syns accordingly.
func (pj *Prjn) Build() bool {
	if !pj.BuildStru() {
		return false
	}
	pj.Syns = make([]Synapse, len(pj.SConIdx))
	return true
}

///////////////////////////////////////////////////////////////////////
//  leabra.PrjnList

// PrjnList is a slice of projections
type PrjnList []*Prjn

// Add adds a projection to the list
func (pl *PrjnList) Add(p *Prjn) {
	(*pl) = append(*pl, p)
}

// Build calls Build on all the prjns in the list
func (pl *PrjnList) Build() {
	for _, pj := range *pl {
		pj.Build()
	}
}

//////////////////////////////////////////////////////////////////////////////////////
//  Init methods

func (pj *Prjn) InitWts() {
	for si := range pj.Syns {
		sy := &pj.Syns[si]
		pj.Learn.InitWts(sy)
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
