// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

import "github.com/emer/emergent/etensor"

// Layer defines the basic interface into neural network layers, used for visualization, I/O, etc
type Layer interface {
	// Shape returns the organization of units in the layer, in terms of an array of dimensions.
	// if 2D, then it is a simple X,Y layer with no sub-structure (unit groups).
	// If 4D, then it is standard unit group X,Y units X,Y.
	Shape() etensor.Shape

	// Unit returns the unit at given index, which must be valid according to shape
	// otherwise a false is returned
	Unit(index []int) (Unit, bool)
}
