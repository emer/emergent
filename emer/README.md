Docs: [GoDoc](https://pkg.go.dev/github.com/emer/emergent/emer)

Package emer provides minimal interfaces for the basic structural elements of neural networks
including:
* emer.Network, emer.Layer, emer.Unit, emer.Prjn (projection that interconnects layers)

These interfaces are intended to be just sufficient to support visualization and generic
analysis kinds of functions, but explicitly avoid exposing ANY of the algorithmic aspects,
so that those can be purely encoded in the implementation structs.

At this point, given the extra complexity it would require, these interfaces do not support
the ability to build or modify networks.

