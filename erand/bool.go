// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package erand

import "math/rand"

// BoolP is a simple method to generate a true value with given probability
// (else false).  is just rand.Float32() < p but this is more readable
// and explicit
func BoolP(p float32) bool {
	return rand.Float32() < p
}
