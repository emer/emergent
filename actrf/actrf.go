// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package actrf

//go:generate core generate -add-types

import (
	"slices"

	"cogentcore.org/core/tensor"
	"cogentcore.org/core/tensor/stats/stats"
)

// RF is used for computing an activation-based receptive field.
// It simply computes the activation weighted average of other
// *source* patterns of activation -- i.e., sum(act * src) / sum(src)
// which then shows you the patterns of source activity for which
// a given unit was active.
// You must call Init to initialize everything, Reset to restart the accumulation of the data,
// and Avg to compute the resulting averages based an accumulated data.
// Avg does not erase the accumulated data so it can continue beyond that point.
type RF struct {

	// name of this RF -- used for management of multiple in RFs
	Name string

	// computed receptive field, as SumProd / SumSrc -- only after Avg has been called
	RF tensor.Float32 `display:"no-inline"`

	// unit normalized version of RF per source (inner 2D dimensions) -- good for display
	NormRF tensor.Float32 `display:"no-inline"`

	// normalized version of SumSrc -- sum of each point in the source -- good for viewing the completeness and uniformity of the sampling of the source space
	NormSrc tensor.Float32 `display:"no-inline"`

	// sum of the products of act * src
	SumProd tensor.Float32 `display:"no-inline"`

	// sum of the sources (denomenator)
	SumSrc tensor.Float32 `display:"no-inline"`

	// temporary destination sum for MPI -- only used when MPISum called
	MPITmp tensor.Float32 `display:"no-inline"`
}

// Init initializes this RF based on name and shapes of given
// tensors representing the activations and source values.
func (af *RF) Init(name string, act, src tensor.Tensor) {
	af.Name = name
	af.InitShape(act, src)
	af.Reset()
}

// InitShape initializes shape for this RF based on shapes of given
// tensors representing the activations and source values.
// does nothing if shape is already correct.
// return shape ints
func (af *RF) InitShape(act, src tensor.Tensor) []int {
	aNy, aNx, _, _ := tensor.Projection2DShape(act.Shape(), false)
	sNy, sNx, _, _ := tensor.Projection2DShape(src.Shape(), false)
	oshp := []int{aNy, aNx, sNy, sNx}
	if slices.Equal(af.RF.Shape().Sizes, oshp) {
		return oshp
	}
	sshp := []int{sNy, sNx}
	af.RF.SetShapeSizes(oshp...)
	af.NormRF.SetShapeSizes(oshp...)
	af.SumProd.SetShapeSizes(oshp...)
	af.NormSrc.SetShapeSizes(sshp...)
	af.SumSrc.SetShapeSizes(sshp...)

	af.ConfigView(&af.RF)
	af.ConfigView(&af.NormRF)
	af.ConfigView(&af.SumProd)
	af.ConfigView(&af.NormSrc)
	af.ConfigView(&af.SumSrc)
	return oshp
}

// ConfigView configures the view params on the tensor
func (af *RF) ConfigView(tsr *tensor.Float32) {
	// todo:meta
	// tsr.SetMetaData("colormap", "Viridis")
	// tsr.SetMetaData("grid-fill", "1") // remove extra lines
	// tsr.SetMetaData("fix-min", "true")
	// tsr.SetMetaData("min", "0")
}

// Reset reinitializes the Sum accumulators -- must have called Init first
func (af *RF) Reset() {
	af.SumProd.SetZeros()
	af.SumSrc.SetZeros()
}

// Add adds one sample based on activation and source tensor values.
// these must be of the same shape as used when Init was called.
// thr is a threshold value on sources below which values are not added (prevents
// numerical issues with very small numbers)
func (af *RF) Add(act, src tensor.Tensor, thr float32) {
	shp := af.InitShape(act, src) // ensure
	aNy, aNx, sNy, sNx := shp[0], shp[1], shp[2], shp[3]
	for sy := 0; sy < sNy; sy++ {
		for sx := 0; sx < sNx; sx++ {
			tv := float32(tensor.Projection2DValue(src, false, sy, sx))
			if tv < thr {
				continue
			}
			af.SumSrc.SetAdd(tv, sy, sx)
			for ay := 0; ay < aNy; ay++ {
				for ax := 0; ax < aNx; ax++ {
					av := float32(tensor.Projection2DValue(act, false, ay, ax))
					af.SumProd.SetAdd(av*tv, ay, ax, sy, sx)
				}
			}
		}
	}
}

// Avg computes RF as SumProd / SumSrc.  Does not Reset sums.
func (af *RF) Avg() {
	aNy := af.SumProd.DimSize(0)
	aNx := af.SumProd.DimSize(1)
	sNy := af.SumProd.DimSize(2)
	sNx := af.SumProd.DimSize(3)
	var maxSrc float32
	for sy := 0; sy < sNy; sy++ {
		for sx := 0; sx < sNx; sx++ {
			src := af.SumSrc.Value(sy, sx)
			if src == 0 {
				continue
			}
			if src > maxSrc {
				maxSrc = src
			}
			for ay := 0; ay < aNy; ay++ {
				for ax := 0; ax < aNx; ax++ {
					oo := af.SumProd.Shape().IndexTo1D(ay, ax, sy, sx)
					af.RF.Values[oo] = af.SumProd.Values[oo] / src
				}
			}
		}
	}
	if maxSrc == 0 {
		maxSrc = 1
	}
	for i, v := range af.SumSrc.Values {
		af.NormSrc.Values[i] = v / maxSrc
	}
}

// Norm computes unit norm of RF values -- must be called after Avg
func (af *RF) Norm() {
	stats.UnitNormOut(&af.RF, &af.NormRF)
}

// AvgNorm computes RF as SumProd / SumTarg and then does Norm.
// This is what you typically want to call before viewing RFs.
// Does not Reset sums.
func (af *RF) AvgNorm() {
	af.Avg()
	af.Norm()
}
