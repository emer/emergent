// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package erand

import (
	"math/rand"
	"time"
)

// Seeds is a set of random seeds, typically used one per Run
type Seeds []int64

// Init allocates given number of seeds and initializes them to
// sequential numbers 1..n
func (rs *Seeds) Init(n int) {
	*rs = make([]int64, n)
	for i := range *rs {
		(*rs)[i] = int64(i) + 1
	}
}

// Set sets the given seed as rand.Seed
func (rs *Seeds) Set(idx int) {
	rand.Seed((*rs)[idx])
}

// NewSeeds sets a new set of random seeds based on current time
func (rs *Seeds) NewSeeds() {
	rn := time.Now().UnixNano()
	for i := range *rs {
		(*rs)[i] = rn + int64(i)
	}
}
