// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package emergent is the overall repository for the emergent neural network simulation
software, written in Go (golang) with Python wrappers.

This top-level of the repository has no functional code -- everything is organized
into the following sub-repositories:

* emer: defines the primary structural interfaces for emergent, at the level of
Network, Layer, and Prjn (projection).  These contain no algorithm-specific code
and are only about the overall structure of a network, sufficient to support general
purpose tools such as the 3D NetView.  It also houses widely-used support classes used
in algorithm-specific code, including things like MinMax and AvgMax, and also the
parameter-styling infrastructure (emer.Params, emer.ParamStyle, emer.ParamSet and
emer.ParamSets).

* erand has misc random-number generation support functionality, including
erand.RndParams for parameterizing the type of random noise to add to a model,
and easier support for making permuted random lists, etc.

* netview provides the NetView interactive 3D network viewer, implemented in the GoGi 3D framework.

* prjn is a separate package for defining patterns of connectivity between layers
(i.e., the ProjectionSpecs from C++ emergent).  This is done using a fully independent
structure that *only* knows about the shapes of the two layers, and it returns a fully general
bitmap representation of the pattern of connectivity between them.  The leabra.Prjn code
then uses these patterns to do all the nitty-gritty of connecting up neurons.
This makes the projection code *much* simpler compared to the ProjectionSpec in C++ emergent,
which was involved in both creating the pattern and also all the complexity of setting up the
actual connections themselves.  This should be the *last* time any of those projection patterns
need to be written (having re-written this code too many times in the C++ version as the details
of memory allocations changed).

* patgen supports various pattern-generation algorithms, as implemented in taDataGen
in C++ emergent (e.g., PermutedBinary and FlipBits).

* timer is a simple interval timing struct, used for benchmarking / profiling etc.

* python contains a template Makefile that uses [GoPy](https://github.com/goki/gopy) to generate
python bindings to the entire emergent system.  See the leabra package version to actually run an example.

*/
package emergent
