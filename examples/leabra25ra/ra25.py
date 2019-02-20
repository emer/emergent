# Copyright (c) 2019, The Emergent Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# labra25ra runs a simple random-associator 5x5 = 25 four-layer leabra network
from emergent.leabra.leabra import leabra
from emergent.emer import emer
from emergent.eplot import eplot
from emergent.patgen import patgen
from emergent.prjn import prjn
from emergent.dtable import dtable

# DefaultPars are the initial default parameters for this simulation
# DefaultPars = emer.ParamStyle{
#     "Prjn": {
#     "Prjn.Learn.Norm.On":     1,
#     "Prjn.Learn.Momentum.On": 1,
#     "Prjn.Learn.WtBal.On":    0,
#     },
#     # "Layer": {
#     #     "Layer.Inhib.Layer.Gi": 1.8, # this is the default
#     # },
#     "#Output": {
#     "Layer.Inhib.Layer.Gi": 1.4, # this turns out to be critical for small output layer
#     },
#     ".Back": {
#     "Prjn.WtScale.Rel": 0.2, # this is generally quite important
#     },
# }

class nil(object):
    def __init__(self):
        self.handle = 0

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
        self.Pats     = dtable.Table()
        self.EpcLog   = dtable.Table()
        # self.Pars     = emer.ParamStyle()
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
        self.CntErr     = 0
        self.Porder     = leabra.SliceOf_int([])
