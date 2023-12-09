// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

import (
	"fmt"

	"github.com/emer/emergent/v2/params"
)

// LaySize contains parameters for size of layers
type LaySize struct {

	// Y (vertical) size of layer -- in units for 2D, or number of pools (outer dimension) for 4D layer
	Y int

	// X (horizontal) size of layer -- in units for 2D, or number of pools (outer dimension) for 4D layer
	X int

	// Y (vertical) size of each pool in units, only for 4D layers (inner dimension)
	PoolY int

	// Y (horizontal) size of each pool in units, only for 4D layers (inner dimension)
	PoolX int
}

// NetSize is a network schema for holding a params for layer sizes.
// Values can be queried for getting sizes when configuring the network.
// Uses params.Flex to support flexible parameter specification
type NetSize params.Flex

func (ns *NetSize) JSONString() string {
	return ((*params.Flex)(ns)).JSONString()
}

// ApplySheet applies given sheet of parameters to each layer
func (ns *NetSize) ApplySheet(sheet *params.Sheet, setMsg bool) {
	((*params.Flex)(ns)).ApplySheet(sheet, setMsg)
}

// AddLayers adds layer(s) of given class -- most efficient
// to add each class separately en-mass.
func (ns *NetSize) AddLayers(names []string, class string) {
	if *ns == nil {
		*ns = make(NetSize)
	}
	for _, nm := range names {
		(*ns)[nm] = &params.FlexVal{Nm: nm, Type: "Layer", Cls: class, Obj: &LaySize{}}
	}
}

// Layer returns the layer size for given layer name -- nil if not found
// and an error is emitted and returned
func (ns *NetSize) Layer(name string) (*LaySize, error) {
	fv, has := (*ns)[name]
	if !has {
		err := fmt.Errorf("emer.NetSize: layer named: %s not found", name)
		return nil, err
	}
	return fv.Obj.(*LaySize), nil
}

// LayX returns the X value = horizontal size of 2D layer or number of pools
// (outer dimension) for 4D layer, for given layer from size, if it set there.
// Otherwise returns the provided default value
func (ns *NetSize) LayX(name string, def int) int {
	ls, err := ns.Layer(name)
	if err != nil || ls.X == 0 {
		return def
	}
	return ls.X
}

// LayY returns the Y value = vertical size of 2D layer or number of pools
// (outer dimension) for 4D layer, for given layer from size, if it set there.
// Otherwise returns the provided default value
func (ns *NetSize) LayY(name string, def int) int {
	ls, err := ns.Layer(name)
	if err != nil || ls.Y == 0 {
		return def
	}
	return ls.Y
}

// PoolX returns the Pool X value (4D inner dim) = size of pool in units
// for given layer from size if it set there.
// Otherwise returns the provided default value
func (ns *NetSize) PoolX(name string, def int) int {
	ls, err := ns.Layer(name)
	if err != nil || ls.PoolX == 0 {
		return def
	}
	return ls.PoolX
}

// PoolY returns the Pool X value (4D inner dim) = size of pool in units
// for given layer from size if it set there.
// Otherwise returns the provided default value
func (ns *NetSize) PoolY(name string, def int) int {
	ls, err := ns.Layer(name)
	if err != nil || ls.PoolY == 0 {
		return def
	}
	return ls.PoolY
}
