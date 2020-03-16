// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prjn

import (
	"log"

	"github.com/emer/emergent/evec"
	"github.com/emer/etable/etensor"
	"github.com/emer/etable/minmax"
	"github.com/goki/mat32"
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
	Recip       bool       `desc:"reciprocal topographic connectivity -- logic runs with recv <-> send -- produces symmetric back-projection or topo prjn when sending layer is larger than recv"`
	Size        evec.Vec2i `desc:"size of receptive field tile, in terms of pools on the sending layer"`
	Skip        evec.Vec2i `desc:"how many pools to skip in tiling over sending layer -- typically 1/2 of Size"`
	Start       evec.Vec2i `desc:"starting pool offset for lower-left corner of first receptive field in sending layer"`
	Wrap        bool       `desc:"if true, pool coordinates wrap around sending shape -- otherwise truncated at edges, which can lead to assymmetries in connectivity etc"`
	GaussFull   GaussTopo  `desc:"gaussian topographic weights / scaling parameters for full receptive field width. multiplies any other factors present"`
	GaussInPool GaussTopo  `desc:"gaussian topographic weights / scaling parameters within individual sending pools (i.e., unit positions within their parent pool drive distance for gaussian) -- this helps organize / differentiate units more within pools, not just across entire receptive field. multiplies any other factors present"`
	// SigmoidTopo SigmoidTopo `desc:"sigmoidal topographic weights / scaling parameters"`
	TopoRange minmax.F32 `desc:"min..max range of topographic weight values to generate"`
}

func NewPoolTile() *PoolTile {
	pt := &PoolTile{}
	pt.Defaults()
	return pt
}

// NewPoolTileRecip creates a new PoolTile that is a recip version of given ff feedforward one
func NewPoolTileRecip(ff *PoolTile) *PoolTile {
	pt := &PoolTile{}
	*pt = *ff
	pt.Recip = true
	return pt
}

func (pt *PoolTile) Defaults() {
	pt.Size.Set(4, 4)
	pt.Skip.Set(2, 2)
	pt.Start.Set(-1, -1)
	pt.Wrap = true
	pt.TopoRange.Min = 0.8
	pt.TopoRange.Max = 1
	pt.GaussFull.Defaults()
	pt.GaussInPool.Defaults()
}

func (pt *PoolTile) Name() string {
	return "PoolTile"
}

func (pt *PoolTile) Connect(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	if pt.Recip {
		return pt.ConnectRecip(send, recv, same)
	}
	sendn, recvn, cons = NewTensors(send, recv)
	if send.NumDims() != 4 || recv.NumDims() != 4 {
		log.Printf("prjn.PoolTile; only valid if both sending and receiving layer are 4D shape with outer 2D being pools and inner 2D are units within pools")
		return
	}
	sNtot := send.Len()
	sNpY := send.Dim(0)
	sNpX := send.Dim(1)
	rNpY := recv.Dim(0)
	rNpX := recv.Dim(1)
	sNu := send.Dim(2) * send.Dim(3)
	rNu := recv.Dim(2) * recv.Dim(3)
	rnv := recvn.Values
	snv := sendn.Values
	var clip bool
	for rpy := 0; rpy < rNpY; rpy++ {
		for rpx := 0; rpx < rNpX; rpx++ {
			rpi := rpy*rNpX + rpx
			ris := rpi * rNu
			for fy := 0; fy < pt.Size.Y; fy++ {
				spy := pt.Start.Y + rpy*pt.Skip.Y + fy
				if spy, clip = Edge(spy, sNpY, pt.Wrap); clip {
					continue
				}
				for fx := 0; fx < pt.Size.X; fx++ {
					spx := pt.Start.X + rpx*pt.Skip.X + fx
					if spx, clip = Edge(spx, sNpX, pt.Wrap); clip {
						continue
					}
					spi := spy*sNpX + spx
					sis := spi * sNu
					for rui := 0; rui < rNu; rui++ {
						ri := ris + rui
						for sui := 0; sui < sNu; sui++ {
							si := sis + sui
							off := ri*sNtot + si
							if off < cons.Len() {
								// if !pt.SelfCon && same && ri == si {
								// 	continue
								// }
								cons.Values.Set(off, true)
								if ri < len(rnv) {
									rnv[ri]++
								}
								if si < len(snv) {
									snv[si]++
								}
							}
						}
					}
				}
			}
		}
	}
	return
}

