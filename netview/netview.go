// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import (
	"github.com/emer/emergent/emer"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/gi3d"
	"github.com/goki/gi/mat32"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
)

// NetView is a GoGi Widget that provides a 3D network view using the GoGi gi3d
// 3D framework.
type NetView struct {
	gi.Layout
	Net      emer.Network `desc:"the network that we're viewing"`
	Var      string       `desc:"current variable that we're viewing"`
	Vars     []string     `desc:"the list of variables to view"`
	UnitSize float32      `desc:"size of a single unit, where 1 = full width and no space.. .9 default"`
	// todo: need a scalebar construct here..
}

var KiT_NetView = kit.Types.AddType(&NetView{}, NetViewProps)

func (nv *NetView) Defaults() {
	nv.UnitSize = .9
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
	nv.Update()
}

func (nv *NetView) HasLayers() bool {
	if nv.Net == nil || nv.Net.NLayers() == 0 {
		return false
	}
	return true
}

// Update updates the display based on current state of network
func (nv *NetView) Update() {
	if !nv.IsVisible() || nv.Net == nil || nv.Net.NLayers() == 0 {
		return
	}
	vs := nv.ViewScene()
	if len(vs.Kids) != nv.Net.NLayers() {
		nv.Config()
	}
	// vs.UpdateMeshes()
	// note: something wrong about update still -- not rendering the lighting -- norms seem wrong.
	vs.InitMeshes()
	vs.UpdateSig()
}

// Config configures the overall view widget
func (nv *NetView) Config() {
	nv.Lay = gi.LayoutHoriz
	if nv.UnitSize == 0 {
		nv.Defaults()
	}
	// nv.SetProp("spacing", gi.StdDialogVSpaceUnits)
	config := nv.StdFrameConfig()
	mods, updt := nv.ConfigChildren(config, false)
	if !mods {
		updt = nv.UpdateStart()
	}
	nv.VarsConfig()
	nv.ViewConfig()
	nv.UpdateEnd(updt)
}

// StdFrameConfig returns a TypeAndNameList for configuring a standard Frame
func (nv *NetView) StdFrameConfig() kit.TypeAndNameList {
	config := kit.TypeAndNameList{}
	config.Add(gi.KiT_Frame, "vars-lay")
	config.Add(gi3d.KiT_Scene, "view-scene")
	return config
}

func (nv *NetView) VarsLay() *gi.Frame {
	return nv.ChildByName("vars-lay", 0).(*gi.Frame)
}

func (nv *NetView) ViewScene() *gi3d.Scene {
	return nv.ChildByName("view-scene", 1).(*gi3d.Scene)
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
	vs := nv.ViewScene()
	if nv.Net == nil || nv.Net.NLayers() == 0 {
		vs.DeleteChildren(true)
		vs.Meshes = nil
		return
	}
	if len(vs.Lights) == 0 {
		nv.ViewDefaults()
	}
	nlay := nv.Net.NLayers()
	if len(vs.Meshes) != nlay {
		vs.Meshes = nil
	}
	layConfig := kit.TypeAndNameList{}
	for li := 0; li < nlay; li++ {
		lay := nv.Net.Layer(li)
		AddNewLayMesh(vs, nv, lay)
		layConfig.Add(gi3d.KiT_Object, lay.Name())
	}
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
	for li, loi := range *vs.Children() {
		ly := nv.Net.Layer(li)
		lo := loi.(*gi3d.Object)
		lo.Defaults()
		lo.SetMesh(vs, ly.Name())
		lp := ly.Pos().Sub(nmin).Mul(nsc).Sub(poff)
		rp := ly.RelPos()
		lo.Pose.Pos.Set(lp.X, lp.Z, lp.Y)
		lo.Pose.Scale.Set(nsc.X*rp.Scale, szc, nsc.Y*rp.Scale)
		lo.Mat.Color.SetUInt8(255, 100, 255, 128)
		lo.Mat.Specular.SetUInt8(128, 128, 128, 255)
		lo.Mat.CullBack = true
		lo.Mat.CullFront = false
		// note: would actually be better to NOT cull back so you can view underneath
		// but then the front and back fight against each other, causing flickering
		// really you ned
	}
	vs.UpdateEnd(updt)
}

// ViewDefaults are the default 3D view params
func (nv *NetView) ViewDefaults() {
	vs := nv.ViewScene()
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
