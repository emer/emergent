// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package patgen

import (
	"fmt"
	"reflect"

	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
)

// InitPats initiates patterns to be used in ConfigPatsTrain
func InitPats(dt *etable.Table, name, desc, inputName, outputName string, listSize, ySize, xSize, poolY, poolX int) {
	dt.SetMetaData("name", name)
	dt.SetMetaData("desc", desc)
	dt.SetFromSchema(etable.Schema{
		{"Name", etensor.STRING, []int{1}, []string{"Name"}},
		{inputName, etensor.FLOAT32, []int{ySize, xSize, poolY, poolX}, []string{"ySize", "xSize", "poolY", "poolX"}},
		{outputName, etensor.FLOAT32, []int{ySize, xSize, poolY, poolX}, []string{"ySize", "xSize", "poolY", "poolX"}},
	}, listSize)
}

// ConfigPats configures patterns using first listSize rows in the vocabulary map
// poolSource order: left right, bottom up
func ConfigPats(dt *etable.Table, mp map[string]*etensor.Float32, colName string, poolSource []string) error {
	name := dt.MetaData["name"]
	listSize := dt.ColByName(colName).Shapes()[0]
	ySize := dt.ColByName(colName).Shapes()[1]
	xSize := dt.ColByName(colName).Shapes()[2]
	for row := 0; row < listSize; row++ {
		dt.CellTensor("Name", row).SetString([]int{0}, fmt.Sprint(name, row))
		npool := 0
		for iY := 0; iY < ySize; iY++ {
			for iX := 0; iX < xSize; iX++ {
				pool := dt.CellTensor(colName, row).SubSpace([]int{iY, iX})
				frmPool := mp[poolSource[npool]].SubSpace([]int{row})
				if !reflect.DeepEqual(pool.Shapes(), frmPool.Shapes()) {
					err := fmt.Errorf("Vocab and pools in the table should have the same shape") // how do I stop the program?
					// fmt.Println(err.Error())
					return err
				}
				pool.CopyFrom(frmPool)
				npool++
			}
		}
	}
	return nil
}
