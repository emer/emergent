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
	"strconv"
	"strings"

	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/ringidx"
	"github.com/emer/etable/eplot"
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/goki/gi/gi"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
	"github.com/goki/mat32"
)

// NetData maintains a record of all the network data that has been displayed
// up to a given maximum number of records (updates), using efficient ring index logic
// with no copying to store in fixed-sized buffers.
type NetData struct {
	Net        emer.Network        `json:"-" desc:"the network that we're viewing"`
	PrjnLay    string              `desc:"name of the layer with unit for viewing projections (connection / synapse-level values)"`
	PrjnUnIdx  int                 `desc:"1D index of unit within PrjnLay for for viewing projections"`
	PrjnType   string              `inactive:"+" desc:"copied from NetView Params: if non-empty, this is the type projection to show when there are multiple projections from the same layer -- e.g., Inhib, Lateral, Forward, etc"`
	Vars       []string            `desc:"the list of unit variables saved"`
	VarIdxs    map[string]int      `desc:"index of each variable in the Vars slice"`
	Ring       ringidx.Idx         `desc:"the circular ring index -- Max here is max number of values to store, Len is number stored, and Idx(Len-1) is the most recent one, etc"`
	LayData    map[string]*LayData `desc:"the layer data -- map keyed by layer name"`
	MinPer     []float32           `desc:"min values for each Ring.Max * variable"`
	MaxPer     []float32           `desc:"max values for each Ring.Max * variable"`
	MinVar     []float32           `desc:"min values for variable"`
	MaxVar     []float32           `desc:"max values for variable"`
	Counters   []string            `desc:"counter strings"`
	RasterCtrs []int               `desc:"raster counter values"`
	RasterMap  map[int]int         `desc:"map of raster counter values to record numbers"`
	RastCtr    int                 `desc:"dummy raster counter when passed a -1"`
}

var KiT_NetData = kit.Types.AddType(&NetData{}, NetDataProps)

// Init initializes the main params and configures the data
func (nd *NetData) Init(net emer.Network, max int) {
	nd.Net = net
	nd.Ring.Max = max
	nd.Config()
	nd.RasterMap = make(map[int]int)
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
	nvars := nd.Net.UnitVarNames()
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
			ld := &LayData{LayName: nm, NUnits: lay.Shape().Len()}
			nd.LayData[nm] = ld
			ld.AllocSendPrjns(lay)
		}
		for li := 0; li < nlay; li++ {
			rlay := nd.Net.Layer(li)
			rld := nd.LayData[rlay.Name()]
			rp := rlay.RecvPrjns()
			rld.RecvPrjns = make([]*PrjnData, len(*rp))
			for ri, rpj := range *rp {
				slay := rpj.SendLay()
				sld := nd.LayData[slay.Name()]
				for _, spj := range sld.SendPrjns {
					if spj.Prjn == rpj {
						rld.RecvPrjns[ri] = spj // link
					}
				}
			}
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
		ld.NUnits = lay.Shape().Len()
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
		nd.RasterCtrs = make([]int, rmax)
	}
}

// Record records the current full set of data from the network,
// and the given counters string (displayed at bottom of window)
// and raster counter value -- if negative, then an internal
// wraping-around counter is used.
func (nd *NetData) Record(ctrs string, rastCtr, rastMax int) {
	nlay := nd.Net.NLayers()
	if nlay == 0 {
		return
	}
	nd.Config() // inexpensive if no diff, and safe..
	vlen := len(nd.Vars)
	nd.Ring.Add(1)
	lidx := nd.Ring.LastIdx()

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
			lay.UnitVals(&dvals, vnm)
			for ui := range dvals {
				vl := dvals[ui]
				if !mat32.IsNaN(vl) {
					*mn = mat32.Min(*mn, vl)
					*mx = mat32.Max(*mx, vl)
				}
			}
		}
	}
	nd.UpdateVarRange()
}

// RecordSyns records synaptic data -- stored separately
// and only needs to be called when synaptic values are updated.
func (nd *NetData) RecordSyns() {
	nlay := nd.Net.NLayers()
	if nlay == 0 {
		return
	}
	nd.Config() // inexpensive if no diff, and safe..
	// mmidx := lidx * vlen
	// for vi := range nd.Vars {
	// 	nd.MinPer[mmidx+vi] = math.MaxFloat32
	// 	nd.MaxPer[mmidx+vi] = -math.MaxFloat32
	// }
	for li := 0; li < nlay; li++ {
		lay := nd.Net.Layer(li)
		laynm := lay.Name()
		ld := nd.LayData[laynm]
		for _, spd := range ld.SendPrjns {
			spd.RecordData()
		}
	}
	// nd.UpdateVarRange()
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
			*vmn = mat32.Min(*vmn, mn)
			*vmx = mat32.Max(*vmx, mx)
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
	ridx := nd.RecIdx(recno)
	return nd.UnitValIdx(laynm, vnm, uidx1d, ridx)
}

