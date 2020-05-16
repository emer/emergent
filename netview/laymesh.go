// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import (
	"github.com/emer/emergent/emer"
	"github.com/emer/etable/etensor"
	"github.com/goki/gi/gi3d"
	"github.com/goki/ki/kit"
	"github.com/goki/mat32"
)

// LayMesh is a gi3d.Mesh that represents a layer -- it is dynamically updated using the
// Update method which only resets the essential Vertex elements.
// The geometry is literal in the layer size: 0,0,0 lower-left corner and increasing X,Z
// for the width and height of the layer, in unit (1) increments per unit..
// NetView applies an overall scaling to make it fit within the larger view.
type LayMesh struct {
	gi3d.MeshBase
	Lay   emer.Layer    `desc:"layer that we render"`
	Shape etensor.Shape `desc:"current shape that has been constructed -- if same, just update"`
	View  *NetView      `desc:"netview that we're in"`
}

var KiT_LayMesh = kit.Types.AddType(&LayMesh{}, nil)

// AddNewLayMesh adds LayMesh mesh to given scene for given layer
func AddNewLayMesh(sc *gi3d.Scene, nv *NetView, lay emer.Layer) *LayMesh {
	lm := &LayMesh{}
	lm.View = nv
	lm.Lay = lay
	lm.Nm = lay.Name()
	sc.AddMesh(lm)
	return lm
}

func (lm *LayMesh) Make(sc *gi3d.Scene) {
	if lm.Lay == nil {
		lm.Shape.SetShape(nil, nil, nil)
		lm.Reset()
	}
	shp := lm.Lay.Shape()
	lm.Reset()
	lm.Shape.CopyShape(shp)

	if lm.Shape.NumDims() == 0 {
		return // nothing
	}

	if lm.Shape.NumDims() == 4 {
		lm.Make4D(true) // true = init
	} else {
		lm.Make2D(true)
	}
}

func (lm *LayMesh) Update(sc *gi3d.Scene) {
	if lm.Shape.NumDims() == 0 {
		return // nothing
	}
	if lm.Shape.NumDims() == 4 {
		lm.Make4D(false) // false = not init
	} else {
		lm.Make2D(false)
	}
	lm.SetVtxData(sc)
	lm.SetColorData(sc)
	lm.SetNormData(sc)
	lm.Activate(sc)
	lm.TransferVectors()
}

// MinUnitHeight ensures that there is always at least some dimensionality
// to the unit cubes -- affects transparency rendering etc
var MinUnitHeight = float32(1.0e-6)

func (lm *LayMesh) Make2D(init bool) {
	lm.Trans = true
	lm.Dynamic = true
	nz := lm.Shape.Dim(0)
	nx := lm.Shape.Dim(1)

	fnz := float32(nz)
	fnx := float32(nx)

	uw := lm.View.Params.UnitSize
	uo := (1.0 - uw)
	segs := 1

	vtxSz, idxSz := lm.PlaneSize(segs, segs)
	nvtx := vtxSz * 5 * nz * nx
	nidx := idxSz * 5 * nz * nx
	lm.Alloc(nvtx, nidx, true)

	pidx := 0 // plane index

	setNorm := true // can change -- always set
	setTex := init
	setIdx := init

	lm.View.ReadLock()
	for zi := nz - 1; zi >= 0; zi-- {
		z0 := uo - float32(zi+1)
		for xi := 0; xi < nx; xi++ {
			poff := pidx * vtxSz * 5
			ioff := pidx * idxSz * 5
			x0 := uo + float32(xi)
			_, scaled, clr := lm.View.UnitVal(lm.Lay, []int{zi, xi})
			ht := 0.5 * mat32.Abs(scaled)
			if ht < MinUnitHeight {
				ht = MinUnitHeight
			}
			if scaled >= 0 {
				lm.SetPlane(poff, ioff, setNorm, setTex, setIdx, mat32.X, mat32.Y, -1, -1, uw, ht, x0, 0, z0, segs, segs, clr)                    // nz
				lm.SetPlane(poff+1*vtxSz, ioff+1*idxSz, setNorm, setTex, setIdx, mat32.Z, mat32.Y, -1, -1, uw, ht, z0, 0, x0+uw, segs, segs, clr) // px
				lm.SetPlane(poff+2*vtxSz, ioff+2*idxSz, setNorm, setTex, setIdx, mat32.Z, mat32.Y, 1, -1, uw, ht, z0, 0, x0, segs, segs, clr)     // nx
				lm.SetPlane(poff+3*vtxSz, ioff+3*idxSz, setNorm, setTex, setIdx, mat32.X, mat32.Z, 1, 1, uw, uw, x0, z0, ht, segs, segs, clr)     // py <-
				lm.SetPlane(poff+4*vtxSz, ioff+4*idxSz, setNorm, setTex, setIdx, mat32.X, mat32.Y, 1, -1, uw, ht, x0, 0, z0+uw, segs, segs, clr)  // pz
			} else {
				lm.SetPlane(poff, ioff, setNorm, setTex, setIdx, mat32.X, mat32.Y, 1, -1, uw, ht, x0, -ht, z0, segs, segs, clr)                    // nz = pz norm
				lm.SetPlane(poff+1*vtxSz, ioff+1*idxSz, setNorm, setTex, setIdx, mat32.Z, mat32.Y, 1, -1, uw, ht, z0, -ht, x0+uw, segs, segs, clr) // px = nx norm
				lm.SetPlane(poff+2*vtxSz, ioff+2*idxSz, setNorm, setTex, setIdx, mat32.Z, mat32.Y, 1, -1, uw, ht, z0, -ht, x0, segs, segs, clr)    // nx
				lm.SetPlane(poff+3*vtxSz, ioff+3*idxSz, setNorm, setTex, setIdx, mat32.X, mat32.Z, 1, 1, uw, uw, x0, z0, -ht, segs, segs, clr)     // ny <-
				lm.SetPlane(poff+4*vtxSz, ioff+4*idxSz, setNorm, setTex, setIdx, mat32.X, mat32.Y, 1, -1, uw, ht, x0, -ht, z0+uw, segs, segs, clr) // pz
			}
			pidx++
		}
	}
	lm.View.ReadUnlock()

	lm.BBoxMu.Lock()
	lm.BBox.SetBounds(mat32.Vec3{0, -0.5, -fnz}, mat32.Vec3{fnx, 0.5, 0})
	lm.BBoxMu.Unlock()
}

