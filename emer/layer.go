// Copyright (c) 2018, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

// Layer defines the basic interface into neural network layers, used for visualization, I/O, etc
type Layer interface {
	// Shape returns the organization of units in the layer, in terms of an array of dimensions
	// If 4D, then it is standard unit group X,Y units X,Y.. for DCNN's it is.. ??
	Shape() Shape

	// Unit returns the unit at given coordinate, which must be valid according to shape
	// otherwise a false is returned
	Unit(coord []int) (Unit, bool)
}
