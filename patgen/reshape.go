// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package patgen

import (
	"log"
	"reflect"

	"cogentcore.org/core/core"
	"github.com/emer/etensor/tensor/table"
)

// ReshapeCpp fixes C++ emergent table shape which is reversed from Go.
// it switches the dimension order in the given table, for all columns
// that are float 2D or 4D columns -- assumes these are layer patterns
// and names dimensions accordingly.
func ReshapeCpp(dt *table.Table) {
	for _, cl := range dt.Columns {
		shp := cl.Shape().Sizes
		if cl.NumDims() == 3 && (cl.DataType() == reflect.Float32 || cl.DataType() == reflect.Float64) {
			revshp := []int{shp[0], shp[2], shp[1]} // [0] = row
			dnms := []string{"Row", "Y", "X"}
			cl.SetShape(revshp, dnms...)
		}
		if cl.NumDims() == 5 && (cl.DataType() == reflect.Float32 || cl.DataType() == reflect.Float64) {
			revshp := []int{shp[0], shp[4], shp[3], shp[2], shp[1]} // [0] = row
			dnms := []string{"Row", "PoolY", "PoolX", "NeurY", "NeurX"}
			cl.SetShape(revshp, dnms...)
		}
	}
}

// ReshapeCppFile fixes C++ emergent table shape which is reversed from Go.
// It loads file from fname and saves to fixnm
func ReshapeCppFile(dt *table.Table, fname, fixnm string) {
	err := dt.OpenCSV(core.Filename(fname), table.Tab)
	if err != nil {
		log.Println(err)
		return
	}
	ReshapeCpp(dt)
	dt.SaveCSV(core.Filename(fixnm), table.Tab, true)
}
