// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prjn

import (
	"fmt"
	"log"

	"github.com/emer/emergent/edge"
	"github.com/emer/emergent/efuns"
	"github.com/emer/emergent/evec"
	"github.com/emer/etable/etensor"
	"github.com/emer/etable/minmax"
	"github.com/goki/mat32"
)

// PoolTileSub implements tiled 2D connectivity between pools within layers, where
// a 2D rectangular receptive field (defined over pools, not units) is tiled
// across the sending layer pools, with specified level of overlap.
// Pools are the outer-most two dimensions of a 4D layer shape.
// Sub version has sub-pools within each pool to encourage more independent
// representations.
// 2D layers are assumed to have 1x1 pool.
// This is a standard form of convolutional connectivity, where pools are
// the filters and the outer dims are locations filtered.
// Various initial weight / scaling patterns are also available -- code
// must specifically apply these to the receptive fields.
type PoolTileSub struct {
	Recip       bool        `desc:"reciprocal topographic connectivity -- logic runs with recv <-> send -- produces symmetric back-projection or topo prjn when sending layer is larger than recv"`
	Size        evec.Vec2i  `desc:"size of receptive field tile, in terms of pools on the sending layer"`
	Skip        evec.Vec2i  `desc:"how many pools to skip in tiling over sending layer -- typically 1/2 of Size"`
	Start       evec.Vec2i  `desc:"starting pool offset for lower-left corner of first receptive field in sending layer"`
	Subs        evec.Vec2i  `desc:"number of sub-pools within each pool"`
	SendSubs    bool        `desc:"sending layer has sub-pools"`
	Wrap        bool        `desc:"if true, pool coordinates wrap around sending shape -- otherwise truncated at edges, which can lead to assymmetries in connectivity etc"`
	GaussFull   GaussTopo   `desc:"gaussian topographic weights / scaling parameters for full receptive field width. multiplies any other factors present"`
	GaussInPool GaussTopo   `desc:"gaussian topographic weights / scaling parameters within individual sending pools (i.e., unit positions within their parent pool drive distance for gaussian) -- this helps organize / differentiate units more within pools, not just across entire receptive field. multiplies any other factors present"`
	SigFull     SigmoidTopo `desc:"sigmoidal topographic weights / scaling parameters for full receptive field width.  left / bottom half have increasing sigmoids, and second half decrease.  Multiplies any other factors present (only used if Gauss versions are not On!)"`
	SigInPool   SigmoidTopo `desc:"sigmoidal topographic weights / scaling parameters within individual sending pools (i.e., unit positions within their parent pool drive distance for sigmoid) -- this helps organize / differentiate units more within pools, not just across entire receptive field. multiplies any other factors present  (only used if Gauss versions are not On!).  left / bottom half have increasing sigmoids, and second half decrease."`
	TopoRange   minmax.F32  `desc:"min..max range of topographic weight values to generate "`
}

func NewPoolTileSub() *PoolTileSub {
	pt := &PoolTileSub{}
	pt.Defaults()
	return pt
}

// NewPoolTileSubRecip creates a new PoolTileSub that is a recip version of given ff feedforward one
func NewPoolTileSubRecip(ff *PoolTileSub) *PoolTileSub {
	pt := &PoolTileSub{}
	*pt = *ff
	pt.Recip = true
	return pt
}

func (pt *PoolTileSub) Defaults() {
	pt.Size.Set(4, 4)
	pt.Skip.Set(2, 2)
	pt.Start.Set(-1, -1)
	pt.Subs.Set(2, 2)
	pt.Wrap = true
	pt.TopoRange.Min = 0.8
	pt.TopoRange.Max = 1
	pt.GaussFull.Defaults()
	pt.GaussInPool.Defaults()
	pt.SigFull.Defaults()
	pt.SigInPool.Defaults()
	pt.GaussFull.On = true
	pt.GaussInPool.On = true
}

func (pt *PoolTileSub) Name() string {
	return "PoolTileSub"
}

