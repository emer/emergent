// Copyright (c) 2021 The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chem

import "github.com/emer/etable/v2/etable"

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
	Log(dt *etable.Table, row int)

	// ConfigLog configures the table Schema to add column(s) for what is logged
	ConfigLog(sch *etable.Schema)
}
