// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

import (
	"io"

	"github.com/emer/emergent/params"
	"github.com/emer/emergent/relpos"
	"github.com/emer/emergent/weights"
	"github.com/emer/etable/etensor"
	"github.com/goki/ki/kit"
	"github.com/goki/mat32"
)

// Layer defines the basic interface for neural network layers, used for managing the structural
// elements of a network, and for visualization, I/O, etc.
// Interfaces are automatically pointers -- think of this as a pointer to your specific layer
// type, with a very basic interface for accessing general structural properties.  Nothing
// algorithm-specific is implemented here -- all of that goes in your specific layer struct.
type Layer interface {
	params.Styler // TypeName, Name, and Class methods for parameter styling

	// InitName MUST be called to initialize the layer's pointer to itself as an emer.Layer
	// which enables the proper interface methods to be called.  Also sets the name, and
	// the parent network that this layer belongs to (which layers may want to retain).
	InitName(lay Layer, name string, net Network)

	// Label satisfies the gi.Labeler interface for getting the name of objects generically
	Label() string

	// SetName sets name of layer
	SetName(nm string)

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
	Shape() *etensor.Shape

	// Is2D() returns true if this is a 2D layer (no Pools)
	Is2D() bool

	// Is4D() returns true if this is a 4D layer (has Pools as inner 2 dimensions)
	Is4D() bool

	// Idx4DFrom2D returns the 4D index from 2D coordinates
	// within which inner dims are interleaved.  Returns false if 2D coords are invalid.
	Idx4DFrom2D(x, y int) ([]int, bool)

	// Type returns the functional type of layer according to LayerType (extensible in
	// more specialized algorithms)
	Type() LayerType

	// SetType sets the functional type of layer
	SetType(typ LayerType)

	// Config configures the basic parameters of the layer
	Config(shape []int, typ LayerType)

	// Thread() returns the thread number (go worker thread) to use in updating this layer.
	// The user is responsible for allocating layers to threads, trying to maintain an even
	// distribution across layers and establishing good break-points.
	Thread() int

	// SetThread sets the thread number (go worker thread) to use in updating this layer.
	SetThread(thr int)

	// RelPos returns the relative 3D position specification for this layer
	// for display in the 3D NetView -- see Pos() for display conventions.
	RelPos() relpos.Rel

	// SetRelPos sets the the relative 3D position specification for this layer
	SetRelPos(r relpos.Rel)

	// Pos returns the 3D position of the lower-left-hand corner of the layer.
	// The 3D view has layers arranged in X-Y planes stacked vertically along the Z axis.
	// Somewhat confusingly, this differs from the standard 3D graphics convention,
	// where the vertical dimension is Y and Z is the depth dimension.  However, in the
	// more "layer-centric" way of thinking about it, it is natural for the width & height
	// to map onto X and Y, and then Z is left over for stacking vertically.
	Pos() mat32.Vec3

	// SetPos sets the 3D position of this layer -- will generally be overwritten by
	// automatic RelPos setting, unless that doesn't specify a valid relative position.
	SetPos(pos mat32.Vec3)

	// Size returns the display size of this layer for the 3D view -- see Pos() for general info.
	// This is multiplied by the RelPos.Scale factor to rescale layer sizes, and takes
	// into account 2D and 4D layer structures.
	Size() mat32.Vec2

	// Index returns a 0..n-1 index of the position of the layer within list of layers
	// in the network.  For backprop networks, index position has computational significance.
	// For Leabra networks, it only has significance in determining who gets which weights for
	// enforcing initial weight symmetry -- higher layers get weights from lower layers.
	Index() int

	// SetIndex sets the layer index
	SetIndex(idx int)

	// UnitVarNames returns a list of variable names available on the units in this layer.
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
	// Note: this is a global list so do not modify!
	UnitVarProps() map[string]string

	// UnitVarIdx returns the index of given variable within the Neuron,
	// according to *this layer's* UnitVarNames() list (using a map to lookup index),
	// or -1 and error message if not found.
	UnitVarIdx(varNm string) (int, error)

	// UnitVarNum returns the number of Neuron-level variables
	// for this layer.  This is needed for extending indexes in derived types.
	UnitVarNum() int

	// UnitVal1D returns value of given variable index on given unit, using 1-dimensional index.
	// returns NaN on invalid index.
	// This is the core unit var access method used by other methods,
	// so it is the only one that needs to be updated for derived layer types.
	UnitVal1D(varIdx int, idx int) float32

	// UnitVals fills in values of given variable name on unit,
	// for each unit in the layer, into given float32 slice (only resized if not big enough).
	// Returns error on invalid var name.
	UnitVals(vals *[]float32, varNm string) error

	// UnitValsTensor fills in values of given variable name on unit
	// for each unit in the layer, into given tensor.
	// If tensor is not already big enough to hold the values, it is
	// set to the same shape as the layer.
	// Returns error on invalid var name.
	UnitValsTensor(tsr etensor.Tensor, varNm string) error

	// UnitValsRepTensor fills in values of given variable name on unit
	// for a smaller subset of representative units in the layer, into given tensor.
	// This is used for computationally intensive stats or displays that work
	// much better with a smaller number of units.
	// The set of representative units are defined by SetRepIdxs -- all units
	// are used if no such subset has been defined.
	// If tensor is not already big enough to hold the values, it is
	// set to RepShape to hold all the values if subset is defined,
	// otherwise it calls UnitValsTensor and is identical to that.
	// Returns error on invalid var name.
	UnitValsRepTensor(tsr etensor.Tensor, varNm string) error

	// RepIdxs returns the current set of representative unit indexes.
	// which are a smaller subset of units that represent the behavior
	// of the layer, for computationally intensive statistics and displays
	// (e.g., PCA, ActRF, NetView rasters).
	// Returns nil if none has been set (in which case all units should be used).
	// See utility function CenterPoolIdxs that returns indexes of
	// units in the central pools of a 4D layer.
	RepIdxs() []int

	// RepShape returns the shape to use for the subset of representative
	// unit indexes, in terms of an array of dimensions.  See Shape() for more info.
	// Layers that set RepIdxs should also set this, otherwise a 1D array
	// of len RepIdxs will be used.
	// See utility function CenterPoolShape that returns shape of
	// units in the central pools of a 4D layer.
	RepShape() *etensor.Shape

	// SetRepIdxsShape sets the RepIdxs, and RepShape and as list of dimension sizes
	SetRepIdxsShape(idxs, shape []int)

	// UnitVal returns value of given variable name on given unit,
	// using shape-based dimensional index.
	// Returns NaN on invalid var name or index.
	UnitVal(varNm string, idx []int) float32

	// NRecvPrjns returns the number of receiving projections
	NRecvPrjns() int

	// RecvPrjn returns a specific receiving projection
	RecvPrjn(idx int) Prjn

	// NSendPrjns returns the number of sending projections
	NSendPrjns() int

	// SendPrjn returns a specific sending projection
	SendPrjn(idx int) Prjn

	// RecvPrjnVals fills in values of given synapse variable name,
	// for projection from given sending layer and neuron 1D index,
	// for all receiving neurons in this layer,
	// into given float32 slice (only resized if not big enough).
	// prjnType is the string representation of the prjn type -- used if non-empty,
	// useful when there are multiple projections between two layers.
	// Returns error on invalid var name.
	// If the receiving neuron is not connected to the given sending layer or neuron
	// then the value is set to mat32.NaN().
	// Returns error on invalid var name or lack of recv prjn (vals always set to nan on prjn err).
	RecvPrjnVals(vals *[]float32, varNm string, sendLay Layer, sendIdx1D int, prjnType string) error

	// SendPrjnVals fills in values of given synapse variable name,
	// for projection into given receiving layer and neuron 1D index,
	// for all sending neurons in this layer,
	// into given float32 slice (only resized if not big enough).
	// prjnType is the string representation of the prjn type -- used if non-empty,
	// useful when there are multiple projections between two layers.
	// Returns error on invalid var name.
	// If the sending neuron is not connected to the given receiving layer or neuron
	// then the value is set to mat32.NaN().
	// Returns error on invalid var name or lack of recv prjn (vals always set to nan on prjn err).
	SendPrjnVals(vals *[]float32, varNm string, recvLay Layer, recvIdx1D int, prjnType string) error

	// Defaults sets default parameter values for all Layer and recv projection parameters
	Defaults()

	// UpdateParams() updates parameter values for all Layer and recv projection parameters,
	// based on any other params that might have changed.
	UpdateParams()

	// ApplyParams applies given parameter style Sheet to this layer and its recv projections.
	// Calls UpdateParams on anything set to ensure derived parameters are all updated.
	// If setMsg is true, then a message is printed to confirm each parameter that is set.
	// it always prints a message if a parameter fails to be set.
	// returns true if any params were set, and error if there were any errors.
	ApplyParams(pars *params.Sheet, setMsg bool) (bool, error)

	// NonDefaultParams returns a listing of all parameters in the Layer that
	// are not at their default values -- useful for setting param styles etc.
	NonDefaultParams() string

	// AllParams returns a listing of all parameters in the Layer
	AllParams() string

	// WriteWtsJSON writes the weights from this layer from the receiver-side perspective
	// in a JSON text format.  We build in the indentation logic to make it much faster and
	// more efficient.
	WriteWtsJSON(w io.Writer, depth int)

	// ReadWtsJSON reads the weights from this layer from the receiver-side perspective
	// in a JSON text format.  This is for a set of weights that were saved *for one layer only*
	// and is not used for the network-level ReadWtsJSON, which reads into a separate
	// structure -- see SetWts method.
	ReadWtsJSON(r io.Reader) error

	// SetWts sets the weights for this layer from weights.Layer decoded values
	SetWts(lw *weights.Layer) error

	// Build constructs the layer and projection state based on the layer shapes
	// and patterns of interconnectivity
	Build() error

	// VarRange returns the min / max values for given variable
	// over the layer
	VarRange(varNm string) (min, max float32, err error)
}

