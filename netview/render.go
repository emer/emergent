// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import (
	"math"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/colors"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/tree"
	"cogentcore.org/core/types"
	"cogentcore.org/core/xyz"
	"github.com/emer/emergent/v2/emer"
)

// UpdateLayers updates the layer display with any structural or
// current data changes.  Very fast if no structural changes.
func (nv *NetView) UpdateLayers() {
	sw := nv.SceneWidget()
	se := sw.SceneXYZ()

	if nv.Net == nil || nv.Net.NumLayers() == 0 {
		se.DeleteChildren()
		se.Meshes.Reset()
		return
	}
	nb := nv.Net.AsEmer()
	if nv.NeedsRebuild() {
		se.Background = colors.Scheme.Background
	}
	nlay := nv.Net.NumLayers()
	laysGp := se.ChildByName("Layers", 0).(*xyz.Group)

	layConfig := tree.TypePlan{}
	for li := range nlay {
		ly := nv.Net.EmerLayer(li)
		layConfig.Add(types.For[xyz.Group](), ly.StyleName())
	}

	if !tree.Update(laysGp, layConfig) {
		for li := range laysGp.Children {
			ly := nv.Net.EmerLayer(li)
			lmesh := errors.Log1(se.MeshByName(ly.StyleName()))
			se.SetMesh(lmesh) // does update
		}
		if nv.hasPaths != nv.Params.Paths {
			nv.UpdatePaths()
		}
		return
	}

	gpConfig := tree.TypePlan{}
	gpConfig.Add(types.For[LayObj](), "layer")
	gpConfig.Add(types.For[LayName](), "name")

	nmin, nmax := nb.MinPos, nb.MaxPos
	nsz := nmax.Sub(nmin).Sub(math32.Vec3(1, 1, 0)).Max(math32.Vec3(1, 1, 1))
	nsc := math32.Vec3(1.0/nsz.X, 1.0/nsz.Y, 1.0/nsz.Z)
	szc := math32.Max(nsc.X, nsc.Y)
	poff := math32.Vector3Scalar(0.5)
	poff.Y = -0.5
	for li, lgi := range laysGp.Children {
		ly := nv.Net.EmerLayer(li)
		lb := ly.AsEmer()
		lmesh, _ := se.MeshByName(ly.StyleName())
		if lmesh == nil {
			NewLayMesh(se, nv, ly)
		} else {
			lmesh.(*LayMesh).Lay = ly // make sure
		}
		lg := lgi.(*xyz.Group)
		gpConfig[1].Name = ly.StyleName() // text2d textures use obj name, so must be unique
		tree.Update(lg, gpConfig)
		lp := lb.Pos.Pos
		lp.Y = -lp.Y // reverse direction
		lp = lp.Sub(nmin).Mul(nsc).Sub(poff)
		lg.Pose.Pos.Set(lp.X, lp.Z, lp.Y)
		lg.Pose.Scale.Set(nsc.X*lb.Pos.Scale, szc, nsc.Y*lb.Pos.Scale)

		lo := lg.Child(0).(*LayObj)
		lo.Defaults()
		lo.LayName = ly.StyleName()
		lo.NetView = nv
		lo.SetMeshName(ly.StyleName())
		lo.Material.Color = colors.FromRGB(255, 100, 255)
		lo.Material.Reflective = 8
		lo.Material.Bright = 8
		lo.Material.Shiny = 30
		// note: would actually be better to NOT cull back so you can view underneath
		// but then the front and back fight against each other, causing flickering

		txt := lg.Child(1).(*LayName)
		txt.Defaults()
		txt.NetView = nv
		txt.SetText(ly.StyleName())
		txt.Pose.Scale = math32.Vector3Scalar(nv.Params.LayNmSize).Div(lg.Pose.Scale)
		txt.Styles.Background = colors.Uniform(colors.Transparent)
		txt.Styles.Text.Align = styles.Start
		txt.Styles.Text.AlignV = styles.Start
	}
	nv.UpdatePaths()
	sw.XYZ.SetNeedsUpdate()
	sw.NeedsRender()
}

