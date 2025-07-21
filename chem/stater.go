// Copyright (c) 2021 The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chem

import "cogentcore.org/core/tensor/table"

// The Stater interface defines the functions implemented for State
// structures containing chem state variables.
// This interface is largely for documentation purposes.
type Stater interface {
	// Init Initializes the state to starting default values (concentrations)
	Init()

	// Zero sets all state variables to zero -- called for deltas after integration
	Zero()

	// Integrate is called with the deltas -- each state value calls Integrate()
	// to update from deltas.
	Integrate(d Stater)

	// Log records relevant state variables in given table at given row
	Log(dt *table.Table, row int)

	// ConfigLog configures the table to add column(s) for what is logged
	ConfigLog(dt *table.Table)
}
