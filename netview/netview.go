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

	"github.com/emer/emergent/emer"
	"github.com/emer/etable/minmax"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/gi3d"
	"github.com/goki/gi/giv"
	"github.com/goki/gi/mat32"
	"github.com/goki/gi/units"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
)

// NetView is a GoGi Widget that provides a 3D network view using the GoGi gi3d
// 3D framework.
type NetView struct {
	gi.Layout
	Net          emer.Network          `desc:"the network that we're viewing"`
	Var          string                `desc:"current variable that we're viewing"`
	Vars         []string              `desc:"the list of variables to view, along with view-specific params"`
	VarParams    map[string]*VarParams `desc:"parameters for the list of variables to view"`
	CurVarParams *VarParams            `json:"-" xml:"-" view:"-" desc:"current var params -- only valid during Update of display"`
	Params       Params                `desc:"parameters controlling how the view is rendered"`
	ColorMap     *giv.ColorMap         `desc:"color map for mapping values to colors -- set by name in Params"`
}

var KiT_NetView = kit.Types.AddType(&NetView{}, NetViewProps)

// AddNewNetView adds a new NetView to given parent node, with given name.
func AddNewNetView(parent ki.Ki, name string) *NetView {
	return parent.AddNewChild(KiT_NetView, name).(*NetView)
}

func (nv *NetView) Defaults() {
	nv.Params.NetView = nv
	nv.Params.Defaults()
	nv.ColorMap = giv.AvailColorMaps[string(nv.Params.ColorMap)]
}

// SetNet sets the network to view and updates view
func (nv *NetView) SetNet(net emer.Network) {
	nv.Defaults()
	nv.Net = net
	nv.Config()
}

// SetVar sets the variable to view and updates the display
func (nv *NetView) SetVar(vr string) {
	nv.Var = vr
	nv.VarsUpdate()
	nv.VarScaleUpdate(nv.Var)
	nv.Update("")
}

func (nv *NetView) HasLayers() bool {
	if nv.Net == nil || nv.Net.NLayers() == 0 {
		return false
	}
	return true
}

// GoUpdate is the update call to make from another go routine
// it does the proper blocking to coordinate with GUI updates
// generated on the main GUI thread.
func (nv *NetView) GoUpdate(counters string) {
	if !nv.IsVisible() || !nv.HasLayers() {
		return
	}
	if nv.Viewport.IsUpdatingNode() {
		return
	}
	nv.Viewport.BlockUpdates()
	vs := nv.Scene()
	updt := vs.UpdateStart()
	nv.UpdateImpl(counters)
	nv.Viewport.UnblockUpdates()
	vs.UpdateEnd(updt)
}

// Update updates the display based on current state of network
// counters string, if non-empty, will be displayed at bottom of view,
// showing current counter state
// This version is for calling within main window eventloop goroutine
// use GoUpdate version for calling outside of main goroutine.
func (nv *NetView) Update(counters string) {
	if !nv.IsVisible() || !nv.HasLayers() {
		return
	}
	vs := nv.Scene()
	updt := vs.UpdateStart()
	nv.UpdateImpl(counters)
	vs.UpdateEnd(updt)
}

func (nv *NetView) UpdateImpl(counters string) {
	if counters != "" {
		nv.SetCounters(counters)
	}

	vp, ok := nv.VarParams[nv.Var]
	if !ok {
		log.Printf("NetView: %v variable: %v not found\n", nv.Nm, nv.Var)
		return
	}
	nv.CurVarParams = vp

	if !vp.Range.FixMin || !vp.Range.FixMax {
		needUpdt := false
		// need to autoscale
		min, max, _ := nv.Net.VarRange(nv.Var)
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
		if needUpdt {
			nv.VarScaleUpdate(nv.Var)
		}
	}

	vs := nv.Scene()
	if len(vs.Kids) != nv.Net.NLayers() {
		nv.Config()
	}
	vs.UpdateMeshes()
}

