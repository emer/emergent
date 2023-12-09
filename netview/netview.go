// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package netview provides the NetView interactive 3D network viewer, implemented in the GoGi 3D framework.
*/
package netview

//go:generate goki generate

import (
	"fmt"
	"image/color"
	"log"
	"reflect"
	"strings"
	"sync"

	"github.com/emer/emergent/v2/emer"
	"goki.dev/colors"
	"goki.dev/colors/colormap"
	"goki.dev/etable/v2/minmax"
	"goki.dev/gi/v2/gi"
	"goki.dev/gi/v2/giv"
	"goki.dev/gi/v2/texteditor"
	"goki.dev/gi/v2/xyzv"
	"goki.dev/girl/styles"
	"goki.dev/goosi/events"
	"goki.dev/goosi/events/key"
	"goki.dev/icons"
	"goki.dev/ki/v2"
	"goki.dev/mat32/v2"
	"goki.dev/xyz"
)

// NetView is a GoGi Widget that provides a 3D network view using the GoGi gi3d
// 3D framework.
type NetView struct {
	gi.Layout

	// the network that we're viewing
	Net emer.Network `set:"-"`

	// current variable that we're viewing
	Var string `set:"-"`

	// current data parallel index di, for networks capable of processing input patterns in parallel.
	Di int

	// the list of variables to view
	Vars []string

	// list of synaptic variables
	SynVars []string

	// map of synaptic variable names to index
	SynVarsMap map[string]int

	// parameters for the list of variables to view
	VarParams map[string]*VarParams

	// current var params -- only valid during Update of display
	CurVarParams *VarParams `json:"-" xml:"-" view:"-"`

	// parameters controlling how the view is rendered
	Params Params

	// color map for mapping values to colors -- set by name in Params
	ColorMap *colormap.Map

	// record number to display -- use -1 to always track latest, otherwise in range
	RecNo int

	// last non-empty counters string provided -- re-used if no new one
	LastCtrs string

	// contains all the network data with history
	Data NetData

	// mutex on data access
	DataMu sync.RWMutex `view:"-" copy:"-" json:"-" xml:"-"`
}

func (nv *NetView) OnInit() {
	nv.Params.NetView = nv
	nv.Params.Defaults()
	nv.ColorMap = colormap.AvailMaps[string(nv.Params.ColorMap)]
	nv.RecNo = -1
	nv.Style(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Grow.Set(1, 1)
		// nv.SetProp("spacing", gi.StdDialogVSpaceUnits)
	})
}

// SetNet sets the network to view and updates view
func (nv *NetView) SetNet(net emer.Network) {
	nv.Net = net
	nv.DataMu.Lock()
	nv.Data.Init(nv.Net, nv.Params.MaxRecs, nv.Params.NoSynData, nv.Net.MaxParallelData())
	nv.DataMu.Unlock()
	nv.ConfigNetView()
}

// SetVar sets the variable to view and updates the display
func (nv *NetView) SetVar(vr string) {
	nv.DataMu.Lock()
	nv.Var = vr
	nv.VarsUpdate()
	nv.VarScaleUpdate(nv.Var)
	nv.DataMu.Unlock()
	nv.Update()
}

// SetMaxRecs sets the maximum number of records that are maintained (default 210)
// resets the current data in the process
func (nv *NetView) SetMaxRecs(max int) {
	nv.Params.MaxRecs = max
	nv.Data.Init(nv.Net, nv.Params.MaxRecs, nv.Params.NoSynData, nv.Net.MaxParallelData())
}

// HasLayers returns true if network has any layers -- else no display
func (nv *NetView) HasLayers() bool {
	if nv.Net == nil || nv.Net.NLayers() == 0 {
		return false
	}
	return true
}

// RecordCounters saves the counters, so they are available for a Current update
func (nv *NetView) RecordCounters(counters string) {
	nv.DataMu.Lock()
	defer nv.DataMu.Unlock()
	if counters != "" {
		nv.LastCtrs = counters
	}
}

// Record records the current state of the network, along with provided counters
// string, which is displayed at the bottom of the view to show the current
// state of the counters. The rastCtr is the raster counter value used for
// the raster plot mode -- use -1 for a default incrementing counter.
// The NetView displays this recorded data when Update is next called.
func (nv *NetView) Record(counters string, rastCtr int) {
	nv.DataMu.Lock()
	defer nv.DataMu.Unlock()
	if counters != "" {
		nv.LastCtrs = counters
	}
	nv.Data.PrjnType = nv.Params.PrjnType
	nv.Data.Record(nv.LastCtrs, rastCtr, nv.Params.Raster.Max)
	nv.RecTrackLatest() // if we make a new record, then user expectation is to track latest..
}

// RecordSyns records synaptic data -- stored separate from unit data
// and only needs to be called when synaptic values are updated.
// Should be done when the DWt values have been computed, before
// updating Wts and zeroing.
// NetView displays this recorded data when Update is next called.
func (nv *NetView) RecordSyns() {
	nv.DataMu.Lock()
	defer nv.DataMu.Unlock()
	nv.Data.RecordSyns()
}

// todo: probably don't need:

// GoUpdateView is the update call to make from another go routine
// it does the proper blocking to coordinate with GUI updates
// generated on the main GUI thread.
func (nv *NetView) GoUpdateView() {
	nv.UpdateView()
}

// UpdateView updates the display based on last recorded state of network.
func (nv *NetView) UpdateView() {
	if !nv.IsVisible() || !nv.HasLayers() {
		return
	}
	vs := nv.Scene3D()
	updt := vs.UpdateStart3D()
	nv.UpdateImpl()
	vs.UpdateEndRender3D(updt)
}

