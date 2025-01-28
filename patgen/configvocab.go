// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package patgen

//go:generate core generate -add-types

import (
	"fmt"
	"log"
	"math"

	"cogentcore.org/core/base/errors"
	"github.com/emer/etensor/tensor"
	"github.com/emer/etensor/tensor/stats/stats"
)

// Vocab is a map of named tensors that contain patterns used for creating
// larger patterns by mixing together.
type Vocab map[string]*tensor.Float32

// ByName looks for vocabulary item of given name, and returns
// (and logs) error message if not found.
func (vc Vocab) ByName(name string) (*tensor.Float32, error) {
	tsr, ok := vc[name]
	if !ok {
		return nil, errors.Log(fmt.Errorf("Vocabulary item named: %s not found", name))
	}
	return tsr, nil
}

// Note: to keep things consistent, all AddVocab functions start with Vocab and name
// args and return the tensor and an error, even if there is no way that they could error.
// Also, all routines should automatically log any error message, and because this is
// "end user" code, it is much better to have error messages instead of crashes
// so we add the extra checks etc.

// NOnInTensor returns the number of bits active in given tensor
func NOnInTensor(trow *tensor.Float32) int {
	return int(stats.SumTensor(trow))
}

// PctActInTensor returns the percent activity in given tensor (NOn / size)
func PctActInTensor(trow *tensor.Float32) float32 {
	return float32(NOnInTensor(trow)) / float32(trow.Len())
}

// NFromPct returns the number of bits for given pct (proportion 0-1),
// relative to total n.  uses math.Round.
func NFromPct(pct float32, n int) int {
	return int(math.Round(float64(n) * float64(pct)))
}

// AddVocabEmpty adds an empty pool to the vocabulary.
// This can be used to make test cases with missing pools.
func AddVocabEmpty(mp Vocab, name string, rows, poolY, poolX int) (*tensor.Float32, error) {
	tsr := tensor.NewFloat32([]int{rows, poolY, poolX}, "row", "Y", "X")
	mp[name] = tsr
	return tsr, nil
}

// AddVocabPermutedBinary adds a permuted binary pool to the vocabulary.
// This is a good source of random patterns with no systematic similarity.
// pctAct = proportion (0-1) bits turned on for a pool.
// minPctDiff = proportion of pctAct (0-1) for minimum difference between
// patterns -- e.g., .5 = each pattern must have half of its bits different
// from each other.  This constraint can be hard to meet depending on the
// number of rows, amount of activity, and minPctDiff level -- an error
// will be returned and printed if it cannot be satisfied in a reasonable
// amount of time.
func AddVocabPermutedBinary(mp Vocab, name string, rows, poolY, poolX int, pctAct, minPctDiff float32) (*tensor.Float32, error) {
	nOn := NFromPct(pctAct, poolY*poolX)
	minDiff := NFromPct(minPctDiff, nOn)
	tsr := tensor.NewFloat32([]int{rows, poolY, poolX}, "row", "Y", "X")
	err := PermutedBinaryMinDiff(tsr, nOn, 1, 0, minDiff)
	mp[name] = tsr
	return tsr, err
}

// AddVocabClone clones an existing pool in the vocabulary to make a new one.
func AddVocabClone(mp Vocab, name string, copyFrom string) (*tensor.Float32, error) {
	cp, err := mp.ByName(copyFrom)
	if err != nil {
		return nil, err
	}
	tsr := cp.Clone().(*tensor.Float32)
	mp[name] = tsr
	return tsr, nil
}

// AddVocabRepeat adds a repeated pool to the vocabulary,
// copying from given row in existing vocabulary item .
func AddVocabRepeat(mp Vocab, name string, rows int, copyFrom string, copyRow int) (*tensor.Float32, error) {
	origItem, err := mp.ByName(copyFrom)
	if err != nil {
		return nil, err
	}
	cp := origItem.Clone()
	tsr := &tensor.Float32{}
	cpshp := cp.Shape().Sizes
	cpshp[0] = rows
	tsr.SetShape(cpshp, cp.Shape().Names...)
	mp[name] = tsr
	cprow := cp.SubSpace([]int{copyRow})
	for i := 0; i < rows; i++ {
		trow := tsr.SubSpace([]int{i})
		trow.CopyFrom(cprow)
	}
	return tsr, nil
}

