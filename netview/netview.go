// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package netview provides the NetView interactive 3D network viewer, implemented in the GoGi 3D framework.
*/
package netview

import (
	"fmt"

	"github.com/emer/emergent/emer"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/gi3d"
	"github.com/goki/gi/giv"
	"github.com/goki/gi/mat32"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
)

// NetViewParams holds parameters controlling how the view is rendered
type NetViewParams struct {
	UnitSize  float32 `min:"0.1" max:"1" step:"0.1" desc:"size of a single unit, where 1 = full width and no space.. .9 default"`
	LayNmSize float32 `min:"0.01" max:".1" step:"0.01" def:"0.05" desc:"size of the layer name labels -- entire network view is unit sized"`
}

func (nv *NetViewParams) Defaults() {
	nv.UnitSize = .9
	nv.LayNmSize = .05
}

// NetView is a GoGi Widget that provides a 3D network view using the GoGi gi3d
// 3D framework.
type NetView struct {
	gi.Layout
	Net    emer.Network  `desc:"the network that we're viewing"`
	Var    string        `desc:"current variable that we're viewing"`
	Vars   []string      `desc:"the list of variables to view"`
	Params NetViewParams `desc:"parameters controlling how the view is rendered"`
	// todo: need a scalebar construct here..
}

var KiT_NetView = kit.Types.AddType(&NetView{}, NetViewProps)

func (nv *NetView) Defaults() {
	nv.Params.Defaults()
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
	nv.Update("")
}

func (nv *NetView) HasLayers() bool {
	if nv.Net == nil || nv.Net.NLayers() == 0 {
		return false
	}
	return true
}

// Update updates the display based on current state of network
// counters string, if non-empty, will be displayed at bottom of view, showing current
// counter state
func (nv *NetView) Update(counters string) {
	if !nv.IsVisible() || nv.Net == nil || nv.Net.NLayers() == 0 {
		return
	}
	if counters != "" {
		nv.SetCounters(counters)
	}
	vs := nv.Scene()
	if len(vs.Kids) != nv.Net.NLayers() {
		nv.Config()
	}
	vs.UpdateMeshes()
	vs.UpdateSig()
}

// Config configures the overall view widget
func (nv *NetView) Config() {
	nv.Lay = gi.LayoutVert
	if nv.Params.UnitSize == 0 {
		nv.Defaults()
	}
	// nv.SetProp("spacing", gi.StdDialogVSpaceUnits)
	config := kit.TypeAndNameList{}
	config.Add(gi.KiT_ToolBar, "tbar")
	config.Add(gi.KiT_Layout, "net")
	config.Add(gi.KiT_Label, "counters")
	mods, updt := nv.ConfigChildren(config, false)
	if !mods {
		updt = nv.UpdateStart()
	}

	nlay := nv.NetLay()
	nlay.Lay = gi.LayoutHoriz
	nlay.SetProp("max-width", -1)
	nlay.SetProp("max-height", -1)

	vncfg := kit.TypeAndNameList{}
	vncfg.Add(gi.KiT_Frame, "vars")
	vncfg.Add(gi3d.KiT_Scene, "scene")
	nlay.ConfigChildren(vncfg, false) // won't do update b/c of above updt

	nv.VarsConfig()
	nv.ViewConfig()
	nv.ToolbarConfig()

	ctrs := nv.Counters()
	ctrs.Redrawable = true
	ctrs.SetText("Counters: ")
	nv.UpdateEnd(updt)
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
	if nv.Net == nil || nv.Net.NLayers() == 0 {
		nv.Vars = nil
		return
	}
	lay := nv.Net.Layer(0)
	unvars := lay.UnitVarNames()
	nv.Vars = make([]string, len(unvars))
	copy(nv.Vars, unvars)
	// todo: get prjn vars
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
	gpConfig.Add(gi3d.KiT_Object, "layer")
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

		lo := lg.Child(0).(*gi3d.Object)
		lo.Defaults()
		lo.SetMeshName(vs, ly.Name())
		lo.Mat.Color.SetUInt8(255, 100, 255, 128)
		lo.Mat.Specular.SetUInt8(128, 128, 128, 255)
		lo.Mat.CullBack = true
		lo.Mat.CullFront = false
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
	dir := gi3d.AddNewDirLight(vs, "dir", 1, gi3d.DirectSun)
	dir.Pos.Set(0, 2, 1)
	// point := gi3d.AddNewPointLight(vs, "point", 1, gi3d.DirectSun)
	// point.Pos.Set(0, 2, 5)
	// spot := gi3d.AddNewSpotLight(vs, "spot", 1, gi3d.DirectSun)
	// spot.Pose.Pos.Set(0, 2, 5)
	// spot.LookAtOrigin()
}

// UnitVal returns the raw value, scaled value, and color representation for given unit of given layer
// scaled is in range -1..1
// todo: could incorporate history etc..
func (nv *NetView) UnitVal(lay emer.Layer, idx []int) (raw, scaled float32, clr gi.Color) {
	raw, _ = lay.UnitVal(nv.Var, idx)
	scaled = mat32.Clamp(raw, -1, 1)
	if scaled < 0 {
		clr.R = uint8(50.0)
		clr.G = uint8(50.0)
		clr.B = uint8(50.0 - scaled*205.0)
		clr.A = uint8(128.0 - scaled*127.0)
	} else {
		clr.R = uint8(50.0 + scaled*205.0)
		clr.G = uint8(50.0)
		clr.B = uint8(50.0)
		clr.A = uint8(128.0 + scaled*127.0)
	}
	return
}

func (nv *NetView) ToolbarConfig() {
	tbar := nv.Toolbar()
	if len(tbar.Kids) != 0 {
		return
	}
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
	tbar.AddSeparator("ctrl")
	tbar.AddAction(gi.ActOpts{Label: "Init", Icon: "update", Tooltip: "fully redraw display"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			nv.Config()
			nv.Update("")
		})
	tbar.AddAction(gi.ActOpts{Label: "Config", Icon: "gear", Tooltip: "set parameters that control display (font size etc)"}, nv.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			giv.StructViewDialog(nv.Viewport, &nv.Params, giv.DlgOpts{Title: nv.Nm + " Params"}, nil, nil)
		})
	// todo: colorbar
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

var NetViewProps = ki.Props{
	"max-width":  -1,
	"max-height": -1,
}