func (pt *PoolTileSub) Connect(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	if pt.Recip {
		return pt.ConnectRecip(send, recv, same)
	}
	sendn, recvn, cons = NewTensors(send, recv)
	sNtot := send.Len()
	sNpY := send.Dim(0)
	sNpX := send.Dim(1)
	rNpY := recv.Dim(0)
	rNpX := recv.Dim(1)
	sNu := 1
	rNu := 1
	if send.NumDims() == 4 {
		sNu = send.Dim(2) * send.Dim(3)
	} else {
		sNpY = 1
		sNpX = 1
		sNu = send.Dim(0) * send.Dim(1)
	}
	if recv.NumDims() == 4 {
		rNu = recv.Dim(2) * recv.Dim(3)
	} else {
		rNpY = 1
		rNpX = 1
		rNu = recv.Dim(0) * recv.Dim(1)
	}
	rnv := recvn.Values
	snv := sendn.Values
	var clip bool
	for rpy := 0; rpy < rNpY; rpy++ {
		rpys := rpy / pt.Subs.Y // sub group
		rpyi := rpy % pt.Subs.Y // index within subgroup
		for rpx := 0; rpx < rNpX; rpx++ {
			rpxs := rpx / pt.Subs.X
			rpxi := rpx % pt.Subs.X
			rpi := rpy*rNpX + rpx
			ris := rpi * rNu
			for fy := 0; fy < pt.Size.Y; fy++ {
				spy := pt.Start.Y + rpys*pt.Skip.Y + fy
				if pt.SendSubs {
					spy = spy*pt.Subs.Y + rpyi
				}
				if spy, clip = edge.Edge(spy, sNpY, pt.Wrap); clip {
					continue
				}
				for fx := 0; fx < pt.Size.X; fx++ {
					spx := pt.Start.X + rpxs*pt.Skip.X + fx
					if pt.SendSubs {
						spx = spx*pt.Subs.X + rpxi
					}
					if spx, clip = edge.Edge(spx, sNpX, pt.Wrap); clip {
						continue
					}
					spi := spy*sNpX + spx
					sis := spi * sNu
					for rui := 0; rui < rNu; rui++ {
						ri := ris + rui
						for sui := 0; sui < sNu; sui++ {
							si := sis + sui
							off := ri*sNtot + si
							if off < cons.Len() && ri < len(rnv) && si < len(snv) {
								// if !pt.SelfCon && same && ri == si {
								// 	continue
								// }
								cons.Values.Set(off, true)
								rnv[ri]++
								snv[si]++
							}
						}
					}
				}
			}
		}
	}
	return
}

func (pt *PoolTileSub) ConnectRecip(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	sendn, recvn, cons = NewTensors(send, recv)
	// all these variables are swapped: s from recv, r from send
	rNtot := send.Len()
	sNpY := recv.Dim(0)
	sNpX := recv.Dim(1)
	rNpY := send.Dim(0)
	rNpX := send.Dim(1)
	sNu := 1
	rNu := 1
	if recv.NumDims() == 4 {
		sNu = recv.Dim(2) * recv.Dim(3)
	} else {
		sNpY = 1
		sNpX = 1
		sNu = recv.Dim(0) * recv.Dim(1)
	}
	if send.NumDims() == 4 {
		rNu = send.Dim(2) * send.Dim(3)
	} else {
		rNpY = 1
		rNpX = 1
		rNu = send.Dim(0) * send.Dim(1)
	}
	snv := recvn.Values
	rnv := sendn.Values
	var clip bool
	for rpy := 0; rpy < rNpY; rpy++ {
		rpys := rpy / pt.Subs.Y // sub group
		rpyi := rpy % pt.Subs.Y // index within subgroup
		for rpx := 0; rpx < rNpX; rpx++ {
			rpxs := rpx / pt.Subs.X
			rpxi := rpx % pt.Subs.X
			rpi := rpy*rNpX + rpx
			ris := rpi * rNu
			for fy := 0; fy < pt.Size.Y; fy++ {
				spy := pt.Start.Y + rpys*pt.Skip.Y + fy
				if pt.SendSubs {
					spy = spy*pt.Subs.Y + rpyi
				}
				if spy, clip = edge.Edge(spy, sNpY, pt.Wrap); clip {
					continue
				}
				for fx := 0; fx < pt.Size.X; fx++ {
					spx := pt.Start.X + rpxs*pt.Skip.X + fx
					if pt.SendSubs {
						spx = spx*pt.Subs.X + rpxi
					}
					if spx, clip = edge.Edge(spx, sNpX, pt.Wrap); clip {
						continue
					}
					spi := spy*sNpX + spx
					sis := spi * sNu
					for sui := 0; sui < sNu; sui++ {
						si := sis + sui
						for rui := 0; rui < rNu; rui++ {
							ri := ris + rui
							off := si*rNtot + ri
							if off < cons.Len() && si < len(snv) && ri < len(rnv) {
								cons.Values.Set(off, true)
								snv[si]++
								rnv[ri]++
							}
						}
					}
				}
			}
		}
	}
	return
}

// HasTopoWts returns true if some form of topographic weight patterns are set
func (pt *PoolTileSub) HasTopoWts() bool {
	return pt.GaussFull.On || pt.GaussInPool.On || pt.SigFull.On || pt.SigInPool.On
}

