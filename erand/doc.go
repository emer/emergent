// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package erand provides randomization functionality built on top of standard math/rand
// random number generation functions.
//
// erand.Rand is an interface that enables calling the standard global rand functions,
// or a rand.Rand separate source, and is used for all methods in this package.
// Methods also take a thr thread arg to support a random generator that handles separate
// threads, such as gosl/slrand.
//
// erand.StdRand implements the interface.
//
//   - RndParams: specifies parameters for random number generation according to various distributions,
//     used e.g., for initializing random weights and generating random noise in neurons
//
// - Permute*: basic convenience methods calling rand.Shuffle on e.g., []int slice
//
// - BoolP: boolean for given probability
package erand
