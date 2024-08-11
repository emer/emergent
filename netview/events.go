// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import (
	"fmt"
	"image"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/xyz"
	"cogentcore.org/core/xyz/xyzcore"
	"github.com/emer/emergent/v2/emer"
)

// Scene is a Widget for managing the 3D Scene of the NetView
type Scene struct {
	xyzcore.Scene

	NetView *NetView
}

func (sw *Scene) Init() {
	sw.Scene.Init()
	sw.On(events.MouseDown, func(e events.Event) {
		sw.MouseDownEvent(e)
		sw.NeedsRender()
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
	// sw.HandleSlideEvents() // TODO: need this
}

func (sw *Scene) MouseDownEvent(e events.Event) {
	pos := e.Pos().Sub(sw.Geom.ContentBBox.Min)
	ns := xyz.NodesUnderPoint(sw.SceneXYZ(), pos)
	for _, n := range ns {
		ln, ok := n.(*LayName)
		if ok {
			lay, _ := ln.NetView.Net.AsEmer().EmerLayerByName(ln.Text)
			if lay != nil {
				FormDialog(sw, lay, "Layer: "+lay.StyleName())
			}
			e.SetHandled()
			return
		}
	}

	lay, _, _, unIndex := sw.LayerUnitAtPoint(e.Pos())
	if lay == nil {
		return
	}
	nv := sw.NetView
	nv.Data.PathUnIndex = unIndex
	nv.Data.PathLay = lay.StyleName()
	nv.UpdateView()
	e.SetHandled()
}

func (sw *Scene) WidgetTooltip(pos image.Point) (string, image.Point) {
	if pos == image.Pt(-1, -1) {
		return "_", image.Point{}
	}

	lay, lx, ly, _ := sw.LayerUnitAtPoint(pos)
	if lay == nil {
		return "", pos
	}
	lb := lay.AsEmer()
	nv := sw.NetView

	tt := ""
	if lb.Is2D() {
		idx := []int{ly, lx}
		val, _, _, hasval := nv.UnitValue(lay, idx)
		if !hasval {
			tt = fmt.Sprintf("[%d,%d]=n/a\n", lx, ly)
		} else {
			tt = fmt.Sprintf("[%d,%d]=%g\n", lx, ly, val)
		}
	} else if lb.Is4D() {
		idx, ok := lb.Index4DFrom2D(lx, ly)
		if !ok {
			return "", pos
		}
		val, _, _, hasval := nv.UnitValue(lay, idx)
		if !hasval {
			tt = fmt.Sprintf("[%d,%d][%d,%d]=n/a\n", idx[1], idx[0], idx[3], idx[2])
		} else {
			tt = fmt.Sprintf("[%d,%d][%d,%d]=%g\n", idx[1], idx[0], idx[3], idx[2], val)
		}
	} else {
		return "", pos // not supported
	}
	return tt, pos
}

func (sw *Scene) LayerUnitAtPoint(pos image.Point) (lay emer.Layer, lx, ly, unIndex int) {
	pos = pos.Sub(sw.Geom.ContentBBox.Min)
	sc := sw.SceneXYZ()
	laysGpi := sc.ChildByName("Layers", 0)
	if laysGpi == nil {
		return
	}
	_, laysGp := xyz.AsNode(laysGpi)
	nv := sw.NetView
	nb := nv.Net.AsEmer()
	nmin, nmax := nb.MinPos, nb.MaxPos
	nsz := nmax.Sub(nmin).Sub(math32.Vec3(1, 1, 0)).Max(math32.Vec3(1, 1, 1))
	nsc := math32.Vec3(1.0/nsz.X, 1.0/nsz.Y, 1.0/nsz.Z)
	szc := math32.Max(nsc.X, nsc.Y)
	poff := math32.Vector3Scalar(0.5)
	poff.Y = -0.5
	for li, lgi := range laysGp.Children {
		lay = nv.Net.EmerLayer(li)
		lb := lay.AsEmer()
		lg := lgi.(*xyz.Group)
		lp := lb.Pos.Pos
		lp.Y = -lp.Y // reverse direction
		lp = lp.Sub(nmin).Mul(nsc).Sub(poff)
		lg.Pose.Pos.Set(lp.X, lp.Z, lp.Y)
		lg.Pose.Scale.Set(nsc.X*lb.Pos.Scale, szc, nsc.Y*lb.Pos.Scale)
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
		lshp := lb.Shape
		if lb.Is2D() {
			idx := []int{ly, lx}
			if !lshp.IndexIsValid(idx) {
				continue
			}
			unIndex = lshp.Offset(idx)
			return
		} else if lb.Is4D() {
			idx, ok := lb.Index4DFrom2D(lx, ly)
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

// FormDialog opens a dialog in a new, separate window
// for viewing / editing the given struct object, in
// the context of the given ctx widget.
func FormDialog(ctx core.Widget, v any, title string) {
	d := core.NewBody().AddTitle(title)
	core.NewForm(d).SetStruct(v)
	if tb, ok := v.(core.ToolbarMaker); ok {
		d.AddAppBar(tb.MakeToolbar)
	}
	d.RunWindowDialog(ctx)
}
