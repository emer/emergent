// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

//go:generate core generate

import (
	"fmt"
	"io"

	"cogentcore.org/core/math32"
	"cogentcore.org/core/tensor"
	"github.com/emer/emergent/v2/params"
	"github.com/emer/emergent/v2/relpos"
	"github.com/emer/emergent/v2/weights"
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

	// Label satisfies the core.Labeler interface for getting the name of objects generically
	Label() string

	// SetName sets name of layer
	SetName(nm string)

	// AddClass adds a CSS-style class name(s) for this layer,
	// ensuring that it is not a duplicate, and properly space separated.
	// Returns Layer so it can be chained to set other properties too
	AddClass(cls ...string) Layer

	// IsOff returns true if layer has been turned Off (lesioned) -- for experimentation
	IsOff() bool

	// SetOff sets the "off" (lesioned) status of layer. Also sets the Off state of all
	// pathways from this layer to other layers.
	SetOff(off bool)

	// Shape returns the organization of units in the layer, in terms of an array of dimensions.
	// Row-major ordering is default (Y then X), outer-most to inner-most.
	// if 2D, then it is a simple Y,X layer with no sub-structure (pools).
	// If 4D, then it number of pools Y, X and then number of units per pool Y, X
	Shape() *tensor.Shape

	// Is2D() returns true if this is a 2D layer (no Pools)
	Is2D() bool

	// Is4D() returns true if this is a 4D layer (has Pools as inner 2 dimensions)
	Is4D() bool

	// Index4DFrom2D returns the 4D index from 2D coordinates
	// within which inner dims are interleaved.  Returns false if 2D coords are invalid.
	Index4DFrom2D(x, y int) ([]int, bool)

	// Type returns the functional type of layer according to LayerType (extensible in
	// more specialized algorithms)
	Type() LayerType

	// SetType sets the functional type of layer
	SetType(typ LayerType)

	// Config configures the basic parameters of the layer
	Config(shape []int, typ LayerType)

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
	Pos() math32.Vector3

	// SetPos sets the 3D position of this layer -- will generally be overwritten by
	// automatic RelPos setting, unless that doesn't specify a valid relative position.
	SetPos(pos math32.Vector3)

	// Size returns the display size of this layer for the 3D view -- see Pos() for general info.
	// This is multiplied by the RelPos.Scale factor to rescale layer sizes, and takes
	// into account 2D and 4D layer structures.
	Size() math32.Vector2

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

	// UnitVarIndex returns the index of given variable within the Neuron,
	// according to *this layer's* UnitVarNames() list (using a map to lookup index),
	// or -1 and error message if not found.
	UnitVarIndex(varNm string) (int, error)

	// UnitVarNum returns the number of Neuron-level variables
	// for this layer.  This is needed for extending indexes in derived types.
	UnitVarNum() int

	// UnitVal1D returns value of given variable index on given unit,
	// using 1-dimensional index, and a data parallel index di,
	// for networks capable of processing multiple input patterns in parallel.
	// returns NaN on invalid index.
	// This is the core unit var access method used by other methods,
	// so it is the only one that needs to be updated for derived layer types.
	UnitVal1D(varIndex int, idx, di int) float32

	// UnitValues fills in values of given variable name on unit,
	// for each unit in the layer, into given float32 slice (only resized if not big enough).
	// di is a data parallel index di, for networks capable of processing input patterns in parallel.
	// Returns error on invalid var name.
	UnitValues(vals *[]float32, varNm string, di int) error

	// UnitValuesTensor fills in values of given variable name on unit
	// for each unit in the layer, into given tensor.
	// di is a data parallel index di, for networks capable of processing input patterns in parallel.
	// If tensor is not already big enough to hold the values, it is
	// set to the same shape as the layer.
	// Returns error on invalid var name.
	UnitValuesTensor(tsr tensor.Tensor, varNm string, di int) error

	// UnitValuesRepTensor fills in values of given variable name on unit
	// for a smaller subset of representative units in the layer, into given tensor.
	// di is a data parallel index di, for networks capable of processing input patterns in parallel.
	// This is used for computationally intensive stats or displays that work
	// much better with a smaller number of units.
	// The set of representative units are defined by SetRepIndexes -- all units
	// are used if no such subset has been defined.
	// If tensor is not already big enough to hold the values, it is
	// set to RepShape to hold all the values if subset is defined,
	// otherwise it calls UnitValuesTensor and is identical to that.
	// Returns error on invalid var name.
	UnitValuesRepTensor(tsr tensor.Tensor, varNm string, di int) error

	// RepIndexes returns the current set of representative unit indexes.
	// which are a smaller subset of units that represent the behavior
	// of the layer, for computationally intensive statistics and displays
	// (e.g., PCA, ActRF, NetView rasters).
	// Returns nil if none has been set (in which case all units should be used).
	// See utility function CenterPoolIndexes that returns indexes of
	// units in the central pools of a 4D layer.
	RepIndexes() []int

	// RepShape returns the shape to use for the subset of representative
	// unit indexes, in terms of an array of dimensions.  See Shape() for more info.
	// Layers that set RepIndexes should also set this, otherwise a 1D array
	// of len RepIndexes will be used.
	// See utility function CenterPoolShape that returns shape of
	// units in the central pools of a 4D layer.
	RepShape() *tensor.Shape

	// SetRepIndexesShape sets the RepIndexes, and RepShape and as list of dimension sizes
	SetRepIndexesShape(idxs, shape []int)

	// UnitVal returns value of given variable name on given unit,
	// using shape-based dimensional index.
	// Returns NaN on invalid var name or index.
	// di is a data parallel index di, for networks capable of processing input patterns in parallel.
	UnitValue(varNm string, idx []int, di int) float32

	// NRecvPaths returns the number of receiving pathways
	NRecvPaths() int

	// RecvPath returns a specific receiving pathway
	RecvPath(idx int) Path

	// NSendPaths returns the number of sending pathways
	NSendPaths() int

	// SendPath returns a specific sending pathway
	SendPath(idx int) Path

	// SendNameTry looks for a pathway connected to this layer whose sender layer has a given name
	SendNameTry(sender string) (Path, error)

	// SendNameTypeTry looks for a pathway connected to this layer whose sender layer has a given name and type
	SendNameTypeTry(sender, typ string) (Path, error)

	// RecvNameTry looks for a pathway connected to this layer whose receiver layer has a given name
	RecvNameTry(recv string) (Path, error)

	// RecvNameTypeTry looks for a pathway connected to this layer whose receiver layer has a given name and type
	RecvNameTypeTry(recv, typ string) (Path, error)

	// RecvPathValues fills in values of given synapse variable name,
	// for pathway from given sending layer and neuron 1D index,
	// for all receiving neurons in this layer,
	// into given float32 slice (only resized if not big enough).
	// pathType is the string representation of the path type -- used if non-empty,
	// useful when there are multiple pathways between two layers.
	// Returns error on invalid var name.
	// If the receiving neuron is not connected to the given sending layer or neuron
	// then the value is set to math32.NaN().
	// Returns error on invalid var name or lack of recv path (vals always set to nan on path err).
	RecvPathValues(vals *[]float32, varNm string, sendLay Layer, sendIndex1D int, pathType string) error

	// SendPathValues fills in values of given synapse variable name,
	// for pathway into given receiving layer and neuron 1D index,
	// for all sending neurons in this layer,
	// into given float32 slice (only resized if not big enough).
	// pathType is the string representation of the path type -- used if non-empty,
	// useful when there are multiple pathways between two layers.
	// Returns error on invalid var name.
	// If the sending neuron is not connected to the given receiving layer or neuron
	// then the value is set to math32.NaN().
	// Returns error on invalid var name or lack of recv path (vals always set to nan on path err).
	SendPathValues(vals *[]float32, varNm string, recvLay Layer, recvIndex1D int, pathType string) error

	// Defaults sets default parameter values for all Layer and recv pathway parameters
	Defaults()

	// UpdateParams() updates parameter values for all Layer and recv pathway parameters,
	// based on any other params that might have changed.
	UpdateParams()

	// ApplyParams applies given parameter style Sheet to this layer and its recv pathways.
	// Calls UpdateParams on anything set to ensure derived parameters are all updated.
	// If setMsg is true, then a message is printed to confirm each parameter that is set.
	// it always prints a message if a parameter fails to be set.
	// returns true if any params were set, and error if there were any errors.
	ApplyParams(pars *params.Sheet, setMsg bool) (bool, error)

	// SetParam sets parameter at given path to given value.
	// returns error if path not found or value cannot be set.
	SetParam(path, val string) error

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

	// Build constructs the layer and pathway state based on the layer shapes
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

// CenterPoolIndexes returns the indexes for n x n center pools of given 4D layer.
// Useful for setting RepIndexes on Layer.
// Will crash if called on non-4D layers.
func CenterPoolIndexes(ly Layer, n int) []int {
	nPy := ly.Shape().DimSize(0)
	nPx := ly.Shape().DimSize(1)
	sPy := (nPy - n) / 2
	sPx := (nPx - n) / 2
	nu := ly.Shape().DimSize(2) * ly.Shape().DimSize(3)
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
	return []int{n, n, ly.Shape().DimSize(2), ly.Shape().DimSize(3)}
}

// Layer2DRepIndexes returns neuron indexes and corresponding 2D shape
// for the representative neurons within a large 2D layer, for passing to
// [SetRepIndexesShape].  These neurons are used for the raster plot
// in the GUI and for computing PCA, among other cases where the full set
// of neurons is problematic. The lower-left corner of neurons up to
// given maxSize is selected.
func Layer2DRepIndexes(ly Layer, maxSize int) (idxs, shape []int) {
	sh := ly.Shape()
	my := min(maxSize, sh.DimSize(0))
	mx := min(maxSize, sh.DimSize(1))
	shape = []int{my, mx}
	idxs = make([]int, my*mx)
	i := 0
	for y := 0; y < my; y++ {
		for x := 0; x < mx; x++ {
			idxs[i] = sh.Offset([]int{y, x})
			i++
		}
	}
	return
}

//////////////////////////////////////////////////////////////////////////////////////
//  Layers

// Layers is a slice of layers
type Layers []Layer

// ElemLabel satisfies the core.SliceLabeler interface to provide labels for slice elements
func (ls *Layers) ElemLabel(idx int) string {
	return (*ls)[idx].Name()
}

//////////////////////////////////////////////////////////////////////////////////////
//  LayerType

// LayerType is the type of the layer: Input, Hidden, Target, Compare.
// Class parameter styles automatically key off of these types.
// Specialized algorithms can extend this to other types, but these types encompass
// most standard neural network models.
type LayerType int32 //enums:enum

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
)

