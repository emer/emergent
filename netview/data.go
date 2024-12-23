// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import (
	"cogentcore.org/core/math32"
	"github.com/emer/emergent/v2/emer"
)

// LayData maintains a record of all the data for a given layer
type LayData struct {

	// the layer name
	LayName string

	// cached number of units
	NUnits int

	// the full data, in that order
	Data []float32

	// receiving pathway data -- shared with SendPaths
	RecvPaths []*PathData

	// sending pathway data -- shared with RecvPaths
	SendPaths []*PathData
}

// AllocSendPaths allocates Sending pathways for given layer.
// does nothing if already allocated.
func (ld *LayData) AllocSendPaths(ly emer.Layer) {
	nsp := ly.NumSendPaths()
	if len(ld.SendPaths) == nsp {
		for si := range ly.NumSendPaths() {
			pt := ly.SendPath(si)
			spd := ld.SendPaths[si]
			spd.Path = pt
		}
		return
	}
	ld.SendPaths = make([]*PathData, nsp)
	for si := range ly.NumSendPaths() {
		pt := ly.SendPath(si)
		pd := &PathData{Send: pt.SendLayer().Label(), Recv: pt.RecvLayer().Label(), Path: pt}
		ld.SendPaths[si] = pd
		pd.Alloc()
	}
}

// FreePaths nils path data -- for NoSynDat
func (ld *LayData) FreePaths() {
	ld.RecvPaths = nil
	ld.SendPaths = nil
}

// PathData holds display state for a pathway
type PathData struct {

	// name of sending layer
	Send string

	// name of recv layer
	Recv string

	// source pathway
	Path emer.Path

	// synaptic data, by variable in SynVars and number of data points
	SynData []float32
}

// Alloc allocates SynData to hold number of variables * nsyn synapses.
// If already has capacity, nothing happens.
func (pd *PathData) Alloc() {
	pt := pd.Path
	nvar := pt.SynVarNum()
	nsyn := pt.NumSyns()
	nt := nvar * nsyn
	if cap(pd.SynData) < nt {
		pd.SynData = make([]float32, nt)
	} else {
		pd.SynData = pd.SynData[:nt]
	}
}

// RecordData records synaptic data from given paths.
// must use sender or recv based depending on natural ordering.
func (pd *PathData) RecordData(nd *NetData) {
	pt := pd.Path
	vnms := pt.SynVarNames()
	nvar := pt.SynVarNum()
	nsyn := pt.NumSyns()
	for vi := 0; vi < nvar; vi++ {
		vnm := vnms[vi]
		si := vi * nsyn
		sv := pd.SynData[si : si+nsyn]
		pt.SynValues(&sv, vnm)
		nvi := nd.SynVarIndexes[vnm]
		mn := &nd.SynMinVar[nvi]
		mx := &nd.SynMaxVar[nvi]
		for _, vl := range sv {
			if !math32.IsNaN(vl) {
				*mn = math32.Min(*mn, vl)
				*mx = math32.Max(*mx, vl)
			}
		}
	}
}
