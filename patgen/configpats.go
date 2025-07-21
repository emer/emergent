// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package patgen

import (
	"fmt"
	"log"
	"reflect"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/tensor/table"
)

// InitPats initiates patterns to be used in MixPats
func InitPats(dt *table.Table, name, desc, inputName, outputName string, listSize, ySize, xSize, poolY, poolX int) {
	dt.DeleteAll()
	dt.SetMetaData("name", name)
	dt.SetMetaData("desc", desc)
	dt.AddStringColumn("Name")
	dt.AddFloat32TensorColumn(inputName, []int{ySize, xSize, poolY, poolX}, "ySize", "xSize", "poolY", "poolX")
	dt.AddFloat32TensorColumn(outputName, []int{ySize, xSize, poolY, poolX}, "ySize", "xSize", "poolY", "poolX")
	dt.SetNumRows(listSize)
}

// MixPats mixes patterns using first listSize rows in the vocabulary map
// poolSource order: left right, bottom up
func MixPats(dt *table.Table, mp Vocab, colName string, poolSource []string) error {
	name := dt.MetaData["name"]
	listSize := errors.Log1(dt.ColumnByName(colName)).Shape().Sizes[0]
	ySize := errors.Log1(dt.ColumnByName(colName)).Shape().Sizes[1]
	xSize := errors.Log1(dt.ColumnByName(colName)).Shape().Sizes[2]
	for row := 0; row < listSize; row++ {
		dt.SetString("Name", row, fmt.Sprint(name, row))
		npool := 0
		for iY := 0; iY < ySize; iY++ {
			for iX := 0; iX < xSize; iX++ {
				trgPool := dt.Tensor(colName, row).SubSpace([]int{iY, iX})
				vocNm := poolSource[npool]
				voc, ok := mp[vocNm]
				if !ok {
					err := fmt.Errorf("Vocab not found: %s", vocNm)
					log.Println(err.Error())
					return err
				}
				vocSize := voc.Shape().Sizes[0]
				effIndex := row % vocSize // be safe and wrap-around to re-use patterns
				frmPool := voc.SubSpace([]int{effIndex})
				if !reflect.DeepEqual(trgPool.Shape().Sizes, frmPool.Shape().Sizes) {
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
func MixPatsN(dt *table.Table, mp Vocab, colName string, poolSource []string, targRow, vocabStart, vocabN int) error {
	name := dt.MetaData["name"]
	_ = name
	ySize := errors.Log1(dt.ColumnByName(colName)).Shape().Sizes[1]
	xSize := errors.Log1(dt.ColumnByName(colName)).Shape().Sizes[2]
	for ri := 0; ri < vocabN; ri++ {
		row := targRow + ri
		vocIndex := vocabStart + ri
		dt.SetString("Name", row, fmt.Sprint(name, row))
		npool := 0
		for iY := 0; iY < ySize; iY++ {
			for iX := 0; iX < xSize; iX++ {
				trgPool := dt.Tensor(colName, row).SubSpace([]int{iY, iX})
				vocNm := poolSource[npool]
				voc, ok := mp[vocNm]
				if !ok {
					err := fmt.Errorf("Vocab not found: %s", vocNm)
					log.Println(err.Error())
					return err
				}
				vocSize := voc.Shape().Sizes[0]
				effIndex := vocIndex % vocSize // be safe and wrap-around to re-use patterns
				frmPool := voc.SubSpace([]int{effIndex})
				if !reflect.DeepEqual(trgPool.Shape().Sizes, frmPool.Shape().Sizes) {
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
