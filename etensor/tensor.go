// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package etensor provides tensor Shape management.
// Based on code from apache/arrow/go/tensor, which is all hidden.
package etensor

// Shape manages a tensor's shape information, including strides and dimension names
// and can compute the flat index into an underlying 1D data storage array based on an
// n-dimensional index (and vice-versa)
type Shape struct {
	shape   []int64
	strides []int64
	names   []string
}

// Len returns the total length of elements in the tensor (i.e., the product of
// the shape sizes)
func (sh *Shape) Len() int64 {
	o := int64(1)
	for _, v := range sh.shape {
		o *= v
	}
	return o
}

func (sh *Shape) Shape() []int64       { return sh.shape }
func (sh *Shape) Strides() []int64     { return sh.strides }
func (sh *Shape) NumDims() int         { return len(sh.shape) }
func (sh *Shape) DimName(i int) string { return sh.names[i] }

func (sh *Shape) IsContiguous() bool {
	return sh.IsRowMajor() || sh.IsColMajor()
}

func (sh *Shape) IsRowMajor() bool {
	strides := RowMajorStrides(sh.shape)
	return EqualInt64s(strides, sh.strides)
}

func (sh *Shape) IsColMajor() bool {
	strides := ColMajorStrides(sh.shape)
	return EqualInt64s(strides, sh.strides)
}

// Offset returns the "flat" 1D array index into an element at the given n-dimensional index
// No checking is done on the length of the index relative to the shape of the tensor.
func (sh *Shape) Offset(index []int64) int64 {
	var offset int64
	for i, v := range index {
		offset += v * sh.strides[i]
	}
	return offset
}

// Index returns the n-dimensional index from a "flat" 1D array index
// No checking is done on the length of the index relative to the shape of the tensor.
// func (sh *Shape) Index(offset int64) []int64 {
// 	index := make([]int64, len(sh.strides))
// 	for i, v := range sh.strides {
// 		index[i] = sh.strides[i]
// 	}
// 	return index
// }

// RowMajorStrides returns strides for shape where the first dimension is outer-most
// and subsequent dimensions are progressively inner
func RowMajorStrides(shape []int64) []int64 {
	rem := int64(1)
	for _, v := range shape {
		rem *= v
	}

	if rem == 0 {
		strides := make([]int64, len(shape))
		rem := int64(1)
		for i := range strides {
			strides[i] = rem
		}
		return strides
	}

	strides := make([]int64, len(shape))
	for i, v := range shape {
		rem /= v
		strides[i] = rem
	}
	return strides
}

// ColMajorStrides returns strides for shape where the first dimension is inner-most
// and subsequent dimensions are progressively outer
func ColMajorStrides(shape []int64) []int64 {
	total := int64(1)
	for _, v := range shape {
		if v == 0 {
			strides := make([]int64, len(shape))
			for i := range strides {
				strides[i] = total
			}
			return strides
		}
	}

	strides := make([]int64, len(shape))
	for i, v := range shape {
		strides[i] = total
		total *= v
	}
	return strides
}

// EqualInt64 compares two int64 slices and returns true if they are equal
func EqualInt64s(a, b []int64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
