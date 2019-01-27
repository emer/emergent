// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/etensor"
	"github.com/emer/emergent/prjn"
)

var TestNet Network
var InPats *etensor.Float32

var Pars = emer.ParamStyle{
	"Prjn": {
		"Prjn.Learn.WtInit.Var": 0, // for reproducibility, identical weights
	},
	".TopDown": {
		"Prjn.WtScale.Rel": 0.2,
	},
}

func TestMakeNet(t *testing.T) {
	TestNet.Name = "TestNet"
	inLay := TestNet.AddLayer("Input", []int{4, 1}, Input)
	hidLay := TestNet.AddLayer("Hidden", []int{4, 1}, Hidden)
	outLay := TestNet.AddLayer("Output", []int{4, 1}, Target)

	TestNet.ConnectLayers(hidLay, inLay, prjn.NewOneToOne())
	TestNet.ConnectLayers(outLay, hidLay, prjn.NewOneToOne())
	outHid := TestNet.ConnectLayers(hidLay, outLay, prjn.NewOneToOne())
	outHid.Class = "TopDown"

	TestNet.Defaults()
	TestNet.StyleParams(Pars)
	TestNet.Build()
	TestNet.InitWts()
	TestNet.TrialInit() // get GeScale

	var buf bytes.Buffer
	TestNet.WriteWtsJSON(&buf)
	wb := buf.Bytes()
	fmt.Printf("TestNet Weights:\n\n%v\n", string(wb))

	fp, err := os.Create("testdata/testnet.wts")
	defer fp.Close()
	if err != nil {
		t.Error(err)
	}
	fp.Write(wb)
}

func TestInPats(t *testing.T) {
	InPats = etensor.NewFloat32([]int{4, 4, 1}, nil, []string{"pat", "Y", "X"})
	for pi := 0; pi < 4; pi++ {
		InPats.Set([]int{pi, pi, 0}, 1)
	}
}

func TestNetAct(t *testing.T) {
	TestNet.InitWts()
	TestNet.InitExt()

	inLay := TestNet.LayerByName("Input").(*Layer)
	hidLay := TestNet.LayerByName("Hidden").(*Layer)
	outLay := TestNet.LayerByName("Output").(*Layer)

	time := NewTime()

	for pi := 0; pi < 4; pi++ {
		inpat, err := InPats.SubSlice(2, []int{pi})
		if err != nil {
			t.Error(err)
		}
		inLay.ApplyExt(inpat)
		outLay.ApplyExt(inpat)

		TestNet.TrialInit()
		time.TrialStart()
		for qtr := 0; qtr < 4; qtr++ {
			for cyc := 0; cyc < time.CycPerQtr; cyc++ {
				TestNet.Cycle()
				time.CycleInc()
			}
			TestNet.QuarterFinal(time)
			time.QuarterInc()

			inActs := inLay.UnitVals("Act")
			hidActs := hidLay.UnitVals("Act")
			hidGes := hidLay.UnitVals("Ge")
			outActs := outLay.UnitVals("Act")
			outGes := outLay.UnitVals("Ge")
			fmt.Printf("pat: %v qtr: %v\nin acts: %v\nhid acts: %v ges: %v\nout acts: %v ges: %v\n", pi, qtr, inActs, hidActs, hidGes, outActs, outGes)

			// C++ emergent:
			// q0: in act: .95
			// hid act: 0.943078, net: 0.475499
			// out act: 0.961023, net: 0.471143
		}
	}
}
