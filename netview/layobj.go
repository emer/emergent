// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import (
	"fmt"

	"github.com/goki/gi/gi"
	"github.com/goki/gi/gi3d"
	"github.com/goki/gi/mat32"
	"github.com/goki/gi/oswin"
	"github.com/goki/gi/oswin/key"
	"github.com/goki/gi/oswin/mouse"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
)

// LayObj is the Layer 3D object
type LayObj struct {
	gi3d.Object
}

var KiT_LayObj = kit.Types.AddType(&LayObj{}, nil)

func (lo *LayObj) ConnectEvents3D(sc *gi3d.Scene) {
	lo.ConnectEvent(sc.Win, oswin.MouseEvent, gi.RegPri, func(recv, send ki.Ki, sig int64, d interface{}) {
		me := d.(*mouse.Event)
		if me.Action != mouse.Press {
			return
		}
		nii, _ := gi3d.KiToNode3D(recv)
		relpos := me.Where.Sub(sc.WinBBox.Min)
		switch {
		case key.HasAllModifierBits(me.Modifiers, key.Shift):
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
			fmt.Printf("selected unit: %v, %v\n", lx, ly)
			me.SetProcessed()
		}
	})
}
