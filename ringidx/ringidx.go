// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package ringidx provides circular indexing logic for writing a given
length of data into a fixed-sized buffer and wrapping around this
buffer, overwriting the oldest data.  No copying is required so
it is highly efficient
*/
package ringidx

// Idx is the ring index structure, maintaining starting index and length
// into a ring-buffer with maximum length Max.  Max must be > 0 and Len <= Max.
// When adding new items would overflow Max, starting index is shifted over
// to overwrite the oldest items with the new ones.  No moving is ever
// required -- just a fixed-length buffer of size Max.
type Idx struct {
	StIdx int `desc:"the starting index where current data starts -- the oldest data is at this index, and continues for Len items, wrapping around at Max, coming back up at most to StIdx-1"`
	Len   int `desc:"the number of items stored starting at StIdx.  Capped at Max"`
	Max   int `desc:"the maximum number of items that can be stored in this ring"`
}

// Idx returns the index of the i'th item starting from StIdx.
// i must be < Len.
func (ri *Idx) Idx(i int) int {
	i += ri.StIdx
	if i >= ri.Max {
		i -= ri.Max
	}
	return i
}

// LastIdx returns the index of the last (most recently added) item in the ring.
// Only valid if Len > 0
func (ri *Idx) LastIdx() int {
	return ri.Idx(ri.Len - 1)
}

// IdxIsValid returns true if given index is valid: >= 0 and < Len
func (ri *Idx) IdxIsValid(i int) bool {
	return i >= 0 && i < ri.Len
}

// Add adds given number of items to the ring (n <= Len.
// Shift is called for Len+n - Max extra items to make room.
func (ri *Idx) Add(n int) {
	over := (ri.Len + n) - ri.Max
	if over > 0 {
		ri.Shift(over)
	}
	ri.Len += n
}

// Shift moves the starting index up by n, and decrements the Len by n as well.
// This is called prior to adding new items if doing so would exceed Max length.
func (ri *Idx) Shift(n int) {
	ri.StIdx = ri.Idx(n)
	ri.Len -= n
}

// Reset initializes start index and length to 0
func (ri *Idx) Reset() {
	ri.StIdx = 0
	ri.Len = 0
}
