// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package patgen

import (
	"fmt"
	"log"
	"reflect"

	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
)

// InitPats initiates patterns to be used in MixPats
func InitPats(dt *etable.Table, name, desc, inputName, outputName string, listSize, ySize, xSize, poolY, poolX int) {
	dt.SetMetaData("name", name)
	dt.SetMetaData("desc", desc)
	dt.SetFromSchema(etable.Schema{
		{"Name", etensor.STRING, nil, nil},
		{inputName, etensor.FLOAT32, []int{ySize, xSize, poolY, poolX}, []string{"ySize", "xSize", "poolY", "poolX"}},
		{outputName, etensor.FLOAT32, []int{ySize, xSize, poolY, poolX}, []string{"ySize", "xSize", "poolY", "poolX"}},
	}, listSize)
}

// MixPats mixes patterns using first listSize rows in the vocabulary map
// poolSource order: left right, bottom up
func MixPats(dt *etable.Table, mp Vocab, colName string, poolSource []string) error {
	name := dt.MetaData["name"]
	listSize := dt.ColByName(colName).Shapes()[0]
	ySize := dt.ColByName(colName).Shapes()[1]
	xSize := dt.ColByName(colName).Shapes()[2]
	for row := 0; row < listSize; row++ {
		dt.SetCellString("Name", row, fmt.Sprint(name, row))
		npool := 0
		for iY := 0; iY < ySize; iY++ {
			for iX := 0; iX < xSize; iX++ {
				trgPool := dt.CellTensor(colName, row).SubSpace([]int{iY, iX})
				vocNm := poolSource[npool]
				voc, ok := mp[vocNm]
				if !ok {
					err := fmt.Errorf("Vocab not found: %s", vocNm)
					log.Println(err.Error())
					return err
				}
				vocSize := voc.Shapes()[0]
				effIdx := row % vocSize // be safe and wrap-around to re-use patterns
				frmPool := voc.SubSpace([]int{effIdx})
				if !reflect.DeepEqual(trgPool.Shapes(), frmPool.Shapes()) {
					err := fmt.Errorf("Vocab and pools in the table should have the same shape")
					log.Println(err.Error())
					return err
				}
				trgPool.CopyFrom(frmPool)
				npool++
			}
		}
	}
	return nil
}

// MixPatsN mixes patterns using specified startVocab and vocabN numbers
// of vocabulary patterns, inserting starting at specified targRow in table.
// poolSource order: left right, bottom up
func MixPatsN(dt *etable.Table, mp Vocab, colName string, poolSource []string, targRow, vocabStart, vocabN int) error {
	name := dt.MetaData["name"]
	_ = name
	ySize := dt.ColByName(colName).Shapes()[1]
	xSize := dt.ColByName(colName).Shapes()[2]
	for ri := 0; ri < vocabN; ri++ {
		row := targRow + ri
		vocIdx := vocabStart + ri
		dt.SetCellString("Name", row, fmt.Sprint(name, row))
		npool := 0
		for iY := 0; iY < ySize; iY++ {
			for iX := 0; iX < xSize; iX++ {
				trgPool := dt.CellTensor(colName, row).SubSpace([]int{iY, iX})
				vocNm := poolSource[npool]
				voc, ok := mp[vocNm]
				if !ok {
					err := fmt.Errorf("Vocab not found: %s", vocNm)
					log.Println(err.Error())
					return err
				}
				vocSize := voc.Shapes()[0]
				effIdx := vocIdx % vocSize // be safe and wrap-around to re-use patterns
				frmPool := voc.SubSpace([]int{effIdx})
				if !reflect.DeepEqual(trgPool.Shapes(), frmPool.Shapes()) {
					err := fmt.Errorf("Vocab and pools in the table should have the same shape")
					log.Println(err.Error())
					return err
				}
				trgPool.CopyFrom(frmPool)
				npool++
			}
		}
	}
	return nil
}