// LayerDimNames2D provides the standard Shape dimension names for 2D layers
var LayerDimNames2D = []string{"Y", "X"}

// LayerDimNames4D provides the standard Shape dimension names for 4D layers
// which have Pools and then neurons within pools.
var LayerDimNames4D = []string{"PoolY", "PoolX", "NeurY", "NeurX"}

// CenterPoolIdxs returns the indexes for n x n center pools of given 4D layer.
// Useful for setting RepIdxs on Layer.
// Will crash if called on non-4D layers.
func CenterPoolIdxs(ly Layer, n int) []int {
	nPy := ly.Shape().Dim(0)
	nPx := ly.Shape().Dim(1)
	sPy := (nPy - n) / 2
	sPx := (nPx - n) / 2
	nu := ly.Shape().Dim(2) * ly.Shape().Dim(3)
	nt := n * n * nu
	idxs := make([]int, nt)
	ix := 0
	for py := 0; py < n; py++ {
		for px := 0; px < n; px++ {
			si := ((py+sPy)*nPx + px + sPx) * nu
			for ui := 0; ui < nu; ui++ {
				idxs[ix+ui] = si + ui
			}
			ix += nu
		}
	}
	return idxs
}

// CenterPoolShape returns shape for n x n center pools of given 4D layer.
// Useful for setting RepShape on Layer.
func CenterPoolShape(ly Layer, n int) []int {
	return []int{n, n, ly.Shape().Dim(2), ly.Shape().Dim(3)}
}

//////////////////////////////////////////////////////////////////////////////////////
//  Layers

// Layers is a slice of layers
type Layers []Layer

// ElemLabel satisfies the gi.SliceLabeler interface to provide labels for slice elements
func (ls *Layers) ElemLabel(idx int) string {
	return (*ls)[idx].Name()
}

//////////////////////////////////////////////////////////////////////////////////////
//  LayerType

// LayerType is the type of the layer: Input, Hidden, Target, Compare.
// Class parameter styles automatically key off of these types.
// Specialized algorithms can extend this to other types, but these types encompass
// most standard neural network models.
type LayerType int32

//go:generate stringer -type=LayerType

var KiT_LayerType = kit.Enums.AddEnum(LayerTypeN, kit.NotBitFlag, nil)

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
