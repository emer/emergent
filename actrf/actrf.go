// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package actrf

import (
	"github.com/emer/etable/etensor"
	"github.com/emer/etable/norm"
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
	Name    string          `desc:"name of this RF -- used for management of multiple in RFs"`
	RF      etensor.Float32 `view:"no-inline" desc:"computed receptive field, as SumProd / SumSrc -- only after Avg has been called"`
	NormRF  etensor.Float32 `view:"no-inline" desc:"unit normalized version of RF per source (inner 2D dimensions) -- good for display"`
	SumProd etensor.Float32 `view:"no-inline" desc:"sum of the products of act * src"`
	SumSrc  etensor.Float32 `view:"no-inline" desc:"sum of the sources (denomenator)"`
}

// Init initializes this RF based on name and shapes of given
// tensors representing the activations and source values.
func (af *RF) Init(name string, act, src etensor.Tensor) {
	af.Name = name
	aNy, aNx, _, _ := etensor.Prjn2DShape(act.ShapeObj(), false)
	sNy, sNx, _, _ := etensor.Prjn2DShape(src.ShapeObj(), false)
	oshp := []int{aNy, aNx, sNy, sNx}
	snm := []string{"ActY", "ActX", "SrcY", "SrcX"}
	af.RF.SetShape(oshp, nil, snm)
	af.NormRF.SetShape(oshp, nil, snm)
	af.SumProd.SetShape(oshp, nil, snm)
	af.SumSrc.SetShape(oshp, nil, snm)
	af.Reset()
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
func (af *RF) Add(act, src etensor.Tensor, thr float32) {
	aNy, aNx, _, _ := etensor.Prjn2DShape(act.ShapeObj(), false)
	sNy, sNx, _, _ := etensor.Prjn2DShape(src.ShapeObj(), false)
	for sy := 0; sy < sNy; sy++ {
		for sx := 0; sx < sNx; sx++ {
			tv := float32(etensor.Prjn2DVal(src, false, sy, sx))
			if tv < thr {
				continue
			}
			for ay := 0; ay < aNy; ay++ {
				for ax := 0; ax < aNx; ax++ {
					av := float32(etensor.Prjn2DVal(act, false, ay, ax))
					oi := []int{ay, ax, sy, sx}
					oo := af.SumProd.Offset(oi)
					af.SumProd.Values[oo] += av * tv
					af.SumSrc.Values[oo] += tv
				}
			}
		}
	}
}

// Avg computes RF as SumProd / SumSrc.  Does not Reset sums.
func (af *RF) Avg() {
	aNy := af.SumProd.Dim(0)
	aNx := af.SumProd.Dim(1)
	sNy := af.SumProd.Dim(2)
	sNx := af.SumProd.Dim(3)
	for ay := 0; ay < aNy; ay++ {
		for ax := 0; ax < aNx; ax++ {
			for sy := 0; sy < sNy; sy++ {
				for sx := 0; sx < sNx; sx++ {
					oi := []int{ay, ax, sy, sx}
					oo := af.SumProd.Offset(oi)
					src := af.SumSrc.Values[oo]
					if src > 0 {
						af.RF.Values[oo] = af.SumProd.Values[oo] / src
					}

				}
			}
		}
	}
}

// Norm computes unit norm of RF values
func (af *RF) Norm() {
	af.NormRF.CopyFrom(&af.RF)
	norm.TensorUnit32(&af.NormRF, 2) // 2 = norm within outer 2 dims = norm each src within
}
