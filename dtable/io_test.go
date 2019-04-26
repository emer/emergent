// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dtable

import (
	"os"
	"strings"
	"testing"

	"github.com/emer/emergent/etensor"

	"github.com/apache/arrow/go/arrow"
)

func TestEmerHeaders(t *testing.T) {
	hdrstr := `_H:	$Name	%Input[2:0,0]<2:5,5>	%Input[2:1,0]	%Input[2:2,0]	%Input[2:3,0]	%Input[2:4,0]	%Input[2:0,1]	%Input[2:1,1]	%Input[2:2,1]	%Input[2:3,1]	%Input[2:4,1]	%Input[2:0,2]	%Input[2:1,2]	%Input[2:2,2]	%Input[2:3,2]	%Input[2:4,2]	%Input[2:0,3]	%Input[2:1,3]	%Input[2:2,3]	%Input[2:3,3]	%Input[2:4,3]	%Input[2:0,4]	%Input[2:1,4]	%Input[2:2,4]	%Input[2:3,4]	%Input[2:4,4]	%Output[2:0,0]<2:5,5>	%Output[2:1,0]	%Output[2:2,0]	%Output[2:3,0]	%Output[2:4,0]	%Output[2:0,1]	%Output[2:1,1]	%Output[2:2,1]	%Output[2:3,1]	%Output[2:4,1]	%Output[2:0,2]	%Output[2:1,2]	%Output[2:2,2]	%Output[2:3,2]	%Output[2:4,2]	%Output[2:0,3]	%Output[2:1,3]	%Output[2:2,3]	%Output[2:3,3]	%Output[2:4,3]	%Output[2:0,4]	%Output[2:1,4]	%Output[2:2,4]	%Output[2:3,4]	%Output[2:4,4]	`

	hdrs := strings.Split(hdrstr, "\t")
	sc, err := SchemaFromHeaders(hdrs)
	if err != nil {
		t.Error(err)
	}
	// fmt.Printf("schema:\n%v\n", sc)
	if len(sc) != 3 {
		t.Errorf("EmerHeaders: len != 3\n")
	}
	if sc[0].Type != etensor.Type(etensor.STRING) {
		t.Errorf("EmerHeaders: sc[0] != STRING\n")
	}
	if sc[1].Type != etensor.Type(etensor.FLOAT32) {
		t.Errorf("EmerHeaders: sc[1] != FLOAT32\n")
	}
	if sc[2].Type != etensor.Type(etensor.FLOAT32) {
		t.Errorf("EmerHeaders: sc[2] != FLOAT32\n")
	}
	if sc[1].CellShape[0] != 5 {
		t.Errorf("EmerHeaders: sc[1].CellShape[0] != 5\n")
	}
	if sc[1].CellShape[1] != 5 {
		t.Errorf("EmerHeaders: sc[1].CellShape[1] != 5\n")
	}
	if sc[2].CellShape[0] != 5 {
		t.Errorf("EmerHeaders: sc[2].CellShape[0] != 5\n")
	}
	if sc[2].CellShape[1] != 5 {
		t.Errorf("EmerHeaders: sc[2].CellShape[1] != 5\n")
	}
	dt := New(sc, 0)
	outh := dt.EmerHeaders()
	// fmt.Printf("headers out:\n%v\n", outh)
	for i := 0; i < 2; i++ { // note: due to diff row-major index ordering, other cols are diff..
		hh := hdrs[i]
		oh := outh[i]
		if hh != oh {
			t.Errorf("EmerHeaders: hdr %v mismatch %v != %v\n", i, hh, oh)
		}
	}
	if hdrs[27] != outh[27] {
		t.Errorf("EmerHeaders: hdr %v mismatch %v != %v\n", 27, hdrs[27], outh[27])
	}
}

func TestReadEmerDat(t *testing.T) {
	for i := 0; i < 2; i++ {
		fp, err := os.Open("testdata/emer_simple_lines_5x5.dat")
		defer fp.Close()
		if err != nil {
			t.Error(err)
		}
		dt := &Table{}
		err = dt.ReadCSV(fp, '\t') // tsv
		if err != nil {
			t.Error(err)
		}
		sc := dt.Cols
		if len(sc) != 3 {
			t.Errorf("EmerHeaders: len != 3\n")
		}
		if sc[0].DataType().ID() != arrow.STRING {
			t.Errorf("EmerHeaders: sc[0] != STRING\n")
		}
		if sc[1].DataType().ID() != arrow.FLOAT32 {
			t.Errorf("EmerHeaders: sc[1] != FLOAT32\n")
		}
		if sc[2].DataType().ID() != arrow.FLOAT32 {
			t.Errorf("EmerHeaders: sc[2] != FLOAT32\n")
		}
		if sc[1].Dim(0) != 6 {
			t.Errorf("EmerHeaders: sc[1].Dim[0] != 6 = %v\n", sc[1].Dim(0))
		}
		if sc[1].Dim(1) != 5 {
			t.Errorf("EmerHeaders: sc[1].Dim[1] != 5\n")
		}
		if sc[2].Dim(0) != 6 {
			t.Errorf("EmerHeaders: sc[2].Dim[0] != 6 = %v\n", sc[2].Dim(0))
		}
		if sc[2].Dim(1) != 5 {
			t.Errorf("EmerHeaders: sc[2].Dim[1] != 5\n")
		}
		fo, err := os.Create("testdata/emer_simple_lines_5x5_rec.dat")
		defer fo.Close()
		if err != nil {
			t.Error(err)
		}
		dt.WriteCSV(fo, '\t', true)
	}
}
