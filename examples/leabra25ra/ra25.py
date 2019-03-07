# Copyright (c) 2019, The Emergent Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# to run this python version of the demo:
# * install gopy, currently in fork at https://github.com/goki/gopy
#   e.g., 'go get github.com/goki/gopy -u ./...' and then cd to that package
#   and do 'go install'
# * go to the python directory in this emergent repository, read README.md there, and 
#   type 'make' -- if that works, then type make install (may need sudo)
# * cd back here, and run 'pyemergent' which was installed into /usr/local/bin
# * then type 'import ra25' and this should run
# * you'll need various standard packages such as pandas, numpy, matplotlib, etc

# labra25ra runs a simple random-associator 5x5 = 25 four-layer leabra network

from emergent import go, leabra, emer, eplot, patgen, prjn, dtable, etensor, rand, erand, gi, giv, svg

# this is in-process and will be an installable module under GoGi later
import pygiv

import importlib as il  #il.reload(ra25) -- doesn't seem to work for reasons unknown
import numpy as np
import matplotlib
matplotlib.use('SVG')
import matplotlib.pyplot as plt
plt.rcParams['svg.fonttype'] = 'none'  # essential for not rendering fonts as paths
import io

# note: xarray or pytorch TensorDataSet can be used instead of pandas for input / output
# patterns and recording of "log" data for plotting
import pandas as pd

# this will become SimState later.. 
Sim = 1

# note: cannot use method callbacks -- must be separate functions
def InitCB(recv, send, sig, data):
    Sim.Init()
    Sim.ClassView.Update()
    Sim.vp.FullRender2DTree()

def TrainCB(recv, send, sig, data):
    Sim.Train()
    Sim.ClassView.Update()
    Sim.vp.FullRender2DTree()

def StopCB(recv, send, sig, data):
    Sim.Stop()
    Sim.vp.FullRender2DTree()

def StepTrialCB(recv, send, sig, data):
    Sim.StepTrial()
    Sim.ClassView.Update()
    Sim.vp.FullRender2DTree()

def StepEpochCB(recv, send, sig, data):
    Sim.StepEpoch()
    Sim.ClassView.Update()
    Sim.vp.FullRender2DTree()

def PlotEpcLogCB(recv, send, sig, data):
    Sim.PlotEpcLog()
    Sim.ClassView.Update()
    Sim.vp.FullRender2DTree()

def SaveEpcPlotCB(recv, send, sig, data):
    Sim.SaveEpcPlot()

def SaveWtsCB(recv, send, sig, data):
    Sim.Net.SaveWtsJSON("ra25_net_trained.wts") # todo: call method to prompt
   
def SaveLogCB(recv, send, sig, data):
    Sim.EpcLog.to_csv("ra25_epc.dat", sep="\t")

def SaveParsCB(recv, send, sig, data):
    print("save params todo")

def NewRndSeedCB(recv, send, sig, data):
    Sim.NewRndSeed()

# DefaultPars are the initial default parameters for this simulation
DefaultPars = emer.ParamStyle({
    "Prjn": emer.Params({
        "Prjn.Learn.Norm.On":     1,
        "Prjn.Learn.Momentum.On": 1,
        "Prjn.Learn.WtBal.On":    0,
        }).handle,
    "#Output": emer.Params({
        "Layer.Inhib.Layer.Gi": 1.4, # this turns out to be critical for small output layer
    }).handle,
    ".Back": emer.Params({
        "Prjn.WtScale.Rel": 0.2, # this is generally quite important
    }).handle,
})

