// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package netview provides the NetView interactive 3D network viewer, implemented in the GoGi 3D framework.
*/
package netview

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"

	"github.com/emer/emergent/emer"
	"github.com/emer/etable/minmax"
	"github.com/goki/gi/colormap"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/gi3d"
	"github.com/goki/gi/gist"
	"github.com/goki/gi/giv"
	"github.com/goki/gi/oswin/key"
	"github.com/goki/gi/units"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
	"github.com/goki/mat32"
)

// NetView is a GoGi Widget that provides a 3D network view using the GoGi gi3d
// 3D framework.
type NetView struct {
	gi.Layout
	Net          emer.Network          `desc:"the network that we're viewing"`
	Var          string                `desc:"current variable that we're viewing"`
	Vars         []string              `desc:"the list of variables to view"`
	SynVars      []string              `desc:"list of synaptic variables"`
	SynVarsMap   map[string]int        `desc:"map of synaptic variable names to index"`
	VarParams    map[string]*VarParams `desc:"parameters for the list of variables to view"`
	CurVarParams *VarParams            `json:"-" xml:"-" view:"-" desc:"current var params -- only valid during Update of display"`
	Params       Params                `desc:"parameters controlling how the view is rendered"`
	ColorMap     *colormap.Map         `desc:"color map for mapping values to colors -- set by name in Params"`
	RecNo        int                   `desc:"record number to display -- use -1 to always track latest, otherwise in range [0..Data.Ring.Len-1]"`
	LastCtrs     string                `desc:"last non-empty counters string provided -- re-used if no new one"`
	Data         NetData               `desc:"contains all the network data with history"`
	DataMu       sync.RWMutex          `view:"-" copy:"-" json:"-" xml:"-" desc:"mutex on data access"`
}

var KiT_NetView = kit.Types.AddType(&NetView{}, NetViewProps)

// AddNewNetView adds a new NetView to given parent node, with given name.
func AddNewNetView(parent ki.Ki, name string) *NetView {
	return parent.AddNewChild(KiT_NetView, name).(*NetView)
}

func (nv *NetView) Defaults() {
	nv.Params.NetView = nv
	nv.Params.Defaults()
	nv.ColorMap = colormap.AvailMaps[string(nv.Params.ColorMap)]
	nv.RecNo = -1
}

// SetNet sets the network to view and updates view
func (nv *NetView) SetNet(net emer.Network) {
	nv.Defaults()
	nv.Net = net
	nv.DataMu.Lock()
	nv.Data.Init(nv.Net, nv.Params.MaxRecs)
	nv.DataMu.Unlock()
	nv.Config()
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
	nv.Data.Init(nv.Net, nv.Params.MaxRecs)
}

// HasLayers returns true if network has any layers -- else no display
func (nv *NetView) HasLayers() bool {
	if nv.Net == nil || nv.Net.NLayers() == 0 {
		return false
	}
	return true
}

// Record records the current state of the network, along with provided counters
// string, which is displayed at the bottom of the view to show the current
// state of the counters. The rastCtr is the raster counter value used for
// the raster plot mode.
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

// RecordSyns records the current state of the network synaptic level state
// which is stored separately and only needs to be called when synaptic values
// are updated. The NetView displays this recorded data when Update is next called.
func (nv *NetView) RecordSyns() {
	nv.DataMu.Lock()
	defer nv.DataMu.Unlock()
	nv.Data.RecordSyns()
}

// GoUpdate is the update call to make from another go routine
// it does the proper blocking to coordinate with GUI updates
// generated on the main GUI thread.
func (nv *NetView) GoUpdate() {
	if !nv.IsVisible() || !nv.HasLayers() {
		return
	}
	mvp := nv.ViewportSafe()
	if mvp.IsUpdatingNode() {
		return
	}
	mvp.BlockUpdates()
	vs := nv.Scene()
	updt := vs.UpdateStart()
	nv.UpdateImpl()
	mvp.UnblockUpdates()
	vs.UpdateEnd(updt)
}

// Update updates the display based on current state of network.
// This version is for calling within main window eventloop goroutine --
// use GoUpdate version for calling outside of main goroutine.
func (nv *NetView) Update() {
	if !nv.IsVisible() || !nv.HasLayers() {
		return
	}
	vs := nv.Scene()
	updt := vs.UpdateStart()
	nv.UpdateImpl()
	vs.UpdateEnd(updt)
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
		nv.Config()
	}
	nv.SetCounters(nv.Data.CounterRec(nv.RecNo))
	nv.UpdateRecNo()
	nv.DataMu.Unlock()
	vs.UpdateMeshes()
}

