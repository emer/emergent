Docs: [GoDoc](https://pkg.go.dev/github.com/emer/emergent/etime)

Everything is organized according to a `ScopeKey`, which is just a `string`, that is formatted to represent two factors: an **evaluation mode** (standard versions defined by `Modes` enum) and a **time scale** (`Times` enum).

Standard evaluation `Modes` are:
* `Train`
* `Test`
* `Validate`
* `Analyze` -- used for internal representational analysis functions such as PCA, ActRF, SimMat, etc.

Standard `Times` are based on the [Env](https://github.com/emer/emergent/wiki/Env) `TimeScales` augmented with Leabra / Axon finer-grained scales, including:
* `Run`
* `Epoch`
* `Trial`
* `Phase`
* `Cycle`

Other arbitrary scope values can be used -- there are `Scope` versions of every method that take an arbitrary `ScopeKey` that can be composed using the `ScopeStr` method from any two strings, along with the "plain" versions of these methods that take the standard `mode` and `time` enums for convenience.  These enums can themselves also be extended but it is probably easier to just use strings.