// UpdatePaths updates the path display.
// Only called when layers have structural changes.
func (nv *NetView) UpdatePaths() {
	sw := nv.SceneWidget()
	se := sw.SceneXYZ()

	nb := nv.Net.AsEmer()
	nlay := nv.Net.NumLayers()
	pathsGp := se.ChildByName("Paths", 0).(*xyz.Group)
	pathsGp.DeleteChildren()

	if !nv.Params.Paths {
		nv.hasPaths = false
		return
	}
	nv.hasPaths = true

	nmin, nmax := nb.MinPos, nb.MaxPos
	nsz := nmax.Sub(nmin).Sub(math32.Vec3(1, 1, 0)).Max(math32.Vec3(1, 1, 1))
	nsc := math32.Vec3(1.0/nsz.X, 1.0/nsz.Y, 1.0/nsz.Z)
	poff := math32.Vector3Scalar(0.5)
	poff.Y = -0.5

	lineWidth := nv.Params.PathWidth

	layPosSize := func(lb *emer.LayerBase) (math32.Vector3, math32.Vector3) {
		lp := lb.Pos.Pos
		lp.Y = -lp.Y
		lp = lp.Sub(nmin).Mul(nsc).Sub(poff)
		lp.Y, lp.Z = lp.Z, lp.Y
		dsz := lb.DisplaySize()
		lsz := math32.Vector3{dsz.X * nsc.X, 0, dsz.Y * nsc.Y}
		return lp, lsz
	}

	// F, L, R, B -- center of each side, z is negative; order favors front in a tie
	sideMids := []math32.Vector3{{0.5, 0, 0}, {0, 0, -0.5}, {1, 0, -0.5}, {0.5, 0, -1}}
	sideDims := []math32.Dims{math32.X, math32.Z, math32.Z, math32.X}

	sideMtx := func(side int, prop float32) math32.Vector3 {
		dim := sideDims[side]
		smat := sideMids[side]
		smat.SetDim(dim, prop)
		if dim == math32.Z {
			smat.Z *= -1
		}
		return smat
	}

	// 0 = forward, "left" side; 1 = lateral, "middle"; 2 = back, "right"
	sideCat := func(rLayY, sLayY float32) int {
		if rLayY < sLayY {
			return 2
		} else if rLayY == sLayY {
			return 1
		}
		return 0
	}

	type sideData struct {
		pth   emer.Path
		rSide int
	}

	for li := range nlay {
		ly := nv.Net.EmerLayer(li)
		lb := ly.AsEmer()
		sLayPos, sLaySz := layPosSize(lb)

		var sides [12][]sideData // by sending side * category
		npt := ly.NumSendPaths()
		for pi := range npt {
			pt := ly.SendPath(pi)
			rb := pt.RecvLayer().AsEmer()
			rLayPos, rLaySz := layPosSize(rb)
			minDist := float32(math.MaxFloat32)
			minSidx := 0
			minRside := 0
			for sSide := range 4 {
				for rSide := range 4 {
					cat := sideCat(rLayPos.Y, sLayPos.Y)
					prop := (float32(cat) + 0.5) * .3333
					smat := sideMtx(sSide, prop)
					rmat := sideMtx(rSide, prop)
					spos := sLayPos.Add(sLaySz.Mul(smat))
					rpos := rLayPos.Add(rLaySz.Mul(rmat))
					dist := rpos.Sub(spos).Length()
					if dist < minDist {
						minDist = dist
						minSidx = sSide*3 + cat
						minRside = rSide
					}
				}
			}
			sides[minSidx] = append(sides[minSidx], sideData{pth: pt, rSide: minRside})
		}
		for sSide := range 4 {
			for cat := range 3 {
				sidx := sSide*3 + cat
				pths := sides[sidx]
				npt := len(pths)
				if npt == 0 {
					continue
				}
				for pi, pd := range pths {
					pt := pd.pth
					rSide := pd.rSide
					rb := pt.RecvLayer().AsEmer()
					sb := pt.AsEmer()
					rLayPos, rLaySz := layPosSize(rb)
					off := float32(0.4)
					if rb.Index < lb.Index {
						off = 0.6
					}
					prop := 0.3333 * (float32(cat) + float32(pi) + off) / float32(npt)
					smat := sideMtx(sSide, prop)
					rmat := sideMtx(rSide, prop)
					spos := sLayPos.Add(sLaySz.Mul(smat))
					rpos := rLayPos.Add(rLaySz.Mul(rmat))
					// xyz.NewLine(se, pathsGp, sb.Name, spos, rpos, lineWidth, clr)
					clr := colors.Spaced(pt.TypeNumber())
					xyz.NewArrow(se, pathsGp, sb.Name, spos, rpos, lineWidth, clr, xyz.NoStartArrow, xyz.EndArrow, 4, .5, 4)
				}
			}
		}
	}
}
