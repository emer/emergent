// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// labra25ra runs a simple random-associator 5x5 = 25 four-layer leabra network
package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/emer/emergent/dtable"
	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/eplot"
	"github.com/emer/emergent/erand"
	"github.com/emer/emergent/etensor"
	"github.com/emer/emergent/leabra/leabra"
	"github.com/emer/emergent/patgen"
	"github.com/emer/emergent/prjn"
	"github.com/emer/emergent/timer"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/gimain"
	"github.com/goki/gi/giv"
	"github.com/goki/gi/svg"
	"github.com/goki/gi/units"
	"github.com/goki/ki/ki"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// this is the stub main for gogi that calls our actual mainrun function, at end of file
func main() {
	gimain.Main(func() {
		mainrun()
	})
}

// DefaultPars are the initial default parameters for this simulation
var DefaultPars = emer.ParamStyle{
	"Prjn": {
		"Prjn.Learn.Norm.On":     1,
		"Prjn.Learn.Momentum.On": 1,
		"Prjn.Learn.WtBal.On":    0,
	},
	// "Layer": {
	// 	"Layer.Inhib.Layer.Gi": 1.8, // this is the default
	// },
	"#Output": {
		"Layer.Inhib.Layer.Gi": 1.4, // this turns out to be critical for small output layer
	},
	".Back": {
		"Prjn.WtScale.Rel": 0.2, // this is generally quite important
	},
}

// these are the plot color names to use in order for successive lines -- feel free to choose your own!
var PlotColorNames = []string{"black", "red", "blue", "ForestGreen", "purple", "orange", "brown", "chartreuse", "navy", "cyan", "magenta", "tan", "salmon", "yellow4", "SkyBlue", "pink"}

// SimState maintains everything about this simulation, and we define all the
// functionality as methods on this type -- this makes it easier to add additional
// state information as needed, and not have to worry about passing arguments around
// and it also makes it much easier to support interactive stepping of the model etc.
// This can be edited directly by the user to access any elements of the simulation.
type SimState struct {
	Net        *leabra.Network `view:"no-inline"`
	Pats       *dtable.Table   `view:"no-inline"`
	EpcLog     *dtable.Table   `view:"no-inline"`
	Pars       emer.ParamStyle `view:"no-inline"`
	MaxEpcs    int             `desc:"maximum number of epochs to run"`
	Epoch      int
	Trial      int
	Time       leabra.Time
	Plot       bool     `desc:"update the epoch plot while running?"`
	PlotVals   []string `desc:"values to plot in epoch plot"`
	Sequential bool     `desc:"set to true to present items in sequential order"`
	Test       bool     `desc:"set to true to not call learning methods"`

	// statistics
	EpcSSE     float32 `inactive:"+" desc:"last epoch's total sum squared error"`
	EpcAvgSSE  float32 `inactive:"+" desc:"last epoch's average sum squared error (average over trials, and over units within layer)"`
	EpcPctErr  float32 `inactive:"+" desc:"last epoch's percent of trials that had SSE > 0 (subject to .5 unit-wise tolerance)"`
	EpcPctCor  float32 `inactive:"+" desc:"last epoch's percent of trials that had SSE == 0 (subject to .5 unit-wise tolerance)"`
	EpcCosDiff float32 `inactive:"+" desc:"last epoch's average cosine difference for output layer (a normalized error measure, maximum of 1 when the minus phase exactly matches the plus)"`

	// internal state - view:"-"
	SumSSE     float32     `view:"-" inactive:"+" desc:"sum to increment as we go through epoch"`
	SumAvgSSE  float32     `view:"-" inactive:"+" desc:"sum to increment as we go through epoch"`
	SumCosDiff float32     `view:"-" inactive:"+" desc:"sum to increment as we go through epoch"`
	CntErr     int         `view:"-" inactive:"+" desc:"sum of errs to increment as we go through epoch"`
	Porder     []int       `view:"-" inactive:"+" desc:"permuted pattern order"`
	EpcPlotSvg *svg.Editor `view:"-" desc:"the epoch plot svg editor"`
	StopNow    bool        `view:"-" desc:"flag to stop running"`
	RndSeed    int64       `view:"-" desc:"the current random seed"`
}

// Sim is the overall state for this simulation
var Sim SimState

