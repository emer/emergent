// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import (
	"cogentcore.org/core/math32"
	"cogentcore.org/core/vgpu/vshape"
	"cogentcore.org/core/xyz"
)

func (lm *LayMesh) RasterSize2D() (nVtx, nIndex int) {
	rs := lm.Lay.RepShape()
	nuz := rs.DimSize(0)
	nux := rs.DimSize(1)
	nz := nuz*nux + nuz - 1
	nx := lm.View.Params.Raster.Max + 1
	segs := 1

	vtxSz, idxSz := vshape.PlaneN(segs, segs)
	nVtx = vtxSz * 5 * nz * nx
	nIndex = idxSz * 5 * nz * nx
	return
}

func (lm *LayMesh) RasterSize4D() (nVtx, nIndex int) {
	rs := lm.Lay.RepShape()
	npz := rs.DimSize(0) // p = pool
	npx := rs.DimSize(1)
	nuz := rs.DimSize(2) // u = unit
	nux := rs.DimSize(3)

	nz := nuz*nux + nuz - 1
	nx := lm.View.Params.Raster.Max + 1

	segs := 1

	vtxSz, idxSz := vshape.PlaneN(segs, segs)
	nVtx = vtxSz * 5 * npz * npx * nz * nx
	nIndex = idxSz * 5 * npz * npx * nz * nx
	return
}

func (lm *LayMesh) RasterSet2DX(sc *xyz.Scene, init bool, vtxAry, normAry, texAry, clrAry math32.ArrayF32, idxAry math32.ArrayU32) {
	rs := lm.Lay.RepShape()
	nuz := rs.DimSize(0)
	nux := rs.DimSize(1)
	nz := nuz*nux + nuz - 1
	nx := lm.View.Params.Raster.Max + 1
	htsc := 0.5 * lm.View.Params.Raster.UnitHeight

	fnoz := float32(lm.Shape.DimSize(0))
	fnox := float32(lm.Shape.DimSize(1))
	fnuz := float32(nuz)
	fnux := float32(nux)
	fnz := float32(nz)
	fnx := float32(nx)

	usz := lm.View.Params.Raster.UnitSize
	uo := (1.0 - usz)

	xsc := fnux / fnx
	zsc := fnuz / fnz

	// rescale rep -> full size
	xsc *= fnox / fnux
	zsc *= fnoz / fnuz

	xuw := xsc * usz
	zuw := zsc * usz

	segs := 1

	vtxSz, idxSz := vshape.PlaneN(segs, segs)
	pidx := 0 // plane index
	pos := math32.Vector3{}

	curRast, _ := lm.View.Data.RasterCtr(-1)

	lm.View.ReadLock()
	for zi := nz - 1; zi >= 0; zi-- {
		z0 := uo - zsc*float32(zi+1)
		uy := zi / (nux + 1)
		ux := zi % (nux + 1)
		xoff := 0
		for xi := 0; xi < nx; xi++ {
			poff := pidx * vtxSz * 5
			ioff := pidx * idxSz * 5
			x0 := uo + xsc*float32(xi)
			_, scaled, clr, _ := lm.View.UnitValRaster(lm.Lay, []int{uy, ux}, xi-xoff)
			if xi-1 == curRast || ux >= nux {
				clr = NilColor
				scaled = 0
			}
			if xi-1 == curRast {
				xoff++
			}
			v4c := math32.NewVector4Color(clr)
			vshape.SetColor(clrAry, poff, 5*vtxSz, v4c)
			ht := htsc * math32.Abs(scaled)
			if ht < MinUnitHeight {
				ht = MinUnitHeight
			}
			if scaled >= 0 {
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff, ioff, math32.X, math32.Y, -1, -1, xuw, ht, x0, 0, z0, segs, segs, pos)                     // nz
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+1*vtxSz, ioff+1*idxSz, math32.Z, math32.Y, -1, -1, zuw, ht, z0, 0, x0+xuw, segs, segs, pos) // px
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+2*vtxSz, ioff+2*idxSz, math32.Z, math32.Y, 1, -1, zuw, ht, z0, 0, x0, segs, segs, pos)      // nx
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+3*vtxSz, ioff+3*idxSz, math32.X, math32.Z, 1, 1, xuw, zuw, x0, z0, ht, segs, segs, pos)     // py <-
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+4*vtxSz, ioff+4*idxSz, math32.X, math32.Y, 1, -1, xuw, ht, x0, 0, z0+zuw, segs, segs, pos)  // pz
			} else {
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff, ioff, math32.X, math32.Y, 1, -1, xuw, ht, x0, -ht, z0, segs, segs, pos)                     // nz = pz norm
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+1*vtxSz, ioff+1*idxSz, math32.Z, math32.Y, 1, -1, zuw, ht, z0, -ht, x0+xuw, segs, segs, pos) // px = nx norm
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+2*vtxSz, ioff+2*idxSz, math32.Z, math32.Y, 1, -1, zuw, ht, z0, -ht, x0, segs, segs, pos)     // nx
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+3*vtxSz, ioff+3*idxSz, math32.X, math32.Z, 1, 1, xuw, zuw, x0, z0, -ht, segs, segs, pos)     // ny <-
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+4*vtxSz, ioff+4*idxSz, math32.X, math32.Y, 1, -1, xuw, ht, x0, -ht, z0+zuw, segs, segs, pos) // pz
			}
			pidx++
		}
	}
	lm.View.ReadUnlock()

	lm.BBoxMu.Lock()
	lm.BBox.SetBounds(math32.Vec3(0, -0.5, -fnz), math32.Vec3(fnx, 0.5, 0))
	lm.BBoxMu.Unlock()
}

