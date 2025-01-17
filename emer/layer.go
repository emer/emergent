// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

import (
	"fmt"
	"io"
	"log"
	"math"

	"cogentcore.org/core/base/slicesx"
	"cogentcore.org/core/math32"
	"cogentcore.org/lab/tensor"
	"github.com/emer/emergent/v2/params"
	"github.com/emer/emergent/v2/relpos"
	"github.com/emer/emergent/v2/weights"
)

var (
	// LayerDimNames2D provides the standard Shape dimension names for 2D layers
	LayerDimNames2D = []string{"Y", "X"}

	// LayerDimNames4D provides the standard Shape dimension names for 4D layers
	// which have Pools and then neurons within pools.
	LayerDimNames4D = []string{"PoolY", "PoolX", "NeurY", "NeurX"}
)

// Layer defines the minimal interface for neural network layers,
// necessary to support the visualization (NetView), I/O,
// and parameter setting functionality provided by emergent.
// Most of the standard expected functionality is defined in the
// LayerBase struct, and this interface only has methods that must be
// implemented specifically for a given algorithmic implementation.
type Layer interface {

	// AsEmer returns the layer as an *emer.LayerBase,
	// to access base functionality.
	AsEmer() *LayerBase

	// Label satisfies the core.Labeler interface for getting
	// the name of objects generically. Use to access Name via interface.
	Label() string

	// TypeName is the type or category of layer, defined
	// by the algorithm (and usually set by an enum).
	TypeName() string

	// TypeNumber is the numerical value for the type or category
	// of layer, defined by the algorithm (and usually set by an enum).
	TypeNumber() int

	// UnitVarIndex returns the index of given variable within
	// the Neuron, according to *this layer's* UnitVarNames() list
	// (using a map to lookup index), or -1 and error message if
	// not found.
	UnitVarIndex(varNm string) (int, error)

	// UnitValue1D returns value of given variable index on given unit,
	// using 1-dimensional index, and a data parallel index di,
	// for networks capable of processing multiple input patterns
	// in parallel. Returns NaN on invalid index.
	// This is the core unit var access method used by other methods,
	// so it is the only one that needs to be updated for derived layer types.
	UnitValue1D(varIndex int, idx, di int) float32

	// VarRange returns the min / max values for given variable
	VarRange(varNm string) (min, max float32, err error)

	// NumRecvPaths returns the number of receiving pathways.
	NumRecvPaths() int

	// RecvPath returns a specific receiving pathway.
	RecvPath(idx int) Path

	// NumSendPaths returns the number of sending pathways.
	NumSendPaths() int

	// SendPath returns a specific sending pathway.
	SendPath(idx int) Path

	// RecvPathValues fills in values of given synapse variable name,
	// for pathway from given sending layer and neuron 1D index,
	// for all receiving neurons in this layer,
	// into given float32 slice (only resized if not big enough).
	// pathType is the string representation of the path type;
	// used if non-empty, useful when there are multiple pathways
	// between two layers.
	// Returns error on invalid var name.
	// If the receiving neuron is not connected to the given sending
	// layer or neuron then the value is set to math32.NaN().
	// Returns error on invalid var name or lack of recv path
	// (vals always set to nan on path err).
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

	// NonDefaultParams returns a listing of all parameters in the Layer that
	// are not at their default values; useful for setting param styles etc.
	NonDefaultParams() string

	// AllParams returns a listing of all parameters in the Layer
	AllParams() string

	// WriteWeightsJSON writes the weights from this layer from the
	// receiver-side perspective in a JSON text format.
	WriteWeightsJSON(w io.Writer, depth int)

	// SetWeights sets the weights for this layer from weights.Layer
	// decoded values
	SetWeights(lw *weights.Layer) error
}

