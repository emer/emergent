// Copyright (c) 2021 The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chem

// CoToN returns N based on concentration, for given volume: co * vol
func CoToN(co, vol float64) float64 {
	return co * vol
}

// CoFmN returns concentration from N, for given volume: co / vol
func CoFmN(n, vol float64) float64 {
	return n / vol
}
