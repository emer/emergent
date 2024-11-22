// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package patgen

import (
	"cogentcore.org/core/base/randx"
	"cogentcore.org/core/tensor/table"
)

// Shuffle shuffles rows in specified columns in the table independently
func Shuffle(dt *table.Table, rows []int, colNames []string, colIndependent bool) {
	cl := dt.Clone()
	if colIndependent { // independent
		for _, colNm := range colNames {
			sfrows := make([]int, len(rows))
			copy(sfrows, rows)
			randx.PermuteInts(sfrows, RandSource)
			for i, row := range rows {
				dt.Column(colNm).RowTensor(row).CopyFrom(cl.Column(colNm).RowTensor(sfrows[i]))
			}
		}
	} else { // shuffle together
		sfrows := make([]int, len(rows))
		copy(sfrows, rows)
		randx.PermuteInts(sfrows, RandSource)
		for _, colNm := range colNames {
			for i, row := range rows {
				dt.Column(colNm).RowTensor(row).CopyFrom(cl.Column(colNm).RowTensor(sfrows[i]))
			}
		}
	}
}