// LayerBase defines the basic shared data for neural network layers,
// used for managing the structural elements of a network,
// and for visualization, I/O, etc.
// Nothing algorithm-specific is implemented here
type LayerBase struct {
	// EmerLayer provides access to the emer.Layer interface
	// methods for functions defined in the LayerBase type.
	// Must set this with a pointer to the actual instance
	// when created, using InitLayer function.
	EmerLayer Layer `display:"-"`

	// Name of the layer, which must be unique within the network.
	// Layers are typically accessed directly by name, via a map.
	Name string

	// Class is for applying parameter styles across multiple layers
	// that all get the same parameters.  This can be space separated
	// with multple classes.
	Class string

	// Doc contains documentation about the layer.
	// This is displayed in a tooltip in the network view.
	Doc string

	// Off turns off the layer, removing from all computations.
	// This provides a convenient way to dynamically test for
	// the contributions of the layer, for example.
	Off bool

	// Shape of the layer, either 2D or 4D.  Although spatial topology
	// is not relevant to all algorithms, the 2D shape is important for
	// efficiently visualizing large numbers of units / neurons.
	// 4D layers have 2D Pools of units embedded within a larger 2D
	// organization of such pools.  This is used for max-pooling or
	// pooled inhibition at a finer-grained level, and biologically
	// corresopnds to hypercolumns in the cortex for example.
	// Order is outer-to-inner (row major), so Y then X for 2D;
	// 4D: Y-X unit pools then Y-X neurons within pools.
	Shape tensor.Shape

	// Pos specifies the relative spatial relationship to another
	// layer, which determines positioning.  Every layer except one
	// "anchor" layer should be positioned relative to another,
	// e.g., RightOf, Above, etc.  This provides robust positioning
	// in the face of layer size changes etc.
	// Layers are arranged in X-Y planes, stacked vertically along the Z axis.
	Pos relpos.Pos `table:"-" display:"inline"`

	// Index is a 0..n-1 index of the position of the layer within
	// the list of layers in the network.
	Index int `display:"-" edit:"-"`

	// SampleIndexes are the current set of "sample" unit indexes,
	// which are a smaller subset of units that represent the behavior
	// of the layer, for computationally intensive statistics and displays
	// (e.g., PCA, ActRF, NetView rasters), when the layer is large.
	// If none have been set, then all units are used.
	// See utility function CenterPoolIndexes that returns indexes of
	// units in the central pools of a 4D layer.
	SampleIndexes []int `table:"-"`

	// SampleShape is the shape to use for the subset of sample
	// unit indexes, in terms of an array of dimensions.
	// See Shape for more info.
	// Layers that set SampleIndexes should also set this,
	// otherwise a 1D array of len SampleIndexes will be used.
	// See utility function CenterPoolShape that returns shape of
	// units in the central pools of a 4D layer.
	SampleShape tensor.Shape `table:"-"`

	// optional metadata that is saved in network weights files,
	// e.g., can indicate number of epochs that were trained,
	// or any other information about this network that would be useful to save.
	MetaData map[string]string
}

// InitLayer initializes the layer, setting the EmerLayer interface
// to provide access to it for LayerBase methods, along with the name.
func InitLayer(l Layer, name string) {
	lb := l.AsEmer()
	lb.EmerLayer = l
	lb.Name = name
}

func (ly *LayerBase) AsEmer() *LayerBase { return ly }
func (ly *LayerBase) Label() string      { return ly.Name }

// AddClass adds a CSS-style class name(s) for this layer,
// ensuring that it is not a duplicate, and properly space separated.
// Returns Layer so it can be chained to set other properties too.
func (ly *LayerBase) AddClass(cls ...string) *LayerBase {
	ly.Class = params.AddClass(ly.Class, cls...)
	return ly
}

// Is2D() returns true if this is a 2D layer (no Pools)
func (ly *LayerBase) Is2D() bool { return ly.Shape.NumDims() == 2 }

// Is4D() returns true if this is a 4D layer (has Pools as inner 2 dimensions)
func (ly *LayerBase) Is4D() bool { return ly.Shape.NumDims() == 4 }

func (ly *LayerBase) NumUnits() int { return ly.Shape.Len() }