func (pt *PoolTile) ConnectRecip(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	sendn, recvn, cons = NewTensors(send, recv)
	if send.NumDims() != 4 || recv.NumDims() != 4 {
		log.Printf("prjn.PoolTile; only valid if both sending and receiving layer are 4D shape with outer 2D being pools and inner 2D are units within pools")
		return
	}
	sNtot := send.Len()
	sNpY := recv.Dim(0) // swapped
	sNpX := recv.Dim(1)
	rNpY := send.Dim(0)
	rNpX := send.Dim(1)
	sNu := send.Dim(2) * send.Dim(3)
	rNu := recv.Dim(2) * recv.Dim(3)
	rnv := recvn.Values
	snv := sendn.Values
	var clip bool
	for rpy := 0; rpy < rNpY; rpy++ {
		for rpx := 0; rpx < rNpX; rpx++ {
			rpi := rpy*rNpX + rpx
			ris := rpi * sNu
			for fy := 0; fy < pt.Size.Y; fy++ {
				spy := pt.Start.Y + rpy*pt.Skip.Y + fy
				if spy, clip = Edge(spy, sNpY, pt.Wrap); clip {
					continue
				}
				for fx := 0; fx < pt.Size.X; fx++ {
					spx := pt.Start.X + rpx*pt.Skip.X + fx
					if spx, clip = Edge(spx, sNpX, pt.Wrap); clip {
						continue
					}
					spi := spy*sNpX + spx
					sis := spi * rNu
					for rui := 0; rui < rNu; rui++ {
						ri := sis + rui
						for sui := 0; sui < sNu; sui++ {
							si := ris + sui
							// note: indexes reversed here
							off := ri*sNtot + si
							if off < cons.Len() {
								cons.Values.Set(off, true)
								if ri < len(rnv) {
									rnv[ri]++
								}
								if si < len(snv) {
									snv[si]++
								}
							}
						}
					}
				}
			}
		}
	}
	return
}

/////////////////////////////////////////////////////
// Topo Wts

type GaussTopo struct {
	On      bool    `desc:"use gaussian topographic weights / scaling values"`
	Sigma   float32 `viewif:"On" def:"0.6" desc:"gaussian sigma (width) in normalized units where entire distance across relevant dimension is 1.0 -- typical useful values range from .3 to 1.5, with .6 default"`
	Wrap    bool    `viewif:"On" desc:"wrap the gaussian around on other sides of the receptive field, with the closest distance being used -- this removes strict topography but ensures a more uniform distribution of weight values so edge units don't have weaker overall weights"`
	CtrMove float32 `viewif:"On" def:"0.8,1" desc:"proportion to move gaussian center relative to the position of the receiving unit within its pool: 1.0 = centers span the entire range of the receptive field.  Typically want to use 1.0 for Wrap = true, and 0.8 for false"`
}

func (gt *GaussTopo) Defaults() {
	gt.On = true
	gt.Sigma = 0.6
	gt.Wrap = true
	gt.CtrMove = 1
}

// DefWrap sets default wrap parameters (which are overall defaults): CtrMove = 1
func (gt *GaussTopo) DefWrap() {
	gt.Wrap = true
	gt.CtrMove = 1
}

// DefNoWrap sets default no-wrap parameters (CtrMove = .8 instead of 1)
func (gt *GaussTopo) DefNoWrap() {
	gt.Wrap = false
	gt.CtrMove = 0.8
}

