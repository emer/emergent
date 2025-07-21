// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package paths

import (
	"cogentcore.org/core/math32"
	"cogentcore.org/core/math32/vecint"
	"cogentcore.org/core/tensor"
	"github.com/emer/emergent/v2/edge"
)

// Rect implements a rectangular pattern of connectivity between two layers
// where the lower-left corner moves in proportion to receiver position with offset
// and multiplier factors (with wrap-around optionally).
// 4D layers are automatically flattened to 2D for this pathway.
type Rect struct {

	// size of rectangle in sending layer that each receiving unit receives from
	Size vecint.Vector2i

	// starting offset in sending layer, for computing the corresponding sending lower-left corner relative to given recv unit position
	Start vecint.Vector2i

	// scaling to apply to receiving unit position to compute corresponding position in sending layer of the lower-left corner of rectangle
	Scale math32.Vector2

	// auto-set the Scale as function of the relative sizes of send and recv layers (e.g., if sending layer is 2x larger than receiving, Scale = 2)
	AutoScale bool

	// if true, use Round when applying scaling factor -- otherwise uses Floor which makes Scale work like a grouping factor -- e.g., .25 will effectively group 4 recv units with same send position
	RoundScale bool

	// if true, connectivity wraps around all edges if it would otherwise go off the edge -- if false, then edges are clipped
	Wrap bool

	// if true, and connecting layer to itself (self pathway), then make a self-connection from unit to itself
	SelfCon bool

	// make the reciprocal of the specified connections -- i.e., symmetric for swapping recv and send
	Recip bool

	// starting position in receiving layer -- if > 0 then units below this starting point remain unconnected
	RecvStart vecint.Vector2i

	// number of units in receiving layer to connect -- if 0 then all (remaining after RecvStart) are connected -- otherwise if < remaining then those beyond this point remain unconnected
	RecvN vecint.Vector2i
}

func NewRect() *Rect {
	cr := &Rect{}
	cr.Defaults()
	return cr
}

// NewRectRecip creates a new Rect that is a Recip version of given ff one
func NewRectRecip(ff *Rect) *Rect {
	cr := &Rect{}
	*cr = *ff
	cr.Recip = true
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

func (cr *Rect) Connect(send, recv *tensor.Shape, same bool) (sendn, recvn *tensor.Int32, cons *tensor.Bits) {
	if cr.Recip {
		return cr.ConnectRecip(send, recv, same)
	}
	sendn, recvn, cons = NewTensors(send, recv)
	sNy, sNx, _, _ := tensor.Projection2DShape(send, false)
	rNy, rNx, _, _ := tensor.Projection2DShape(recv, false)

	rnv := recvn.Values
	snv := sendn.Values
	sNtot := send.Len()

	rNyEff := rNy
	if cr.RecvN.Y > 0 {
		rNyEff = min(rNy, cr.RecvN.Y)
	}
	if cr.RecvStart.Y > 0 {
		rNyEff = min(rNyEff, rNy-cr.RecvStart.Y)
	}

	rNxEff := rNx
	if cr.RecvN.X > 0 {
		rNxEff = min(rNx, cr.RecvN.X)
	}
	if cr.RecvStart.X > 0 {
		rNxEff = min(rNxEff, rNx-cr.RecvStart.X)
	}

	sc := cr.Scale
	if cr.AutoScale {
		ssz := math32.Vec2(float32(sNx), float32(sNy))
		rsz := math32.Vec2(float32(rNxEff), float32(rNyEff))
		sc = ssz.Div(rsz)
	}

	for ry := cr.RecvStart.Y; ry < rNyEff+cr.RecvStart.Y; ry++ {
		for rx := cr.RecvStart.X; rx < rNxEff+cr.RecvStart.X; rx++ {
			ri := tensor.Projection2DIndex(recv, false, ry, rx)
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
					si := tensor.Projection2DIndex(send, false, sy, sx)
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

func (cr *Rect) ConnectRecip(send, recv *tensor.Shape, same bool) (sendn, recvn *tensor.Int32, cons *tensor.Bits) {
	sendn, recvn, cons = NewTensors(send, recv)
	sNy, sNx, _, _ := tensor.Projection2DShape(recv, false) // swapped!
	rNy, rNx, _, _ := tensor.Projection2DShape(send, false)

	rnv := recvn.Values
	snv := sendn.Values
	sNtot := send.Len()

	rNyEff := rNy
	if cr.RecvN.Y > 0 {
		rNyEff = min(rNy, cr.RecvN.Y)
	}
	if cr.RecvStart.Y > 0 {
		rNyEff = min(rNyEff, rNy-cr.RecvStart.Y)
	}

	rNxEff := rNx
	if cr.RecvN.X > 0 {
		rNxEff = min(rNx, cr.RecvN.X)
	}
	if cr.RecvStart.X > 0 {
		rNxEff = min(rNxEff, rNx-cr.RecvStart.X)
	}

	sc := cr.Scale
	if cr.AutoScale {
		ssz := math32.Vec2(float32(sNx), float32(sNy))
		rsz := math32.Vec2(float32(rNxEff), float32(rNyEff))
		sc = ssz.Div(rsz)
	}

	for ry := cr.RecvStart.Y; ry < rNyEff+cr.RecvStart.Y; ry++ {
		for rx := cr.RecvStart.X; rx < rNxEff+cr.RecvStart.X; rx++ {
			ri := tensor.Projection2DIndex(send, false, ry, rx)
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
					si := tensor.Projection2DIndex(recv, false, sy, sx)
					off := si*sNtot + ri
					if !cr.SelfCon && same && ri == si {
						continue
					}
					cons.Values.Set(off, true)
					rnv[si]++
					snv[ri]++
				}
			}
		}
	}
	return
}
