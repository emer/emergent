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

// ConfigPatsTrain configures training patterns using first listSize rows in the vocabulary list
// design order: left right, bottom up
func ConfigPatsTrain(dt, poolVocab *etable.Table, design []string) {
	name := dt.MetaData["name"]
	inputName := dt.ColName(1)
	outputName := dt.ColName(2)
	listSize := dt.ColByName(inputName).Shapes()[0]
	ySize := dt.ColByName(inputName).Shapes()[1]
	xSize := dt.ColByName(inputName).Shapes()[2]
	for row := 0; row < listSize; row++ {
		dt.CellTensor("Name", row).SetString([]int{0}, fmt.Sprint(name, row))
		npool := 0
		for pooly := 0; pooly < ySize; pooly++ {
			for poolx := 0; poolx < xSize; poolx++ {
				tmpInp := dt.CellTensor(inputName, row).SubSpace([]int{pooly, poolx})
				tmpOut := dt.CellTensor(outputName, row).SubSpace([]int{pooly, poolx})
				tmpInp.CopyFrom(poolVocab.CellTensor(design[npool], row))
				tmpOut.CopyFrom(tmpInp)
				npool++
			}
		}
	}
}

// ConfigPatsTest configures testing patterns using training patterns
// voidDesign: bool slices to decide which pools will be turned to void in Input
func ConfigPatsTest(dt, trainDT *etable.Table, voidDesign []bool) error {
	inputName := dt.ColName(1)
	outputName := dt.ColName(2)
	testShapes := dt.ColByName(inputName).Shapes()
	trainShapes := trainDT.ColByName(inputName).Shapes()

	// check shape
	fmt.Println(testShapes)
	fmt.Println(trainShapes)
	if !reflect.DeepEqual(testShapes, trainShapes) {
		fmt.Println("hit an error but not terminating?")
		err := fmt.Errorf("Train and Test should have the same shape") // how do I stop the program?
		fmt.Println(err.Error())
	}

	name := dt.MetaData["name"]
	listSize := testShapes[0]
	ySize := testShapes[1]
	xSize := testShapes[2]
	for row := 0; row < listSize; row++ {
		dt.CellTensor("Name", row).SetString([]int{0}, fmt.Sprint(name, row))
		npool := 0
		for pooly := 0; pooly < ySize; pooly++ {
			for poolx := 0; poolx < xSize; poolx++ {
				tmpInp := dt.CellTensor(inputName, row).SubSpace([]int{pooly, poolx})
				tmpOut := dt.CellTensor(outputName, row).SubSpace([]int{pooly, poolx})
				tmpInp.CopyFrom(trainDT.CellTensor(inputName, row).SubSpace([]int{pooly, poolx}))
				tmpOut.CopyFrom(trainDT.CellTensor(outputName, row).SubSpace([]int{pooly, poolx}))
				if voidDesign[npool] {
					tmpInp.SetZeros()
				}
				npool++
			}
		}
	}
	return nil
}

// PatsFlipBits flips nFlipBits bits in pools specified by flipDesign
func PatsFlipBits(dt *etable.Table, flipDesign []bool, nFlipBits int) {
	inputName := dt.ColName(1)
	outputName := dt.ColName(2)
	dtShapes := dt.ColByName(inputName).Shapes()
	listSize := dtShapes[0]
	ySize := dtShapes[1]
	xSize := dtShapes[2]
	for row := 0; row < listSize; row++ {
		npool := 0
		for pooly := 0; pooly < ySize; pooly++ {
			for poolx := 0; poolx < xSize; poolx++ {
				tmpInp := dt.CellTensor(inputName, row).SubSpace([]int{pooly, poolx})
				tmpOut := dt.CellTensor(outputName, row).SubSpace([]int{pooly, poolx})
				if flipDesign[npool] {
					FlipBits(tmpInp, nFlipBits, nFlipBits, 1, 0)
					tmpOut.CopyFrom(tmpInp)
				}
				npool++
			}
		}
	}
}

// Example codes for TrainAB and TestAB in Hip.go
// ss.ConfigVocab(ss.PoolVocab, ss.ECPool.Y, ss.ECPool.X, ss.ECPctAct, []string{"A", "B", "ctxt1a", "ctxt1b", "ctxt1c", "ctxt1d"})
// for _, poolname := range []string{"ctxt1a", "ctxt1b", "ctxt1c", "ctxt1d"} { // the context for A-B should remain the same
// 	for _, row := range []int{1, 2, 3, 4, 5, 6, 7, 8, 9} { // thus we copy the tensors from row1 to row9 from row0
// 		ss.PoolVocab.CellTensor(poolname, row).CopyFrom(ss.PoolVocab.CellTensor(poolname, 0))
// 	}
// }
// patgen.InitPats(ss.TrainAB, "TrainAB", "AB Training Patterns", "Input", "ECout", 10, ss.YSize, ss.XSize, ss.ECPool.Y, ss.ECPool.X)
// patgen.InitPats(ss.TestAB, "TestAB", "AB Testing Patterns", "Input", "ECout", 10, ss.YSize, ss.XSize, ss.ECPool.Y, ss.ECPool.X)
// patgen.ConfigPatsTrain(ss.TrainAB, ss.PoolVocab, []string{"A", "B", "ctxt1a", "ctxt1b", "ctxt1c", "ctxt1d"})
// patgen.PatsFlipBits(ss.TrainAB, []bool{false, false, true, true, true, true}, 1) // bool slices to decide which pool to flip
// patgen.ConfigPatsTest(ss.TestAB, ss.TrainAB, []bool{false, true, false, false, false, false}) // bool slices to decide which pool to zero in Input
