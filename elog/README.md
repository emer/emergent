Docs: [GoDoc](https://pkg.go.dev/github.com/emer/emergent/elog)

`elog` provides a full infrastructure for recording data of all sorts at multiple time scales and evaluation modes (training, testing, validation, etc).

The `elog.Item` provides a full definition of each distinct item that is logged with a map of Write functions keyed by a scope string that reflects the time scale and mode.  The same function can be used across multiple scopes, or a different function for each scope, etc.

The Items are written to the table *in the order added*, so you can take advantage of previously-computed item values based on the actual ordering of item code.  For example, intermediate values can be stored / retrieved from Stats, or from other items on a log, e.g., using `Context.LogItemFloat` function.

The Items are then processed in `CreateTables()` to create a set of `etable.Table` tables to hold the data.

The `elog.Logs` struct holds all the relevant data and functions for managing the logging process.

* `Log(mode, time)` does logging, adding a new row

* `LogRow(mode, time, row)` does logging at given row

Both of these functions automatically write incrementally to a `tsv` File if it has been opened.

The `Context` object is passed to the Item Write functions, and has all the info typically needed -- must call `SetContext(stats, net)` on the Logs to provide those elements.  Write functions can do most standard things by calling methods on Context -- see that in Docs above for more info.

# Scopes

Everything is organized according to a `ScopeKey`, which is just a `string`, that is formatted to represent two factors: an **evaluation mode** (standard versions defined by `EvalModes` enum) and a **time scale** (`Times` enum).

Standard `EvalModes` are:
* `Train`
* `Test`
* `Validate`
* `Analyze` -- used for internal representational analysis functions such as PCA, ActRF, SimMat, etc.

Standard `Times` are based on the [Env](https://github.com/emer/emergent/wiki/Env) `TimeScales` augmented with Leabra / Axon finer-grained scales, including:
* `Cycle`
* `Trial`
* `Epoch`
* `Run`

Other arbitrary scope values can be used -- there are `Scope` versions of every method that take an arbitrary `ScopeKey` that can be composed using the `ScopeStr` method from any two strings, along with the "plain" versions of these methods that take the standard `mode` and `time` enums for convenience.  These enums can themselves also be extended but it is probably easier to just use strings.

# Examples

The [ra25](https://github.com/emer/axon/tree/master/examples/ra25) example has a fully updated implementation of this new logging infrastructure.  The individual log Items are added in the `logitems.go` file, which keeps the main sim file smaller and easier to navigate.  It is also a good idea to put the params in a separate `params.go` file, as we now do in this example.

## Main Config and Log functions

The `ConfigLogs` function configures the items, creates the tables, and configures any other log-like entities including spike rasters.

```Go
func (ss *Sim) ConfigLogs() {
    ss.ConfigLogItems()
    ss.Logs.CreateTables()
    ss.Logs.SetContext(&ss.Stats, ss.Net)
    // don't plot certain combinations we don't use
    ss.Logs.NoPlot(elog.Train, elog.Cycle)
    ss.Logs.NoPlot(elog.Test, elog.Run)
    // note: Analyze not plotted by default
    ss.Logs.SetMeta(elog.Train, elog.Run, "LegendCol", "Params")
    ss.Stats.ConfigRasters(ss.Net, ss.Net.LayersByClass())
}
```

There is one master `Log` function that handles any details associated with different levels of logging -- it is called with the scope elements, e.g., `ss.Log(elog.Train, elog.Trial)`

```Go
// Log is the main logging function, handles special things for different scopes
func (ss *Sim) Log(mode elog.EvalModes, time elog.Times) {
    dt := ss.Logs.Table(mode, time)
    row := dt.Rows
    switch {
    case mode == elog.Test && time == elog.Epoch:
        ss.LogTestErrors()
    case mode == elog.Train && time == elog.Epoch:
        epc := ss.TrainEnv.Epoch.Cur
        if (ss.PCAInterval > 0) && ((epc-1)%ss.PCAInterval == 0) { // -1 so runs on first epc
            ss.PCAStats()
        }
    case time == elog.Cycle:
        row = ss.Stats.Int("Cycle")
    case time == elog.Trial:
        row = ss.Stats.Int("Trial")
    }

    ss.Logs.LogRow(mode, time, row) // also logs to file, etc
    if time == elog.Cycle {
        ss.GUI.UpdateCyclePlot(elog.Test, ss.Time.Cycle)
    } else {
        ss.GUI.UpdatePlot(mode, time)
    }

    // post-logging special statistics
    switch {
    case mode == elog.Train && time == elog.Run:
        ss.LogRunStats()
    case mode == elog.Train && time == elog.Trial:
        epc := ss.TrainEnv.Epoch.Cur
        if (ss.PCAInterval > 0) && (epc%ss.PCAInterval == 0) {
            ss.Log(elog.Analyze, elog.Trial)
        }
    }
}
```

There are various additional analysis functions called here, for example this one that generates summary statistics about the overall performance across runs -- these are stored in the `MiscTables` in the `Logs` object:

```Go
// LogRunStats records stats across all runs, at Train Run scope
func (ss *Sim) LogRunStats() {
    sk := elog.Scope(elog.Train, elog.Run)
    lt := ss.Logs.TableDetailsScope(sk)
    ix, _ := lt.NamedIdxView("RunStats")

    spl := split.GroupBy(ix, []string{"Params"})
    split.Desc(spl, "FirstZero")
    split.Desc(spl, "PctCor")
    ss.Logs.MiscTables["RunStats"] = spl.AggsToTable(etable.AddAggName)
}
```

## Counter Items

All counters of interest should be written to [estats](https://github.com/emer/emergent/tree/master/estats) `Stats` elements, whenever the counters might be updated, and then logging just reads those stats.  Here's a `StatCounters` function:

```Go
// StatCounters saves current counters to Stats, so they are available for logging etc
// Also saves a string rep of them to the GUI, if the GUI is active
func (ss *Sim) StatCounters(train bool) {
    ev := ss.TrainEnv
    if !train {
        ev = ss.TestEnv
    }
    ss.Stats.SetInt("Run", ss.TrainEnv.Run.Cur)
    ss.Stats.SetInt("Epoch", ss.TrainEnv.Epoch.Cur)
    ss.Stats.SetInt("Trial", ev.Trial.Cur)
    ss.Stats.SetString("TrialName", ev.TrialName.Cur)
    ss.Stats.SetInt("Cycle", ss.Time.Cycle)
    ss.GUI.NetViewText = ss.Stats.Print([]string{"Run", "Epoch", "Trial", "TrialName", "Cycle", "TrlUnitErr", "TrlErr", "TrlCosDiff"})
}
```

Then they are easily logged -- just showing different Scope expressions here:

```Go
    ss.Logs.AddItem(&elog.Item{
        Name: "Run",
        Type: etensor.INT64,
        Plot: elog.DFalse,
        Write: elog.WriteMap{
            elog.Scope(elog.AllModes, elog.AllTimes): func(ctx *elog.Context) {
                ctx.SetStatInt("Run")
            }}})
```

```Go
    ss.Logs.AddItem(&elog.Item{
        Name: "Epoch",
        Type: etensor.INT64,
        Plot: elog.DFalse,
        Write: elog.WriteMap{
            elog.Scopes([]elog.EvalModes{elog.AllModes}, []elog.Times{elog.Epoch, elog.Trial}): func(ctx *elog.Context) {
                ctx.SetStatInt("Epoch")
            }}})
```            

```Go
    ss.Logs.AddItem(&elog.Item{
        Name: "Trial",
        Type: etensor.INT64,
        Write: elog.WriteMap{
            elog.Scope(elog.AllModes, elog.Trial): func(ctx *elog.Context) {
                ctx.SetStatInt("Trial")
            }}})
```

## Performance Stats

Overall summary performance statistics have multiple Write functions for different scopes, performing aggregation over log data at lower levels:

```Go
    ss.Logs.AddItem(&elog.Item{
        Name: "UnitErr",
        Type: etensor.FLOAT64,
        Plot: elog.DFalse,
        Write: elog.WriteMap{
            elog.Scope(elog.AllModes, elog.Trial): func(ctx *elog.Context) {
                ctx.SetStatFloat("TrlUnitErr")
            }, elog.Scope(elog.AllModes, elog.Epoch): func(ctx *elog.Context) {
                ctx.SetAgg(ctx.Mode, elog.Trial, agg.AggMean)
            }, elog.Scope(elog.AllModes, elog.Run): func(ctx *elog.Context) {
                ix := ctx.LastNRows(ctx.Mode, elog.Epoch, 5)
                ctx.SetFloat64(agg.Mean(ix, ctx.Item.Name)[0])
            }}})
```

## Layer Stats

Iterate over layers of interest (use `LayersByClass` function). It is *essential* to create a local variable inside the loop for the `lnm` variable, which is then captured by the closure (see https://github.com/golang/go/wiki/CommonMistakes):

```Go
    // Standard stats for Ge and AvgAct tuning -- for all hidden, output layers
    layers := ss.Net.LayersByClass("Hidden", "Target")
    for _, lnm := range layers {
        clnm := lnm
        ss.Logs.AddItem(&elog.Item{
            Name:   clnm + "_ActAvg",
            Type:   etensor.FLOAT64,
            Plot:   elog.DFalse,
            FixMax: elog.DFalse,
            Range:  minmax.F64{Max: 1},
            Write: elog.WriteMap{
                elog.Scope(elog.Train, elog.Epoch): func(ctx *elog.Context) {
                    ly := ctx.Layer(clnm).(axon.AxonLayer).AsAxon()
                    ctx.SetFloat32(ly.ActAvg.ActMAvg)
                }}})
          ...
    }
```

## Layer Activity Patterns

A log column can be a tensor of any shape -- the `SetLayerTensor` method on the Context grabs the data from the layer into a reused tensor (no memory churning after first initialization), and then stores that tensor into the log column.

```Go
    // input / output layer activity patterns during testing
    layers = ss.Net.LayersByClass("Input", "Target")
    for _, lnm := range layers {
        clnm := lnm
        cly := ss.Net.LayerByName(clnm)
        ss.Logs.AddItem(&elog.Item{
            Name:      clnm + "_Act",
            Type:      etensor.FLOAT64,
            CellShape: cly.Shape().Shp,
            FixMax:    elog.DTrue,
            Range:     minmax.F64{Max: 1},
            Write: elog.WriteMap{
                elog.Scope(elog.Test, elog.Trial): func(ctx *elog.Context) {
                    ctx.SetLayerTensor(clnm, "Act")
                }}})
```

## PCA on Activity

Computing stats on the principal components of variance (PCA) across different input patterns is very informative about the nature of the internal representations in hidden layers.  The [estats](https://github.com/emer/emergent/tree/master/estats) package has support for this -- it is fairly expensive computationally so we only do this every N epochs (10 or so), calling this method:

```Go
// PCAStats computes PCA statistics on recorded hidden activation patterns
// from Analyze, Trial log data
func (ss *Sim) PCAStats() {
    ss.Stats.PCAStats(ss.Logs.IdxView(elog.Analyze, elog.Trial), "ActM", ss.Net.LayersByClass("Hidden"))
    ss.Logs.ResetLog(elog.Analyze, elog.Trial)
}
```

Here's how you record the data and log the resulting stats, using the `Analyze` `EvalMode`: 

```Go
    // hidden activities for PCA analysis, and PCA results
    layers = ss.Net.LayersByClass("Hidden")
    for _, lnm := range layers {
        clnm := lnm
        cly := ss.Net.LayerByName(clnm)
        ss.Logs.AddItem(&elog.Item{
            Name:      clnm + "_ActM",
            Type:      etensor.FLOAT64,
            CellShape: cly.Shape().Shp,
            FixMax:    elog.DTrue,
            Range:     minmax.F64{Max: 1},
            Write: elog.WriteMap{
                elog.Scope(elog.Analyze, elog.Trial): func(ctx *elog.Context) {
                    ctx.SetLayerTensor(clnm, "ActM")
                }}})
        ss.Logs.AddItem(&elog.Item{
            Name: clnm + "_PCA_NStrong",
            Type: etensor.FLOAT64,
            Plot: elog.DFalse,
            Write: elog.WriteMap{
                elog.Scope(elog.Train, elog.Epoch): func(ctx *elog.Context) {
                    ctx.SetStatFloat(ctx.Item.Name)
                }, elog.Scope(elog.AllModes, elog.Run): func(ctx *elog.Context) {
                    ix := ctx.LastNRows(ctx.Mode, elog.Epoch, 5)
                    ctx.SetFloat64(agg.Mean(ix, ctx.Item.Name)[0])
                }}})
       ...
    }
```

## Error by Input Category

This item creates a tensor column that records the average error for each category of input stimulus (e.g., for images from object categories), using the `split.GroupBy` function for `etable`.  The `IdxView` function (see also `NamedIdxView`) automatically manages the `etable.IdxView` indexed view onto a log table, which is used for all aggregation and further analysis of data, so that you can efficiently analyze filtered subsets of the original data.

```Go
    ss.Logs.AddItem(&elog.Item{
        Name:      "CatErr",
        Type:      etensor.FLOAT64,
        CellShape: []int{20},
        DimNames:  []string{"Cat"},
        Plot:      elog.DTrue,
        Range:     minmax.F64{Min: 0},
        TensorIdx: -1, // plot all values
        Write: elog.WriteMap{
            elog.Scope(elog.Test, elog.Epoch): func(ctx *elog.Context) {
                ix := ctx.Logs.IdxView(elog.Test, elog.Trial)
                spl := split.GroupBy(ix, []string{"Cat"})
                split.AggTry(spl, "Err", agg.AggMean)
                cats := spl.AggsToTable(etable.ColNameOnly)
                ss.Logs.MiscTables[ctx.Item.Name] = cats
                ctx.SetTensor(cats.Cols[1])
            }}})
```

## Confusion matricies


## Closest Pattern Stat

The [estats](https://github.com/emer/emergent/tree/master/estats) package has a `ClosestPat` function that grabs the activity from a given variable in a given layer, and compares it to a list of patterns in a table, returning the pattern that is closest to the layer activity pattern, using the Correlation metric, which is the most robust metric in terms of ignoring differences in overall activity levels.  You can also compare that closest pattern name to a (list of) acceptable target names and use that as an error measure.

```Go
    row, cor, cnm := ss.Stats.ClosestPat(ss.Net, "Output", "ActM", ss.Pats, "Output", "Name")
    ss.Stats.SetString("TrlClosest", cnm)
    ss.Stats.SetFloat("TrlCorrel", float64(cor))
    tnm := ss.TrainEnv.TrialName
    if cnm == tnm {
        ss.Stats.SetFloat("TrlErr", 0)
    } else {
        ss.Stats.SetFloat("TrlErr", 1)
    }
```

## Activation-based Receptive Fields

The [estats](https://github.com/emer/emergent/tree/master/estats) package has support for recording activation-based receptive fields ([actrf](https://github.com/emer/emergent/tree/master/actrf)), which are very useful for decoding what units represent.

First, initialize the ActRFs in the `ConfigLogs` function, using strings that specify the layer name to record activity from, followed by the source data for the receptive field, which can be *anything* that might help you understand what the units are responding to, including the name of another layer.  If it is not another layer name, then the code will look for the name in the `Stats.F32Tensors` map of named tensors.

```Go
    ss.Stats.SetF32Tensor("Image", &ss.TestEnv.Vis.ImgTsr) // image used for actrfs, must be there first
    ss.Stats.InitActRFs(ss.Net, []string{"V4:Image", "V4:Output", "IT:Image", "IT:Output"}, "ActM")
```

To add tabs in the gui to visualize the resulting RFs, add this in your `ConfigGui` (note also adding a tab to visualize the input Image that is being presented to the network):

```Go
    tg := ss.GUI.TabView.AddNewTab(etview.KiT_TensorGrid, "Image").(*etview.TensorGrid)
    tg.SetStretchMax()
    ss.GUI.SetGrid("Image", tg)
    tg.SetTensor(&ss.TrainEnv.Vis.ImgTsr)

    ss.GUI.AddActRFGridTabs(&ss.Stats.ActRFs)
```

At the relevant `Trial` level, call the function to update the RF data based on current network activity state:

```Go
    ss.Stats.UpdateActRFs(ss.Net, "ActM", 0.01)
```

Here's a `TestAll` function that manages the testing of a large number of inputs to compute the RFs (often need a large amount of testing data to sample the space sufficiently to get meaningful results):

```Go
// TestAll runs through the full set of testing items
func (ss *Sim) TestAll() {
    ss.TestEnv.Init(ss.TrainEnv.Run.Cur)
    ss.Stats.ActRFs.Reset() // initialize prior to testing
    for {
        ss.TestTrial(true)
        ss.Stats.UpdateActRFs(ss.Net, "ActM", 0.01)
        _, _, chg := ss.TestEnv.Counter(env.Epoch)
        if chg || ss.StopNow {
            break
        }
    }
    ss.Stats.ActRFsAvgNorm() // final 
    ss.GUI.ViewActRFs(&ss.Stats.ActRFs)
}
```

## Representational Similarity Analysis (SimMat)

## Cluster Plots


