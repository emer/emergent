// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/metadata"
	"cogentcore.org/core/core"
	"cogentcore.org/core/math32"
	"cogentcore.org/lab/lab"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/plotcore"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensorfs"
	"github.com/emer/emergent/v2/emer"
	"github.com/emer/emergent/v2/ringidx"
)

// NetData maintains a record of all the network data that has been displayed
// up to a given maximum number of records (updates), using efficient ring index logic
// with no copying to store in fixed-sized buffers.
type NetData struct { //types:add

	// the network that we're viewing
	Net emer.Network `json:"-"`

	// copied from Params -- do not record synapse level data -- turn this on for very large networks where recording the entire synaptic state would be prohibitive
	NoSynData bool

	// name of the layer with unit for viewing pathways (connection / synapse-level values)
	PathLay string

	// 1D index of unit within PathLay for for viewing pathways
	PathUnIndex int

	// copied from NetView Params: if non-empty, this is the type pathway to show when there are multiple pathways from the same layer -- e.g., Inhib, Lateral, Forward, etc
	PathType string `edit:"-"`

	// the list of unit variables saved
	UnVars []string

	// index of each variable in the Vars slice
	UnVarIndexes map[string]int

	// the list of synaptic variables saved
	SynVars []string

	// index of synaptic variable in the SynVars slice
	SynVarIndexes map[string]int

	// the circular ring index -- Max here is max number of values to store, Len is number stored, and Index(Len-1) is the most recent one, etc
	Ring ringidx.Index

	// max data parallel data per unit
	MaxData int

	// the layer data -- map keyed by layer name
	LayData map[string]*LayData

	// unit var min values for each Ring.Max * variable
	UnMinPer []float32

	// unit var max values for each Ring.Max * variable
	UnMaxPer []float32

	// min values for unit variables
	UnMinVar []float32

	// max values for unit variables
	UnMaxVar []float32

	// min values for syn variables
	SynMinVar []float32

	// max values for syn variables
	SynMaxVar []float32

	// counter strings
	Counters []string

	// raster counter values
	RasterCtrs []int

	// map of raster counter values to record numbers
	RasterMap map[int]int

	// dummy raster counter when passed a -1 -- increments and wraps around
	RastCtr int
}

// Init initializes the main params and configures the data
func (nd *NetData) Init(net emer.Network, max int, noSynData bool, maxData int) {
	nd.Net = net
	nd.Ring.Max = max
	nd.MaxData = maxData
	nd.NoSynData = noSynData
	nd.Config()
	nd.RastCtr = 0
	nd.RasterMap = make(map[int]int)
}