func (lm *LayMesh) Make4D(init bool) {
	lm.Trans = true
	lm.Dynamic = true
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

	vtxSz, idxSz := lm.PlaneSize(segs, segs)
	nvtx := vtxSz * 5 * npz * npx * nuz * nux
	nidx := idxSz * 5 * npz * npx * nuz * nux
	lm.Alloc(nvtx, nidx, true)

	pidx := 0 // plane index

	setNorm := true // can change -- always set
	setTex := init
	setIdx := init

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
					_, scaled, clr := lm.View.UnitVal(lm.Lay, []int{zpi, xpi, zui, xui})
					ht := 0.5 * mat32.Abs(scaled)
					if ht < MinUnitHeight {
						ht = MinUnitHeight
					}
					if scaled >= 0 {
						lm.SetPlane(poff, ioff, setNorm, setTex, setIdx, mat32.X, mat32.Y, -1, -1, xuw, ht, x0, 0, z0, segs, segs, clr)                     // nz
						lm.SetPlane(poff+1*vtxSz, ioff+1*idxSz, setNorm, setTex, setIdx, mat32.Z, mat32.Y, -1, -1, zuw, ht, z0, 0, x0+xuw, segs, segs, clr) // px
						lm.SetPlane(poff+2*vtxSz, ioff+2*idxSz, setNorm, setTex, setIdx, mat32.Z, mat32.Y, 1, -1, zuw, ht, z0, 0, x0, segs, segs, clr)      // nx
						lm.SetPlane(poff+3*vtxSz, ioff+3*idxSz, setNorm, setTex, setIdx, mat32.X, mat32.Z, 1, 1, xuw, zuw, x0, z0, ht, segs, segs, clr)     // py <-
						lm.SetPlane(poff+4*vtxSz, ioff+4*idxSz, setNorm, setTex, setIdx, mat32.X, mat32.Y, 1, -1, xuw, ht, x0, 0, z0+zuw, segs, segs, clr)  // pz
					} else {
						lm.SetPlane(poff, ioff, setNorm, setTex, setIdx, mat32.X, mat32.Y, 1, -1, xuw, ht, x0, -ht, z0, segs, segs, clr)                     // nz = pz norm
						lm.SetPlane(poff+1*vtxSz, ioff+1*idxSz, setNorm, setTex, setIdx, mat32.Z, mat32.Y, 1, -1, zuw, ht, z0, -ht, x0+xuw, segs, segs, clr) // px = nx norm
						lm.SetPlane(poff+2*vtxSz, ioff+2*idxSz, setNorm, setTex, setIdx, mat32.Z, mat32.Y, 1, -1, zuw, ht, z0, -ht, x0, segs, segs, clr)     // nx
						lm.SetPlane(poff+3*vtxSz, ioff+3*idxSz, setNorm, setTex, setIdx, mat32.X, mat32.Z, 1, 1, xuw, zuw, x0, z0, -ht, segs, segs, clr)     // ny <-
						lm.SetPlane(poff+4*vtxSz, ioff+4*idxSz, setNorm, setTex, setIdx, mat32.X, mat32.Y, 1, -1, xuw, ht, x0, -ht, z0+zuw, segs, segs, clr) // pz
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
