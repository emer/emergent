// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import (
	"github.com/emer/emergent/v2/emer"
	"goki.dev/etable/v2/etensor"
	"goki.dev/mat32/v2"
	"goki.dev/vgpu/v2/vshape"
	"goki.dev/xyz"
)

// LayMesh is a xyz.Mesh that represents a layer -- it is dynamically updated using the
// Update method which only resets the essential Vertex elements.
// The geometry is literal in the layer size: 0,0,0 lower-left corner and increasing X,Z
// for the width and height of the layer, in unit (1) increments per unit..
// NetView applies an overall scaling to make it fit within the larger view.
type LayMesh struct {
	xyz.MeshBase

	// layer that we render
	Lay emer.Layer

	// current shape that has been constructed -- if same, just update
	Shape etensor.Shape

	// netview that we're in
	View *NetView
}

// NewLayMesh adds LayMesh mesh to given scene for given layer
func NewLayMesh(sc *xyz.Scene, nv *NetView, lay emer.Layer) *LayMesh {
	lm := &LayMesh{}
	lm.View = nv
	lm.Lay = lay
	lm.Nm = lay.Name()
	sc.AddMesh(lm)
	return lm
}

func (lm *LayMesh) Sizes() (nVtx, nIdx int, hasColor bool) {
	lm.Trans = true
	lm.Dynamic = true
	lm.Color = true
	if lm.Lay == nil {
		return 0, 0, true
	}
	shp := lm.Lay.Shape()
	lm.Shape.CopyShape(shp)
	if lm.View.Params.Raster.On {
		if shp.NumDims() == 4 {
			lm.NVtx, lm.NIdx = lm.RasterSize4D()
		} else {
			lm.NVtx, lm.NIdx = lm.RasterSize2D()
		}
	} else {
		if shp.NumDims() == 4 {
			lm.NVtx, lm.NIdx = lm.Size4D()
		} else {
			lm.NVtx, lm.NIdx = lm.Size2D()
		}
	}
	return lm.NVtx, lm.NIdx, lm.Color
}

func (lm *LayMesh) Size2D() (nVtx, nIdx int) {
	nz := lm.Shape.Dim(0)
	nx := lm.Shape.Dim(1)
	segs := 1

	vtxSz, idxSz := vshape.PlaneN(segs, segs)
	nVtx = vtxSz * 5 * nz * nx
	nIdx = idxSz * 5 * nz * nx
	return
}

func (lm *LayMesh) Size4D() (nVtx, nIdx int) {
	npz := lm.Shape.Dim(0) // p = pool
	npx := lm.Shape.Dim(1)
	nuz := lm.Shape.Dim(2) // u = unit
	nux := lm.Shape.Dim(3)

	segs := 1

	vtxSz, idxSz := vshape.PlaneN(segs, segs)
	nVtx = vtxSz * 5 * npz * npx * nuz * nux
	nIdx = idxSz * 5 * npz * npx * nuz * nux
	return
}

func (lm *LayMesh) Set(sc *xyz.Scene, vtxAry, normAry, texAry, clrAry mat32.ArrayF32, idxAry mat32.ArrayU32) {
	if lm.Lay == nil || lm.Shape.NumDims() == 0 {
		return // nothing
	}
	// true = init
	if lm.View.Params.Raster.On {
		if lm.View.Params.Raster.XAxis {
			if lm.Shape.NumDims() == 4 {
				lm.RasterSet4DX(sc, true, vtxAry, normAry, texAry, clrAry, idxAry)
			} else {
				lm.RasterSet2DX(sc, true, vtxAry, normAry, texAry, clrAry, idxAry)
			}
		} else {
			if lm.Shape.NumDims() == 4 {
				lm.RasterSet4DZ(sc, true, vtxAry, normAry, texAry, clrAry, idxAry)
			} else {
				lm.RasterSet2DZ(sc, true, vtxAry, normAry, texAry, clrAry, idxAry)
			}
		}
	} else {
		if lm.Shape.NumDims() == 4 {
			lm.Set4D(sc, true, vtxAry, normAry, texAry, clrAry, idxAry)
		} else {
			lm.Set2D(sc, true, vtxAry, normAry, texAry, clrAry, idxAry)
		}
	}
	lm.SetMod(sc)

}

