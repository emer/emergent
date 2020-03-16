// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prjn

import (
	"github.com/emer/emergent/evec"
	"github.com/emer/etable/etensor"
	"github.com/goki/mat32"
)

// Rect implements a rectangular pattern of connectivity between two layers
// where the lower-left corner moves in proportion to receiver position with offset
// and multiplier factors (with wrap-around optionally).
// 4D layers are automatically flattened to 2D for this connection.
type Rect struct {
	Start      evec.Vec2i `desc:"starting offset in sending layer, for computing the corresponding sending lower-left corner relative to given recv unit position"`
	Size       evec.Vec2i `desc:"size of rectangle"`
	Scale      mat32.Vec2 `desc:"scaling to apply to receiving unit position to compute corresponding position in sending layer"`
	AutoScale  bool       `desc:"auto-scale sending positions as function of relative sizes of send and recv layers"`
	Wrap       bool       `desc:"if true, connectivity wraps around edges"`
	SelfCon    bool       `desc:"if true, and connecting layer to itself (self projection), then make a self-connection from unit to itself"`
	RoundScale bool       `desc:"if true, use Round when applying scaling factor -- otherwise uses Floor which makes Scale work like a grouping factor -- e.g., .25 will effectively group 4 recv units with same send position"`
}

func NewRect() *Rect {
	cr := &Rect{}
	cr.Defaults()
	return cr
}

func (cr *Rect) Defaults() {
	cr.Wrap = true
	cr.Size.Set(2, 2)
	cr.Scale.SetScalar(1)
}

func (cr *Rect) Name() string {
	return "Rect"
}

func (cr *Rect) Connect(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	sendn, recvn, cons = NewTensors(send, recv)
	sNy, sNx, _, _ := etensor.Prjn2DShape(send, false)
	rNy, rNx, _, _ := etensor.Prjn2DShape(recv, false)

	rnv := recvn.Values
	snv := sendn.Values
	sNtot := send.Len()

	sc := cr.Scale
	if cr.AutoScale {
		ssz := mat32.Vec2{float32(sNx), float32(sNy)}
		rsz := mat32.Vec2{float32(rNx), float32(rNy)}
		sc = ssz.Div(rsz)
	}

	for ry := 0; ry < rNy; ry++ {
		for rx := 0; rx < rNx; rx++ {
			sst := cr.Start
			if cr.RoundScale {
				sst.X += int(mat32.Round(float32(rx) * sc.X))
				sst.Y += int(mat32.Round(float32(ry) * sc.Y))
			} else {
				sst.X += int(mat32.Floor(float32(rx) * sc.X))
				sst.Y += int(mat32.Floor(float32(ry) * sc.Y))
			}
			for y := 0; y < cr.Size.Y; y++ {
				sy, clipy := Edge(sst.Y+y, sNy, cr.Wrap)
				if clipy {
					continue
				}
				for x := 0; x < cr.Size.X; x++ {
					sx, clipx := Edge(sst.X+x, sNx, cr.Wrap)
					if clipx {
						continue
					}
					ri := etensor.Prjn2DIdx(recv, false, ry, rx)
					si := etensor.Prjn2DIdx(send, false, sy, sx)
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
	return
}