// Current records the current state of the network, including synaptic values,
// and updates the display.  Use this when switching to NetView tab after network
// has been running while viewing another tab, because the network state
// is typically not recored then.
func (nv *NetView) Current() { //gti:add
	nv.Record("", -1)
	nv.RecordSyns()
	nv.Update()
}

// UpdateImpl does the guts of updating -- backend for Update or GoUpdate
func (nv *NetView) UpdateImpl() {
	nv.DataMu.Lock()
	vp, ok := nv.VarParams[nv.Var]
	if !ok {
		nv.DataMu.Unlock()
		log.Printf("NetView: %v variable: %v not found\n", nv.Nm, nv.Var)
		return
	}
	nv.CurVarParams = vp

	if !vp.Range.FixMin || !vp.Range.FixMax {
		needUpdt := false
		// need to autoscale
		min, max, ok := nv.Data.VarRange(nv.Var)
		if ok {
			vp.MinMax.Set(min, max)
			if !vp.Range.FixMin {
				nmin := float32(minmax.NiceRoundNumber(float64(min), true)) // true = below
				if vp.Range.Min != nmin {
					vp.Range.Min = nmin
					needUpdt = true
				}
			}
			if !vp.Range.FixMax {
				nmax := float32(minmax.NiceRoundNumber(float64(max), false)) // false = above
				if vp.Range.Max != nmax {
					vp.Range.Max = nmax
					needUpdt = true
				}
			}
			if vp.ZeroCtr && !vp.Range.FixMin && !vp.Range.FixMax {
				bmax := mat32.Max(mat32.Abs(vp.Range.Max), mat32.Abs(vp.Range.Min))
				if !needUpdt {
					if vp.Range.Max != bmax || vp.Range.Min != -bmax {
						needUpdt = true
					}
				}
				vp.Range.Max = bmax
				vp.Range.Min = -bmax
			}
			if needUpdt {
				nv.VarScaleUpdate(nv.Var)
			}
		}
	}

	vs := nv.Scene()
	laysGp, err := vs.ChildByNameTry("Layers", 0)
	if err != nil || laysGp.NumChildren() != nv.Net.NLayers() {
		nv.ConfigNetView()
	}
	nv.SetCounters(nv.Data.CounterRec(nv.RecNo))
	nv.UpdateRecNo()
	nv.DataMu.Unlock()
	vs.UpdateMeshes()
}

func (nv *NetView) ConfigWidget() {
	nv.ConfigNetView()
}

// ConfigNetView configures the overall view widget
func (nv *NetView) ConfigNetView() {
	updt := nv.UpdateStart()
	defer nv.UpdateEndLayout(updt)

	cmap, ok := colormap.AvailMaps[string(nv.Params.ColorMap)]
	if ok {
		nv.ColorMap = cmap
	} else {
		log.Printf("NetView: %v  ColorMap named: %v not found in colormap.AvailMaps\n", nv.Nm, nv.Params.ColorMap)
	}
	if !nv.HasChildren() {
		tb := gi.NewToolbar(nv, "tbar")
		nlay := gi.NewLayout(nv, "net")
		nlay.Style(func(s *styles.Style) {
			s.Direction = styles.Row
			s.Grow.Set(1, 1)
		})
		gi.NewLabel(nv, "counters")
		vb := gi.NewToolbar(nv, "vbar")

		vlay := gi.NewFrame(nlay, "vars")
		vlay.Style(func(s *styles.Style) {
			s.Display = styles.Grid
			s.Columns = nv.Params.NVarCols
			s.Grow.Set(0, 1)
			s.Overflow.Y = styles.OverflowAuto
			s.BackgroundColor.SetSolid(colors.Scheme.SurfaceContainerLow)
		})

		xyzv.NewScene3D(nlay, "scene")

		nv.ConfigToolbar(tb)
		nv.ConfigViewbar(vb)
	}

	nv.VarsConfig()
	nv.ViewConfig()

	ctrs := nv.Counters()
	ctrs.SetText("Counters: " + strings.Repeat(" ", 100))

	nv.DataMu.Lock()
	nv.Data.Init(nv.Net, nv.Params.MaxRecs, nv.Params.NoSynData, nv.Net.MaxParallelData())
	nv.DataMu.Unlock()
	nv.ReconfigMeshes()
}

// ReconfigMeshes reconfigures the layer meshes
func (nv *NetView) ReconfigMeshes() {
	vs := nv.Scene()
	if vs.IsConfiged() {
		vs.ReconfigMeshes()
	}
}

func (nv *NetView) Toolbar() *gi.Toolbar {
	return nv.ChildByName("tbar", 0).(*gi.Toolbar)
}

func (nv *NetView) NetLay() *gi.Layout {
	return nv.ChildByName("net", 1).(*gi.Layout)
}

func (nv *NetView) Counters() *gi.Label {
	return nv.ChildByName("counters", 2).(*gi.Label)
}

func (nv *NetView) Viewbar() *gi.Toolbar {
	return nv.ChildByName("vbar", 3).(*gi.Toolbar)
}

func (nv *NetView) Scene3D() *xyzv.Scene3D {
	return nv.NetLay().ChildByName("scene", 1).(*xyzv.Scene3D)
}

func (nv *NetView) Scene() *xyz.Scene {
	return &(nv.Scene3D().Scene)
}

