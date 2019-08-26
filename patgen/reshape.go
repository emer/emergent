// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package patgen

import (
	"log"

	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/goki/gi/gi"
)

// ReshapeCpp fixes C++ emergent table shape which is reversed from Go.
// it switches the dimension order in the given table, for all columns
// that are float 2D or 4D columns -- assumes these are layer patterns
// and names dimensions accordingly.
func ReshapeCpp(dt *etable.Table) {
	for _, cl := range dt.Cols {
		shp := cl.Shapes()
		if cl.NumDims() == 3 && (cl.DataType() == etensor.FLOAT32 || cl.DataType() == etensor.FLOAT64) {
			revshp := []int{shp[0], shp[2], shp[1]} // [0] = row
			dnms := []string{"Row", "Y", "X"}
			cl.SetShape(revshp, nil, dnms)
		}
		if cl.NumDims() == 5 && (cl.DataType() == etensor.FLOAT32 || cl.DataType() == etensor.FLOAT64) {
			revshp := []int{shp[0], shp[4], shp[3], shp[2], shp[1]} // [0] = row
			dnms := []string{"Row", "PoolY", "PoolX", "NeurY", "NeurX"}
			cl.SetShape(revshp, nil, dnms)
		}
	}
}

// ReshapeCppFile fixes C++ emergent table shape which is reversed from Go.
// It loads file from fname and saves to fixnm
func ReshapeCppFile(dt *etable.Table, fname, fixnm string) {
	err := dt.OpenCSV(gi.FileName(fname), etable.Tab)
	if err != nil {
		log.Println(err)
		return
	}
	ReshapeCpp(dt)
	dt.SaveCSV(gi.FileName(fixnm), etable.Tab, true)
}
