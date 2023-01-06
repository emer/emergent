// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ringidx

//gosl: start ringidx

// FIx is a fixed-length ring index structure -- does not grow
// or shrink dynamically.
type FIx struct {
	Zi  int32 `desc:"the zero index position -- where logical 0 is in physical buffer"`
	Len int32 `desc:"the length of the buffer -- wraps around at this modulus"`
}

// Idx returns the physical index of the logical index i.
// i must be < Len.
func (fi *FIx) Idx(i int) int {
	i += int(fi.Zi)
	if i >= int(fi.Len) {
		i -= int(fi.Len)
	}
	return i
}

// IdxIsValid returns true if given index is valid: >= 0 and < Len
func (fi *FIx) IdxIsValid(i int) bool {
	return i >= 0 && i < int(fi.Len)
}

// Shift moves the zero index up by n.
func (fi *FIx) Shift(n int) {
	fi.Zi = int32(fi.Idx(n))
}

//gosl: end ringidx