// New creates new blank elements
func (ss *SimState) New() {
	ss.Net = &leabra.Network{}
	ss.Pats = &dtable.Table{}
	ss.EpcLog = &dtable.Table{}
	ss.Pars = DefaultPars
	ss.RndSeed = 1
}

// Config configures all the elements using the standard functions
func (ss *SimState) Config() {
	ss.ConfigNet()
	ss.OpenPats()
	ss.ConfigEpcLog()
}

// Init restarts the run, and initializes everything, including network weights
// and resets the epoch log table
func (ss *SimState) Init() {
	rand.Seed(ss.RndSeed)
	if ss.MaxEpcs == 0 { // allow user override
		ss.MaxEpcs = 50
	}
	ss.Epoch = 0
	ss.Trial = 0
	ss.StopNow = false
	ss.Time.Reset()
	np := ss.Pats.NumRows()
	ss.Porder = rand.Perm(np)         // always start with new one so random order is identical
	ss.Net.StyleParams(ss.Pars, true) // set msg
	ss.Net.InitWts()
	ss.EpcLog.SetNumRows(0)
}

// NewRndSeed gets a new random seed based on current time -- otherwise uses
// the same random seed for every run
func (ss *SimState) NewRndSeed() {
	ss.RndSeed = time.Now().UnixNano()
}

// RunTrial runs one alpha-trial (100 msec, 4 quarters)			 of processing
// this does NOT call TrialInc (so it can be used flexibly)
// but it does use the Trial counter to determine which pattern to present.
func (ss *SimState) RunTrial() {
	inLay := ss.Net.LayerByName("Input").(*leabra.Layer)
	outLay := ss.Net.LayerByName("Output").(*leabra.Layer)
	inPats := ss.Pats.ColByName("Input").(*etensor.Float32)
	outPats := ss.Pats.ColByName("Output").(*etensor.Float32)

	pidx := ss.Trial
	if !ss.Sequential {
		pidx = ss.Porder[ss.Trial]
	}

	inp, _ := inPats.SubSlice(2, []int{pidx})
	outp, _ := outPats.SubSlice(2, []int{pidx})
	inLay.ApplyExt(inp)
	outLay.ApplyExt(outp)

	ss.Net.TrialInit()
	ss.Time.TrialStart()
	for qtr := 0; qtr < 4; qtr++ {
		for cyc := 0; cyc < ss.Time.CycPerQtr; cyc++ {
			ss.Net.Cycle(&ss.Time)
			ss.Time.CycleInc()
		}
		ss.Net.QuarterFinal(&ss.Time)
		ss.Time.QuarterInc()
	}

	if !ss.Test {
		ss.Net.DWt()
		ss.Net.WtFmDWt()
	}
}

// TrialInc increments counters after one trial of processing
func (ss *SimState) TrialInc() {
	ss.Trial++
	np := ss.Pats.NumRows()
	if ss.Trial >= np {
		ss.LogEpoch()
		if ss.Plot {
			ss.PlotEpcLog()
		}
		ss.EpochInc()
	}
}

// TrialStats computes the trial-level statistics and adds them to the epoch accumulators if
// accum is true
func (ss *SimState) TrialStats(accum bool) (sse, avgsse, cosdiff float32) {
	outLay := ss.Net.LayerByName("Output").(*leabra.Layer)
	cosdiff = outLay.CosDiff.Cos
	sse, avgsse = outLay.MSE(0.5) // 0.5 = per-unit tolerance -- right side of .5
	if accum {
		ss.SumSSE += sse
		ss.SumAvgSSE += avgsse
		ss.SumCosDiff += cosdiff
		if sse != 0 {
			ss.CntErr++
		}
	}
	return
}

// EpochInc increments counters after one epoch of processing and updates a new random
// order of permuted inputs for the next epoch
func (ss *SimState) EpochInc() {
	ss.Trial = 0
	ss.Epoch++
	erand.PermuteInts(ss.Porder)
}

