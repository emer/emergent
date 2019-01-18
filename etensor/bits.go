// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etensor

import "github.com/emer/emergent/bitslice"

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
	bt.Values = bitslice.Make(int(ln), 0)
	return bt
}

// Value returns value at given tensor index
func (tsr *Bits) Value(i []int) bool {
	j := int(tsr.Offset(i))
	return tsr.Values.Index(j)
}

// Set sets value at given tensor index
func (tsr *Bits) Set(i []int, val bool) {
	j := int(tsr.Offset(i))
	tsr.Values.Set(j, val)
}
