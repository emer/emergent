// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package erand

import "math/rand"

// PermuteInts permutes (shuffles) the order of elements in the given int slice
// using the standard Fisher-Yates shuffle
// https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle
// So you don't have to remember how to call rand.Shuffle
func PermuteInts(ins []int) {
	rand.Shuffle(len(ins), func(i, j int) {
		ins[i], ins[j] = ins[j], ins[i]
	})
}

// PermuteStrings permutes (shuffles) the order of elements in the given string slice
// using the standard Fisher-Yates shuffle
// https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle
// So you don't have to remember how to call rand.Shuffle
func PermuteStrings(ins []string) {
	rand.Shuffle(len(ins), func(i, j int) {
		ins[i], ins[j] = ins[j], ins[i]
	})
}

// PermuteFloat32s permutes (shuffles) the order of elements in the given float32 slice
// using the standard Fisher-Yates shuffle
// https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle
// So you don't have to remember how to call rand.Shuffle
func Permutefloat32s(ins []float32) {
	rand.Shuffle(len(ins), func(i, j int) {
		ins[i], ins[j] = ins[j], ins[i]
	})
}

// PermuteFloat64s permutes (shuffles) the order of elements in the given float64 slice
// using the standard Fisher-Yates shuffle
// https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle
// So you don't have to remember how to call rand.Shuffle
func Permutefloat64s(ins []float64) {
	rand.Shuffle(len(ins), func(i, j int) {
		ins[i], ins[j] = ins[j], ins[i]
	})
}