func (nv *NetView) VarsLay() *gi.Frame {
	return nv.NetLay().ChildByName("vars", 0).(*gi.Frame)
}

// SetCounters sets the counters widget view display at bottom of netview
func (nv *NetView) SetCounters(ctrs string) {
	ct := nv.Counters()
	if ct.Text != ctrs {
		ct.SetText(ctrs)
	}
}

// UpdateRecNo updates the record number viewing
func (nv *NetView) UpdateRecNo() {
	vbar := nv.Viewbar()
	rlbl := vbar.ChildByName("rec", 10).(*gi.Label)
	rlbl.SetText(fmt.Sprintf("%4d ", nv.RecNo))
}

// RecFullBkwd move view record to start of history.
func (nv *NetView) RecFullBkwd() bool {
	if nv.RecNo == 0 {
		return false
	}
	nv.RecNo = 0
	return true
}

// RecFastBkwd move view record N (default 10) steps backward. Returns true if updated.
func (nv *NetView) RecFastBkwd() bool {
	if nv.RecNo == 0 {
		return false
	}
	if nv.RecNo < 0 {
		nv.RecNo = nv.Data.Ring.Len - nv.Params.NFastSteps
	} else {
		nv.RecNo -= nv.Params.NFastSteps
	}
	if nv.RecNo < 0 {
		nv.RecNo = 0
	}
	return true
}

// RecBkwd move view record 1 steps backward. Returns true if updated.
func (nv *NetView) RecBkwd() bool {
	if nv.RecNo == 0 {
		return false
	}
	if nv.RecNo < 0 {
		nv.RecNo = nv.Data.Ring.Len - 1
	} else {
		nv.RecNo -= 1
	}
	if nv.RecNo < 0 {
		nv.RecNo = 0
	}
	return true
}

// RecFwd move view record 1 step forward. Returns true if updated.
func (nv *NetView) RecFwd() bool {
	if nv.RecNo >= nv.Data.Ring.Len-1 {
		nv.RecNo = nv.Data.Ring.Len - 1
		return false
	}
	if nv.RecNo < 0 {
		return false
	}
	nv.RecNo += 1
	if nv.RecNo >= nv.Data.Ring.Len-1 {
		nv.RecNo = nv.Data.Ring.Len - 1
	}
	return true
}

// RecFastFwd move view record N (default 10) steps forward. Returns true if updated.
func (nv *NetView) RecFastFwd() bool {
	if nv.RecNo >= nv.Data.Ring.Len-1 {
		nv.RecNo = nv.Data.Ring.Len - 1
		return false
	}
	if nv.RecNo < 0 {
		return false
	}
	nv.RecNo += nv.Params.NFastSteps
	if nv.RecNo >= nv.Data.Ring.Len-1 {
		nv.RecNo = nv.Data.Ring.Len - 1
	}
	return true
}

// RecTrackLatest sets view to track latest record (-1).  Returns true if updated.
func (nv *NetView) RecTrackLatest() bool {
	if nv.RecNo == -1 {
		return false
	}
	nv.RecNo = -1
	return true
}

// NetVarsList returns the list of layer and prjn variables for given network.
// layEven ensures that the number of layer variables is an even number if true
// (used for display but not storage).
func (nv *NetView) NetVarsList(net emer.Network, layEven bool) (nvars, synvars []string) {
	if net == nil || net.NLayers() == 0 {
		return nil, nil
	}
	unvars := net.UnitVarNames()
	synvars = net.SynVarNames()
	ulen := len(unvars)
	ncols := NVarCols // nv.Params.NVarCols
	nr := ulen % ncols
	if layEven && nr != 0 { // make it an even number
		ulen += ncols - nr
	}

	tlen := ulen + 2*len(synvars)
	nvars = make([]string, tlen)
	copy(nvars, unvars)
	st := ulen
	for pi := 0; pi < len(synvars); pi++ {
		nvars[st+2*pi] = "r." + synvars[pi]
		nvars[st+2*pi+1] = "s." + synvars[pi]
	}
	return
}

// VarsListUpdate updates the list of network variables
func (nv *NetView) VarsListUpdate() {
	nvars, synvars := nv.NetVarsList(nv.Net, true) // true = layEven
	if len(nvars) == len(nv.Vars) {
		return
	}
	nv.Vars = nvars
	nv.VarParams = make(map[string]*VarParams, len(nv.Vars))

	nv.SynVars = synvars
	nv.SynVarsMap = make(map[string]int, len(synvars))
	for i, vn := range nv.SynVars {
		nv.SynVarsMap[vn] = i
	}

	unprops := nv.Net.UnitVarProps()
	prjnprops := nv.Net.SynVarProps()
	for _, nm := range nv.Vars {
		vp := &VarParams{Var: nm}
		vp.Defaults()
		var vtag string
		if strings.HasPrefix(nm, "r.") || strings.HasPrefix(nm, "s.") {
			vtag = prjnprops[nm[2:]]
		} else {
			vtag = unprops[nm]
		}
		if vtag != "" {
			vp.SetProps(vtag)
		}
		nv.VarParams[nm] = vp
	}
}

// VarsUpdate updates the selection status of the variables
// and the view range state too
func (nv *NetView) VarsUpdate() {
	vl := nv.VarsLay()
	updt := vl.UpdateStart()
	for _, vbi := range *vl.Children() {
		vb := vbi.(*gi.Button)
		vb.SetSelected(vb.Text == nv.Var)
	}
	tbar := nv.Toolbar()
	cmapi := tbar.ChildByName("cmap", 5)
	if cmapi != nil {
		cmap := cmapi.(*giv.ColorMapView)
		cmap.Map = nv.ColorMap
		cmap.SetNeedsRender(updt)
	}
	vl.UpdateEndRender(updt)
}