func (lm *LayMesh) RasterSet2DZ(sc *xyz.Scene, init bool, vtxAry, normAry, texAry, clrAry math32.ArrayF32, idxAry math32.ArrayU32) {
	rs := lm.Lay.RepShape()
	nuz := rs.DimSize(0)
	nux := rs.DimSize(1)
	nx := nuz*nux + nuz - 1
	nz := lm.View.Params.Raster.Max + 1
	htsc := 0.5 * lm.View.Params.Raster.UnitHeight

	fnoz := float32(lm.Shape.DimSize(0))
	fnox := float32(lm.Shape.DimSize(1))
	fnuz := float32(nuz)
	fnux := float32(nux)
	fnz := float32(nz)
	fnx := float32(nx)

	usz := lm.View.Params.Raster.UnitSize
	uo := (1.0 - usz)

	xsc := fnux / fnx
	zsc := fnuz / fnz

	// rescale rep -> full size
	xsc *= fnox / fnux
	zsc *= fnoz / fnuz

	xuw := xsc * usz
	zuw := zsc * usz

	segs := 1

	vtxSz, idxSz := vshape.PlaneN(segs, segs)
	pidx := 0 // plane index
	pos := math32.Vector3{}

	curRast, _ := lm.View.Data.RasterCtr(-1)

	lm.View.ReadLock()
	zoff := 1
	for zi := nz - 1; zi >= 0; zi-- {
		z0 := uo - zsc*float32(zi+1)
		for xi := 0; xi < nx; xi++ {
			uy := xi / (nux + 1)
			ux := xi % (nux + 1)
			poff := pidx * vtxSz * 5
			ioff := pidx * idxSz * 5
			x0 := uo + xsc*float32(xi)
			_, scaled, clr, _ := lm.View.UnitValRaster(lm.Lay, []int{uy, ux}, zi-zoff)
			if zi-1 == curRast || ux >= nux {
				clr = NilColor
				scaled = 0
			}
			if zi-1 == curRast {
				zoff = 0
			}
			v4c := math32.NewVector4Color(clr)
			vshape.SetColor(clrAry, poff, 5*vtxSz, v4c)
			ht := htsc * math32.Abs(scaled)
			if ht < MinUnitHeight {
				ht = MinUnitHeight
			}
			if scaled >= 0 {
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff, ioff, math32.X, math32.Y, -1, -1, xuw, ht, x0, 0, z0, segs, segs, pos)                     // nz
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+1*vtxSz, ioff+1*idxSz, math32.Z, math32.Y, -1, -1, zuw, ht, z0, 0, x0+xuw, segs, segs, pos) // px
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+2*vtxSz, ioff+2*idxSz, math32.Z, math32.Y, 1, -1, zuw, ht, z0, 0, x0, segs, segs, pos)      // nx
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+3*vtxSz, ioff+3*idxSz, math32.X, math32.Z, 1, 1, xuw, zuw, x0, z0, ht, segs, segs, pos)     // py <-
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+4*vtxSz, ioff+4*idxSz, math32.X, math32.Y, 1, -1, xuw, ht, x0, 0, z0+zuw, segs, segs, pos)  // pz
			} else {
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff, ioff, math32.X, math32.Y, 1, -1, xuw, ht, x0, -ht, z0, segs, segs, pos)                     // nz = pz norm
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+1*vtxSz, ioff+1*idxSz, math32.Z, math32.Y, 1, -1, zuw, ht, z0, -ht, x0+xuw, segs, segs, pos) // px = nx norm
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+2*vtxSz, ioff+2*idxSz, math32.Z, math32.Y, 1, -1, zuw, ht, z0, -ht, x0, segs, segs, pos)     // nx
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+3*vtxSz, ioff+3*idxSz, math32.X, math32.Z, 1, 1, xuw, zuw, x0, z0, -ht, segs, segs, pos)     // ny <-
				vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+4*vtxSz, ioff+4*idxSz, math32.X, math32.Y, 1, -1, xuw, ht, x0, -ht, z0+zuw, segs, segs, pos) // pz
			}
			pidx++
		}
	}
	lm.View.ReadUnlock()

	lm.BBoxMu.Lock()
	lm.BBox.SetBounds(math32.Vec3(0, -0.5, -fnz), math32.Vec3(fnx, 0.5, 0))
	lm.BBoxMu.Unlock()
}

