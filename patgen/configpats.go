// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package patgen

import (
	"fmt"
	"log"
	"slices"

	"cogentcore.org/core/base/metadata"
	"cogentcore.org/core/tensor/table"
)

// InitPats initiates patterns to be used in MixPats
func InitPats(dt *table.Table, name, doc, inputName, outputName string, listSize, ySize, xSize, poolY, poolX int) {
	dt.DeleteAll()
	metadata.SetName(dt, name)
	metadata.SetDoc(dt, doc)
	dt.AddStringColumn("Name")
	dt.AddFloat32Column(inputName, ySize, xSize, poolY, poolX)
	dt.AddFloat32Column(outputName, ySize, xSize, poolY, poolX)
	dt.SetNumRows(listSize)
}

// MixPats mixes patterns using first listSize rows in the vocabulary map
// poolSource order: left right, bottom up
func MixPats(dt *table.Table, mp Vocab, colName string, poolSource []string) error {
	name := metadata.Name(dt)
	listSize := dt.Column(colName).DimSize(0)
	ySize := dt.Column(colName).DimSize(1)
	xSize := dt.Column(colName).DimSize(2)
	for row := 0; row < listSize; row++ {
		dt.Column("Name").SetString1D(fmt.Sprint(name, row), row)
		npool := 0
		for iY := 0; iY < ySize; iY++ {
			for iX := 0; iX < xSize; iX++ {
				trgPool := dt.Column(colName).SubSpace(row, iY, iX)
				vocNm := poolSource[npool]
				voc, ok := mp[vocNm]
				if !ok {
					err := fmt.Errorf("Vocab not found: %s", vocNm)
					log.Println(err.Error())
					return err
				}
				vocSize := voc.DimSize(0)
				effIndex := row % vocSize // be safe and wrap-around to re-use patterns
				frmPool := voc.SubSpace(effIndex)
				if !slices.Equal(trgPool.Shape().Sizes, frmPool.Shape().Sizes) {
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
	name := metadata.Name(dt)
	_ = name
	ySize := dt.Column(colName).DimSize(1)
	xSize := dt.Column(colName).DimSize(2)
	for ri := 0; ri < vocabN; ri++ {
		row := targRow + ri
		vocIndex := vocabStart + ri
		dt.Column("Name").SetString1D(fmt.Sprint(name, row), row)
		npool := 0
		for iY := 0; iY < ySize; iY++ {
			for iX := 0; iX < xSize; iX++ {
				trgPool := dt.Column(colName).SubSpace(row, iY, iX)
				vocNm := poolSource[npool]
				voc, ok := mp[vocNm]
				if !ok {
					err := fmt.Errorf("Vocab not found: %s", vocNm)
					log.Println(err.Error())
					return err
				}
				vocSize := voc.Shape().Sizes[0]
				effIndex := vocIndex % vocSize // be safe and wrap-around to re-use patterns
				frmPool := voc.SubSpace(effIndex)
				if !slices.Equal(trgPool.Shape().Sizes, frmPool.Shape().Sizes) {
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