// VarScaleUpdate updates display of the scaling params
// for given variable (use nv.Var for current)
// returns true if any setting changed (update always triggered)
func (nv *NetView) VarScaleUpdate(varNm string) bool {
	vp := nv.VarParams[varNm]

	tb := nv.Toolbar()
	mod := false

	if ci := tb.ChildByName("mnsw", 4); ci != nil {
		sw := ci.(*gi.Switch)
		if sw.IsChecked() != vp.Range.FixMin {
			mod = true
			sw.SetChecked(vp.Range.FixMin)
			sw.SetNeedsRender(true)
		}
	}
	if ci := tb.ChildByName("mxsw", 6); ci != nil {
		sw := ci.(*gi.Switch)
		if sw.IsChecked() != vp.Range.FixMax {
			mod = true
			sw.SetChecked(vp.Range.FixMax)
			sw.SetNeedsRender(true)
		}
	}
	if ci := tb.ChildByName("mnsp", 5); ci != nil {
		sp := ci.(*gi.Spinner)
		mnv := float32(vp.Range.Min)
		if sp.Value != mnv {
			mod = true
			sp.SetValue(mnv)
			sp.SetNeedsRender(true)
		}
	}
	if ci := tb.ChildByName("mxsp", 7); ci != nil {
		sp := ci.(*gi.Spinner)
		mxv := float32(vp.Range.Max)
		if sp.Value != mxv {
			mod = true
			sp.SetValue(mxv)
			sp.SetNeedsRender(true)
		}
	}
	if ci := tb.ChildByName("zcsw", 8); ci != nil {
		sw := ci.(*gi.Switch)
		if sw.IsChecked() != vp.ZeroCtr {
			mod = true
			sw.SetChecked(vp.ZeroCtr)
			sw.SetNeedsRender(true)
		}
	}
	return mod
}

// VarsConfig configures the variables
func (nv *NetView) VarsConfig() {
	vl := nv.VarsLay()
	nv.VarsListUpdate()
	if len(nv.Vars) == 0 {
		vl.DeleteChildren(true)
		return
	}
	if len(vl.Kids) == len(nv.Vars) {
		return
	}
	unprops := nv.Net.UnitVarProps()
	prjnprops := nv.Net.SynVarProps()
	for _, vn := range nv.Vars {
		vn := vn
		vb := gi.NewButton(vl).SetText(vn).SetType(gi.ButtonAction)
		pstr := ""
		if strings.HasPrefix(vn, "r.") || strings.HasPrefix(vn, "s.") {
			pstr = prjnprops[vn[2:]]
		} else {
			pstr = unprops[vn]
		}
		if pstr != "" {
			rstr := reflect.StructTag(pstr)
			if desc, ok := rstr.Lookup("desc"); ok {
				vb.Tooltip = vn + ": " + desc
			}
		}
		vb.SetSelected(vn == nv.Var)
		vb.OnClick(func(e events.Event) {
			nv.SetVar(vn)
		})
	}
}

// ViewConfig configures the 3D view
func (nv *NetView) ViewConfig() {
	vs := nv.Scene()
	if nv.Net == nil || nv.Net.NLayers() == 0 {
		vs.DeleteChildren(true)
		vs.Meshes.Reset()
		return
	}
	if vs.Lights.Len() == 0 {
		nv.ViewDefaults()
	}
	// todo:
	// vs.BgColor = gi.Prefs.Colors.Background // reset in case user changes
	nlay := nv.Net.NLayers()
	laysGp, err := vs.ChildByNameTry("Layers", 0)
	if err != nil {
		laysGp = xyz.NewGroup(vs, "Layers")
	}
	layConfig := ki.Config{}
	for li := 0; li < nlay; li++ {
		lay := nv.Net.Layer(li)
		lmesh := vs.MeshByName(lay.Name())
		if lmesh == nil {
			NewLayMesh(vs, nv, lay)
		} else {
			lmesh.(*LayMesh).Lay = lay // make sure
		}
		layConfig.Add(xyz.GroupType, lay.Name())
	}
	gpConfig := ki.Config{}
	gpConfig.Add(LayObjType, "layer")
	gpConfig.Add(LayNameType, "name")

	_, updt := laysGp.ConfigChildren(layConfig)

	nmin, nmax := nv.Net.Bounds()
	nsz := nmax.Sub(nmin).Sub(mat32.Vec3{1, 1, 0}).Max(mat32.Vec3{1, 1, 1})
	nsc := mat32.Vec3{1.0 / nsz.X, 1.0 / nsz.Y, 1.0 / nsz.Z}
	szc := mat32.Max(nsc.X, nsc.Y)
	poff := mat32.NewVec3Scalar(0.5)
	poff.Y = -0.5
	for li, lgi := range *laysGp.Children() {
		ly := nv.Net.Layer(li)
		lg := lgi.(*xyz.Group)
		lg.ConfigChildren(gpConfig) // won't do update b/c of above
		lp := ly.Pos()
		lp.Y = -lp.Y // reverse direction
		lp = lp.Sub(nmin).Mul(nsc).Sub(poff)
		rp := ly.RelPos()
		lg.Pose.Pos.Set(lp.X, lp.Z, lp.Y)
		lg.Pose.Scale.Set(nsc.X*rp.Scale, szc, nsc.Y*rp.Scale)

		lo := lg.Child(0).(*LayObj)
		lo.Defaults()
		lo.LayName = ly.Name()
		lo.NetView = nv
		lo.SetMeshName(ly.Name())
		lo.Mat.Color = colors.FromRGB(255, 100, 255)
		lo.Mat.Reflective = 8
		lo.Mat.Bright = 8
		lo.Mat.Shiny = 30
		// note: would actually be better to NOT cull back so you can view underneath
		// but then the front and back fight against each other, causing flickering

		txt := lg.Child(1).(*LayName)
		txt.Nm = "layname:" + ly.Name()
		txt.Defaults()
		txt.NetView = nv
		txt.SetText(ly.Name())
		txt.Pose.Scale = mat32.NewVec3Scalar(nv.Params.LayNmSize).Div(lg.Pose.Scale)
		// todo:
		// txt.SetProp("text-align", styles.Start)
		// txt.SetProp("text-vertical-align", styles.Start)
	}
	laysGp.UpdateEnd(updt)
}

