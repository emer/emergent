// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package erand provides randomization functionality built on top of standard math/rand
// random number generation functions.  Includes:
// *  RndParams: specifies parameters for random number generation according to various distributions
//    used e.g., for initializing random weights and generating random noise in neurons
// *  Permute*: basic convenience methods calling rand.Shuffle on e.g., []int slice
//
package erand