// LogEpoch adds data from current epoch to the EpochLog table -- computes epoch
// averages prior to logging.
// Epoch counter is assumed to not have yet been incremented.
func (ss *SimState) LogEpoch() {
	ss.EpcLog.SetNumRows(ss.Epoch + 1)
	hid1Lay := ss.Net.LayerByName("Hidden1").(*leabra.Layer)
	hid2Lay := ss.Net.LayerByName("Hidden2").(*leabra.Layer)
	outLay := ss.Net.LayerByName("Output").(*leabra.Layer)

	np := float32(ss.Pats.NumRows())
	ss.EpcSSE = ss.SumSSE / np
	ss.SumSSE = 0
	ss.EpcAvgSSE = ss.SumAvgSSE / np
	ss.SumAvgSSE = 0
	ss.EpcPctErr = float32(ss.CntErr) / np
	ss.CntErr = 0
	ss.EpcPctCor = 1 - ss.EpcPctErr
	ss.EpcCosDiff = ss.SumCosDiff / np
	ss.SumCosDiff = 0

	epc := ss.Epoch

	ss.EpcLog.ColByName("Epoch").SetFloat1D(epc, float64(epc))
	ss.EpcLog.ColByName("SSE").SetFloat1D(epc, float64(ss.EpcSSE))
	ss.EpcLog.ColByName("Avg SSE").SetFloat1D(epc, float64(ss.EpcAvgSSE))
	ss.EpcLog.ColByName("Pct Err").SetFloat1D(epc, float64(ss.EpcPctErr))
	ss.EpcLog.ColByName("Pct Cor").SetFloat1D(epc, float64(ss.EpcPctCor))
	ss.EpcLog.ColByName("CosDiff").SetFloat1D(epc, float64(ss.EpcCosDiff))
	ss.EpcLog.ColByName("Hid1 ActAvg").SetFloat1D(epc, float64(hid1Lay.Pools[0].ActAvg.ActPAvgEff))
	ss.EpcLog.ColByName("Hid2 ActAvg").SetFloat1D(epc, float64(hid2Lay.Pools[0].ActAvg.ActPAvgEff))
	ss.EpcLog.ColByName("Out ActAvg").SetFloat1D(epc, float64(outLay.Pools[0].ActAvg.ActPAvgEff))

}

// StepTrial does one alpha trial of processing and increments everything etc
// for interactive running.
func (ss *SimState) StepTrial() {
	ss.RunTrial()
	ss.TrialStats(!ss.Test) // accumulate if not doing testing
	ss.TrialInc()           // does LogEpoch, EpochInc automatically
}

// StepEpoch runs for remainder of this epoch
func (ss *SimState) StepEpoch() {
	curEpc := ss.Epoch
	for {
		ss.RunTrial()
		ss.TrialStats(!ss.Test) // accumulate if not doing testing
		ss.TrialInc()           // does LogEpoch, EpochInc automatically
		if ss.StopNow || ss.Epoch > curEpc {
			break
		}
	}
}

// Train runs the full training from this point onward
func (ss *SimState) Train() {
	ss.StopNow = false
	tmr := timer.Time{}
	tmr.Start()
	for {
		ss.StepTrial()
		if ss.StopNow || ss.Epoch >= ss.MaxEpcs {
			break
		}
	}
	tmr.Stop()
	epcs := ss.Epoch
	fmt.Printf("Took %6g secs for %v epochs, avg per epc: %6g\n", tmr.TotalSecs(), epcs, tmr.TotalSecs()/float64(epcs))
}

// Stop tells the sim to stop running
func (ss *SimState) Stop() {
	ss.StopNow = true
}

////////////////////////////////////////////////////////////////////////////////////////////
// Config methods

func (ss *SimState) ConfigNet() {
	net := ss.Net
	net.InitName(net, "RA25")
	inLay := net.AddLayer2D("Input", 5, 5, emer.Input)
	hid1Lay := net.AddLayer2D("Hidden1", 7, 7, emer.Hidden)
	hid2Lay := net.AddLayer2D("Hidden2", 7, 7, emer.Hidden)
	outLay := net.AddLayer2D("Output", 5, 5, emer.Target)

	net.ConnectLayers(inLay, hid1Lay, prjn.NewFull(), emer.Forward)
	net.ConnectLayers(hid1Lay, hid2Lay, prjn.NewFull(), emer.Forward)
	net.ConnectLayers(hid2Lay, outLay, prjn.NewFull(), emer.Forward)

	net.ConnectLayers(outLay, hid2Lay, prjn.NewFull(), emer.Back)
	net.ConnectLayers(hid2Lay, hid1Lay, prjn.NewFull(), emer.Back)

	// if Thread {
	// 	hid2Lay.SetThread(1)
	// 	outLay.SetThread(1)
	// }

	net.Defaults()
	net.StyleParams(ss.Pars, true) // set msg
	net.Build()
	net.InitWts()
}