// TopoWts sets values in given 6D tensor according to Topo settings.
// wts is shaped with first 2 outer-most dims as Y, X of units within pool
// of recv layer (these are units over which topography is defined)
// and remaing 4D is for receptive field Size by units within pool size for
// sending layer.
func (pt *PoolTile) TopoWts(send, recv *etensor.Shape, wts *etensor.Float32) {
	if send.NumDims() != 4 || recv.NumDims() != 4 {
		log.Printf("prjn.PoolTile; only valid if both sending and receiving layer are 4D shape with outer 2D being pools and inner 2D are units within pools")
		return
	}
	if pt.GaussFull.Sigma == 0 {
		pt.GaussFull.Defaults()
	}
	if pt.GaussInPool.Sigma == 0 {
		pt.GaussInPool.Defaults()
	}
	sNuY := send.Dim(2)
	sNuX := send.Dim(3)
	rNuY := recv.Dim(2)
	rNuX := recv.Dim(3)
	wshp := []int{rNuY, rNuX, pt.Size.Y, pt.Size.X, sNuY, sNuX}
	wts.SetShape(wshp, nil, []string{"rNuY", "rNuX", "szY", "szX", "sNuY", "sNuX"})

	fsz := mat32.Vec2{float32(pt.Size.X*sNuX - 1), float32(pt.Size.Y*sNuY - 1)} // full rf size
	hfsz := fsz.MulScalar(0.5)                                                  // half rf
	fsig := pt.GaussFull.Sigma * hfsz.X                                         // full sigma
	if fsig <= 0 {
		fsig = pt.GaussFull.Sigma
	}

	psz := mat32.Vec2{float32(sNuX), float32(sNuY)} // within-pool rf size
	if sNuX > 1 {
		psz.X -= 1
	}
	if sNuY > 1 {
		psz.Y -= 1
	}
	hpsz := psz.MulScalar(0.5)            // half rf
	psig := pt.GaussInPool.Sigma * hpsz.X // pool sigma
	if psig <= 0 {
		psig = pt.GaussInPool.Sigma
	}

	rsz := mat32.Vec2{float32(rNuX), float32(rNuY)} // recv units-in-pool size
	if rNuX > 1 {
		rsz.X -= 1
	}
	if rNuY > 1 {
		rsz.Y -= 1
	}
	hrsz := rsz.MulScalar(0.5)
	for ruy := 0; ruy < rNuY; ruy++ {
		for rux := 0; rux < rNuX; rux++ {
			rpos := mat32.Vec2{float32(rux), float32(ruy)}.Sub(hrsz).Div(hrsz) // -1..1 normalized r unit pos
			rfpos := rpos.MulScalar(pt.GaussFull.CtrMove)
			rppos := rpos.MulScalar(pt.GaussInPool.CtrMove)
			sfctr := rfpos.Mul(hfsz).Add(hfsz) // sending center for full
			spctr := rppos.Mul(hpsz).Add(hpsz) // sending center for within-pool
			for fy := 0; fy < pt.Size.Y; fy++ {
				for fx := 0; fx < pt.Size.X; fx++ {
					for suy := 0; suy < sNuY; suy++ {
						for sux := 0; sux < sNuX; sux++ {
							fwt := float32(1)
							if pt.GaussFull.On {
								sf := mat32.Vec2{float32(fx*sNuX + sux), float32(fy*sNuY + suy)}
								if pt.GaussFull.Wrap {
									sf.X = WrapMinDist(sf.X, fsz.X, sfctr.X)
									sf.Y = WrapMinDist(sf.Y, fsz.Y, sfctr.Y)
								}
								fwt = evec.GaussVecDistNoNorm(sf, sfctr, fsig)
							}
							pwt := float32(1)
							if pt.GaussInPool.On {
								sp := mat32.Vec2{float32(sux), float32(suy)}
								if pt.GaussInPool.Wrap {
									sp.X = WrapMinDist(sp.X, psz.X, spctr.X)
									sp.Y = WrapMinDist(sp.Y, psz.Y, spctr.Y)
								}
								pwt = evec.GaussVecDistNoNorm(sp, spctr, psig)
							}
							wt := fwt * pwt
							rwt := pt.TopoRange.ProjVal(wt)
							wts.Set([]int{ruy, rux, fy, fx, suy, sux}, rwt)
						}
					}
				}
			}
		}
	}
}
