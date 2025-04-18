// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package paths

import (
	"cogentcore.org/core/math32"
	"cogentcore.org/core/math32/vecint"
	"cogentcore.org/lab/tensor"
	"github.com/emer/emergent/v2/edge"
)

// PoolRect implements a rectangular pattern of connectivity between
// two 4D layers, in terms of their pool-level shapes,
// where the lower-left corner moves in proportion to receiver
// pool position with offset and multiplier factors (with wrap-around optionally).
type PoolRect struct {

	// size of rectangle (of pools) in sending layer that each receiving unit receives from
	Size vecint.Vector2i

	// starting pool offset in sending layer, for computing the corresponding sending lower-left corner relative to given recv pool position
	Start vecint.Vector2i

	// scaling to apply to receiving pool osition to compute corresponding position in sending layer of the lower-left corner of rectangle
	Scale math32.Vector2

	// auto-set the Scale as function of the relative pool sizes of send and recv layers (e.g., if sending layer is 2x larger than receiving, Scale = 2)
	AutoScale bool

	// if true, use Round when applying scaling factor -- otherwise uses Floor which makes Scale work like a grouping factor -- e.g., .25 will effectively group 4 recv pools with same send position
	RoundScale bool

	// if true, connectivity wraps around all edges if it would otherwise go off the edge -- if false, then edges are clipped
	Wrap bool

	// if true, and connecting layer to itself (self pathway), then make a self-connection from unit to itself
	SelfCon bool

	// starting pool position in receiving layer -- if > 0 then pools below this starting point remain unconnected
	RecvStart vecint.Vector2i

	// number of pools in receiving layer to connect -- if 0 then all (remaining after RecvStart) are connected -- otherwise if < remaining then those beyond this point remain unconnected
	RecvN vecint.Vector2i
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

func (cr *PoolRect) Connect(send, recv *tensor.Shape, same bool) (sendn, recvn *tensor.Int32, cons *tensor.Bool) {
	sendn, recvn, cons = NewTensors(send, recv)
	sNy := send.DimSize(0)
	sNx := send.DimSize(1)
	rNy := recv.DimSize(0)
	rNx := recv.DimSize(1)

	sNn := 1
	rNn := 1

	if send.NumDims() == 4 {
		sNn = send.DimSize(2) * send.DimSize(3)
	} else { // 2D
		sNn = sNy * sNx
		sNy = 1
		sNx = 1
	}
	if recv.NumDims() == 4 {
		rNn = recv.DimSize(2) * recv.DimSize(3)
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
		ssz := math32.Vec2(float32(sNx), float32(sNy))
		rsz := math32.Vec2(float32(rNx), float32(rNy))
		sc = ssz.Div(rsz)
	}

	rNyEff := rNy
	if cr.RecvN.Y > 0 {
		rNyEff = min(rNy, cr.RecvStart.Y+cr.RecvN.Y)
	}
	rNxEff := rNx
	if cr.RecvN.X > 0 {
		rNxEff = min(rNx, cr.RecvStart.X+cr.RecvN.X)
	}

	for ry := cr.RecvStart.Y; ry < rNyEff; ry++ {
		for rx := cr.RecvStart.X; rx < rNxEff; rx++ {
			rpi := ry*rNx + rx
			ris := rpi * rNn
			sst := cr.Start
			if cr.RoundScale {
				sst.X += int(math32.Round(float32(rx-cr.RecvStart.X) * sc.X))
				sst.Y += int(math32.Round(float32(ry-cr.RecvStart.Y) * sc.Y))
			} else {
				sst.X += int(math32.Floor(float32(rx-cr.RecvStart.X) * sc.X))
				sst.Y += int(math32.Floor(float32(ry-cr.RecvStart.Y) * sc.Y))
			}
			for y := 0; y < cr.Size.Y; y++ {
				sy, clipy := edge.Edge(sst.Y+y, sNy, cr.Wrap)
				if clipy {
					continue
				}
				for x := 0; x < cr.Size.X; x++ {
					sx, clipx := edge.Edge(sst.X+x, sNx, cr.Wrap)
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
							cons.Values.Set(true, off)
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
