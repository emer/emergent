// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import (
	"github.com/goki/gi/gi3d"
	"github.com/goki/mat32"
	"github.com/goki/vgpu/vshape"
)

func (lm *LayMesh) RasterSize2D() (nVtx, nIdx int) {
	nUy := lm.Shape.Dim(0)
	nUx := lm.Shape.Dim(1)
	nz := nUy*nUx + nUy - 1
	nx := lm.View.Params.Raster.Max + 1
	segs := 1

	vtxSz, idxSz := vshape.PlaneN(segs, segs)
	nVtx = vtxSz * 5 * nz * nx
	nIdx = idxSz * 5 * nz * nx
	return
}

func (lm *LayMesh) RasterSize4D() (nVtx, nIdx int) {
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

func (lm *LayMesh) RasterSet2DX(sc *gi3d.Scene, init bool, vtxAry, normAry, texAry, clrAry mat32.ArrayF32, idxAry mat32.ArrayU32) {
	nUy := lm.Shape.Dim(0)
	nUx := lm.Shape.Dim(1)
	nz := nUy*nUx + nUy - 1
	nx := lm.View.Params.Raster.Max + 1
	htsc := lm.View.Params.Raster.UnitHeight

	fnz := float32(nz)
	fnx := float32(nx)

	usz := lm.View.Params.Raster.UnitSize
	uo := (1.0 - usz)

	xsc := float32(nUx) / fnx
	zsc := float32(nUy) / fnz

	xuw := xsc * usz
	zuw := zsc * usz

	segs := 1

	vtxSz, idxSz := vshape.PlaneN(segs, segs)
	pidx := 0 // plane index
	pos := mat32.Vec3{}

	curRast, _ := lm.View.Data.RasterCtr(-1)

	lm.View.ReadLock()
	for zi := nz - 1; zi >= 0; zi-- {
		z0 := uo - zsc*float32(zi+1)
		uy := zi / (nUx + 1)
		ux := zi % (nUx + 1)
		xoff := 0
		for xi := 0; xi < nx; xi++ {
			poff := pidx * vtxSz * 5
			ioff := pidx * idxSz * 5
			x0 := uo + xsc*float32(xi)
			_, scaled, clr, _ := lm.View.UnitValRaster(lm.Lay, []int{uy, ux}, xi-xoff)
			if xi-1 == curRast || ux >= nUx {
				clr = NilColor
				scaled = 0
				xoff++
			}
			v4c := mat32.NewVec4Color(clr)
			vshape.SetColor(clrAry, poff, 5*vtxSz, v4c)
			ht := htsc * 0.5 * mat32.Abs(scaled)
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
	lm.View.ReadUnlock()

	lm.BBoxMu.Lock()
	lm.BBox.SetBounds(mat32.Vec3{0, -0.5, -fnz}, mat32.Vec3{fnx, 0.5, 0})
	lm.BBoxMu.Unlock()
}

func (lm *LayMesh) RasterSet2DZ(sc *gi3d.Scene, init bool, vtxAry, normAry, texAry, clrAry mat32.ArrayF32, idxAry mat32.ArrayU32) {
	nUy := lm.Shape.Dim(0)
	nUx := lm.Shape.Dim(1)
	nx := nUy*nUx + nUy - 1
	nz := lm.View.Params.Raster.Max + 1
	htsc := lm.View.Params.Raster.UnitHeight

	fnz := float32(nz)
	fnx := float32(nx)

	usz := lm.View.Params.Raster.UnitSize
	uo := (1.0 - usz)

	xsc := float32(nUx) / fnx
	zsc := float32(nUy) / fnz

	xuw := xsc * usz
	zuw := zsc * usz

	segs := 1

	vtxSz, idxSz := vshape.PlaneN(segs, segs)
	pidx := 0 // plane index
	pos := mat32.Vec3{}

	curRast, _ := lm.View.Data.RasterCtr(-1)

	lm.View.ReadLock()
	zoff := 1
	for zi := nz - 1; zi >= 0; zi-- {
		z0 := uo - zsc*float32(zi+1)
		for xi := 0; xi < nx; xi++ {
			uy := xi / (nUx + 1)
			ux := xi % (nUx + 1)
			poff := pidx * vtxSz * 5
			ioff := pidx * idxSz * 5
			x0 := uo + xsc*float32(xi)
			_, scaled, clr, _ := lm.View.UnitValRaster(lm.Lay, []int{uy, ux}, zi-zoff)
			if zi-1 == curRast || ux >= nUx {
				clr = NilColor
				scaled = 0
				zoff = 0
			}
			v4c := mat32.NewVec4Color(clr)
			vshape.SetColor(clrAry, poff, 5*vtxSz, v4c)
			ht := htsc * 0.5 * mat32.Abs(scaled)
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
	lm.View.ReadUnlock()

	lm.BBoxMu.Lock()
	lm.BBox.SetBounds(mat32.Vec3{0, -0.5, -fnz}, mat32.Vec3{fnx, 0.5, 0})
	lm.BBoxMu.Unlock()
}

func (lm *LayMesh) RasterSet4D(sc *gi3d.Scene, init bool, vtxAry, normAry, texAry, clrAry mat32.ArrayF32, idxAry mat32.ArrayU32) {
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
