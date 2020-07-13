// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import (
	"fmt"

	"github.com/goki/gi/gi"
	"github.com/goki/gi/gi3d"
	"github.com/goki/gi/oswin"
	"github.com/goki/gi/oswin/mouse"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
	"github.com/goki/mat32"
)

// LayObj is the Layer 3D object within the NetView
type LayObj struct {
	gi3d.Solid
	LayName string   `desc:"name of the layer we represent"`
	NetView *NetView `copy:"-" json:"-" xml:"-" view:"-" desc:"our netview"`
}

var KiT_LayObj = kit.Types.AddType(&LayObj{}, LayObjProps)

func (lo *LayObj) ConnectEvents3D(sc *gi3d.Scene) {
	lo.ConnectEvent(sc.Win, oswin.MouseEvent, gi.RegPri, func(recv, send ki.Ki, sig int64, d interface{}) {
		me := d.(*mouse.Event)
		if me.Action != mouse.Press || !sc.IsVisible() {
			return
		}
		// note: could conditionalize on modifier but easier to just always be able to click!
		// if key.HasAllModifierBits(me.Modifiers, key.Shift)
		nii, _ := gi3d.KiToNode3D(recv)
		relpos := me.Where.Sub(sc.ObjBBox.Min)
		ray := nii.RayPick(relpos, sc)
		// layer is in XZ plane with norm pointing up in Y axis
		// offset is 0 in local coordinates
		plane := mat32.Plane{Norm: mat32.Vec3{0, 1, 0}, Off: 0}
		pt, ok := ray.IntersectPlane(plane)
		if !ok || pt.Z > 0 { // Z > 0 means clicked "in front" of plane -- where labels are
			return
		}
		lx := int(pt.X)
		ly := -int(pt.Z)
		// fmt.Printf("selected unit: %v, %v\n", lx, ly)
		if lx < 0 || ly < 0 {
			return
		}
		nv := lo.NetView
		lay := nv.Net.LayerByName(lo.LayName)
		if lay == nil {
			return
		}
		lshp := lay.Shape()
		if lay.Is2D() {
			idx := []int{ly, lx}
			if !lshp.IdxIsValid(idx) {
				return
			}
			nv.Data.PrjnUnIdx = lshp.Offset(idx)
		} else if lay.Is4D() {
			idx, ok := lay.Idx4DFrom2D(lx, ly)
			if !ok {
				return
			}
			nv.Data.PrjnUnIdx = lshp.Offset(idx)
		} else {
			return // not supported
		}
		nv.Data.PrjnLay = lo.LayName
		nv.Record("") // requires new update
		nv.Update()
		me.SetProcessed()
	})
	lo.ConnectEvent(sc.Win, oswin.MouseHoverEvent, gi.RegPri, func(recv, send ki.Ki, sig int64, d interface{}) {
		if !sc.IsVisible() {
			return
		}
		me := d.(*mouse.HoverEvent)
		me.SetProcessed()
		nii, _ := gi3d.KiToNode3D(recv)
		relpos := me.Where.Sub(sc.ObjBBox.Min)
		ray := nii.RayPick(relpos, sc)
		// layer is in XZ plane with norm pointing up in Y axis
		// offset is 0 in local coordinates
		plane := mat32.Plane{Norm: mat32.Vec3{0, 1, 0}, Off: 0}
		pt, ok := ray.IntersectPlane(plane)
		if !ok {
			return
		}
		lx := int(pt.X)
		ly := -int(pt.Z)
		// fmt.Printf("selected unit: %v, %v\n", lx, ly)
		if lx < 0 || ly < 0 {
			return
		}
		nv := lo.NetView
		lay := nv.Net.LayerByName(lo.LayName)
		if lay == nil {
			return
		}
		lshp := lay.Shape()
		sval := ""
		if lay.Is2D() {
			idx := []int{ly, lx}
			if !lshp.IdxIsValid(idx) {
				return
			}
			val, _, _, hasval := nv.UnitVal(lay, idx)
			if !hasval {
				sval = fmt.Sprintf("[%d,%d]=n/a\n", lx, ly)
			} else {
				sval = fmt.Sprintf("[%d,%d]=%g\n", lx, ly, val)
			}
		} else if lay.Is4D() {
			idx, ok := lay.Idx4DFrom2D(lx, ly)
			if !ok {
				return
			}
			val, _, _, hasval := nv.UnitVal(lay, idx)
			if !hasval {
				sval = fmt.Sprintf("[%d,%d][%d,%d]=n/a\n", idx[1], idx[0], idx[3], idx[2])
			} else {
				sval = fmt.Sprintf("[%d,%d][%d,%d]=%g\n", idx[1], idx[0], idx[3], idx[2], val)
			}
		} else {
			return // not supported
		}
		pos := me.Where
		gi.PopupTooltip(sval, pos.X, pos.Y, sc.Win.Viewport, lo.LayName)
	})
}

var LayObjProps = ki.Props{
	"EnumType:Flag": gi.KiT_NodeFlags,
}