// ViewDefaults are the default 3D view params
func (nv *NetView) ViewDefaults() {
	vs := nv.Scene()
	vs.Defaults()
	vs.Camera.Pose.Pos.Set(0, 1.5, 2.5) // more "top down" view shows more of layers
	// 	vs.Camera.Pose.Pos.Set(0, 1, 2.75) // more "head on" for larger / deeper networks
	vs.Camera.Near = 0.1
	vs.Camera.LookAt(mat32.Vec3{0, 0, 0}, mat32.Vec3{0, 1, 0})
	// todo:
	// vs.BgColor = gi.Prefs.Colors.Background
	xyz.NewAmbientLight(vs, "ambient", 0.1, xyz.DirectSun)
	dir := xyz.NewDirLight(vs, "dirUp", 0.3, xyz.DirectSun)
	dir.Pos.Set(0, 1, 0)
	dir = xyz.NewDirLight(vs, "dirBack", 0.3, xyz.DirectSun)
	dir.Pos.Set(0, 1, -2.5)
	// point := xyz.NewPointLight(vs, "point", 1, xyz.DirectSun)
	// point.Pos.Set(0, 2, 5)
	// spot := xyz.NewSpotLight(vs, "spot", 1, xyz.DirectSun)
	// spot.Pose.Pos.Set(0, 2, 5)
	// spot.LookAtOrigin()
}

// ReadLock locks data for reading -- call ReadUnlock when done.
// Call this surrounding calls to UnitVal.
func (nv *NetView) ReadLock() {
	nv.DataMu.RLock()
}

// ReadUnlock unlocks data for reading.
func (nv *NetView) ReadUnlock() {
	nv.DataMu.RUnlock()
}

// UnitVal returns the raw value, scaled value, and color representation
// for given unit of given layer. scaled is in range -1..1
func (nv *NetView) UnitVal(lay emer.Layer, idx []int) (raw, scaled float32, clr color.RGBA, hasval bool) {
	idx1d := lay.Shape().Offset(idx)
	if idx1d >= lay.Shape().Len() {
		raw, hasval = 0, false
	} else {
		raw, hasval = nv.Data.UnitVal(lay.Name(), nv.Var, idx1d, nv.RecNo, nv.Di)
	}
	scaled, clr = nv.UnitValColor(lay, idx1d, raw, hasval)
	return
}

// UnitValRaster returns the raw value, scaled value, and color representation
// for given unit of given layer, and given raster counter index value (0..RasterMax)
// scaled is in range -1..1
func (nv *NetView) UnitValRaster(lay emer.Layer, idx []int, rCtr int) (raw, scaled float32, clr color.RGBA, hasval bool) {
	rs := lay.RepShape()
	idx1d := rs.Offset(idx)
	ridx := lay.RepIdxs()
	if len(ridx) == 0 { // no rep
		if idx1d >= lay.Shape().Len() {
			raw, hasval = 0, false
		} else {
			raw, hasval = nv.Data.UnitValRaster(lay.Name(), nv.Var, idx1d, rCtr, nv.Di)
		}
	} else {
		if idx1d >= len(ridx) {
			raw, hasval = 0, false
		} else {
			idx1d = ridx[idx1d]
			raw, hasval = nv.Data.UnitValRaster(lay.Name(), nv.Var, idx1d, rCtr, nv.Di)
		}
	}
	scaled, clr = nv.UnitValColor(lay, idx1d, raw, hasval)
	return
}

var NilColor = color.RGBA{0x20, 0x20, 0x20, 0x40}

// UnitValColor returns the raw value, scaled value, and color representation
// for given unit of given layer. scaled is in range -1..1
func (nv *NetView) UnitValColor(lay emer.Layer, idx1d int, raw float32, hasval bool) (scaled float32, clr color.RGBA) {
	if nv.CurVarParams == nil || nv.CurVarParams.Var != nv.Var {
		ok := false
		nv.CurVarParams, ok = nv.VarParams[nv.Var]
		if !ok {
			return
		}
	}
	if !hasval {
		scaled = 0
		if lay.Name() == nv.Data.PrjnLay && idx1d == nv.Data.PrjnUnIdx {
			clr = color.RGBA{0x20, 0x80, 0x20, 0x80}
		} else {
			clr = NilColor
		}
	} else {
		clp := nv.CurVarParams.Range.ClipVal(raw)
		norm := nv.CurVarParams.Range.NormVal(clp)
		var op float32
		if nv.CurVarParams.ZeroCtr {
			scaled = float32(2*norm - 1)
			op = (nv.Params.ZeroAlpha + (1-nv.Params.ZeroAlpha)*mat32.Abs(scaled))
		} else {
			scaled = float32(norm)
			op = (nv.Params.ZeroAlpha + (1-nv.Params.ZeroAlpha)*0.8) // no meaningful alpha -- just set at 80\%
		}
		clr = colors.WithAF32(nv.ColorMap.Map(float64(norm)), op)
	}
	return
}

