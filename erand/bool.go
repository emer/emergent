// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package erand

// BoolP is a simple method to generate a true value with given probability
// (else false).  is just rand.Float32() < p but this is more readable
// and explicit.
// Thr is an optional parallel thread index (-1 for none).
// Optionally can pass a single Rand interface to use --
// otherwise uses system global Rand source.
func BoolP(p float32, thr int, randOpt ...Rand) bool {
	var rnd Rand
	if len(randOpt) == 0 {
		rnd = NewGlobalRand()
	} else {
		rnd = randOpt[0]
	}
	return rnd.Float32(thr) < p
}