// Config configures the overall view widget
func (nv *NetView) Config() {
	nv.Lay = gi.LayoutVert
	if nv.Params.UnitSize == 0 {
		nv.Defaults()
	}
	cmap, ok := colormap.AvailMaps[string(nv.Params.ColorMap)]
	if ok {
		nv.ColorMap = cmap
	} else {
		log.Printf("NetView: %v  ColorMap named: %v not found in colormap.AvailMaps\n", nv.Nm, nv.Params.ColorMap)
	}
	nv.SetProp("spacing", gi.StdDialogVSpaceUnits)
	config := kit.TypeAndNameList{}
	config.Add(gi.KiT_ToolBar, "tbar")
	config.Add(gi.KiT_Layout, "net")
	config.Add(gi.KiT_Label, "counters")
	config.Add(gi.KiT_ToolBar, "vbar")
	mods, updt := nv.ConfigChildren(config)
	if !mods {
		updt = nv.UpdateStart()
	}

	nlay := nv.NetLay()
	nlay.Lay = gi.LayoutHoriz
	nlay.SetProp("max-width", -1)
	nlay.SetProp("max-height", -1)
	nlay.SetProp("spacing", gi.StdDialogVSpaceUnits)

	vncfg := kit.TypeAndNameList{}
	vncfg.Add(gi.KiT_Frame, "vars")
	vncfg.Add(gi3d.KiT_Scene, "scene")
	nlay.ConfigChildren(vncfg) // won't do update b/c of above updt

	nv.VarsConfig()
	nv.ViewConfig()
	nv.ToolbarConfig()
	nv.ViewbarConfig()

	ctrs := nv.Counters()
	ctrs.Redrawable = true
	ctrs.SetText("Counters: ")

	nv.DataMu.Lock()
	nv.Data.Init(nv.Net, nv.Params.MaxRecs)
	nv.DataMu.Unlock()
	nv.ReconfigMeshes()
	nv.UpdateEnd(updt)
}

// ReconfigMeshes reconfigures the layer meshes
func (nv *NetView) ReconfigMeshes() {
	vs := nv.Scene()
	if vs.IsConfiged() {
		vs.ReconfigMeshes()
	}
}

// IsConfiged returns true if widget is fully configured
func (nv *NetView) IsConfiged() bool {
	if len(nv.Kids) == 0 {
		return false
	}
	nl := nv.NetLay()
	if len(nl.Kids) == 0 {
		return false
	}
	return true
}

func (nv *NetView) Toolbar() *gi.ToolBar {
	return nv.ChildByName("tbar", 0).(*gi.ToolBar)
}

func (nv *NetView) NetLay() *gi.Layout {
	return nv.ChildByName("net", 1).(*gi.Layout)
}

func (nv *NetView) Counters() *gi.Label {
	return nv.ChildByName("counters", 2).(*gi.Label)
}

func (nv *NetView) Viewbar() *gi.ToolBar {
	return nv.ChildByName("vbar", 3).(*gi.ToolBar)
}

