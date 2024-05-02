// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ringidx

//gosl:start ringidx

// FIx is a fixed-length ring index structure -- does not grow
// or shrink dynamically.
type FIx struct {

	// the zero index position -- where logical 0 is in physical buffer
	Zi uint32

	// the length of the buffer -- wraps around at this modulus
	Len uint32

	pad, pad1 uint32
}

// Index returns the physical index of the logical index i.
// i must be < Len.
func (fi *FIx) Index(i uint32) uint32 {
	i += fi.Zi
	if i >= fi.Len {
		i -= fi.Len
	}
	return i
}

// IndexIsValid returns true if given index is valid: >= 0 and < Len
func (fi *FIx) IndexIsValid(i uint32) bool {
	return i < fi.Len
}

// Shift moves the zero index up by n.
func (fi *FIx) Shift(n uint32) {
	fi.Zi = uint32(fi.Index(n))
}

//gosl:end ringidx
