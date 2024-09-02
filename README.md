# Emergent Neural Network Simulation Framework

[![Go Report Card](https://goreportcard.com/badge/github.com/emer/emergent)](https://goreportcard.com/report/github.com/emer/emergent)
[![Go Reference](https://pkg.go.dev/badge/github.com/emer/emergent.svg)](https://pkg.go.dev/github.com/emer/emergent)
[![CI](https://github.com/emer/emergent/actions/workflows/ci.yml/badge.svg)](https://github.com/emer/emergent/actions/workflows/ci.yml)
[![Codecov](https://codecov.io/gh/emer/emergent/branch/main/graph/badge.svg?token=Hw5cInAxY3)](https://codecov.io/gh/emer/emergent)

The *emergent* neural network simulation framework provides a toolkit in the Go programming language (golang) for developing neural network models across multiple levels of complexity, from biologically-detailed spiking networks in the [axon](https://github.com/emer/axon) package, to PyTorch abstract deep networks in the [eTorch](https://github.com/emer/etorch) package.  It builds on the [Cogent Core](https://cogentcore.org/core) GUI framework to provide dynamic graphical interfaces for visualizing and manipulating networks and data, making the models uniquely accessible for teaching (e.g., see the [Computational Cognitive Neuroscience](https://github.com/CompCogNeuro/sims) simulations) and supporting the development of complex dynamical models for research.

See [cogent core install](https://www.cogentcore.org/core/setup/install) instructions for general installation instructions. 

The [Wiki Rationale](https://github.com/emer/emergent/wiki/Rationale) and [History](https://github.com/emer/emergent/wiki/History) pages for a more detailed rationale for this version of emergent, and a history of emergent (and its predecessors).  The Wiki tends to be a bit out of date, but can have some useful information.  In general it is best to take the plunge and "use the source" directly :)

The single clearest motivation for using Go vs. the ubiquitous Python, is that Python is too slow to implement the full computational model: it can only serve as a wrapper around backend code which is often written in C or C++.  By contrast, *Go can implement the entire model* in one coherent language.  This, along with the advantages of the strongly typed, rapidly compiled Go language vs. duck typed Python for developing large scale frameworks, and the many other benefits of the Go language environment for reproducible, reliable builds across platforms, results in a satisfying and productive programming experience.

Furthermore, the _Go shader language_ [gosl](https://github.com/cogentcore/core/tree/main/vgpu/gosl) in Cogent Core enables Go to run efficiently on the GPU as well, enabling the same code base to be used for both CPU and GPU execution.  This enables even very complex, biologically-detailed models as in the [axon](https://github.com/emer/axon) framework to take full advantage of GPU acceleration, resulting in 10x or more speedup factors over CPU.

See the [ra25 example](https://github.com/emer/axon/tree/main/ra25/README.md) in the [axon](https://github.com/emer/axon) package for a complete working example (intended to be a good starting point for creating your own models), and any of the 26 models in the [Comp Cog Neuro sims](https://github.com/CompCogNeuro/sims) repository which also provide good starting points, using the [leabra](https://github.com/emer/leabra) framework. 

# Current Status / News

* June 2024: Major changes ongoing in coordination with the [Cogent Core](https://cogentcore.org/core) framework development over the prior year, replacing the previous [GoKi](https://github.com/goki) GUI framework.  Many packages have migrated to Cogent Core, which is a much cleaner major rewrite, which should be stable and released in beta status soon.  [axon](https://github.com/emer/axon) is staying updated but everything else should use the [v1](https://github.com/emer/emergent/tree/v1) branch. [Leabra](https://github.com/emer/leabra) and [Comp Cog Neuro sims](https://github.com/CompCogNeuro/sims) will be updated after the dust settles, hopefully in time for Fall 2024 classes.

* Nov 2020: Full Python conversions of CCN sims complete, and [eTorch](https://github.com/emer/etorch) for viewing and interacting with PyTorch models.

* April 2020: Version 1.0 of GoGi GUI is now released, and we have updated all module dependencies accordingly. *We now recommend using the go modules instead of GOPATH* -- the [Wiki Install](https://github.com/emer/emergent/wiki/Install) instructions have been updated accordingly.

* 12/30/2019: Version 1.0.0 released!  The [Comp Cog Neuro sims](https://github.com/CompCogNeuro/sims) that accompany the [CCN Textbook](https://github.com/CompCogNeuro/book) are now complete and have driven extensive testing and bugfixing.

* 3/2019: Python interface is up and running!  See the `python` directory in `leabra` for the [README](https://github.com/emer/leabra/blob/main/python/README.md) status and how to give it a try.  You can run the full `leabra/examples/ra25` code using Python, including the GUI etc.

* 2/2019: Initial implementation and benchmarking (see `examples/bench` for details -- shows that the Go version is comparable in speed to C++).

# Design / Organization

* The `emergent` repository contains a collection of packages supporting the implementation of biologically based neural networks.  The main package is `emer` which specifies a minimal abstract interface for a neural network.

* Go uses `interfaces` to represent abstract collections of functionality (i.e., sets of methods).  The `emer` package provides a set of _minimal_ interfaces for each structural level (e.g., `emer.Layer` etc), along with concrete "Base" types that implement a lot of shared functionality (e.g., `emer.LayerBase`), which are available as `AsEmer()` from the interface.  Each algorithm must implement the interface methods to support the Network view, logging, parameter setting and other shared emergent functionality.

* The emer interfaces are designed to support generic access to network state, e.g., for the 3D network viewer, but specifically avoid anything algorithmic or structural, so that most of the algorithm-specific code uses direct slices and methods that return algorithm-specific types.

* There are 3 main levels of structure: `Network`, `Layer` and `Path` (pathway of connectivity between layers, also known as a Projection).  The Network typically calls methods on its Layers, and Layers have methods to access Neuron or Unit-level data.  The `Path` fully manages everything about a pathway of connectivity between two layers, including access to Synapse or Connection level state.

* Layers have a `Shape` property, using the `tensor.Shape` type (from the Cogent Core [tensor](https://github.com/cogentcore/core/tree/main/tensor) package), which specifies their n-dimensional (tensor) shape.  Standard layers are expected to use a 2D Y*X shape (note: dimension order is outer-to-inner or *RowMajor*), and a 4D shape that enables `Pools` as hypercolumn-like structures within a layer that can have their own local level of inihbition, and are also used extensively for organizing patterns of connectivity.

# Packages

Here are some of the additional supporting packages, organized by overall functionality:

## Core Network

* [emer](emer): the primary abstract `Network`, `Layer`, `Path` interfaces.

* [params](params): a parameter-styling infrastructure (e.g., `params.Set`, `params.Sheet`, `params.Sel`), which implement a powerful, flexible, and efficient CSS style-sheet approach to parameters.  See the [Wiki Params](https://github.com/emer/emergent/wiki/Params) page for more info.

* [netview](netview): the `NetView` interactive 3D network viewer, implemented in the Cogent Core [xyz](https://github.com/cogentcore/core/tree/main/xyz) 3D framework.

* [paths](paths) is a separate package for defining patterns of connectivity between layers.  This is done using a fully independent structure that *only* knows about the shapes of the two layers, and it returns a fully general bitmap representation of the pattern of connectivity between them.  The algorithm-specific code then uses these patterns to do all the nitty-gritty of connecting up neurons.  This makes the pathway code *much* simpler compared to earlier implementations that combined both of these functions.

* [relpos](relpos) provides relative positioning of layers (right of, above, etc).

* [weights](weights) provides weight-file parsing / loading routines: much easier to read into a temporary structure and then apply to the network.

## Environment: input / output patterns

* [env](env) has an interface for environments, which encapsulates all the counters and timing information for patterns that are presented to the network, and enables more of a mix-and-match ability for using different environments with different networks.  See [Wiki Env](https://github.com/emer/emergent/wiki/Env) page for more info, and the [envs](https://github.com/emer/envs) repository for various specialized environments that can be a good starting point.

* [patgen](patgen) supports various general-purpose pattern-generation algorithms (e.g., `PermutedBinary` and `FlipBits`).

## Running, Logging, Stats, GUI toolkit

The following all work together to provide a convenient layer of abstraction for running, logging & statistics, and the GUI interface:

* [etime](time) provides the core time scales and training / testing etc modes used in the rest of the packages.

* [looper](looper) provides a fully step-able hierarchical looping control framework (e.g., Run / Epoch / Trial / Cycle) where you can insert functions to run at different points (start, end, or specific counter value).  [Axon](https;//github.com/emer/axon) uses this natively and has support functions for configuring standard looper functions.

* [estats](estats) manages statistics as maps of name, value for various types, along with network-relevant statistics such as `ClosestPat`, PCA stats, Cluster plots, decoders, raster plots, etc.

* [elog](elog) has comprehensive support for logging data at different time scales and evaluation modes -- saves a lot of boilerplate code for configuring, updating.

* [egui](egui) implements a standard simulation GUI, with a toolbar, tabs of different views, and a Sim struct view on the left.

* [econfig](econfig) manages command-line args and configuration files.

## Other Misc

* [actrf](actrf) provides activation-based receptive field stats (reverse correlation, spike-triggered averaging) for decoding internal representations.

* [chem](chem) provides basic chemistry simulation mechanisms for chemical reactions characterized by rate constants and concentrations, including diffusion.  This can be used for detailed biochemical models of neural function, as in the [Urakubo et al (2008)](https://github.com/ccnlab/kinase/sims/urakubo) model of synaptic plasticity.

* [confusion](confusion) provides confusion matricies for model output vs. target output.

* [decoder](decoder) provides simple linear, sigmoid, and softmax decoders for interpreting network activity states according to hypothesized variables of interest.

* [efuns](efuns) has misc special functions such as Gaussian and Sigmoid.

* [esg](esg) is the *emergent stochastic / sentence generator* -- parses simple grammars that generate random events (sentences) -- can be a good starting point for generating more complex environments.

* [popcode](popcode) supports the encoding and decoding of population codes -- distributed representations of numeric quantities across a population of neurons.  This is the `ScalarVal` functionality from C++ emergent, but now completely independent of any specific algorithm so it can be used anywhere.

* [ringidx](ringidx) provides a wrap-around ring index for efficient use of a fixed buffer that overwrites the oldest items without any copying.

# Other Packages

Here are the other packages from [Cogent Core](https://github.com/cogentcore/core) and within `emer` that provide infrastructure and other optional elements for simulations:

* [tensor](https://github.com/cogentcore/core/tree/main/tensor) in Cogent Core provides critical data management infrastructure in the form of n-dimensional `Tensor` data types, and the  [table.Table](https://github.com/cogentcore/core/tree/main/tensor/table), which is a collection of tensors and provides the same functionality as `pandas` or `xarray` in Python.  There are packages for standard data analysis tools like similarity / distance matricies, PCA, cluster plots, etc.  The [plot](https://github.com/cogentcore/core/tree/main/plot) package supports interactive plotting of data.

* [mpi](https://github.com/cogentcore/core/tree/main/base/mpi) in Cogent Core provides an MPI (message passing interface) distributed memory implementation, for running multiple coordinated instances of models across multiple compute nodes, and aggregating weight changes and results across them.

* [envs](https://github.com/emer/envs) has misc standalone environments that can be good starting points, including managing files, visual images, etc.

* [ttail](https://github.com/cogentcore/core/tree/main/tensor/cmd/ttail) is a `tail` program for interactively viewing tabular (csv, tsv, etc) log files in a terminal CLI environment!  `go install cogentcore.org/core/tensor/cmd/ttail@latest` from anywhere to install.

* [eTorch](https://github.com/emer/etorch) is the emergent interface to PyTorch models, providing emergent GUI NetView etc for these models.

* [physics](https://github.com/cogentcore/core/tree/main/xyz/physics) is a basic physics engine and collision detection framework, with 3D visualization using Cogent Core XYZ, for dynamic realistic environments and their visualization.

* [vision](https://github.com/emer/vision) and [auditory](https://github.com/emer/auditory) provide low-level filtering on sensory inputs reflecting corresponding biological mechanisms.

* [vgpu](https://github.com/cogentcore/core/tree/main/vgpu) is a GPU library using Vulkan for both graphics and compute functionality, which also has the [gosl](https://github.com/cogentcore/core/tree/main/vgpu/gosl) Go shader language system, which together enable complex models written in Go to run efficiently on the GPU, as used in the [axon](https;//github.com/emer/axon) package.  This system converts Go to HLSL shader code that can then be compiled and run in the VGPU framework.

* [Numbers](https://github.com/cogentcore/cogent/tree/main/numbers) is an interactive data analysis and numerical computing framework that can be used for analyzing simulation data.  It also has a `DataBrowser` that can be configured with convenient shell scripts for managing the process of running simulations on a remote cluster, see [axon simscripts](https://github.com/emer/axon/tree/main/simscripts) for examples.