// ConfigLabels ensures that given label xyz.Text2D objects are created and initialized
// in a top-level group called Labels.  Use LabelByName() to get a given label, and
// LayerByName() to get a Layer group, whose Pose can be copied to put a label in
// position relative to a layer.  Default alignment is Left, Top.
// Returns true set of labels was changed (mods).
func (nv *NetView) ConfigLabels(labs []string) bool {
	vs := nv.Scene()
	lgp, err := vs.ChildByNameTry("Labels", 1)
	if err != nil {
		lgp = xyz.NewGroup(vs, "Labels")
	}

	lbConfig := ki.Config{}
	for _, ls := range labs {
		lbConfig.Add(xyz.Text2DType, ls)
	}
	mods, updt := lgp.ConfigChildren(lbConfig)
	if mods {
		for i, ls := range labs {
			lb := lgp.ChildByName(ls, i).(*xyz.Text2D)
			lb.Defaults()
			lb.SetText(ls)
			// todo:
			// lb.SetProp("text-align", styles.Start)
			// lb.SetProp("vertical-align", styles.Start)
			// lb.SetProp("white-space", styles.WhiteSpacePre)
		}
	}
	lgp.UpdateEnd(updt)
	return mods
}

// LabelByName returns given Text2D label (see ConfigLabels).
// nil if not found.
func (nv *NetView) LabelByName(lab string) *xyz.Text2D {
	vs := nv.Scene()
	lgp, err := vs.ChildByNameTry("Labels", 1)
	if err != nil {
		return nil
	}
	txt, err := lgp.ChildByNameTry(lab, 0)
	if err != nil {
		return nil
	}
	return txt.(*xyz.Text2D)
}

// LayerByName returns the xyz.Group that represents layer of given name.
// nil if not found.
func (nv *NetView) LayerByName(lay string) *xyz.Group {
	vs := nv.Scene()
	lgp, err := vs.ChildByNameTry("Layers", 0)
	if err != nil {
		return nil
	}
	ly, err := lgp.ChildByNameTry(lay, 0)
	if err != nil {
		return nil
	}
	return ly.(*xyz.Group)
}