// we keep these here to make it easier for other packages to implement the emer.Layer interface
// by just calling these methods
func SendNameTry(l Layer, sender string) (Path, error) {
	for pi := 0; pi < l.NRecvPaths(); pi++ {
		pj := l.RecvPath(pi)
		if pj.SendLay().Name() == sender {
			return pj, nil
		}
	}
	return nil, fmt.Errorf("sending layer: %v not found in list of pathways", sender)
}

func RecvNameTry(l Layer, recv string) (Path, error) {
	for pi := 0; pi < l.NSendPaths(); pi++ {
		pj := l.SendPath(pi)
		if pj.RecvLay().Name() == recv {
			return pj, nil
		}
	}
	return nil, fmt.Errorf("receiving layer: %v not found in list of pathways", recv)
}

func SendNameTypeTry(l Layer, sender, typ string) (Path, error) {
	for pi := 0; pi < l.NRecvPaths(); pi++ {
		pj := l.RecvPath(pi)
		if pj.SendLay().Name() == sender && pj.PathTypeName() == typ {
			return pj, nil
		}
	}
	return nil, fmt.Errorf("sending layer: %v not found in list of pathways", sender)
}

func RecvNameTypeTry(l Layer, recv, typ string) (Path, error) {
	for pi := 0; pi < l.NSendPaths(); pi++ {
		pj := l.SendPath(pi)
		if pj.RecvLay().Name() == recv && pj.PathTypeName() == typ {
			return pj, nil
		}
	}
	return nil, fmt.Errorf("receiving layer: %v, type: %v not found in list of pathways", recv, typ)
}