// Config configures the data storage for given network
// only re-allocates if needed.
func (nd *NetData) Config() {
	nlay := nd.Net.NumLayers()
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
	nvars := nd.Net.UnitVarNames()
	vlen := len(nvars)
	if len(nd.UnVars) != vlen {
		nd.UnVars = nvars
		nd.UnVarIndexes = make(map[string]int, vlen)
		for vi, vn := range nd.UnVars {
			nd.UnVarIndexes[vn] = vi
		}
	}
	svars := nd.Net.SynVarNames()
	svlen := len(svars)
	if len(nd.SynVars) != svlen {
		nd.SynVars = svars
		nd.SynVarIndexes = make(map[string]int, svlen)
		for vi, vn := range nd.SynVars {
			nd.SynVarIndexes[vn] = vi
		}
	}
makeData:
	if len(nd.LayData) != nlay {
		nd.LayData = make(map[string]*LayData, nlay)
		for li := range nlay {
			lay := nd.Net.EmerLayer(li).AsEmer()
			nm := lay.Name
			ld := &LayData{LayName: nm, NUnits: lay.Shape.Len()}
			nd.LayData[nm] = ld
			if nd.NoSynData {
				ld.FreePaths()
			} else {
				ld.AllocSendPaths(lay.EmerLayer)
			}
		}
		if !nd.NoSynData {
			for li := range nlay {
				rlay := nd.Net.EmerLayer(li)
				rld := nd.LayData[rlay.Label()]
				rld.RecvPaths = make([]*PathData, rlay.NumRecvPaths())
				for ri := 0; ri < rlay.NumRecvPaths(); ri++ {
					rpj := rlay.RecvPath(ri)
					slay := rpj.SendLayer()
					sld := nd.LayData[slay.Label()]
					for _, spj := range sld.SendPaths {
						if spj.Path == rpj {
							rld.RecvPaths[ri] = spj // link
						}
					}
				}
			}
		}
	} else {
		for li := range nlay {
			lay := nd.Net.EmerLayer(li)
			ld := nd.LayData[lay.Label()]
			if nd.NoSynData {
				ld.FreePaths()
			} else {
				ld.AllocSendPaths(lay)
			}
		}
	}
	vmax := vlen * rmax * nd.MaxData
	for li := range nlay {
		lay := nd.Net.EmerLayer(li).AsEmer()
		nm := lay.Name
		ld, ok := nd.LayData[nm]
		if !ok {
			nd.LayData = nil
			goto makeData
		}
		ld.NUnits = lay.Shape.Len()
		nu := ld.NUnits
		ltot := vmax * nu
		if len(ld.Data) != ltot {
			ld.Data = make([]float32, ltot)
		}
	}
	if len(nd.UnMinPer) != vmax {
		nd.UnMinPer = make([]float32, vmax)
		nd.UnMaxPer = make([]float32, vmax)
	}
	if len(nd.UnMinVar) != vlen {
		nd.UnMinVar = make([]float32, vlen)
		nd.UnMaxVar = make([]float32, vlen)
	}
	if len(nd.SynMinVar) != svlen {
		nd.SynMinVar = make([]float32, svlen)
		nd.SynMaxVar = make([]float32, svlen)
	}
	if len(nd.Counters) != rmax {
		nd.Counters = make([]string, rmax)
		nd.RasterCtrs = make([]int, rmax)
	}
}

// Record records the current full set of data from the network,
// and the given counters string (displayed at bottom of window)
// and raster counter value -- if negative, then an internal
// wraping-around counter is used.
func (nd *NetData) Record(ctrs string, rastCtr, rastMax int) {
	nlay := nd.Net.NumLayers()
	if nlay == 0 {
		return
	}
	nd.Config() // inexpensive if no diff, and safe..
	vlen := len(nd.UnVars)
	nd.Ring.Add(1)
	lidx := nd.Ring.LastIndex()
	maxData := nd.MaxData

	if rastCtr < 0 {
		rastCtr = nd.RastCtr
		nd.RastCtr++
		if nd.RastCtr >= rastMax {
			nd.RastCtr = 0
		}
	}

	nd.Counters[lidx] = ctrs
	nd.RasterCtrs[lidx] = rastCtr
	nd.RasterMap[rastCtr] = lidx

	mmidx := lidx * vlen
	for vi := range nd.UnVars {
		nd.UnMinPer[mmidx+vi] = math.MaxFloat32
		nd.UnMaxPer[mmidx+vi] = -math.MaxFloat32
	}
	for li := range nlay {
		lay := nd.Net.EmerLayer(li).AsEmer()
		laynm := lay.Name
		ld := nd.LayData[laynm]
		nu := lay.Shape.Len()
		nvu := vlen * maxData * nu
		for vi, vnm := range nd.UnVars {
			mn := &nd.UnMinPer[mmidx+vi]
			mx := &nd.UnMaxPer[mmidx+vi]
			for di := 0; di < maxData; di++ {
				idx := lidx*nvu + vi*maxData*nu + di*nu
				dvals := ld.Data[idx : idx+nu]
				lay.UnitValues(&dvals, vnm, di)
				for ui := range dvals {
					vl := dvals[ui]
					if !math32.IsNaN(vl) {
						*mn = math32.Min(*mn, vl)
						*mx = math32.Max(*mx, vl)
					}
				}
			}
		}
	}
	nd.UpdateUnVarRange()
}