func (lm *LayMesh) RasterSet4DX(sc *xyz.Scene, init bool, vtxAry, normAry, texAry, clrAry math32.ArrayF32, idxAry math32.ArrayU32) {
	rs := lm.Lay.RepShape()
	npz := rs.DimSize(0) // p = pool
	npx := rs.DimSize(1)
	nuz := rs.DimSize(2) // u = unit
	nux := rs.DimSize(3)

	nz := nuz*nux + nuz - 1
	nx := lm.View.Params.Raster.Max + 1
	htsc := 0.5 * lm.View.Params.Raster.UnitHeight

	fnpoz := float32(lm.Shape.DimSize(0))
	fnpox := float32(lm.Shape.DimSize(1))
	fnpz := float32(npz)
	fnpx := float32(npx)
	fnuz := float32(nuz)
	fnux := float32(nux)
	fnx := float32(nx)
	fnz := float32(nz)

	usz := lm.View.Params.UnitSize
	uo := 2.0 * (1.0 - usz) // offset = space

	// for 4D, we build in spaces between groups without changing the overall size of layer
	// by shrinking the spacing of each unit according to the spaces we introduce
	// these scales are for overall group positioning
	xsc := (fnpx * fnux) / ((fnpx-1)*uo + (fnpx * fnux))
	zsc := (fnpz * fnuz) / ((fnpz-1)*uo + (fnpz * fnuz))

	// rescale rep -> full size
	xsc *= fnpox / fnpx
	zsc *= fnpoz / fnpz

	// these are for the raster within
	xscr := xsc * (fnux / fnx)
	zscr := zsc * (fnuz / fnz)

	uszr := lm.View.Params.Raster.UnitSize
	uor := (1.0 - uszr) // offset = space

	xuw := xscr * uszr
	zuw := zscr * uszr

	segs := 1

	vtxSz, idxSz := vshape.PlaneN(segs, segs)
	pidx := 0 // plane index
	pos := math32.Vector3{}

	curRast, _ := lm.View.Data.RasterCtr(-1)

	lm.View.ReadLock()
	for zpi := npz - 1; zpi >= 0; zpi-- {
		zp0 := zsc * (-float32(zpi) * (uo + fnuz))
		for xpi := 0; xpi < npx; xpi++ {
			xp0 := xsc * (float32(xpi)*uo + float32(xpi)*fnux)
			for zi := nz - 1; zi >= 0; zi-- {
				z0 := zp0 + zscr*(uor-float32(zi+1))
				uy := zi / (nux + 1)
				ux := zi % (nux + 1)
				xoff := 0
				for xi := 0; xi < nx; xi++ {
					poff := pidx * vtxSz * 5
					ioff := pidx * idxSz * 5
					x0 := xp0 + xscr*(uor+float32(xi))
					_, scaled, clr, _ := lm.View.UnitValRaster(lm.Lay, []int{zpi, xpi, uy, ux}, xi-xoff)
					if xi-1 == curRast || ux >= nux {
						clr = NilColor
						scaled = 0
					}
					if xi-1 == curRast {
						xoff++
					}
					v4c := math32.NewVector4Color(clr)
					vshape.SetColor(clrAry, poff, 5*vtxSz, v4c)
					ht := htsc * math32.Abs(scaled)
					if ht < MinUnitHeight {
						ht = MinUnitHeight
					}
					if scaled >= 0 {
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff, ioff, math32.X, math32.Y, -1, -1, xuw, ht, x0, 0, z0, segs, segs, pos)                     // nz
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+1*vtxSz, ioff+1*idxSz, math32.Z, math32.Y, -1, -1, zuw, ht, z0, 0, x0+xuw, segs, segs, pos) // px
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+2*vtxSz, ioff+2*idxSz, math32.Z, math32.Y, 1, -1, zuw, ht, z0, 0, x0, segs, segs, pos)      // nx
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+3*vtxSz, ioff+3*idxSz, math32.X, math32.Z, 1, 1, xuw, zuw, x0, z0, ht, segs, segs, pos)     // py <-
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+4*vtxSz, ioff+4*idxSz, math32.X, math32.Y, 1, -1, xuw, ht, x0, 0, z0+zuw, segs, segs, pos)  // pz
					} else {
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff, ioff, math32.X, math32.Y, 1, -1, xuw, ht, x0, -ht, z0, segs, segs, pos)                     // nz = pz norm
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+1*vtxSz, ioff+1*idxSz, math32.Z, math32.Y, 1, -1, zuw, ht, z0, -ht, x0+xuw, segs, segs, pos) // px = nx norm
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+2*vtxSz, ioff+2*idxSz, math32.Z, math32.Y, 1, -1, zuw, ht, z0, -ht, x0, segs, segs, pos)     // nx
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+3*vtxSz, ioff+3*idxSz, math32.X, math32.Z, 1, 1, xuw, zuw, x0, z0, -ht, segs, segs, pos)     // ny <-
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+4*vtxSz, ioff+4*idxSz, math32.X, math32.Y, 1, -1, xuw, ht, x0, -ht, z0+zuw, segs, segs, pos) // pz
					}
					pidx++
				}
			}
		}
	}
	lm.View.ReadUnlock()

	lm.BBoxMu.Lock()
	lm.BBox.SetBounds(math32.Vec3(0, -0.5, -fnpoz*fnuz), math32.Vec3(fnpox*fnux, 0.5, 0))
	lm.BBoxMu.Unlock()
}

