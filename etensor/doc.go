// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package etensor provides a basic set of tensor data structures
// based on apache/arrow/go/tensor and intercompatible with those
// structures.
//
// The primary differences are:
// * pure simple unidimensional Go slice used as the backing data array, auto allocated
// * fully modifiable data -- arrow is designed to be read-only
// * Shape struct is fully usable separate from the tensor data
// * Everything exported, e.g., Offset method
// * int used instead of int64 to make everything easier -- target platforms
//   are all 64bit and have 64bit int in Go by default
//
package etensor
