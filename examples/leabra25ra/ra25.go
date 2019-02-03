// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// ra25 runs a simple random-associator 5x5 = 25 four-layer leabra network
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/emer/emergent/basic/leabra"
	"github.com/emer/emergent/dtable"
	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/etensor"
	"github.com/emer/emergent/prjn"
)

var Net *leabra.Network
var Pats *dtable.Table
var EpcLog *dtable.Table
var Thread = false // much slower for small net

var Pars = emer.ParamStyle{
	"Prjn": {
		"Prjn.Learn.Norm.On":     1,
		"Prjn.Learn.Momentum.On": 1,
		"Prjn.Learn.WtBal.On":    0,
	},
	"Layer": {
		"Layer.Inhib.Layer.Gi": 1.8,
	},
	"#Output": {
		"Layer.Inhib.Layer.Gi": 1.4, // this turns out to be critical for small output layer
	},
	".TopDown": {
		"Prjn.WtScale.Rel": 0.2, // this is generally quite important
	},
}

func ConfigNet(net *leabra.Network) {
	net.Name = "RA25"
	inLay := net.AddLayer("Input", []int{5, 5}, leabra.Input)
	hid1Lay := net.AddLayer("Hidden1", []int{7, 7}, leabra.Hidden)
	hid2Lay := net.AddLayer("Hidden2", []int{7, 7}, leabra.Hidden)
	outLay := net.AddLayer("Output", []int{5, 5}, leabra.Target)

	net.ConnectLayers(inLay, hid1Lay, prjn.NewFull())
	net.ConnectLayers(hid1Lay, hid2Lay, prjn.NewFull())
	net.ConnectLayers(hid2Lay, outLay, prjn.NewFull())

	outHid2 := net.ConnectLayers(outLay, hid2Lay, prjn.NewFull())
	outHid2.Class = "TopDown"
	hid2Hid1 := net.ConnectLayers(hid2Lay, hid1Lay, prjn.NewFull())
	hid2Hid1.Class = "TopDown"

	if Thread {
		hid2Lay.Thread = 1
		outLay.Thread = 1
	}

	net.Defaults()
	net.StyleParams(Pars)
	net.Build()
	net.InitWts()
}

func ConfigPats(dt *dtable.Table) {
	dt.SetFromSchema(dtable.Schema{
		{"Name", etensor.STRING, nil, nil},
		{"Input", etensor.FLOAT32, []int{5, 5}, []string{"Y", "X"}},
		{"Output", etensor.FLOAT32, []int{5, 5}, []string{"Y", "X"}},
	}, 25)

	// todo: write code to generate the patterns using bit flip logic..
}

func OpenPats(dt *dtable.Table) {
	err := dt.OpenCSV("random_5x5_25.dat", '\t')
	if err != nil {
		log.Println(err)
	}
}

func ConfigEpcLog(dt *dtable.Table) {
	dt.SetFromSchema(dtable.Schema{
		{"Epoch", etensor.INT64, nil, nil},
		{"CosDiff", etensor.FLOAT32, nil, nil},
		{"AvgCosDiff", etensor.FLOAT32, nil, nil},
		{"SSE", etensor.FLOAT32, nil, nil},
		{"Avg SSE", etensor.FLOAT32, nil, nil},
		{"Count Err", etensor.FLOAT32, nil, nil},
		{"Pct Err", etensor.FLOAT32, nil, nil},
		{"Pct Cor", etensor.FLOAT32, nil, nil},
		{"Hid1 ActAvg", etensor.FLOAT32, nil, nil},
		{"Hid2 ActAvg", etensor.FLOAT32, nil, nil},
		{"Out ActAvg", etensor.FLOAT32, nil, nil},
	}, 0)
}

