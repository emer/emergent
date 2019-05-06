# emergent reboot in Go

[![Go Report Card](https://goreportcard.com/badge/github.com/emer/emergent)](https://goreportcard.com/report/github.com/emer/emergent)
[![GoDoc](https://godoc.org/github.com/emer/emergent?status.svg)](https://godoc.org/github.com/emer/emergent)

This is the new home of the *emergent* neural network simulation software, developed primarily by the CCN lab at the University of Colorado Boulder.  We have decided to completely reboot the entire enterprise from the ground up, with a much more open, general-purpose design and approach.

See [Wiki Install](https://github.com/emer/emergent/wiki/Install) for installation instructions, and the [Wiki Rationale](https://github.com/emer/emergent/wiki/Rationale) and [History](https://github.com/emer/emergent/wiki/History) pages for a more detailed rationale for the new version of emergent, and a history of emergent (and its predecessors).

# Current Status / News

* 4/2019: separated the `leabra` and `etable` repositories from the overall `emergent` repository, to make it easier to fork and save / modify just the algorithm components of the system independent of the overall emergent infrastructure, and because `etable` (and associated `etensor` and `bitslice`) packages are fully independent and useful more generally.  This means that `emergent` is just a toolkit library with no runnable `examples` executables etc -- all of that has moved over to the `leabra` repository including the `python` wrapper.  You just need to replace "github.com/emer/emergent/leabra/leabra" -> "github.com/emer/leabra/leabra" in your imports, and likewise "github.com/emer/emergent/etable" -> "github.com/emer/etable/etable", "github.com/emer/emergent/etensor" -> "github.com/emer/etable/etensor".

* 3/2019: Python interface is up and running!  See the `python` directory in `leabra` for the [README](https://github.com/emer/leabra/blob/master/python/README.md) status and how to give it a try.  You can run the full `leabra/examples/ra25` code using Python, including the GUI etc.

* 2/2019: An initial complete basic-level implementation is now in place, and you can actually run `Leabra` models in the new emergent!  See the `examples/ra25` directory for a runnable, standalone Go program that you can compile use to train a "random associator" test model.  This is definitely the place to start in understanding how everything works.

* 2/2019: Initial benchmarking (see `examples/bench` for details) shows that the Go version is roughly 20-30% slower than C++ emergent for larger-sized models on a single processor, and while it does benefit significantly from multi-CPU processors, it does so less than the C++ version, which can be 2x faster than the Go version for some sizes and numbers of processors.  Nevertheless, we think the massive improvement in code simplicity and flexibility makes these performance tradeoffs worth it for most typical applications.

# Key Features

* Currently focused exclusively on implementing the biologically-based `Leabra` algorithm (now in a separate repository), which is not at all suited to implementation in current popular neural network frameworks such as `PyTorch`.  Leabra uses point-neurons and competitive inhibition, and has sparse activity levels and ubiquitous fully recurrent bidirectional processing, which enable / require novel optimizations for how simulated neurons communicate, etc.

* Go-based code can be compiled to run entire models.  Instead of creating and running everything in the *emergent* GUI, the process is much more similar to how e.g., PyTorch and other current frameworks work.  You write code to configure your model, and directly call functions that run your model, etc.  This gives you full, direct, transparent control over everything that happens in your model, as opposed to the previous highly opaque nature of C++ emergent.

* Although we will be updating our core library (`package` in Go) code with bug fixes, performance improvements, and new algorithms, we encourage users who have invested in developing a particular model to fork their own copy of the codebase and use that to maintain control over everything.  Once we make our official release of the code, the raw algorithm code is essentially guaranteed to remain fairly stable and encapsulated, so further changes should be relatively minimal, but nevertheless, it would be good to have an insurance policy!  The code is very compact and having your own fork should be very manageable.

* The `emergent` repository will host additional Go packages that provide support for models.  These are all designed to be usable as independently and optionally as possible.  Users running Leabra from Python for example will likely rely on relevant tools in that ecosystem instead.  An overview of some of those packages is provided below.

* We are committed to making the system fully usable from within Python, given the extensive base of Python users.  See the [leabra python README](https://github.com/emer/leabra/blob/master/python/README.md).  This includes interoperating with [PsyNeuLink](https://princetonuniversity.github.io/PsyNeuLink/) to make Leabra models accessible in that framework, and vice-versa.  Furthermore, interactive, IDE-level tools such as `Jupyter` and `nteract` can be used to interactively develop and analyze the models, etc. 

* We are leveraging the [GoGi Gui](https://github.com/goki/gi) to provide interactive 2D and 3D GUI interfaces to models, capturing the essential functionality of the original C++ emergent interface, but in a much more a-la-carte fashion.  We will also support the [GoNum](https://github.com/gonum) framework for analyzing and plotting results within Go.

# Design

* In general, *emergent* works by compiling programs into executables which you then run like any other executable. This is very different from the C++ version of emergent which was a single monolithic program attempting to have all functionality built-in. Instead, the new model is the more prevalent approach of writing more specific code to achieve more specific goals, which is more flexible and allows individuals to be more in control of their own destiny..
    + To make your own simulations, start with e.g., the `examples/ra25/ra25.go` code (or that of a more appropriate example) and copy that to your own repository, and edit accordingly.

* The `emergent` repository contains a collection of packages supporting the implementation of biologically-based neural networks.  The main package is `emer` which specifies a minimal abstract interface for a neural network.  The `etable` `etable.Table` data structure (DataTable in C++) is in a separate repository under the overall `emer` project umbrella, as are specific algorithms such as `leabra` which implement the `emer` interface.

* Go uses `interfaces` to represent abstract collections of functionality (i.e., sets of methods).  The `emer` package provides a set of interfaces for each structural level (e.g., `emer.Layer` etc) -- any given specific layer must implement all of these methods, and the structural containers (e.g., the list of layers in a network) are lists of these interfaces.  An interface is implicitly a *pointer* to an actual concrete object that implements the interface.  Thus, we typically need to convert this interface into the pointer to the actual concrete type, as in:

```Go
func (nt *Network) InitActs() {
	for _, ly := range nt.Layers {
		if ly.IsOff() {
			continue
		}
		ly.(*Layer).InitActs() // ly is the emer.Layer interface -- (*Layer) converts to leabra.Layer
	}
}
```

* The emer interfaces are designed to support generic access to network state, e.g., for the 3D network viewer, but specifically avoid anything algorithmic.  Thus, they should allow viewing of any kind of network, including PyTorch backprop nets.

* There are 3 main levels of structure: `Network`, `Layer` and `Prjn` (projection).  The network calls methods on its Layers, and Layers iterate over both `Neuron` data structures (which have only a minimal set of methods) and the `Prjn`s, to implement the relevant computations.  The `Prjn` fully manages everything about a projection of connectivity between two layers, including the full list of `Syanpse` elements in the connection.  There is no "ConGroup" or "ConState" level as was used in C++, which greatly simplifies many things.  The Layer also has a set of `Pool` elements, one for each level at which inhibition is computed (there is always one for the Layer, and then optionally one for each Sub-Pool of units (*Pool* is the new simpler term for "Unit Group" from C++ emergent).

* Layers have a `Shape` property, using the `etensor.Shape` type (see `etable` package), which specifies their n-dimensional (tensor) shape.  Standard layers are expected to use a 2D Y*X shape (note: dimension order is now outer-to-inner or *RowMajor* now), and a 4D shape then enables `Pools` ("unit groups") as hypercolumn-like structures within a layer that can have their own local level of inihbition, and are also used extensively for organizing patterns of connectivity.

# Packages

Here are some of the additional packages beyond the Leabra algorithm:

* `emer` is intended to hold all the widely-used support classes used in our algorithm-specific code, including things like `MinMax` and `AvgMax`, and also the parameter-styling infrastructure (`emer.Params`, `emer.ParamStyle`, `emer.ParamSet` and `emer.ParamSets`).

* `erand` has misc random-number generation support functionality, including `erand.RndParams` for parameterizing the type of random noise to add to a model, and easier support for making permuted random lists, etc.

* `prjn` is a separate package for defining patterns of connectivity between layers (i.e., the `ProjectionSpec`s from C++ emergent).  This is done using a fully independent structure that *only* knows about the shapes of the two layers, and it returns a fully general bitmap representation of the pattern of connectivity between them.  The `leabra.Prjn` code then uses these patterns to do all the nitty-gritty of connecting up neurons.  This makes the projection code *much* simpler compared to the ProjectionSpec in C++ emergent, which was involved in both creating the pattern and also all the complexity of setting up the actual connections themselves.  This should be the *last* time any of those projection patterns need to be written (having re-written this code too many times in the C++ version as the details of memory allocations changed).

* `patgen` supports various pattern-generation algorithms, as implemented in `taDataGen` in C++ emergent (e.g., `PermutedBinary` and `FlipBits`).

* `timer` is a simple interval timing struct, used for benchmarking / profiling etc.

* `python` contains a template `Makefile` that uses [GoPy](https://github.com/goki/gopy) to generate python bindings to the entire emergent system.  See the `leabra` package version to actually run an example.

