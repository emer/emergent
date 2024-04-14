// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prjn

import (
	"cogentcore.org/core/math32"
	"github.com/emer/emergent/v2/edge"
	"github.com/emer/emergent/v2/efuns"
	"github.com/emer/emergent/v2/evec"
	"github.com/emer/etable/v2/etensor"
)

// Circle implements a circular pattern of connectivity between two layers
// where the center moves in proportion to receiver position with offset
// and multiplier factors, and a given radius is used (with wrap-around
// optionally).  A corresponding Gaussian bump of TopoWts is available as well.
// Makes for a good center-surround connectivity pattern.
// 4D layers are automatically flattened to 2D for this connection.
type Circle struct {

	// radius of the circle, in units from center in sending layer
	Radius int

	// starting offset in sending layer, for computing the corresponding sending center relative to given recv unit position
	Start evec.Vector2i

	// scaling to apply to receiving unit position to compute sending center as function of recv unit position
	Scale math32.Vector2

	// auto-scale sending center positions as function of relative sizes of send and recv layers -- if Start is positive then assumes it is a border, subtracted from sending size
	AutoScale bool

	// if true, connectivity wraps around edges
	Wrap bool

	// if true, this prjn should set gaussian topographic weights, according to following parameters
	TopoWts bool

	// gaussian sigma (width) as a proportion of the radius of the circle
	Sigma float32

	// maximum weight value for GaussWts function -- multiplies values
	MaxWt float32

	// if true, and connecting layer to itself (self projection), then make a self-connection from unit to itself
	SelfCon bool
}

func NewCircle() *Circle {
	cr := &Circle{}
	cr.Defaults()
	return cr
}

func (cr *Circle) Defaults() {
	cr.Wrap = true
	cr.Radius = 8
	cr.Scale.SetScalar(1)
	cr.Sigma = 0.5
	cr.MaxWt = 1
}

func (cr *Circle) Name() string {
	return "Circle"
}

func (cr *Circle) Connect(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	sendn, recvn, cons = NewTensors(send, recv)
	sNy, sNx, _, _ := etensor.Prjn2DShape(send, false)
	rNy, rNx, _, _ := etensor.Prjn2DShape(recv, false)

	rnv := recvn.Values
	snv := sendn.Values
	sNtot := send.Len()

	sc := cr.Scale
	if cr.AutoScale {
		ssz := math32.Vec2(float32(sNx), float32(sNy))
		if cr.Start.X >= 0 && cr.Start.Y >= 0 {
			ssz.X -= float32(2 * cr.Start.X)
			ssz.Y -= float32(2 * cr.Start.Y)
		}
		rsz := math32.Vec2(float32(rNx), float32(rNy))
		sc = ssz.Div(rsz)
	}

	for ry := 0; ry < rNy; ry++ {
		for rx := 0; rx < rNx; rx++ {
			sctr := math32.Vec2(float32(rx)*sc.X+float32(cr.Start.X), float32(ry)*sc.Y+float32(cr.Start.Y))
			for sy := 0; sy < sNy; sy++ {
				for sx := 0; sx < sNx; sx++ {
					sp := math32.Vec2(float32(sx), float32(sy))
					if cr.Wrap {
						sp.X = edge.WrapMinDist(sp.X, float32(sNx), sctr.X)
						sp.Y = edge.WrapMinDist(sp.Y, float32(sNy), sctr.Y)
					}
					d := int(math32.Round(sp.DistTo(sctr)))
					if d <= cr.Radius {
						ri := etensor.Prjn2DIndex(recv, false, ry, rx)
						si := etensor.Prjn2DIndex(send, false, sy, sx)
						off := ri*sNtot + si
						if !cr.SelfCon && same && ri == si {
							continue
						}
						cons.Values.Set(off, true)
						rnv[ri]++
						snv[si]++
					}
				}
			}
		}
	}
	return
}

// GaussWts returns gaussian weight value for given unit indexes in
// given send and recv layers according to Gaussian Sigma and MaxWt.
// Can be used for a Prjn.SetScalesFunc or SetWtsFunc
func (cr *Circle) GaussWts(si, ri int, send, recv *etensor.Shape) float32 {
	sNy, sNx, _, _ := etensor.Prjn2DShape(send, false)
	rNy, rNx, _, _ := etensor.Prjn2DShape(recv, false)

	ry := ri / rNx // todo: this is not right for 4d!
	rx := ri % rNx
	sy := si / sNx
	sx := si % sNx

	fsig := cr.Sigma * float32(cr.Radius)

	sc := cr.Scale
	if cr.AutoScale {
		ssz := math32.Vec2(float32(sNx), float32(sNy))
		if cr.Start.X >= 0 && cr.Start.Y >= 0 {
			ssz.X -= float32(2 * cr.Start.X)
			ssz.Y -= float32(2 * cr.Start.Y)
		}
		rsz := math32.Vec2(float32(rNx), float32(rNy))
		sc = ssz.Div(rsz)
	}
	sctr := math32.Vec2(float32(rx)*sc.X+float32(cr.Start.X), float32(ry)*sc.Y+float32(cr.Start.Y))
	sp := math32.Vec2(float32(sx), float32(sy))
	if cr.Wrap {
		sp.X = edge.WrapMinDist(sp.X, float32(sNx), sctr.X)
		sp.Y = edge.WrapMinDist(sp.Y, float32(sNy), sctr.Y)
	}
	wt := cr.MaxWt * efuns.GaussVecDistNoNorm(sp, sctr, fsig)
	return wt
}
