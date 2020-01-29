// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package patgen

import (
	"fmt"

	"github.com/emer/emergent/erand"
	"github.com/emer/etable/etable"
)

// Shuffle shuffles rows in specified columns in the table independently
func Shuffle(dt *etable.Table, rows []int, colNames []string) {
	cl := dt.Clone()
	for _, colNm := range colNames {
		orirows := make([]int, len(rows))
		copy(orirows, rows)
		erand.PermuteInts(rows)
		fmt.Println(orirows)
		fmt.Println(rows)
		for i, row := range rows {
			dt.CellTensor(colNm, row).CopyFrom(cl.CellTensor(colNm, orirows[i]))
		}
	}
}
