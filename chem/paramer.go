// Copyright (c) 2021 The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chem

// The Paramer interface defines functions implemented for Params
// structures, containing chem React, Enz, etc functions.
// This interface is largely for documentation purposes.
type Paramer interface {
	// Defaults sets default parameters
	Defaults()

	// Step computes deltas d based on current values c
	Step(c, d Paramer)
}
