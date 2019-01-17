// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

// Unit provides the interface for displaying unit state in a visualization etc
type Unit interface {
	// VarNames returns a list of variable names available on this unit
	VarNames() []string

	// VarByName returns the value of a variable by name, false if not a valid name
	VarByName(varNm string) (float32, bool)
}