// Index4DFrom2D returns the 4D index from 2D coordinates
// within which inner dims are interleaved.  Returns false if 2D coords are invalid.
func (ly *LayerBase) Index4DFrom2D(x, y int) ([]int, bool) {
	lshp := ly.Shape
	nux := lshp.DimSize(3)
	nuy := lshp.DimSize(2)
	ux := x % nux
	uy := y % nuy
	px := x / nux
	py := y / nuy
	idx := []int{py, px, uy, ux}
	if !lshp.IndexIsValid(idx...) {
		return nil, false
	}
	return idx, true
}

// PlaceRightOf positions the layer to the right of the other layer,
// with given spacing, using default YAlign = Front alignment.
func (ly *LayerBase) PlaceRightOf(other Layer, space float32) {
	ly.Pos.SetRightOf(other.AsEmer().Name, space)
}

// PlaceBehind positions the layer behind the other layer,
// with given spacing, using default XAlign = Left alignment.
func (ly *LayerBase) PlaceBehind(other Layer, space float32) {
	ly.Pos.SetBehind(other.AsEmer().Name, space)
}

// PlaceAbove positions the layer above the other layer,
// using default XAlign = Left, YAlign = Front alignment.
func (ly *LayerBase) PlaceAbove(other Layer) {
	ly.Pos.SetAbove(other.AsEmer().Name)
}

// DisplaySize returns the display size of this layer for the 3D view.
// see Pos field for general info.
// This is multiplied by the Pos.Scale factor to rescale
// layer sizes, and takes into account 2D and 4D layer structures.
func (ly *LayerBase) DisplaySize() math32.Vector2 {
	if ly.Pos.Scale == 0 {
		ly.Pos.Defaults()
	}
	var sz math32.Vector2
	switch {
	case ly.Is2D():
		sz = math32.Vec2(float32(ly.Shape.DimSize(1)), float32(ly.Shape.DimSize(0))) // Y, X
	case ly.Is4D():
		// note: pool spacing is handled internally in display and does not affect overall size
		sz = math32.Vec2(float32(ly.Shape.DimSize(1)*ly.Shape.DimSize(3)), float32(ly.Shape.DimSize(0)*ly.Shape.DimSize(2))) // Y, X
	default:
		sz = math32.Vec2(float32(ly.Shape.Len()), 1)
	}
	return sz.MulScalar(ly.Pos.Scale)
}

// SetShape sets the layer shape and also uses default dim names.
func (ly *LayerBase) SetShape(shape ...int) {
	ly.Shape.SetShapeSizes(shape...)
}

// SetSampleShape sets the SampleIndexes,
// and SampleShape and as list of dimension sizes,
// for a subset sample of units to represent the entire layer.
// This is critical for large layers that are otherwise unwieldy
// to visualize and for computationally-intensive statistics.
func (ly *LayerBase) SetSampleShape(idxs, shape []int) {
	ly.SampleIndexes = idxs
	ly.SampleShape.SetShapeSizes(shape...)
}

// GetSampleShape returns the shape to use for representative units.
func (ly *LayerBase) GetSampleShape() *tensor.Shape {
	sz := len(ly.SampleIndexes)
	if sz == 0 {
		return &ly.Shape
	}
	if ly.SampleShape.Len() != sz {
		ly.SampleShape.SetShapeSizes(sz)
	}
	return &ly.SampleShape
}

// NSubPools returns the number of sub-pools of neurons
// according to the shape parameters.  2D shapes have 0 sub pools.
// For a 4D shape, the pools are the first set of 2 Y,X dims
// and then the neurons within the pools are the 2nd set of 2 Y,X dims.
func (ly *LayerBase) NumPools() int {
	if ly.Shape.NumDims() != 4 {
		return 0
	}
	return ly.Shape.DimSize(0) * ly.Shape.DimSize(1)
}

// UnitValues fills in values of given variable name on unit,
// for each unit in the layer, into given float32 slice
// (only resized if not big enough).
// di is a data parallel index di, for networks capable of
// processing input patterns in parallel.
// Returns error on invalid var name.
func (ly *LayerBase) UnitValues(vals *[]float32, varNm string, di int) error {
	nn := ly.NumUnits()
	*vals = slicesx.SetLength(*vals, nn)
	vidx, err := ly.EmerLayer.UnitVarIndex(varNm)
	if err != nil {
		nan := math32.NaN()
		for lni := range nn {
			(*vals)[lni] = nan
		}
		return err
	}
	for lni := range nn {
		(*vals)[lni] = ly.EmerLayer.UnitValue1D(vidx, lni, di)
	}
	return nil
}

