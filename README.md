# emergent reboot in Go

[![Go Report Card](https://goreportcard.com/badge/github.com/emer/emergent)](https://goreportcard.com/report/github.com/emer/emergent)
[![GoDoc](https://godoc.org/github.com/emer/emergent?status.svg)](https://godoc.org/github.com/emer/emergent)
[![Travis](https://travis-ci.com/emer/emergent.svg?branch=master)](https://travis-ci.com/emer/emergent)

This is the new home of the *emergent* neural network simulation software, developed primarily by the CCN lab at the University of Colorado Boulder.  We have decided to completely reboot the entire enterprise from the ground up, with a much more open, general-purpose design and approach.

See [Wiki Install](https://github.com/emer/emergent/wiki/Install) for installation instructions, and the [Wiki Rationale](https://github.com/emer/emergent/wiki/Rationale) and [History](https://github.com/emer/emergent/wiki/History) pages for a more detailed rationale for the new version of emergent, and a history of emergent (and its predecessors).

# Current Status / News

* 9/10/2019: First batch of [Comp Cog Neuro sims](https://github.com/CompCogNeuro/sims) complete, for chapters 2-4, which drove development of lots of good stuff in [etable](https://github.com/emer/etable), including `simat` package for similarity / distance matricies, PCA, and cluster plots -- and a few key advances in eplot functionality.  See the [etable wiki](https://github.com/emer/etable/wiki) for docs and example code, and the `family_trees` as an example textbook sim with lots of this code.

* 8/26/2019: Starting on [CCN Textbook](https://grey.colorado.edu/CompCogNeuro/index.php/CCNBook/Main) simulations using new emergent: https://github.com/CompCogNeuro/sims -- in course of doing so, finding and fixing a number of issues, including getting the OpenWtsJSON function to work (using a corrected JSON format), and improving the NetView etc.  For those following along with their own forks, this is a burst of change that should settle down in a week or so.

* 6/12/2019: **Initial beta release!!**  The [leabra ra25](https://github.com/emer/leabra/blob/master/examples/ra25/README.md) example code is now sufficiently feature complete to provide a reasonable starting point for creating your own simulations, with both the Go and Python versions closely matched in functionality so you can choose either way (or explore both!).  We expect that although there will certainly be changes and improvements etc (there are many planned additions already), the existing code should be reasonably stable at this point, with only relatively minor changes to the API -- it is now safe to start building your own models!   Please file issues on github (use the emergent repository's issue tracker) for any bugs or other issues you find.

* 5/2019: NetView nearly fully functional. `eplot` package and `Plot2D` widget provides basic dynamic gui for 2d plots.  See [TODO](#todo) section below for current planning roadmap for a push to get all the basic functionality in place so people can start using it!

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

Here are some of the additional supporting packages:

* `emer` *only* has the primary abstract Network interfaces (previously had put other random things in there, but the new policy is to keep everything in separate packages, as that seems to be where things end up eventually as they are better developed).

* `params` has the parameter-styling infrastructure (e.g., `params.Set`, `params.Sheet`, `params.Sel`), which implement a powerful, flexible, and efficient CSS style-sheet approach to parameters.  See the [Wiki Params](https://github.com/emer/emergent/wiki/Params) page for more info.

* `env` has an interface for environments, which encapsulates all the counters and timing information for patterns that are presented to the network, and enables more of a mix-and-match ability for using different environments with different networks.  See [Wiki Env](https://github.com/emer/emergent/wiki/Env) page for more info.

* `netview` provides the `NetView` interactive 3D network viewer, implemented in the GoGi 3D framework.

* `prjn` is a separate package for defining patterns of connectivity between layers (i.e., the `ProjectionSpec`s from C++ emergent).  This is done using a fully independent structure that *only* knows about the shapes of the two layers, and it returns a fully general bitmap representation of the pattern of connectivity between them.  The `leabra.Prjn` code then uses these patterns to do all the nitty-gritty of connecting up neurons.  This makes the projection code *much* simpler compared to the ProjectionSpec in C++ emergent, which was involved in both creating the pattern and also all the complexity of setting up the actual connections themselves.  This should be the *last* time any of those projection patterns need to be written (having re-written this code too many times in the C++ version as the details of memory allocations changed).

* `patgen` supports various general-purpose pattern-generation algorithms, as implemented in `taDataGen` in C++ emergent (e.g., `PermutedBinary` and `FlipBits`).

* `popcode` supports the encoding and decoding of population codes -- distributed representations of numeric quantities across a population of neurons.  This is the `ScalarVal` functionality from C++ emergent, but now completely independent of any specific algorithm so it can be used anywhere.

* `erand` has misc random-number generation support functionality, including `erand.RndParams` for parameterizing the type of random noise to add to a model, and easier support for making permuted random lists, etc.

* `timer` is a simple interval timing struct, used for benchmarking / profiling etc.

* `python` contains a template `Makefile` that uses [GoPy](https://github.com/goki/gopy) to generate python bindings to the entire emergent system.  See the `leabra` package version to actually run an example.

* The [etable](https://github.com/emer/etable) repository holds all of the more general-purpose "DataTable" or DataFrame (`etable.Table`) related code, which is our version of something like `pandas` or `xarray` in Python.  This includes the `etensor` n-dimensional array, `eplot` for interactive plotting of data, and basic utility packages like `minmax` and `bitslice`, and lots of data analysis tools like similarity / distance matricies, PCA, cluster plots, etc.

# TODO

Last updated: 10/28/2019

This list is not strictly in order, but roughly so..

- [ ] fix bugs, try to update glfw including keyboard input fixes

- [ ] write converter from Go to Python

- [x] do basic test of deep leabra

- [ ] MPI -- see [MPI Wiki page](https://github.com/emer/emergent/wiki/DMem)

- [ ] add python example code for interchange between pandas, xarray, tensorflow tensor stuff and etable.Table -- right now the best is to just save as .csv and load from there (esp for pandas which doesn't have tensors) -- should be able to use arrow stuff so it would be good to look into that.

- [ ] td/rl -- include ability to simulate drugs!

- [x] pbwm -- in process now

- [ ] pvlv

- [x] virtual environment -- [eve](https://github.com/emer/eve) is under way -- no actual physics yet but core infrastructure in place and usable for basic boxy objects under external control.

- [ ] GPU -- see https://github.com/gorgonia/gorgonia for existing CUDA impl -- alternatively, maybe try using opengl or vulkan directly within existing gogi/gpu framework -- would work on any GPU and seems like it wouldn't be very hard and gives full control -- https://www.khronos.org/opengl/wiki/Compute_Shader -- 4.3 min version though -- maybe better to just go to vulkan?  https://community.khronos.org/t/opencl-vs-vulkan-compute/7132/6

