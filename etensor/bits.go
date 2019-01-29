// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etensor

import (
	"github.com/apache/arrow/go/arrow"
	"github.com/emer/emergent/bitslice"
	"github.com/goki/ki/ints"
	"github.com/goki/ki/kit"
)

// BoolType not in arrow..

type BoolType struct{}

func (t *BoolType) ID() arrow.Type { return arrow.BOOL }
func (t *BoolType) Name() string   { return "bool" }
func (t *BoolType) BitWidth() int  { return 1 }

// etensor.Bits is a tensor of bits backed by a bitslice.Slice for efficient storage
// of binary data
type Bits struct {
	Shape
	Values bitslice.Slice
}

// NewBits returns a new n-dimensional array of bits
// If strides is nil, row-major strides will be inferred.
// If names is nil, a slice of empty strings will be created.
func NewBits(shape, strides []int, names []string) *Bits {
	bt := &Bits{}
	bt.SetShape(shape, strides, names)
	ln := bt.Len()
	bt.Values = bitslice.Make(ln, 0)
	return bt
}

// NewBitsShape returns a new n-dimensional array of bits
// If strides is nil, row-major strides will be inferred.
// If names is nil, a slice of empty strings will be created.
func NewBitsShape(shape *Shape) *Bits {
	bt := &Bits{}
	bt.CopyShape(shape)
	ln := bt.Len()
	bt.Values = bitslice.Make(ln, 0)
	return bt
}

func (tsr *Bits) DataType() arrow.DataType { return &BoolType{} }

func (tsr *Bits) Value(i []int) bool        { j := int(tsr.Offset(i)); return tsr.Values.Index(j) }
func (tsr *Bits) Set(i []int, val bool)     { j := int(tsr.Offset(i)); tsr.Values.Set(j, val) }
func (tsr *Bits) IsNull(i []int) bool       { return false }
func (tsr *Bits) SetNull(i []int, nul bool) {}

func Float64ToBool(val float64) bool {
	bv := true
	if val == 0 {
		bv = false
	}
	return bv
}

func BoolToFloat64(bv bool) float64 {
	if bv {
		return 1
	} else {
		return 0
	}
}

func (tsr *Bits) Float64Val(i []int) float64 {
	j := tsr.Offset(i)
	return BoolToFloat64(tsr.Values.Index(j))
}
func (tsr *Bits) SetFloat64(i []int, val float64) {
	j := tsr.Offset(i)
	tsr.Values.Set(j, Float64ToBool(val))
}

func (tsr *Bits) StringVal(i []int) string {
	j := tsr.Offset(i)
	return kit.ToString(tsr.Values.Index(j))
}

func (tsr *Bits) SetString(i []int, val string) {
	if bv, ok := kit.ToBool(val); ok {
		j := tsr.Offset(i)
		tsr.Values.Set(j, bv)
	}
}

// AggFloat64 applies given aggregation function to each element in the tensor, using float64
// conversions of the values.  init is the initial value for the agg variable.  returns final
// aggregate value
func (tsr *Bits) AggFloat64(fun func(val float64, agg float64) float64, ini float64) float64 {
	ln := tsr.Len()
	ag := ini
	for j := 0; j < ln; j++ {
		val := BoolToFloat64(tsr.Values.Index(j))
		ag = fun(val, ag)
	}
	return ag
}

// EvalFloat64 applies given function to each element in the tensor, using float64
// conversions of the values, and puts the results into given float64 slice, which is
// ensured to be of the proper length
func (tsr *Bits) EvalFloat64(fun func(val float64) float64, res *[]float64) {
	ln := tsr.Len()
	if len(*res) != ln {
		*res = make([]float64, ln)
	}
	for j := 0; j < ln; j++ {
		val := BoolToFloat64(tsr.Values.Index(j))
		(*res)[j] = fun(val)
	}
}

// UpdtFloat64 applies given function to each element in the tensor, using float64
// conversions of the values, and writes the results back into the same tensor values
func (tsr *Bits) UpdtFloat64(fun func(val float64) float64) {
	ln := tsr.Len()
	for j := 0; j < ln; j++ {
		val := BoolToFloat64(tsr.Values.Index(j))
		tsr.Values.Set(j, Float64ToBool(fun(val)))
	}
}

// Clone creates a new tensor that is a copy of the existing tensor, with its own
// separate memory -- changes to the clone will not affect the source.
func (tsr *Bits) Clone() *Bits {
	csr := NewBitsShape(&tsr.Shape)
	csr.Values = tsr.Values.Clone()
	return csr
}

// CloneTensor creates a new tensor that is a copy of the existing tensor, with its own
// separate memory -- changes to the clone will not affect the source.
func (tsr *Bits) CloneTensor() Tensor {
	return tsr.Clone()
}

// SetShape sets the shape params, resizing backing storage appropriately
func (tsr *Bits) SetShape(shape, strides []int, names []string) {
	tsr.SetShape(shape, strides, names)
	nln := tsr.Len()
	tsr.Values.SetLen(nln)
}

// AddRows adds n rows (outer-most dimension) to RowMajor organized tensor.
func (tsr *Bits) AddRows(n int) {
	if !tsr.IsRowMajor() {
		return
	}
	cln := tsr.Len()
	rows := tsr.Dim(0)
	inln := cln / rows // length of inner dims
	nln := (rows + n) * inln
	tsr.Shape.shape[0] += n
	tsr.Values.SetLen(nln)
}

// SetNumRows sets the number of rows (outer-most dimension) in a RowMajor organized tensor.
func (tsr *Bits) SetNumRows(rows int) {
	if !tsr.IsRowMajor() {
		return
	}
	rows = ints.MaxInt(1, rows) // must be > 0
	cln := tsr.Len()
	crows := tsr.Dim(0)
	inln := cln / crows // length of inner dims
	nln := rows * inln
	tsr.Shape.shape[0] = rows
	tsr.Values.SetLen(nln)
}
