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

// PoolRect implements a rectangular pattern of connectivity between
// two 4D layers, in terms of their pool-level shapes,
// where the lower-left corner moves in proportion to receiver
// pool position with offset and multiplier factors (with wrap-around optionally).
type PoolRect struct {
	Size       evec.Vec2i `desc:"size of rectangle (of pools) in sending layer that each receiving unit receives from"`
	Start      evec.Vec2i `desc:"starting pool offset in sending layer, for computing the corresponding sending lower-left corner relative to given recv pool position"`
	Scale      mat32.Vec2 `desc:"scaling to apply to receiving pool osition to compute corresponding position in sending layer of the lower-left corner of rectangle"`
	AutoScale  bool       `desc:"auto-set the Scale as function of the relative pool sizes of send and recv layers (e.g., if sending layer is 2x larger than receiving, Scale = 2)"`
	RoundScale bool       `desc:"if true, use Round when applying scaling factor -- otherwise uses Floor which makes Scale work like a grouping factor -- e.g., .25 will effectively group 4 recv pools with same send position"`
	Wrap       bool       `desc:"if true, connectivity wraps around all edges if it would otherwise go off the edge -- if false, then edges are clipped"`
	SelfCon    bool       `desc:"if true, and connecting layer to itself (self projection), then make a self-connection from unit to itself"`
	RecvStart  evec.Vec2i `desc:"starting pool position in receiving layer -- if > 0 then pools below this starting point remain unconnected"`
	RecvN      evec.Vec2i `desc:"number of pools in receiving layer to connect -- if 0 then all (remaining after RecvStart) are connected -- otherwise if < remaining then those beyond this point remain unconnected"`
}

func NewPoolRect() *PoolRect {
	cr := &PoolRect{}
	cr.Defaults()
	return cr
}

func (cr *PoolRect) Defaults() {
	cr.Wrap = true
	cr.Size.Set(1, 1)
	cr.Scale.SetScalar(1)
}

func (cr *PoolRect) Name() string {
	return "PoolRect"
}

func (cr *PoolRect) Connect(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	sendn, recvn, cons = NewTensors(send, recv)
	sNy := send.Dim(0)
	sNx := send.Dim(1)
	rNy := recv.Dim(0)
	rNx := recv.Dim(1)

	sNn := 1
	rNn := 1

	if send.NumDims() == 4 {
		sNn = send.Dim(2) * send.Dim(3)
	} else { // 2D
		sNn = sNy * sNx
		sNy = 1
		sNx = 1
	}
	if recv.NumDims() == 4 {
		rNn = recv.Dim(2) * recv.Dim(3)
	} else { // 2D
		rNn = rNy * rNx
		rNy = 1
		rNx = 1
	}

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
			rpi := ry*rNx + rx
			ris := rpi * rNn
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
					spi := sy*sNx + sx
					sis := spi * sNn

					for r := 0; r < rNn; r++ {
						ri := ris + r
						for s := 0; s < sNn; s++ {
							si := sis + s
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
	}
	return
}
