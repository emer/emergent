// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package netview provides the NetView interactive 3D network viewer,
implemented in the Cogent Core 3D framework.
*/
package netview

//go:generate core generate -add-types

import (
	"fmt"
	"image/color"
	"log"
	"log/slog"
	"reflect"
	"strings"
	"sync"
	"time"

	"cogentcore.org/core/base/errors"
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
	"cogentcore.org/core/system"
	"cogentcore.org/core/texteditor"
	"cogentcore.org/core/tree"
	"cogentcore.org/core/types"
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
	CurVarParams *VarParams `json:"-" xml:"-" display:"-"`

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

	// current counters
	CurCtrs string

	// contains all the network data with history
	Data NetData

	// mutex on data access
	DataMu sync.RWMutex `display:"-" copier:"-" json:"-" xml:"-"`
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

	tree.AddChildAt(nv, "tbar", func(w *core.Toolbar) {
		w.Maker(nv.MakeToolbar)
	})
	tree.AddChildAt(nv, "netframe", func(w *core.Frame) {
		w.Styler(func(s *styles.Style) {
			s.Direction = styles.Row
			s.Grow.Set(1, 1)
		})
		nv.makeVars(w)
		tree.AddChildAt(w, "scene", func(w *Scene) {
			w.NetView = nv
			se := w.SceneXYZ()
			nv.ViewDefaults(se)
			laysGp := xyz.NewGroup(se)
			laysGp.Name = "Layers"
		})
	})
	tree.AddChildAt(nv, "counters", func(w *core.Text) {
		w.SetText("Counters: " + strings.Repeat(" ", 200)).
			Styler(func(s *styles.Style) {
				s.Grow.Set(1, 0)
			})
		w.Updater(func() {
			if w.Text != nv.CurCtrs && nv.CurCtrs != "" {
				w.SetText(nv.CurCtrs)
			}
		})
	})
	tree.AddChildAt(nv, "vbar", func(w *core.Toolbar) {
		w.Maker(nv.MakeViewbar)
	})
}

// SetNet sets the network to view and updates view
func (nv *NetView) SetNet(net emer.Network) {
	nv.Net = net
	nv.DataMu.Lock()
	nv.Data.Init(nv.Net, nv.Params.MaxRecs, nv.Params.NoSynData, nv.Net.MaxParallelData())
	nv.DataMu.Unlock()
	nv.UpdateTree() // need children
	nv.UpdateLayers()
	nv.Current()
}

// SetVar sets the variable to view and updates the display
func (nv *NetView) SetVar(vr string) {
	nv.DataMu.Lock()
	nv.Var = vr
	nv.VarsFrame().Update()
	nv.DataMu.Unlock()
	nv.Toolbar().Update()
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
	if nv.Net == nil || nv.Net.NumLayers() == 0 {
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
	se := sw.SceneXYZ()
	sw.Scene.AsyncLock()
	nv.UpdateImpl()
	se.SetNeedsRender()
	sw.NeedsRender()
	sw.Scene.AsyncUnlock()
	if core.TheApp.Platform() == system.Web {
		time.Sleep(time.Millisecond) // critical to prevent hanging!
	}
}

// UpdateView updates the display based on last recorded state of network.
func (nv *NetView) UpdateView() {
	if !nv.IsVisible() || !nv.HasLayers() {
		return
	}
	sw := nv.SceneWidget()
	nv.UpdateImpl()
	sw.XYZ.SetNeedsRender()
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
				tb := nv.Toolbar()
				tb.UpdateTree()
				tb.NeedsRender()
			}
		}
	}

	nv.SetCounters(nv.Data.CounterRec(nv.RecNo))
	nv.UpdateRecNo()
	nv.DataMu.Unlock()
	nv.UpdateLayers()
}

// // ReconfigMeshes reconfigures the layer meshes
// func (nv *NetView) ReconfigMeshes() {
// 	se := nv.SceneXYZ()
// 	se.ReconfigMeshes()
// }

func (nv *NetView) Toolbar() *core.Toolbar {
	return nv.ChildByName("tbar", 0).(*core.Toolbar)
}