// UnitValuesTensor fills in values of given variable name
// on unit for each unit in the layer, into given tensor.
// di is a data parallel index di, for networks capable of
// processing input patterns in parallel.
// If tensor is not already big enough to hold the values, it is
// set to the same shape as the layer.
// Returns error on invalid var name.
func (ly *LayerBase) UnitValuesTensor(tsr tensor.Values, varNm string, di int) error {
	if tsr == nil {
		err := fmt.Errorf("emer.UnitValuesTensor: Tensor is nil")
		log.Println(err)
		return err
	}
	nn := ly.NumUnits()
	tsr.SetShapeSizes(ly.Shape.Sizes...)
	vidx, err := ly.EmerLayer.UnitVarIndex(varNm)
	if err != nil {
		nan := math.NaN()
		for lni := 0; lni < nn; lni++ {
			tsr.SetFloat1D(nan, lni)
		}
		return err
	}
	for lni := 0; lni < nn; lni++ {
		v := ly.EmerLayer.UnitValue1D(vidx, lni, di)
		if math32.IsNaN(v) {
			tsr.SetFloat1D(math.NaN(), lni)
		} else {
			tsr.SetFloat1D(float64(v), lni)
		}
	}
	return nil
}

// UnitValuesSampleTensor fills in values of given variable name
// on unit for a smaller subset of representative units
// in the layer, into given tensor.
// di is a data parallel index di, for networks capable of
// processing input patterns in parallel.
// This is used for computationally intensive stats or displays that work
// much better with a smaller number of units.
// The set of representative units are defined by SetSampleIndexes -- all units
// are used if no such subset has been defined.
// If tensor is not already big enough to hold the values, it is
// set to SampleShape to hold all the values if subset is defined,
// otherwise it calls UnitValuesTensor and is identical to that.
// Returns error on invalid var name.
func (ly *LayerBase) UnitValuesSampleTensor(tsr tensor.Values, varNm string, di int) error {
	nu := len(ly.SampleIndexes)
	if nu == 0 {
		return ly.UnitValuesTensor(tsr, varNm, di)
	}
	if tsr == nil {
		err := fmt.Errorf("emer.UnitValuesSampleTensor: Tensor is nil")
		log.Println(err)
		return err
	}
	if tsr.Len() != nu {
		rs := ly.GetSampleShape()
		tsr.SetShapeSizes(rs.Sizes...)
	}
	vidx, err := ly.EmerLayer.UnitVarIndex(varNm)
	if err != nil {
		nan := math.NaN()
		for i, _ := range ly.SampleIndexes {
			tsr.SetFloat1D(nan, i)
		}
		return err
	}
	for i, ui := range ly.SampleIndexes {
		v := ly.EmerLayer.UnitValue1D(vidx, ui, di)
		if math32.IsNaN(v) {
			tsr.SetFloat1D(math.NaN(), i)
		} else {
			tsr.SetFloat1D(float64(v), i)
		}
	}
	return nil
}

// UnitValue returns value of given variable name on given unit,
// using shape-based dimensional index.
// Returns NaN on invalid var name or index.
// di is a data parallel index di, for networks capable of
// processing input patterns in parallel.
func (ly *LayerBase) UnitValue(varNm string, idx []int, di int) float32 {
	vidx, err := ly.EmerLayer.UnitVarIndex(varNm)
	if err != nil {
		return math32.NaN()
	}
	fidx := ly.Shape.IndexTo1D(idx...)
	return ly.EmerLayer.UnitValue1D(vidx, fidx, di)
}

