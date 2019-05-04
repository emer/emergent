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

var KiT_NetView = kit.Types.AddType(&NetView{}, nil) // todo props

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
	if !nv.HasLayers() {
		return
	}
	vs := nv.ViewScene()
	vs.UpdateMeshes() // todo: actually call Update() on all layermesh
	vs.Render()
	vs.Win.DirectUpdate(vs)
	vs.Win.Publish()
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
	nv.VarsListUpdate()
	if len(nv.Vars) == 0 {
		vl.DeleteChildren(true)
		return
	}
	config := kit.TypeAndNameList{}
	for _, vn := range nv.Vars {
		config.Add(gi.KiT_Button, vn)
	}
	mods, updt := vl.ConfigChildren(config, false)
	if !mods {
		updt = vl.UpdateStart()
	}
	for i, vbi := range *vl.Children() {
		vb := vbi.(*gi.Button)
		vn := nv.Vars[i]
		vb.SetText(vn)
		vb.ButtonSig.Connect(nv.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
			if sig == int64(gi.ButtonClicked) {
				nvv := recv.Embed(KiT_NetView).(*NetView)
				vbv := send.(*gi.Button)
				nvv.SetVar(vbv.Text)
			}
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
	yht := 1.0 / float32(nlay+1)
	for li, loi := range *vs.Children() {
		lay := nv.Net.Layer(li)
		lo := loi.(*gi3d.Object)
		lo.Defaults()
		lo.SetMesh(vs, lay.Name())
		lo.Pose.Pos.Y = float32(li) / float32(nlay)
		lo.Pose.Scale.Y = yht
		lo.Mat.Color.SetUInt8(255, 100, 255, 128)
		lo.Mat.Specular.SetUInt8(128, 128, 128, 255)
	}
	vs.UpdateEnd(updt)
}

// ViewDefaults are the default 3D view params
func (nv *NetView) ViewDefaults() {
	vs := nv.ViewScene()
	vs.SetStretchMaxWidth()
	vs.SetStretchMaxHeight()
	vs.Defaults()
	vs.Camera.Pose.Pos.Set(-.2, 1, 3)
	vs.Camera.LookAt(mat32.Vec3{.1, .3, 0}, mat32.Vec3{0, 1, 0})
	vs.BgColor.SetUInt8(255, 255, 255, 255) // white
	gi3d.AddNewAmbientLight(vs, "ambient", 0.2, gi3d.DirectSun)
	dir := gi3d.AddNewDirLight(vs, "dir", 1, gi3d.DirectSun)
	dir.Pos.Set(0, 0, 1)
}

// UnitVal returns the raw value, scaled value, and color representation for given unit of given layer
// scaled is in range -0.5..0.5
// todo: could incorporate history etc..
func (nv *NetView) UnitVal(lay emer.Layer, idx []int) (raw, scaled float32, clr gi.Color) {
	raw, _ = lay.UnitVal(nv.Var, idx)
	scaled = mat32.Clamp(0.5*raw, -0.5, 0.5)
	if scaled < 0 {
		clr.R = uint8(50.0)
		clr.G = uint8(50.0)
		clr.B = uint8(50.0 - 2*scaled*205.0)
		clr.A = uint8(128.0 - 2*scaled*127.0)
	} else {
		clr.R = uint8(50.0 + 2*scaled*205.0)
		clr.G = uint8(50.0)
		clr.B = uint8(50.0)
		clr.A = uint8(128.0 + 2*scaled*127.0)
	}
	return
}

// func (nv *NetView) Render2D() {
// 	if nv.FullReRenderIfNeeded() {
// 		return
// 	}
// 	if nv.PushBounds() {
// 		nv.Frame.Render2D()
// 		updt := nv.UpdateStart()
// 		nv.Update()
// 		nv.UpdateEndNoSig(updt)
// 		nv.PopBounds()
// 	}
// }
