Docs: [GoDoc](https://pkg.go.dev/github.com/emer/emergent/estats)

`estats.Stats` provides maps for storing statistics as named scalar and tensor values.  These stats are available in the `elog.Context` for use during logging -- see [elog Docs](https://pkg.go.dev/github.com/emer/emergent/elog).

To make relevant stats visible to users, call the `Print` function with a list of stat names -- this can be passed to the `Netview` Record method to show these stats at the bottom of the network view, and / or displayed in a Sims field.

There are 3 main data types supported: `Float` (`float64`), `String`, and `Int`. The Float interface to Tables uses float64 so for simple scalar values, it is simpler to just use the float64 instead of also supporting float32.  However, for Tensor data, network data is often float32 so we have `F32Tensor` and `F64Tensor` for `float32` and `float64` respectively.

There are also various utility functions for computing various useful statistics.

# Examples

A common use-case for example is to use `F32Tensor` to manage a tensor that is reused every time you need to access values on a given layer (this was commonly named `ValsTsr` in existing Sims):

```Go
    ly := ctxt.Net.LayerByName(lnm)
    tsr := ctxt.Stats.F32TEnsorr(lnm)
    ly.UnitValsTensor(tsr, "Act")
    // tsr now has the "Act" values from given layer -- can be logged, computed on, etc..
```

The above also now available as a convenience function named `SetLayerTensor` (also present in `elog.Context`).

# Stats functions

* `SetLayerTensor` does the above storing of unit values to a tensor.

* `ClosestPat` finds the closest pattern in given column of given table of possible patterns, based on unit layer activations from `SetLayerTensor`

* `PCAStats` computes PCA (principal components analysis) statistics on activity patterns in a table -- Helpful for measuring the overall information (variance) in the representations, to detect a common failure mode where a few patterns dominate over everything ("hogs").

* `Raster` functions store raster-based tensor data with X axis = time and Y axis = unit values.

