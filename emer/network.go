// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

import (
	"io"

	"github.com/emer/emergent/params"
	"github.com/emer/emergent/prjn"
	"github.com/emer/emergent/weights"
	"github.com/goki/gi/gi"
	"github.com/goki/mat32"
)

// Network defines the basic interface for a neural network, used for managing the structural
// elements of a network, and for visualization, I/O, etc
type Network interface {
	// InitName MUST be called to initialize the network's pointer to itself as an emer.Network
	// which enables the proper interface methods to be called.  Also sets the name.
	InitName(net Network, name string)

	// Name() returns name of the network
	Name() string

	// Label satisfies the gi.Labeler interface for getting the name of objects generically
	Label() string

	// NLayers returns the number of layers in the network
	NLayers() int

	// Layer returns layer (as emer.Layer interface) at given index -- does not
	// do extra bounds checking
	Layer(idx int) Layer

	// LayerByName returns layer of given name, nil if not found.
	// Layer names must be unique and a map is used so this is a fast operation
	LayerByName(name string) Layer

	// LayerByNameTry returns layer of given name,
	// returns error and emits a log message if not found.
	// Layer names must be unique and a map is used so this is a fast operation
	LayerByNameTry(name string) (Layer, error)

	// Defaults sets default parameter values for everything in the Network
	Defaults()

	// UpdateParams() updates parameter values for all Network parameters,
	// based on any other params that might have changed.
	UpdateParams()

	// ApplyParams applies given parameter style Sheet to layers and prjns in this network.
	// Calls UpdateParams on anything set to ensure derived parameters are all updated.
	// If setMsg is true, then a message is printed to confirm each parameter that is set.
	// it always prints a message if a parameter fails to be set.
	// returns true if any params were set, and error if there were any errors.
	ApplyParams(pars *params.Sheet, setMsg bool) (bool, error)

	// NonDefaultParams returns a listing of all parameters in the Network that
	// are not at their default values -- useful for setting param styles etc.
	NonDefaultParams() string

	// AllParams returns a listing of all parameters in the Network
	AllParams() string

	// UnitVarNames returns a list of variable names available on the units in this network.
	// This list determines what is shown in the NetView (and the order of vars list).
	// Not all layers need to support all variables, but must safely return mat32.NaN() for
	// unsupported ones.
	// This is typically a global list so do not modify!
	UnitVarNames() []string

	// UnitVarProps returns a map of unit variable properties, with the key being the
	// name of the variable, and the value gives a space-separated list of
	// go-tag-style properties for that variable.
	// The NetView recognizes the following properties:
	// range:"##" = +- range around 0 for default display scaling
	// min:"##" max:"##" = min, max display range
	// auto-scale:"+" or "-" = use automatic scaling instead of fixed range or not.
	// zeroctr:"+" or "-" = control whether zero-centering is used
	// desc:"txt" tooltip description of the variable
	// Note: this is typically a global list so do not modify!
	UnitVarProps() map[string]string

	// SynVarNames returns the names of all the variables on the synapses in this network.
	// This list determines what is shown in the NetView (and the order of vars list).
	// Not all projections need to support all variables, but must safely return mat32.NaN() for
	// unsupported ones.
	// This is typically a global list so do not modify!
	SynVarNames() []string

	// SynVarProps returns a map of synapse variable properties, with the key being the
	// name of the variable, and the value gives a space-separated list of
	// go-tag-style properties for that variable.
	// The NetView recognizes the following properties:
	// range:"##" = +- range around 0 for default display scaling
	// min:"##" max:"##" = min, max display range
	// auto-scale:"+" or "-" = use automatic scaling instead of fixed range or not.
	// zeroctr:"+" or "-" = control whether zero-centering is used
	// Note: this is typically a global list so do not modify!
	SynVarProps() map[string]string

	// WriteWtsJSON writes network weights (and any other state that adapts with learning)
	// to JSON-formatted output.
	WriteWtsJSON(w io.Writer) error

	// ReadWtsJSON reads network weights (and any other state that adapts with learning)
	// from JSON-formatted input.  Reads into a temporary weights.Network structure that
	// is then passed to SetWts to actually set the weights.
	ReadWtsJSON(r io.Reader) error

	// SetWts sets the weights for this network from weights.Network decoded values
	SetWts(nw *weights.Network) error

	// SaveWtsJSON saves network weights (and any other state that adapts with learning)
	// to a JSON-formatted file.  If filename has .gz extension, then file is gzip compressed.
	SaveWtsJSON(filename gi.FileName) error

	// OpenWtsJSON opens network weights (and any other state that adapts with learning)
	// from a JSON-formatted file.  If filename has .gz extension, then file is gzip uncompressed.
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
