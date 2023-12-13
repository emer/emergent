// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import (
	"fmt"

	"github.com/emer/emergent/v2/emer"
	"goki.dev/gi/v2/gi"
	"goki.dev/gi/v2/xyzv"
	"goki.dev/goosi/events"
	"goki.dev/mat32/v2"
	"goki.dev/xyz"
)

// Scene3D is a Widget for managing the 3D Scene
type Scene3D struct {
	xyzv.Scene3D

	NetView *NetView
}

func (se *Scene3D) OnInit() {
	se.Scene3D.OnInit()
	se.HandleEvents()
}

func (se *Scene3D) HandleEvents() {
	se.On(events.MouseDown, func(e events.Event) {
		pos := se.Geom.ContentBBox.Min
		e.SetLocalOff(e.LocalOff().Add(pos))
		se.MouseDownEvent(e)
		se.SetNeedsRender(true)
	})
	se.On(events.LongHoverStart, func(e events.Event) {
		pos := se.Geom.ContentBBox.Min
		e.SetLocalOff(e.LocalOff().Add(pos))
		se.LongHoverEvent(e)
	})
	se.On(events.SlideMove, func(e events.Event) {
		pos := se.Geom.ContentBBox.Min
		e.SetLocalOff(e.LocalOff().Add(pos))
		se.Scene.SlideMoveEvent(e)
		se.SetNeedsRender(true)
	})
	se.On(events.Scroll, func(e events.Event) {
		pos := se.Geom.ContentBBox.Min
		e.SetLocalOff(e.LocalOff().Add(pos))
		se.Scene.MouseScrollEvent(e.(*events.MouseScroll))
		se.SetNeedsRender(true)
	})
	se.On(events.KeyChord, func(e events.Event) {
		se.Scene.KeyChordEvent(e)
		se.SetNeedsRender(true)
	})
}

func (se *Scene3D) MouseDownEvent(e events.Event) {
	lay, _, _, unIdx := se.LayerUnitAtPoint(e)
	if lay == nil {
		return
	}
	nv := se.NetView
	nv.Data.PrjnUnIdx = unIdx
	nv.Data.PrjnLay = lay.Name()
	nv.UpdateView()
	e.SetHandled()
}

func (se *Scene3D) LongHoverEvent(e events.Event) {
	lay, lx, ly, _ := se.LayerUnitAtPoint(e)
	if lay == nil {
		return
	}
	nv := se.NetView

	sval := ""
	if lay.Is2D() {
		idx := []int{ly, lx}
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
	// TODO: it would be better to use the 2D layer position here
	gi.NewTooltipTextAt(se, sval, e.Pos(), lay.Size().ToPoint()).Run()
	e.SetHandled()
}

func (se *Scene3D) LayerUnitAtPoint(e events.Event) (lay emer.Layer, lx, ly, unIdx int) {
	pos := e.LocalPos()
	sc := se.Scene
	laysGp, err := sc.ChildByNameTry("Layers", 0)
	if err != nil {
		return
	}
	nv := se.NetView
	nmin, nmax := nv.Net.Bounds()
	nsz := nmax.Sub(nmin).Sub(mat32.Vec3{1, 1, 0}).Max(mat32.Vec3{1, 1, 1})
	nsc := mat32.Vec3{1.0 / nsz.X, 1.0 / nsz.Y, 1.0 / nsz.Z}
	szc := mat32.Max(nsc.X, nsc.Y)
	poff := mat32.NewVec3Scalar(0.5)
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
		plane := mat32.Plane{Norm: mat32.Vec3{0, 1, 0}, Off: 0}
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
			if !lshp.IdxIsValid(idx) {
				continue
			}
			unIdx = lshp.Offset(idx)
			return
		} else if lay.Is4D() {
			idx, ok := lay.Idx4DFrom2D(lx, ly)
			if !ok {
				continue
			}
			unIdx = lshp.Offset(idx)
			return
		} else {
			continue // not supported
		}
	}
	lay = nil
	return
}
