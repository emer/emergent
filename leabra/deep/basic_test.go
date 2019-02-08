// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package deep

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/chewxy/math32"
	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/etensor"
	"github.com/emer/emergent/leabra/leabra"
	"github.com/emer/emergent/prjn"
)

// Note: this test project exactly reproduces the configuration and behavior of
// C++ emergent/demo/leabra/basic_leabra_test.proj  in version 8.5.6 svn 11492

// difTol is the numerical difference tolerance for comparing vs. target values
const difTol = float32(1.0e-10)

var TestNet Network
var InPats *etensor.Float32

// number of distinct sets of learning parameters to test
const NLrnPars = 4

var Pars = [NLrnPars]emer.ParamStyle{
	{
		"Prjn": {
			"Prjn.Learn.WtInit.Var":  0, // for reproducibility, identical weights
			"Prjn.Learn.Norm.On":     0,
			"Prjn.Learn.Momentum.On": 0,
		},
		".Back": {
			"Prjn.WtScale.Rel": 0.2,
		},
	},
	{
		"Prjn": {
			"Prjn.Learn.WtInit.Var":  0, // for reproducibility, identical weights
			"Prjn.Learn.Norm.On":     1,
			"Prjn.Learn.Momentum.On": 0,
		},
		".Back": {
			"Prjn.WtScale.Rel": 0.2,
		},
	},
	{
		"Prjn": {
			"Prjn.Learn.WtInit.Var":  0, // for reproducibility, identical weights
			"Prjn.Learn.Norm.On":     0,
			"Prjn.Learn.Momentum.On": 1,
		},
		".Back": {
			"Prjn.WtScale.Rel": 0.2,
		},
	},
	{
		"Prjn": {
			"Prjn.Learn.WtInit.Var":  0, // for reproducibility, identical weights
			"Prjn.Learn.Norm.On":     1,
			"Prjn.Learn.Momentum.On": 1,
		},
		".Back": {
			"Prjn.WtScale.Rel": 0.2,
		},
	},
}