// TopoWts sets values in given 4D or 6D tensor according to *Topo settings.
// wts is shaped with first 2 outer-most dims as Y, X of units within layer / pool
// of recv layer (these are units over which topography is defined)
// and remaing 2D or 4D is for receptive field Size by units within pool size for
// sending layer.
func (pt *PoolTileSub) TopoWts(send, recv *etensor.Shape, wts *etensor.Float32) error {
	if pt.GaussFull.On || pt.GaussInPool.On {
		if send.NumDims() == 2 {
			return pt.TopoWtsGauss2D(send, recv, wts)
		} else {
			return pt.TopoWtsGauss4D(send, recv, wts)
		}
	}
	if pt.SigFull.On || pt.SigInPool.On {
		if send.NumDims() == 2 {
			return pt.TopoWtsSigmoid2D(send, recv, wts)
		} else {
			return pt.TopoWtsSigmoid4D(send, recv, wts)
		}
	}
	err := fmt.Errorf("PoolTileSub:TopoWts no Gauss or Sig params turned on")
	log.Println(err)
	return err
}

// GaussOff turns off gaussian weights
func (pt *PoolTileSub) GaussOff() {
	pt.GaussFull.On = false
	pt.GaussInPool.On = false
}

// TopoWtsGauss2D sets values in given 4D tensor according to *Topo settings.
// wts is shaped with first 2 outer-most dims as Y, X of units within layer / pool
// of recv layer (these are units over which topography is defined)
// and remaing 2D is for sending layer size (2D = sender)
func (pt *PoolTileSub) TopoWtsGauss2D(send, recv *etensor.Shape, wts *etensor.Float32) error {
	if pt.GaussFull.Sigma == 0 {
		pt.GaussFull.Defaults()
	}
	if pt.GaussInPool.Sigma == 0 {
		pt.GaussInPool.Defaults()
	}
	sNuY := send.Dim(0)
	sNuX := send.Dim(1)
	rNuY := recv.Dim(0) // ok if recv is 2D
	rNuX := recv.Dim(1)
	if recv.NumDims() == 4 {
		rNuY = recv.Dim(2)
		rNuX = recv.Dim(3)
	}
	wshp := []int{rNuY, rNuX, sNuY, sNuX}
	wts.SetShape(wshp, nil, []string{"rNuY", "rNuX", "szY", "szX"})

	fsz := mat32.Vec2{float32(sNuX - 1), float32(sNuY - 1)} // full rf size
	hfsz := fsz.MulScalar(0.5)                              // half rf
	fsig := pt.GaussFull.Sigma * hfsz.X                     // full sigma
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
			for suy := 0; suy < sNuY; suy++ {
				for sux := 0; sux < sNuX; sux++ {
					fwt := float32(1)
					if pt.GaussFull.On {
						sf := mat32.Vec2{float32(sux), float32(suy)}
						if pt.GaussFull.Wrap {
							sf.X = edge.WrapMinDist(sf.X, fsz.X, sfctr.X)
							sf.Y = edge.WrapMinDist(sf.Y, fsz.Y, sfctr.Y)
						}
						fwt = efuns.GaussVecDistNoNorm(sf, sfctr, fsig)
					}
					pwt := float32(1)
					if pt.GaussInPool.On {
						sp := mat32.Vec2{float32(sux), float32(suy)}
						if pt.GaussInPool.Wrap {
							sp.X = edge.WrapMinDist(sp.X, psz.X, spctr.X)
							sp.Y = edge.WrapMinDist(sp.Y, psz.Y, spctr.Y)
						}
						pwt = efuns.GaussVecDistNoNorm(sp, spctr, psig)
					}
					wt := fwt * pwt
					rwt := pt.TopoRange.ProjVal(wt)
					wts.Set([]int{ruy, rux, suy, sux}, rwt)
				}
			}
		}
	}
	return nil
}

