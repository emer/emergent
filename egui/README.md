Docs: [GoDoc](https://pkg.go.dev/github.com/emer/emergent/egui)

`egui` handles all the GUI elements for a typical simulation, reducing boilerplate code in models.

The [ra25](https://github.com/emer/axon/tree/master/examples/ra25) example has a fully updated implementation of this new GUI infrastructure. 

# Examples

Here's the start of the main ConfigGUI method:

```Go
// ConfigGUI configures the Cogent Core GUI interface for this simulation.
func (ss *Sim) ConfigGUI() *core.Window {
	title := "Leabra Random Associator"
	ss.GUI.MakeWindow(ss, "ra25", title, `This demonstrates a basic Leabra model. See <a href="https://github.com/emer/emergent">emergent on GitHub</a>.</p>`)
	ss.GUI.CycleUpdateInterval = 10
	ss.GUI.NetView.SetNet(ss.Net)

    // optionally reconfigure the netview:
	ss.GUI.NetView.Scene().Camera.Pose.Pos.Set(0, 1, 2.75) 
	ss.GUI.NetView.Scene().Camera.LookAt(math32.V3(0, 0, 0), math32.V3(0, 1, 0)) 
	ss.GUI.AddPlots(title, &ss.Logs) // automatically adds all configured plots
```


## Toolbar Items

The `ToolbarItem` class provides toolbar configuration options, taking the place of `core.ActOpts` from existing code that operates directly at the `GoGi` level.  The main differences are

* The standard `UpdateFunc` options of either making the action active or inactive while the sim is running are now handled using `Active: equi.ActiveStopped` or `egui.ActiveRunning` or `egui.ActiveAlways`

* The action function is just a simple `Func: func() {` with no args -- use context capture of closures to access any relevant state.

* Use `ss.GUI.UpdateWindow()` inside any action function instead of `vp.SetNeedsFullRender()`

Here is a typical item:

```Go
    ss.GUI.AddToolbarItem(egui.ToolbarItem{Label: "Init", Icon: "update",
        Tooltip: "Initialize everything including network weights, and start over.  Also applies current params.",
        Active:  egui.ActiveStopped,
        Func: func() {
            ss.Init()
            ss.GUI.UpdateWindow()
        },
    })
```

For actions that take any significant amount of time, call the function in a separate routine using `go`, and use the GUI based variables:

```Go
    ss.GUI.AddToolbarItem(egui.ToolbarItem{Label: "Train",
        Icon:    "run",
        Tooltip: "Starts the network training, picking up from wherever it may have left off.  If not stopped, training will complete the specified number of Runs through the full number of Epochs of training, with testing automatically occuring at the specified interval.",
        Active:  egui.ActiveStopped,
        Func: func() {
            if !ss.GUI.IsRunning {
                ss.GUI.IsRunning = true
                ss.GUI.ToolBar.UpdateActions()
                go ss.Train()
            }
        },
    })
```

Here's an `ActiveRunning` case:

```Go
    ss.GUI.AddToolbarItem(egui.ToolbarItem{Label: "Stop",
        Icon:    "stop",
        Tooltip: "Interrupts running.  Hitting Train again will pick back up where it left off.",
        Active:  egui.ActiveRunning,
        Func: func() {
            ss.Stop()
        },
    })
```

## Spike Rasters

```Go
	stb := ss.GUI.TabView.AddNewTab(core.KiT_Layout, "Spike Rasters").(*core.Layout)
	stb.Lay = core.LayoutVert
	stb.SetStretchMax()
	for _, lnm := range ss.Stats.Rasters {
		sr := ss.Stats.F32Tensor("Raster_" + lnm)
		ss.GUI.ConfigRasterGrid(stb, lnm, sr)
	}
```    

## Tensor Grid (e.g., of an Image)

```Go
	tg := ss.GUI.TabView.AddNewTab(etview.KiT_TensorGrid, "Image").(*etview.TensorGrid)
	tg.SetStretchMax()
	ss.GUI.SetGrid("Image", tg)
	tg.SetTensor(&ss.TrainEnv.Img.Tsr)
```

## Activation-based Receptive Fields

```Go
	ss.GUI.AddActRFGridTabs(&ss.Stats.ActRFs)
```


