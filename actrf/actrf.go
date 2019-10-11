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
// *target* patterns of activation -- i.e., sum (act * targ) / sum (targ)
// which then shows you the patterns of target activity for which
// a given unit was active.
// You must call Init to initialize everything, Reset to restart the accumulation of the data,
// and Avg to compute the resulting averages based an accumulated data.
// Avg does not erase the accumulated data so it can continue beyond that point.
type RF struct {
	Name    string          `desc:"name of this RF -- used for management of multiple in RFs"`
	RF      etensor.Float32 `view:"no-inline" desc:"computed receptive field, as SumProd / SumTarg -- only after Avg has been called"`
	NormRF  etensor.Float32 `view:"no-inline" desc:"unit normalized version of RF per target (inner 2D dimensions) -- good for display"`
	SumProd etensor.Float32 `view:"no-inline" desc:"sum of the products of act * targ"`
	SumTarg etensor.Float32 `view:"no-inline" desc:"sum of the targets (denomenator)"`
}

// Init initializes this RF based on name and shapes of given
// tensors representing the activations and target values.
func (af *RF) Init(name string, act, trg etensor.Tensor) {
	af.Name = name
	aNy, aNx, _, _ := etensor.Prjn2DShape(act, false)
	tNy, tNx, _, _ := etensor.Prjn2DShape(trg, false)
	oshp := []int{aNy, aNx, tNy, tNx}
	snm := []string{"ActY", "ActX", "TrgY", "TrgX"}
	af.RF.SetShape(oshp, nil, snm)
	af.NormRF.SetShape(oshp, nil, snm)
	af.SumProd.SetShape(oshp, nil, snm)
	af.SumTarg.SetShape(oshp, nil, snm)
	af.Reset()
}

// Reset reinitializes the Sum accumulators -- must have called Init first
func (af *RF) Reset() {
	af.SumProd.SetZeros()
	af.SumTarg.SetZeros()
}

// Add adds one sample based on activation and target tensor values.
// these must be of the same shape as used when Init was called.
// thr is a threshold value on targets below which values are not added (prevents
// numerical issues with very small numbers)
func (af *RF) Add(act, trg etensor.Tensor, thr float32) {
	aNy, aNx, _, _ := etensor.Prjn2DShape(act, false)
	tNy, tNx, _, _ := etensor.Prjn2DShape(trg, false)
	for ty := 0; ty < tNy; ty++ {
		for tx := 0; tx < tNx; tx++ {
			tv := float32(etensor.Prjn2DVal(trg, false, ty, tx))
			if tv < thr {
				continue
			}
			for ay := 0; ay < aNy; ay++ {
				for ax := 0; ax < aNx; ax++ {
					av := float32(etensor.Prjn2DVal(act, false, ay, ax))
					oi := []int{ay, ax, ty, tx}
					oo := af.SumProd.Offset(oi)
					af.SumProd.Values[oo] += av * tv
					af.SumTarg.Values[oo] += tv
				}
			}
		}
	}
}

// Avg computes RF as SumProd / SumTarg.  Does not Reset sums.
func (af *RF) Avg() {
	aNy := af.SumProd.Dim(0)
	aNx := af.SumProd.Dim(1)
	tNy := af.SumProd.Dim(2)
	tNx := af.SumProd.Dim(3)
	for ay := 0; ay < aNy; ay++ {
		for ax := 0; ax < aNx; ax++ {
			for ty := 0; ty < tNy; ty++ {
				for tx := 0; tx < tNx; tx++ {
					oi := []int{ay, ax, ty, tx}
					oo := af.SumProd.Offset(oi)
					trg := af.SumTarg.Values[oo]
					if trg > 0 {
						af.RF.Values[oo] = af.SumProd.Values[oo] / trg
					}

				}
			}
		}
	}
}

// Norm computes unit norm of RF values
func (af *RF) Norm() {
	af.NormRF.CopyFrom(&af.RF)
	norm.TensorUnit32(&af.NormRF, 2) // 2 = norm within outer 2 dims = norm each trg within
}