// RasterCtr returns the raster counter value at given record number (-1 = current)
func (nd *NetData) RasterCtr(recno int) (int, bool) {
	if nd.Ring.Len == 0 {
		return 0, false
	}
	ridx := nd.RecIdx(recno)
	return nd.RasterCtrs[ridx], true
}

// UnitValRaster returns the value for given layer, variable name, unit index, and
// raster counter number.
// Returns false if value unavailable for any reason (including recorded as such as NaN).
func (nd *NetData) UnitValRaster(laynm string, vnm string, uidx1d int, rastCtr int) (float32, bool) {
	ridx, has := nd.RasterMap[rastCtr]
	if !has {
		return 0, false
	}
	return nd.UnitValIdx(laynm, vnm, uidx1d, ridx)
}

// UnitValIdx returns the value for given layer, variable name, unit index, and stored idx
// Returns false if value unavailable for any reason (including recorded as such as NaN).
func (nd *NetData) UnitValIdx(laynm string, vnm string, uidx1d int, ridx int) (float32, bool) {
	if strings.HasPrefix(vnm, "r.") {
		svar := vnm[2:]
		return nd.RecvUnitVal(laynm, svar, uidx1d)
	} else if strings.HasPrefix(vnm, "s.") {
		svar := vnm[2:]
		return nd.SendUnitVal(laynm, svar, uidx1d)
	}
	vi, ok := nd.VarIdxs[vnm]
	if !ok {
		return 0, false
	}
	vlen := len(nd.Vars)
	ld, ok := nd.LayData[laynm]
	if !ok {
		return 0, false
	}
	nu := ld.NUnits
	nvu := vlen * nu
	idx := ridx*nvu + vi*nu + uidx1d
	val := ld.Data[idx]
	if mat32.IsNaN(val) {
		return 0, false
	}
	return val, true
}

// RecvUnitVal returns the value for given layer, variable name, unit index,
// for receiving projection variable, based on recorded synaptic projection data.
// Returns false if value unavailable for any reason (including recorded as such as NaN).
func (nd *NetData) RecvUnitVal(laynm string, vnm string, uidx1d int) (float32, bool) {
	ld, ok := nd.LayData[laynm]
	if !ok || nd.PrjnLay == "" {
		return 0, false
	}
	recvLay := nd.Net.LayerByName(nd.PrjnLay)
	if recvLay == nil {
		return 0, false
	}
	var pj emer.Prjn
	var err error
	if nd.PrjnType != "" {
		pj, err = recvLay.RecvPrjns().SendNameTypeTry(laynm, nd.PrjnType)
		if pj == nil {
			pj, err = recvLay.RecvPrjns().SendNameTry(laynm)
		}
	} else {
		pj, err = recvLay.RecvPrjns().SendNameTry(laynm)
	}
	if pj == nil {
		return 0, false
	}
	var spd *PrjnData
	for _, pd := range ld.SendPrjns {
		if pd.Prjn == pj {
			spd = pd
			break
		}
	}
	if spd == nil {
		return 0, false
	}
	varIdx, err := pj.SynVarIdx(vnm)
	if err != nil {
		return 0, false
	}
	synIdx := pj.SynIdx(uidx1d, nd.PrjnUnIdx)
	if synIdx < 0 {
		return 0, false
	}
	nsyn := pj.Syn1DNum()
	val := spd.SynData[varIdx*nsyn+synIdx]
	return val, true
}

// SendUnitVal returns the value for given layer, variable name, unit index,
// for sending projection variable, based on recorded synaptic projection data.
// Returns false if value unavailable for any reason (including recorded as such as NaN).
func (nd *NetData) SendUnitVal(laynm string, vnm string, uidx1d int) (float32, bool) {
	ld, ok := nd.LayData[laynm]
	if !ok || nd.PrjnLay == "" {
		return 0, false
	}
	sendLay := nd.Net.LayerByName(nd.PrjnLay)
	if sendLay == nil {
		return 0, false
	}
	var pj emer.Prjn
	var err error
	if nd.PrjnType != "" {
		pj, err = sendLay.SendPrjns().RecvNameTypeTry(laynm, nd.PrjnType)
		if pj == nil {
			pj, err = sendLay.SendPrjns().RecvNameTry(laynm)
		}
	} else {
		pj, err = sendLay.SendPrjns().RecvNameTry(laynm)
	}
	if pj == nil {
		return 0, false
	}
	var rpd *PrjnData
	for _, pd := range ld.RecvPrjns {
		if pd.Prjn == pj {
			rpd = pd
			break
		}
	}
	if rpd == nil {
		return 0, false
	}
	varIdx, err := pj.SynVarIdx(vnm)
	if err != nil {
		return 0, false
	}
	synIdx := pj.SynIdx(nd.PrjnUnIdx, uidx1d)
	if synIdx < 0 {
		return 0, false
	}
	nsyn := pj.Syn1DNum()
	val := rpd.SynData[varIdx*nsyn+synIdx]
	return val, true
}