// TopoWtsGauss4D sets values in given 6D tensor according to *Topo settings.
// wts is shaped with first 2 outer-most dims as Y, X of units within layer / pool
// of recv layer (these are units over which topography is defined)
// and remaing 4D is for receptive field Size by units within pool size for
// sending layer.
func (pt *PoolTileSub) TopoWtsGauss4D(send, recv *etensor.Shape, wts *etensor.Float32) error {
	if pt.GaussFull.Sigma == 0 {
		pt.GaussFull.Defaults()
	}
	if pt.GaussInPool.Sigma == 0 {
		pt.GaussInPool.Defaults()
	}
	sNuY := send.Dim(2)
	sNuX := send.Dim(3)
	rNuY := recv.Dim(0) // ok if recv is 2D
	rNuX := recv.Dim(1)
	if recv.NumDims() == 4 {
		rNuY = recv.Dim(2)
		rNuX = recv.Dim(3)
	}
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
									sf.X = edge.WrapMinDist(sf.X, fsz.X, sfctr.X)
									sf.Y = edge.WrapMinDist(sf.Y, fsz.Y, sfctr.Y)
								}
								fwt = efuns.GaussVecDistNoNorm(sf, sfctr, fsig)
							}
							pwt := float32(1)
							if pt.GaussInPool.On {
								sp := mat32.Vec2{float32(sux), float32(suy)}
								if pt.GaussInPool.Wrap {
									sp.X = edge.WrapMinDist(sp.X, psz.X, spctr.X)
									sp.Y = edge.WrapMinDist(sp.Y, psz.Y, spctr.Y)
								}
								pwt = efuns.GaussVecDistNoNorm(sp, spctr, psig)
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
	return nil
}

/////////////////////////////////////////////////////
// SigmoidTopo Wts

// TopoWtsSigmoid2D sets values in given 4D tensor according to Topo settings.
// wts is shaped with first 2 outer-most dims as Y, X of units within pool
// of recv layer (these are units over which topography is defined)
// and remaing 2D is for sending layer (2D = sender).
func (pt *PoolTileSub) TopoWtsSigmoid2D(send, recv *etensor.Shape, wts *etensor.Float32) error {
	if pt.SigFull.Gain == 0 {
		pt.SigFull.Defaults()
	}
	if pt.SigInPool.Gain == 0 {
		pt.SigInPool.Defaults()
	}
	sNuY := send.Dim(0)
	sNuX := send.Dim(1)
	rNuY := recv.Dim(0) // ok if recv is 2D
	rNuX := recv.Dim(1)
	if recv.NumDims() == 4 {
		rNuY = recv.Dim(2)
		rNuX = recv.Dim(3)
	}
	wshp := []int{rNuY, rNuX, sNuY, sNuX}
	wts.SetShape(wshp, nil, []string{"rNuY", "rNuX", "sNuY", "sNuX"})

	fsz := mat32.Vec2{float32(sNuX - 1), float32(sNuY - 1)} // full rf size
	hfsz := fsz.MulScalar(0.5)                              // half rf
	fgain := pt.SigFull.Gain * hfsz.X                       // full gain

	psz := mat32.Vec2{float32(sNuX), float32(sNuY)} // within-pool rf size
	if sNuX > 1 {
		psz.X -= 1
	}
	if sNuY > 1 {
		psz.Y -= 1
	}
	hpsz := psz.MulScalar(0.5)          // half rf
	pgain := pt.SigInPool.Gain * hpsz.X // pool sigma

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
			rpos := mat32.Vec2{float32(rux), float32(ruy)}.Div(hrsz) // 0..2 normalized r unit pos
			sgn := mat32.Vec2{1, 1}
			rfpos := rpos.SubScalar(0.5).MulScalar(pt.SigFull.CtrMove).AddScalar(0.5)
			rppos := rpos.SubScalar(0.5).MulScalar(pt.SigInPool.CtrMove).AddScalar(0.5)
			if rpos.X >= 1 { // flip direction half-way through
				sgn.X = -1
				rpos.X = -rpos.X + 1
				rfpos.X = (rpos.X+0.5)*pt.SigFull.CtrMove - 0.5
				rppos.X = (rpos.X+0.5)*pt.SigInPool.CtrMove - 0.5
			}
			if rpos.Y >= 1 {
				sgn.Y = -1
				rpos.Y = -rpos.Y + 1
				rfpos.Y = (rpos.Y+0.5)*pt.SigFull.CtrMove - 0.5
				rfpos.Y = (rpos.Y+0.5)*pt.SigInPool.CtrMove - 0.5
			}
			sfctr := rfpos.Mul(fsz) // sending center for full
			spctr := rppos.Mul(psz) // sending center for within-pool
			for suy := 0; suy < sNuY; suy++ {
				for sux := 0; sux < sNuX; sux++ {
					fwt := float32(1)
					if pt.SigFull.On {
						sf := mat32.Vec2{float32(sux), float32(suy)}
						sigx := efuns.Logistic(sgn.X*sf.X, fgain, sfctr.X)
						sigy := efuns.Logistic(sgn.Y*sf.Y, fgain, sfctr.Y)
						fwt = sigx * sigy
					}
					pwt := float32(1)
					if pt.SigInPool.On {
						sp := mat32.Vec2{float32(sux), float32(suy)}
						sigx := efuns.Logistic(sgn.X*sp.X, pgain, spctr.X)
						sigy := efuns.Logistic(sgn.Y*sp.Y, pgain, spctr.Y)
						pwt = sigx * sigy
					}
					wt := fwt * pwt
					rwt := pt.TopoRange.ProjVal(wt)
					wts.Set([]int{ruy, rux, suy, sux}, rwt)
				}
			}
		}
	}
	return nil
}

