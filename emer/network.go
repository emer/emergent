// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

// Network defines the basic interface for a neural network, used for managing the structural
// elements of a network, and for visualization, I/O, etc
type Network interface {
	// NetName returns the name of this network
	NetName() string

	// NLayers returns the number of layers in the network
	NLayers() int

	// LayerIndex returns layer (as emer.Layer interface) at given index -- does not
	// do extra bounds checking
	LayerIndex(idx int) Layer

	// LayerByName returns layer of given name, nil if not found
	LayerByName(name string) Layer

	// LayerByNameErrMsg returns layer of given name, emits a log error message and returns false if not found
	LayerByNameErrMsg(name string) (Layer, bool)
}