func (nv *NetView) Scene() *gi3d.Scene {
	return nv.NetLay().ChildByName("scene", 1).(*gi3d.Scene)
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
func NetVarsList(net emer.Network, layEven bool) (nvars, synvars []string) {
	if net == nil || net.NLayers() == 0 {
		return nil, nil
	}
	unvars := net.UnitVarNames()
	synvars = net.SynVarNames()
	ulen := len(unvars)
	if layEven && ulen%2 != 0 { // make it an even number, for 2 column layout
		ulen++
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
	nvars, synvars := NetVarsList(nv.Net, true) // true = layEven
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
		vb := vbi.(*gi.Action)
		if vb.Text == nv.Var {
			vb.SetSelected()
		} else {
			vb.ClearSelected()
		}
	}
	tbar := nv.Toolbar()
	cmap := tbar.ChildByName("cmap", 5).(*giv.ColorMapView)
	cmap.Map = nv.ColorMap
	cmap.UpdateSig()
	vl.UpdateEnd(updt)
}

// VarScaleUpdate updates display of the scaling params
// for given variable (use nv.Var for current)
// returns true if any setting changed (update always triggered)
func (nv *NetView) VarScaleUpdate(varNm string) bool {
	vp := nv.VarParams[varNm]

	tbar := nv.Toolbar()
	mncb := tbar.ChildByName("mncb", 4).(*gi.CheckBox)
	mnsb := tbar.ChildByName("mnsb", 5).(*gi.SpinBox)
	mxcb := tbar.ChildByName("mxcb", 6).(*gi.CheckBox)
	mxsb := tbar.ChildByName("mxsb", 7).(*gi.SpinBox)
	zccb := tbar.ChildByName("zccb", 8).(*gi.CheckBox)

	mod := false
	updt := false
	if mncb.IsChecked() != vp.Range.FixMin {
		updt = tbar.UpdateStart()
		mod = true
		mncb.SetChecked(vp.Range.FixMin)
	}
	if mxcb.IsChecked() != vp.Range.FixMax {
		if !mod {
			updt = tbar.UpdateStart()
			mod = true
		}
		mxcb.SetChecked(vp.Range.FixMax)
	}
	mnv := float32(vp.Range.Min)
	if mnsb.Value != mnv {
		if !mod {
			updt = tbar.UpdateStart()
			mod = true
		}
		mnsb.SetValue(mnv)
	}
	mxv := float32(vp.Range.Max)
	if mxsb.Value != mxv {
		if !mod {
			updt = tbar.UpdateStart()
			mod = true
		}
		mxsb.SetValue(mxv)
	}
	if zccb.IsChecked() != vp.ZeroCtr {
		if !mod {
			updt = tbar.UpdateStart()
			mod = true
		}
		zccb.SetChecked(vp.ZeroCtr)
	}
	tbar.UpdateEnd(updt)
	return mod
}

// VarsConfig configures the variables
func (nv *NetView) VarsConfig() {
	vl := nv.VarsLay()
	vl.SetReRenderAnchor()
	vl.Lay = gi.LayoutGrid
	vl.SetProp("columns", 2)
	vl.SetProp("spacing", 0)
	vl.SetProp("vertical-align", gist.AlignTop)
	nv.VarsListUpdate()
	if len(nv.Vars) == 0 {
		vl.DeleteChildren(true)
		return
	}
	config := kit.TypeAndNameList{}
	for _, vn := range nv.Vars {
		config.Add(gi.KiT_Action, vn)
	}
	mods, updt := vl.ConfigChildren(config)
	if !mods {
		updt = vl.UpdateStart()
	}
	unprops := nv.Net.UnitVarProps()
	prjnprops := nv.Net.SynVarProps()
	for i, vbi := range *vl.Children() {
		vb := vbi.(*gi.Action)
		vb.SetProp("margin", 0)
		vb.SetProp("max-width", -1)
		vn := nv.Vars[i]
		vb.SetText(vn)
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
		if vn == nv.Var {
			vb.SetSelected()
		} else {
			vb.ClearSelected()
		}
		vb.ActionSig.Connect(nv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			vbv := send.(*gi.Action)
			nvv.SetVar(vbv.Text)
		})
	}
	vl.UpdateEnd(updt)
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
	vs.BgColor = gi.Prefs.Colors.Background // reset in case user changes
	nlay := nv.Net.NLayers()
	laysGp, err := vs.ChildByNameTry("Layers", 0)
	if err != nil {
		laysGp = gi3d.AddNewGroup(vs, vs, "Layers")
	}
	layConfig := kit.TypeAndNameList{}
	for li := 0; li < nlay; li++ {
		lay := nv.Net.Layer(li)
		lmesh := vs.MeshByName(lay.Name())
		if lmesh == nil {
			AddNewLayMesh(vs, nv, lay)
		} else {
			lmesh.(*LayMesh).Lay = lay // make sure
		}
		layConfig.Add(gi3d.KiT_Group, lay.Name())
	}
	gpConfig := kit.TypeAndNameList{}
	gpConfig.Add(KiT_LayObj, "layer")
	gpConfig.Add(KiT_LayName, "name")

	_, updt := laysGp.ConfigChildren(layConfig)
	// if !mods {
	// 	updt = laysGp.UpdateStart()
	// }
	nmin, nmax := nv.Net.Bounds()
	nsz := nmax.Sub(nmin).Sub(mat32.Vec3{1, 1, 0}).Max(mat32.Vec3{1, 1, 1})
	nsc := mat32.Vec3{1.0 / nsz.X, 1.0 / nsz.Y, 1.0 / nsz.Z}
	szc := mat32.Max(nsc.X, nsc.Y)
	poff := mat32.NewVec3Scalar(0.5)
	poff.Y = -0.5
	for li, lgi := range *laysGp.Children() {
		ly := nv.Net.Layer(li)
		lg := lgi.(*gi3d.Group)
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
		lo.SetMeshName(vs, ly.Name())
		lo.Mat.Color.SetUInt8(255, 100, 255, 255)
		lo.Mat.Reflective = 1
		lo.Mat.Bright = 8
		// lo.Mat.Shiny = 10
		// note: would actually be better to NOT cull back so you can view underneath
		// but then the front and back fight against each other, causing flickering

		txt := lg.Child(1).(*LayName)
		txt.Nm = "layname:" + ly.Name()
		txt.Defaults(vs)
		txt.NetView = nv
		txt.SetText(vs, ly.Name())
		txt.Pose.Scale = mat32.NewVec3Scalar(nv.Params.LayNmSize).Div(lg.Pose.Scale)
		txt.SetProp("text-align", gist.AlignLeft)
		txt.SetProp("text-vertical-align", gist.AlignTop)
	}
	laysGp.UpdateEnd(updt)
}