// Config configures the overall view widget
func (nv *NetView) Config() {
	nv.Lay = gi.LayoutVert
	if nv.Params.UnitSize == 0 {
		nv.Defaults()
	}
	cmap, ok := giv.AvailColorMaps[string(nv.Params.ColorMap)]
	if ok {
		nv.ColorMap = cmap
	} else {
		log.Printf("NetView: %v  ColorMap named: %v not found in AvailColorMaps\n", nv.Nm, nv.Params.ColorMap)
	}
	nv.SetProp("spacing", gi.StdDialogVSpaceUnits)
	config := kit.TypeAndNameList{}
	config.Add(gi.KiT_ToolBar, "tbar")
	config.Add(gi.KiT_Layout, "net")
	config.Add(gi.KiT_Label, "counters")
	config.Add(gi.KiT_ToolBar, "vbar")
	mods, updt := nv.ConfigChildren(config, false)
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
	nlay.ConfigChildren(vncfg, false) // won't do update b/c of above updt

	nv.VarsConfig()
	nv.ViewConfig()
	nv.ToolbarConfig()
	nv.ViewbarConfig()

	ctrs := nv.Counters()
	ctrs.Redrawable = true
	ctrs.SetText("Counters: ")
	nv.UpdateEnd(updt)
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

func (nv *NetView) SetCounters(ctrs string) {
	ct := nv.Counters()
	if ct.Text != ctrs {
		ct.SetText(ctrs)
	}
}

// VarsListUpdate updates the list of network variables
func (nv *NetView) VarsListUpdate() {
	if !nv.HasLayers() {
		nv.Vars = nil
		return
	}
	lay := nv.Net.Layer(0)
	unvars := lay.UnitVarNames()
	if len(unvars) == len(nv.Vars) {
		return
	}
	nv.Vars = make([]string, len(unvars))
	copy(nv.Vars, unvars)

	nv.VarParams = make(map[string]*VarParams, len(nv.Vars))
	for _, nm := range nv.Vars {
		vp := &VarParams{Var: nm}
		vp.Defaults()
		nv.VarParams[nm] = vp
	}
	// todo: get prjn vars
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
	if mod {
		nv.Update("")
	}
	return mod
}

// VarsConfig configures the variables
func (nv *NetView) VarsConfig() {
	vl := nv.VarsLay()
	vl.SetReRenderAnchor()
	vl.Lay = gi.LayoutVert
	vl.SetProp("spacing", 0)
	vl.SetProp("vertical-align", gi.AlignTop)
	nv.VarsListUpdate()
	if len(nv.Vars) == 0 {
		vl.DeleteChildren(true)
		return
	}
	config := kit.TypeAndNameList{}
	for _, vn := range nv.Vars {
		config.Add(gi.KiT_Action, vn)
	}
	mods, updt := vl.ConfigChildren(config, false)
	if !mods {
		updt = vl.UpdateStart()
	}
	for i, vbi := range *vl.Children() {
		vb := vbi.(*gi.Action)
		vb.SetProp("margin", 0)
		vb.SetProp("max-width", -1)
		vn := nv.Vars[i]
		vb.SetText(vn)
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
		vs.Meshes = nil
		return
	}
	if len(vs.Lights) == 0 {
		nv.ViewDefaults()
	}
	nlay := nv.Net.NLayers()
	if len(vs.Meshes) != nlay+1 { // one extra for the text plane mesh..
		vs.Meshes = nil
	}
	layConfig := kit.TypeAndNameList{}
	for li := 0; li < nlay; li++ {
		lay := nv.Net.Layer(li)
		AddNewLayMesh(vs, nv, lay)
		layConfig.Add(gi3d.KiT_Group, lay.Name())
	}
	gpConfig := kit.TypeAndNameList{}
	gpConfig.Add(KiT_LayObj, "layer")
	gpConfig.Add(gi3d.KiT_Text2D, "name")

	mods, updt := vs.ConfigChildren(layConfig, false)
	if !mods {
		updt = vs.UpdateStart()
	}
	nmin, nmax := nv.Net.Bounds()
	nsz := nmax.Sub(nmin).Sub(mat32.Vec3{1, 1, 0}).Max(mat32.Vec3{1, 1, 1})
	nsc := mat32.Vec3{1.0 / nsz.X, 1.0 / nsz.Y, 1.0 / nsz.Z}
	szc := mat32.Max(nsc.X, nsc.Y)
	poff := mat32.NewVec3Scalar(0.5)
	poff.Y = -0.5
	for li, lgi := range *vs.Children() {
		ly := nv.Net.Layer(li)
		lg := lgi.(*gi3d.Group)
		lg.ConfigChildren(gpConfig, false) // won't do update b/c of above
		lp := ly.Pos().Sub(nmin).Mul(nsc).Sub(poff)
		rp := ly.RelPos()
		lg.Pose.Pos.Set(lp.X, lp.Z, lp.Y)
		lg.Pose.Scale.Set(nsc.X*rp.Scale, szc, nsc.Y*rp.Scale)

		lo := lg.Child(0).(*LayObj)
		lo.Defaults()
		lo.SetMeshName(vs, ly.Name())
		lo.Mat.Color.SetUInt8(255, 100, 255, 128)
		lo.Mat.Specular.SetUInt8(128, 128, 128, 255)
		lo.Mat.CullBack = true
		lo.Mat.CullFront = false
		// lo.Mat.Shiny = 10
		// note: would actually be better to NOT cull back so you can view underneath
		// but then the front and back fight against each other, causing flickering
		// really you ned

		txt := lg.Child(1).(*gi3d.Text2D)
		txt.Defaults(vs)
		txt.SetText(vs, ly.Name())
		txt.Pose.Scale = mat32.NewVec3Scalar(nv.Params.LayNmSize).Div(lg.Pose.Scale)
		txt.SetProp("text-align", gi.AlignLeft)
		txt.SetProp("vertical-align", gi.AlignTop)
	}
	vs.InitMeshes()
	vs.UpdateEnd(updt)
}

// ViewDefaults are the default 3D view params
func (nv *NetView) ViewDefaults() {
	vs := nv.Scene()
	vs.SetStretchMaxWidth()
	vs.SetStretchMaxHeight()
	vs.Defaults()
	vs.Camera.Pose.Pos.Set(0, 1, 2.75)
	vs.Camera.Near = 0.1
	vs.Camera.LookAt(mat32.Vec3{0, 0, 0}, mat32.Vec3{0, 1, 0})
	vs.BgColor.SetUInt8(255, 255, 255, 255) // white
	gi3d.AddNewAmbientLight(vs, "ambient", 0.3, gi3d.DirectSun)
	dir := gi3d.AddNewDirLight(vs, "dirUp", 0.3, gi3d.DirectSun)
	dir.Pos.Set(0, 1, 0)
	dir = gi3d.AddNewDirLight(vs, "dirBack", 0.6, gi3d.DirectSun)
	dir.Pos.Set(0, 1, -2.5)
	// point := gi3d.AddNewPointLight(vs, "point", 1, gi3d.DirectSun)
	// point.Pos.Set(0, 2, 5)
	// spot := gi3d.AddNewSpotLight(vs, "spot", 1, gi3d.DirectSun)
	// spot.Pose.Pos.Set(0, 2, 5)
	// spot.LookAtOrigin()
}

// UnitVal returns the raw value, scaled value, and color representation for given unit of given layer
// scaled is in range -1..1
// todo: incorporate history etc..
func (nv *NetView) UnitVal(lay emer.Layer, idx []int) (raw, scaled float32, clr gi.Color) {
	raw = lay.UnitVal(nv.Var, idx)
	if nv.CurVarParams == nil || nv.CurVarParams.Var != nv.Var {
		ok := false
		nv.CurVarParams, ok = nv.VarParams[nv.Var]
		if !ok {
			return
		}
	}
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
	return
}

func (nv *NetView) ToolbarConfig() {
	tbar := nv.Toolbar()
	if len(tbar.Kids) != 0 {
		return
	}
	tbar.SetStretchMaxWidth()
	tbar.AddAction(gi.ActOpts{Label: "Init", Icon: "update", Tooltip: "fully redraw display"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nv.Config()
			nv.Update("")
			nv.VarsUpdate()
		})
	tbar.AddAction(gi.ActOpts{Label: "Config", Icon: "gear", Tooltip: "set parameters that control display (font size etc)"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			giv.StructViewDialog(nv.Viewport, &nv.Params, giv.DlgOpts{Title: nv.Nm + " Params"}, nil, nil)
		})
	tbar.AddSeparator("file")
	tbar.AddAction(gi.ActOpts{Label: "Save Wts", Icon: "file-save"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			giv.CallMethod(nv, "SaveWeights", nv.Viewport) // this auto prompts for filename using file chooser
		})
	tbar.AddAction(gi.ActOpts{Label: "Open Wts", Icon: "file-open"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			giv.CallMethod(nv, "OpenWeights", nv.Viewport) // this auto prompts for filename using file chooser
		})
	tbar.AddAction(gi.ActOpts{Label: "Non Def Params", Icon: "info", Tooltip: "shows all the parameters that are not at default values -- useful for setting params"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nv.ShowNonDefaultParams()
		})
	tbar.AddAction(gi.ActOpts{Label: "All Params", Icon: "info", Tooltip: "shows all the parameters in the network"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nv.ShowAllParams()
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
			nvv.Update("")
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
			nvv.Update("")
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
			nvv.VarScaleUpdate(nvv.Var)
			nvv.Update("")
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
				nvv.Update("")
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
	tbar.AddAction(gi.ActOpts{Icon: "pan", Tooltip: "return to default pan / orbit mode where mouse drags move camera around (Shift = pan, Alt = pan target)"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			fmt.Printf("this will select pan mode\n")
		})
	tbar.AddAction(gi.ActOpts{Icon: "arrow", Tooltip: "turn on select mode for selecting units and layers with mouse clicks"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			fmt.Printf("this will select select mode\n")
		})
	tbar.AddSeparator("zoom")
	tbar.AddAction(gi.ActOpts{Icon: "update", Tooltip: "reset to default initial display"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nv.Scene().SetCamera("default")
			nv.Scene().UpdateSig()
		})
	tbar.AddAction(gi.ActOpts{Icon: "zoom-in", Tooltip: "zoom in"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nv.Scene().Camera.Zoom(-.05)
			nv.Scene().UpdateSig()
		})
	tbar.AddAction(gi.ActOpts{Icon: "zoom-out", Tooltip: "zoom out"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nv.Scene().Camera.Zoom(.05)
			nv.Scene().UpdateSig()
		})
	tbar.AddSeparator("rot")
	gi.AddNewLabel(tbar, "rot", "Rot:")
	tbar.AddAction(gi.ActOpts{Icon: "wedge-left"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nv.Scene().Camera.Orbit(5, 0)
			nv.Scene().UpdateSig()
		})
	tbar.AddAction(gi.ActOpts{Icon: "wedge-up"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nv.Scene().Camera.Orbit(0, 5)
			nv.Scene().UpdateSig()
		})
	tbar.AddAction(gi.ActOpts{Icon: "wedge-down"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nv.Scene().Camera.Orbit(0, -5)
			nv.Scene().UpdateSig()
		})
	tbar.AddAction(gi.ActOpts{Icon: "wedge-right"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nv.Scene().Camera.Orbit(-5, 0)
			nv.Scene().UpdateSig()
		})
	tbar.AddSeparator("pan")
	gi.AddNewLabel(tbar, "pan", "Pan:")
	tbar.AddAction(gi.ActOpts{Icon: "wedge-left"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nv.Scene().Camera.Pan(-.2, 0)
			nv.Scene().UpdateSig()
		})
	tbar.AddAction(gi.ActOpts{Icon: "wedge-up"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nv.Scene().Camera.Pan(0, .2)
			nv.Scene().UpdateSig()
		})
	tbar.AddAction(gi.ActOpts{Icon: "wedge-down"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nv.Scene().Camera.Pan(0, -.2)
			nv.Scene().UpdateSig()
		})
	tbar.AddAction(gi.ActOpts{Icon: "wedge-right"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nv.Scene().Camera.Pan(.2, 0)
			nv.Scene().UpdateSig()
		})
	tbar.AddSeparator("save")
	gi.AddNewLabel(tbar, "save", "Save:")
	tbar.AddAction(gi.ActOpts{Label: "1", Icon: "save", Tooltip: "first click saves current view, second click restores to saved state"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			err := nv.Scene().SetCamera("1")
			if err != nil {
				nv.Scene().SaveCamera("1")
			}
			nv.Scene().UpdateSig()
		})
	tbar.AddAction(gi.ActOpts{Label: "2", Icon: "save", Tooltip: "first click saves current view, second click restores to saved state"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			err := nv.Scene().SetCamera("2")
			if err != nil {
				nv.Scene().SaveCamera("2")
			}
			nv.Scene().UpdateSig()
		})
	tbar.AddAction(gi.ActOpts{Label: "3", Icon: "save", Tooltip: "first click saves current view, second click restores to saved state"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			err := nv.Scene().SetCamera("3")
			if err != nil {
				nv.Scene().SaveCamera("3")
			}
			nv.Scene().UpdateSig()
		})
	tbar.AddAction(gi.ActOpts{Label: "4", Icon: "save", Tooltip: "first click saves current view, second click restores to saved state"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			err := nv.Scene().SetCamera("4")
			if err != nil {
				nv.Scene().SaveCamera("4")
			}
			nv.Scene().UpdateSig()
		})
}

// func (nv *NetView) Render2D() {
// 	if nv.FullReRenderIfNeeded() {
// 		return
// 	}
// 	if nv.PushBounds() {
// 		nv.This().(gi.Node2D).ConnectEvents2D()
// 		nv.RenderScrolls()
// 		nv.Render2DChildren()
// 		nv.PopBounds()
// 	} else {
// 		nv.DisconnectAllEvents(gi.AllPris) // uses both Low and Hi
// 	}
// }

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
	giv.TextViewDialog(nv.Viewport, []byte(nds), giv.DlgOpts{Title: "Non Default Params"})
	return nds
}

// ShowAllParams shows a dialog of all the parameters in the network.
func (nv *NetView) ShowAllParams() string {
	nds := nv.Net.AllParams()
	giv.TextViewDialog(nv.Viewport, []byte(nds), giv.DlgOpts{Title: "All Params"})
	return nds
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
					"ext": ".wts",
				}},
			},
		}},
		{"OpenWeights", ki.Props{
			"desc": "open network weights from file",
			"icon": "file-open",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".wts",
				}},
			},
		}},
	},
}
