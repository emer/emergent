// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package patgen

import (
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
)

// ConfigVocab configures enough trials of vocabularies for ConfigPats
// There's always a "void" column where the patterns are zeros
// Index columns by pool names instead of numbers
func ConfigVocab(dt *etable.Table, poolY, poolX int, ecPctAct float32, poolNames []string) {
	rows := 1000
	nOn := int(float32(poolY) * float32(poolX) * ecPctAct)
	dt.SetFromSchema(etable.Schema{
		{"void", etensor.FLOAT32, []int{poolY, poolX}, []string{"Y", "X"}},
	}, rows)
	for i, name := range poolNames {
		dt.AddCol(dt.Cols[0].Clone(), name)
		PermutedBinaryRows(dt.Cols[i+1], nOn, 1, 0)
	}
}
