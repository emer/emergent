// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package patgen

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/emer/etable/etensor"
)

// AddVocabVoid adds a void pool to the vocabulary
func AddVocabVoid(mp map[string]*etensor.Float32, rows, poolY, poolX int, poolName string) {
	tsr := etensor.NewFloat32([]int{rows, poolY, poolX}, nil, []string{"row", "Y", "X"})
	mp[poolName] = tsr
}

// AddVocab adds a normal pool to the vocabulary
func AddVocab(mp map[string]*etensor.Float32, rows, poolY, poolX int, pctAct float32, poolName string) {
	nOn := int(math.Round(float64(poolY) * float64(poolX) * float64(pctAct)))
	tsr := etensor.NewFloat32([]int{rows, poolY, poolX}, nil, []string{"row", "Y", "X"})
	PermutedBinaryRows(tsr, nOn, 1, 0)
	mp[poolName] = tsr
}

// AddVocabRepeat adds a repeated pool to the vocabulary
func AddVocabRepeat(mp map[string]*etensor.Float32, rows, poolY, poolX int, pctAct float32, poolName string) {
	nOn := int(math.Round(float64(poolY) * float64(poolX) * float64(pctAct)))
	tsr := etensor.NewFloat32([]int{rows, poolY, poolX}, nil, []string{"row", "Y", "X"})
	for i := 0; i < rows; i++ {
		subtsr := tsr.SubSpace([]int{i})
		if i == 0 {
			PermutedBinary(subtsr, nOn, 1, 0)
		} else {
			subtsr.CopyFrom(tsr.SubSpace([]int{i - 1}))
		}
	}
	mp[poolName] = tsr
}

// AddVocabDrift adds a temporally drifted pool to the vocabulary
func AddVocabDrift(mp map[string]*etensor.Float32, rows, poolY, poolX int, pctAct, pctDrift float32, poolName string) {
	nOn := int(math.Round(float64(poolY) * float64(poolX) * float64(pctAct)))
	nDrift := int(math.Round(float64(nOn) * float64(pctDrift)))
	AddVocab(mp, rows, poolY, poolX, pctAct, poolName)
	VocabDrift(mp, nDrift, poolName)
}

// VocabDrift makes a pool in the vocabulary temporally drifted
func VocabDrift(mp map[string]*etensor.Float32, nDrift int, poolName string) {
	tsr := mp[poolName]
	rows := tsr.Shapes()[0]
	for i := 0; i < rows; i++ {
		subtsr := tsr.SubSpace([]int{i})
		if i > 0 {
			subtsr.CopyFrom(tsr.SubSpace([]int{i - 1}))
			FlipBits(subtsr, nDrift, nDrift, 1, 0)
		}
	}
}

// VocabClone clones an old pool in the vocabulary and make it a new one
func VocabClone(mp map[string]*etensor.Float32, frmPool, toNewPool string) {
	mp[toNewPool] = mp[frmPool].Clone().(*etensor.Float32)
}

// VocabShuffle shuffles a pool in the vocabulary on its first dimension (row)
func VocabShuffle(mp map[string]*etensor.Float32, shuffleKeys []string) {
	for _, shuffleKey := range shuffleKeys {
		tsr := mp[shuffleKey]
		rows := tsr.Shapes()[0]
		poolY := tsr.Shapes()[1]
		poolX := tsr.Shapes()[2]
		sRows := rand.Perm(rows)
		sTsr := etensor.NewFloat32([]int{rows, poolY, poolX}, nil, []string{"row", "Y", "X"})
		for iRow, sRow := range sRows {
			sTsr.SubSpace([]int{iRow}).CopyFrom(tsr.SubSpace([]int{sRow}))
		}
		mp[shuffleKey] = sTsr
	}
}

// VocabConcat contatenates several pools in the vocabulary and make it a new one
func VocabConcat(mp map[string]*etensor.Float32, newKey string, frmKeys []string) {
	tsr := mp[frmKeys[0]].Clone().(*etensor.Float32)
	for i, key := range frmKeys {
		// check dimension, how
		if i > 0 {
			currows := tsr.Shapes()[0]
			approws := mp[key].Shapes()[0]
			tsr.SetShape([]int{currows + approws, tsr.Shapes()[1], tsr.Shapes()[2]}, nil, []string{"row", "Y", "X"})
			for iRow := 0; iRow < approws; iRow++ {
				subtsr := tsr.SubSpace([]int{iRow + currows})
				subtsr.CopyFrom(mp[key].SubSpace([]int{iRow}))
			}
		}
	}
	mp[newKey] = tsr
}

// VocabSlice slices a pool in the vocabulary into new ones
func VocabSlice(mp map[string]*etensor.Float32, frmKey string, newKeys []string, sliceOffs []int) error {
	oriTsr := mp[frmKey]
	poolY := oriTsr.Shapes()[1]
	poolX := oriTsr.Shapes()[2]

	// check newKeys and sliceOffs have same length
	if len(newKeys)+1 != len(sliceOffs) {
		err := fmt.Errorf("sliceOffs should have one more element than newKeys") // how do I stop the program?
		fmt.Println(err.Error())
	}

	// check sliceOffs is in right order
	preVal := sliceOffs[0]
	for i, curVal := range sliceOffs {
		if i > 0 {
			if preVal < curVal {
				preVal = curVal
			} else {
				err := fmt.Errorf("sliceOffs should increase progressively") // how do I stop the program?
				fmt.Println(err.Error())
			}
		}
	}

	// slice
	frmOff := sliceOffs[0]
	for i := range newKeys {
		toOff := sliceOffs[i+1]
		newKey := newKeys[i]
		newTsr := etensor.NewFloat32([]int{toOff - frmOff, poolY, poolX}, nil, []string{"row", "Y", "X"})
		for off := frmOff; off < toOff; off++ {
			newTsr.SubSpace([]int{off - frmOff}).CopyFrom(oriTsr.SubSpace([]int{off}))
		}
		mp[newKey] = newTsr
		frmOff = toOff
	}
	return nil
}
