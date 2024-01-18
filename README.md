# emergent reboot in Go

[![Go Report Card](https://goreportcard.com/badge/github.com/emer/emergent)](https://goreportcard.com/report/github.com/emer/emergent)
[![Go Reference](https://pkg.go.dev/badge/github.com/emer/emergent.svg)](https://pkg.go.dev/github.com/emer/emergent)
[![CI](https://github.com/emer/emergent/actions/workflows/ci.yml/badge.svg)](https://github.com/emer/emergent/actions/workflows/ci.yml)
[![Codecov](https://codecov.io/gh/emer/emergent/branch/master/graph/badge.svg?token=Hw5cInAxY3)](https://codecov.io/gh/emer/emergent)

This is the new home of the *emergent* neural network simulation software, developed primarily by the CCN lab, originally at CU Boulder, and now at UC Davis: https://ccnlab.org.  We have decided to completely reboot the entire enterprise from the ground up, with a much more open, general-purpose design and approach.

See [Wiki Install](https://github.com/emer/emergent/wiki/Install) for installation instructions (note: Go 1.18 required), and the [Wiki Rationale](https://github.com/emer/emergent/wiki/Rationale) and [History](https://github.com/emer/emergent/wiki/History) pages for a more detailed rationale for the new version of emergent, and a history of emergent (and its predecessors).

The single clearest motivation for using Go vs. the ubiquitous Python, is that Python is too slow to implement the full computational model: it can only serve as a wrapper around backend code which is often written in C or C++.  By contrast, *Go can implement the entire model* in one coherent language.  This, along with the advantages of the strongly typed, rapidly compiled Go language vs. duck typed Python for developing large scale frameworks, and the many other benefits of the Go language environment for reproducible, reliable builds across platforms, results in a satisfying and productive programming experience.  Note also that we are not able to leverage existing Python backends such as PyTorch or TensorFlow due to the unique demands of biologically-based neural models.

See the [ra25 example](https://github.com/emer/leabra/blob/master/examples/ra25/README.md) in the `leabra` package for a complete working example (intended to be a good starting point for creating your own models), and any of the 26 models in the [Comp Cog Neuro sims](https://github.com/CompCogNeuro/sims) repository which also provide good starting points.  See the [etable wiki](https://github.com/emer/etable/v2/wiki) for docs and example code for the widely-used etable data table structure, and the `family_trees` example in the CCN textbook sims which has good examples of many standard network representation analysis techniques (PCA, cluster plots, RSA).

See [python README](https://github.com/emer/leabra/blob/master/python/README.md) and [Python Wiki](https://github.com/emer/emergent/wiki/Python) for info on using Python to run models.  See [eTorch](https://github.com/emer/etorch) for how to get an interactive 3D NetView for PyTorch models.

# Current Status / News

* Nov 2020: Full Python conversions of CCN sims complete, and [eTorch](https://github.com/emer/etorch) for viewing and interacting with PyTorch models.

* April 2020: Version 1.0 of GoGi GUI is now released, and we have updated all module dependencies accordingly. *We now recommend using the go modules instead of GOPATH* -- the [Wiki Install](https://github.com/emer/emergent/wiki/Install) instructions have been updated accordingly.

* 12/30/2019: Version 1.0.0 released!  The [Comp Cog Neuro sims](https://github.com/CompCogNeuro/sims) that accompany the [CCN Textbook](https://github.com/CompCogNeuro/ed4) are now complete and have driven extensive testing and bugfixing.

* 3/2019: Python interface is up and running!  See the `python` directory in `leabra` for the [README](https://github.com/emer/leabra/blob/master/python/README.md) status and how to give it a try.  You can run the full `leabra/examples/ra25` code using Python, including the GUI etc.

* 2/2019: Initial implementation and benchmarking (see `examples/bench` for details -- shows that the Go version is comparable in speed to C++).

# Key Features

* Currently focused exclusively on implementing the biologically-based `Leabra` algorithm, which is not at all suited to implementation in current popular neural network frameworks such as `PyTorch`.  Leabra uses point-neurons and competitive inhibition, and has sparse activity levels and ubiquitous fully recurrent bidirectional processing, which enable / require novel optimizations for how simulated neurons communicate, etc.

* Go-based code can be compiled to run entire models.  Instead of creating and running everything in the *emergent* GUI, the process is much more similar to how e.g., PyTorch and other current frameworks work.  You write code to configure your model, and directly call functions that run your model, etc.  This gives you full, direct, transparent control over everything that happens in your model, as opposed to the previous highly opaque nature of [C++ emergent](https://github.com/emer/cemer).

* Although we will be updating our core library (`package` in Go) code with bug fixes, performance improvements, and new algorithms, we encourage users who have invested in developing a particular model to fork their own copy of the codebase and use that to maintain control over everything.  Once we make our official release of the code, the raw algorithm code is essentially guaranteed to remain fairly stable and encapsulated, so further changes should be relatively minimal, but nevertheless, it would be good to have an insurance policy!  The code is very compact and having your own fork should be easily manageable.

* The `emergent` repository will host additional Go packages that provide support for models.  These are all designed to be usable as independently and optionally as possible.  An overview of some of those packages is provided below.

* The system is fully usable from within Python -- see the [Python Wiki](https://github.com/emer/emergent/wiki/Python).  This includes interoperating with PyTorch via [eTorch](https://github.com/emer/etorch), and [PsyNeuLink](https://princetonuniversity.github.io/PsyNeuLink/) to make Leabra models accessible in that framework, and vice-versa.  Furthermore, interactive, IDE-level tools such as `Jupyter` and `nteract` can be used to interactively develop and analyze the models, etc.

* We are leveraging the [GoGi Gui](https://cogentcore.org/core/gi/v2) to provide interactive 2D and 3D GUI interfaces to models, capturing the essential functionality of the original C++ emergent interface, but in a much more a-la-carte fashion.  We also use and support the [GoNum](https://github.com/gonum) framework for analyzing and plotting results within Go.

# Design / Organization

* The `emergent` repository contains a collection of packages supporting the implementation of biologically-based neural networks.  The main package is `emer` which specifies a minimal abstract interface for a neural network.  The `etable` `etable.Table` data structure (DataTable in C++) is in a separate repository under the overall `emer` project umbrella, as are specific algorithms such as `leabra` which implement the `emer` interface.

* Go uses `interfaces` to represent abstract collections of functionality (i.e., sets of methods).  The `emer` package provides a set of interfaces for each structural level (e.g., `emer.Layer` etc) -- any given specific layer must implement all of these methods, and the structural containers (e.g., the list of layers in a network) are lists of these interfaces.  An interface is implicitly a *pointer* to an actual concrete object that implements the interface.

* To allow for specialized algorithms to extend the basic Leabra algorithm functionality, we have additional algorithm-specific interfaces in `leabra/leabra/leabra.go`, called `LeabraNetwork`, `LeabraLayer`, and `LeabraPrjn` -- all functions should go through this interface so that the final actual function called can be either the default version defined on `leabra.Layer` or a more specialized type (e.g., for simulating the PFC, hippocampus, BG etc). This is what it looks like for example:

```Go
func (nt *Network) InitActs() {
	for _, ly := range nt.Layers {
		if ly.IsOff() {
			continue
		}
		ly.(LeabraLayer).InitActs() // ly is the emer.Layer interface -- convert to (LeabraLayer) interface
	}
}
```

* The emer interfaces are designed to support generic access to network state, e.g., for the 3D network viewer, but specifically avoid anything algorithmic.  Thus, they allow viewing of any kind of network, including PyTorch backprop nets in the eTorch package.

* There are 3 main levels of structure: `Network`, `Layer` and `Prjn` (projection).  The Network calls methods on its Layers, and Layers iterate over both `Neuron` data structures (which have only a minimal set of methods) and the `Prjn`s, to implement the relevant computations.  The `Prjn` fully manages everything about a projection of connectivity between two layers, including the full list of `Syanpse` elements in the connection.  There is no "ConGroup" or "ConState" level as was used in C++, which greatly simplifies many things.  The Layer also has a set of `Pool` elements, one for each level at which inhibition is computed (there is always one for the Layer, and then optionally one for each Sub-Pool of units (*Pool* is the new simpler term for "Unit Group" from C++ emergent).

* Layers have a `Shape` property, using the `etensor.Shape` type (see `etable` package), which specifies their n-dimensional (tensor) shape.  Standard layers are expected to use a 2D Y*X shape (note: dimension order is now outer-to-inner or *RowMajor* now), and a 4D shape then enables `Pools` ("unit groups") as hypercolumn-like structures within a layer that can have their own local level of inihbition, and are also used extensively for organizing patterns of connectivity.

# Packages

Here are some of the additional supporting packages, organized by overall functionality:

## Core Network

* [emer](emer) *only* has the primary abstract Network interfaces.

* [params](params) has the parameter-styling infrastructure (e.g., `params.Set`, `params.Sheet`, `params.Sel`), which implement a powerful, flexible, and efficient CSS style-sheet approach to parameters.  See the [Wiki Params](https://github.com/emer/emergent/wiki/Params) page for more info.

* [netview](netview) provides the `NetView` interactive 3D network viewer, implemented in the GoGi 3D framework.

* [prjn](prjn) is a separate package for defining patterns of connectivity between layers (i.e., the `ProjectionSpec`s from C++ emergent).  This is done using a fully independent structure that *only* knows about the shapes of the two layers, and it returns a fully general bitmap representation of the pattern of connectivity between them.  The `leabra.Prjn` code then uses these patterns to do all the nitty-gritty of connecting up neurons.  This makes the projection code *much* simpler compared to the ProjectionSpec in C++ emergent, which was involved in both creating the pattern and also all the complexity of setting up the actual connections themselves.  This should be the *last* time any of those projection patterns need to be written (having re-written this code too many times in the C++ version as the details of memory allocations changed).

* [relpos](relpos) provides relative positioning of layers (right of, above, etc).

* [weights](weights) provides weight-file parsing / loading routines: much easier to read into a temporary structure and then apply to the network.

## Environment: input / output patterns

* [env](env) has an interface for environments, which encapsulates all the counters and timing information for patterns that are presented to the network, and enables more of a mix-and-match ability for using different environments with different networks.  See [Wiki Env](https://github.com/emer/emergent/wiki/Env) page for more info, and the [envs](https://github.com/emer/envs) repository for various specialized environments that can be a good starting point.

* [patgen](patgen) supports various general-purpose pattern-generation algorithms, as implemented in `taDataGen` in C++ emergent (e.g., `PermutedBinary` and `FlipBits`).

## Running, Logging, Stats, GUI toolkit

The following all work together to provide a convenient layer of abstraction for running, logging & statistics, and the GUI interface:

* [etime](time) provides the core time scales and training / testing etc modes used in the rest of the packages.

* [looper](looper) provides a fully step-able hierarchical looping control framework (e.g., Run / Epoch / Trial / Cycle) where you can insert functions to run at different points (start, end, or specific counter value).  [Axon](https;//github.com/emer/axon) uses this natively and has support functions for configuring standard looper functions.

* [estats](estats) manages statistics as maps of name, value for various types, along with network-relevant statistics such as `ClosestPat`, PCA stats, Cluster plots, decoders, raster plots, etc.

* [elog](elog) has comprehensive support for logging data at different time scales and evaluation modes -- saves a lot of boilerplate code for configuring, updating.

* [egui](egui) implements a standard simulation GUI, with a toolbar, tabs of different views, and a Sim struct view on the left.

* [ecmd](ecmd) manages command-line args and standard defaults / methdods around these.

## Other Misc

* [actrf](actrf) provides activation-based receptive field stats (reverse correlation, spike-triggered averaging) for decoding internal representations.

* [chem](chem) provides basic chemistry simulation mechanisms for chemical reactions characterized by rate constants and concentrations, including diffusion.  This can be used for detailed biochemical models of neural function, as in the [Urakubo et al (2008)](https://github.com/ccnlab/kinase/sims/urakubo) model of synaptic plasticity.

* [confusion](confusion) provides confusion matricies for model output vs. target output.

* [decoder](decoder) provides simple linear, sigmoid, and softmax decoders for interpreting network activity states according to hypothesized variables of interest.

* [efuns](efuns) has misc special functions such as Gaussian and Sigmoid.

* [erand](erand) has misc random-number generation support functionality, including `erand.RndParams` for parameterizing the type of random noise to add to a model, and easier support for making permuted random lists, etc.

* [esg](esg) is the *emergent stochastic / sentence generator* -- parses simple grammars that generate random events (sentences) -- can be a good starting point for generating more complex environments.

* [evec](evec) has `Vec2i` which uses plain `int` X, Y fields, whereas the `mat32` package uses `int32` which are needed for graphics but int is more convenient in models.

* [popcode](popcode) supports the encoding and decoding of population codes -- distributed representations of numeric quantities across a population of neurons.  This is the `ScalarVal` functionality from C++ emergent, but now completely independent of any specific algorithm so it can be used anywhere.

* [ringidx](ringidx) provides a wrap-around ring index for efficient use of a fixed buffer that overwrites the oldest items without any copying.

* [stepper](stepper) provides dynamic stepping control at multiple levels -- used in `pvlv` model (contributed by Randy Gobbel).  This functionality is now available in [looper](looper) in a more robust and integrated form.

* [timer](timer) is a simple interval timing struct, used for benchmarking / profiling etc.

# Repositories

Here are the other repositories within `emer` that provide additional, optional elements for simulations:

* [etable](https://github.com/emer/etable/v2) repository holds all of the more general-purpose "DataTable" or DataFrame (`etable.Table`) related code, which is our version of something like `pandas` or `xarray` in Python.  This includes the `etensor` n-dimensional array, `eplot` for interactive plotting of data, and basic utility packages like `minmax` and `bitslice`, and lots of data analysis tools like similarity / distance matricies, PCA, cluster plots, etc.

* [eMPI](https://github.com/emer/empi) provides an MPI (message passing interface) distributed memory implementation -- see [MPI Wiki page](https://github.com/emer/emergent/wiki/MPI)

* [envs](https://github.com/emer/envs) has misc standalone environments that can be good starting points, including managing files, visual images, etc.

* [etail](https://github.com/emer/etail) is the emergent `tail` program -- a separate command-line tool for looking at tabular (csv, tsv, etc) log files from simulations -- very useful!  `go get github.com/emer/etail` from anywhere to install (in go modules mode).

* [eTorch](https://github.com/emer/etorch) is the emergent interface to PyTorch models, providing emergent GUI NetView etc for these models.

* [eve](https://github.com/emer/eve) is the emergent virtual environment -- provides a physics engine and collision detection that interfaces with the GoGi 3D for visualization.  For constructing more realistic environments for your models.

* [grunt](https://github.com/emer/grunt) is the git-based run tool -- it handles the grunt work for running simulations on a cluster, by pushing to git repositories hosted on the cluster, which has a daemon running on it monitoring for these git updates.  It pushes back updates and results from the cluster.  There is a GUI for controlling and managing a potentially large history of jobs -- invaluable for any significant simulation to keep track of various parameter searches, changes over time etc.

* [vision](https://github.com/emer/vision) and [auditory](https://github.com/emer/auditory) provide low-level filtering on sensory inputs reflecting corresponding biological mechanisms.

* [vgpu](https://github.com/goki/vgpu) is a GPU library using Vulkan for both graphics and compute functionality.  Used now in [Axon](https;//github.com/emer/axon), via [gosl](https://github.com/goki/gosl) which is a Go shader language that converts Go -> HLSL shader code that can then be compiled and run in the VGPU framework.

