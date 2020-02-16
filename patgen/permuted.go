// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package patgen

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"

	"github.com/chewxy/math32"
	"github.com/emer/emergent/erand"
	"github.com/emer/etable/etensor"
	"github.com/emer/etable/metric"
)

// PermutedBinary sets the given tensor to contain nOn onVal values and the
// remainder are offVal values, using a permuted order of tensor elements (i.e.,
// randomly shuffled or permuted).
func PermutedBinary(tsr etensor.Tensor, nOn int, onVal, offVal float64) {
	ln := tsr.Len()
	if ln == 0 {
		return
	}
	pord := rand.Perm(ln)
	for i := 0; i < ln; i++ {
		if i < nOn {
			tsr.SetFloat1D(pord[i], onVal)
		} else {
			tsr.SetFloat1D(pord[i], offVal)
		}
	}
}

// PermutedBinaryRows treats the tensor as a column of rows as in a etable.Table
// and sets each row to contain nOn onVal values and the remainder are offVal values,
// using a permuted order of tensor elements (i.e., randomly shuffled or permuted).
func PermutedBinaryRows(tsr etensor.Tensor, nOn int, onVal, offVal float64) {
	rows, cells := tsr.RowCellSize()
	if rows == 0 || cells == 0 {
		return
	}
	pord := rand.Perm(cells)
	for rw := 0; rw < rows; rw++ {
		stidx := rw * cells
		for i := 0; i < cells; i++ {
			if i < nOn {
				tsr.SetFloat1D(stidx+pord[i], onVal)
			} else {
				tsr.SetFloat1D(stidx+pord[i], offVal)
			}
		}
		erand.PermuteInts(pord)
	}
}

// PermutedBinaryMinDiff treats the tensor as a column of rows as in a etable.Table
// and sets each row to contain nOn onVal values and the remainder are offVal values,
// using a permuted order of tensor elements (i.e., randomly shuffled or permuted).
// This version ensures that all patterns have at least a given minimum distance from each other,
// expressed using minDiff = number of bits that must be different (can't be > nOn).
// If the mindiff constraint cannot be met within a reasonable number of iterations,
// then an error is returned.
func PermutedBinaryMinDiff(tsr *etensor.Float32, nOn int, onVal, offVal float32, minDiff int) error {
	rows, cells := tsr.RowCellSize()
	if rows == 0 || cells == 0 {
		return errors.New("empty tensor")
	}
	pord := rand.Perm(cells)
	fails := 0
	for rw := 0; rw < rows; rw++ {
		stidx := rw * cells
		iters := 100 + (10 * (rw + 1)) // 100 plus 10 more for every new rew
		got := false
		for itr := 0; itr < iters; itr++ {
			for i := 0; i < cells; i++ {
				if i < nOn {
					tsr.Values[stidx+pord[i]] = onVal
				} else {
					tsr.Values[stidx+pord[i]] = offVal
				}
			}
			erand.PermuteInts(pord)
			if rw == 0 {
				got = true
				break
			}
			min, _ := RowVsPrevDist32(tsr, rw, metric.Hamming32)
			df := int(math.Round(float64(.5 * min))) // diff
			if df >= minDiff {
				got = true
				break
			}
		}
		if !got {
			fails++
		}
	}
	if fails > 0 {
		err := fmt.Errorf("PermutedBinaryMinDiff: minimum difference of: %d was not met: %d times, rows: %d", minDiff, fails, rows)
		log.Println(err)
		return err
	}
	return nil
}

// RowVsPrevDist32 returns the minimum and maximum distance between the given row
// in tensor and all previous rows.  Row must be >= 1 and < total rows.
// (outer-most dimension is row, as in columns of etable.Table).
func RowVsPrevDist32(tsr *etensor.Float32, row int, fun metric.Func32) (min, max float32) {
	if row < 1 {
		return
	}
	min = float32(math.MaxFloat32)
	max = float32(-math.MaxFloat32)
	lrow := tsr.SubSpace([]int{row}).(*etensor.Float32)
	for i := 0; i <= row-1; i++ {
		crow := tsr.SubSpace([]int{i}).(*etensor.Float32)
		dst := fun(lrow.Values, crow.Values)
		min = math32.Min(min, dst)
		max = math32.Max(max, dst)
	}
	return
}
