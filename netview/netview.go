// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package netview provides the NetView interactive 3D network viewer, implemented in the Cogent Core 3D framework.
*/
package netview

//go:generate core generate -add-types

import (
	"fmt"
	"image/color"
	"log"
	"reflect"
	"strings"
	"sync"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/colors/colormap"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/math32/minmax"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/abilities"
	"cogentcore.org/core/texteditor"
	"cogentcore.org/core/tree"
	"cogentcore.org/core/xyz"
	"github.com/emer/emergent/v2/emer"
)

// NetView is a Cogent Core Widget that provides a 3D network view using the Cogent Core gi3d
// 3D framework.
type NetView struct {
	core.Frame

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

	// color map value representing ColorMap
	ColorMapButton *core.ColorMapButton

	// record number to display -- use -1 to always track latest, otherwise in range
	RecNo int

	// last non-empty counters string provided -- re-used if no new one
	LastCtrs string

	// contains all the network data with history
	Data NetData

	// mutex on data access
	DataMu sync.RWMutex `view:"-" copier:"-" json:"-" xml:"-"`
}

func (nv *NetView) Init() {
	nv.Frame.Init()
	nv.Params.NetView = nv
	nv.Params.Defaults()
	nv.ColorMap = colormap.AvailableMaps[string(nv.Params.ColorMap)]
	nv.RecNo = -1
	nv.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Grow.Set(1, 1)
	})

	core.AddChildAt(nv, "tbar", func(w *core.Toolbar) {
		nv.ConfigToolbar(w)
	})
	core.AddChildAt(nv, "nframe", func(w *core.Frame) {
		w.Styler(func(s *styles.Style) {
			s.Direction = styles.Row
			s.Grow.Set(1, 1)
		})
		core.AddChildAt(w, "vars", func(w *core.Frame) {
			w.Styler(func(s *styles.Style) {
				s.Display = styles.Grid
				s.Columns = nv.Params.NVarCols
				s.Grow.Set(0, 1)
				s.Overflow.Y = styles.OverflowAuto
				s.Background = colors.C(colors.Scheme.SurfaceContainerLow)
			})
			w.Maker(nv.makeVars)
		})
		core.AddChildAt(w, "scene", func(w *Scene) {
			w.NetView = nv
			nv.ViewConfig()
		})
	})
	core.AddChildAt(nv, "counters", func(w *core.Text) {
		w.SetText("Counters: " + strings.Repeat(" ", 200))
	})
	core.AddChildAt(nv, "vbar", func(w *core.Toolbar) {
		nv.ConfigViewbar(w)
	})

	nv.Data.Init(nv.Net, nv.Params.MaxRecs, nv.Params.NoSynData, nv.Net.MaxParallelData())
	nv.ReconfigMeshes()
}

// SetNet sets the network to view and updates view
func (nv *NetView) SetNet(net emer.Network) {
	nv.Net = net
	nv.DataMu.Lock()
	nv.Data.Init(nv.Net, nv.Params.MaxRecs, nv.Params.NoSynData, nv.Net.MaxParallelData())
	nv.DataMu.Unlock()
	nv.Update()
}

