// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package patgen

import (
	"errors"
	"fmt"
	"log"
	"math"

	"github.com/emer/emergent/erand"
	"github.com/emer/etable/etensor"
	"github.com/emer/etable/metric"
	"github.com/goki/ki/ints"
	"github.com/goki/mat32"
)

// PermutedBinary sets the given tensor to contain nOn onVal values and the
// remainder are offVal values, using a permuted order of tensor elements (i.e.,
// randomly shuffled or permuted).
func PermutedBinary(tsr etensor.Tensor, nOn int, onVal, offVal float64) {
	ln := tsr.Len()
	if ln == 0 {
		return
	}
	pord := RandSource.Perm(ln, -1)
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
	pord := RandSource.Perm(cells, -1)
	for rw := 0; rw < rows; rw++ {
		stidx := rw * cells
		for i := 0; i < cells; i++ {
			if i < nOn {
				tsr.SetFloat1D(stidx+pord[i], onVal)
			} else {
				tsr.SetFloat1D(stidx+pord[i], offVal)
			}
		}
		erand.PermuteInts(pord, RandSource)
	}
}

// MinDiffPrintIters set this to true to see the iteration stats for
// PermutedBinaryMinDiff -- for large, long-running cases.
var MinDiffPrintIters = false

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
	pord := RandSource.Perm(cells, -1)
	iters := 100
	nunder := make([]int, rows) // per row
	fails := 0
	for itr := 0; itr < iters; itr++ {
		for rw := 0; rw < rows; rw++ {
			if itr > 0 && nunder[rw] == 0 {
				continue
			}
			stidx := rw * cells
			for i := 0; i < cells; i++ {
				if i < nOn {
					tsr.Values[stidx+pord[i]] = onVal
				} else {
					tsr.Values[stidx+pord[i]] = offVal
				}
			}
			erand.PermuteInts(pord, RandSource)
		}
		for i := range nunder {
			nunder[i] = 0
		}
		nbad := 0
		mxnun := 0
		for r1 := 0; r1 < rows; r1++ {
			r1v := tsr.SubSpace([]int{r1}).(*etensor.Float32)
			for r2 := r1 + 1; r2 < rows; r2++ {
				r2v := tsr.SubSpace([]int{r2}).(*etensor.Float32)
				dst := metric.Hamming32(r1v.Values, r2v.Values)
				df := int(math.Round(float64(.5 * dst)))
				if df < minDiff {
					nunder[r1]++
					mxnun = ints.MaxInt(mxnun, nunder[r1])
					nunder[r2]++
					mxnun = ints.MaxInt(mxnun, nunder[r2])
					nbad++
				}
			}
		}
		if nbad == 0 {
			break
		}
		fails++
		if MinDiffPrintIters {
			fmt.Printf("PermutedBinaryMinDiff: Itr: %d  NBad: %d  MaxN: %d\n", itr, nbad, mxnun)
		}
	}
	if fails == iters {
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
		min = mat32.Min(min, dst)
		max = mat32.Max(max, dst)
	}
	return
}