func (nv *NetView) ConfigToolbar(tb *gi.Toolbar) {
	giv.NewFuncButton(tb, nv.Update).SetText("Init").SetIcon(icons.Update)
	giv.NewFuncButton(tb, nv.Current).SetIcon(icons.Update)
	gi.NewButton(tb).SetText("Config").SetIcon(icons.Settings).
		SetTooltip("set parameters that control display (font size etc)").
		OnClick(func(e events.Event) {
			d := gi.NewBody().AddTitle(nv.Nm + " Params")
			giv.NewStructView(d).SetStruct(&nv.Params)
			d.NewFullDialog(nv).Run()
		})
	gi.NewSeparator(tb)
	gi.NewButton(tb).SetText("Weights").SetType(gi.ButtonAction).SetMenu(func(m *gi.Scene) {
		giv.NewFuncButton(m, nv.SaveWeights).SetIcon(icons.Save)
		giv.NewFuncButton(m, nv.OpenWeights).SetIcon(icons.Open)
	})
	gi.NewButton(tb).SetText("Params").SetIcon(icons.Info).SetMenu(func(m *gi.Scene) {
		giv.NewFuncButton(m, nv.ShowNonDefaultParams).SetIcon(icons.Info)
		giv.NewFuncButton(m, nv.ShowAllParams).SetIcon(icons.Info)
		giv.NewFuncButton(m, nv.ShowKeyLayerParams).SetIcon(icons.Info)
		giv.NewFuncButton(m, nv.ShowKeyPrjnParams).SetIcon(icons.Info)
	})
	gi.NewButton(tb).SetText("Net Data").SetIcon(icons.Save).SetMenu(func(m *gi.Scene) {
		giv.NewFuncButton(m, nv.Data.SaveJSON).SetText("Save Net Data").SetIcon(icons.Save)
		giv.NewFuncButton(m, nv.Data.OpenJSON).SetText("Open Net Data").SetIcon(icons.Open)
		gi.NewSeparator(m)
		// giv.NewFuncButton(m, nv.PlotSelectedUnit).SetIcon(icons.Open)
	})
	gi.NewSeparator(tb)
	ditp := "data parallel index -- for models running multiple input patterns in parallel, this selects which one is viewed"
	gi.NewLabel(tb).SetText("Di:").SetTooltip(ditp)
	dis := gi.NewSpinner(tb).SetTooltip(ditp).SetMin(0).SetStep(1).SetValue(float32(nv.Di))
	dis.OnChange(func(e events.Event) {
		maxData := nv.Net.MaxParallelData()
		md := int(dis.Value)
		if md < maxData && md >= 0 {
			nv.Di = md
		}
		dis.SetValue(float32(nv.Di))
		nv.UpdateView()
	})
	gi.NewSeparator(tb)
	rchk := gi.NewSwitch(tb).SetText("Raster").
		SetTooltip("Toggles raster plot mode -- displays values on one axis (Z by default) and raster counter (time) along the other (X by default)").
		SetChecked(nv.Params.Raster.On)
	rchk.OnChange(func(e events.Event) {
		nv.Params.Raster.On = rchk.IsChecked()
		nv.ReconfigMeshes()
		nv.UpdateView()
	})
	xchk := gi.NewSwitch(tb).SetText("X").
		SetTooltip("If checked, the raster (time) dimension is plotted along the X (horizontal) axis of the layers, otherwise it goes in the depth (Z) dimension").
		SetChecked(nv.Params.Raster.XAxis)
	xchk.OnChange(func(e events.Event) {
		nv.Params.Raster.XAxis = xchk.IsChecked()
		nv.Update()
	})

	vp, ok := nv.VarParams[nv.Var]
	if !ok {
		vp = &VarParams{}
		vp.Defaults()
	}

	gi.NewSeparator(tb)
	mnsw := gi.NewSwitch(tb, "mnsw").SetText("Min").
		SetTooltip("Fix the minimum end of the displayed value range to value shown in next box.  Having both min and max fixed is recommended where possible for speed and consistent interpretability of the colors.").
		SetChecked(vp.Range.FixMin)
	mnsw.OnChange(func(e events.Event) {
		vp.Range.FixMin = mnsw.IsChecked()
		nv.VarScaleUpdate(nv.Var) // todo: before update?
		nv.UpdateView()
	})
	mnsp := gi.NewSpinner(tb, "mnsp").SetValue(float32(vp.Range.Min))
	mnsp.OnChange(func(e events.Event) {
		vp.Range.SetMin(mnsp.Value)
		if vp.ZeroCtr && vp.Range.Min < 0 && vp.Range.FixMax {
			vp.Range.SetMax(-vp.Range.Min)
		}
		nv.VarScaleUpdate(nv.Var)
		nv.UpdateView()
	})

	cmap := giv.NewColorMapView(tb, "cmap").SetColorMap(nv.ColorMap)
	cmap.SetTooltip("Color map for translating values into colors -- click to select alternative.").
		Style(func(s *styles.Style) {
			s.Min.X.Em(8)
			s.Min.Y.Em(1.2)
			s.Grow.Set(0, 0)
		})
	cmap.OnClick(func(e events.Event) {
		if cmap.Map != nil {
			nv.Params.ColorMap = giv.ColorMapName(cmap.Map.Name)
			nv.ColorMap = cmap.Map
			nv.UpdateView()
		}
	})

	mxsw := gi.NewSwitch(tb, "mxsw").SetText("Max").
		SetTooltip("Fix the maximum end of the displayed value range to value shown in next box.  Having both min and max fixed is recommended where possible for speed and consistent interpretability of the colors.").
		SetChecked(vp.Range.FixMax)
	mxsw.OnChange(func(e events.Event) {
		vp.Range.FixMax = mxsw.IsChecked()
		nv.VarScaleUpdate(nv.Var)
		nv.UpdateView()
	})
	mxsp := gi.NewSpinner(tb, "mxsp").SetValue(float32(vp.Range.Max))
	mxsp.OnChange(func(e events.Event) {
		vp.Range.SetMax(mxsp.Value)
		if vp.ZeroCtr && vp.Range.Max > 0 && vp.Range.FixMin {
			vp.Range.SetMin(-vp.Range.Max)
		}
		nv.VarScaleUpdate(nv.Var)
		nv.UpdateView()
	})
	zcsw := gi.NewSwitch(tb, "zcsw").SetText("ZeroCtr").
		SetTooltip("keep Min - Max centered around 0, and use negative heights for units -- else use full min-max range for height (no negative heights)").
		SetChecked(vp.ZeroCtr)
	zcsw.OnChange(func(e events.Event) {
		vp.ZeroCtr = zcsw.IsChecked()
		nv.VarScaleUpdate(nv.Var)
		nv.Update()
	})
}

