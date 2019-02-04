# emergent reboot in Go

[![Go Report Card](https://goreportcard.com/badge/github.com/emer/emergent)](https://goreportcard.com/report/github.com/emer/emergent)
[![GoDoc](https://godoc.org/github.com/emer/emergent?status.svg)](https://godoc.org/github.com/emer/emergent)

This is the new home of the *emergent* neural network simulation software, developed primarily by the CCN lab at the University of Colorado Boulder.  We have decided to completely reboot the entire enterprise from the ground up, with a much more open, general-purpose design and approach.

See the [Wiki Rationale](https://github.com/emer/emergent/wiki/Rationale) and [History](https://github.com/emer/emergent/wiki/History) pages for a more detailed rationale for the new version of emergent, and a history of emergent (and its predecessors).

# Current Status / News

* 2/2019: An initial complete basic-level implementation is now in place, and you can actually run `Leabra` models in the new emergent!  See the `examples/leabra25ra` directory for a runnable, standalone Go program that you can compile use to train a "random associator" test model.  This is definitely the place to start in understanding how everything works.

* 2/2019: Initial benchmarking (see `examples/bench` for details) shows that the Go version is roughly 20-30% slower than C++ emergent for larger-sized models on a single processor, and while it does benefit significantly from multi-CPU processors, it does so less than the C++ version, which can be 2x faster than the Go version for some sizes and numbers of processors.  Nevertheless, we think the massive improvement in code simplicity and flexibility makes these performance tradeoffs worth it for most typical applications.

# Key Features

* Currently focused exclusively on implementing the biologically-based `Leabra` algorithm, which is not at all suited to implementation in current popular neural network frameworks such as `PyTorch`.  Leabra uses point-neurons and competitive inhibition, and has sparse activity levels and ubiquitous fully recurrent bidirectional processing, which enable / require novel optimizations for how simulated neurons communicate, etc.

* Go-based code can be compiled to run entire models.  Instead of creating and running everything in the *emergent* GUI, the process is much more similar to how e.g., PyTorch and other current frameworks work.  You write code to configure your model, and directly call functions that run your model, etc.  This gives you full, direct, transparent control over everything that happens in your model, as opposed to the previous highly opaque nature of C++ emergent.

* We provide clean, well-organized implementations of core Leabra algorithms and Network structures, starting with a `basic/leabra` version that only implements the core elements of Leabra activation and learning.  More specialized modifications such as `DeepLeabra` or `PBWM` or `PVLV` are all (going to be) implemented as additional specialized code that builds on / replaces elements of the basic version.  The goal is to make all of the code simpler, more transparent, and more easily modified by end users.  You should not have to dig through endless chains of C++ inheritance to find out what is going on.

* Although we will be updating our core library (`package` in Go) code with bug fixes, performance improvements, and new algorithms, we encourage users who have invested in developing a particular model to fork their own copy of the codebase and use that to maintain control over everything.  Once we make our official release of the code, the raw algorithm code is essentially guaranteed to remain fairly stable and encapsulated, so further changes should be relatively minimal, but nevertheless, it would be good to have an insurance policy!  The code is very compact and having your own fork should be very manageable.

* In addition to the core algorithms, the `emergent` repository will host many additional Go packages that provide support for models.  These are all designed to be usable as independently and optionally as possible.  Users running Leabra from Python for example will likely rely on relevant tools in that ecosystem instead.  An overview of some of those packages is provided below.

* We are committed to making the system fully usable from within Python, given the extensive base of Python users.  Details on this forthcoming soon.  This includes interoperating with [PsyNeuLink](https://princetonuniversity.github.io/PsyNeuLink/) to make Leabra models accessible in that framework, and vice-versa.  Furthermore, interactive, IDE-level tools such as `Jupyter` and `nteract` can be used to interactively develop and analyze the models, etc. 

* We will be leveraging the [GoGi Gui](https://github.com/goki/gi) to provide interactive 2D and 3D GUI interfaces to models, capturing the essential functionality of the original C++ emergent interface, but in a much more a-la-carte fashion.  We will also support the [GoNum](https://github.com/gonum) framework for analyzing and plotting results within Go.

# Design

* `ActParams` (in `act.go`), `InhibParams` (in `inhib.go`), and `LearnNeurParams` / `LearnSynParams` (in `learn.go`) provide the core parameters and functions used, including the X-over-X-plus-1 activation function, FFFB inhibition, and the XCal BCM-like learning rule, etc.  This function-based organization should be clearer than the purely structural organization used in C++ emergent.

* There are 3 main levels of structure: `Network`, `Layer` and `Prjn` (projection).  The network calls methods on its Layers, and Layers iterate over both `Neuron` data structures (which have only a minimal set of methods) and the `Prjn`s, to implement the relevant computations.  The `Prjn` fully manages everything about a projection of connectivity between two layers, including the full list of `Syanpse` elements in the connection.  There is no "ConGroup" or "ConState" level as was used in C++, which greatly simplifies many things.  The Layer also has a set of `Pool` elements, one for each level at which inhibition is computed (there is always one for the Layer, and then optionally one for each "Unit Group" within that).

* The `NetworkStru` and `LayerStru` structs manage all the core structural aspects of things (data structures etc), and then the algorithm-specific versions (e.g., `leabra.Network`) use Go's anonymous embedding (akin to inheritance in C++) to transparently get all that functionality, while then directly implementing the algorithm code.  Almost every step of computation has an associated method in `leabra.Layer`, so look first in `basic/leabra/layer.go` to see how something is implemented.  

* Each structural element directly has all the parameters controlling its behavior -- e.g., the `Layer` contains an `ActParams` field (named `Act`), etc, instead of using a separate `Spec` structure as in C++ emergent.  The Spec-like ability to share parameter settings across multiple layers etc is instead achieved through a **styling**-based paradigm -- you apply parameter "styles" to relevant layers instead of assigning different specs to them.  This paradigm should be less confusing and less likely to result in accidental or poorly-understood parameter applications.  We adopt the CSS (cascading-style-sheets) standard where parameters can be specifed in terms of the Name of an object (e.g., `#Hidden`), the *Class* of an object (e.g., `.TopDown` -- where the class name TopDown is manually assigned to relevant elements), and the *Type* of an object (e.g., `Layer` applies to all layers).  Multiple space-separated classes can be assigned to any given element, enabling a powerful combinatorial styling strategy to be used.

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

* The emer interfaces are designed to support generic access to network state, e.g., for the 3D network viewer, but specifically avoid anything algorithmic.  Thus, they should allow viewing of any kind of network, including PyTorch backprop nets.  We will likely be introducing an additional algorithm-specific set of interfaces that provide suitable abstractions over the broad computational stages of processing, to enable more specialized algorithm variants to be more easily "plugged into" a heterogenous network.  This is current work-in-progress as we start implementing these variants.

* Layers have a `Shape` property, using the `etensor.Shape` type (see below), which specifies their n-dimensional (tensor) shape.  Standard layers are expected to use a 2D Y*X shape (note: dimension order is now outer-to-inner or *RowMajor* now), and a 4D shape then enables "unit groups" as hypercolumn-like structures within a layer that can have their own local level of inihbition, and are also used extensively for organizing patterns of connectivity.

# Packages

Here are some of the additional packages beyond the Leabra algorithm:

* `emer` is intended to hold all the widely-used support classes used in our algorithm-specific code, including things like `MinMax` and `AvgMax`, and also the parameter-styling infrastructure (`emer.Params`, `emer.ParamStyle`, `emer.ParamSet` and `emer.ParamSets`).

* `erand` has misc random-number generation support functionality, including `erand.RndParams` for parameterizing the type of random noise to add to a model, and easier support for making permuted random lists, etc.

* `prjn` is a separate package for defining patterns of connectivity between layers (i.e., the `ProjectionSpec`s from C++ emergent).  This is done using a fully independent structure that *only* knows about the shapes of the two layers, and it returns a fully general bitmap representation of the pattern of connectivity between them.  The `leabra.Prjn` code then uses these patterns to do all the nitty-gritty of connecting up neurons.  This makes the projection code *much* simpler compared to the ProjectionSpec in C++ emergent, which was involved in both creating the pattern and also all the complexity of setting up the actual connections themselves.  This should be the *last* time any of those projection patterns need to be written (having re-written this code too many times in the C++ version as the details of memory allocations changed).

* `etensor` is our own implementation of a Tensor object, which corresponds to the `Matrix` type in C++ emergent.  `etensor.Tensor` is an interface that applies to many different type-specific instances, such as `etensor.Float32`.  A tensor is just a `etensor.Shape` plus a slice holding the specific data type.  Our tensor is based directly on the [Apache Arrow](https://github.com/apache/arrow/tree/master/go) project's tensor, and it fully interoperates with it.  Arrow tensors are designed to be read-only, and we needed some extra support to make our `dtable.Table` work well, so we had to roll our own.  Our tensors will also interoperate fully with Gonum's 2D-specific Matrix type.

* `dtable` is our Go version of `DataTable` from C++ emergent, which is widely useful for holding input patterns to present to the network, and logs of output from the network, among many other uses.  A `dtable.Table` is a collection of `etensor.Tensor` columns, that are all aligned along the outer-most *row* dimension.  We are keeping the index-based indirection outside of the core type, which greatly simplifies many things.  The `dtable.Table` should interoperate with the under-development gonum `DataFrame` structure among others.  The use of this data structure is always optional and orthogonal to the core network algorithm code -- in Python the `pandas` library has a suitable `DataFrame` structure that can be used instead.

* `bitslice` is a Go slice of bytes `[]byte` that has methods for setting individual bits, as if it was a slice of bools, while being 8x more memory efficient.  This is used in `prjn` for representing the pattern of connectivity, for encoding null entries in  `etensor`, and as a Tensor of bool / bits there as well.

* `patgen` supports various pattern-generation algorithms, as implemented in `taDataGen` in C++ emergent (e.g., `PermutedBinary` and `FlipBits`).

* `timer` is a simple interval timing struct, used for benchmarking / profiling etc.