// TopoWtsSigmoid4D sets values in given 6D tensor according to Topo settings.
// wts is shaped with first 2 outer-most dims as Y, X of units within pool
// of recv layer (these are units over which topography is defined)
// and remaing 2D is for receptive field Size by units within pool size for
// sending layer.
func (pt *PoolTileSub) TopoWtsSigmoid4D(send, recv *etensor.Shape, wts *etensor.Float32) error {
	if pt.SigFull.Gain == 0 {
		pt.SigFull.Defaults()
	}
	if pt.SigInPool.Gain == 0 {
		pt.SigInPool.Defaults()
	}
	sNuY := send.Dim(2)
	sNuX := send.Dim(3)
	rNuY := recv.Dim(0) // ok if recv is 2D
	rNuX := recv.Dim(1)
	if recv.NumDims() == 4 {
		rNuY = recv.Dim(2)
		rNuX = recv.Dim(3)
	}
	wshp := []int{rNuY, rNuX, pt.Size.Y, pt.Size.X, sNuY, sNuX}
	wts.SetShape(wshp, nil, []string{"rNuY", "rNuX", "szY", "szX", "sNuY", "sNuX"})

	fsz := mat32.Vec2{float32(pt.Size.X*sNuX - 1), float32(pt.Size.Y*sNuY - 1)} // full rf size
	hfsz := fsz.MulScalar(0.5)                                                  // half rf
	fgain := pt.SigFull.Gain * hfsz.X                                           // full gain

	psz := mat32.Vec2{float32(sNuX), float32(sNuY)} // within-pool rf size
	if sNuX > 1 {
		psz.X -= 1
	}
	if sNuY > 1 {
		psz.Y -= 1
	}
	hpsz := psz.MulScalar(0.5)          // half rf
	pgain := pt.SigInPool.Gain * hpsz.X // pool sigma

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
			rpos := mat32.Vec2{float32(rux), float32(ruy)}.Div(hrsz) // 0..2 normalized r unit pos
			sgn := mat32.Vec2{1, 1}
			rfpos := rpos.SubScalar(0.5).MulScalar(pt.SigFull.CtrMove).AddScalar(0.5)
			rppos := rpos.SubScalar(0.5).MulScalar(pt.SigInPool.CtrMove).AddScalar(0.5)
			if rpos.X >= 1 { // flip direction half-way through
				sgn.X = -1
				rpos.X = -rpos.X + 1
				rfpos.X = (rpos.X+0.5)*pt.SigFull.CtrMove - 0.5
				rppos.X = (rpos.X+0.5)*pt.SigInPool.CtrMove - 0.5
			}
			if rpos.Y >= 1 {
				sgn.Y = -1
				rpos.Y = -rpos.Y + 1
				rfpos.Y = (rpos.Y+0.5)*pt.SigFull.CtrMove - 0.5
				rfpos.Y = (rpos.Y+0.5)*pt.SigInPool.CtrMove - 0.5
			}
			sfctr := rfpos.Mul(fsz) // sending center for full
			spctr := rppos.Mul(psz) // sending center for within-pool
			for fy := 0; fy < pt.Size.Y; fy++ {
				for fx := 0; fx < pt.Size.X; fx++ {
					for suy := 0; suy < sNuY; suy++ {
						for sux := 0; sux < sNuX; sux++ {
							fwt := float32(1)
							if pt.SigFull.On {
								sf := mat32.Vec2{float32(fx*sNuX + sux), float32(fy*sNuY + suy)}
								sigx := efuns.Logistic(sgn.X*sf.X, fgain, sfctr.X)
								sigy := efuns.Logistic(sgn.Y*sf.Y, fgain, sfctr.Y)
								fwt = sigx * sigy
							}
							pwt := float32(1)
							if pt.SigInPool.On {
								sp := mat32.Vec2{float32(sux), float32(suy)}
								sigx := efuns.Logistic(sgn.X*sp.X, pgain, spctr.X)
								sigy := efuns.Logistic(sgn.Y*sp.Y, pgain, spctr.Y)
								pwt = sigx * sigy
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
	return nil
}