class SimState(object):
    """
    SimState maintains everything about this simulation, and we define all the
    functionality as methods on this type -- this makes it easier to add additional
    state information as needed, and not have to worry about passing arguments around
    and it also makes it much easier to support interactive stepping of the model etc.
    This can be edited directly by the user to access any elements of the simulation.
    """
    def __init__(self):
        self.Net = leabra.Network()
        self.Pats     = pd.DataFrame()
        self.EpcLog   = pd.DataFrame()
        self.Pars     = DefaultPars
        self.MaxEpcs  =  100
        self.Epoch    = 0
        self.Trial    = 0
        self.Time     = leabra.Time()
        self.Plot     = True
        self.PlotVals  = ["SSE", "Pct Err"]
        self.Sequential = False
        self.Test      = False
        
        # statistics
        self.EpcSSE     = 0.0
        self.EpcAvgSSE  = 0.0
        self.EpcPctErr  = 0.0
        self.EpcPctCor  = 0.0
        self.EpcCosDiff = 0.0
        
        # internal state - view:"-"
        self.SumSSE     = 0.0
        self.SumAvgSSE  = 0.0
        self.SumCosDiff = 0.0
        self.CntErr     = 0.0
        self.Porder     = [1]
        self.EpcPlotSvg = svg.Editor()
        self.StopNow    = False
        self.RndSeed    = 0
        
        # ClassView tags for controlling display of fields
        self.Tags = {
            'EpcSSE': 'inactive:"+"',
            'EpcAvgSSE': 'inactive:"+"',
            'EpcPctErr': 'inactive:"+"',
            'EpcPctCor': 'inactive:"+"',
            'EpcCosDiff': 'inactive:"+"',
            'SumSSE': 'view:"-"',
            'SumAvgSSE': 'view:"-"',
            'SumCosDiff': 'view:"-"',
            'CntErr': 'view:"-"',
            'Porder': 'view:"-"',
            'EpcPlotSvg': 'view:"-"',
            'StopNow': 'view:"-"',
            'RndSeed': 'view:"-"',
            'win': 'view:"-"',
            'vp': 'view:"-"',
            'ClassView': 'view:"-"',
            'Tags': 'view:"-"',
        }


    def Config(self):
        """Config configures all the elements using the standard functions"""
        self.ConfigNet()
        self.OpenPats()
        self.ConfigEpcLog()

    def Init(self):
        """Init restarts the run, and initializes everything, including network weights and resets the epoch log table"""
        rand.Seed(self.RndSeed)
        if self.MaxEpcs == 0: # allow user override
            self.MaxEpcs = 100
        self.Epoch = 0
        self.Trial = 0
        self.StopNow = False
        self.Time.Reset()
        npat = len(self.Pats.index)
        self.Porder = np.random.permutation(npat)         # always start with new one so random order is identical
        self.Net.StyleParams(self.Pars, True) # set msg
        self.Net.InitWts()
        self.EpcLog = self.EpcLog.iloc[0:] # keep columns, delete rows

    def NewRndSeed(self):
        """NewRndSeed gets a new random seed based on current time -- otherwise uses the same random seed for every run"""
        # self.RndSeed = time.Now().UnixNano()

    def ApplyExt(self, lay, nparray):
        """ApplyExt applies external input to given layer from given numpy array source"""
        flt = np.ndarray.flatten(nparray, 'C')
        # print(len(flt))
        # print(flt)
        slc = go.Slice_float32(flt)
        # print(len(slc))
        # for i in slc:
        #     print(i)
        lay.ApplyExt1D(slc)
    
    def RunTrial(self):
        """RunTrial runs one alpha-trial (100 msec, 4 quarters) of processing
        this does NOT call TrialInc (so it can be used flexibly)
        but it does use the Trial counter to determine which pattern to present."""
    
        inLay = leabra.Layer(self.Net.LayerByName("Input"))
        outLay = leabra.Layer(self.Net.LayerByName("Output"))
        # inPats = etensor.Float32(self.Pats.ColByName("Input"))
        # outPats = etensor.Float32(self.Pats.ColByName("Output"))
        
        pidx = self.Trial
        if not self.Sequential:
            pidx = self.Porder[self.Trial]

        # note: these indexes must be updated based on columns in patterns..
        inp = self.Pats.iloc[pidx,1:26].values
        outp = self.Pats.iloc[pidx,26:26+25].values

        self.ApplyExt(inLay, inp)
        self.ApplyExt(outLay, outp)
        
        self.Net.TrialInit()
        self.Time.TrialStart()
        for qtr in range(4):
            for cyc in range(self.Time.CycPerQtr):
                self.Net.Cycle(self.Time)
                self.Time.CycleInc()
            self.Net.QuarterFinal(self.Time)
            self.Time.QuarterInc()

        
        if not self.Test:
            self.Net.DWt()
            self.Net.WtFmDWt()

    def TrialInc(self):
        """TrialInc increments counters after one trial of processing"""
        self.Trial += 1
        np = len(self.Pats.index)
        if self.Trial >= np:
            self.LogEpoch()
            if self.Plot:
                self.PlotEpcLog()
            self.EpochInc()

    def TrialStats(self, accum):
        """TrialStats computes the trial-level statistics and adds them to the
        epoch accumulators if accum is True"""
        outLay = leabra.Layer(self.Net.LayerByName("Output"))
        cosdiff = outLay.CosDiff.Cos
        sse = outLay.SSE(0.5) # 0.5 = per-unit tolerance -- right side of .5
        if accum:
            self.SumSSE += sse
            self.SumAvgSSE += sse # not accurate
            self.SumCosDiff += cosdiff
            if sse != 0:
                self.CntErr += 1.0

    def EpochInc(self):
        """EpochInc increments counters after one epoch of processing and updates a new random
        order of permuted inputs for the next epoch"""
        self.Trial = 0
        self.Epoch += 1
        self.Porder = np.random.permutation(self.Porder)

    def LogEpoch(self):
        """LogEpoch adds data from current epoch to the EpochLog table -- computes epoch
        averages prior to logging.
        Epoch counter is assumed to not have yet been incremented."""
        hid1Lay = leabra.Layer(self.Net.LayerByName("Hidden1"))
        hid2Lay = leabra.Layer(self.Net.LayerByName("Hidden2"))
        outLay = leabra.Layer(self.Net.LayerByName("Output"))
        
        npat = len(self.Pats.index)
        self.EpcSSE = self.SumSSE / npat
        self.SumSSE = 0.0
        self.EpcAvgSSE = self.SumAvgSSE / npat
        self.SumAvgSSE = 0.0
        self.EpcPctErr = self.CntErr / npat
        self.CntErr = 0.0
        self.EpcPctCor = 1.0 - self.EpcPctErr
        self.EpcCosDiff = self.SumCosDiff / npat
        self.SumCosDiff = 0.0
        
        epc = self.Epoch
        
        nwdat = [epc, self.EpcSSE, self.EpcAvgSSE, self.EpcPctErr, self.EpcPctCor, self.EpcCosDiff, 0, 0, 0]
        # self.EpcLog.ColByName("Hid1 ActAvg").SetFloat1D(epc, hid1Lay.Pools[0].ActAvg.ActPAvgEff)
        # self.EpcLog.ColByName("Hid2 ActAvg").SetFloat1D(epc, hid2Lay.Pools[0].ActAvg.ActPAvgEff)
        # self.EpcLog.ColByName("Out ActAvg").SetFloat1D(epc, outLay.Pools[0].ActAvg.ActPAvgEff)

        nrow = len(self.EpcLog.index)
        # note: this is reportedly rather slow
        self.EpcLog.loc[nrow] = nwdat

    def StepTrial(self):
        """StepTrial does one alpha trial of processing and increments everything etc
        for interactive running."""
        self.RunTrial()
        self.TrialStats(not self.Test) # accumulate if not doing testing
        self.TrialInc()           # does LogEpoch, EpochInc automatically

    def StepEpoch(self):
        """StepEpoch runs for remainder of this epoch"""
        curEpc = self.Epoch
        while True:
            self.RunTrial()
            self.TrialStats(not self.Test) # accumulate if not doing testing
            self.TrialInc()           # does LogEpoch, EpochInc automatically
            if self.StopNow or self.Epoch > curEpc:
                break
        self.ClassView.Update()

    def Train(self):
        """Train runs the full training from this point onward"""
        self.StopNow = False
        # tmr = timer.Time{}
        # tmr.Start()
        while True:
            self.StepTrial()
            if self.StopNow or self.Epoch >= self.MaxEpcs:
                break
        # tmr.Stop()
        epcs = self.Epoch
        # print("Took %6g secs for %v epochs, avg per epc: %6g\n", tmr.TotalSecs(), epcs, tmr.TotalSecs()/float64(epcs))

    def Stop(self):
        """Stop tells the sim to stop running"""
        self.StopNow = True

    # --------- Config methods -----------

    def ConfigNet(self):
        net = self.Net
        net.InitName(net, "RA25")
        inLay = net.AddLayer2D("Input", 5, 5, emer.Input)
        hid1Lay = net.AddLayer2D("Hidden1", 7, 7, emer.Hidden)
        hid2Lay = net.AddLayer2D("Hidden2", 7, 7, emer.Hidden)
        outLay = net.AddLayer2D("Output", 5, 5, emer.Target)
        
        net.ConnectLayers(inLay, hid1Lay, prjn.NewFull(), emer.Forward)
        net.ConnectLayers(hid1Lay, hid2Lay, prjn.NewFull(), emer.Forward)
        net.ConnectLayers(hid2Lay, outLay, prjn.NewFull(), emer.Forward)
        
        net.ConnectLayers(outLay, hid2Lay, prjn.NewFull(), emer.Back)
        net.ConnectLayers(hid2Lay, hid1Lay, prjn.NewFull(), emer.Back)
        
        net.Defaults()
        # net.StyleParams(self.Pars, True) # set msg
        net.Build()
        net.InitWts()

    def ConfigPats(self):
        # note: this is all go-based for using dtable.Table instead of pandas
        dt = self.Pats
        schema = dtable.Schema()
        schema.append(dtable.Column("Name", etensor.STRING, nil, nil))
        schema.append(dtable.Column("Input", etensor.FLOAT32, go.Slice_int([5, 5]), go.Slice_string(["Y", "X"])))
        schema.append(dtable.Column("Output", etensor.FLOAT32, go.Slice_int([5, 5]), go.Slice_string(["Y", "X"])))
        dt.SetFromSchema(schema, 25)
            
        patgen.PermutedBinaryRows(dt.Cols[1], 6, 1, 0)
        patgen.PermutedBinaryRows(dt.Cols[2], 6, 1, 0)
        dt.SaveCSV("random_5x5_25_gen.dat", ',', True)

    def OpenPats(self):
        dt = pd.read_csv("random_5x5_25.dat", sep='\t')
        dt = dt.drop(columns="_H:")
        self.Pats = dt

    def ConfigEpcLog(self):
        self.EpcLog = pd.DataFrame(columns=["Epoch", "SSE", "Avg SSE", "Pct Err", "Pct Cor", "CosDiff", "Hid1 ActAvg", "Hid2 ActAvg", "Out ActAvg"])
        self.PlotVals = ["SSE", "Pct Err"]
        self.Plot = True

    def PlotEpcLog(self):
        """PlotEpcLog plots given epoch log using PlotVals Y axis columns into EpcPlotSvg"""
        dt = self.EpcLog
        epc = dt['Epoch'].values
        for cl in self.PlotVals:
            yv = dt[cl].values
            plt.plot(epc, yv, label=cl)

        plt.xlabel("Epoch")
        plt.ylabel("Values")
        plt.legend()
        plt.title("Random Associator Epoch Log")
        
        svgstr = io.StringIO()
        plt.savefig(svgstr, format='svg')
        eplot.StringViewSVG(svgstr.getvalue(), self.EpcPlotSvg, 2)
        plt.close()
        return svgstr

    def SaveEpcPlot(self, fname):
        """SaveEpcLog plots given epoch log using PlotVals Y axis columns into .svg file"""
        svgstr = self.PlotEpcLog()
        f = open(fname,"w+")
        f.write(svgstr.getvalue())
        f.close()

    def ConfigGui(self):
        """ConfigGui configures the GoGi gui interface for this simulation"""
        width = 1600
        height = 1200
        
        gi.SetAppName("leabra25ra")
        gi.SetAppAbout('This demonstrates a basic Leabra model. See <a href="https:#github.com/emer/emergent">emergent on GitHub</a>.</p>')
        
        win = gi.NewWindow2D("leabra25ra", "Leabra Random Associator", width, height, True)
        vp = win.WinViewport2D()
        
        self.win = win
        self.vp = vp
        
        updt = vp.UpdateStart()
         
        mfr = win.SetMainFrame()
        
        tbar = gi.ToolBar(mfr.AddNewChild(gi.KiT_ToolBar(), "tbar"))
        tbar.Lay = gi.LayoutHoriz
        tbar.SetStretchMaxWidth()
        
        split = gi.SplitView(mfr.AddNewChild(gi.KiT_SplitView(), "split"))
        split.Dim = gi.X
        # split.SetProp("horizontal-align", "center")
        # split.SetProp("margin", 2.0) # raw numbers = px = 96 dpi pixels
        split.SetStretchMaxWidth()
        split.SetStretchMaxHeight()
         
        self.ClassView = pygiv.ClassView("ra25sv", self.Tags)
        self.ClassView.AddFrame(split)
        self.ClassView.SetClass(self)
        
        svge = svg.Editor(split.AddNewChild(svg.KiT_Editor(), "svg"))
        svge.InitScale()
        svge.Fill = True
        svge.SetPropStr("background-color", "white")
        svge.SetPropStr("width", "800px")
        svge.SetPropStr("height", "1100px")
        svge.SetStretchMaxWidth()
        svge.SetStretchMaxHeight()
        self.EpcPlotSvg = svge
         
        # split.SetSplits(.3, .7)  # not avail due to var-arg
        
        recv = win.This()
        
        tbar.AddAction(gi.ActOpts(Label="Init", Icon="update"), recv, InitCB)
        tbar.AddAction(gi.ActOpts(Label="Train", Icon="run"), recv, TrainCB)
        tbar.AddAction(gi.ActOpts(Label="Stop", Icon="stop"), recv, StopCB)
        tbar.AddAction(gi.ActOpts(Label="Step Trial", Icon="step-fwd"), recv, StepTrialCB)
        tbar.AddAction(gi.ActOpts(Label="Step Epoch", Icon="fast-fwd"), recv, StepEpochCB)
        
        # tbar.AddSep("file")
        
        tbar.AddAction(gi.ActOpts(Label="Epoch Plot", Icon="update"), recv, PlotEpcLogCB)
        tbar.AddAction(gi.ActOpts(Label="Save Wts", Icon="file-save"), recv, SaveWtsCB)
        tbar.AddAction(gi.ActOpts(Label="Save Log", Icon="file-save"), recv, SaveLogCB)
        tbar.AddAction(gi.ActOpts(Label="Save Plot", Icon="file-save"), recv, SaveEpcPlotCB)
        tbar.AddAction(gi.ActOpts(Label="Save Pars", Icon="file-save"), recv, SaveParsCB)
        tbar.AddAction(gi.ActOpts(Label="New Seed", Icon="new"), recv, NewRndSeedCB)
        
        # main menu
        appnm = gi.AppName()
        mmen = win.MainMenu
        mmen.ConfigMenus(go.Slice_string([appnm, "File", "Edit", "Window"]))
        
        amen = gi.Action(win.MainMenu.ChildByName(appnm, 0))
        amen.Menu.AddAppMenu(win)
        
        emen = gi.Action(win.MainMenu.ChildByName("Edit", 1))
        emen.Menu.AddCopyCutPaste(win)
        
        # note: Command in shortcuts is automatically translated into Control for
        # Linux, Windows or Meta for MacOS
        # fmen = win.MainMenu.ChildByName("File", 0).(*gi.Action)
        # fmen.Menu = make(gi.Menu, 0, 10)
        # fmen.Menu.AddAction(gi.ActOpts{Label: "Open", Shortcut: "Command+O"},
        #     recv, func(recv, send ki.Ki, sig int64, data interface{}) {
        #     FileViewOpenSVG(vp)
        #     })
        # fmen.Menu.AddSeparator("csep")
        # fmen.Menu.AddAction(gi.ActOpts{Label: "Close Window", Shortcut: "Command+W"},
        #     recv, func(recv, send ki.Ki, sig int64, data interface{}) {
        #     win.CloseReq()
        #     })
                
        #    win.SetCloseCleanFunc(func(w *gi.Window) {
        #         gi.Quit() # once main window is closed, quit
        #     })
        #         
        win.MainMenuUpdated()
        vp.UpdateEndNoSig(updt)
        win.GoStartEventLoop()
        

# Sim is the overall state for this simulation
Sim = SimState()

Sim.Config()
Sim.Init()
Sim.ConfigGui()
# Sim.Train()
# Sim.EpcLog.SaveCSV("ra25_epc.dat", ord(','), True)

    
