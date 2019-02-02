// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

import (
	"io"

	"github.com/goki/gi/gi"
)

// Network defines the basic interface for a neural network, used for managing the structural
// elements of a network, and for visualization, I/O, etc
type Network interface {
	// NetName returns the name of this network
	NetName() string

	// Label satisfies the gi.Labeler interface for getting the name of objects generically
	Label() string

	// NLayers returns the number of layers in the network
	NLayers() int

	// LayerIndex returns layer (as emer.Layer interface) at given index -- does not
	// do extra bounds checking
	LayerIndex(idx int) Layer

	// LayerByName returns layer of given name, nil if not found
	LayerByName(name string) Layer

	// LayerByNameErrMsg returns layer of given name, emits a log error message and returns false if not found
	LayerByNameErrMsg(name string) (Layer, bool)

	// Defaults sets default parameter values for everything in the Network
	Defaults()

	// UpdateParams() updates parameter values for all Network parameters,
	// based on any other params that might have changed.
	UpdateParams()

	// StyleParams applies a given ParamStyle style sheet to the layers and projections in network
	StyleParams(psty ParamStyle)

	// WriteWtsJSON writes network weights (and any other state that adapts with learning)
	// to JSON-formatted output
	WriteWtsJSON(w io.Writer)

	// ReadWtsJSON reads network weights (and any other state that adapts with learning)
	// from JSON-formatted input
	ReadWtsJSON(r io.Reader) error

	// SaveWtsJSON saves network weights (and any other state that adapts with learning)
	// to a JSON-formatted file
	SaveWtsJSON(filename gi.FileName) error

	// OpenWtsJSON opens network weights (and any other state that adapts with learning)
	// from a JSON-formatted file
	OpenWtsJSON(filename gi.FileName) error
}
