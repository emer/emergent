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

func (lm *LayMesh) Make() {
	if lm.Lay == nil {
		lm.Shape.SetShape(nil, nil, nil)
		lm.Reset()
	}
	shp := lm.Lay.LayShape()
	// todo: optimize
	// if lm.Shape.IsEqual(shp) {
	// 	lm.Update()
	// 	return
	// }
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

func (lm *LayMesh) Make4D() {
}

func (lm *LayMesh) Make2D() {
	lm.Trans = true
	nz := lm.Shape.Dim(0)
	nx := lm.Shape.Dim(1)

	fnz := float32(nz)
	fnx := float32(nx)

	xw := lm.View.UnitSize / fnx
	xo := (1.0 - lm.View.UnitSize) / fnx
	zw := lm.View.UnitSize / fnz
	zo := (1.0 - lm.View.UnitSize) / fnz

	segs := 1

	for zi := nz - 1; zi >= 0; zi-- {
		z0 := zo + -0.5 + (fnz-(float32(zi)+1))/fnz
		for xi := 0; xi < nx; xi++ {
			x0 := xo + -0.5 + float32(xi)/fnx
			_, scaled, clr := lm.View.UnitVal(lm.Lay, []int{zi, xi})
			// clr = gi.Color{}
			ht := mat32.Abs(scaled)
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

	lm.BBox.BBox.Min = mat32.Vec3{-0.5, -0.5, -0.5}
	lm.BBox.BBox.Max = mat32.Vec3{0.5, 0.5, 0.5}
	lm.BBox.BSphere.Radius = lm.BBox.BBox.Min.Length()
	lm.BBox.Area = 2 + 2 + 2
	lm.BBox.Volume = 1
}

func (lm *LayMesh) Update() {
	lm.Make()
}