func (nv *NetView) ConfigViewbar(tb *gi.Toolbar) {
	gi.NewButton(tb).SetIcon(icons.Update).SetTooltip("reset to default initial display").
		OnClick(func(e events.Event) {
			nv.Scene().SetCamera("default")
			nv.UpdateView()
		})
	gi.NewButton(tb).SetIcon(icons.ZoomIn).SetTooltip("zoom in").
		OnClick(func(e events.Event) {
			nv.Scene().Camera.Zoom(-.05)
			nv.UpdateView()
		})
	gi.NewButton(tb).SetIcon(icons.ZoomOut).SetTooltip("zoom out").
		OnClick(func(e events.Event) {
			nv.Scene().Camera.Zoom(.05)
			nv.UpdateView()
		})
	gi.NewSeparator(tb)
	gi.NewLabel(tb).SetText("Rot:").SetTooltip("rotate display")
	gi.NewButton(tb).SetIcon(icons.KeyboardArrowLeft).OnClick(func(e events.Event) {
		nv.Scene().Camera.Orbit(5, 0)
		nv.UpdateView()
	})
	gi.NewButton(tb).SetIcon(icons.KeyboardArrowUp).OnClick(func(e events.Event) {
		nv.Scene().Camera.Orbit(0, 5)
		nv.UpdateView()
	})
	gi.NewButton(tb).SetIcon(icons.KeyboardArrowDown).OnClick(func(e events.Event) {
		nv.Scene().Camera.Orbit(0, -5)
		nv.UpdateView()
	})
	gi.NewButton(tb).SetIcon(icons.KeyboardArrowRight).OnClick(func(e events.Event) {
		nv.Scene().Camera.Orbit(-5, 0)
		nv.UpdateView()
	})
	gi.NewSeparator(tb)

	gi.NewLabel(tb).SetText("Pan:").SetTooltip("pan display")
	gi.NewButton(tb).SetIcon(icons.KeyboardArrowLeft).OnClick(func(e events.Event) {
		nv.Scene().Camera.Pan(-.2, 0)
		nv.UpdateView()
	})
	gi.NewButton(tb).SetIcon(icons.KeyboardArrowUp).OnClick(func(e events.Event) {
		nv.Scene().Camera.Pan(0, .2)
		nv.UpdateView()
	})
	gi.NewButton(tb).SetIcon(icons.KeyboardArrowDown).OnClick(func(e events.Event) {
		nv.Scene().Camera.Pan(0, -.2)
		nv.UpdateView()
	})
	gi.NewButton(tb).SetIcon(icons.KeyboardArrowRight).OnClick(func(e events.Event) {
		nv.Scene().Camera.Pan(.2, 0)
		nv.UpdateView()
	})
	gi.NewSeparator(tb)

	gi.NewLabel(tb).SetText("Save:")
	for i := 1; i <= 4; i++ {
		i := i
		nm := fmt.Sprintf("%d", i)
		gi.NewButton(tb).SetText(nm).SetIcon(icons.Save).
			SetTooltip("first click (or + Shift) saves current view, second click restores to saved state").
			OnClick(func(e events.Event) {
				sc := nv.Scene()
				cam := nm
				if e.HasAllModifiers(e.Modifiers(), key.Shift) {
					sc.SaveCamera(cam)
				} else {
					err := sc.SetCamera(cam)
					if err != nil {
						sc.SaveCamera(cam)
					}
				}
				fmt.Printf("Camera %s: %v\n", cam, sc.Camera.GenGoSet(""))
				nv.UpdateView()
			})
	}
	gi.NewSeparator(tb)

	gi.NewLabel(tb).SetText("Time:").
		SetTooltip("states are recorded over time -- last N can be reviewed using these buttons")

	gi.NewLabel(tb, "rec").SetText(fmt.Sprintf("%4d ", nv.RecNo)).
		SetTooltip("current view record: -1 means latest, 0 = earliest")
	gi.NewButton(tb).SetIcon(icons.FirstPage).SetTooltip("move to first record (start of history)").
		OnClick(func(e events.Event) {
			if nv.RecFullBkwd() {
				nv.UpdateView()
			}
		})
	gi.NewButton(tb).SetIcon(icons.FastRewind).SetTooltip("move earlier by N records (default 10)").
		OnClick(func(e events.Event) {
			if nv.RecFastBkwd() {
				nv.Update()
			}
		})
	gi.NewButton(tb).SetIcon(icons.SkipPrevious).SetTooltip("move earlier by 1").
		OnClick(func(e events.Event) {
			if nv.RecBkwd() {
				nv.Update()
			}
		})
	gi.NewButton(tb).SetIcon(icons.PlayArrow).SetTooltip("move to latest and always display latest (-1)").
		OnClick(func(e events.Event) {
			if nv.RecTrackLatest() {
				nv.Update()
			}
		})
	gi.NewButton(tb).SetIcon(icons.SkipNext).SetTooltip("move later by 1").
		OnClick(func(e events.Event) {
			if nv.RecFwd() {
				nv.Update()
			}
		})
	gi.NewButton(tb).SetIcon(icons.FastForward).SetTooltip("move later by N (default 10)").
		OnClick(func(e events.Event) {
			if nv.RecFastFwd() {
				nv.Update()
			}
		})
	gi.NewButton(tb).SetIcon(icons.LastPage).SetTooltip("move to end (current time, tracking latest updates)").
		OnClick(func(e events.Event) {
			if nv.RecTrackLatest() {
				nv.Update()
			}
		})
}

// SaveWeights saves the network weights -- when called with giv.CallMethod
// it will auto-prompt for filename
func (nv *NetView) SaveWeights(filename gi.FileName) {
	nv.Net.SaveWtsJSON(filename)
}

// OpenWeights opens the network weights -- when called with giv.CallMethod
// it will auto-prompt for filename
func (nv *NetView) OpenWeights(filename gi.FileName) {
	nv.Net.OpenWtsJSON(filename)
}

// ShowNonDefaultParams shows a dialog of all the parameters that
// are not at their default values in the network.  Useful for setting params.
func (nv *NetView) ShowNonDefaultParams() string {
	nds := nv.Net.NonDefaultParams()
	texteditor.TextDialog(nv, "Non Default Params: "+nv.Nm, nds)
	return nds
}

// ShowAllParams shows a dialog of all the parameters in the network.
func (nv *NetView) ShowAllParams() string {
	nds := nv.Net.AllParams()
	texteditor.TextDialog(nv, "All Params: "+nv.Nm, nds)
	return nds
}

// ShowKeyLayerParams shows a dialog with a listing for all layers in the network,
// of the most important layer-level params (specific to each algorithm)
func (nv *NetView) ShowKeyLayerParams() string {
	nds := nv.Net.KeyLayerParams()
	texteditor.TextDialog(nv, "Key Layer Params: "+nv.Nm, nds)
	return nds
}

// ShowKeyPrjnParams shows a dialog with a listing for all Recv projections in the network,
// of the most important projection-level params (specific to each algorithm)
func (nv *NetView) ShowKeyPrjnParams() string {
	nds := nv.Net.KeyPrjnParams()
	texteditor.TextDialog(nv, "Key Prjn Params: "+nv.Nm, nds)
	return nds
}