func (ss *SimState) ConfigPats() {
	dt := ss.Pats
	dt.SetFromSchema(dtable.Schema{
		{"Name", etensor.STRING, nil, nil},
		{"Input", etensor.FLOAT32, []int{5, 5}, []string{"Y", "X"}},
		{"Output", etensor.FLOAT32, []int{5, 5}, []string{"Y", "X"}},
	}, 25)

	patgen.PermutedBinaryRows(dt.Cols[1], 6, 1, 0)
	patgen.PermutedBinaryRows(dt.Cols[2], 6, 1, 0)
	dt.SaveCSV("random_5x5_25_gen.dat", ',', true)
}

func (ss *SimState) OpenPats() {
	dt := ss.Pats
	err := dt.OpenCSV("random_5x5_25.dat", '\t')
	if err != nil {
		log.Println(err)
	}
}

func (ss *SimState) ConfigEpcLog() {
	dt := ss.EpcLog
	dt.SetFromSchema(dtable.Schema{
		{"Epoch", etensor.INT64, nil, nil},
		{"SSE", etensor.FLOAT32, nil, nil},
		{"Avg SSE", etensor.FLOAT32, nil, nil},
		{"Pct Err", etensor.FLOAT32, nil, nil},
		{"Pct Cor", etensor.FLOAT32, nil, nil},
		{"CosDiff", etensor.FLOAT32, nil, nil},
		{"Hid1 ActAvg", etensor.FLOAT32, nil, nil},
		{"Hid2 ActAvg", etensor.FLOAT32, nil, nil},
		{"Out ActAvg", etensor.FLOAT32, nil, nil},
	}, 0)
	ss.PlotVals = []string{"SSE", "Pct Err"}
	ss.Plot = true
}

// PlotEpcLog plots given epoch log using PlotVals Y axis columns into EpcPlotSvg
func (ss *SimState) PlotEpcLog() *plot.Plot {
	dt := ss.EpcLog
	plt, _ := plot.New() // todo: keep around?
	plt.Title.Text = "Random Associator Epoch Log"
	plt.X.Label.Text = "Epoch"
	plt.Y.Label.Text = "Y"

	const lineWidth = 1

	for i, cl := range ss.PlotVals {
		xy, _ := eplot.NewTableXYNames(dt, "Epoch", cl)
		l, _ := plotter.NewLine(xy)
		l.LineStyle.Width = vg.Points(lineWidth)
		clr, _ := gi.ColorFromString(PlotColorNames[i%len(PlotColorNames)], nil)
		l.LineStyle.Color = clr
		plt.Add(l)
		plt.Legend.Add(cl, l)
	}
	plt.Legend.Top = true
	eplot.PlotViewSVG(plt, ss.EpcPlotSvg, 5, 5, 2)
	return plt
}

// SaveEpcPlot plots given epoch log using PlotVals Y axis columns and saves to .svg file
func (ss *SimState) SaveEpcPlot(fname string) {
	plt := ss.PlotEpcLog()
	plt.Save(5, 5, fname)
}