func (nv *NetView) NetFrame() *core.Frame {
	return nv.ChildByName("netframe", 1).(*core.Frame)
}

func (nv *NetView) Counters() *core.Text {
	return nv.ChildByName("counters", 2).(*core.Text)
}

func (nv *NetView) Viewbar() *core.Toolbar {
	return nv.ChildByName("vbar", 3).(*core.Toolbar)
}

func (nv *NetView) SceneWidget() *Scene {
	return nv.NetFrame().ChildByName("scene", 1).(*Scene)
}

func (nv *NetView) SceneXYZ() *xyz.Scene {
	return nv.SceneWidget().SceneXYZ()
}

func (nv *NetView) VarsFrame() *core.Tabs {
	return nv.NetFrame().ChildByName("vars", 0).(*core.Tabs)
}

// SetCounters sets the counters widget view display at bottom of netview
func (nv *NetView) SetCounters(ctrs string) {
	if ctrs == "" {
		return
	}
	nv.CurCtrs = ctrs
	ct := nv.Counters()
	ct.UpdateWidget().NeedsRender()
}

// UpdateRecNo updates the record number viewing
func (nv *NetView) UpdateRecNo() {
	vbar := nv.Viewbar()
	rlbl := vbar.ChildByName("rec", 10)
	if rlbl != nil {
		rlbl.(*core.Text).UpdateWidget().NeedsRender()
	}
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
	if net == nil || net.NumLayers() == 0 {
		return nil, nil
	}
	unvars := net.UnitVarNames()
	synvars = net.SynVarNames()
	ulen := len(unvars)
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

// makeVars configures the variables
func (nv *NetView) makeVars(netframe *core.Frame) {
	nv.VarsListUpdate()
	if nv.Net == nil {
		return
	}
	unprops := nv.Net.UnitVarProps()
	pathprops := nv.Net.SynVarProps()
	cats := nv.Net.VarCategories()
	if len(cats) == 0 {
		cats = []emer.VarCategory{
			{"Unit", "unit variables"},
			{"Wt", "connection weight variables"},
		}
	}
	tree.AddChildAt(netframe, "vars", func(w *core.Tabs) {
		w.Styler(func(s *styles.Style) {
			s.Grow.Set(0, 1)
			s.Overflow.Y = styles.OverflowAuto
		})
		tabs := make(map[string]*core.Frame)
		for _, ct := range cats {
			tf, tb := w.NewTab(ct.Cat)
			tb.Tooltip = ct.Desc
			tabs[ct.Cat] = tf
			tf.Styler(func(s *styles.Style) {
				s.Display = styles.Grid
				s.Columns = nv.Params.NVarCols
				s.Grow.Set(1, 1)
				s.Overflow.Y = styles.OverflowAuto
				s.Background = colors.Scheme.SurfaceContainerLow
			})
		}
		for _, vn := range nv.Vars {
			cat := ""
			pstr := ""
			desc := ""
			if strings.HasPrefix(vn, "r.") || strings.HasPrefix(vn, "s.") {
				pstr = pathprops[vn[2:]]
				cat = "Wt" // default
			} else {
				pstr = unprops[vn]
				cat = "Unit"
			}
			if pstr != "" {
				rstr := reflect.StructTag(pstr)
				desc = rstr.Get("desc")
				cat = rstr.Get("cat")
				if rstr.Get("display") == "-" {
					continue
				}
			}
			tf, ok := tabs[cat]
			if !ok {
				slog.Error("emergent.NetView UnitVarProps 'cat' name not found in VarCategories list", "cat", cat, "variable", vn)
				cat = cats[0].Cat
				tf = tabs[cat]
			}
			w := core.NewButton(tf).SetText(vn)
			if desc != "" {
				w.Tooltip = vn + ": " + desc
			}
			w.SetText(vn).SetType(core.ButtonAction)
			w.OnClick(func(e events.Event) {
				nv.SetVar(vn)
			})
			w.Updater(func() {
				w.SetSelected(w.Text == nv.Var)
			})
		}
	})
}

// UpdateLayers updates the layer display with any structural or
// current data changes.  Very fast if no structural changes.
func (nv *NetView) UpdateLayers() {
	sw := nv.SceneWidget()
	se := sw.SceneXYZ()

	if nv.Net == nil || nv.Net.NumLayers() == 0 {
		se.DeleteChildren()
		se.Meshes.Reset()
		return
	}
	nb := nv.Net.AsEmer()
	if nv.NeedsRebuild() {
		se.Background = colors.Scheme.Background
	}
	nlay := nv.Net.NumLayers()
	laysGp := se.ChildByName("Layers", 0).(*xyz.Group)

	layConfig := tree.TypePlan{}
	for li := range nlay {
		ly := nv.Net.EmerLayer(li)
		layConfig.Add(types.For[xyz.Group](), ly.StyleName())
	}

	if !tree.Update(laysGp, layConfig) {
		for li := range laysGp.Children {
			ly := nv.Net.EmerLayer(li)
			lmesh := errors.Log1(se.MeshByName(ly.StyleName()))
			se.SetMesh(lmesh) // does update
		}
		return
	}

	gpConfig := tree.TypePlan{}
	gpConfig.Add(types.For[LayObj](), "layer")
	gpConfig.Add(types.For[LayName](), "name")

	nmin, nmax := nb.MinPos, nb.MaxPos
	nsz := nmax.Sub(nmin).Sub(math32.Vec3(1, 1, 0)).Max(math32.Vec3(1, 1, 1))
	nsc := math32.Vec3(1.0/nsz.X, 1.0/nsz.Y, 1.0/nsz.Z)
	szc := math32.Max(nsc.X, nsc.Y)
	poff := math32.Vector3Scalar(0.5)
	poff.Y = -0.5
	for li, lgi := range laysGp.Children {
		ly := nv.Net.EmerLayer(li)
		lb := ly.AsEmer()
		lmesh, _ := se.MeshByName(ly.StyleName())
		if lmesh == nil {
			NewLayMesh(se, nv, ly)
		} else {
			lmesh.(*LayMesh).Lay = ly // make sure
		}
		lg := lgi.(*xyz.Group)
		gpConfig[1].Name = ly.StyleName() // text2d textures use obj name, so must be unique
		tree.Update(lg, gpConfig)
		lp := lb.Pos.Pos
		lp.Y = -lp.Y // reverse direction
		lp = lp.Sub(nmin).Mul(nsc).Sub(poff)
		lg.Pose.Pos.Set(lp.X, lp.Z, lp.Y)
		lg.Pose.Scale.Set(nsc.X*lb.Pos.Scale, szc, nsc.Y*lb.Pos.Scale)

		lo := lg.Child(0).(*LayObj)
		lo.Defaults()
		lo.LayName = ly.StyleName()
		lo.NetView = nv
		lo.SetMeshName(ly.StyleName())
		lo.Material.Color = colors.FromRGB(255, 100, 255)
		lo.Material.Reflective = 8
		lo.Material.Bright = 8
		lo.Material.Shiny = 30
		// note: would actually be better to NOT cull back so you can view underneath
		// but then the front and back fight against each other, causing flickering

		txt := lg.Child(1).(*LayName)
		txt.Defaults()
		txt.NetView = nv
		txt.SetText(ly.StyleName())
		txt.Pose.Scale = math32.Vector3Scalar(nv.Params.LayNmSize).Div(lg.Pose.Scale)
		txt.Styles.Background = colors.Uniform(colors.Transparent)
		txt.Styles.Text.Align = styles.Start
		txt.Styles.Text.AlignV = styles.Start
	}
	sw.XYZ.SetNeedsUpdate()
	sw.NeedsRender()
}

// ViewDefaults are the default 3D view params
func (nv *NetView) ViewDefaults(se *xyz.Scene) {
	se.Camera.Pose.Pos.Set(0, 1.5, 2.5) // more "top down" view shows more of layers
	// 	vs.Camera.Pose.Pos.Set(0, 1, 2.75) // more "head on" for larger / deeper networks
	se.Camera.Near = 0.1
	se.Camera.LookAt(math32.Vec3(0, 0, 0), math32.Vec3(0, 1, 0))
	nv.Styler(func(s *styles.Style) {
		se.Background = colors.Scheme.Background
	})
	xyz.NewAmbient(se, "ambient", 0.1, xyz.DirectSun)
	xyz.NewDirectional(se, "directional", 0.5, xyz.DirectSun).Pos.Set(0, 2, 5)
	xyz.NewPoint(se, "point", .2, xyz.DirectSun).Pos.Set(0, 2, -5)
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
	lb := lay.AsEmer()
	idx1d := lb.Shape.Offset(idx)
	if idx1d >= lb.Shape.Len() {
		raw, hasval = 0, false
	} else {
		raw, hasval = nv.Data.UnitValue(lb.Name, nv.Var, idx1d, nv.RecNo, nv.Di)
	}
	scaled, clr = nv.UnitValColor(lay, idx1d, raw, hasval)
	return
}

// UnitValRaster returns the raw value, scaled value, and color representation
// for given unit of given layer, and given raster counter index value (0..RasterMax)
// scaled is in range -1..1
func (nv *NetView) UnitValRaster(lay emer.Layer, idx []int, rCtr int) (raw, scaled float32, clr color.RGBA, hasval bool) {
	lb := lay.AsEmer()
	idx1d := lb.SampleShape.Offset(idx)
	ridx := lb.SampleIndexes
	if len(ridx) == 0 { // no rep
		if idx1d >= lb.Shape.Len() {
			raw, hasval = 0, false
		} else {
			raw, hasval = nv.Data.UnitValRaster(lb.Name, nv.Var, idx1d, rCtr, nv.Di)
		}
	} else {
		if idx1d >= len(ridx) {
			raw, hasval = 0, false
		} else {
			idx1d = ridx[idx1d]
			raw, hasval = nv.Data.UnitValRaster(lb.Name, nv.Var, idx1d, rCtr, nv.Di)
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
		if lay.StyleName() == nv.Data.PathLay && idx1d == nv.Data.PathUnIndex {
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

func (nv *NetView) Layers() *xyz.Group {
	se := nv.SceneXYZ()
	lgpi := se.ChildByName("Layers", 0)
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
		lbConfig.Add(types.For[xyz.Text2D](), ls)
	}
	if tree.Update(lgp, lbConfig) {
		for i, ls := range labs {
			lb := lgp.ChildByName(ls, i).(*xyz.Text2D)
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
	lgp := nv.Layers()
	ly := lgp.ChildByName(lay, 0)
	if ly == nil {
		return nil
	}
	return ly.(*xyz.Group)
}

func (nv *NetView) MakeToolbar(p *tree.Plan) {
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(nv.Update).SetText("Init").SetIcon(icons.Update)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(nv.Current).SetIcon(icons.Update)
	})
	tree.Add(p, func(w *core.Button) {
		w.SetText("Config").SetIcon(icons.Settings).
			SetTooltip("set parameters that control display (font size etc)").
			OnClick(func(e events.Event) {
				FormDialog(nv, &nv.Params, nv.Name+" Params")
			})
	})
	tree.Add(p, func(w *core.Separator) {})
	tree.Add(p, func(w *core.Separator) {})
	tree.Add(p, func(w *core.Button) {
		w.SetText("Weights").SetType(core.ButtonAction).SetMenu(func(m *core.Scene) {
			fb := core.NewFuncButton(m).SetFunc(nv.SaveWeights)
			fb.SetIcon(icons.Save)
			fb.Args[0].SetTag(`extension:".wts,.wts.gz"`)
			fb = core.NewFuncButton(m).SetFunc(nv.OpenWeights)
			fb.SetIcon(icons.Open)
			fb.Args[0].SetTag(`extension:".wts,.wts.gz"`)
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetText("Params").SetIcon(icons.Info).SetMenu(func(m *core.Scene) {
			core.NewFuncButton(m).SetFunc(nv.ShowNonDefaultParams).SetIcon(icons.Info)
			core.NewFuncButton(m).SetFunc(nv.ShowAllParams).SetIcon(icons.Info)
			core.NewFuncButton(m).SetFunc(nv.ShowKeyLayerParams).SetIcon(icons.Info)
			core.NewFuncButton(m).SetFunc(nv.ShowKeyPathParams).SetIcon(icons.Info)
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetText("Net Data").SetIcon(icons.Save).SetMenu(func(m *core.Scene) {
			core.NewFuncButton(m).SetFunc(nv.Data.SaveJSON).SetText("Save Net Data").SetIcon(icons.Save)
			core.NewFuncButton(m).SetFunc(nv.Data.OpenJSON).SetText("Open Net Data").SetIcon(icons.Open)
			core.NewSeparator(m)
			core.NewFuncButton(m).SetFunc(nv.PlotSelectedUnit).SetIcon(icons.Open)
		})
	})
	tree.Add(p, func(w *core.Separator) {})
	ditp := "data parallel index -- for models running multiple input patterns in parallel, this selects which one is viewed"
	tree.Add(p, func(w *core.Text) {
		w.SetText("Di:").SetTooltip(ditp)
	})
	tree.Add(p, func(w *core.Spinner) {
		w.SetMin(0).SetStep(1).SetValue(float32(nv.Di)).SetTooltip(ditp)
		w.OnChange(func(e events.Event) {
			maxData := nv.Net.MaxParallelData()
			md := int(w.Value)
			if md < maxData && md >= 0 {
				nv.Di = md
			}
			w.SetValue(float32(nv.Di))
			nv.UpdateView()
		})
	})

	tree.Add(p, func(w *core.Separator) {})

	tree.Add(p, func(w *core.Switch) {
		w.SetText("Raster").SetChecked(nv.Params.Raster.On).
			SetTooltip("Toggles raster plot mode -- displays values on one axis (Z by default) and raster counter (time) along the other (X by default)").
			OnChange(func(e events.Event) {
				nv.Params.Raster.On = w.IsChecked()
				// nv.ReconfigMeshes()
				nv.UpdateView()
			})
	})
	tree.Add(p, func(w *core.Switch) {
		w.SetText("X").SetType(core.SwitchCheckbox).SetChecked(nv.Params.Raster.XAxis).
			SetTooltip("If checked, the raster (time) dimension is plotted along the X (horizontal) axis of the layers, otherwise it goes in the depth (Z) dimension").
			OnChange(func(e events.Event) {
				nv.Params.Raster.XAxis = w.IsChecked()
				nv.UpdateView()
			})
	})
	vp, ok := nv.VarParams[nv.Var]
	if !ok {
		vp = &VarParams{}
		vp.Defaults()
	}

	var minSpin, maxSpin *core.Spinner
	var minSwitch, maxSwitch *core.Switch

	tree.Add(p, func(w *core.Separator) {})
	tree.AddAt(p, "minSwitch", func(w *core.Switch) {
		minSwitch = w
		w.SetText("Min").SetType(core.SwitchCheckbox).SetChecked(vp.Range.FixMin).
			SetTooltip("Fix the minimum end of the displayed value range to value shown in next box.  Having both min and max fixed is recommended where possible for speed and consistent interpretability of the colors.").
			OnChange(func(e events.Event) {
				vp := nv.VarParams[nv.Var]
				vp.Range.FixMin = w.IsChecked()
				minSpin.UpdateWidget().NeedsRender()
				nv.UpdateView()
			})
		w.Updater(func() {
			vp := nv.VarParams[nv.Var]
			if vp != nil {
				w.SetChecked(vp.Range.FixMin)
			}
		})
	})
	tree.AddAt(p, "minSpin", func(w *core.Spinner) {
		minSpin = w
		w.SetValue(vp.Range.Min).
			OnChange(func(e events.Event) {
				vp := nv.VarParams[nv.Var]
				vp.Range.SetMin(w.Value)
				vp.Range.FixMin = true
				minSwitch.UpdateWidget().NeedsRender()
				if vp.ZeroCtr && vp.Range.Min < 0 && vp.Range.FixMax {
					vp.Range.SetMax(-vp.Range.Min)
				}
				if vp.ZeroCtr {
					maxSpin.UpdateWidget().NeedsRender()
				}
				nv.UpdateView()
			})
		w.Updater(func() {
			vp := nv.VarParams[nv.Var]
			if vp != nil {
				w.SetValue(vp.Range.Min)
			}
		})
	})

	tree.AddAt(p, "cmap", func(w *core.ColorMapButton) {
		nv.ColorMapButton = w
		w.MapName = string(nv.Params.ColorMap)
		w.SetTooltip("Color map for translating values into colors -- click to select alternative.")
		w.Styler(func(s *styles.Style) {
			s.Min.X.Em(10)
			s.Min.Y.Em(1.2)
			s.Grow.Set(0, 1)
		})
		w.OnChange(func(e events.Event) {
			cmap, ok := colormap.AvailableMaps[string(nv.ColorMapButton.MapName)]
			if ok {
				nv.ColorMap = cmap
			}
			nv.UpdateView()
		})
	})

	tree.AddAt(p, "maxSwitch", func(w *core.Switch) {
		maxSwitch = w
		w.SetText("Max").SetType(core.SwitchCheckbox).SetChecked(vp.Range.FixMax).
			SetTooltip("Fix the maximum end of the displayed value range to value shown in next box.  Having both min and max fixed is recommended where possible for speed and consistent interpretability of the colors.").
			OnChange(func(e events.Event) {
				vp := nv.VarParams[nv.Var]
				vp.Range.FixMax = w.IsChecked()
				maxSpin.UpdateWidget().NeedsRender()
				nv.UpdateView()
			})
		w.Updater(func() {
			vp := nv.VarParams[nv.Var]
			if vp != nil {
				w.SetChecked(vp.Range.FixMax)
			}
		})
	})

	tree.AddAt(p, "maxSpin", func(w *core.Spinner) {
		maxSpin = w
		w.SetValue(vp.Range.Max).OnChange(func(e events.Event) {
			vp := nv.VarParams[nv.Var]
			vp.Range.SetMax(w.Value)
			vp.Range.FixMax = true
			maxSwitch.UpdateWidget().NeedsRender()
			if vp.ZeroCtr && vp.Range.Max > 0 && vp.Range.FixMin {
				vp.Range.SetMin(-vp.Range.Max)
			}
			if vp.ZeroCtr {
				minSpin.UpdateWidget().NeedsRender()
			}
			nv.UpdateView()
		})
		w.Updater(func() {
			vp := nv.VarParams[nv.Var]
			if vp != nil {
				w.SetValue(vp.Range.Max)
			}
		})
	})

	tree.AddAt(p, "zeroCtrSwitch", func(w *core.Switch) {
		w.SetText("ZeroCtr").SetChecked(vp.ZeroCtr).
			SetTooltip("keep Min - Max centered around 0, and use negative heights for units -- else use full min-max range for height (no negative heights)").
			OnChange(func(e events.Event) {
				vp := nv.VarParams[nv.Var]
				vp.ZeroCtr = w.IsChecked()
				nv.UpdateView()
			})
		w.Updater(func() {
			vp := nv.VarParams[nv.Var]
			if vp != nil {
				w.SetChecked(vp.ZeroCtr)
			}
		})
	})
}

func (nv *NetView) MakeViewbar(p *tree.Plan) {
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.Update).SetTooltip("reset to default initial display").
			OnClick(func(e events.Event) {
				nv.SceneXYZ().SetCamera("default")
				nv.UpdateView()
			})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.ZoomIn).SetTooltip("zoom in")
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Zoom(-.05)
			nv.UpdateView()
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.ZoomOut).SetTooltip("zoom out")
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Zoom(.05)
			nv.UpdateView()
		})
	})
	tree.Add(p, func(w *core.Separator) {})
	tree.Add(p, func(w *core.Text) {
		w.SetText("Rot:").SetTooltip("rotate display")
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.KeyboardArrowLeft)
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Orbit(5, 0)
			nv.UpdateView()
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.KeyboardArrowUp)
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Orbit(0, 5)
			nv.UpdateView()
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.KeyboardArrowDown)
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Orbit(0, -5)
			nv.UpdateView()
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.KeyboardArrowRight)
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Orbit(-5, 0)
			nv.UpdateView()
		})
	})
	tree.Add(p, func(w *core.Separator) {})

	tree.Add(p, func(w *core.Text) {
		w.SetText("Pan:").SetTooltip("pan display")
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.KeyboardArrowLeft)
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Pan(-.2, 0)
			nv.UpdateView()
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.KeyboardArrowUp)
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Pan(0, .2)
			nv.UpdateView()
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.KeyboardArrowDown)
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Pan(0, -.2)
			nv.UpdateView()
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.KeyboardArrowRight)
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			nv.SceneXYZ().Camera.Pan(.2, 0)
			nv.UpdateView()
		})
	})
	tree.Add(p, func(w *core.Separator) {})

	tree.Add(p, func(w *core.Text) { w.SetText("Save:") })

	for i := 1; i <= 4; i++ {
		nm := fmt.Sprintf("%d", i)
		tree.AddAt(p, "saved-"+nm, func(w *core.Button) {
			w.SetText(nm).
				SetTooltip("first click (or + Shift) saves current view, second click restores to saved state")
			w.OnClick(func(e events.Event) {
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
		})
	}
	tree.Add(p, func(w *core.Separator) {})

	tree.Add(p, func(w *core.Text) {
		w.SetText("Time:").
			SetTooltip("states are recorded over time -- last N can be reviewed using these buttons")
	})

	tree.AddAt(p, "rec", func(w *core.Text) {
		w.SetText(fmt.Sprintf("  %4d  ", nv.RecNo)).
			SetTooltip("current view record: -1 means latest, 0 = earliest")
		w.Styler(func(s *styles.Style) {
			s.Min.X.Ch(5)
		})
		w.Updater(func() {
			w.SetText(fmt.Sprintf("  %4d  ", nv.RecNo))
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.FirstPage).SetTooltip("move to first record (start of history)")
		w.OnClick(func(e events.Event) {
			if nv.RecFullBkwd() {
				nv.UpdateView()
			}
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.FastRewind).SetTooltip("move earlier by N records (default 10)")
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			if nv.RecFastBkwd() {
				nv.UpdateView()
			}
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.SkipPrevious).SetTooltip("move earlier by 1")
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			if nv.RecBkwd() {
				nv.UpdateView()
			}
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.PlayArrow).SetTooltip("move to latest and always display latest (-1)")
		w.OnClick(func(e events.Event) {
			if nv.RecTrackLatest() {
				nv.UpdateView()
			}
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.SkipNext).SetTooltip("move later by 1")
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			if nv.RecFwd() {
				nv.UpdateView()
			}
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.FastForward).SetTooltip("move later by N (default 10)")
		w.Styler(func(s *styles.Style) {
			s.SetAbilities(true, abilities.RepeatClickable)
		})
		w.OnClick(func(e events.Event) {
			if nv.RecFastFwd() {
				nv.UpdateView()
			}
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.LastPage).SetTooltip("move to end (current time, tracking latest updates)")
		w.OnClick(func(e events.Event) {
			if nv.RecTrackLatest() {
				nv.UpdateView()
			}
		})
	})
}

// SaveWeights saves the network weights.
func (nv *NetView) SaveWeights(filename core.Filename) { //types:add
	nv.Net.AsEmer().SaveWeightsJSON(filename)
}

// OpenWeights opens the network weights.
func (nv *NetView) OpenWeights(filename core.Filename) { //types:add
	nv.Net.AsEmer().OpenWeightsJSON(filename)
}

// ShowNonDefaultParams shows a dialog of all the parameters that
// are not at their default values in the network.  Useful for setting params.
func (nv *NetView) ShowNonDefaultParams() string { //types:add
	nds := nv.Net.AsEmer().NonDefaultParams()
	texteditor.TextDialog(nv, "Non Default Params: "+nv.Name, nds)
	return nds
}

// ShowAllParams shows a dialog of all the parameters in the network.
func (nv *NetView) ShowAllParams() string { //types:add
	nds := nv.Net.AsEmer().AllParams()
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
