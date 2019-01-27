// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etensor

// Tensor is the general interface for n-dimensional tensors
type Tensor interface {
	// Len returns the number of elements in the tensor.
	Len() int

	// Shapes returns the size in each dimension of the tensor. (Shape is the full Shape struct)
	Shapes() []int

	// Strides returns the number of elements to step in each dimension when traversing the tensor.
	Strides() []int

	// Shape64 returns the size in each dimension using int64 (arrow compatbile)
	Shape64() []int64

	// Strides64 returns the strides in each dimension using int64 (arrow compatbile)
	Strides64() []int64

	// NumDims returns the number of dimensions of the tensor.
	NumDims() int

	// DimNames returns the string slice of dimension names
	DimNames() []string

	// DimName returns the name of the i-th dimension.
	DimName(i int) string

	IsContiguous() bool
	IsRowMajor() bool
	IsColMajor() bool

	// Offset returns the flat 1D array / slice index into an element at the given n-dimensional index.
	// No checking is done on the length or size of the index values relative to the shape of the tensor.
	Offset(i []int) int
}

// Check impl
var _ Tensor = (*Float32)(nil)
