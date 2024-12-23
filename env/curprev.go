// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env

import "cogentcore.org/lab/tensor"

// CurPrev manages current and previous values for basic data types.
type CurPrev[T tensor.DataTypes] struct {
	Cur, Prev T
}

// Set sets the new current value, after saving Cur to Prev.
func (cv *CurPrev[T]) Set(cur T) {
	cv.Prev = cv.Cur
	cv.Cur = cur
}

type CurPrevString = CurPrev[string]
