// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/chewxy/math32"
	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/etensor"
	"github.com/emer/emergent/prjn"
)

// Note: this test project exactly reproduces the configuration and behavior of
// C++ emergent/demo/leabra/basic_leabra_test.proj  in version 8.5.6 svn 11492

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
	// fmt.Printf("TestNet Weights:\n\n%v\n", string(wb))

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

func CmprFloats(out, cor []float32, msg string, t *testing.T) {
	for i := range out {
		dif := math32.Abs(out[i] - cor[i])
		if dif > difTol { // allow for small numerical diffs
			t.Errorf("%v err: out: %v, cor: %v, dif: %v\n", msg, out[i], cor[i], dif)
		}
	}
}

func TestNetAct(t *testing.T) {
	TestNet.InitWts()
	TestNet.InitExt()

	inLay := TestNet.LayerByName("Input").(*Layer)
	hidLay := TestNet.LayerByName("Hidden").(*Layer)
	outLay := TestNet.LayerByName("Output").(*Layer)

	time := NewTime()

	printCycs := false
	printQtrs := false

	qtr0HidActs := []float32{0.9427379, 2.4012093e-33, 2.4012093e-33, 2.4012093e-33}
	qtr0HidGes := []float32{0.47417355, 0, 0, 0}
	qtr0HidGis := []float32{0.45752862, 0.45752862, 0.45752862, 0.45752862}
	qtr0OutActs := []float32{0.94144684, 2.4021936e-33, 2.4021936e-33, 2.4021936e-33}
	qtr0OutGes := []float32{0.47107852, 0, 0, 0}
	qtr0OutGis := []float32{0.45534685, 0.45534685, 0.45534685, 0.45534685}

	qtr3HidActs := []float32{0.9431544, 4e-45, 4e-45, 4e-45}
	qtr3HidGes := []float32{0.47499993, 0, 0, 0}
	qtr3HidGis := []float32{0.45816946, 0.45816946, 0.45816946, 0.45816946}
	qtr3OutActs := []float32{0.95, 0, 0, 0}
	qtr3OutGes := []float32{0.47114015, 0, 0, 0}
	qtr3OutGis := []float32{0.45951304, 0.45951304, 0.45951304, 0.45951304}

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

				if printCycs {
					inActs := inLay.UnitVals("Act")
					hidActs := hidLay.UnitVals("Act")
					hidGes := hidLay.UnitVals("Ge")
					hidGis := hidLay.UnitVals("Gi")
					outActs := outLay.UnitVals("Act")
					outGes := outLay.UnitVals("Ge")
					outGis := outLay.UnitVals("Gi")
					fmt.Printf("pat: %v qtr: %v cyc: %v\nin acts: %v\nhid acts: %v ges: %v gis: %v\nout acts: %v ges: %v gis: %v\n", pi, qtr, cyc, inActs, hidActs, hidGes, hidGis, outActs, outGes, outGis)
				}
			}
			TestNet.QuarterFinal(time)
			time.QuarterInc()

			if printCycs && printQtrs {
				fmt.Printf("=============================\n")
			}

			inActs := inLay.UnitVals("Act")
			hidActs := hidLay.UnitVals("Act")
			hidGes := hidLay.UnitVals("Ge")
			hidGis := hidLay.UnitVals("Gi")
			outActs := outLay.UnitVals("Act")
			outGes := outLay.UnitVals("Ge")
			outGis := outLay.UnitVals("Gi")

			if printQtrs {
				fmt.Printf("pat: %v qtr: %v cyc: %v\nin acts: %v\nhid acts: %v ges: %v gis: %v\nout acts: %v ges: %v gis: %v\n", pi, qtr, time.Cycle, inActs, hidActs, hidGes, hidGis, outActs, outGes, outGis)
			}

			if printCycs && printQtrs {
				fmt.Printf("=============================\n")
			}

			if pi == 0 && qtr == 0 {
				CmprFloats(hidActs, qtr0HidActs, "qtr 0 hidActs", t)
				CmprFloats(hidGes, qtr0HidGes, "qtr 0 hidGes", t)
				CmprFloats(hidGis, qtr0HidGis, "qtr 0 hidGis", t)
				CmprFloats(outActs, qtr0OutActs, "qtr 0 outActs", t)
				CmprFloats(outGes, qtr0OutGes, "qtr 0 outGes", t)
				CmprFloats(outGis, qtr0OutGis, "qtr 0 outGis", t)
			}
			if pi == 0 && qtr == 3 {
				CmprFloats(hidActs, qtr3HidActs, "qtr 3 hidActs", t)
				CmprFloats(hidGes, qtr3HidGes, "qtr 3 hidGes", t)
				CmprFloats(hidGis, qtr3HidGis, "qtr 3 hidGis", t)
				CmprFloats(outActs, qtr3OutActs, "qtr 3 outActs", t)
				CmprFloats(outGes, qtr3OutGes, "qtr 3 outGes", t)
				CmprFloats(outGis, qtr3OutGis, "qtr 3 outGis", t)
			}
		}

		if printQtrs {
			fmt.Printf("=============================\n")
		}
	}
}

