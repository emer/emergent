// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import (
	"fmt"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/views"
	"cogentcore.org/core/xyz"
	"cogentcore.org/core/xyzview"
	"github.com/emer/emergent/v2/emer"
)

// Scene is a Widget for managing the 3D Scene of the NetView
type Scene struct {
	xyzview.Scene

	NetView *NetView
}

func (sw *Scene) OnInit() {
	sw.Scene.OnInit()
	sw.HandleEvents()
}

func (sw *Scene) HandleEvents() {
	sw.On(events.MouseDown, func(e events.Event) {
		pos := sw.Geom.ContentBBox.Min
		e.SetLocalOff(e.LocalOff().Add(pos))
		sw.MouseDownEvent(e)
		sw.NeedsRender()
	})
	sw.On(events.LongHoverStart, func(e events.Event) {
		pos := sw.Geom.ContentBBox.Min
		e.SetLocalOff(e.LocalOff().Add(pos))
		sw.LongHoverEvent(e)
	})
	sw.On(events.Scroll, func(e events.Event) {
		pos := sw.Geom.ContentBBox.Min
		e.SetLocalOff(e.LocalOff().Add(pos))
		sw.SceneXYZ().MouseScrollEvent(e.(*events.MouseScroll))
		sw.NeedsRender()
	})
	sw.On(events.KeyChord, func(e events.Event) {
		sw.SceneXYZ().KeyChordEvent(e)
		sw.NeedsRender()
	})
	sw.HandleSlideEvents()
}

func (sw *Scene) MouseDownEvent(e events.Event) {
	ns := xyz.NodesUnderPoint(sw.SceneXYZ(), e.Pos())
	for _, n := range ns {
		ln, ok := n.(*LayName)
		if ok {
			lay := ln.NetView.Net.LayerByName(ln.Text)
			if lay != nil {
				views.StructViewDialog(sw, lay, "Layer: "+lay.Name(), true)
			}
			e.SetHandled()
			return
		}
	}

	lay, _, _, unIndex := sw.LayerUnitAtPoint(e)
	if lay == nil {
		return
	}
	nv := sw.NetView
	nv.Data.PrjnUnIndex = unIndex
	nv.Data.PrjnLay = lay.Name()
	nv.UpdateView()
	e.SetHandled()
}

func (sw *Scene) LongHoverEvent(e events.Event) {
	lay, lx, ly, _ := sw.LayerUnitAtPoint(e)
	if lay == nil {
		return
	}
	nv := sw.NetView

	sval := ""
	if lay.Is2D() {
		idx := []int{ly, lx}
		val, _, _, hasval := nv.UnitValue(lay, idx)
		if !hasval {
			sval = fmt.Sprintf("[%d,%d]=n/a\n", lx, ly)
		} else {
			sval = fmt.Sprintf("[%d,%d]=%g\n", lx, ly, val)
		}
	} else if lay.Is4D() {
		idx, ok := lay.Index4DFrom2D(lx, ly)
		if !ok {
			return
		}
		val, _, _, hasval := nv.UnitValue(lay, idx)
		if !hasval {
			sval = fmt.Sprintf("[%d,%d][%d,%d]=n/a\n", idx[1], idx[0], idx[3], idx[2])
		} else {
			sval = fmt.Sprintf("[%d,%d][%d,%d]=%g\n", idx[1], idx[0], idx[3], idx[2], val)
		}
	} else {
		return // not supported
	}
	core.NewTooltipTextAt(sw, sval, e.WindowPos(), lay.Size().ToPoint()).Run()
	e.SetHandled()
}

func (sw *Scene) LayerUnitAtPoint(e events.Event) (lay emer.Layer, lx, ly, unIndex int) {
	pos := e.Pos()
	sc := sw.SceneXYZ()
	laysGp := sc.ChildByName("Layers", 0)
	if laysGp == nil {
		return
	}
	nv := sw.NetView
	nmin, nmax := nv.Net.Bounds()
	nsz := nmax.Sub(nmin).Sub(math32.Vec3(1, 1, 0)).Max(math32.Vec3(1, 1, 1))
	nsc := math32.Vec3(1.0/nsz.X, 1.0/nsz.Y, 1.0/nsz.Z)
	szc := math32.Max(nsc.X, nsc.Y)
	poff := math32.Vector3Scalar(0.5)
	poff.Y = -0.5
	for li, lgi := range *laysGp.Children() {
		lay = nv.Net.Layer(li)
		lg := lgi.(*xyz.Group)
		lp := lay.Pos()
		lp.Y = -lp.Y // reverse direction
		lp = lp.Sub(nmin).Mul(nsc).Sub(poff)
		rp := lay.RelPos()
		lg.Pose.Pos.Set(lp.X, lp.Z, lp.Y)
		lg.Pose.Scale.Set(nsc.X*rp.Scale, szc, nsc.Y*rp.Scale)
		lo := lg.Child(0).(*LayObj)
		ray := lo.RayPick(pos)
		// layer is in XZ plane with norm pointing up in Y axis
		// offset is 0 in local coordinates
		plane := math32.Plane{Norm: math32.Vec3(0, 1, 0), Off: 0}
		pt, ok := ray.IntersectPlane(plane)
		if !ok || pt.Z > 0 { // Z > 0 means clicked "in front" of plane -- where labels are
			continue
		}
		lx = int(pt.X)
		ly = -int(pt.Z)
		// fmt.Printf("selected unit: %v, %v\n", lx, ly)
		if lx < 0 || ly < 0 {
			continue
		}
		lshp := lay.Shape()
		if lay.Is2D() {
			idx := []int{ly, lx}
			if !lshp.IndexIsValid(idx) {
				continue
			}
			unIndex = lshp.Offset(idx)
			return
		} else if lay.Is4D() {
			idx, ok := lay.Index4DFrom2D(lx, ly)
			if !ok {
				continue
			}
			unIndex = lshp.Offset(idx)
			return
		} else {
			continue // not supported
		}
	}
	lay = nil
	return
}