func TestMakeNet(t *testing.T) {
	TestNet.InitName(&TestNet, "TestNet")
	inLay := TestNet.AddLayer("Input", []int{4, 1}, emer.Input)
	hidLay := TestNet.AddLayer("Hidden", []int{4, 1}, emer.Hidden)
	outLay := TestNet.AddLayer("Output", []int{4, 1}, emer.Target)

	TestNet.ConnectLayers(inLay, hidLay, prjn.NewOneToOne(), emer.Forward)
	TestNet.ConnectLayers(hidLay, outLay, prjn.NewOneToOne(), emer.Forward)
	TestNet.ConnectLayers(outLay, hidLay, prjn.NewOneToOne(), emer.Back)

	TestNet.Defaults()
	TestNet.StyleParams(Pars[0], false) // false) // true) // no msg
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

	time := leabra.NewTime()

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
					inActs, _ := inLay.UnitVals("Act")
					hidActs, _ := hidLay.UnitVals("Act")
					hidGes, _ := hidLay.UnitVals("GeInc")
					hidGis, _ := hidLay.UnitVals("Gi")
					outActs, _ := outLay.UnitVals("Act")
					outGes, _ := outLay.UnitVals("Ge")
					outGis, _ := outLay.UnitVals("Gi")
					fmt.Printf("pat: %v qtr: %v cyc: %v\nin acts: %v\nhid acts: %v ges: %v gis: %v\nout acts: %v ges: %v gis: %v\n", pi, qtr, cyc, inActs, hidActs, hidGes, hidGis, outActs, outGes, outGis)
				}
			}
			TestNet.QuarterFinal(time)
			time.QuarterInc()

			if printCycs && printQtrs {
				fmt.Printf("=============================\n")
			}

			inActs, _ := inLay.UnitVals("Act")
			hidActs, _ := hidLay.UnitVals("Act")
			hidGes, _ := hidLay.UnitVals("Ge")
			hidGis, _ := hidLay.UnitVals("Gi")
			outActs, _ := outLay.UnitVals("Act")
			outGes, _ := outLay.UnitVals("Ge")
			outGis, _ := outLay.UnitVals("Gi")

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
	inLay := TestNet.LayerByName("Input").(*Layer)
	hidLay := TestNet.LayerByName("Hidden").(*Layer)
	outLay := TestNet.LayerByName("Output").(*Layer)

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

	trl0HidAvgL := []float32{0.3975, 0.3975, 0.3975, 0.3975}
	trl1HidAvgL := []float32{0.5935205, 0.35775128, 0.35775128, 0.35775128}
	trl2HidAvgL := []float32{0.5341764, 0.55774546, 0.32197616, 0.32197616}
	trl3HidAvgL := []float32{0.48075876, 0.5019788, 0.5255478, 0.28977853}

	trl1HidAvgLLrn := []float32{0.0008553083, 0.00034286897, 0.00034286897, 0.00034286897}
	trl2HidAvgLLrn := []float32{0.0007263252, 0.00077755196, 0.00026511253, 0.00026511253}
	trl3HidAvgLLrn := []float32{0.00061022304, 0.0006563444, 0.0007075711, 0.00019513167}

	// these are organized by pattern within and then by test iteration (params) outer
	hidDwts := []float32{3.376007e-06, 1.1105859e-05, 9.811188e-06, 8.4557105e-06,
		0.00050640106, 0.0016658787, 0.0014716781, 0.0012683566,
		3.376007e-07, 1.1105858e-06, 9.811188e-07, 8.4557104e-07,
		5.0640105e-05, 0.00016658788, 0.00014716781, 0.00012683566}
	outDwts := []float32{2.8908253e-05, 2.9251574e-05, 2.9251574e-05, 2.9251574e-05,
		0.004336238, 0.0043877363, 0.0043877363, 0.0043877363,
		2.8908253e-06, 2.9251576e-06, 2.9251576e-06, 2.9251576e-06,
		0.0004336238, 0.00043877363, 0.00043877363, 0.00043877363}
	hidNorms := []float32{0, 0, 0, 0, 8.440018e-05, 0.00027764647, 0.0002452797, 0.00021139276,
		0, 0, 0, 0, 8.440018e-05, 0.00027764647, 0.0002452797, 0.00021139276}
	outNorms := []float32{0, 0, 0, 0, 0.0007227063, 0.0007312894, 0.0007312894, 0.0007312894,
		0, 0, 0, 0, 0.0007227063, 0.0007312894, 0.0007312894, 0.0007312894}
	hidMoments := []float32{0, 0, 0, 0, 0, 0, 0, 0,
		8.440018e-05, 0.00027764647, 0.0002452797, 0.00021139276,
		8.440018e-05, 0.00027764647, 0.0002452797, 0.00021139276}
	outMoments := []float32{0, 0, 0, 0, 0, 0, 0, 0,
		0.0007227063, 0.0007312894, 0.0007312894, 0.0007312894,
		0.0007227063, 0.0007312894, 0.0007312894, 0.0007312894}
	hidWts := []float32{0.50001, 0.50003326, 0.5000293, 0.5000254,
		0.50151914, 0.5049973, 0.5044148, 0.5038051,
		0.5000011, 0.5000032, 0.50000286, 0.5000025,
		0.500152, 0.5004996, 0.5004417, 0.5003805}
	outWts := []float32{0.50008655, 0.5000876, 0.5000876, 0.5000876,
		0.51300585, 0.5131602, 0.5131602, 0.5131602,
		0.5000086, 0.50000894, 0.50000894, 0.50000894,
		0.5013011, 0.5013164, 0.5013164, 0.5013164}

	hiddwt := make([]float32, 4*NLrnPars)
	outdwt := make([]float32, 4*NLrnPars)
	hidwt := make([]float32, 4*NLrnPars)
	outwt := make([]float32, 4*NLrnPars)
	hidnorm := make([]float32, 4*NLrnPars)
	outnorm := make([]float32, 4*NLrnPars)
	hidmoment := make([]float32, 4*NLrnPars)
	outmoment := make([]float32, 4*NLrnPars)

	for ti := 0; ti < NLrnPars; ti++ {
		TestNet.Defaults()
		TestNet.StyleParams(Pars[ti], false) // no msg
		TestNet.InitWts()
		TestNet.InitExt()

		time := leabra.NewTime()

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

					hidAct, _ := hidLay.UnitVals("Act")
					hidGes, _ := hidLay.UnitVals("Ge")
					hidGis, _ := hidLay.UnitVals("Gi")
					hidAvgSS, _ := hidLay.UnitVals("AvgSS")
					hidAvgS, _ := hidLay.UnitVals("AvgS")
					hidAvgM, _ := hidLay.UnitVals("AvgM")

					outAvgS, _ := outLay.UnitVals("AvgS")
					outAvgM, _ := outLay.UnitVals("AvgM")

					if printCycs {
						fmt.Printf("pat: %v qtr: %v cyc: %v\nhid act: %v ges: %v gis: %v\nhid avgss: %v avgs: %v avgm: %v\nout avgs: %v avgm: %v\n", pi, qtr, time.Cycle, hidAct, hidGes, hidGis, hidAvgSS, hidAvgS, hidAvgM, outAvgS, outAvgM)
					}

				}
				TestNet.QuarterFinal(time)
				time.QuarterInc()

				hidAvgS, _ := hidLay.UnitVals("AvgS")
				hidAvgM, _ := hidLay.UnitVals("AvgM")

				outAvgS, _ := outLay.UnitVals("AvgS")
				outAvgM, _ := outLay.UnitVals("AvgM")

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

			hidAvgL, _ := hidLay.UnitVals("AvgL")
			hidAvgLLrn, _ := hidLay.UnitVals("AvgLLrn")
			outAvgL, _ := outLay.UnitVals("AvgL")
			outAvgLLrn, _ := outLay.UnitVals("AvgLLrn")
			_ = outAvgL
			_ = outAvgLLrn

			// fmt.Printf("hid cosdif stats: %v\nhid avgl:   %v\nhid avgllrn: %v\n", hidLay.CosDiff, hidAvgL, hidAvgLLrn)
			// fmt.Printf("out cosdif stats: %v\nout avgl:   %v\nout avgllrn: %v\n", outLay.CosDiff, outAvgL, outAvgLLrn)

			TestNet.DWt()

			didx := ti*4 + pi

			hiddwt[didx], err = hidLay.RecvPrjns[0].SynVal("DWt", pi, pi)
			if err != nil {
				t.Error(err)
			}
			outdwt[didx], err = outLay.RecvPrjns[0].SynVal("DWt", pi, pi)
			if err != nil {
				t.Error(err)
			}
			hidnorm[didx], _ = hidLay.RecvPrjns[0].SynVal("Norm", pi, pi)
			outnorm[didx], _ = outLay.RecvPrjns[0].SynVal("Norm", pi, pi)
			hidmoment[didx], _ = hidLay.RecvPrjns[0].SynVal("Moment", pi, pi)
			outmoment[didx], _ = outLay.RecvPrjns[0].SynVal("Moment", pi, pi)

			TestNet.WtFmDWt()

			hidwt[didx], _ = hidLay.RecvPrjns[0].SynVal("Wt", pi, pi)
			outwt[didx], _ = outLay.RecvPrjns[0].SynVal("Wt", pi, pi)

			switch pi {
			case 0:
				CmprFloats(hidAvgL, trl0HidAvgL, "trl 0 hidAvgL", t)
			case 1:
				CmprFloats(hidAvgL, trl1HidAvgL, "trl 1 hidAvgL", t)
				CmprFloats(hidAvgLLrn, trl1HidAvgLLrn, "trl 1 hidAvgLLrn", t)
			case 2:
				CmprFloats(hidAvgL, trl2HidAvgL, "trl 2 hidAvgL", t)
				CmprFloats(hidAvgLLrn, trl2HidAvgLLrn, "trl 2 hidAvgLLrn", t)
			case 3:
				CmprFloats(hidAvgL, trl3HidAvgL, "trl 3 hidAvgL", t)
				CmprFloats(hidAvgLLrn, trl3HidAvgLLrn, "trl 3 hidAvgLLrn", t)
			}

		}
	}

	//	fmt.Printf("hid dwt: %v\nout dwt: %v\nhid norm: %v\n hid moment: %v\nout norm: %v\nout moment: %v\nhid wt: %v\nout wt: %v\n", hiddwt, outdwt, hidnorm, hidmoment, outnorm, outmoment, hidwt, outwt)

	CmprFloats(hiddwt, hidDwts, "hid DWt", t)
	CmprFloats(outdwt, outDwts, "out DWt", t)
	CmprFloats(hidnorm, hidNorms, "hid Norm", t)
	CmprFloats(outnorm, outNorms, "out Norm", t)
	CmprFloats(hidmoment, hidMoments, "hid Moment", t)
	CmprFloats(outmoment, outMoments, "out Moment", t)
	CmprFloats(hidwt, hidWts, "hid Wt", t)
	CmprFloats(outwt, outWts, "out Wt", t)

	var buf bytes.Buffer
	TestNet.WriteWtsJSON(&buf)
	wb := buf.Bytes()
	// fmt.Printf("TestNet Trained Weights:\n\n%v\n", string(wb))

	fp, err := os.Create("testdata/testnet_train.wts")
	defer fp.Close()
	if err != nil {
		t.Error(err)
	}
	fp.Write(wb)
}
