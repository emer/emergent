// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// bench runs a benchmark model with 5 layers (3 hidden, Input, Output) all of the same
// size, for benchmarking different size networks.  These are not particularly realistic
// models for actual applications (e.g., large models tend to have much more topographic
// patterns of connectivity and larger layers with fewer connections), but they are
// easy to run..
package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"

	"github.com/emer/emergent/dtable"
	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/erand"
	"github.com/emer/emergent/etensor"
	"github.com/emer/emergent/leabra/leabra"
	"github.com/emer/emergent/patgen"
	"github.com/emer/emergent/prjn"
	"github.com/emer/emergent/timer"
)

var Net *leabra.Network
var Pats *dtable.Table
var EpcLog *dtable.Table
var Thread = false // much slower for small net

var Pars = emer.ParamStyle{
	"Prjn": {
		"Prjn.Learn.Norm.On":     1,
		"Prjn.Learn.Momentum.On": 1,
		"Prjn.Learn.WtBal.On":    1,
	},
	".Back": {
		"Prjn.WtScale.Rel": 0.2, // this is generally quite important
	},
}

func ConfigNet(net *leabra.Network, threads, units int) {
	net.InitName(net, "BenchNet")

	squn := int(math.Sqrt(float64(units)))
	shp := []int{squn, squn}

	inLay := net.AddLayer("Input", shp, emer.Input)
	hid1Lay := net.AddLayer("Hidden1", shp, emer.Hidden)
	hid2Lay := net.AddLayer("Hidden2", shp, emer.Hidden)
	hid3Lay := net.AddLayer("Hidden3", shp, emer.Hidden)
	outLay := net.AddLayer("Output", shp, emer.Target)

	net.ConnectLayers(inLay, hid1Lay, prjn.NewFull(), emer.Forward)
	net.ConnectLayers(hid1Lay, hid2Lay, prjn.NewFull(), emer.Forward)
	net.ConnectLayers(hid2Lay, hid3Lay, prjn.NewFull(), emer.Forward)
	net.ConnectLayers(hid3Lay, outLay, prjn.NewFull(), emer.Forward)

	net.ConnectLayers(outLay, hid3Lay, prjn.NewFull(), emer.Back)
	net.ConnectLayers(hid3Lay, hid2Lay, prjn.NewFull(), emer.Back)
	net.ConnectLayers(hid2Lay, hid1Lay, prjn.NewFull(), emer.Back)

	switch threads {
	case 2:
		hid3Lay.SetThread(1)
		outLay.SetThread(1)
	case 4:
		hid2Lay.SetThread(1)
		hid3Lay.SetThread(2)
		outLay.SetThread(3)
	}

	net.Defaults()
	net.StyleParams(Pars, false) // no msg
	net.Build()
	net.InitWts()
}

func ConfigPats(dt *dtable.Table, pats, units int) {
	squn := int(math.Sqrt(float64(units)))
	shp := []int{squn, squn}
	fmt.Printf("shape: %v\n", shp)

	dt.SetFromSchema(dtable.Schema{
		{"Name", etensor.STRING, nil, nil},
		{"Input", etensor.FLOAT32, shp, []string{"Y", "X"}},
		{"Output", etensor.FLOAT32, shp, []string{"Y", "X"}},
	}, pats)

	// note: actually can learn if activity is .15 instead of .25
	// but C++ benchmark is for .25..
	nOn := units / 6

	patgen.PermutedBinaryRows(dt.Cols[1], nOn, 1, 0)
	patgen.PermutedBinaryRows(dt.Cols[2], nOn, 1, 0)
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
	porder := rand.Perm(np) // randomly permuted order of ints

	epcLog.SetNumRows(epcs)

	inLay := net.LayerByName("Input").(*leabra.Layer)
	hid1Lay := net.LayerByName("Hidden1").(*leabra.Layer)
	hid2Lay := net.LayerByName("Hidden2").(*leabra.Layer)
	outLay := net.LayerByName("Output").(*leabra.Layer)

	_ = hid1Lay
	_ = hid2Lay

	inPats := pats.ColByName("Input").(*etensor.Float32)
	outPats := pats.ColByName("Output").(*etensor.Float32)

	tmr := timer.Time{}
	tmr.Start()
	for epc := 0; epc < epcs; epc++ {
		erand.PermuteInts(porder)
		outCosDiff := float32(0)
		cntErr := 0
		sse := float32(0)
		avgSSE := float32(0)
		for pi := 0; pi < np; pi++ {
			ppi := porder[pi]
			inp, _ := inPats.SubSlice(2, []int{ppi})
			outp, _ := outPats.SubSlice(2, []int{ppi})

			inLay.ApplyExt(inp)
			outLay.ApplyExt(outp)

			net.TrialInit()
			ltime.TrialStart()
			for qtr := 0; qtr < 4; qtr++ {
				for cyc := 0; cyc < ltime.CycPerQtr; cyc++ {
					net.Cycle(ltime)
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
		epcLog.ColByName("Epoch").SetFloat1D(epc, float64(epc))
		epcLog.ColByName("CosDiff").SetFloat1D(epc, float64(outCosDiff))
		epcLog.ColByName("AvgCosDiff").SetFloat1D(epc, float64(outLay.CosDiff.Avg))
		epcLog.ColByName("SSE").SetFloat1D(epc, float64(sse))
		epcLog.ColByName("Avg SSE").SetFloat1D(epc, float64(avgSSE))
		epcLog.ColByName("Count Err").SetFloat1D(epc, float64(cntErr))
		epcLog.ColByName("Pct Err").SetFloat1D(epc, float64(pctErr))
		epcLog.ColByName("Pct Cor").SetFloat1D(epc, float64(pctCor))
		epcLog.ColByName("Hid1 ActAvg").SetFloat1D(epc, float64(hid1Lay.Pools[0].ActAvg.ActPAvgEff))
		epcLog.ColByName("Hid2 ActAvg").SetFloat1D(epc, float64(hid2Lay.Pools[0].ActAvg.ActPAvgEff))
		epcLog.ColByName("Out ActAvg").SetFloat1D(epc, float64(outLay.Pools[0].ActAvg.ActPAvgEff))
	}
	tmr.Stop()
	fmt.Printf("Took %6.4g secs for %v epochs, avg per epc: %6.4g\n", tmr.TotalSecs(), epcs, tmr.TotalSecs()/float64(epcs))
	net.TimerReport()
}

func main() {
	var threads int
	var epochs int
	var pats int
	var units int

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	// process command args
	flag.IntVar(&threads, "threads", 1, "number of threads (goroutines) to use")
	flag.IntVar(&epochs, "epochs", 2, "number of epochs to run")
	flag.IntVar(&pats, "pats", 10, "number of patterns per epoch")
	flag.IntVar(&units, "units", 100, "number of units per layer -- uses NxN where N = sqrt(units)")
	flag.Parse()

	fmt.Printf("Running bench with: %v threads, %v epochs, %v pats, %v units\n", threads, epochs, pats, units)

	Net = &leabra.Network{}
	ConfigNet(Net, threads, units)

	Pats = &dtable.Table{}
	ConfigPats(Pats, pats, units)

	EpcLog = &dtable.Table{}
	ConfigEpcLog(EpcLog)

	TrainNet(Net, Pats, EpcLog, epochs)

	EpcLog.SaveCSV("bench_epc.dat", ',', true)
}
