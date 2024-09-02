// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package netview provides the NetView interactive 3D network viewer,
// implemented in the Cogent Core 3D framework.
package netview

//go:generate core generate -add-types

import (
	"image/color"
	"log"
	"log/slog"
	"reflect"
	"strings"
	"sync"
	"time"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/colors/colormap"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/math32/minmax"
	"cogentcore.org/core/styles"
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
	VarOptions map[string]*VarOptions

	// current var params -- only valid during Update of display
	CurVarOptions *VarOptions `json:"-" xml:"-" display:"-"`

	// parameters controlling how the view is rendered
	Options Options

	// color map for mapping values to colors -- set by name in Options
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

	// these are used to detect need to update
	layerNameSizeShown float32
	hasPaths           bool
	pathTypeShown      string
	pathWidthShown     float32
}

func (nv *NetView) Init() {
	nv.Frame.Init()
	nv.Options.Defaults()
	nv.ColorMap = colormap.AvailableMaps[string(nv.Options.ColorMap)]
	nv.RecNo = -1
	nv.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Grow.Set(1, 1)
	})

	tree.AddChildAt(nv, "tbar", func(w *core.Toolbar) {
		w.FinalStyler(func(s *styles.Style) {
			s.Wrap = true
		})
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
			pathsGp := xyz.NewGroup(se)
			pathsGp.Name = "Paths"
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
	nv.Data.Init(nv.Net, nv.Options.MaxRecs, nv.Options.NoSynData, nv.Net.MaxParallelData())
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
	nv.Options.MaxRecs = max
	nv.Data.Init(nv.Net, nv.Options.MaxRecs, nv.Options.NoSynData, nv.Net.MaxParallelData())
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
	nv.Data.PathType = nv.Options.PathType
	nv.Data.Record(nv.LastCtrs, rastCtr, nv.Options.Raster.Max)
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
	vp, ok := nv.VarOptions[nv.Var]
	if !ok {
		nv.DataMu.Unlock()
		log.Printf("NetView: %v variable: %v not found\n", nv.Name, nv.Var)
		return
	}
	nv.CurVarOptions = vp

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
		nv.RecNo = nv.Data.Ring.Len - nv.Options.NFastSteps
	} else {
		nv.RecNo -= nv.Options.NFastSteps
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
	nv.RecNo += nv.Options.NFastSteps
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
	nv.VarOptions = make(map[string]*VarOptions, len(nv.Vars))

	nv.SynVars = synvars
	nv.SynVarsMap = make(map[string]int, len(synvars))
	for i, vn := range nv.SynVars {
		nv.SynVarsMap[vn] = i
	}

	unprops := nv.Net.UnitVarProps()
	pathprops := nv.Net.SynVarProps()
	for _, nm := range nv.Vars {
		vp := &VarOptions{Var: nm}
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
		nv.VarOptions[nm] = vp
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
			tb.Tooltip = ct.Doc
			tabs[ct.Cat] = tf
			tf.Styler(func(s *styles.Style) {
				s.Display = styles.Grid
				s.Columns = nv.Options.NVarCols
				s.Grow.Set(1, 1)
				s.Overflow.Y = styles.OverflowAuto
				s.Background = colors.Scheme.SurfaceContainerLow
			})
		}
		for _, vn := range nv.Vars {
			cat := ""
			pstr := ""
			doc := ""
			if strings.HasPrefix(vn, "r.") || strings.HasPrefix(vn, "s.") {
				pstr = pathprops[vn[2:]]
				cat = "Wt" // default
			} else {
				pstr = unprops[vn]
				cat = "Unit"
			}
			if pstr != "" {
				rstr := reflect.StructTag(pstr)
				doc = rstr.Get("doc")
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
			if doc != "" {
				w.Tooltip = vn + ": " + doc
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
	if nv.CurVarOptions == nil || nv.CurVarOptions.Var != nv.Var {
		ok := false
		nv.CurVarOptions, ok = nv.VarOptions[nv.Var]
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
		clp := nv.CurVarOptions.Range.ClipValue(raw)
		norm := nv.CurVarOptions.Range.NormValue(clp)
		var op float32
		if nv.CurVarOptions.ZeroCtr {
			scaled = float32(2*norm - 1)
			op = (nv.Options.ZeroAlpha + (1-nv.Options.ZeroAlpha)*math32.Abs(scaled))
		} else {
			scaled = float32(norm)
			op = (nv.Options.ZeroAlpha + (1-nv.Options.ZeroAlpha)*0.8) // no meaningful alpha -- just set at 80\%
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