func TrainNet(net *leabra.Network, pats, epcLog *dtable.Table, epcs int) {
	ltime := leabra.NewTime()
	net.InitWts()
	np := pats.NumRows()

	epcLog.SetNumRows(epcs)

	inLay := net.LayerByName("Input").(*leabra.Layer)
	hid1Lay := net.LayerByName("Hidden1").(*leabra.Layer)
	hid2Lay := net.LayerByName("Hidden2").(*leabra.Layer)
	outLay := net.LayerByName("Output").(*leabra.Layer)

	_ = hid1Lay
	_ = hid2Lay

	inPats := pats.ColByName("Input").(*etensor.Float32)
	outPats := pats.ColByName("Output").(*etensor.Float32)

	stts := time.Now()
	for epc := 0; epc < epcs; epc++ {
		// todo: shuffle order
		outCosDiff := float32(0)
		cntErr := 0
		sse := float32(0)
		avgSSE := float32(0)
		for pi := 0; pi < np; pi++ {
			inp, _ := inPats.SubSlice(2, []int{pi})
			outp, _ := outPats.SubSlice(2, []int{pi})

			inLay.ApplyExt(inp)
			outLay.ApplyExt(outp)

			net.TrialInit()
			ltime.TrialStart()
			for qtr := 0; qtr < 4; qtr++ {
				for cyc := 0; cyc < ltime.CycPerQtr; cyc++ {
					net.Cycle()
					ltime.CycleInc()
				}
				net.QuarterFinal(ltime)
				ltime.QuarterInc()
			}
			net.DWt()
			net.WtFmDWt()
			outCosDiff += outLay.CosDiff.Cos
			pSSE, pAvgSSE := outLay.SSE(0.5)
			sse += pSSE
			avgSSE += pAvgSSE
			if pSSE != 0 {
				cntErr++
			}
		}
		outCosDiff /= float32(np)
		sse /= float32(np)
		avgSSE /= float32(np)
		pctErr := float32(cntErr) / float32(np)
		pctCor := 1 - pctErr
		// fmt.Printf("epc: %v  \tCosDiff: %v \tAvgCosDif: %v\n", epc, outCosDiff, outLay.CosDiff.Avg)
		epcLog.ColByName("Epoch").SetFlatFloat64(epc, float64(epc))
		epcLog.ColByName("CosDiff").SetFlatFloat64(epc, float64(outCosDiff))
		epcLog.ColByName("AvgCosDiff").SetFlatFloat64(epc, float64(outLay.CosDiff.Avg))
		epcLog.ColByName("SSE").SetFlatFloat64(epc, float64(sse))
		epcLog.ColByName("Avg SSE").SetFlatFloat64(epc, float64(avgSSE))
		epcLog.ColByName("Count Err").SetFlatFloat64(epc, float64(cntErr))
		epcLog.ColByName("Pct Err").SetFlatFloat64(epc, float64(pctErr))
		epcLog.ColByName("Pct Cor").SetFlatFloat64(epc, float64(pctCor))
		epcLog.ColByName("Hid1 ActAvg").SetFlatFloat64(epc, float64(hid1Lay.Pools[0].ActAvg.ActPAvgEff))
		epcLog.ColByName("Hid2 ActAvg").SetFlatFloat64(epc, float64(hid2Lay.Pools[0].ActAvg.ActPAvgEff))
		epcLog.ColByName("Out ActAvg").SetFlatFloat64(epc, float64(outLay.Pools[0].ActAvg.ActPAvgEff))
	}
	etts := time.Now()
	secs := float64(etts.Sub(stts)) / float64(time.Second)
	fmt.Printf("Took %v secs for %v epochs\n", secs, epcs)
}

func main() {
	Net = &leabra.Network{}
	ConfigNet(Net)
	Net.SaveWtsJSON("ra25_net_init.wts")

	Pats = &dtable.Table{}
	OpenPats(Pats)

	EpcLog = &dtable.Table{}
	ConfigEpcLog(EpcLog)

	TrainNet(Net, Pats, EpcLog, 100)
	Net.SaveWtsJSON("ra25_net_trained.wts")

	EpcLog.SaveCSV("ra25_epc.dat", ',', true)
}
