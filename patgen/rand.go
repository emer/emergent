// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package patgen

import "cogentcore.org/core/base/randx"

var (
	// RandSource is a random source to use for all random numbers used in patgen
	// By default it just uses the standard Go math/rand source.
	// If initialized, e.g., by calling NewRand(seed), then a separate stream of
	// random numbers will be generated for all patgen calls, and the
	// seed is saved as RandSeed -- it can be reinstated by calling RestoreSeed.
	// Can also set RandSource to another existing randx.Rand source to use it.
	RandSource = &randx.SysRand{}

	// Random seed last set by NewRand or SetRandSeed.
	RandSeed int64
)

// NewRand sets RandSource to a new separate random number stream
// using given seed, which is saved as RandSeed -- see RestoreSeed.
func NewRand(seed int64) {
	RandSource = randx.NewSysRand(seed)
	RandSeed = seed
}

// SetRandSeed sets existing random number stream to use given random
// seed, starting from the next call.  Saves the seed in RandSeed -- see RestoreSeed.
func SetRandSeed(seed int64) {
	RandSeed = seed
	RestoreSeed()
}

// RestoreSeed restores the random seed last used -- random number sequence
// will repeat what was generated from that point onward.
func RestoreSeed() {
	RandSource.Seed(RandSeed)
}
