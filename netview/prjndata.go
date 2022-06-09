// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import (
	"github.com/emer/emergent/emer"
	"github.com/goki/mat32"
)

// LayData maintains a record of all the data for a given layer
type LayData struct {
	LayName   string      `desc:"the layer name"`
	NUnits    int         `desc:"cached number of units"`
	Data      []float32   `desc:"the full data, Ring.Max * len(Vars) * NUnits in that order"`
	RecvPrjns []*PrjnData `desc:"receiving projection data -- shared with SendPrjns"`
	SendPrjns []*PrjnData `desc:"sending projection data -- shared with RecvPrjns"`
}

// AllocSendPrjns allocates Sending projections for given layer.
// does nothing if already allocated.
func (ld *LayData) AllocSendPrjns(ly emer.Layer) {
	// nrp := ly.NRecvPrjns()
	// len(ld.RecvPrjns) == nrp &&
	nsp := ly.NSendPrjns()
	if len(ld.SendPrjns) == nsp {
		return
	}
	ld.SendPrjns = make([]*PrjnData, nsp)
	sp := ly.SendPrjns()
	for si, pj := range *sp {
		pd := &PrjnData{Send: pj.SendLay().Name(), Recv: pj.RecvLay().Name(), Prjn: pj}
		ld.SendPrjns[si] = pd
		// todo: need a Recv option
		pd.Alloc()
	}
}

// FreePrjns nils prjn data -- for NoSynDat
func (ld *LayData) FreePrjns() {
	ld.RecvPrjns = nil
	ld.SendPrjns = nil
}

// PrjnData holds display state for a projection
type PrjnData struct {
	Send    string    `desc:"name of sending layer"`
	Recv    string    `desc:"name of recv layer"`
	Prjn    emer.Prjn `desc:"source projection"`
	SynData []float32 `desc:"synaptic data, by variable in SynVars and number of data points"`
}

// Alloc allocates SynData to hold number of variables * nsyn synapses.
// If already has capacity, nothing happens.
func (pd *PrjnData) Alloc() {
	pj := pd.Prjn
	nvar := pj.SynVarNum()
	nsyn := pj.Syn1DNum()
	nt := nvar * nsyn
	if cap(pd.SynData) < nt {
		pd.SynData = make([]float32, nt)
	} else {
		pd.SynData = pd.SynData[:nt]
	}
}

// RecordData records synaptic data from given prjn.
// must use sender or recv based depending on natural ordering.
func (pd *PrjnData) RecordData(nd *NetData) {
	pj := pd.Prjn
	vnms := pj.SynVarNames()
	nvar := pj.SynVarNum()
	nsyn := pj.Syn1DNum()
	for vi := 0; vi < nvar; vi++ {
		vnm := vnms[vi]
		si := vi * nsyn
		sv := pd.SynData[si : si+nsyn]
		pj.SynVals(&sv, vnm)
		nvi := nd.SynVarIdxs[vnm]
		mn := &nd.SynMinVar[nvi]
		mx := &nd.SynMaxVar[nvi]
		for _, vl := range sv {
			if !mat32.IsNaN(vl) {
				*mn = mat32.Min(*mn, vl)
				*mx = mat32.Max(*mx, vl)
			}
		}
	}
}
