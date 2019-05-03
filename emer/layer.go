// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

import (
	"io"

	"github.com/emer/emergent/relpos"
	"github.com/emer/etable/etensor"
	"github.com/goki/ki/kit"
)

// Layer defines the basic interface for neural network layers, used for managing the structural
// elements of a network, and for visualization, I/O, etc.
// Interfaces are automatically pointers -- think of this as a pointer to your specific layer
// type, with a very basic interface for accessing general structural properties.  Nothing
// algorithm-specific is implemented here -- all of that goes in your specific layer struct.
type Layer interface {
	// InitName MUST be called to initialize the layer's pointer to itself as an emer.Layer
	// which enables the proper interface methods to be called.  Also sets the name.
	InitName(lay Layer, name string)

	// LayName returns the name of this layer
	LayName() string

	// Label satisfies the gi.Labeler interface for getting the name of objects generically
	Label() string

	// LayClass is for applying parameter styles, CSS-style -- can be space-separated multple tags
	LayClass() string

	// SetClass sets CSS-style class name(s) for this layer (space-separated if multiple)
	SetClass(cls string)

	// IsOff returns true if layer has been turned Off (lesioned) -- for experimentation
	IsOff() bool

	// SetOff sets the "off" (lesioned) status of layer
	SetOff(off bool)

	// Shape returns the organization of units in the layer, in terms of an array of dimensions.
	// Row-major ordering is default (Y then X), outer-most to inner-most.
	// if 2D, then it is a simple Y,X layer with no sub-structure (pools).
	// If 4D, then it number of pools Y, X and then number of units per pool Y, X
	LayShape() *etensor.Shape

	// LayType returns the functional type of layer according to LayerType (extensible in
	// more specialized versions of Leabra)
	LayType() LayerType

	// SetType sets the functional type of layer
	SetType(typ LayerType)

	// Config configures the basic parameters of the layer
	Config(shape []int, typ LayerType)

	// LayThread() returns the thread number (go worker thread) to use in updating this layer.
	// The user is responsible for allocating layers to threads, trying to maintain an even
	// distribution across layers and establishing good break-points.
	LayThread() int

	// SetThread sets the thread number (go worker thread) to use in updating this layer.
	SetThread(thr int)

	// LayRel returns the relative 3D position specification for this layer
	LayRel() relpos.Rel

	// SetLayRel sets the the relative 3D position specification for this layer
	SetLayRel(r relpos.Rel)

	// LayPos returns the 3D position of the lower-left-hand corner of the layer
	LayPos() relpos.Pos3D

	// LayIndex returns a 0..n-1 index of the position of the layer within list of layers
	// in the network.  For backprop networks, index position has computational significance.
	// For Leabra networks, it only has significance in determining who gets which weights for
	// enforcing initial weight symmetry -- higher layers get weights from lower layers.
	LayIndex() int

	// SetIndex sets the layer index
	SetIndex(idx int)

	// UnitVarNames returns a list of variable names available on the units in this layer.
	// this is a global list so do not modify!
	UnitVarNames() []string

	// UnitVals returns values of given variable name on unit for each unit in the layer,
	// as a float32 slice (which is created de-novo)
	UnitVals(varnm string) ([]float32, error)

	// UnitVal returns value of given variable name on given unit,
	// using shape-based dimensional index.
	UnitVal(varnm string, idx []int) (float32, error)

	// UnitVal1D returns value of given variable name on given unit,
	// using 1-dimensional index.
	UnitVal1D(varnm string, idx int) (float32, error)

	// RecvPrjnList returns the full list of receiving projections
	RecvPrjnList() *PrjnList

	// NRecvPrjns returns the number of receiving projections
	NRecvPrjns() int

	// RecvPrjn returns a specific receiving projection
	RecvPrjn(idx int) Prjn

	// SendPrjnList returns the full list of sending projections
	SendPrjnList() *PrjnList

	// NSendPrjns returns the number of sending projections
	NSendPrjns() int

	// SendPrjn returns a specific sending projection
	SendPrjn(idx int) Prjn

	// Defaults sets default parameter values for all Layer and recv projection parameters
	Defaults()

	// UpdateParams() updates parameter values for all Layer and recv projection parameters,
	// based on any other params that might have changed.
	UpdateParams()

	// SetParams sets given parameters to this layer, if the target type is Layer
	// calls UpdateParams to ensure derived parameters are all updated.
	// If setMsg is true, then a message is printed to confirm each parameter that is set.
	// it always prints a message if a parameter fails to be set.
	SetParams(pars Params, setMsg bool) bool

	// StyleParam applies a given style to either this layer or the receiving projections in this layer
	// depending on the style specification (.Class, #Name, Type) and target value of params.
	// .LayerType is automatically recognized as a .Class type (e.g., .Hidden vs. .Input etc)
	// Returns true if applied successfully.
	// If setMsg is true, then a message is printed to confirm each parameter that is set.
	// it always prints a message if a parameter fails to be set.
	StyleParam(sty string, pars Params, setMsg bool) bool

	// StyleParams applies a given ParamStyle style sheet to the layer and recv projections
	// If setMsg is true, then a message is printed to confirm each parameter that is set.
	// it always prints a message if a parameter fails to be set.
	StyleParams(psty ParamStyle, setMsg bool)

	// WriteWtsJSON writes the weights from this layer from the receiver-side perspective
	// in a JSON text format.  We build in the indentation logic to make it much faster and
	// more efficient.
	WriteWtsJSON(w io.Writer, depth int)

	// ReadWtsJSON reads the weights from this layer from the receiver-side perspective
	// in a JSON text format.
	ReadWtsJSON(r io.Reader) error

	// Build constructs the layer and projection state based on the layer shapes
	// and patterns of interconnectivity
	Build() error
}

//////////////////////////////////////////////////////////////////////////////////////
//  LayerType

// LayerType is the type of the layer: Input, Hidden, Target, Compare.
// Class parameter styles automatically key off of these types.
// Specialized algorithms can extend this to other types, but these types encompass
// most standard neural network models.
type LayerType int32

//go:generate stringer -type=LayerType

var KiT_LayerType = kit.Enums.AddEnum(LayerTypeN, false, nil)

func (ev LayerType) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *LayerType) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }

// The layer types
const (
	// Hidden is an internal representational layer that does not receive direct input / targets
	Hidden LayerType = iota

	// Input is a layer that receives direct external input in its Ext inputs
	Input

	// Target is a layer that receives direct external target inputs used for driving plus-phase learning
	Target

	// Compare is a layer that receives external comparison inputs, which drive statistics but
	// do NOT drive activation or learning directly
	Compare

	LayerTypeN
)
