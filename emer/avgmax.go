// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

import "math"

// AvgMax holds average and max statistics
type AvgMax struct {
	Avg    float32
	Max    float32
	MaxIdx int     `desc:"index of max item"`
	Sum    float32 `desc:"sum for computing average"`
	N      int     `desc:"number of items in sum"`
}

// Init initializes prior to new updates
func (am *AvgMax) Init() {
	am.Avg = 0
	am.Sum = 0
	am.N = 0
	am.Max = -math.MaxFloat32
	am.MaxIdx = -1
}

// UpdateVal updates stats from given value
func (am *AvgMax) UpdateVal(val float32, idx int) {
	am.Sum += val
	am.N++
	if val > am.Max {
		am.Max = val
		am.MaxIdx = idx
	}
}

// UpdateFrom updates these values from other AvgMax
func (am *AvgMax) UpdateFrom(oth *AvgMax) {
	am.Sum += oth.Sum
	am.N += oth.N
	if oth.Max > am.Max {
		am.Max = oth.Max
		am.MaxIdx = oth.MaxIdx
	}
}

// CopyFrom copies from other AvgMax
func (am *AvgMax) CopyFrom(oth *AvgMax) {
	am.Avg = oth.Avg
	am.Max = oth.Max
	am.MaxIdx = oth.MaxIdx
	am.Sum = oth.Sum
	am.N = oth.N
}

// CalcAvg computes the average given the current Sum and N values
func (am *AvgMax) CalcAvg() {
	if am.N > 0 {
		am.Avg = am.Sum / float32(am.N)
	} else {
		am.Avg = am.Sum
		am.Max = am.Avg // prevents Max from being -MaxFloat..
	}
}
