// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

import (
	"github.com/emer/emergent/etensor"
)

// Layer defines the basic interface for neural network layers, used for managing the structural
// elements of a network, and for visualization, I/O, etc
// Interfaces are automatically pointers -- think of this as a pointer to your specific layer
// type, with a very basic interface for accessing general structural properties.  Nothing
// algorithm-specific is implemented here -- all of that goes in your specific layer struct.
type Layer interface {
	// LayName returns the name of this layer
	LayName() string

	// LayClass is for applying parameter styles, CSS-style -- can be space-separated multple tags
	LayClass() string

	// IsOff returns true if layer has been turned Off -- for experimentation
	IsOff() bool

	// Shape returns the organization of units in the layer, in terms of an array of dimensions.
	// if 2D, then it is a simple X,Y layer with no sub-structure (unit groups).
	// If 4D, then it is standard unit group X,Y units X,Y.
	LayShape() *etensor.Shape

	// LayPos returns the 3D position of the lower-left-hand corner of the layer
	LayPos() Vec3i

	// LayIndex returns a 0..n-1 index of the position of the layer within list of layers
	// in the network.  For backprop networks, index position has computational significance.
	// For Leabra networks, it only has significance in determining who gets which weights for
	// enforcing initial weight symmetry -- higher layers get weights from lower layers.
	LayIndex() int

	// Unit returns the unit at given index, which must be valid according to shape
	// otherwise a nil is returned
	Unit(index []int) Unit

	// UnitVals returns values of given variable name on unit for each unit in the layer, as a float32 slice
	UnitVals(varnm string) []float32

	// RecvPrjnList returns the full list of receiving projections
	RecvPrjnList() PrjnList

	// NRecvPrjns returns the number of receiving projections
	NRecvPrjns() int

	// RecvPrjn returns a specific receiving projection
	RecvPrjn(idx int) Prjn

	// SendPrjnList returns the full list of sending projections
	SendPrjnList() PrjnList

	// NSendPrjns returns the number of sending projections
	NSendPrjns() int

	// SendPrjn returns a specific sending projection
	SendPrjn(idx int) Prjn

	// Defaults sets default parameter values for all Layer and recv projection parameters
	Defaults()

	// UpdateParams() updates parameter values for all Layer and recv projection parameters,
	// based on any other params that might have changed.
	UpdateParams()

	// StyleParams applies a given ParamStyle style sheet to the layer and recv projections
	StyleParams(psty ParamStyle)
}