// ViewDefaults are the default 3D view params
func (nv *NetView) ViewDefaults() {
	vs := nv.Scene()
	vs.SetStretchMax()
	vs.Defaults()
	vs.Camera.Pose.Pos.Set(0, 1.5, 2.5) // more "top down" view shows more of layers
	// 	vs.Camera.Pose.Pos.Set(0, 1, 2.75) // more "head on" for larger / deeper networks
	vs.Camera.Near = 0.1
	vs.Camera.LookAt(mat32.Vec3{0, 0, 0}, mat32.Vec3{0, 1, 0})
	vs.BgColor = gi.Prefs.Colors.Background
	gi3d.AddNewAmbientLight(vs, "ambient", 0.1, gi3d.DirectSun)
	dir := gi3d.AddNewDirLight(vs, "dirUp", 0.3, gi3d.DirectSun)
	dir.Pos.Set(0, 1, 0)
	dir = gi3d.AddNewDirLight(vs, "dirBack", 0.3, gi3d.DirectSun)
	dir.Pos.Set(0, 1, -2.5)
	// point := gi3d.AddNewPointLight(vs, "point", 1, gi3d.DirectSun)
	// point.Pos.Set(0, 2, 5)
	// spot := gi3d.AddNewSpotLight(vs, "spot", 1, gi3d.DirectSun)
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
func (nv *NetView) UnitVal(lay emer.Layer, idx []int) (raw, scaled float32, clr gist.Color, hasval bool) {
	idx1d := lay.Shape().Offset(idx)
	if idx1d >= lay.Shape().Len() {
		raw, hasval = 0, false
	} else {
		raw, hasval = nv.Data.UnitVal(lay.Name(), nv.Var, idx1d, nv.RecNo)
	}
	scaled, clr = nv.UnitValColor(lay, idx1d, raw, hasval)
	return
}

// UnitValRaster returns the raw value, scaled value, and color representation
// for given unit of given layer, and given raster counter index value (0..RasterMax)
// scaled is in range -1..1
func (nv *NetView) UnitValRaster(lay emer.Layer, idx []int, rCtr int) (raw, scaled float32, clr gist.Color, hasval bool) {
	idx1d := lay.Shape().Offset(idx)
	if idx1d >= lay.Shape().Len() {
		raw, hasval = 0, false
	} else {
		raw, hasval = nv.Data.UnitValRaster(lay.Name(), nv.Var, idx1d, rCtr)
	}
	scaled, clr = nv.UnitValColor(lay, idx1d, raw, hasval)
	return
}

var NilColor = gist.Color{0x20, 0x20, 0x20, 0x40}

// UnitValColor returns the raw value, scaled value, and color representation
// for given unit of given layer. scaled is in range -1..1
func (nv *NetView) UnitValColor(lay emer.Layer, idx1d int, raw float32, hasval bool) (scaled float32, clr gist.Color) {
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
			clr.SetUInt8(0x20, 0x80, 0x20, 0x80)
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
		clr = nv.ColorMap.Map(float64(norm))
		r, g, b, a := clr.ToNPFloat32()
		clr.SetNPFloat32(r, g, b, a*op)
	}
	return
}

// ConfigLabels ensures that given label gi3d.Text2D objects are created and initialized
// in a top-level group called Labels.  Use LabelByName() to get a given label, and
// LayerByName() to get a Layer group, whose Pose can be copied to put a label in
// position relative to a layer.  Default alignment is Left, Top.
// Returns true set of labels was changed (mods).
func (nv *NetView) ConfigLabels(labs []string) bool {
	vs := nv.Scene()
	lgp, err := vs.ChildByNameTry("Labels", 1)
	if err != nil {
		lgp = gi3d.AddNewGroup(vs, vs, "Labels")
	}

	lbConfig := kit.TypeAndNameList{}
	for _, ls := range labs {
		lbConfig.Add(gi3d.KiT_Text2D, ls)
	}
	mods, updt := lgp.ConfigChildren(lbConfig)
	if mods {
		for i, ls := range labs {
			lb := lgp.ChildByName(ls, i).(*gi3d.Text2D)
			lb.Defaults(vs)
			lb.SetText(vs, ls)
			lb.SetProp("text-align", gist.AlignLeft)
			lb.SetProp("vertical-align", gist.AlignTop)
			lb.SetProp("white-space", gist.WhiteSpacePre)
		}
	}
	lgp.UpdateEnd(updt)
	return mods
}

