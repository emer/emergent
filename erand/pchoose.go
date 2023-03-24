// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package erand

// PChoose32 chooses an index in given slice of float32's at random according
// to the probilities of each item (must be normalized to sum to 1).
// Thr is an optional parallel thread index (-1 for none).
// Optionally can pass a single Rand interface to use --
// otherwise uses system global Rand source.
func PChoose32(ps []float32, thr int, randOpt ...Rand) int {
	var rnd Rand
	if len(randOpt) == 0 {
		rnd = NewGlobalRand()
	} else {
		rnd = randOpt[0]
	}
	pv := rnd.Float32(thr)
	sum := float32(0)
	for i, p := range ps {
		sum += p
		if pv < sum { // note: lower values already excluded
			return i
		}
	}
	return len(ps) - 1
}

// PChoose64 chooses an index in given slice of float64's at random according
// to the probilities of each item (must be normalized to sum to 1)
// Thr is an optional parallel thread index (-1 for none).
// Optionally can pass a single Rand interface to use --
// otherwise uses system global Rand source.
func PChoose64(ps []float64, thr int, randOpt ...Rand) int {
	var rnd Rand
	if len(randOpt) == 0 {
		rnd = NewGlobalRand()
	} else {
		rnd = randOpt[0]
	}
	pv := rnd.Float64(thr)
	sum := float64(0)
	for i, p := range ps {
		sum += p
		if pv < sum { // note: lower values already excluded
			return i
		}
	}
	return len(ps) - 1
}