////////////////////////////////////////////////////////////////
//   IO

// OpenJSON opens colors from a JSON-formatted file.
func (nd *NetData) OpenJSON(filename gi.FileName) error {
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
func (nd *NetData) SaveJSON(filename gi.FileName) error {
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
	nan := mat32.NaN()
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
			if mat32.IsNaN(ld.Data[i]) {
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
// currently-selected unit.
// Useful for replaying detailed trace for units of interest.
func (nv *NetView) PlotSelectedUnit() (*gi.Window, *etable.Table, *eplot.Plot2D) {
	width := 1600
	height := 1200

	nd := &nv.Data

	if nd.PrjnLay == "" || nd.PrjnUnIdx < 0 {
		fmt.Printf("NetView:PlotSelectedUnit -- no unit selected\n")
		return nil, nil, nil
	}

	selnm := nd.PrjnLay + fmt.Sprintf("[%d]", nd.PrjnUnIdx)

	win := gi.NewMainWindow("netview-selectedunit", "NetView SelectedUnit Plot: "+selnm, width, height)
	vp := win.WinViewport2D()
	updt := vp.UpdateStart()
	mfr := win.SetMainFrame()

	plt := mfr.AddNewChild(eplot.KiT_Plot2D, "plot").(*eplot.Plot2D)
	plt.Params.Title = "NetView " + selnm
	plt.Params.XAxisCol = "Rec"

	dt := nd.SelectedUnitTable()

	plt.SetTable(dt)

	for _, vnm := range nd.Vars {
		vp, ok := nv.VarParams[vnm]
		if !ok {
			continue
		}
		disp := (vnm == nv.Var)
		min := vp.Range.Min
		if min < 0 && vp.Range.FixMin && vp.MinMax.Min >= 0 {
			min = 0 // netview uses -1..1 but not great for graphs unless needed
		}
		plt.SetColParams(vnm, disp, vp.Range.FixMin, float64(min), vp.Range.FixMax, float64(vp.Range.Max))
	}

	vp.UpdateEndNoSig(updt)
	win.GoStartEventLoop() // in a separate goroutine
	return win, dt, plt
}

// SelectedUnitTable returns a table with all of the data for the
// currently-selected unit.
func (nd *NetData) SelectedUnitTable() *etable.Table {
	if nd.PrjnLay == "" || nd.PrjnUnIdx < 0 {
		fmt.Printf("NetView:SelectedUnitTable -- no unit selected\n")
		return nil
	}

	ld, ok := nd.LayData[nd.PrjnLay]
	if !ok {
		fmt.Printf("NetView:SelectedUnitTable -- layer name incorrect\n")
		return nil
	}

	selnm := nd.PrjnLay + fmt.Sprintf("[%d]", nd.PrjnUnIdx)

	dt := &etable.Table{}
	dt.SetMetaData("name", "NetView: "+selnm)
	dt.SetMetaData("read-only", "true")
	dt.SetMetaData("precision", strconv.Itoa(4))

	ln := nd.Ring.Len
	vlen := len(nd.Vars)
	nu := ld.NUnits
	nvu := vlen * nu
	uidx1d := nd.PrjnUnIdx

	sch := etable.Schema{
		{"Rec", etensor.INT64, nil, nil},
	}
	for _, vnm := range nd.Vars {
		sch = append(sch, etable.Column{vnm, etensor.FLOAT64, nil, nil})
	}
	dt.SetFromSchema(sch, ln)

	for ri := 0; ri < ln; ri++ {
		ridx := nd.RecIdx(ri)
		dt.SetCellFloatIdx(0, ri, float64(ri))
		for vi := 0; vi < vlen; vi++ {
			idx := ridx*nvu + vi*nu + uidx1d
			val := ld.Data[idx]
			dt.SetCellFloatIdx(vi+1, ri, float64(val))
		}
	}
	return dt
}

var NetDataProps = ki.Props{
	"CallMethods": ki.PropSlice{
		{"SaveJSON", ki.Props{
			"desc": "save recorded network view data to file",
			"icon": "file-save",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".netdat,.netdat.gz",
				}},
			},
		}},
		{"OpenJSON", ki.Props{
			"desc": "open recorded network view data from file",
			"icon": "file-open",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".netdat,.netdat.gz",
				}},
			},
		}},
	},
}
