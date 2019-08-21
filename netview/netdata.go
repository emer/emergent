// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import (
	"math"
	"strings"

	"github.com/chewxy/math32"
	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/ringidx"
)

// LayData maintains a record of all the data for a given layer
type LayData struct {
	LayName string    `desc:"the layer name"`
	Data    []float32 `desc:"the full data, Ring.Max * len(Vars) * NUnits in that order"`
}

// NetData maintains a record of all the network data to that has been displayed
// up to a given maximum number of records (updates), using efficient ring index logic
// with no copying to store in fixed-sized buffers.
type NetData struct {
	Net       emer.Network        `desc:"the network that we're viewing"`
	PrjnLay   string              `desc:"name of the layer with unit for viewing projections (connection / synapse-level values)"`
	PrjnUnIdx int                 `desc:"1D index of unit within PrjnLay for for viewing projections"`
	Vars      []string            `desc:"the list of variables saved -- copied from NetView"`
	VarIdxs   map[string]int      `desc:"index of each variable in the Vars slice"`
	Ring      ringidx.Idx         `desc:"the circular ring index -- Max here is max number of values to store, Len is number stored, and Idx(Len-1) is the most recent one, etc"`
	LayData   map[string]*LayData `desc:"the layer data -- map keyed by layer name"`
	MinPer    []float32           `desc:"min values for each Ring.Max * variable"`
	MaxPer    []float32           `desc:"max values for each Ring.Max * variable"`
	MinVar    []float32           `desc:"min values for variable"`
	MaxVar    []float32           `desc:"max values for variable"`
}

// Init initializes the main params and configures the data
func (nd *NetData) Init(net emer.Network, max int) {
	nd.Net = net
	nd.Ring.Max = max
	nd.Config()
}

// Config configures the data storage for given network
// only re-allocates if needed.
func (nd *NetData) Config() {
	nlay := nd.Net.NLayers()
	if nlay == 0 {
		return
	}
	if nd.Ring.Max == 0 {
		nd.Ring.Max = 2
	}
	rmax := nd.Ring.Max
	if rmax > nd.Ring.Len {
		nd.Ring.Reset()
	}
	nvars := NetVarsList(nd.Net, false) // not even
	vlen := len(nvars)
	if len(nd.Vars) != vlen {
		nd.Vars = nvars
	}
makeData:
	if len(nd.LayData) != nlay {
		nd.LayData = make(map[string]*LayData, nlay)
		for li := 0; li < nlay; li++ {
			lay := nd.Net.Layer(li)
			nm := lay.Name()
			nd.LayData[nm] = &LayData{LayName: nm}
		}
	}
	vmax := vlen * rmax
	for li := 0; li < nlay; li++ {
		lay := nd.Net.Layer(li)
		nm := lay.Name()
		ld, ok := nd.LayData[nm]
		if !ok {
			nd.LayData = nil
			goto makeData
		}
		nu := lay.Shape().Len()
		ltot := vmax * nu
		if len(ld.Data) != ltot {
			ld.Data = make([]float32, ltot)
		}
	}
	if len(nd.MinPer) != vmax {
		nd.MinPer = make([]float32, vmax)
		nd.MaxPer = make([]float32, vmax)
	}
	if len(nd.MinVar) != vlen {
		nd.MinVar = make([]float32, vlen)
		nd.MaxVar = make([]float32, vlen)
	}
}

// RecvPrjnValFrom returns the receiving projection value to cur prjn unit
// from given 1D unit index in given layer.
// Returns false if no connection or no valid prjn unit.
func (nd *NetData) RecvPrjnValFrom(svar string, lay emer.Layer, idx1d int) (float32, bool) {
	play := nd.Net.LayerByName(nd.PrjnLay)
	if play == nil {
		return 0, false
	}
	pp := play.RecvPrjns().SendName(lay.Name())
	if pp == nil {
		return 0, false
	}
	sval, err := pp.SynValTry(svar, idx1d, nd.PrjnUnIdx)
	if err != nil {
		return 0, false
	}
	return sval, true
}

// SendPrjnValTo returns the sending projection value from cur prjn unit
// to given 1D unit index in given layer.
// Returns false if no connection or no valid prjn unit.
func (nd *NetData) SendPrjnValFrom(svar string, lay emer.Layer, idx1d int) (float32, bool) {
	play := nd.Net.LayerByName(nd.PrjnLay)
	if play == nil {
		return 0, false
	}
	pp := play.SendPrjns().RecvName(lay.Name())
	if pp == nil {
		return 0, false
	}
	sval, err := pp.SynValTry(svar, nd.PrjnUnIdx, idx1d)
	if err != nil {
		return 0, false
	}
	return sval, true
}

// Record records the current full set of data from the network
func (nd *NetData) Record() {
	nlay := nd.Net.NLayers()
	if nlay == 0 {
		return
	}
	nd.Config() // inexpensive if no diff, and safe..
	vlen := len(nd.Vars)
	nd.Ring.Add(1)
	lidx := nd.Ring.LastIdx()
	mmidx := lidx * vlen
	for vi := range nd.Vars {
		nd.MinPer[mmidx+vi] = math.MaxFloat32
		nd.MaxPer[mmidx+vi] = -math.MaxFloat32
	}
	for li := 0; li < nlay; li++ {
		lay := nd.Net.Layer(li)
		laynm := lay.Name()
		ld := nd.LayData[laynm]
		nu := lay.Shape().Len()
		nvu := vlen * nu
		for vi, vnm := range nd.Vars {
			mn := &nd.MinPer[mmidx+vi]
			mx := &nd.MaxPer[mmidx+vi]
			for ui := 0; ui < nu; ui++ {
				idx := lidx*vi*nvu + vi*nu + ui
				hasval := true
				raw := float32(0)
				if strings.HasPrefix(vnm, "r.") {
					svar := vnm[2:]
					raw, hasval = nd.RecvPrjnValFrom(svar, lay, ui)
				} else if strings.HasPrefix(vnm, "s.") {
					svar := vnm[2:]
					raw, hasval = nd.SendPrjnValFrom(svar, lay, ui)
				} else {
					raw = lay.UnitVal1D(vnm, ui)
				}
				if hasval {
					ld.Data[idx] = raw
					*mn = math32.Min(*mn, raw)
					*mx = math32.Max(*mx, raw)
				} else {
					ld.Data[idx] = math.MaxFloat32 // hack flag
				}
			}
		}
	}
}
