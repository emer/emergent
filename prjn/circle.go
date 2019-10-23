// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prjn

import (
	"github.com/emer/emergent/evec"
	"github.com/emer/etable/etensor"
	"github.com/goki/gi/mat32"
)

// Circle implements a circular pattern of connectivity between two layers
// where the center moves in proportion to receiver position with offset
// and multiplier factors, and a given radius is used (with wrap-around
// optionally).  A corresponding Gaussian bump of TopoWts is available as well.
// Makes for a good center-surround connectivity pattern.
// 4D layers are automatically flattened to 2D for this connection.
type Circle struct {
	Radius    int        `desc:"radius of the circle, in units from center in sending layer"`
	Start     evec.Vec2i `desc:"starting offset in sending layer, for computing the corresponding sending center relative to given recv unit position"`
	Scale     mat32.Vec2 `desc:"scaling to apply to receiving unit position to compute sending center as function of recv unit position"`
	AutoScale bool       `desc:"auto-scale sending center positions as function of relative sizes of send and recv layers -- if Start is positive then assumes it is a border, subtracted from sending size"`
	SelfCon   bool       `desc:"if true, and connecting layer to itself (self projection), then make a self-connection from unit to itself"`
}

func NewCircle() *Circle {
	cr := &Circle{}
	cr.Defaults()
	return cr
}

func (cr *Circle) Defaults() {
	if cr.Radius == 0 {
		cr.Radius = 8
	}
	if cr.Scale.X == 0 || cr.Scale.Y == 0 {
		cr.Scale.SetScalar(1)
	}
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
		ssz := mat32.Vec2{float32(sNx - 2*cr.Start.X), float32(sNy - 2*cr.Start.Y)}
		rsz := mat32.Vec2{float32(rNx), float32(rNy)}
		sc = ssz.Div(rsz)
	}

	for ry := 0; ry < rNy; ry++ {
		for rx := 0; rx < rNx; rx++ {
			sctr := mat32.Vec2{float32(rx)*sc.X + float32(cr.Start.X), float32(ry)*sc.Y + float32(cr.Start.Y)}
			for sy := 0; sy < sNy; sy++ {
				for sx := 0; sx < sNx; sx++ {
					sp := mat32.Vec2{float32(sx), float32(sy)}
					sp.X = WrapMinDist(sp.X, float32(sNx-1), sctr.X)
					sp.Y = WrapMinDist(sp.Y, float32(sNy-1), sctr.Y)
					d := int(mat32.Round(sp.DistTo(sctr)))
					if d <= cr.Radius {
						ri := ry*rNx + rx
						si := sy*sNx + sx
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
