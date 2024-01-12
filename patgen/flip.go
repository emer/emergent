// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package patgen

import (
	"github.com/emer/emergent/v2/erand"
	"github.com/emer/etable/v2/etensor"
)

// FlipBits turns nOff bits that are currently On to Off and
// nOn bits that are currently Off to On, using permuted lists.
func FlipBits(tsr etensor.Tensor, nOff, nOn int, onVal, offVal float64) {
	ln := tsr.Len()
	if ln == 0 {
		return
	}
	var ons, offs []int
	for i := 0; i < ln; i++ {
		vl := tsr.FloatVal1D(i)
		if vl == offVal {
			offs = append(offs, i)
		} else {
			ons = append(ons, i)
		}
	}
	erand.PermuteInts(ons, RandSource)
	erand.PermuteInts(offs, RandSource)
	if nOff > len(ons) {
		nOff = len(ons)
	}
	if nOn > len(offs) {
		nOn = len(offs)
	}
	for i := 0; i < nOff; i++ {
		tsr.SetFloat1D(ons[i], offVal)
	}
	for i := 0; i < nOn; i++ {
		tsr.SetFloat1D(offs[i], onVal)
	}
}

// FlipBitsRows turns nOff bits that are currently On to Off and
// nOn bits that are currently Off to On, using permuted lists.
// Iterates over the outer-most tensor dimension as rows.
func FlipBitsRows(tsr etensor.Tensor, nOff, nOn int, onVal, offVal float64) {
	rows, _ := tsr.RowCellSize()
	for i := 0; i < rows; i++ {
		trow := tsr.SubSpace([]int{i})
		FlipBits(trow, nOff, nOn, onVal, offVal)
	}
}