// LabelByName returns given Text2D label (see ConfigLabels).
// nil if not found.
func (nv *NetView) LabelByName(lab string) *gi3d.Text2D {
	vs := nv.Scene()
	lgp, err := vs.ChildByNameTry("Labels", 1)
	if err != nil {
		return nil
	}
	txt, err := lgp.ChildByNameTry(lab, 0)
	if err != nil {
		return nil
	}
	return txt.(*gi3d.Text2D)
}

// LayerByName returns the gi3d.Group that represents layer of given name.
// nil if not found.
func (nv *NetView) LayerByName(lay string) *gi3d.Group {
	vs := nv.Scene()
	lgp, err := vs.ChildByNameTry("Layers", 0)
	if err != nil {
		return nil
	}
	ly, err := lgp.ChildByNameTry(lay, 0)
	if err != nil {
		return nil
	}
	return ly.(*gi3d.Group)
}

func (nv *NetView) ToolbarConfig() {
	tbar := nv.Toolbar()
	if len(tbar.Kids) != 0 {
		return
	}
	tbar.SetStretchMaxWidth()
	tbar.AddAction(gi.ActOpts{Label: "Init", Icon: "update", Tooltip: "fully redraw display"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			nvv.Config()
			nvv.Update()
			nvv.VarsUpdate()
		})
	tbar.AddAction(gi.ActOpts{Label: "Config", Icon: "gear", Tooltip: "set parameters that control display (font size etc)"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			giv.StructViewDialog(nvv.ViewportSafe(), &nvv.Params, giv.DlgOpts{Title: nvv.Nm + " Params"}, nil, nil)
		})
	tbar.AddSeparator("file")
	wtsmen := tbar.AddAction(gi.ActOpts{Label: "Weights", Icon: "file-save"}, nil, nil)
	wtsmen.Menu.AddAction(gi.ActOpts{Label: "Save Wts", Icon: "file-save"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			giv.CallMethod(nvv, "SaveWeights", nvv.ViewportSafe()) // this auto prompts for filename using file chooser
		})
	wtsmen.Menu.AddAction(gi.ActOpts{Label: "Open Wts", Icon: "file-open"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			giv.CallMethod(nvv, "OpenWeights", nvv.ViewportSafe()) // this auto prompts for filename using file chooser
		})
	parsmen := tbar.AddAction(gi.ActOpts{Label: "Params", Icon: "info"}, nil, nil)
	parsmen.Menu.AddAction(gi.ActOpts{Label: "Non Def Params", Icon: "info", Tooltip: "shows all the parameters that are not at default values -- useful for setting params"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			nvv.ShowNonDefaultParams()
		})
	parsmen.Menu.AddAction(gi.ActOpts{Label: "All Params", Icon: "info", Tooltip: "shows all the parameters in the network"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			nvv.ShowAllParams()
		})
	ndmen := tbar.AddAction(gi.ActOpts{Label: "Net Data", Icon: "file-save"}, nil, nil)
	ndmen.Menu.AddAction(gi.ActOpts{Label: "Save Net Data", Icon: "file-save"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			giv.CallMethod(&nvv.Data, "SaveJSON", nvv.ViewportSafe()) // this auto prompts for filename using file chooser
		})
	ndmen.Menu.AddAction(gi.ActOpts{Label: "Open Net Data", Icon: "file-open"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			giv.CallMethod(&nvv.Data, "OpenJSON", nvv.ViewportSafe()) // this auto prompts for filename using file chooser
		})
	ndmen.Menu.AddSeparator("plotneur")
	rchk := gi.AddNewCheckBox(tbar, "raster")
	rchk.SetChecked(nv.Params.Raster.On)
	rchk.SetText("Raster")
	rchk.Tooltip = "Toggles raster plot mode -- displays values on one axis (Z by default) and raster counter (time) along the other (X by default)"
	rchk.ButtonSig.Connect(nv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		if sig == int64(gi.ButtonToggled) {
			cb := send.(*gi.CheckBox)
			nv.Params.Raster.On = cb.IsChecked()
			nv.ReconfigMeshes()
			nv.Update()
		}
	})
	xchk := gi.AddNewCheckBox(tbar, "raster-x")
	xchk.SetChecked(nv.Params.Raster.XAxis)
	xchk.SetText("X")
	xchk.Tooltip = "If checked, the raster (time) dimension is plotted along the X (horizontal) axis of the layers, otherwise it goes in the depth (Z) dimension"
	xchk.ButtonSig.Connect(nv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		if sig == int64(gi.ButtonToggled) {
			cb := send.(*gi.CheckBox)
			nv.Params.Raster.XAxis = cb.IsChecked()
			nv.Update()
		}
	})

	ndmen.Menu.AddAction(gi.ActOpts{Label: "Plot Selected Unit", Icon: "image", Tooltip: "opens up a window with a plot of all saved data for currently-selected unit"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			nvv.PlotSelectedUnit()
		})

	vp, ok := nv.VarParams[nv.Var]
	if !ok {
		vp = &VarParams{}
		vp.Defaults()
	}

	tbar.AddSeparator("cbar")
	mncb := gi.AddNewCheckBox(tbar, "mncb")
	mncb.Text = "Min"
	mncb.Tooltip = "Fix the minimum end of the displayed value range to value shown in next box.  Having both min and max fixed is recommended where possible for speed and consistent interpretability of the colors."
	mncb.SetChecked(vp.Range.FixMin)
	mncb.ButtonSig.Connect(nv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		if sig == int64(gi.ButtonToggled) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			vpp, ok := nvv.VarParams[nvv.Var]
			if ok {
				cbb := send.(*gi.CheckBox)
				vpp.Range.FixMin = cbb.IsChecked()
				nvv.Update()
				nvv.VarScaleUpdate(nvv.Var)
			}
		}
	})
	mnsb := gi.AddNewSpinBox(tbar, "mnsb")
	mnsb.SetValue(float32(vp.Range.Min))
	mnsb.SpinBoxSig.Connect(nv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		nvv := recv.Embed(KiT_NetView).(*NetView)
		vpp, ok := nvv.VarParams[nvv.Var]
		if ok {
			sbb := send.(*gi.SpinBox)
			vpp.Range.SetMin(sbb.Value)
			if vpp.ZeroCtr && vpp.Range.Min < 0 && vpp.Range.FixMax {
				vpp.Range.SetMax(-vpp.Range.Min)
			}
			nvv.VarScaleUpdate(nvv.Var)
			nvv.Update()
		}
	})

	cmap := giv.AddNewColorMapView(tbar, "cmap", nv.ColorMap)
	cmap.SetProp("min-width", units.NewEm(4))
	cmap.SetStretchMaxHeight()
	cmap.SetStretchMaxWidth()
	cmap.Tooltip = "Color map for translating values into colors -- click to select alternative."
	cmap.ColorMapSig.Connect(nv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		nvv := recv.Embed(KiT_NetView).(*NetView)
		cmm := send.(*giv.ColorMapView)
		if cmm.Map != nil {
			nvv.Params.ColorMap = giv.ColorMapName(cmm.Map.Name)
			nvv.ColorMap = cmm.Map
			nvv.Update()
		}
	})

	mxcb := gi.AddNewCheckBox(tbar, "mxcb")
	mxcb.SetChecked(vp.Range.FixMax)
	mxcb.Text = "Max"
	mxcb.Tooltip = "Fix the maximum end of the displayed value range to value shown in next box.  Having both min and max fixed is recommended where possible for speed and consistent interpretability of the colors."
	mxcb.ButtonSig.Connect(nv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		if sig == int64(gi.ButtonToggled) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			vpp, ok := nvv.VarParams[nvv.Var]
			if ok {
				cbb := send.(*gi.CheckBox)
				vpp.Range.FixMax = cbb.IsChecked()
				nvv.Update()
				nvv.VarScaleUpdate(nvv.Var)
			}
		}
	})
	mxsb := gi.AddNewSpinBox(tbar, "mxsb")
	mxsb.SetValue(float32(vp.Range.Max))
	mxsb.SpinBoxSig.Connect(nv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		nvv := recv.Embed(KiT_NetView).(*NetView)
		vpp, ok := nvv.VarParams[nvv.Var]
		if ok {
			sbb := send.(*gi.SpinBox)
			vpp.Range.SetMax(sbb.Value)
			if vpp.ZeroCtr && vpp.Range.Max > 0 && vpp.Range.FixMin {
				vpp.Range.SetMin(-vpp.Range.Max)
			}
			nvv.Update()
			nvv.VarScaleUpdate(nvv.Var)
		}
	})
	zccb := gi.AddNewCheckBox(tbar, "zccb")
	zccb.SetChecked(vp.ZeroCtr)
	zccb.Text = "ZeroCtr"
	zccb.Tooltip = "keep Min - Max centered around 0, and use negative heights for units -- else use full min-max range for height (no negative heights)"
	zccb.ButtonSig.Connect(nv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		if sig == int64(gi.ButtonToggled) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			vpp, ok := nvv.VarParams[nvv.Var]
			if ok {
				cbb := send.(*gi.CheckBox)
				vpp.ZeroCtr = cbb.IsChecked()
				nvv.Update()
				nvv.VarScaleUpdate(nvv.Var)
			}
		}
	})
}

