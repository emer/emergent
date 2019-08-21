// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import (
	"github.com/goki/gi/gi"
	"github.com/goki/gi/gi3d"
	"github.com/goki/gi/mat32"
	"github.com/goki/gi/oswin"
	"github.com/goki/gi/oswin/mouse"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
)

// LayObj is the Layer 3D object within the NetView
type LayObj struct {
	gi3d.Object
	LayName string   `desc:"name of the layer we represent"`
	NetView *NetView `copy:"-" json:"-" xml:"-" view:"-" desc:"our netview"`
}

var KiT_LayObj = kit.Types.AddType(&LayObj{}, nil)

func (lo *LayObj) ConnectEvents3D(sc *gi3d.Scene) {
	lo.ConnectEvent(sc.Win, oswin.MouseEvent, gi.RegPri, func(recv, send ki.Ki, sig int64, d interface{}) {
		me := d.(*mouse.Event)
		if me.Action != mouse.Press {
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
		lshp := lay.Shape()
		if lay.Is2D() {
			idx := []int{ly, lx}
			if !lshp.IdxIsValid(idx) {
				return
			}
			nv.PrjnUnIdx = lshp.Offset(idx)
		} else if lay.Is4D() {
			idx, ok := lay.Idx4DFrom2D(lx, ly)
			if !ok {
				return
			}
			nv.PrjnUnIdx = lshp.Offset(idx)
		} else {
			return // not supported
		}
		nv.PrjnLay = lo.LayName
		nv.Update("")
		me.SetProcessed()
	})
}
