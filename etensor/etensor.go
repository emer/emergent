// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etensor

import "github.com/apache/arrow/go/arrow"

//go:generate tmpl -i -data=numeric.tmpldata numeric.gen.go.tmpl

// Tensor is the general interface for n-dimensional tensors
type Tensor interface {
	// Len returns the number of elements in the tensor (product of shape dimensions).
	Len() int

	// DataType returns the type of data, using arrow.DataType (ID() is the arrow.Type enum value)
	DataType() arrow.DataType

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

	// Dim returns the size of the given dimension
	Dim(i int) int

	// DimNames returns the string slice of dimension names
	DimNames() []string

	// DimName returns the name of the i-th dimension.
	DimName(i int) string

	IsContiguous() bool
	IsRowMajor() bool
	IsColMajor() bool

	// RowCellSize returns the size of the outer-most Row shape dimension, and the size of all the
	// remaining inner dimensions (the "cell" size) -- e.g., for Tensors that are columns in a
	// data table. Only valid for RowMajor organization.
	RowCellSize() (rows, cells int)

	// Offset returns the flat 1D array / slice index into an element at the given n-dimensional index.
	// No checking is done on the length or size of the index values relative to the shape of the tensor.
	Offset(i []int) int

	// Generic accessor routines support Float (float64) or String, either full dimensional or 1D

	// FloatVal returns the value of given index as a float64
	FloatVal(i []int) float64

	// SetFloat sets the value of given index as a float64
	SetFloat(i []int, val float64)

	// StringVal returns the value of given index as a string
	StringVal(i []int) string

	// SetString sets the value of given index as a string
	SetString(i []int, val string)

	// FloatVal1D returns the value of given 1-dimensional index (0-Len()-1) as a float64
	FloatVal1D(off int) float64

	// SetFloat1D sets the value of given 1-dimensional index (0-Len()-1) as a float64
	SetFloat1D(off int, val float64)

	// StringVal1D returns the value of given 1-dimensional index (0-Len()-1) as a string
	StringVal1D(off int) string

	// SetString1D sets the value of given 1-dimensional index (0-Len()-1) as a string
	SetString1D(off int, val string)

	// AggFloat applies given aggregation function to each element in the tensor, using float64
	// conversions of the values.  init is the initial value for the agg variable.  returns final
	// aggregate value
	AggFloat(fun func(val float64, agg float64) float64, init float64) float64

	// EvalFloat applies given function to each element in the tensor, using float64
	// conversions of the values, and puts the results into given float64 slice, which is
	// ensured to be of the proper length
	EvalFloat(fun func(val float64) float64, res *[]float64)

	// UpdtFloat applies given function to each element in the tensor, using float64
	// conversions of the values, and writes the results back into the same tensor values
	UpdtFloat(fun func(val float64) float64)

	// CloneTensor clones this tensor returning a Tensor interface.
	// There is a type-specific Clone() method as well for each tensor.
	CloneTensor() Tensor

	// SetShape sets the shape parameters of the tensor, and resizes backing storage appropriately.
	// existing RowMajor or ColMajor stride preference will be used if strides is nil, and
	// existing names will be preserved if nil
	SetShape(shape, strides []int, names []string)

	// AddRows adds n rows (outer-most dimension) to RowMajor organized tensor.
	// Does nothing for other stride layouts
	AddRows(n int)

	// SetNumRows sets the number of rows (outer-most dimension) in a RowMajor organized tensor.
	// Does nothing for other stride layouts
	SetNumRows(rows int)
}

// Check impl
var _ Tensor = (*Float32)(nil)