// ConfigGui configures the GoGi gui interface for this simulation,
func (ss *SimState) ConfigGui() *gi.Window {
	width := 1600
	height := 1200

	gi.SetAppName("leabra25ra")
	gi.SetAppAbout(`This demonstrates a basic Leabra model. See <a href="https://github.com/emer/emergent">emergent on GitHub</a>.</p>`)

	plot.DefaultFont = "Helvetica"

	win := gi.NewWindow2D("leabra25ra", "Leabra Random Associator", width, height, true)

	vp := win.WinViewport2D()
	updt := vp.UpdateStart()

	mfr := win.SetMainFrame()

	tbar := gi.AddNewToolBar(mfr, "tbar")
	tbar.SetStretchMaxWidth()

	split := gi.AddNewSplitView(mfr, "split")
	split.Dim = gi.X
	// split.SetProp("horizontal-align", "center")
	// split.SetProp("margin", 2.0) // raw numbers = px = 96 dpi pixels
	split.SetStretchMaxWidth()
	split.SetStretchMaxHeight()

	sv := giv.AddNewStructView(split, "sv")
	sv.SetStruct(ss, nil)
	// sv.SetStretchMaxWidth()
	// sv.SetStretchMaxHeight()

	tv := gi.AddNewTabView(split, "tv")
	svge := tv.AddNewTab(svg.KiT_Editor, "Epc Plot").(*svg.Editor)
	svge.InitScale()
	svge.Fill = true
	svge.SetProp("background-color", "white")
	svge.SetProp("width", units.NewValue(float32(width/2), units.Px))
	svge.SetProp("height", units.NewValue(float32(height-100), units.Px))
	svge.SetStretchMaxWidth()
	svge.SetStretchMaxHeight()
	ss.EpcPlotSvg = svge

	split.SetSplits(.3, .7)

	tbar.AddAction(gi.ActOpts{Label: "Init", Icon: "update"}, win.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			ss.Init()
			vp.FullRender2DTree()
		})

	tbar.AddAction(gi.ActOpts{Label: "Train", Icon: "run"}, win.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			go ss.Train()
		})

	tbar.AddAction(gi.ActOpts{Label: "Stop", Icon: "stop"}, win.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			ss.Stop()
			vp.FullRender2DTree()
		})

	tbar.AddAction(gi.ActOpts{Label: "Step Trial", Icon: "step-fwd"}, win.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			ss.StepTrial()
			vp.FullRender2DTree()
		})

	tbar.AddAction(gi.ActOpts{Label: "Step Epoch", Icon: "fast-fwd"}, win.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			ss.StepEpoch()
			vp.FullRender2DTree()
		})

	// tbar.AddSep("file")

	tbar.AddAction(gi.ActOpts{Label: "Epoch Plot", Icon: "update"}, win.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			ss.PlotEpcLog()
		})

	tbar.AddAction(gi.ActOpts{Label: "Save Wts", Icon: "file-save"}, win.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			ss.Net.SaveWtsJSON("ra25_net_trained.wts") // todo: call method to prompt
		})

	tbar.AddAction(gi.ActOpts{Label: "Save Log", Icon: "file-save"}, win.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			ss.EpcLog.SaveCSV("ra25_epc.dat", ',', true)
		})

	tbar.AddAction(gi.ActOpts{Label: "Save Plot", Icon: "file-save"}, win.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			ss.SaveEpcPlot("ra25_cur_epc_plot.svg")
		})

	tbar.AddAction(gi.ActOpts{Label: "Save Pars", Icon: "file-save"}, win.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			// todo: need save / load methods for these
			// ss.EpcLog.SaveCSV("ra25_epc.dat", ',', true)
		})

	tbar.AddAction(gi.ActOpts{Label: "New Seed", Icon: "new"}, win.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			ss.NewRndSeed()
		})

	vp.UpdateEndNoSig(updt)

	// main menu
	appnm := gi.AppName()
	mmen := win.MainMenu
	mmen.ConfigMenus([]string{appnm, "File", "Edit", "Window"})

	amen := win.MainMenu.ChildByName(appnm, 0).(*gi.Action)
	amen.Menu.AddAppMenu(win)

	emen := win.MainMenu.ChildByName("Edit", 1).(*gi.Action)
	emen.Menu.AddCopyCutPaste(win)

	// note: Command in shortcuts is automatically translated into Control for
	// Linux, Windows or Meta for MacOS
	// fmen := win.MainMenu.ChildByName("File", 0).(*gi.Action)
	// fmen.Menu.AddAction(gi.ActOpts{Label: "Open", Shortcut: "Command+O"},
	// 	win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
	// 		FileViewOpenSVG(vp)
	// 	})
	// fmen.Menu.AddSeparator("csep")
	// fmen.Menu.AddAction(gi.ActOpts{Label: "Close Window", Shortcut: "Command+W"},
	// 	win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
	// 		win.Close()
	// 	})

	win.SetCloseCleanFunc(func(w *gi.Window) {
		go gi.Quit() // once main window is closed, quit
	})

	win.MainMenuUpdated()
	return win
}

func mainrun() {

	// todo: args
	Sim.New()
	Sim.Config()
	Sim.Init()
	win := Sim.ConfigGui()
	win.StartEventLoop()
}
