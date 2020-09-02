// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prjn

import (
	"github.com/emer/emergent/evec"
	"github.com/emer/etable/etensor"
	"github.com/goki/ki/ints"
	"github.com/goki/mat32"
)

// Rect implements a rectangular pattern of connectivity between two layers
// where the lower-left corner moves in proportion to receiver position with offset
// and multiplier factors (with wrap-around optionally).
// 4D layers are automatically flattened to 2D for this projection.
type Rect struct {
	Size       evec.Vec2i `desc:"size of rectangle in sending layer that each receiving unit receives from"`
	Start      evec.Vec2i `desc:"starting offset in sending layer, for computing the corresponding sending lower-left corner relative to given recv unit position"`
	Scale      mat32.Vec2 `desc:"scaling to apply to receiving unit position to compute corresponding position in sending layer of the lower-left corner of rectangle"`
	AutoScale  bool       `desc:"auto-set the Scale as function of the relative sizes of send and recv layers (e.g., if sending layer is 2x larger than receiving, Scale = 2)"`
	RoundScale bool       `desc:"if true, use Round when applying scaling factor -- otherwise uses Floor which makes Scale work like a grouping factor -- e.g., .25 will effectively group 4 recv units with same send position"`
	Wrap       bool       `desc:"if true, connectivity wraps around all edges if it would otherwise go off the edge -- if false, then edges are clipped"`
	SelfCon    bool       `desc:"if true, and connecting layer to itself (self projection), then make a self-connection from unit to itself"`
	RecvStart  evec.Vec2i `desc:"starting position in receiving layer -- if > 0 then units below this starting point remain unconnected"`
	RecvN      evec.Vec2i `desc:"number of units in receiving layer to connect -- if 0 then all (remaining after RecvStart) are connected -- otherwise if < remaining then those beyond this point remain unconnected"`
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

	rNyEff := rNy
	if cr.RecvN.Y > 0 {
		rNyEff = ints.MinInt(rNy, cr.RecvStart.Y+cr.RecvN.Y)
	}
	rNxEff := rNx
	if cr.RecvN.X > 0 {
		rNxEff = ints.MinInt(rNx, cr.RecvStart.X+cr.RecvN.X)
	}

	for ry := cr.RecvStart.Y; ry < rNyEff; ry++ {
		for rx := cr.RecvStart.X; rx < rNxEff; rx++ {
			ri := etensor.Prjn2DIdx(recv, false, ry, rx)
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
