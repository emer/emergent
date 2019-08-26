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
	NUnits  int       `desc:"cached number of units"`
	Data    []float32 `desc:"the full data, Ring.Max * len(Vars) * NUnits in that order"`
}

// NetData maintains a record of all the network data that has been displayed
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
	Counters  []string            `desc:"counter strings"`
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
	if nd.Ring.Len > rmax {
		nd.Ring.Reset()
	}
	nvars := NetVarsList(nd.Net, false) // not even
	vlen := len(nvars)
	if len(nd.Vars) != vlen {
		nd.Vars = nvars
		nd.VarIdxs = make(map[string]int, vlen)
		for vi, vn := range nd.Vars {
			nd.VarIdxs[vn] = vi
		}
	}
makeData:
	if len(nd.LayData) != nlay {
		nd.LayData = make(map[string]*LayData, nlay)
		for li := 0; li < nlay; li++ {
			lay := nd.Net.Layer(li)
			nm := lay.Name()
			nd.LayData[nm] = &LayData{LayName: nm, NUnits: lay.Shape().Len()}
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
		nu := ld.NUnits
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
	if len(nd.Counters) != rmax {
		nd.Counters = make([]string, rmax)
	}
}

// Record records the current full set of data from the network, and the given counters string
func (nd *NetData) Record(ctrs string) {
	nlay := nd.Net.NLayers()
	if nlay == 0 {
		return
	}
	nd.Config() // inexpensive if no diff, and safe..
	vlen := len(nd.Vars)
	nd.Ring.Add(1)
	lidx := nd.Ring.LastIdx()

	nd.Counters[lidx] = ctrs

	prjnlay := nd.Net.LayerByName(nd.PrjnLay)

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
			idx := lidx*nvu + vi*nu
			dvals := ld.Data[idx : idx+nu]
			if strings.HasPrefix(vnm, "r.") {
				svar := vnm[2:]
				lay.SendPrjnVals(&dvals, svar, prjnlay, nd.PrjnUnIdx)
			} else if strings.HasPrefix(vnm, "s.") {
				svar := vnm[2:]
				lay.RecvPrjnVals(&dvals, svar, prjnlay, nd.PrjnUnIdx)
			} else {
				lay.UnitVals(&dvals, vnm)
			}
			for ui := range dvals {
				vl := dvals[ui]
				if !math32.IsNaN(vl) {
					*mn = math32.Min(*mn, vl)
					*mx = math32.Max(*mx, vl)
				}
			}
		}
	}
	nd.UpdateVarRange()
}

// UpdateVarRange updates the range for variables
func (nd *NetData) UpdateVarRange() {
	vlen := len(nd.Vars)
	rlen := nd.Ring.Len
	for vi := range nd.Vars {
		vmn := &nd.MinVar[vi]
		vmx := &nd.MaxVar[vi]
		*vmn = math.MaxFloat32
		*vmx = -math.MaxFloat32

		for ri := 0; ri < rlen; ri++ {
			mmidx := ri * vlen
			mn := nd.MinPer[mmidx+vi]
			mx := nd.MaxPer[mmidx+vi]
			*vmn = math32.Min(*vmn, mn)
			*vmx = math32.Max(*vmx, mx)
		}
	}
}

// VarRange returns the current min, max range for given variable.
// Returns false if not found or no data.
func (nd *NetData) VarRange(vnm string) (float32, float32, bool) {
	if nd.Ring.Len == 0 {
		return 0, 0, false
	}
	vi, ok := nd.VarIdxs[vnm]
	if !ok {
		return 0, 0, false
	}
	return nd.MinVar[vi], nd.MaxVar[vi], true
}

// RecIdx returns record index for given record number,
// which is -1 for current (last) record, or in [0..Len-1] for prior records.
func (nd *NetData) RecIdx(recno int) int {
	ridx := nd.Ring.LastIdx()
	if nd.Ring.IdxIsValid(recno) {
		ridx = nd.Ring.Idx(recno)
	}
	return ridx
}

// CounterRec returns counter string for given record,
// which is -1 for current (last) record, or in [0..Len-1] for prior records.
func (nd *NetData) CounterRec(recno int) string {
	if nd.Ring.Len == 0 {
		return ""
	}
	ridx := nd.RecIdx(recno)
	return nd.Counters[ridx]
}

// UnitVal returns the value for given layer, variable name, unit index, and record number,
// which is -1 for current (last) record, or in [0..Len-1] for prior records.
// Returns false if value unavailable for any reason (including recorded as such as NaN).
func (nd *NetData) UnitVal(laynm string, vnm string, uidx1d int, recno int) (float32, bool) {
	if nd.Ring.Len == 0 {
		return 0, false
	}
	vi, ok := nd.VarIdxs[vnm]
	if !ok {
		return 0, false
	}
	vlen := len(nd.Vars)
	ridx := nd.RecIdx(recno)
	ld, ok := nd.LayData[laynm]
	if !ok {
		return 0, false
	}
	nu := ld.NUnits
	nvu := vlen * nu
	idx := ridx*nvu + vi*nu + uidx1d
	val := ld.Data[idx]
	if math32.IsNaN(val) {
		return 0, false
	}
	return val, true
}