// SetVar sets the variable to view and updates the display
func (nv *NetView) SetVar(vr string) {
	nv.DataMu.Lock()
	nv.Var = vr
	nv.VarsUpdate()
	nv.VarScaleUpdate(nv.Var)
	nv.DataMu.Unlock()
	nv.UpdateView()
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
	nv.Data.PathType = nv.Params.PathType
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

// GoUpdateView is the update call to make from another go routine
// it does the proper blocking to coordinate with GUI updates
// generated on the main GUI thread.
func (nv *NetView) GoUpdateView() {
	if !nv.IsVisible() || !nv.HasLayers() {
		return
	}
	sw := nv.SceneWidget()
	sc := sw.SceneXYZ()
	sw.Scene.AsyncLock()
	nv.UpdateImpl()
	sc.NeedsRender = true
	sw.Scene.AsyncUnlock()
}

// UpdateView updates the display based on last recorded state of network.
func (nv *NetView) UpdateView() {
	if !nv.IsVisible() || !nv.HasLayers() {
		return
	}
	sw := nv.SceneWidget()
	nv.UpdateImpl()
	sw.XYZ.NeedsRender = true
	sw.NeedsRender()
}

// Current records the current state of the network, including synaptic values,
// and updates the display.  Use this when switching to NetView tab after network
// has been running while viewing another tab, because the network state
// is typically not recored then.
func (nv *NetView) Current() { //types:add
	nv.Record("", -1)
	nv.RecordSyns()
	nv.UpdateView()
}

// UpdateImpl does the guts of updating -- backend for Update or GoUpdate
func (nv *NetView) UpdateImpl() {
	nv.DataMu.Lock()
	vp, ok := nv.VarParams[nv.Var]
	if !ok {
		nv.DataMu.Unlock()
		log.Printf("NetView: %v variable: %v not found\n", nv.Name, nv.Var)
		return
	}
	nv.CurVarParams = vp

	if !vp.Range.FixMin || !vp.Range.FixMax {
		needUpdate := false
		// need to autoscale
		min, max, ok := nv.Data.VarRange(nv.Var)
		if ok {
			vp.MinMax.Set(min, max)
			if !vp.Range.FixMin {
				nmin := float32(minmax.NiceRoundNumber(float64(min), true)) // true = below
				if vp.Range.Min != nmin {
					vp.Range.Min = nmin
					needUpdate = true
				}
			}
			if !vp.Range.FixMax {
				nmax := float32(minmax.NiceRoundNumber(float64(max), false)) // false = above
				if vp.Range.Max != nmax {
					vp.Range.Max = nmax
					needUpdate = true
				}
			}
			if vp.ZeroCtr && !vp.Range.FixMin && !vp.Range.FixMax {
				bmax := math32.Max(math32.Abs(vp.Range.Max), math32.Abs(vp.Range.Min))
				if !needUpdate {
					if vp.Range.Max != bmax || vp.Range.Min != -bmax {
						needUpdate = true
					}
				}
				vp.Range.Max = bmax
				vp.Range.Min = -bmax
			}
			if needUpdate {
				nv.VarScaleUpdate(nv.Var)
			}
		}
	}

	se := nv.SceneXYZ()
	// laysGp := se.ChildByName("Layers", 0).AsTree()
	// if laysGp == nil || laysGp.NumChildren() != nv.Net.NLayers() {
	// 	nv.ConfigNetView()
	// }
	nv.SetCounters(nv.Data.CounterRec(nv.RecNo))
	nv.UpdateRecNo()
	nv.DataMu.Unlock()
	se.UpdateMeshes()
}

/*
// ConfigNetView configures the overall view widget
func (nv *NetView) ConfigNetView() {
	tb := core.NewToolbar(nv)
	tb.SetName("tbar")
	nlay := core.NewFrame(nv)
	nlay.SetName("net")
	nlay.Styler(func(s *styles.Style) {
		s.Direction = styles.Row
		s.Grow.Set(1, 1)
	})
	core.NewText(nv).SetName("counters").SetText(strings.Repeat(" ", 200))
	vb := core.NewToolbar(nv, "vbar")

	vlay := core.NewFrame(nlay, "vars")
	vlay.Styler(func(s *styles.Style) {
		s.Display = styles.Grid
		s.Columns = nv.Params.NVarCols
		s.Grow.Set(0, 1)
		s.Overflow.Y = styles.OverflowAuto
		s.Background = colors.C(colors.Scheme.SurfaceContainerLow)
	})

	sw := NewScene(nlay, "scene")
	sw.NetView = nv

	nv.ConfigToolbar(tb)
	nv.ConfigViewbar(vb)

	nv.VarsConfig()
	nv.ViewConfig()

	ctrs := nv.Counters()
	ctrs.SetText("Counters: " + strings.Repeat("-", 200))

	nv.DataMu.Lock()
	nv.Data.Init(nv.Net, nv.Params.MaxRecs, nv.Params.NoSynData, nv.Net.MaxParallelData())
	nv.DataMu.Unlock()
	nv.ReconfigMeshes()
	nv.NeedsLayout()
}
*/

// ReconfigMeshes reconfigures the layer meshes
func (nv *NetView) ReconfigMeshes() {
	se := nv.SceneXYZ()
	// if se.IsConfiged() {
	se.ReconfigMeshes()
	// }
}

func (nv *NetView) Toolbar() *core.Toolbar {
	return nv.ChildByName("tbar", 0).(*core.Toolbar)
}

func (nv *NetView) NetLay() *core.Frame {
	return nv.ChildByName("net", 1).(*core.Frame)
}

func (nv *NetView) Counters() *core.Text {
	return nv.ChildByName("counters", 2).(*core.Text)
}

func (nv *NetView) Viewbar() *core.Toolbar {
	return nv.ChildByName("vbar", 3).(*core.Toolbar)
}

func (nv *NetView) SceneWidget() *Scene {
	return nv.NetLay().ChildByName("scene", 1).(*Scene)
}

func (nv *NetView) SceneXYZ() *xyz.Scene {
	return nv.SceneWidget().Scene.XYZ

}

func (nv *NetView) VarsLay() *core.Frame {
	return nv.NetLay().ChildByName("vars", 0).(*core.Frame)
}

// SetCounters sets the counters widget view display at bottom of netview
func (nv *NetView) SetCounters(ctrs string) {
	ct := nv.Counters()
	if ct.Text != ctrs {
		ct.SetText(ctrs).Update()
	}
}

// UpdateRecNo updates the record number viewing
func (nv *NetView) UpdateRecNo() {
	vbar := nv.Viewbar()
	rlbl := vbar.ChildByName("rec", 10).(*core.Text)
	rlbl.SetText(fmt.Sprintf("%4d ", nv.RecNo)).Update()
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

// NetVarsList returns the list of layer and path variables for given network.
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
	pathprops := nv.Net.SynVarProps()
	for _, nm := range nv.Vars {
		vp := &VarParams{Var: nm}
		vp.Defaults()
		var vtag string
		if strings.HasPrefix(nm, "r.") || strings.HasPrefix(nm, "s.") {
			vtag = pathprops[nm[2:]]
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
	for _, vbi := range vl.Children {
		vb := vbi.(*core.Button)
		vb.SetSelected(vb.Text == nv.Var)
	}
	nv.ColorMapButton.Update()
	vl.NeedsRender()
}

// VarScaleUpdate updates display of the scaling params
// for given variable (use nv.Var for current)
// returns true if any setting changed (update always triggered)
func (nv *NetView) VarScaleUpdate(varNm string) bool {
	vp := nv.VarParams[varNm]

	tb := nv.Toolbar()
	mod := false

	if ci := tb.ChildByName("mnsw", 4); ci != nil {
		sw := ci.(*core.Switch)
		if sw.IsChecked() != vp.Range.FixMin {
			mod = true
			sw.SetChecked(vp.Range.FixMin)
			sw.NeedsRender()
		}
	}
	if ci := tb.ChildByName("mxsw", 6); ci != nil {
		sw := ci.(*core.Switch)
		if sw.IsChecked() != vp.Range.FixMax {
			mod = true
			sw.SetChecked(vp.Range.FixMax)
			sw.NeedsRender()
		}
	}
	if ci := tb.ChildByName("mnsp", 5); ci != nil {
		sp := ci.(*core.Spinner)
		mnv := vp.Range.Min
		if sp.Value != mnv {
			mod = true
			sp.SetValue(mnv)
			sp.NeedsRender()
		}
	}
	if ci := tb.ChildByName("mxsp", 7); ci != nil {
		sp := ci.(*core.Spinner)
		mxv := vp.Range.Max
		if sp.Value != mxv {
			mod = true
			sp.SetValue(mxv)
			sp.NeedsRender()
		}
	}
	if ci := tb.ChildByName("zcsw", 8); ci != nil {
		sw := ci.(*core.Switch)
		if sw.IsChecked() != vp.ZeroCtr {
			mod = true
			sw.SetChecked(vp.ZeroCtr)
			sw.NeedsRender()
		}
	}
	return mod
}

// makeVars configures the variables
func (nv *NetView) makeVars(p *core.Plan) {
	nv.VarsListUpdate()
	unprops := nv.Net.UnitVarProps()
	pathprops := nv.Net.SynVarProps()
	for _, vn := range nv.Vars {
		core.AddAt(p, vn, func(w *core.Button) {
			w.SetText(vn).SetType(core.ButtonAction)
			pstr := ""
			if strings.HasPrefix(vn, "r.") || strings.HasPrefix(vn, "s.") {
				pstr = pathprops[vn[2:]]
			} else {
				pstr = unprops[vn]
			}
			if pstr != "" {
				rstr := reflect.StructTag(pstr)
				if desc, ok := rstr.Lookup("desc"); ok {
					w.Tooltip = vn + ": " + desc
				}
			}
			w.SetSelected(vn == nv.Var)
			w.OnClick(func(e events.Event) {
				nv.SetVar(vn)
			})
		})
	}
}

// ViewConfig configures the 3D view
func (nv *NetView) ViewConfig() {
	sw := nv.SceneWidget()

	se := sw.Scene.XYZ
	if nv.Net == nil || nv.Net.NLayers() == 0 {
		se.DeleteChildren()
		se.Meshes.Reset()
		return
	}
	if se.Lights.Len() == 0 {
		nv.ViewDefaults()
	}
	// todo:
	// vs.BgColor = core.Prefs.Colors.Background // reset in case user changes
	nlay := nv.Net.NLayers()
	laysGpi := se.ChildByName("Layers", 0)
	var laysGp *xyz.Group
	if laysGpi == nil {
		laysGp = xyz.NewGroup(se)
		laysGp.Name = "Layers"
	} else {
		laysGp = laysGpi.(*xyz.Group)
	}
	layConfig := tree.TypePlan{}
	for li := 0; li < nlay; li++ {
		lay := nv.Net.Layer(li)
		lmesh := se.MeshByName(lay.Name())
		if lmesh == nil {
			NewLayMesh(se, nv, lay)
		} else {
			lmesh.(*LayMesh).Lay = lay // make sure
		}
		layConfig.Add(xyz.GroupType, lay.Name())
	}
	gpConfig := tree.TypePlan{}
	gpConfig.Add(LayObjType, "layer")
	gpConfig.Add(LayNameType, "name")

	tree.Update(laysGp, layConfig)

	nmin, nmax := nv.Net.Bounds()
	nsz := nmax.Sub(nmin).Sub(math32.Vec3(1, 1, 0)).Max(math32.Vec3(1, 1, 1))
	nsc := math32.Vec3(1.0/nsz.X, 1.0/nsz.Y, 1.0/nsz.Z)
	szc := math32.Max(nsc.X, nsc.Y)
	poff := math32.Vector3Scalar(0.5)
	poff.Y = -0.5
	for li, lgi := range laysGp.Children {
		ly := nv.Net.Layer(li)
		lg := lgi.(*xyz.Group)
		tree.Update(lg, gpConfig)
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
		lo.Material.Color = colors.FromRGB(255, 100, 255)
		lo.Material.Reflective = 8
		lo.Material.Bright = 8
		lo.Material.Shiny = 30
		// note: would actually be better to NOT cull back so you can view underneath
		// but then the front and back fight against each other, causing flickering

		txt := lg.Child(1).(*LayName)
		txt.Name = "layname:" + ly.Name()
		txt.Defaults()
		txt.NetView = nv
		txt.SetText(ly.Name())
		txt.Pose.Scale = math32.Vector3Scalar(nv.Params.LayNmSize).Div(lg.Pose.Scale)
		txt.Styles.Background = colors.C(colors.Transparent)
		txt.Styles.Text.Align = styles.Start
		txt.Styles.Text.AlignV = styles.Start
	}
	sw.XYZ.NeedsConfig = true
	sw.NeedsRender()
}

// ViewDefaults are the default 3D view params
func (nv *NetView) ViewDefaults() {
	se := nv.SceneXYZ()
	se.Camera.Pose.Pos.Set(0, 1.5, 2.5) // more "top down" view shows more of layers
	// 	vs.Camera.Pose.Pos.Set(0, 1, 2.75) // more "head on" for larger / deeper networks
	se.Camera.Near = 0.1
	se.Camera.LookAt(math32.Vec3(0, 0, 0), math32.Vec3(0, 1, 0))
	// todo:
	// vs.BgColor = core.Prefs.Colors.Background
	xyz.NewAmbientLight(se, "ambient", 0.1, xyz.DirectSun)
	dir := xyz.NewDirLight(se, "dirUp", 0.3, xyz.DirectSun)
	dir.Pos.Set(0, 1, 0)
	dir = xyz.NewDirLight(se, "dirBack", 0.3, xyz.DirectSun)
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
func (nv *NetView) UnitValue(lay emer.Layer, idx []int) (raw, scaled float32, clr color.RGBA, hasval bool) {
	idx1d := lay.Shape().Offset(idx)
	if idx1d >= lay.Shape().Len() {
		raw, hasval = 0, false
	} else {
		raw, hasval = nv.Data.UnitValue(lay.Name(), nv.Var, idx1d, nv.RecNo, nv.Di)
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
	ridx := lay.RepIndexes()
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
		if lay.Name() == nv.Data.PathLay && idx1d == nv.Data.PathUnIndex {
			clr = color.RGBA{0x20, 0x80, 0x20, 0x80}
		} else {
			clr = NilColor
		}
	} else {
		clp := nv.CurVarParams.Range.ClipValue(raw)
		norm := nv.CurVarParams.Range.NormValue(clp)
		var op float32
		if nv.CurVarParams.ZeroCtr {
			scaled = float32(2*norm - 1)
			op = (nv.Params.ZeroAlpha + (1-nv.Params.ZeroAlpha)*math32.Abs(scaled))
		} else {
			scaled = float32(norm)
			op = (nv.Params.ZeroAlpha + (1-nv.Params.ZeroAlpha)*0.8) // no meaningful alpha -- just set at 80\%
		}
		clr = colors.WithAF32(nv.ColorMap.Map(norm), op)
	}
	return
}

func (nv *NetView) Labels() *xyz.Group {
	se := nv.SceneXYZ()
	lgpi := se.ChildByName("Labels", 1)
	if lgpi == nil {
		return nil
	}
	return lgpi.(*xyz.Group)
}

// ConfigLabels ensures that given label xyz.Text2D objects are created and initialized
// in a top-level group called Labels.  Use LabelByName() to get a given label, and
// LayerByName() to get a Layer group, whose Pose can be copied to put a label in
// position relative to a layer.  Default alignment is Left, Top.
// Returns true set of labels was changed (mods).
func (nv *NetView) ConfigLabels(labs []string) bool {
	se := nv.SceneXYZ()
	lgp := nv.Labels()
	if lgp == nil {
		lgp = xyz.NewGroup(se)
		lgp.Name = "Labels"
	}

	lbConfig := tree.TypePlan{}
	for _, ls := range labs {
		lbConfig.Add(xyz.Text2DType, ls)
	}
	if tree.Update(lgp, lbConfig) {
		for i, ls := range labs {
			lb := lgp.ChildByName(ls, i).(*xyz.Text2D)
			// lb.Defaults()
			lb.SetText(ls)
			// todo:
			// lb.SetProperty("text-align", styles.Start)
			// lb.SetProperty("vertical-align", styles.Start)
			// lb.SetProperty("white-space", styles.WhiteSpacePre)
		}
		return true
	}
	return false
}

// LabelByName returns given Text2D label (see ConfigLabels).
// nil if not found.
func (nv *NetView) LabelByName(lab string) *xyz.Text2D {
	lgp := nv.Labels()
	txt := lgp.ChildByName(lab, 0)
	if txt == nil {
		return nil
	}
	return txt.(*xyz.Text2D)
}

// LayerByName returns the xyz.Group that represents layer of given name.
// nil if not found.
func (nv *NetView) LayerByName(lay string) *xyz.Group {
	lgp := nv.Labels()
	ly := lgp.ChildByName(lay, 0)
	if ly == nil {
		return nil
	}
	return ly.(*xyz.Group)
}

func (nv *NetView) ConfigToolbar(tb *core.Toolbar) {
	core.NewFuncButton(tb, nv.Update).SetText("Init").SetIcon(icons.Update)
	core.NewFuncButton(tb, nv.Current).SetIcon(icons.Update)
	core.NewButton(tb).SetText("Config").SetIcon(icons.Settings).
		SetTooltip("set parameters that control display (font size etc)").
		OnClick(func(e events.Event) {
			core.FormDialog(nv, &nv.Params, nv.Name+" Params", true)
		})
	core.NewSeparator(tb)
	core.NewButton(tb).SetText("Weights").SetType(core.ButtonAction).SetMenu(func(m *core.Scene) {
		fb := core.NewFuncButton(m, nv.SaveWeights)
		fb.SetIcon(icons.Save)
		fb.Args[0].SetTag(`"ext:".wts,.wts.gz"`)
		fb = core.NewFuncButton(m, nv.OpenWeights)
		fb.SetIcon(icons.Open)
		fb.Args[0].SetTag(`"ext:".wts,.wts.gz"`)
	})
	core.NewButton(tb).SetText("Params").SetIcon(icons.Info).SetMenu(func(m *core.Scene) {
		core.NewFuncButton(m, nv.ShowNonDefaultParams).SetIcon(icons.Info)
		core.NewFuncButton(m, nv.ShowAllParams).SetIcon(icons.Info)
		core.NewFuncButton(m, nv.ShowKeyLayerParams).SetIcon(icons.Info)
		core.NewFuncButton(m, nv.ShowKeyPathParams).SetIcon(icons.Info)
	})
	core.NewButton(tb).SetText("Net Data").SetIcon(icons.Save).SetMenu(func(m *core.Scene) {
		core.NewFuncButton(m, nv.Data.SaveJSON).SetText("Save Net Data").SetIcon(icons.Save)
		core.NewFuncButton(m, nv.Data.OpenJSON).SetText("Open Net Data").SetIcon(icons.Open)
		core.NewSeparator(m)
		core.NewFuncButton(m, nv.PlotSelectedUnit).SetIcon(icons.Open)
	})
	core.NewSeparator(tb)
	ditp := "data parallel index -- for models running multiple input patterns in parallel, this selects which one is viewed"
	core.NewText(tb).SetText("Di:").SetTooltip(ditp)
	dis := core.NewSpinner(tb)
	dis.SetMin(0).SetStep(1).SetValue(float32(nv.Di)).SetTooltip(ditp)
	dis.OnChange(func(e events.Event) {
		maxData := nv.Net.MaxParallelData()
		md := int(dis.Value)
		if md < maxData && md >= 0 {
			nv.Di = md
		}
		dis.SetValue(float32(nv.Di))
		nv.UpdateView()
	})
	core.NewSeparator(tb)
	rchk := core.NewSwitch(tb)
	rchk.SetText("Raster").SetChecked(nv.Params.Raster.On).
		SetTooltip("Toggles raster plot mode -- displays values on one axis (Z by default) and raster counter (time) along the other (X by default)")
	rchk.OnChange(func(e events.Event) {
		nv.Params.Raster.On = rchk.IsChecked()
		nv.ReconfigMeshes()
		nv.UpdateView()
	})
	xchk := core.NewSwitch(tb)
	xchk.SetText("X").SetType(core.SwitchCheckbox).SetChecked(nv.Params.Raster.XAxis).
		SetTooltip("If checked, the raster (time) dimension is plotted along the X (horizontal) axis of the layers, otherwise it goes in the depth (Z) dimension").
		OnChange(func(e events.Event) {
			nv.Params.Raster.XAxis = xchk.IsChecked()
			nv.UpdateView()
		})

	vp, ok := nv.VarParams[nv.Var]
	if !ok {
		vp = &VarParams{}
		vp.Defaults()
	}

	core.NewSeparator(tb)
	mnsw := core.NewSwitch(tb)
	mnsw.Name = "mnsw"
	mnsw.SetText("Min").SetType(core.SwitchCheckbox).SetChecked(vp.Range.FixMin).
		SetTooltip("Fix the minimum end of the displayed value range to value shown in next box.  Having both min and max fixed is recommended where possible for speed and consistent interpretability of the colors.").OnChange(func(e events.Event) {
		vp := nv.VarParams[nv.Var]
		vp.Range.FixMin = mnsw.IsChecked()
		nv.VarScaleUpdate(nv.Var) // todo: before update?
		nv.UpdateView()
	})
	mnsp := core.NewSpinner(tb)
	mnsp.Name = "mnsp"
	mnsp.SetValue(vp.Range.Min).
		OnChange(func(e events.Event) {
			vp := nv.VarParams[nv.Var]
			vp.Range.SetMin(mnsp.Value)
			if vp.ZeroCtr && vp.Range.Min < 0 && vp.Range.FixMax {
				vp.Range.SetMax(-vp.Range.Min)
			}
			nv.VarScaleUpdate(nv.Var)
			nv.UpdateView()
		})

	cmb := core.NewColorMapButton(tb)
	nv.ColorMapButton = cmb
	cmb.MapName = string(nv.Params.ColorMap)
	cmb.Name = "cmap"
	cmb.SetTooltip("Color map for translating values into colors -- click to select alternative.").
		Styler(func(s *styles.Style) {
			s.Min.X.Em(10)
			s.Min.Y.Em(1.2)
			s.Grow.Set(0, 1)
		}).OnChange(func(e events.Event) {
		cmap, ok := colormap.AvailableMaps[string(nv.ColorMapButton.MapName)]
		if ok {
			nv.ColorMap = cmap
		}
		nv.UpdateView()
	})

	mxsw := core.NewSwitch(tb)
	mxsw.Name = "mxsw"
	mxsw.SetText("Max").SetType(core.SwitchCheckbox).SetChecked(vp.Range.FixMax).
		SetTooltip("Fix the maximum end of the displayed value range to value shown in next box.  Having both min and max fixed is recommended where possible for speed and consistent interpretability of the colors.").
		OnChange(func(e events.Event) {
			vp := nv.VarParams[nv.Var]
			vp.Range.FixMax = mxsw.IsChecked()
			nv.VarScaleUpdate(nv.Var)
			nv.UpdateView()
		})
	mxsp := core.NewSpinner(tb)
	mxsp.Name = "mxsp"
	mxsp.SetValue(vp.Range.Max).OnChange(func(e events.Event) {
		vp := nv.VarParams[nv.Var]
		vp.Range.SetMax(mxsp.Value)
		if vp.ZeroCtr && vp.Range.Max > 0 && vp.Range.FixMin {
			vp.Range.SetMin(-vp.Range.Max)
		}
		nv.VarScaleUpdate(nv.Var)
		nv.UpdateView()
	})
	zcsw := core.NewSwitch(tb)
	zcsw.Name = "zcsw"
	zcsw.SetText("ZeroCtr").SetChecked(vp.ZeroCtr).
		SetTooltip("keep Min - Max centered around 0, and use negative heights for units -- else use full min-max range for height (no negative heights)").
		OnChange(func(e events.Event) {
			vp := nv.VarParams[nv.Var]
			vp.ZeroCtr = zcsw.IsChecked()
			nv.VarScaleUpdate(nv.Var)
			nv.UpdateView()
		})
}

func (nv *NetView) ConfigViewbar(tb *core.Toolbar) {
	core.NewButton(tb).SetIcon(icons.Update).SetTooltip("reset to default initial display").
		OnClick(func(e events.Event) {
			nv.SceneXYZ().SetCamera("default")
			nv.UpdateView()
		})
	core.NewButton(tb).SetIcon(icons.ZoomIn).SetTooltip("zoom in").
		Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		}).
		OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Zoom(-.05)
			nv.UpdateView()
		})
	core.NewButton(tb).SetIcon(icons.ZoomOut).SetTooltip("zoom out").
		Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		}).
		OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Zoom(.05)
			nv.UpdateView()
		})
	core.NewSeparator(tb)
	core.NewText(tb).SetText("Rot:").SetTooltip("rotate display")
	core.NewButton(tb).SetIcon(icons.KeyboardArrowLeft).
		Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		}).
		OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Orbit(5, 0)
			nv.UpdateView()
		})
	core.NewButton(tb).SetIcon(icons.KeyboardArrowUp).
		Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		}).
		OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Orbit(0, 5)
			nv.UpdateView()
		})
	core.NewButton(tb).SetIcon(icons.KeyboardArrowDown).
		Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		}).
		OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Orbit(0, -5)
			nv.UpdateView()
		})
	core.NewButton(tb).SetIcon(icons.KeyboardArrowRight).
		Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		}).
		OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Orbit(-5, 0)
			nv.UpdateView()
		})
	core.NewSeparator(tb)

	core.NewText(tb).SetText("Pan:").SetTooltip("pan display")
	core.NewButton(tb).SetIcon(icons.KeyboardArrowLeft).
		Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		}).
		OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Pan(-.2, 0)
			nv.UpdateView()
		})
	core.NewButton(tb).SetIcon(icons.KeyboardArrowUp).
		Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		}).
		OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Pan(0, .2)
			nv.UpdateView()
		})
	core.NewButton(tb).SetIcon(icons.KeyboardArrowDown).
		Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		}).
		OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Pan(0, -.2)
			nv.UpdateView()
		})
	core.NewButton(tb).SetIcon(icons.KeyboardArrowRight).
		Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		}).
		OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Pan(.2, 0)
			nv.UpdateView()
		})
	core.NewSeparator(tb)

	core.NewText(tb).SetText("Save:")
	for i := 1; i <= 4; i++ {
		i := i
		nm := fmt.Sprintf("%d", i)
		core.NewButton(tb).SetText(nm).
			SetTooltip("first click (or + Shift) saves current view, second click restores to saved state").
			OnClick(func(e events.Event) {
				sc := nv.SceneXYZ()
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
	core.NewSeparator(tb)

	core.NewText(tb).SetText("Time:").
		SetTooltip("states are recorded over time -- last N can be reviewed using these buttons")

	rec := core.NewText(tb).SetText(fmt.Sprintf("%4d ", nv.RecNo))
	rec.Name = "rec"
	rec.SetTooltip("current view record: -1 means latest, 0 = earliest")
	core.NewButton(tb).SetIcon(icons.FirstPage).SetTooltip("move to first record (start of history)").
		OnClick(func(e events.Event) {
			if nv.RecFullBkwd() {
				nv.UpdateView()
			}
		})
	core.NewButton(tb).SetIcon(icons.FastRewind).SetTooltip("move earlier by N records (default 10)").
		Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		}).
		OnClick(func(e events.Event) {
			if nv.RecFastBkwd() {
				nv.UpdateView()
			}
		})
	core.NewButton(tb).SetIcon(icons.SkipPrevious).SetTooltip("move earlier by 1").
		Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		}).
		OnClick(func(e events.Event) {
			if nv.RecBkwd() {
				nv.UpdateView()
			}
		})
	core.NewButton(tb).SetIcon(icons.PlayArrow).SetTooltip("move to latest and always display latest (-1)").
		OnClick(func(e events.Event) {
			if nv.RecTrackLatest() {
				nv.UpdateView()
			}
		})
	core.NewButton(tb).SetIcon(icons.SkipNext).SetTooltip("move later by 1").
		Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		}).
		OnClick(func(e events.Event) {
			if nv.RecFwd() {
				nv.UpdateView()
			}
		})
	core.NewButton(tb).SetIcon(icons.FastForward).SetTooltip("move later by N (default 10)").
		Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		}).
		OnClick(func(e events.Event) {
			if nv.RecFastFwd() {
				nv.UpdateView()
			}
		})
	core.NewButton(tb).SetIcon(icons.LastPage).SetTooltip("move to end (current time, tracking latest updates)").
		OnClick(func(e events.Event) {
			if nv.RecTrackLatest() {
				nv.UpdateView()
			}
		})
}

