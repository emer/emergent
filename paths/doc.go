// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package path is a separate package for defining patterns of connectivity between layers
(i.e., the ProjectionSpecs from C++ emergent).  This is done using a fully independent
structure that *only* knows about the shapes of the two layers, and it returns a fully general
bitmap representation of the pattern of connectivity between them.

The algorithm-specific leabra.Path code then uses these patterns to do all the nitty-gritty
of connecting up neurons.

This makes the pathway code *much* simpler compared to the ProjectionSpec in C++ emergent,
which was involved in both creating the pattern and also all the complexity of setting up the
actual connections themselves.  This should be the *last* time any of those pathway patterns
need to be written (having re-written this code too many times in the C++ version as the details
of memory allocations changed).

A Pattern maintains nothing about a specific pathway -- it only has the parameters that
are applied in creating a new pattern of connectivity, so it can be shared among any number
of pathways that need the same connectivity parameters.

All Patttern types have a New<Name> where <Name> is the type name, that creates a new
instance of given pattern initialized with default values.

Individual Pattern types may have a Defaults() method to initialize default values, but it is
not mandatory.
*/
package paths
