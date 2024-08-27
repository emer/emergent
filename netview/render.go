// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import (
	"fmt"
	"math"
	"strings"

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

	if !tree.Update(laysGp, layConfig) && nv.layerNameSizeShown == nv.Options.LayerNameSize {
		for li := range laysGp.Children {
			ly := nv.Net.EmerLayer(li)
			lmesh := errors.Log1(se.MeshByName(ly.StyleName()))
			se.SetMesh(lmesh) // does update
		}
		if nv.hasPaths != nv.Options.Paths || nv.pathTypeShown != nv.Options.PathType ||
			nv.pathWidthShown != nv.Options.PathWidth {
			nv.UpdatePaths()
		}
		return
	}
	nv.layerNameSizeShown = nv.Options.LayerNameSize

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
		txt.Pose.Scale = math32.Vector3Scalar(nv.Options.LayerNameSize).Div(lg.Pose.Scale)
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

	if !nv.Options.Paths {
		nv.hasPaths = false
		return
	}
	nv.hasPaths = true
	nv.pathTypeShown = nv.Options.PathType
	nv.pathWidthShown = nv.Options.PathWidth

	nmin, nmax := nb.MinPos, nb.MaxPos
	nsz := nmax.Sub(nmin).Sub(math32.Vec3(1, 1, 0)).Max(math32.Vec3(1, 1, 1))
	nsc := math32.Vec3(1.0/nsz.X, 1.0/nsz.Y, 1.0/nsz.Z)
	poff := math32.Vector3Scalar(0.5)
	poff.Y = -0.5

	lineWidth := nv.Options.PathWidth

	// weight factors applied to distance for the different sides,
	// to encourage / discourage choice of sides.
	// In general the sides are preferred, and back is discouraged.
	sideWeights := [4]float32{1.1, 1, 1, 1.1}

	type pathData struct {
		path               emer.Path
		sSide, rSide, cat  int
		sIdx, sN, rIdx, rN int // indexes and numbers for each side
		sPos, rPos         math32.Vector3
	}

	pdIdx := func(side, cat int) int {
		return side*3 + cat
	}

	type layerData struct {
		paths     [12][]*pathData // by side * category
		selfPaths []*pathData
	}

	layPaths := make([]layerData, nlay)

	// 0 = forward, "left" side; 1 = lateral, "middle"; 2 = back, "right"
	sideCat := func(rLayY, sLayY float32) int {
		if rLayY < sLayY {
			return 2
		} else if rLayY == sLayY {
			return 1
		}
		return 0
	}

	// returns layer position and size in normalized display coordinates (NDC)
	// using the correct rendering coordinate system: X = X, Y <-> Z
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

	// returns the matrix
	sideMtx := func(side int, prop float32) math32.Vector3 {
		dim := sideDims[side]
		smat := sideMids[side]
		smat.SetDim(dim, prop)
		if dim == math32.Z {
			smat.Z *= -1
		}
		return smat
	}

	laySidePos := func(lb *emer.LayerBase, side, cat, idx, n int, off float32) math32.Vector3 {
		prop := (float32(cat) / 3) + (float32(idx)+off)/float32(3*n)
		pos, sz := layPosSize(lb)
		mat := sideMtx(side, prop)
		return pos.Add(sz.Mul(mat))
	}

	// returns the sending, recv positions of the path,
	// for point at given index along side, cat
	setPathPos := func(pd *pathData) {
		pt := pd.path
		sb := pt.SendLayer().AsEmer()
		rb := pt.RecvLayer().AsEmer()
		off := float32(0.4)
		if rb.Index < sb.Index {
			off = 0.6
		}
		pd.sPos = laySidePos(sb, pd.sSide, pd.cat, pd.sIdx, pd.sN, off)
		pd.rPos = laySidePos(rb, pd.rSide, pd.cat, pd.rIdx, pd.rN, off)
		return
	}

	// first pass: find the side to make connections on, based on shortest weighted length
	for si := range nlay {
		sl := nv.Net.EmerLayer(si)
		sb := sl.AsEmer()
		slayData := &layPaths[sb.Index]
		sLayPos, _ := layPosSize(sb)

		npt := sl.NumSendPaths()
		for pi := range npt {
			pt := sl.SendPath(pi)
			if !nv.pathTypeNameMatch(pt.TypeName()) {
				continue
			}
			rb := pt.RecvLayer().AsEmer()
			if sb.Index == rb.Index { // self
				slayData.selfPaths = append(slayData.selfPaths, &pathData{path: pt, cat: 1})
				continue
			}
			minDist := float32(math.MaxFloat32)
			var minData *pathData
			for sSide := range 4 {
				swt := sideWeights[sSide]
				for rSide := range 4 {
					rwt := sideWeights[rSide]
					rLayPos, _ := layPosSize(rb)
					cat := sideCat(rLayPos.Y, sLayPos.Y)
					pd := &pathData{path: pt, sSide: sSide, rSide: rSide, cat: cat, sN: 1, rN: 1}
					setPathPos(pd)
					dist := pd.rPos.Sub(pd.sPos).Length() * swt * rwt
					if dist < minDist {
						minDist = dist
						minData = pd
					}
				}
			}
			i := pdIdx(minData.sSide, minData.cat)
			minData.sIdx = len(slayData.paths[i])
			slayData.paths[i] = append(slayData.paths[i], minData)
			for _, pd := range slayData.paths[i] {
				pd.sN = len(slayData.paths[i])
			}
			rlayData := &layPaths[rb.Index]
			i = pdIdx(minData.rSide, minData.cat)
			minData.rIdx = len(rlayData.paths[i])
			rlayData.paths[i] = append(rlayData.paths[i], minData)
			for _, pd := range rlayData.paths[i] {
				pd.rN = len(rlayData.paths[i])
			}
		}
	}
	// now we have the full set of data, sort positions
	// orderChanged := false
	for range 1 {
		for li := range nlay {
			ly := nv.Net.EmerLayer(li)
			lb := ly.AsEmer()
			layData := &layPaths[lb.Index]
			for side := range 4 {
				for cat := range 3 {
					pidx := pdIdx(side, cat)
					pths := layData.paths[pidx]
					npt := len(pths)
					if npt == 0 {
						continue
					}
					for _, pd := range pths {
						if pd.path.RecvLayer() != ly {
							continue
						}
						setPathPos(pd)
					}
					// slices.SortStableFunc(pths, func(a, b *pathData) int {
					// 	return -cmp.Compare(a.spos.Dim(sideDims[rSide]), b.spos.Dim(sideDims[rSide]))
					// })
				}
			}
		}
	}

	// final render
	for li := range nlay {
		ly := nv.Net.EmerLayer(li)
		lb := ly.AsEmer()
		layData := &layPaths[lb.Index]
		for side := range 4 {
			for cat := range 3 {
				pidx := pdIdx(side, cat)
				pths := layData.paths[pidx]
				for _, pd := range pths {
					if pd.path.RecvLayer() != ly {
						continue
					}
					pt := pd.path
					pb := pt.AsEmer()
					clr := colors.Spaced(pt.TypeNumber())
					xyz.NewArrow(se, pathsGp, pb.Name, pd.sPos, pd.rPos, lineWidth, clr, xyz.NoStartArrow, xyz.EndArrow, 4, .5, 4)
				}
			}
		}
		npt := len(layData.selfPaths)
		if npt == 0 {
			continue
		}
		// determine which side to put the self connections on.
		// they will show up in the front by default.
		var totLeft, totRight int
		for side := 1; side <= 2; side++ { // left, right
			for cat := range 3 {
				pidx := pdIdx(side, cat)
				if side == 1 {
					totLeft += len(layData.paths[pidx])
				} else {
					totRight += len(layData.paths[pidx])
				}
			}
		}
		selfSide := 1 // left
		if totRight < totLeft {
			selfSide = 2 // right
		}
		for pi, pd := range layData.selfPaths {
			pt := pd.path
			pb := pt.AsEmer()
			pd.sSide, pd.rSide = selfSide, selfSide
			clr := colors.Spaced(pt.TypeNumber())
			spm := nv.selfPrjn(se, pd.sSide)
			sfgp := xyz.NewGroup(pathsGp)
			sfgp.SetName(pb.Name)
			sfp := xyz.NewSolid(sfgp).SetMesh(spm).SetColor(clr)
			sfp.SetName(pb.Name)
			sfp.Pose.Pos = laySidePos(lb, selfSide, 1, pi, npt, 0.2)
		}
	}
}

func (nv *NetView) pathTypeNameMatch(ptyp string) bool {
	if len(nv.Options.PathType) == 0 {
		return true
	}
	ptyp = strings.ToLower(ptyp)
	fs := strings.Fields(nv.Options.PathType)
	for _, pt := range fs {
		pt = strings.ToLower(pt)
		if strings.Contains(ptyp, pt) {
			return true
		}
	}
	return false
}

// returns the self projection mesh, either left = 1 or right = 2
func (nv *NetView) selfPrjn(se *xyz.Scene, side int) xyz.Mesh {
	selfnm := fmt.Sprintf("selfPathSide%d", side)
	sm, err := se.MeshByName(selfnm)
	if err == nil {
		return sm
	}
	lineWidth := 1.5 * nv.Options.PathWidth
	size := float32(0.015)
	sideFact := float32(1.5)
	if side == 1 {
		sideFact = -1.5
	}
	sm = xyz.NewLines(se, selfnm, []math32.Vector3{{0, 0, -size}, {sideFact * size, 0, -size}, {sideFact * size, 0, size}, {0, 0, size}}, math32.Vec2(lineWidth, lineWidth), xyz.OpenLines)
	return sm
}
