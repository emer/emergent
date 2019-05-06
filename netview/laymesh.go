// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import (
	"github.com/emer/emergent/emer"
	"github.com/emer/etable/etensor"
	"github.com/goki/gi/gi3d"
	"github.com/goki/gi/mat32"
	"github.com/goki/ki/kit"
)

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

	if lm.Shape.Len() == 0 {
		return // nothing
	}

	if lm.Shape.NumDims() == 4 {
		lm.Make4D()
	} else {
		lm.Make2D()
	}
}

func (lm *LayMesh) Update(sc *gi3d.Scene) {
	lm.Make(sc)
	// todo: something is wrong with this Update2D method!
	// if lm.Shape.Len() == 0 {
	// 	return // nothing
	// }
	// if lm.Shape.NumDims() == 4 {
	// 	lm.Update4D()
	// } else {
	// 	lm.Update2D()
	// }
	lm.Activate(sc)
	lm.SetVtxData(sc)
	lm.SetColorData(sc)
	lm.TransferVectors()
}

func (lm *LayMesh) Make2D() {
	lm.Trans = true
	lm.Dynamic = true
	nz := lm.Shape.Dim(0)
	nx := lm.Shape.Dim(1)

	fnz := float32(nz)
	fnx := float32(nx)

	xw := lm.View.UnitSize
	xo := (1.0 - lm.View.UnitSize)
	zw := lm.View.UnitSize
	zo := (1.0 - lm.View.UnitSize)

	segs := 1

	for zi := nz - 1; zi >= 0; zi-- {
		z0 := zo - float32(zi+1)
		for xi := 0; xi < nx; xi++ {
			x0 := xo + float32(xi)
			_, scaled, clr := lm.View.UnitVal(lm.Lay, []int{zi, xi})
			ht := 0.5 * mat32.Abs(scaled)
			if scaled >= 0 {
				lm.AddPlane(mat32.X, mat32.Y, -1, -1, xw, ht, x0, 0, z0, segs, segs, clr)    // nz
				lm.AddPlane(mat32.Z, mat32.Y, -1, -1, zw, ht, z0, 0, x0+xw, segs, segs, clr) // px
				lm.AddPlane(mat32.Z, mat32.Y, 1, -1, zw, ht, z0, 0, x0, segs, segs, clr)     // nx
				lm.AddPlane(mat32.X, mat32.Z, 1, 1, xw, zw, x0, z0, ht, segs, segs, clr)     // py <-
				lm.AddPlane(mat32.X, mat32.Y, 1, -1, xw, ht, x0, 0, z0+zw, segs, segs, clr)  // pz
			} else {
				lm.AddPlane(mat32.X, mat32.Y, 1, -1, xw, ht, x0, -ht, z0, segs, segs, clr)    // nz = pz norm
				lm.AddPlane(mat32.Z, mat32.Y, 1, -1, zw, ht, z0, -ht, x0+xw, segs, segs, clr) // px = nx norm
				lm.AddPlane(mat32.Z, mat32.Y, 1, -1, zw, ht, z0, -ht, x0, segs, segs, clr)    // nx
				lm.AddPlane(mat32.X, mat32.Z, 1, 1, xw, zw, x0, z0, -ht, segs, segs, clr)     // ny <-
				lm.AddPlane(mat32.X, mat32.Y, 1, -1, xw, ht, x0, -ht, z0+zw, segs, segs, clr) // pz
			}
		}
	}

	lm.BBox.SetBounds(mat32.Vec3{0, -0.5, 0}, mat32.Vec3{fnx, 0.5, fnz})
}

func (lm *LayMesh) Update2D() {
	nz := lm.Shape.Dim(0)
	nx := lm.Shape.Dim(1)

	xw := lm.View.UnitSize
	xo := (1.0 - lm.View.UnitSize)
	zw := lm.View.UnitSize
	zo := (1.0 - lm.View.UnitSize)

	segs := 1

	psz := lm.PlaneSize(segs, segs)
	pidx := 0 // plane index
	for zi := nz - 1; zi >= 0; zi-- {
		z0 := zo - float32(zi+1)
		for xi := 0; xi < nx; xi++ {
			poff := pidx * psz * 5
			x0 := xo + float32(xi)
			_, scaled, clr := lm.View.UnitVal(lm.Lay, []int{zi, xi})
			ht := 0.5 * mat32.Abs(scaled)
			if scaled >= 0 {
				lm.SetPlaneVtx(poff, mat32.X, mat32.Y, -1, -1, xw, ht, x0, 0, z0, segs, segs, clr)          // nz
				lm.SetPlaneVtx(poff+1*psz, mat32.Z, mat32.Y, -1, -1, zw, ht, z0, 0, x0+xw, segs, segs, clr) // px
				lm.SetPlaneVtx(poff+2*psz, mat32.Z, mat32.Y, 1, -1, zw, ht, z0, 0, x0, segs, segs, clr)     // nx
				lm.SetPlaneVtx(poff+3*psz, mat32.X, mat32.Z, 1, 1, xw, zw, x0, z0, ht, segs, segs, clr)     // py <-
				lm.SetPlaneVtx(poff+4*psz, mat32.X, mat32.Y, 1, -1, xw, ht, x0, 0, z0+zw, segs, segs, clr)  // pz
			} else {
				lm.SetPlaneVtx(poff, mat32.X, mat32.Y, 1, -1, xw, ht, x0, -ht, z0, segs, segs, clr)          // nz = pz norm
				lm.SetPlaneVtx(poff+1*psz, mat32.Z, mat32.Y, 1, -1, zw, ht, z0, -ht, x0+xw, segs, segs, clr) // px = nx norm
				lm.SetPlaneVtx(poff+2*psz, mat32.Z, mat32.Y, 1, -1, zw, ht, z0, -ht, x0, segs, segs, clr)    // nx
				lm.SetPlaneVtx(poff+3*psz, mat32.X, mat32.Z, 1, 1, xw, zw, x0, z0, -ht, segs, segs, clr)     // ny <-
				lm.SetPlaneVtx(poff+4*psz, mat32.X, mat32.Y, 1, -1, xw, ht, x0, -ht, z0+zw, segs, segs, clr) // pz
			}
			pidx++
		}
	}
}

func (lm *LayMesh) Make4D() {
}

func (lm *LayMesh) Update4D() {
}
