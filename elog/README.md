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

Standard `Times` are based on the [Env](https://github.com/emer/emergent/wiki/Env) `TimeScales` augmented with Leabra / Axon finer-grained scales, including:
* `Cycle`
* `Trial`
* `Epoch`
* `Run`

Other arbitrary scope values can be used -- there are `Scope` versions of every method that take an arbitrary `ScopeKey` that can be composed using the `ScopeStr` method from any two strings, along with the "plain" versions of these methods that take the standard `mode` and `time` enums for convenience.  These enums can themselves also be extended but it is probably easier to just use strings.