func (lm *LayMesh) RasterSet4DZ(sc *xyz.Scene, init bool, vtxAry, normAry, texAry, clrAry math32.ArrayF32, idxAry math32.ArrayU32) {
	rs := lm.Lay.RepShape()
	npz := rs.DimSize(0) // p = pool
	npx := rs.DimSize(1)
	nuz := rs.DimSize(2) // u = unit
	nux := rs.DimSize(3)

	nx := nuz*nux + nuz - 1
	nz := lm.View.Params.Raster.Max + 1
	htsc := 0.5 * lm.View.Params.Raster.UnitHeight

	fnpoz := float32(lm.Shape.DimSize(0))
	fnpox := float32(lm.Shape.DimSize(1))
	fnpz := float32(npz)
	fnpx := float32(npx)
	fnuz := float32(nuz)
	fnux := float32(nux)
	fnx := float32(nx)
	fnz := float32(nz)

	usz := lm.View.Params.UnitSize
	uo := 2.0 * (1.0 - usz) // offset = space

	// for 4D, we build in spaces between groups without changing the overall size of layer
	// by shrinking the spacing of each unit according to the spaces we introduce
	// these scales are for overall group positioning
	xsc := (fnpx * fnux) / ((fnpx-1)*uo + (fnpx * fnux))
	zsc := (fnpz * fnuz) / ((fnpz-1)*uo + (fnpz * fnuz))

	// rescale rep -> full size
	xsc *= fnpox / fnpx
	zsc *= fnpoz / fnpz

	// these are for the raster within
	xscr := xsc * (fnux / fnx)
	zscr := zsc * (fnuz / fnz)

	uszr := lm.View.Params.Raster.UnitSize
	uor := (1.0 - uszr) // offset = space

	xuw := xscr * uszr
	zuw := zscr * uszr

	segs := 1

	vtxSz, idxSz := vshape.PlaneN(segs, segs)
	pidx := 0 // plane index
	pos := math32.Vector3{}

	curRast, _ := lm.View.Data.RasterCtr(-1)

	lm.View.ReadLock()
	for zpi := npz - 1; zpi >= 0; zpi-- {
		zp0 := zsc * (-float32(zpi) * (uo + fnuz))
		for xpi := 0; xpi < npx; xpi++ {
			xp0 := xsc * (float32(xpi)*uo + float32(xpi)*fnux)
			zoff := 1
			for zi := nz - 1; zi >= 0; zi-- {
				z0 := zp0 + zscr*(uor-float32(zi+1))
				for xi := 0; xi < nx; xi++ {
					uy := xi / (nux + 1)
					ux := xi % (nux + 1)
					poff := pidx * vtxSz * 5
					ioff := pidx * idxSz * 5
					x0 := xp0 + xscr*(uor+float32(xi))
					_, scaled, clr, _ := lm.View.UnitValRaster(lm.Lay, []int{zpi, xpi, uy, ux}, zi-zoff)
					if zi-1 == curRast || ux >= nux {
						clr = NilColor
						scaled = 0
					}
					if zi-1 == curRast {
						zoff = 0
					}
					v4c := math32.NewVector4Color(clr)
					vshape.SetColor(clrAry, poff, 5*vtxSz, v4c)
					ht := htsc * math32.Abs(scaled)
					if ht < MinUnitHeight {
						ht = MinUnitHeight
					}
					if scaled >= 0 {
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff, ioff, math32.X, math32.Y, -1, -1, xuw, ht, x0, 0, z0, segs, segs, pos)                     // nz
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+1*vtxSz, ioff+1*idxSz, math32.Z, math32.Y, -1, -1, zuw, ht, z0, 0, x0+xuw, segs, segs, pos) // px
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+2*vtxSz, ioff+2*idxSz, math32.Z, math32.Y, 1, -1, zuw, ht, z0, 0, x0, segs, segs, pos)      // nx
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+3*vtxSz, ioff+3*idxSz, math32.X, math32.Z, 1, 1, xuw, zuw, x0, z0, ht, segs, segs, pos)     // py <-
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+4*vtxSz, ioff+4*idxSz, math32.X, math32.Y, 1, -1, xuw, ht, x0, 0, z0+zuw, segs, segs, pos)  // pz
					} else {
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff, ioff, math32.X, math32.Y, 1, -1, xuw, ht, x0, -ht, z0, segs, segs, pos)                     // nz = pz norm
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+1*vtxSz, ioff+1*idxSz, math32.Z, math32.Y, 1, -1, zuw, ht, z0, -ht, x0+xuw, segs, segs, pos) // px = nx norm
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+2*vtxSz, ioff+2*idxSz, math32.Z, math32.Y, 1, -1, zuw, ht, z0, -ht, x0, segs, segs, pos)     // nx
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+3*vtxSz, ioff+3*idxSz, math32.X, math32.Z, 1, 1, xuw, zuw, x0, z0, -ht, segs, segs, pos)     // ny <-
						vshape.SetPlane(vtxAry, normAry, texAry, idxAry, poff+4*vtxSz, ioff+4*idxSz, math32.X, math32.Y, 1, -1, xuw, ht, x0, -ht, z0+zuw, segs, segs, pos) // pz
					}
					pidx++
				}
			}
		}
	}
	lm.View.ReadUnlock()

	lm.BBoxMu.Lock()
	lm.BBox.SetBounds(math32.Vec3(0, -0.5, -fnpoz*fnuz), math32.Vec3(fnpox*fnux, 0.5, 0))
	lm.BBoxMu.Unlock()
}