#        self.EpcPlotSvg *svg.Editor
        self.StopNow    = False
        self.RndSeed    = 0


    def Config(self):
        """Config configures all the elements using the standard functions"""
        self.ConfigNet()
        self.OpenPats()
        self.ConfigEpcLog()

    def Init(self):
        """Init restarts the run, and initializes everything, including network weights and resets the epoch log table"""
        # rand.Seed(self.RndSeed)
        if self.MaxEpcs == 0: # allow user override
            self.MaxEpcs = 100
        self.Epoch = 0
        self.Trial = 0
        self.StopNow = false
        self.Time.Reset()
        np = self.Pats.NumRows()
        # self.Porder = rand.Perm(np)         # always start with new one so random order is identical
        # self.Net.StyleParams(self.Pars, true) # set msg
        self.Net.InitWts()
        self.EpcLog.SetNumRows(0)

    def NewRndSeed(self):
        """NewRndSeed gets a new random seed based on current time -- otherwise uses the same random seed for every run"""
        # self.RndSeed = time.Now().UnixNano()
        
    def RunTrial(self):
        """RunTrial runs one alpha-trial (100 msec, 4 quarters) of processing
        this does NOT call TrialInc (so it can be used flexibly)
        but it does use the Trial counter to determine which pattern to present."""
    
        inLay = self.Net.LayerByName("Input")
        outLay = self.Net.LayerByName("Output")
        inPats = self.Pats.ColByName("Input")
        outPats = self.Pats.ColByName("Output")
        
        pidx = self.Trial
        if not self.Sequential:
            pidx = self.Porder[self.Trial]
            
        pslc = leabra.SliceOf_int([pidx])
        
        inp = inPats.SubSlice(2, pslc)
        outp = outPats.SubSlice(2, pslc)
        inLay.ApplyExt(inp)
        outLay.ApplyExt(outp)
        
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
        np = self.Pats.NumRows()
        if self.Trial >= np:
            self.LogEpoch()
            if self.Plot:
                self.PlotEpcLog()
            self.EpochInc()

    def TrialStats(self, accum):
        """TrialStats computes the trial-level statistics and adds them to the
        epoch accumulators if accum is true"""
        outLay = self.Net.LayerByName("Output")
        cosdiff = outLay.CosDiff.Cos
        # todo: multi-return val not there:
        # sse, avgsse = outLay.SSE(0.5) # 0.5 = per-unit tolerance -- right side of .5
        # todo: this whole method returns multiple values..
        if accum:
            self.SumSSE += sse
            self.SumAvgSSE += avgsse
            self.SumCosDiff += cosdiff
            if sse != 0:
                self.CntErr+= 1

    def EpochInc(self):
        """EpochInc increments counters after one epoch of processing and updates a new random
        order of permuted inputs for the next epoch"""
        self.Trial = 0
        self.Epoch += 1
        erand.PermuteInts(self.Porder)

    def LogEpoch(self):
        """LogEpoch adds data from current epoch to the EpochLog table -- computes epoch
        averages prior to logging.
        Epoch counter is assumed to not have yet been incremented."""
        self.EpcLog.SetNumRows(self.Epoch + 1)
        hid1Lay = self.Net.LayerByName("Hidden1")
        hid2Lay = self.Net.LayerByName("Hidden2")
        outLay = self.Net.LayerByName("Output")
        
        np = self.Pats.NumRows()
        self.EpcSSE = self.SumSSE / np
        self.SumSSE = 0.0
        self.EpcAvgSSE = self.SumAvgSSE / np
        self.SumAvgSSE = 0.0
        self.EpcPctErr = self.CntErr / np
        self.CntErr = 0.0
        self.EpcPctCor = 1.0 - self.EpcPctErr
        self.EpcCosDiff = self.SumCosDiff / np
        self.SumCosDiff = 0.0
        
        epc = self.Epoch
        
        self.EpcLog.ColByName("Epoch").SetFloat1D(epc, epc)
        self.EpcLog.ColByName("SSE").SetFloat1D(epc, self.EpcSSE)
        self.EpcLog.ColByName("Avg SSE").SetFloat1D(epc, self.EpcAvgSSE)
        self.EpcLog.ColByName("Pct Err").SetFloat1D(epc, self.EpcPctErr)
        self.EpcLog.ColByName("Pct Cor").SetFloat1D(epc, self.EpcPctCor)
        self.EpcLog.ColByName("CosDiff").SetFloat1D(epc, self.EpcCosDiff)
        self.EpcLog.ColByName("Hid1 ActAvg").SetFloat1D(epc, hid1Lay.Pools[0].ActAvg.ActPAvgEff)
        self.EpcLog.ColByName("Hid2 ActAvg").SetFloat1D(epc, hid2Lay.Pools[0].ActAvg.ActPAvgEff)
        self.EpcLog.ColByName("Out ActAvg").SetFloat1D(epc, outLay.Pools[0].ActAvg.ActPAvgEff)

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

    def Train(self):
        """Train runs the full training from this point onward"""
        self.StopNow = false
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
        self.StopNow = true

    # --------- Config methods -----------

    def ConfigNet(self):
        net = self.Net
        net.InitName(net, "RA25")
        inLay = net.AddLayer("Input", leabra.SliceOf_int([5, 5]), emer.Input)
        hid1Lay = net.AddLayer("Hidden1", leabra.SliceOf_int([7, 7]), emer.Hidden)
        hid2Lay = net.AddLayer("Hidden2", leabra.SliceOf_int([7, 7]), emer.Hidden)
        outLay = net.AddLayer("Output", leabra.SliceOf_int([5, 5]), emer.Target)
        
        net.ConnectLayers(inLay, hid1Lay, prjn.NewFull(), emer.Forward)
        net.ConnectLayers(hid1Lay, hid2Lay, prjn.NewFull(), emer.Forward)
        net.ConnectLayers(hid2Lay, outLay, prjn.NewFull(), emer.Forward)
        
        net.ConnectLayers(outLay, hid2Lay, prjn.NewFull(), emer.Back)
        net.ConnectLayers(hid2Lay, hid1Lay, prjn.NewFull(), emer.Back)
        
        net.Defaults()
        # net.StyleParams(self.Pars, true) # set msg
        net.Build()
        net.InitWts()

    def ConfigPats(self):
        dt = self.Pats
        schema = dtable.Schema()
        schema.append(dtable.Column("Name", etensor.STRING, nil, nil))
        schema.append(dtable.Column("Input", etensor.FLOAT32, dtable.SliceOf_int([5, 5]), dtable.SliceOf_string(["Y", "X"])))
        schema.append(dtable.Column("Output", etensor.FLOAT32, dtable.SliceOf_int([5, 5]), dtable.SliceOf_string(["Y", "X"])))
        dt.SetFromSchema(schema, 25)
            
        patgen.PermutedBinaryRows(dt.Cols[1], 6, 1, 0)
        patgen.PermutedBinaryRows(dt.Cols[2], 6, 1, 0)
        dt.SaveCSV("random_5x5_25_gen.dat", ',', true)

    def OpenPats(self):
        dt = self.Pats
        dt.OpenCSV("random_5x5_25.dat", '\t')
        # if err != nil:
        #     log.Println(err)

    def ConfigEpcLog(self):
        dt = self.EpcLog
        schema = dtable.Schema()
        schema.append(dtable.Column("Epoch", etensor.INT64, nil, nil)),
        schema.append(dtable.Column("SSE", etensor.FLOAT32, nil, nil)),
        schema.append(dtable.Column("Avg SSE", etensor.FLOAT32, nil, nil)),
        schema.append(dtable.Column("Pct Err", etensor.FLOAT32, nil, nil)),
        schema.append(dtable.Column("Pct Cor", etensor.FLOAT32, nil, nil)),
        schema.append(dtable.Column("CosDiff", etensor.FLOAT32, nil, nil)),
        schema.append(dtable.Column("Hid1 ActAvg", etensor.FLOAT32, nil, nil)),
        schema.append(dtable.Column("Hid2 ActAvg", etensor.FLOAT32, nil, nil)),
        schema.append(dtable.Column("Out ActAvg", etensor.FLOAT32, nil, nil)),
        dt.SetFromSchema(schema, 0)
            
        self.PlotVals = dtable.SliceOf_string(["SSE", "Pct Err"])
        self.Plot = true

    def PlotEpcLog(self):
        """PlotEpcLog plots given epoch log using given Y axis columns into EpcPlotSvg"""
        dt = self.EpcLog
        # plt = plot.New() # todo: keep around?
        # plt.Title.Text = "Random Associator Epoch Log"
        # plt.X.Label.Text = "Epoch"
        # plt.Y.Label.Text = "Y"
        
        # for cl in range(self.PlotVals):
        #     xy = eplot.NewTableXYNames(dt, "Epoch", cl)
            # plotutil.AddLines(plt, cl, xy)

        #eplot.PlotViewSVG(plt, self.EpcPlotSvg, 5, 5, 2)


    # def ConfigGui() *gi.Window {
    #     """ConfigGui configures the GoGi gui interface for this simulation"""
    #     width = 1600
    #     height = 1200
    #     
    #     oswin.TheApp.SetName("leabra25ra")
    #     oswin.TheApp.SetAbout(`This demonstrates a basic Leabra model. See <a href="https:#github.com/emer/emergent">emergent on GitHub</a>.</p>`)
    #     
    #     win = gi.NewWindow2D("leabra25ra", "Leabra Random Associator", width, height, true)
    #     
    #     vp = win.WinViewport2D()
    #     updt = vp.UpdateStart()
    #     
    #     mfr = win.SetMainFrame()
    #     
    #     tbar = mfr.AddNewChild(gi.KiT_ToolBar, "tbar").(*gi.ToolBar)
    #     tbar.Lay = gi.LayoutHoriz
    #     tbar.SetStretchMaxWidth()
    #     
    #     split = mfr.AddNewChild(gi.KiT_SplitView, "split").(*gi.SplitView)
    #     split.Dim = gi.X
    #     # split.SetProp("horizontal-align", "center")
    #     # split.SetProp("margin", 2.0) # raw numbers = px = 96 dpi pixels
    #     split.SetStretchMaxWidth()
    #     split.SetStretchMaxHeight()
    #     
    #     # todo: add a splitview here
    #     
    #     sv = split.AddNewChild(giv.KiT_StructView, "sv").(*giv.StructView)
    #     sv.SetStruct(ss, nil)
    #     # sv.SetStretchMaxWidth()
    #     # sv.SetStretchMaxHeight()
    #     
    #     svge = split.AddNewChild(svg.KiT_Editor, "svg").(*svg.Editor)
    #     svge.InitScale()
    #     svge.Fill = true
    #     svge.SetProp("background-color", "white")
    #     svge.SetProp("width", units.NewValue(float32(width/2), units.Px))
    #     svge.SetProp("height", units.NewValue(float32(height-100), units.Px))
    #     svge.SetStretchMaxWidth()
    #     svge.SetStretchMaxHeight()
    #     self.EpcPlotSvg = svge
    #     
    #     split.SetSplits(.3, .7)
    #     
    #     tbar.AddAction(gi.ActOpts{Label: "Init", Icon: "update"}, win.This(),
    #     func(recv, send ki.Ki, sig int64, data interface{}) {
    #         self.Init()
    #         vp.FullRender2DTree()
    #     })
    #         
    #     tbar.AddAction(gi.ActOpts{Label: "Train", Icon: "run"}, win.This(),
    #     func(recv, send ki.Ki, sig int64, data interface{}) {
    #         go self.Train()
    #      })
    #         
    #     tbar.AddAction(gi.ActOpts{Label: "Stop", Icon: "stop"}, win.This(),
    #     func(recv, send ki.Ki, sig int64, data interface{}) {
    #         self.Stop()
    #         vp.FullRender2DTree()
    #         })
    #                 
    #     tbar.AddAction(gi.ActOpts{Label: "Step Trial", Icon: "step-fwd"}, win.This(),
    #     func(recv, send ki.Ki, sig int64, data interface{}) {
    #         self.StepTrial()
    #         vp.FullRender2DTree()
    #         })
    #                     
    #     tbar.AddAction(gi.ActOpts{Label: "Step Epoch", Icon: "fast-fwd"}, win.This(),
    #     func(recv, send ki.Ki, sig int64, data interface{}) {
    #         self.StepEpoch()
    #         vp.FullRender2DTree()
    #         })
    #                         
    #     # tbar.AddSep("file")
    #                         
    #     tbar.AddAction(gi.ActOpts{Label: "Epoch Plot", Icon: "update"}, win.This(),
    #     func(recv, send ki.Ki, sig int64, data interface{}) {
    #         self.PlotEpcLog()
    #         })
    #                             
    #     tbar.AddAction(gi.ActOpts{Label: "Save Wts", Icon: "file-save"}, win.This(),
    #     func(recv, send ki.Ki, sig int64, data interface{}) {
    #         self.Net.SaveWtsJSON("ra25_net_trained.wts") # todo: call method to prompt
    #         })
    #                                 
    #     tbar.AddAction(gi.ActOpts{Label: "Save Log", Icon: "file-save"}, win.This(),
    #     func(recv, send ki.Ki, sig int64, data interface{}) {
    #         self.EpcLog.SaveCSV("ra25_epc.dat", ',', true)
    #         })
    #                                     
    #     tbar.AddAction(gi.ActOpts{Label: "Save Pars", Icon: "file-save"}, win.This(),
    #     func(recv, send ki.Ki, sig int64, data interface{}) {
    #         # todo: need save / load methods for these
    #         # self.EpcLog.SaveCSV("ra25_epc.dat", ',', true)
    #         })
    #                                         
    #     tbar.AddAction(gi.ActOpts{Label: "New Seed", Icon: "new"}, win.This(),
    #     func(recv, send ki.Ki, sig int64, data interface{}) {
    #         self.NewRndSeed()
    #         })
    #                                             
    #     vp.UpdateEndNoSig(updt)
    #     
    #     # main menu
    #     appnm = oswin.TheApp.Name()
    #     mmen = win.MainMenu
    #     mmen.ConfigMenus([]string{appnm, "File", "Edit", "Window"})
    #     
    #     amen = win.MainMenu.ChildByName(appnm, 0).(*gi.Action)
    #     amen.Menu = make(gi.Menu, 0, 10)
    #     amen.Menu.AddAppMenu(win)
    #     
    #     emen = win.MainMenu.ChildByName("Edit", 1).(*gi.Action)
    #     emen.Menu = make(gi.Menu, 0, 10)
    #     emen.Menu.AddCopyCutPaste(win)
    #     
    #     # note: Command in shortcuts is automatically translated into Control for
    #     # Linux, Windows or Meta for MacOS
    #     # fmen = win.MainMenu.ChildByName("File", 0).(*gi.Action)
    #     # fmen.Menu = make(gi.Menu, 0, 10)
    #     # fmen.Menu.AddAction(gi.ActOpts{Label: "Open", Shortcut: "Command+O"},
    #     #     win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
    #     #     FileViewOpenSVG(vp)
    #     #     })
    #     # fmen.Menu.AddSeparator("csep")
    #     # fmen.Menu.AddAction(gi.ActOpts{Label: "Close Window", Shortcut: "Command+W"},
    #     #     win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
    #     #     win.OSWin.Close()
    #     #     })
    #     
    #     win.OSWin.SetCloseCleanFunc(func(w oswin.Window) {
    #         go oswin.TheApp.Quit() # once main window is closed, quit
    #     })
    #         
    #     win.MainMenuUpdated()
    #     return win


# Sim is the overall state for this simulation
Sim = SimState()

Sim.Config()
Sim.Init()
#win = Sim.ConfigGui()
Sim.Run()