func (lm *LayMesh) Update(sc *xyz.Scene, vtxAry, normAry, texAry, clrAry mat32.ArrayF32, idxAry mat32.ArrayU32) {
	if lm.Lay == nil || lm.Shape.NumDims() == 0 {
		return // nothing
	}
	// false = not init
	// todo: not using the init flag to save minor amount of index non-updating..

	if lm.View.Params.Raster.On {
		if lm.View.Params.Raster.XAxis {
			if lm.Shape.NumDims() == 4 {
				lm.RasterSet4DX(sc, false, vtxAry, normAry, texAry, clrAry, idxAry)
			} else {
				lm.RasterSet2DX(sc, false, vtxAry, normAry, texAry, clrAry, idxAry)
			}
		} else {
			if lm.Shape.NumDims() == 4 {
				lm.RasterSet4DZ(sc, false, vtxAry, normAry, texAry, clrAry, idxAry)
			} else {
				lm.RasterSet2DZ(sc, false, vtxAry, normAry, texAry, clrAry, idxAry)
			}
		}
	} else {
		if lm.Shape.NumDims() == 4 {
			lm.Set4D(sc, false, vtxAry, normAry, texAry, clrAry, idxAry)
		} else {
			lm.Set2D(sc, false, vtxAry, normAry, texAry, clrAry, idxAry)
		}
	}
	lm.SetMod(sc)
}

// MinUnitHeight ensures that there is always at least some dimensionality
// to the unit cubes -- affects transparency rendering etc
var MinUnitHeight = float32(1.0e-6)

func (lm *LayMesh) Set2D(sc *xyz.Scene, init bool, vtxAry, normAry, texAry, clrAry mat32.ArrayF32, idxAry mat32.ArrayU32) {
	nz := lm.Shape.Dim(0)
	nx := lm.Shape.Dim(1)

	fnz := float32(nz)
	fnx := float32(nx)

	uw := lm.View.Params.UnitSize
	uo := (1.0 - uw)
	segs := 1

	vtxSz, idxSz := vshape.PlaneN(segs, segs)
	pidx := 0 // plane index
	pos := mat32.Vec3{}

	lm.View.ReadLock()
	for zi := nz - 1; zi >= 0; zi-- {
		z0 := uo - float32(zi+1)
		for xi := 0; xi < nx; xi++ {
			poff := pidx * vtxSz * 5
			ioff := pidx * idxSz * 5
			x0 := uo + float32(xi)
			_, scaled, clr, _ := lm.View.UnitVal(lm.Lay, []int{zi, xi})
			v4c := mat32.NewVec4Color(clr)
			vshape.SetColor(clrAry, poff, 5*vtxSz, v4c)
			ht := 0.5 * mat32.Abs(scaled)
			if ht < MinUnitHeight {
				ht = MinUnitHeight
			}
			if scaled >= 0 {
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff, ioff, mat32.X, mat32.Y, -1, -1, uw, ht, x0, 0, z0, segs, segs, pos)                    // nz
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+1*vtxSz, ioff+1*idxSz, mat32.Z, mat32.Y, -1, -1, uw, ht, z0, 0, x0+uw, segs, segs, pos) // px
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+2*vtxSz, ioff+2*idxSz, mat32.Z, mat32.Y, 1, -1, uw, ht, z0, 0, x0, segs, segs, pos)     // nx
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+3*vtxSz, ioff+3*idxSz, mat32.X, mat32.Z, 1, 1, uw, uw, x0, z0, ht, segs, segs, pos)     // py <-
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+4*vtxSz, ioff+4*idxSz, mat32.X, mat32.Y, 1, -1, uw, ht, x0, 0, z0+uw, segs, segs, pos)  // pz
			} else {
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff, ioff, mat32.X, mat32.Y, 1, -1, uw, ht, x0, -ht, z0, segs, segs, pos)                    // nz = pz norm
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+1*vtxSz, ioff+1*idxSz, mat32.Z, mat32.Y, 1, -1, uw, ht, z0, -ht, x0+uw, segs, segs, pos) // px = nx norm
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+2*vtxSz, ioff+2*idxSz, mat32.Z, mat32.Y, 1, -1, uw, ht, z0, -ht, x0, segs, segs, pos)    // nx
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+3*vtxSz, ioff+3*idxSz, mat32.X, mat32.Z, 1, 1, uw, uw, x0, z0, -ht, segs, segs, pos)     // ny <-
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+4*vtxSz, ioff+4*idxSz, mat32.X, mat32.Y, 1, -1, uw, ht, x0, -ht, z0+uw, segs, segs, pos) // pz
			}
			pidx++
		}
	}
	lm.View.ReadUnlock()

	lm.BBoxMu.Lock()
	lm.BBox.SetBounds(mat32.Vec3{0, -0.5, -fnz}, mat32.Vec3{fnx, 0.5, 0})
	lm.BBoxMu.Unlock()
}

