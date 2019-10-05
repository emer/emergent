// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prjn

import (
	"github.com/emer/emergent/evec"
	"github.com/emer/etable/etensor"
)

// PoolTile implements tiled 2D connectivity between pools within layers, where
// a 2D rectangular receptive field (defined over pools, not units) is tiled
// across the sending layer pools, with specified level of overlap.
// Pools are the outer-most two dimensions of a 4D layer shape, and both layers
// must have pools.
// This is a standard form of convolutional connectivity, where pools are
// the filters and the outer dims are locations filtered.
// Various initial weight / scaling patterns are also available -- code
// must specifically apply these to the receptive fields.
type PoolTile struct {
	Size  evec.Vec2i `desc:"size of receptive field tile, in terms of pools on the sending layer"`
	Skip  evec.Vec2i `desc:"how many pools to skip in tiling over sending layer -- typically 1/2 of Size"`
	Start evec.Vec2i `desc:"starting pool offset for lower-left corner of first receptive field in sending layer"`
	Wrap  bool       `desc:"if true, pool coordinates wrap around sending shape -- otherwise truncated at edges, which can lead to assymmetries in connectivity etc"`
}

func NewPoolTile() *PoolTile {
	pt := &PoolTile{}
	pt.Defaults()
	return pt
}

func (pt *PoolTile) Defaults() {
	pt.Size.Set(4, 4)
	pt.Skip.Set(2, 2)
	pt.Start.Set(-1, -1)
	pt.Wrap = true
}

func (pt *PoolTile) Name() string {
	return "PoolTile"
}

func (pt *PoolTile) Connect(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	sendn, recvn, cons = NewTensors(send, recv)
	sNtot := send.Len()
	sNy := send.Dim(0)
	sNx := send.Dim(1)
	rNy := recv.Dim(0)
	rNx := recv.Dim(1)
	sNu := send.Dim(2) * send.Dim(3)
	rNu := recv.Dim(2) * recv.Dim(3)
	rnv := recvn.Values
	snv := sendn.Values
	var clip bool
	for ry := 0; ry < rNy; ry++ {
		for rx := 0; rx < rNx; rx++ {
			rpi := ry*rNx + rx
			ris := rpi * rNu
			for fy := 0; fy < pt.Size.Y; fy++ {
				sy := pt.Start.Y + ry*pt.Skip.Y + fy
				if sy, clip = Edge(sy, sNy, pt.Wrap); clip {
					continue
				}
				for fx := 0; fx < pt.Size.X; fx++ {
					sx := pt.Start.X + rx*pt.Skip.X + fx
					if sx, clip = Edge(sx, sNx, pt.Wrap); clip {
						continue
					}
					spi := sy*sNx + sx
					sis := spi * sNu
					for rui := 0; rui < rNu; rui++ {
						ri := ris + rui
						for sui := 0; sui < sNu; sui++ {
							si := sis + sui
							off := ri*sNtot + si
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