// SaveWeights saves the network weights.
func (nv *NetView) SaveWeights(filename core.Filename) { //types:add
	nv.Net.SaveWtsJSON(filename)
}

// OpenWeights opens the network weights.
func (nv *NetView) OpenWeights(filename core.Filename) { //types:add
	nv.Net.OpenWtsJSON(filename)
}

// ShowNonDefaultParams shows a dialog of all the parameters that
// are not at their default values in the network.  Useful for setting params.
func (nv *NetView) ShowNonDefaultParams() string { //types:add
	nds := nv.Net.NonDefaultParams()
	texteditor.TextDialog(nv, "Non Default Params: "+nv.Name, nds)
	return nds
}

// ShowAllParams shows a dialog of all the parameters in the network.
func (nv *NetView) ShowAllParams() string { //types:add
	nds := nv.Net.AllParams()
	texteditor.TextDialog(nv, "All Params: "+nv.Name, nds)
	return nds
}

// ShowKeyLayerParams shows a dialog with a listing for all layers in the network,
// of the most important layer-level params (specific to each algorithm)
func (nv *NetView) ShowKeyLayerParams() string { //types:add
	nds := nv.Net.KeyLayerParams()
	texteditor.TextDialog(nv, "Key Layer Params: "+nv.Name, nds)
	return nds
}

// ShowKeyPathParams shows a dialog with a listing for all Recv pathways in the network,
// of the most important pathway-level params (specific to each algorithm)
func (nv *NetView) ShowKeyPathParams() string { //types:add
	nds := nv.Net.KeyPathParams()
	texteditor.TextDialog(nv, "Key Path Params: "+nv.Name, nds)
	return nds
}