func (nv *NetView) ViewbarConfig() {
	tbar := nv.Viewbar()
	if len(tbar.Kids) != 0 {
		return
	}
	tbar.SetStretchMaxWidth()
	tbar.AddAction(gi.ActOpts{Icon: "update", Tooltip: "reset to default initial display"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			nvv.Scene().SetCamera("default")
			nvv.Scene().UpdateSig()
		})
	tbar.AddAction(gi.ActOpts{Icon: "zoom-in", Tooltip: "zoom in"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			nvv.Scene().Camera.Zoom(-.05)
			nvv.Scene().UpdateSig()
		})
	tbar.AddAction(gi.ActOpts{Icon: "zoom-out", Tooltip: "zoom out"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			nvv.Scene().Camera.Zoom(.05)
			nvv.Scene().UpdateSig()
		})
	tbar.AddSeparator("rot")
	gi.AddNewLabel(tbar, "rot", "Rot:")
	tbar.AddAction(gi.ActOpts{Icon: "wedge-left"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			nvv.Scene().Camera.Orbit(5, 0)
			nvv.Scene().UpdateSig()
		})
	tbar.AddAction(gi.ActOpts{Icon: "wedge-up"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			nvv.Scene().Camera.Orbit(0, 5)
			nvv.Scene().UpdateSig()
		})
	tbar.AddAction(gi.ActOpts{Icon: "wedge-down"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			nvv.Scene().Camera.Orbit(0, -5)
			nvv.Scene().UpdateSig()
		})
	tbar.AddAction(gi.ActOpts{Icon: "wedge-right"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			nvv.Scene().Camera.Orbit(-5, 0)
			nvv.Scene().UpdateSig()
		})
	tbar.AddSeparator("pan")
	gi.AddNewLabel(tbar, "pan", "Pan:")
	tbar.AddAction(gi.ActOpts{Icon: "wedge-left"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			nvv.Scene().Camera.Pan(-.2, 0)
			nvv.Scene().UpdateSig()
		})
	tbar.AddAction(gi.ActOpts{Icon: "wedge-up"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			nvv.Scene().Camera.Pan(0, .2)
			nvv.Scene().UpdateSig()
		})
	tbar.AddAction(gi.ActOpts{Icon: "wedge-down"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			nvv.Scene().Camera.Pan(0, -.2)
			nvv.Scene().UpdateSig()
		})
	tbar.AddAction(gi.ActOpts{Icon: "wedge-right"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			nvv.Scene().Camera.Pan(.2, 0)
			nvv.Scene().UpdateSig()
		})
	tbar.AddSeparator("save")
	gi.AddNewLabel(tbar, "save", "Save:")
	tbar.AddAction(gi.ActOpts{Label: "1", Icon: "save", Tooltip: "first click (or + Shift) saves current view, second click restores to saved state"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			em := nvv.EventMgr2D()
			scc := nvv.Scene()
			cam := "1"
			if key.HasAllModifierBits(em.LastModBits, key.Shift) {
				scc.SaveCamera(cam)
			} else {
				err := scc.SetCamera(cam)
				if err != nil {
					scc.SaveCamera(cam)
				}
			}
			fmt.Printf("Camera %s: %v\n", cam, scc.Camera.GenGoSet(""))
			scc.UpdateSig()
		})
	tbar.AddAction(gi.ActOpts{Label: "2", Icon: "save", Tooltip: "first click (or + Shift) saves current view, second click restores to saved state"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			em := nvv.EventMgr2D()
			scc := nvv.Scene()
			cam := "2"
			if key.HasAllModifierBits(em.LastModBits, key.Shift) {
				scc.SaveCamera(cam)
			} else {
				err := scc.SetCamera(cam)
				if err != nil {
					scc.SaveCamera(cam)
				}
			}
			fmt.Printf("Camera %s: %v\n", cam, scc.Camera.GenGoSet(""))
			scc.UpdateSig()
		})
	tbar.AddAction(gi.ActOpts{Label: "3", Icon: "save", Tooltip: "first click (or + Shift) saves current view, second click restores to saved state"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			em := nvv.EventMgr2D()
			scc := nvv.Scene()
			cam := "3"
			if key.HasAllModifierBits(em.LastModBits, key.Shift) {
				scc.SaveCamera(cam)
			} else {
				err := scc.SetCamera(cam)
				if err != nil {
					scc.SaveCamera(cam)
				}
			}
			fmt.Printf("Camera %s: %v\n", cam, scc.Camera.GenGoSet(""))
			scc.UpdateSig()
		})
	tbar.AddAction(gi.ActOpts{Label: "4", Icon: "save", Tooltip: "first click (or + Shift) saves current view, second click restores to saved state"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			em := nvv.EventMgr2D()
			scc := nvv.Scene()
			cam := "4"
			if key.HasAllModifierBits(em.LastModBits, key.Shift) {
				scc.SaveCamera(cam)
			} else {
				err := scc.SetCamera(cam)
				if err != nil {
					scc.SaveCamera(cam)
				}
			}
			fmt.Printf("Camera %s: %v\n", cam, scc.Camera.GenGoSet(""))
			scc.UpdateSig()
		})
	tbar.AddSeparator("time")
	tlbl := gi.AddNewLabel(tbar, "time", "Time:")
	tlbl.Tooltip = "states are recorded over time -- last N can be reviewed using these buttons"
	rlbl := gi.AddNewLabel(tbar, "rec", fmt.Sprintf("%4d ", nv.RecNo))
	rlbl.Redrawable = true
	rlbl.Tooltip = "current view record: -1 means latest, 0 = earliest"
	tbar.AddAction(gi.ActOpts{Icon: "fast-bkwd", Tooltip: "move to first record (start of history)"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			if nvv.RecFullBkwd() {
				nvv.Update()
			}
		})
	tbar.AddAction(gi.ActOpts{Icon: "backward", Tooltip: "move earlier by N records (default 10)"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			if nvv.RecFastBkwd() {
				nvv.Update()
			}
		})
	tbar.AddAction(gi.ActOpts{Icon: "step-bkwd", Tooltip: "move earlier by 1"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			if nvv.RecBkwd() {
				nvv.Update()
			}
		})
	tbar.AddAction(gi.ActOpts{Icon: "play", Tooltip: "move to latest and always display latest (-1)"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			if nvv.RecTrackLatest() {
				nvv.Update()
			}
		})
	tbar.AddAction(gi.ActOpts{Icon: "step-fwd", Tooltip: "move later by 1"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			if nvv.RecFwd() {
				nvv.Update()
			}
		})
	tbar.AddAction(gi.ActOpts{Icon: "forward", Tooltip: "move later by N (default 10)"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			if nvv.RecFastFwd() {
				nvv.Update()
			}
		})
	tbar.AddAction(gi.ActOpts{Icon: "fast-fwd", Tooltip: "move to end (current time, tracking latest updates)"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nvv := recv.Embed(KiT_NetView).(*NetView)
			if nvv.RecTrackLatest() {
				nvv.Update()
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
	giv.TextViewDialog(nv.ViewportSafe(), []byte(nds), giv.DlgOpts{Title: "Non Default Params"})
	return nds
}

// ShowAllParams shows a dialog of all the parameters in the network.
func (nv *NetView) ShowAllParams() string {
	nds := nv.Net.AllParams()
	giv.TextViewDialog(nv.ViewportSafe(), []byte(nds), giv.DlgOpts{Title: "All Params"})
	return nds
}

func (nv *NetView) Render2D() {
	if gist.RebuildDefaultStyles {
		vs := nv.Scene()
		if vs != nil {
			vs.BgColor = gi.Prefs.Colors.Background // reset in case user changes
		}
	}
	nv.Layout.Render2D()
}

var NetViewProps = ki.Props{
	"max-width":  -1,
	"max-height": -1,
	// "width":      units.NewEm(5), // this gives the entire plot the scrollbars
	// "height":     units.NewEm(5),
	"CallMethods": ki.PropSlice{
		{"SaveWeights", ki.Props{
			"desc": "save network weights to file",
			"icon": "file-save",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".wts,.wts.gz",
				}},
			},
		}},
		{"OpenWeights", ki.Props{
			"desc": "open network weights from file",
			"icon": "file-open",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".wts,.wts.gz",
				}},
			},
		}},
	},
}