func (lm *LayMesh) Set4D(sc *xyz.Scene, init bool, vtxAry, normAry, texAry, clrAry mat32.ArrayF32, idxAry mat32.ArrayU32) {
	npz := lm.Shape.Dim(0) // p = pool
	npx := lm.Shape.Dim(1)
	nuz := lm.Shape.Dim(2) // u = unit
	nux := lm.Shape.Dim(3)

	fnpz := float32(npz)
	fnpx := float32(npx)
	fnuz := float32(nuz)
	fnux := float32(nux)

	usz := lm.View.Params.UnitSize
	uo := (1.0 - usz) // offset = space

	// for 4D, we build in spaces between groups without changing the overall size of layer
	// by shrinking the spacing of each unit according to the spaces we introduce
	xsc := (fnpx * fnux) / ((fnpx-1)*uo + (fnpx * fnux))
	zsc := (fnpz * fnuz) / ((fnpz-1)*uo + (fnpz * fnuz))

	xuw := xsc * usz
	zuw := zsc * usz

	segs := 1

	vtxSz, idxSz := vshape.PlaneN(segs, segs)
	pidx := 0 // plane index
	pos := mat32.Vec3{}

	lm.View.ReadLock()
	for zpi := npz - 1; zpi >= 0; zpi-- {
		zp0 := zsc * (-float32(zpi) * (uo + fnuz))
		for xpi := 0; xpi < npx; xpi++ {
			xp0 := xsc * (float32(xpi)*uo + float32(xpi)*fnux)
			for zui := nuz - 1; zui >= 0; zui-- {
				z0 := zp0 + zsc*(uo-float32(zui+1))
				for xui := 0; xui < nux; xui++ {
					poff := pidx * vtxSz * 5
					ioff := pidx * idxSz * 5
					x0 := xp0 + xsc*(uo+float32(xui))
					_, scaled, clr, _ := lm.View.UnitVal(lm.Lay, []int{zpi, xpi, zui, xui})
					v4c := mat32.NewVec4Color(clr)
					vshape.SetColor(clrAry, poff, 5*vtxSz, v4c)
					ht := 0.5 * mat32.Abs(scaled)
					if ht < MinUnitHeight {
						ht = MinUnitHeight
					}
					if scaled >= 0 {
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff, ioff, mat32.X, mat32.Y, -1, -1, xuw, ht, x0, 0, z0, segs, segs, pos)                     // nz
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+1*vtxSz, ioff+1*idxSz, mat32.Z, mat32.Y, -1, -1, zuw, ht, z0, 0, x0+xuw, segs, segs, pos) // px
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+2*vtxSz, ioff+2*idxSz, mat32.Z, mat32.Y, 1, -1, zuw, ht, z0, 0, x0, segs, segs, pos)      // nx
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+3*vtxSz, ioff+3*idxSz, mat32.X, mat32.Z, 1, 1, xuw, zuw, x0, z0, ht, segs, segs, pos)     // py <-
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+4*vtxSz, ioff+4*idxSz, mat32.X, mat32.Y, 1, -1, xuw, ht, x0, 0, z0+zuw, segs, segs, pos)  // pz
					} else {
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff, ioff, mat32.X, mat32.Y, 1, -1, xuw, ht, x0, -ht, z0, segs, segs, pos)                     // nz = pz norm
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+1*vtxSz, ioff+1*idxSz, mat32.Z, mat32.Y, 1, -1, zuw, ht, z0, -ht, x0+xuw, segs, segs, pos) // px = nx norm
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+2*vtxSz, ioff+2*idxSz, mat32.Z, mat32.Y, 1, -1, zuw, ht, z0, -ht, x0, segs, segs, pos)     // nx
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+3*vtxSz, ioff+3*idxSz, mat32.X, mat32.Z, 1, 1, xuw, zuw, x0, z0, -ht, segs, segs, pos)     // ny <-
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+4*vtxSz, ioff+4*idxSz, mat32.X, mat32.Y, 1, -1, xuw, ht, x0, -ht, z0+zuw, segs, segs, pos) // pz
					}
					pidx++
				}
			}
		}
	}
	lm.View.ReadUnlock()

	lm.BBoxMu.Lock()
	lm.BBox.SetBounds(mat32.Vec3{0, -0.5, -fnpz * fnuz}, mat32.Vec3{fnpx * fnux, 0.5, 0})
	lm.BBoxMu.Unlock()
}
