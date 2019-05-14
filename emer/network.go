// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

import (
	"io"

	"github.com/emer/emergent/prjn"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/mat32"
)

// Network defines the basic interface for a neural network, used for managing the structural
// elements of a network, and for visualization, I/O, etc
type Network interface {
	// InitName MUST be called to initialize the network's pointer to itself as an emer.Network
	// which enables the proper interface methods to be called.  Also sets the name.
	InitName(net Network, name string)

	// Name returns the name of this network
	Name() string

	// Label satisfies the gi.Labeler interface for getting the name of objects generically
	Label() string

	// NLayers returns the number of layers in the network
	NLayers() int

	// Layer returns layer (as emer.Layer interface) at given index -- does not
	// do extra bounds checking
	Layer(idx int) Layer

	// LayerByName returns layer of given name, nil if not found
	LayerByName(name string) Layer

	// LayerByNameTry returns layer of given name, returns error and emits a log message
	// if not found
	LayerByNameTry(name string) (Layer, error)

	// Defaults sets default parameter values for everything in the Network
	Defaults()

	// UpdateParams() updates parameter values for all Network parameters,
	// based on any other params that might have changed.
	UpdateParams()

	// StyleParams applies a given ParamStyle style sheet to the layers and projections in network
	// If setMsg is true, then a message is printed to confirm each parameter that is set.
	// it always prints a message if a parameter fails to be set.
	StyleParams(psty ParamStyle, setMsg bool)

	// StyleParamSet applies given set of ParamStyles to the layers and projections in network
	// If setMsg is true, then a message is printed to confirm each parameter that is set.
	// it always prints a message if a parameter fails to be set.
	StyleParamSet(pset ParamSet, setMsg bool)

	// NonDefaultParams returns a listing of all parameters in the Network that
	// are not at their default values -- useful for setting param styles etc.
	NonDefaultParams() string

	// WriteWtsJSON writes network weights (and any other state that adapts with learning)
	// to JSON-formatted output.
	WriteWtsJSON(w io.Writer)

	// ReadWtsJSON reads network weights (and any other state that adapts with learning)
	// from JSON-formatted input.
	ReadWtsJSON(r io.Reader) error

	// SaveWtsJSON saves network weights (and any other state that adapts with learning)
	// to a JSON-formatted file.
	SaveWtsJSON(filename gi.FileName) error

	// OpenWtsJSON opens network weights (and any other state that adapts with learning)
	// from a JSON-formatted file.
	OpenWtsJSON(filename gi.FileName) error

	// NewLayer creates a new concrete layer of appropriate type for this network
	NewLayer() Layer

	// NewPrjn creates a new concrete projection of appropriate type for this network
	NewPrjn() Prjn

	// ConnectLayerNames establishes a projection between two layers, referenced by name
	// adding to the recv and send projection lists on each side of the connection.
	// Returns error if not successful.
	// Does not yet actually connect the units within the layers -- that requires Build.
	ConnectLayerNames(send, recv string, pat prjn.Pattern, typ PrjnType) (rlay, slay Layer, pj Prjn, err error)

	// ConnectLayers establishes a projection between two layers,
	// adding to the recv and send projection lists on each side of the connection.
	// Returns false if not successful. Does not yet actually connect the units within the layers -- that
	// requires Build.
	ConnectLayers(send, recv Layer, pat prjn.Pattern, typ PrjnType) Prjn

	// Bounds returns the minimum and maximum display coordinates of the network for 3D display
	Bounds() (min, max mat32.Vec3)

	// VarRange returns the min / max values for given variable
	VarRange(varNm string) (min, max float32, err error)
}