// AddVocabDrift adds a row-by-row drifting pool to the vocabulary,
// starting from the given row in existing vocabulary item
// (which becomes starting row in this one -- drift starts in second row).
// The current row patterns are generated by taking the previous row
// pattern and flipping pctDrift percent of active bits (min of 1 bit).
func AddVocabDrift(mp Vocab, name string, rows int, pctDrift float32, copyFrom string, copyRow int) (*tensor.Float32, error) {
	cp, err := mp.ByName(copyFrom)
	if err != nil {
		return nil, err
	}
	tsr := &tensor.Float32{}
	cpshp := cp.Shape().Sizes
	cpshp[0] = rows
	tsr.SetShape(cpshp, cp.Shape().Names...)
	mp[name] = tsr
	cprow := cp.SubSpace([]int{copyRow}).(*tensor.Float32)
	trow := tsr.SubSpace([]int{0})
	trow.CopyFrom(cprow)
	nOn := NOnInTensor(cprow)
	rmdr := 0.0                               // remainder carryover in drift
	drift := float64(nOn) * float64(pctDrift) // precise fractional amount of drift
	for i := 1; i < rows; i++ {
		srow := tsr.SubSpace([]int{i - 1})
		trow := tsr.SubSpace([]int{i})
		trow.CopyFrom(srow)
		curDrift := math.Round(drift + rmdr) // integer amount
		nDrift := int(curDrift)
		if nDrift > 0 {
			FlipBits(trow, nDrift, nDrift, 1, 0)
		}
		rmdr += drift - curDrift // accumulate remainder
	}
	return tsr, nil
}

// VocabShuffle shuffles a pool in the vocabulary on its first dimension (row).
func VocabShuffle(mp Vocab, shufflePools []string) {
	for _, key := range shufflePools {
		tsr := mp[key]
		rows := tsr.Shape().Sizes[0]
		poolY := tsr.Shape().Sizes[1]
		poolX := tsr.Shape().Sizes[2]
		sRows := RandSource.Perm(rows)
		sTsr := tensor.NewFloat32([]int{rows, poolY, poolX}, "row", "Y", "X")
		for iRow, sRow := range sRows {
			sTsr.SubSpace([]int{iRow}).CopyFrom(tsr.SubSpace([]int{sRow}))
		}
		mp[key] = sTsr
	}
}

// VocabConcat contatenates several pools in the vocabulary and store it into newPool (could be one of the previous pools).
func VocabConcat(mp Vocab, newPool string, frmPools []string) error {
	tsr := mp[frmPools[0]].Clone().(*tensor.Float32)
	for i, key := range frmPools {
		if i > 0 {
			// check pool shape
			if !(tsr.SubSpace([]int{0}).(*tensor.Float32).Shp.IsEqual(&mp[key].SubSpace([]int{0}).(*tensor.Float32).Shp)) {
				err := fmt.Errorf("shapes of input pools must be the same") // how do I stop the program?
				log.Println(err.Error())
				return err
			}

			currows := tsr.Shape().Sizes[0]
			approws := mp[key].Shape().Sizes[0]
			tsr.SetShape([]int{currows + approws, tsr.Shape().Sizes[1], tsr.Shape().Sizes[2]}, "row", "Y", "X")
			for iRow := 0; iRow < approws; iRow++ {
				subtsr := tsr.SubSpace([]int{iRow + currows})
				subtsr.CopyFrom(mp[key].SubSpace([]int{iRow}))
			}
		}
	}
	mp[newPool] = tsr
	return nil
}

// VocabSlice slices a pool in the vocabulary into new ones.
// SliceOffs is the cutoff points in the original pool, should have one more element than newPools.
func VocabSlice(mp Vocab, frmPool string, newPools []string, sliceOffs []int) error {
	oriTsr := mp[frmPool]
	poolY := oriTsr.Shape().Sizes[1]
	poolX := oriTsr.Shape().Sizes[2]

	// check newPools and sliceOffs have same length
	if len(newPools)+1 != len(sliceOffs) {
		err := fmt.Errorf("sliceOffs should have one more element than newPools") // how do I stop the program?
		log.Println(err.Error())
		return err
	}

	// check sliceOffs is in right order
	preVal := sliceOffs[0]
	for i, curVal := range sliceOffs {
		if i > 0 {
			if preVal < curVal {
				preVal = curVal
			} else {
				err := fmt.Errorf("sliceOffs should increase progressively") // how do I stop the program?
				log.Println(err.Error())
				return err
			}
		}
	}

	// slice
	frmOff := sliceOffs[0]
	for i := range newPools {
		toOff := sliceOffs[i+1]
		newPool := newPools[i]
		newTsr := tensor.NewFloat32([]int{toOff - frmOff, poolY, poolX}, "row", "Y", "X")
		for off := frmOff; off < toOff; off++ {
			newTsr.SubSpace([]int{off - frmOff}).CopyFrom(oriTsr.SubSpace([]int{off}))
		}
		mp[newPool] = newTsr
		frmOff = toOff
	}
	return nil
}
