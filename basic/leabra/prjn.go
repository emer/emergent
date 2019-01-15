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
type PrjnStru struct {
	Off       bool       `desc:"inactivate this projection -- for testing"`
	Notes     string     `desc:"can record notes about this projection here"`
	Recv      emer.Layer `desc:"receiving layer for this projection -- the emer.Layer interface can be converted to the specific Layer type you are using, e.g., rlay := prjn.Recv.(*leabra.Layer)"`
	Send      emer.Layer `desc:"sending layer for this projection"`
	Pat       prjn.Pat   `desc:"pattern of connectivity"`
	RConN     []int32    `desc:"number of connections for each neuron in the receiving layer, as a flat list"`
	RConIdxSt []int32    `desc:"starting index into ConIdx list for each neuron in receiving layer -- just a list incremented by ConN"`
	RConIdx   []int32    `desc:"index of other neuron on sending side of projection, ordered by the receiving layer's order of units as the outer loop (each start is in ConIdxSt), and then by the sending layer's units within that"`
	RSynIdx   []int32    `desc:"index of synaptic state values for each recv unit x connection, for the receiver projection which does not own the synapses, and instead indexes into sender-ordered list"`
	SConN     []int32    `desc:"number of connections for each neuron in the sending layer, as a flat list"`
	SConIdxSt []int32    `desc:"starting index into ConIdx list for each neuron in sending layer -- just a list incremented by ConN"`
	SConIdx   []int32    `desc:"index of other neuron on receiving side of projection, ordered by the sending layer's order of units as the outer loop (each start is in ConIdxSt), and then by the sending layer's units within that"`
}

// Connect sets the connectivity between two layers and the pattern to use in interconnecting them
func (ps *PrjnStru) Connect(rlay, slay emer.Layer, pat prjn.Pat) {
	ps.Recv = rlay
	ps.Send = slay
	ps.Pat = pat
}

// Validate tests for non-nil settings for the projection -- returns error
// message or nil if no problems (and logs them if log = true)
func (ps *PrjnStru) Validate(log bool) error {
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
		if log {
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
	rsh := ps.Recv.Shape()
	ssh := ps.Send.Shape()
	recvn, sendn, cons := ps.Pat.Connect(rsh, ssh, ps.Recv == ps.Send)
	rlen := rsh.Len()
	slen := ssh.Len()
	tconr := ps.SetNIdxSt(&ps.RConN, &ps.RConIdxSt, recvn)
	tcons := ps.SetNIdxSt(&ps.SConN, &ps.SConIdxSt, sendn)
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
		rci := 0
		for si := 0; si < slen; si++ {
			if !cbits.Index(rbi + si) { // no connection
				continue
			}
			sst := ps.SConIdxSt[si]
			if rci >= rtcn {
				log.Printf("%v programmer error: recv target total con number: %v exceeded at recv idx: %v, send idx: %v\n", ps.String(), rtcn, ri, si)
				break
			}
			ps.RConIdx[rst+rci] = si

			sci := sconN[si]
			stcn := ps.SConN[si]
			if sci >= stcn {
				log.Printf("%v programmer error: send target total con number: %v exceeded at recv idx: %v, send idx: %v\n", ps.String(), stcn, ri, si)
				break
			}
			ps.SConIdx[sst+sci] = ri
			(sconN[si])++
		}
	}
}

// SetNIdxSt sets the *ConN and *ConIdxSt values given n tensor from Pat.
// Returns total number of connections for this direction.
func (ps *PrjnStru) SetNIdxSt(n, idxst *[]int32, tn *tensor.Int32) int32 {
	ln := tn.Len()
	tnv := tn.Int32Values()
	*n = make([]int32, ln)
	*idxst = make([]int32, ln)
	idx := int32(0)
	for i := 0; i < ln; i++ {
		nv := tnv[i]
		(*n)[i] = nv
		(*idxst)[i] = idx
		idx += nv
	}
	return idx
}

// String satisfies fmt.Stringer for prjn
func (ps *PrjnStru) String() string {
	str := ""
	if ps.Recv == nil {
		str += "recv=nil; "
	} else {
		str += ps.Recv.Name() + " <- "
	}
	if ps.Send == nil {
		str += "send=nil"
	} else {
		str += ps.Send.Name()
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
	Learn LearnSyn  `desc:"synaptic-level learning parameters"`
	Syns  []Synapse `desc:"synaptic state values, ordered by the sending layer units which "owns" them -- one-to-one with SConIdx array"`
}

// Build constructs the full connectivity among the layers as specified in this projection.
// Calls PrjnStru.BuildStru and then allocates the synaptic values in Syns accordingly.
func (pj *Prjn) Build() bool {
	if !pj.BuildStru() {
		return false
	}
	pj.Syns = make([]Synapse, len(pj.SConIdx))
}

// PrjnList is a slice of projections
type PrjnList []*Prjn

// Add adds a projection to the list
func (pl *PrjnList) Add(p *Prjn) {
	(*pl) = append(*pl, p)
}