// RecordLastCtrs records just the last counter string to be the given string
// overwriting what was there before.
func (nd *NetData) RecordLastCtrs(ctrs string) {
	lidx := nd.Ring.LastIndex()
	nd.Counters[lidx] = ctrs
}

// UpdateUnVarRange updates the range for unit variables, integrating over
// the entire range of stored values, so it is valid when iterating
// over history.
func (nd *NetData) UpdateUnVarRange() {
	vlen := len(nd.UnVars)
	rlen := nd.Ring.Len
	for vi := range nd.UnVars {
		vmn := &nd.UnMinVar[vi]
		vmx := &nd.UnMaxVar[vi]
		*vmn = math.MaxFloat32
		*vmx = -math.MaxFloat32

		for ri := 0; ri < rlen; ri++ {
			ridx := nd.Ring.Index(ri)
			mmidx := ridx * vlen
			mn := nd.UnMinPer[mmidx+vi]
			mx := nd.UnMaxPer[mmidx+vi]
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
	if strings.HasPrefix(vnm, "r.") || strings.HasPrefix(vnm, "s.") {
		vnm = vnm[2:]
		vi, ok := nd.SynVarIndexes[vnm]
		if !ok {
			return 0, 0, false
		}
		return nd.SynMinVar[vi], nd.SynMaxVar[vi], true
	}
	vi, ok := nd.UnVarIndexes[vnm]
	if !ok {
		return 0, 0, false
	}
	return nd.UnMinVar[vi], nd.UnMaxVar[vi], true
}

// RecordSyns records synaptic data -- stored separate from unit data
// and only needs to be called when synaptic values are updated.
// Should be done when the DWt values have been computed, before
// updating Wts and zeroing.
// NetView displays this recorded data when Update is next called.
func (nd *NetData) RecordSyns() {
	if nd.NoSynData {
		return
	}
	nlay := nd.Net.NumLayers()
	if nlay == 0 {
		return
	}
	nd.Config() // inexpensive if no diff, and safe..
	for vi := range nd.SynVars {
		nd.SynMinVar[vi] = math.MaxFloat32
		nd.SynMaxVar[vi] = -math.MaxFloat32
	}
	for li := range nlay {
		lay := nd.Net.EmerLayer(li)
		laynm := lay.Label()
		ld := nd.LayData[laynm]
		for si := 0; si < lay.NumSendPaths(); si++ {
			spd := ld.SendPaths[si]
			spd.RecordData(nd)
		}
	}
}

// RecIndex returns record index for given record number,
// which is -1 for current (last) record, or in [0..Len-1] for prior records.
func (nd *NetData) RecIndex(recno int) int {
	ridx := nd.Ring.LastIndex()
	if nd.Ring.IndexIsValid(recno) {
		ridx = nd.Ring.Index(recno)
	}
	return ridx
}

// CounterRec returns counter string for given record,
// which is -1 for current (last) record, or in [0..Len-1] for prior records.
func (nd *NetData) CounterRec(recno int) string {
	if nd.Ring.Len == 0 {
		return ""
	}
	ridx := nd.RecIndex(recno)
	return nd.Counters[ridx]
}

// UnitVal returns the value for given layer, variable name, unit index, data parallel idx di,
// and record number, which is -1 for current (last) record, or in [0..Len-1] for prior records.
// Returns false if value unavailable for any reason (including recorded as such as NaN).
func (nd *NetData) UnitValue(laynm string, vnm string, uidx1d int, recno int, di int) (float32, bool) {
	if nd.Ring.Len == 0 {
		return 0, false
	}
	ridx := nd.RecIndex(recno)
	return nd.UnitValueIndex(laynm, vnm, uidx1d, ridx, di)
}

// RasterCtr returns the raster counter value at given record number (-1 = current)
func (nd *NetData) RasterCtr(recno int) (int, bool) {
	if nd.Ring.Len == 0 {
		return 0, false
	}
	ridx := nd.RecIndex(recno)
	return nd.RasterCtrs[ridx], true
}

// UnitValRaster returns the value for given layer, variable name, unit index, and
// raster counter number.
// Returns false if value unavailable for any reason (including recorded as such as NaN).
func (nd *NetData) UnitValRaster(laynm string, vnm string, uidx1d int, rastCtr int, di int) (float32, bool) {
	ridx, has := nd.RasterMap[rastCtr]
	if !has {
		return 0, false
	}
	return nd.UnitValueIndex(laynm, vnm, uidx1d, ridx, di)
}

// UnitValueIndex returns the value for given layer, variable name, unit index, stored idx,
// and data parallel index.
// Returns false if value unavailable for any reason (including recorded as such as NaN).
func (nd *NetData) UnitValueIndex(laynm string, vnm string, uidx1d int, ridx int, di int) (float32, bool) {
	if strings.HasPrefix(vnm, "r.") {
		svar := vnm[2:]
		return nd.RecvUnitValue(laynm, svar, uidx1d)
	} else if strings.HasPrefix(vnm, "s.") {
		svar := vnm[2:]
		return nd.SendUnitValue(laynm, svar, uidx1d)
	}
	vi, ok := nd.UnVarIndexes[vnm]
	if !ok {
		return 0, false
	}
	vlen := len(nd.UnVars)
	ld, ok := nd.LayData[laynm]
	if !ok {
		return 0, false
	}
	nu := ld.NUnits
	nvu := vlen * nd.MaxData * nu
	idx := ridx*nvu + vi*nd.MaxData*nu + di*nu + uidx1d
	val := ld.Data[idx]
	if math32.IsNaN(val) {
		return 0, false
	}
	return val, true
}

// RecvUnitVal returns the value for given layer, variable name, unit index,
// for receiving pathway variable, based on recorded synaptic pathway data.
// Returns false if value unavailable for any reason (including recorded as such as NaN).
func (nd *NetData) RecvUnitValue(laynm string, vnm string, uidx1d int) (float32, bool) {
	ld, ok := nd.LayData[laynm]
	if nd.NoSynData || !ok || nd.PathLay == "" {
		return 0, false
	}
	recvLay := errors.Ignore1(nd.Net.AsEmer().EmerLayerByName(nd.PathLay)).AsEmer()
	if recvLay == nil {
		return 0, false
	}
	var pj emer.Path
	var err error
	if nd.PathType != "" {
		pj, err = recvLay.RecvPathBySendNameType(laynm, nd.PathType)
		if pj == nil {
			pj, err = recvLay.RecvPathBySendName(laynm)
		}
	} else {
		pj, err = recvLay.RecvPathBySendName(laynm)
	}
	if pj == nil {
		return 0, false
	}
	var spd *PathData
	for _, pd := range ld.SendPaths {
		if pd.Path == pj {
			spd = pd
			break
		}
	}
	if spd == nil {
		return 0, false
	}
	varIndex, err := pj.SynVarIndex(vnm)
	if err != nil {
		return 0, false
	}
	synIndex := pj.SynIndex(uidx1d, nd.PathUnIndex)
	if synIndex < 0 {
		return 0, false
	}
	nsyn := pj.NumSyns()
	val := spd.SynData[varIndex*nsyn+synIndex]
	return val, true
}

// SendUnitVal returns the value for given layer, variable name, unit index,
// for sending pathway variable, based on recorded synaptic pathway data.
// Returns false if value unavailable for any reason (including recorded as such as NaN).
func (nd *NetData) SendUnitValue(laynm string, vnm string, uidx1d int) (float32, bool) {
	ld, ok := nd.LayData[laynm]
	if nd.NoSynData || !ok || nd.PathLay == "" {
		return 0, false
	}
	sendLay := errors.Ignore1(nd.Net.AsEmer().EmerLayerByName(nd.PathLay)).AsEmer()
	if sendLay == nil {
		return 0, false
	}
	var pj emer.Path
	var err error
	if nd.PathType != "" {
		pj, err = sendLay.SendPathByRecvNameType(laynm, nd.PathType)
		if pj == nil {
			pj, err = sendLay.SendPathByRecvName(laynm)
		}
	} else {
		pj, err = sendLay.SendPathByRecvName(laynm)
	}
	if pj == nil {
		return 0, false
	}
	var rpd *PathData
	for _, pd := range ld.RecvPaths {
		if pd.Path == pj {
			rpd = pd
			break
		}
	}
	if rpd == nil {
		return 0, false
	}
	varIndex, err := pj.SynVarIndex(vnm)
	if err != nil {
		return 0, false
	}
	synIndex := pj.SynIndex(nd.PathUnIndex, uidx1d)
	if synIndex < 0 {
		return 0, false
	}
	nsyn := pj.NumSyns()
	val := rpd.SynData[varIndex*nsyn+synIndex]
	return val, true
}

////////////////////////////////////////////////////////////////
//   IO

// OpenJSON opens colors from a JSON-formatted file.
func (nd *NetData) OpenJSON(filename core.Filename) error { //types:add
	fp, err := os.Open(string(filename))
	defer fp.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	ext := filepath.Ext(string(filename))
	if ext == ".gz" {
		gzr, err := gzip.NewReader(fp)
		defer gzr.Close()
		if err != nil {
			log.Println(err)
			return err
		}
		return nd.ReadJSON(gzr)
	} else {
		return nd.ReadJSON(bufio.NewReader(fp))
	}
}

// SaveJSON saves colors to a JSON-formatted file.
func (nd *NetData) SaveJSON(filename core.Filename) error { //types:add
	fp, err := os.Create(string(filename))
	defer fp.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	ext := filepath.Ext(string(filename))
	if ext == ".gz" {
		gzr := gzip.NewWriter(fp)
		err = nd.WriteJSON(gzr)
		gzr.Close()
	} else {
		bw := bufio.NewWriter(fp)
		err = nd.WriteJSON(bw)
		bw.Flush()
	}
	return err
}

// ReadJSON reads netdata from JSON format
func (nd *NetData) ReadJSON(r io.Reader) error {
	dec := json.NewDecoder(r)
	err := dec.Decode(nd) // this is way to do it on reader instead of bytes
	nan := math32.NaN()
	for _, ld := range nd.LayData {
		for i := range ld.Data {
			if ld.Data[i] == NaNSub {
				ld.Data[i] = nan
			}
		}
	}
	if err == nil || err == io.EOF {
		return nil
	}
	log.Println(err)
	return err
}

// NaNSub is used to replace NaN values for saving -- JSON doesn't handle nan's
const NaNSub = -1.11e-37

// WriteJSON writes netdata to JSON format
func (nd *NetData) WriteJSON(w io.Writer) error {
	for _, ld := range nd.LayData {
		for i := range ld.Data {
			if math32.IsNaN(ld.Data[i]) {
				ld.Data[i] = NaNSub
			}
		}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	err := enc.Encode(nd)
	if err != nil {
		log.Println(err)
	}
	return err
}

// func (ld *LayData) MarshalJSON() ([]byte, error) {
//
// }

// PlotSelectedUnit opens a window with a plot of all the data for the
// currently selected unit, saving data to the [tensorfs.CurRoot]/NetView
// directory.
// Useful for replaying detailed trace for units of interest.
func (nv *NetView) PlotSelectedUnit() (*table.Table, *plotcore.PlotEditor) { //types:add
	nd := &nv.Data
	if nd.PathLay == "" || nd.PathUnIndex < 0 {
		fmt.Printf("NetView:PlotSelectedUnit -- no unit selected\n")
		return nil, nil
	}

	selnm := nd.PathLay + fmt.Sprintf("[%d]", nd.PathUnIndex)
	dt := nd.SelectedUnitTable(nv.Di)
	for _, vnm := range nd.UnVars {
		vp, ok := nv.VarOptions[vnm]
		if !ok {
			continue
		}
		disp := (vnm == nv.Var)
		min := vp.Range.Min
		if min < 0 && vp.Range.FixMin && vp.MinMax.Min >= 0 {
			min = 0 // netview uses -1..1 but not great for graphs unless needed
		}
		dc := dt.Column(vnm)
		plot.AddStylerTo(dc, func(s *plot.Style) {
			s.On = disp
			s.Range.SetMin(float64(min)).SetMax(float64(vp.Range.Max))
		})
	}
	if tensorfs.CurRoot != nil && lab.CurTabber != nil {
		dir := tensorfs.CurRoot.Dir("NetView")
		udir := dir.Dir(selnm)
		tensorfs.DirFromTable(udir, dt)
		plt := lab.CurTabber.PlotTensorFS(udir)
		return dt, plt
	} else {
		b := core.NewBody("netview-selectedunit").SetTitle("NetView SelectedUnit Plot: " + selnm)
		plt := plotcore.NewPlotEditor(b)
		plt.SetTable(dt)
		b.AddTopBar(func(bar *core.Frame) {
			core.NewToolbar(bar).Maker(plt.MakeToolbar)
		})
		b.RunWindow()
		return dt, plt
	}
}

// SelectedUnitTable returns a table with all of the data for the
// currently-selected unit, and data parallel index.
func (nd *NetData) SelectedUnitTable(di int) *table.Table {
	if nd.PathLay == "" || nd.PathUnIndex < 0 {
		fmt.Printf("NetView:SelectedUnitTable -- no unit selected\n")
		return nil
	}

	ld, ok := nd.LayData[nd.PathLay]
	if !ok {
		fmt.Printf("NetView:SelectedUnitTable -- layer name incorrect\n")
		return nil
	}

	selnm := nd.PathLay + fmt.Sprintf("[%d]", nd.PathUnIndex)

	dt := table.New()
	metadata.SetName(dt, "NetView: "+selnm)
	metadata.SetTo(dt, "read-only", true)
	tensor.SetPrecision(dt, 4)

	ln := nd.Ring.Len
	vlen := len(nd.UnVars)
	nu := ld.NUnits
	nvu := vlen * nd.MaxData * nu
	uidx1d := nd.PathUnIndex

	dt.AddIntColumn("Rec")
	for _, vnm := range nd.UnVars {
		dt.AddFloat64Column(vnm)
	}
	dt.SetNumRows(ln)

	for ri := 0; ri < ln; ri++ {
		ridx := nd.RecIndex(ri)
		dt.Columns.Values[0].SetFloat1D(float64(ri), ri)
		for vi := 0; vi < vlen; vi++ {
			idx := ridx*nvu + vi*nd.MaxData*nu + di*nu + uidx1d
			val := ld.Data[idx]
			dt.Columns.Values[vi+1].SetFloat1D(float64(val), ri)
		}
	}
	return dt
}

/*
var NetDataProps = tree.Props{
	"CallMethods": tree.PropSlice{
		{"SaveJSON", tree.Props{
			"desc": "save recorded network view data to file",
			"icon": "file-save",
			"Args": tree.PropSlice{
				{"File Name", tree.Props{
					"ext": ".netdat,.netdat.gz",
				}},
			},
		}},
		{"OpenJSON", tree.Props{
			"desc": "open recorded network view data from file",
			"icon": "file-open",
			"Args": tree.PropSlice{
				{"File Name", tree.Props{
					"ext": ".netdat,.netdat.gz",
				}},
			},
		}},
	},
}
*/
