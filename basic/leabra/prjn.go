// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import "github.com/emer/emergent/prjn"

// PrjnStru contains the basic structural information for specifying a projection of synaptic
// connections between two layers
type PrjnStru struct {
	Recv      *Layer   `desc:"receiving layer for this projection"`
	Send      *Layer   `desc:"sending layer for this projection"`
	Pat       prjn.Pat `desc:"pattern of connectivity"`
	RConN     []int32  `desc:"number of connections for each neuron in the receiving layer, as a flat list"`
	RConIdxSt []int32  `desc:"starting index into ConIdx list for each neuron in receiving layer -- just a list incremented by ConN"`
	RConIdx   []int32  `desc:"index of other neuron on sending side of projection, ordered by the receiving layer's order of units as the outer loop (each start is in ConIdxSt), and then by the sending layer's units within that"`
	RSynIdx   []int32  `desc:"index of synaptic state values, for the receiver projection which does not own the synapses, and instead indexes into sender-ordered list"`
	SConN     []int32  `desc:"number of connections for each neuron in the sending layer, as a flat list"`
	SConIdxSt []int32  `desc:"starting index into ConIdx list for each neuron in sending layer -- just a list incremented by ConN"`
	SConIdx   []int32  `desc:"index of other neuron on receiving side of projection, ordered by the sending layer's order of units as the outer loop (each start is in ConIdxSt), and then by the sending layer's units within that"`
}

// leabra.Prjn is a basic Leabra projection with synaptic learning parameters
type Prjn struct {
	PrjnStru
	Learn LearnSyn  `desc:"synaptic-level learning parameters"`
	Syns  []Synapse `desc:"synaptic state values, ordered by the sending layer units which "owns" them -- one-to-one with SConIdx array"`
}

// PrjnList is a slice of projections
type PrjnList []*Prjn

// Add adds a projection to the list
func (pl *PrjnList) Add(p *Prjn) {
	(*pl) = append(*pl, p)
}