func TestNetLearn(t *testing.T) {
	TestNet.InitWts()
	TestNet.InitExt()

	inLay := TestNet.LayerByName("Input").(*Layer)
	hidLay := TestNet.LayerByName("Hidden").(*Layer)
	outLay := TestNet.LayerByName("Output").(*Layer)

	time := NewTime()

	printCycs := false
	printQtrs := false

	qtr0HidAvgS := []float32{0.9422413, 6.034972e-08, 6.034972e-08, 6.034972e-08}
	qtr0HidAvgM := []float32{0.8162388, 0.013628835, 0.013628835, 0.013628835}
	qtr0OutAvgS := []float32{0.93967456, 6.034972e-08, 6.034972e-08, 6.034972e-08}
	qtr0OutAvgM := []float32{0.7438192, 0.013628835, 0.013628835, 0.013628835}

	qtr3HidAvgS := []float32{0.94315434, 6.0347804e-30, 6.0347804e-30, 6.0347804e-30}
	qtr3HidAvgM := []float32{0.94308215, 5.042516e-06, 5.042516e-06, 5.042516e-06}
	qtr3OutAvgS := []float32{0.9499999, 6.0347804e-30, 6.0347804e-30, 6.0347804e-30}
	qtr3OutAvgM := []float32{0.9492211, 5.042516e-06, 5.042516e-06, 5.042516e-06}

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

				hidAct := hidLay.UnitVals("Act")
				hidGes := hidLay.UnitVals("Ge")
				hidGis := hidLay.UnitVals("Gi")
				hidAvgSS := hidLay.UnitVals("AvgSS")
				hidAvgS := hidLay.UnitVals("AvgS")
				hidAvgM := hidLay.UnitVals("AvgM")

				outAvgS := outLay.UnitVals("AvgS")
				outAvgM := outLay.UnitVals("AvgM")

				if printCycs {
					fmt.Printf("pat: %v qtr: %v cyc: %v\nhid act: %v ges: %v gis: %v\nhid avgss: %v avgs: %v avgm: %v\nout avgs: %v avgm: %v\n", pi, qtr, time.Cycle, hidAct, hidGes, hidGis, hidAvgSS, hidAvgS, hidAvgM, outAvgS, outAvgM)
				}

			}
			TestNet.QuarterFinal(time)
			time.QuarterInc()

			hidAvgS := hidLay.UnitVals("AvgS")
			hidAvgM := hidLay.UnitVals("AvgM")

			outAvgS := outLay.UnitVals("AvgS")
			outAvgM := outLay.UnitVals("AvgM")

			if printQtrs {
				fmt.Printf("pat: %v qtr: %v cyc: %v\nhid avgs: %v avgm: %v\nout avgs: %v avgm: %v\n", pi, qtr, time.Cycle, hidAvgS, hidAvgM, outAvgS, outAvgM)
			}

			if pi == 0 && qtr == 0 {
				CmprFloats(hidAvgS, qtr0HidAvgS, "qtr 0 hidAvgS", t)
				CmprFloats(hidAvgM, qtr0HidAvgM, "qtr 0 hidAvgM", t)
				CmprFloats(outAvgS, qtr0OutAvgS, "qtr 0 outAvgS", t)
				CmprFloats(outAvgM, qtr0OutAvgM, "qtr 0 outAvgM", t)
			}
			if pi == 0 && qtr == 3 {
				CmprFloats(hidAvgS, qtr3HidAvgS, "qtr 3 hidAvgS", t)
				CmprFloats(hidAvgM, qtr3HidAvgM, "qtr 3 hidAvgM", t)
				CmprFloats(outAvgS, qtr3OutAvgS, "qtr 3 outAvgS", t)
				CmprFloats(outAvgM, qtr3OutAvgM, "qtr 3 outAvgM", t)
			}
		}

		if printQtrs {
			fmt.Printf("=============================\n")
		}
	}
}