// CenterPoolIndexes returns the indexes for n x n center pools of given 4D layer.
// Useful for setting SampleIndexes on Layer.
// Will crash if called on non-4D layers.
func CenterPoolIndexes(ly Layer, n int) []int {
	lb := ly.AsEmer()
	nPy := lb.Shape.DimSize(0)
	nPx := lb.Shape.DimSize(1)
	sPy := (nPy - n) / 2
	sPx := (nPx - n) / 2
	nu := lb.Shape.DimSize(2) * lb.Shape.DimSize(3)
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
// Useful for setting SampleShape on Layer.
func CenterPoolShape(ly Layer, n int) []int {
	lb := ly.AsEmer()
	return []int{n, n, lb.Shape.DimSize(2), lb.Shape.DimSize(3)}
}

// Layer2DSampleIndexes returns neuron indexes and corresponding 2D shape
// for the representative neurons within a large 2D layer, for passing to
// [SetSampleShape].  These neurons are used for the raster plot
// in the GUI and for computing PCA, among other cases where the full set
// of neurons is problematic. The lower-left corner of neurons up to
// given maxSize is selected.
func Layer2DSampleIndexes(ly Layer, maxSize int) (idxs, shape []int) {
	lb := ly.AsEmer()
	sh := lb.Shape
	my := min(maxSize, sh.DimSize(0))
	mx := min(maxSize, sh.DimSize(1))
	shape = []int{my, mx}
	idxs = make([]int, my*mx)
	i := 0
	for y := 0; y < my; y++ {
		for x := 0; x < mx; x++ {
			idxs[i] = sh.IndexTo1D(y, x)
			i++
		}
	}
	return
}

// RecvPathBySendName returns the receiving Path with given
// sending layer name (the first one if multiple exist).
func (ly *LayerBase) RecvPathBySendName(sender string) (Path, error) {
	el := ly.EmerLayer
	for pi := range el.NumRecvPaths() {
		pt := el.RecvPath(pi)
		if pt.SendLayer().Label() == sender {
			return pt, nil
		}
	}
	return nil, fmt.Errorf("sending layer named: %s not found in list of receiving pathways", sender)
}

// SendPathByRecvName returns the sending Path with given
// recieving layer name (the first one if multiple exist).
func (ly *LayerBase) SendPathByRecvName(recv string) (Path, error) {
	el := ly.EmerLayer
	for pi := range el.NumSendPaths() {
		pt := el.SendPath(pi)
		if pt.RecvLayer().Label() == recv {
			return pt, nil
		}
	}
	return nil, fmt.Errorf("receiving layer named: %s not found in list of sending pathways", recv)
}

// RecvPathBySendName returns the receiving Path with given
// sending layer name, with the given type name
// (the first one if multiple exist).
func (ly *LayerBase) RecvPathBySendNameType(sender, typeName string) (Path, error) {
	el := ly.EmerLayer
	for pi := range el.NumRecvPaths() {
		pt := el.RecvPath(pi)
		if pt.SendLayer().Label() == sender && pt.TypeName() == typeName {
			return pt, nil
		}
	}
	return nil, fmt.Errorf("sending layer named: %s of type %s not found in list of receiving pathways", sender, typeName)
}

// SendPathByRecvName returns the sending Path with given
// recieving layer name, with the given type name
// (the first one if multiple exist).
func (ly *LayerBase) SendPathByRecvNameType(recv, typeName string) (Path, error) {
	el := ly.EmerLayer
	for pi := range el.NumSendPaths() {
		pt := el.SendPath(pi)
		if pt.RecvLayer().Label() == recv && pt.TypeName() == typeName {
			return pt, nil
		}
	}
	return nil, fmt.Errorf("receiving layer named: %s, type: %s not found in list of sending pathways", recv, typeName)
}

//////// Params

// NonDefaultParams returns a listing of all parameters in the Layer that
// are not at their default values -- useful for setting param styles etc.
func (ly *LayerBase) NonDefaultParams() string {
	// nds := reflectx.NonDefaultFields(ly.Params) // todo:
	nds := "non default field strings todo"
	//Str(ly.AxonLay.AsAxon().Params, ly.Name)
	el := ly.EmerLayer
	for pi := range el.NumRecvPaths() {
		pt := el.RecvPath(pi).AsEmer()
		pnd := pt.NonDefaultParams()
		nds += pnd
	}
	return nds
}
