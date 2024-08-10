// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

//go:generate core generate -add-types

import (
	"fmt"
	"io"
	"strings"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/randx"
	"cogentcore.org/core/core"
	"cogentcore.org/core/math32"
	"github.com/emer/emergent/v2/params"
	"github.com/emer/emergent/v2/relpos"
	"github.com/emer/emergent/v2/weights"
)

// Network defines the minimal interface for a neural network,
// used for managing the structural elements of a network,
// and for visualization, I/O, etc.
// Most of the standard expected functionality is defined in the
// NetworkBase struct, and this interface only has methods that must be
// implemented specifically for a given algorithmic implementation.
type Network interface {
	// AsEmer returns the network as an *emer.NetworkBase,
	// to access base functionality.
	AsEmer() *NetworkBase

	// Label satisfies the core.Labeler interface for getting
	// the name of objects generically.
	Label() string

	// NumLayers returns the number of layers in the network.
	NumLayers() int

	// EmerLayer returns layer as emer.Layer interface at given index.
	// Does not do extra bounds checking.
	EmerLayer(idx int) Layer

	// MaxParallelData returns the maximum number of data inputs that can be
	// processed in parallel by the network.
	// The NetView supports display of up to this many data elements.
	MaxParallelData() int

	// NParallelData returns the current number of data inputs currently being
	// processed in parallel by the network.
	// Logging supports recording each of these where appropriate.
	NParallelData() int

	// Defaults sets default parameter values for everything in the Network.
	Defaults()

	// UpdateParams() updates parameter values for all Network parameters,
	// based on any other params that might have changed.
	UpdateParams()

	// ApplyParams applies given parameter style Sheet to layers
	// and paths in this network.
	// Calls UpdateParams on anything set to ensure derived parameters
	// are all updated.
	// If setMsg is true, then a message is printed to confirm each
	// parameter that is set.
	// it always prints a message if a parameter fails to be set.
	// returns true if any params were set, and error if there were any errors.
	ApplyParams(pars *params.Sheet, setMsg bool) (bool, error)

	// NonDefaultParams returns a listing of all parameters in the Network that
	// are not at their default values -- useful for setting param styles etc.
	NonDefaultParams() string

	// AllParams returns a listing of all parameters in the Network
	AllParams() string

	// KeyLayerParams returns a listing for all layers in the network,
	// of the most important layer-level params (specific to each algorithm).
	KeyLayerParams() string

	// KeyPathParams returns a listing for all Recv pathways in the network,
	// of the most important pathway-level params (specific to each algorithm).
	KeyPathParams() string

	// UnitVarNames returns a list of variable names available on
	// the units in this network.
	// This list determines what is shown in the NetView
	// (and the order of vars list).
	// Not all layers need to support all variables,
	// but must safely return math32.NaN() for unsupported ones.
	// This is typically a global list so do not modify!
	UnitVarNames() []string

	// UnitVarProps returns a map of unit variable properties,
	// with the key being the name of the variable,
	// and the value gives a space-separated list of
	// go-tag-style properties for that variable.
	// The NetView recognizes the following properties:
	// range:"##" = +- range around 0 for default display scaling
	// min:"##" max:"##" = min, max display range
	// auto-scale:"+" or "-" = use automatic scaling instead of fixed range or not.
	// zeroctr:"+" or "-" = control whether zero-centering is used
	// desc:"txt" tooltip description of the variable
	// Note: this is typically a global list so do not modify!
	UnitVarProps() map[string]string

	// SynVarNames returns the names of all the variables
	// on the synapses in this network.
	// This list determines what is shown in the NetView
	// (and the order of vars list).
	// Not all pathways need to support all variables,
	// but must safely return math32.NaN() for
	// unsupported ones.
	// This is typically a global list so do not modify!
	SynVarNames() []string

	// SynVarProps returns a map of synapse variable properties,
	// with the key being the name of the variable,
	// and the value gives a space-separated list of
	// go-tag-style properties for that variable.
	// The NetView recognizes the following properties:
	// range:"##" = +- range around 0 for default display scaling
	// min:"##" max:"##" = min, max display range
	// auto-scale:"+" or "-" = use automatic scaling instead of fixed range or not.
	// zeroctr:"+" or "-" = control whether zero-centering is used
	// Note: this is typically a global list so do not modify!
	SynVarProps() map[string]string

	// WriteWtsJSON writes network weights (and any other state
	// that adapts with learning) to JSON-formatted output.
	WriteWtsJSON(w io.Writer) error

	// ReadWtsJSON reads network weights (and any other state
	// that adapts with learning) from JSON-formatted input.
	// Reads into a temporary weights.Network structure that
	// is then passed to SetWts to actually set the weights.
	ReadWtsJSON(r io.Reader) error

	// SetWts sets the weights for this network from weights.Network
	// decoded values.
	SetWts(nw *weights.Network) error

	// SaveWtsJSON saves network weights (and any other state
	// that adapts with learning) to a JSON-formatted file.
	// If filename has .gz extension, then file is gzip compressed.
	SaveWtsJSON(filename core.Filename) error

	// OpenWtsJSON opens network weights (and any other state that
	// adapts with learning) from a JSON-formatted file.
	// If filename has .gz extension, then file is gzip uncompressed.
	OpenWtsJSON(filename core.Filename) error

	// VarRange returns the min / max values for given variable
	VarRange(varNm string) (min, max float32, err error)
}

// NetworkBase defines the basic data for a neural network,
// used for managing the structural elements of a network,
// and for visualization, I/O, etc.
type NetworkBase struct {
	// EmerNetwork provides access to the emer.Network interface
	// methods for functions defined in the NetworkBase type.
	// Must set this with a pointer to the actual instance
	// when created, using InitNetwork function.
	EmerNetwork Network

	// overall name of network, which helps discriminate if there are multiple.
	Name string

	// filename of last weights file loaded or saved.
	WeightsFile string

	// map of name to layers, for EmerLayerByName methods
	LayerNameMap map[string]Layer `display:"-"`

	// map from class name to layer names.
	LayerClassMap map[string][]string `display:"-"`

	// minimum display position in network
	MinPos math32.Vector3 `display:"-"`

	// maximum display position in network
	MaxPos math32.Vector3 `display:"-"`

	// optional metadata that is saved in network weights files,
	// e.g., can indicate number of epochs that were trained,
	// or any other information about this network that would be useful to save.
	MetaData map[string]string

	// random number generator for the network.
	// all random calls must use this.
	// Set seed here for weight initialization values.
	Rand randx.SysRand `display:"-"`

	// Random seed to be set at the start of configuring
	// the network and initializing the weights.
	// Set this to get a different set of weights.
	RandSeed int64 `edit:"-"`
}

// InitNetwork initializes the network, setting the EmerNetwork interface
// to provide access to it for NetworkBase methods, along with the name.
func InitNetwork(nt Network, name string) {
	nb := nt.AsEmer()
	nb.EmerNetwork = nt
	nb.Name = name
}

func (nt *NetworkBase) AsEmer() *NetworkBase { return nt }

func (nt *NetworkBase) Label() string { return nt.Name }

// UpdateLayerMaps updates the LayerNameMap and LayerClassMap.
// Call this when the network is built.
func (nt *NetworkBase) UpdateLayerMaps() {
	nt.LayerNameMap = make(map[string]Layer)
	nt.LayerClassMap = make(map[string][]string)
	nl := nt.EmerNetwork.NumLayers()
	for li := range nl {
		ly := nt.EmerNetwork.EmerLayer(li)
		lnm := ly.StyleName()
		nt.LayerNameMap[lnm] = ly
		cls := strings.Split(ly.StyleClass(), " ")
		for _, cl := range cls {
			ll := nt.LayerClassMap[cl]
			ll = append(ll, lnm)
			nt.LayerClassMap[cl] = ll
		}
	}
}

// EmerLayerByName returns a layer by looking it up by name.
// returns error message if layer is not found.
func (nt *NetworkBase) EmerLayerByName(name string) (Layer, error) {
	if nt.LayerNameMap == nil || len(nt.LayerNameMap) != nt.EmerNetwork.NumLayers() {
		nt.UpdateLayerMaps()
	}
	if ly, ok := nt.LayerNameMap[name]; ok {
		return ly, nil
	}
	err := fmt.Errorf("Layer named: %s not found in Network: %s", name, nt.Name)
	return nil, err
}

// EmerPathByName returns a path by looking it up by name.
// Paths are named SendToRecv = sending layer name "To" recv layer name.
// returns error message if path is not found.
func (nt *NetworkBase) EmerPathByName(name string) (Path, error) {
	ti := strings.Index(name, "To")
	if ti < 0 {
		return nil, errors.Log(fmt.Errorf("EmerPathByName: path name must contain 'To': %s", name))
	}
	sendNm := name[:ti]
	recvNm := name[ti+2:]
	_, err := nt.EmerLayerByName(sendNm)
	if errors.Log(err) != nil {
		return nil, err
	}
	recv, err := nt.EmerLayerByName(recvNm)
	if errors.Log(err) != nil {
		return nil, err
	}
	path, err := recv.AsEmer().RecvPathBySendName(sendNm)
	if errors.Log(err) != nil {
		return nil, err
	}
	return path, nil
}

// LayersByClass returns a list of layer names by given class(es).
// Lists are compiled when network Build() function called,
// or now if not yet present.
// The layer Type is always included as a Class, along with any other
// space-separated strings specified in Class for parameter styling, etc.
// If no classes are passed, all layer names in order are returned.
func (nt *NetworkBase) LayersByClass(classes ...string) []string {
	if nt.LayerClassMap == nil {
		nt.UpdateLayerMaps()
	}
	var nms []string
	nl := nt.EmerNetwork.NumLayers()
	if len(classes) == 0 {
		for li := range nl {
			ly := nt.EmerNetwork.EmerLayer(li).AsEmer()
			if ly.Off {
				continue
			}
			nms = append(nms, ly.Name)
		}
		return nms
	}
	for _, lc := range classes {
		nms = append(nms, nt.LayerClassMap[lc]...)
	}
	// only get unique layers
	layers := []string{}
	has := map[string]bool{}
	for _, nm := range nms {
		if has[nm] {
			continue
		}
		layers = append(layers, nm)
		has[nm] = true
	}
	if len(layers) == 0 {
		panic(fmt.Sprintf("No Layers found for query: %#v.", classes))
	}
	return layers
}

// LayoutLayers computes the 3D layout of layers based on their relative
// position settings.
func (nt *NetworkBase) LayoutLayers() {
	en := nt.EmerNetwork
	nlay := en.NumLayers()
	for range 5 {
		var lstly *LayerBase
		for li := range nlay {
			ly := en.EmerLayer(li).AsEmer()
			var oly *LayerBase
			if lstly != nil && ly.Pos.Rel == relpos.NoRel {
				if ly.Pos.Pos.X != 0 || ly.Pos.Pos.Y != 0 || ly.Pos.Pos.Z != 0 {
					// Position has been modified, don't mess with it.
					continue
				}
				oly = lstly
				ly.Pos = relpos.Pos{Rel: relpos.Above, Other: lstly.Name, XAlign: relpos.Middle, YAlign: relpos.Front}
			} else {
				if ly.Pos.Other != "" {
					olyi, err := nt.EmerLayerByName(ly.Pos.Other)
					if errors.Log(err) != nil {
						continue
					}
					oly = olyi.AsEmer()
				} else if lstly != nil {
					oly = lstly
					ly.Pos = relpos.Pos{Rel: relpos.Above, Other: lstly.Name, XAlign: relpos.Middle, YAlign: relpos.Front}
				}
			}
			if oly != nil {
				ly.Pos.SetPos(oly.Pos.Pos, oly.DisplaySize(), ly.DisplaySize())
			}
			lstly = ly
		}
	}
	nt.layoutBoundsUpdate()
}

// layoutBoundsUpdate updates the Min / Max display bounds for 3D display.
func (nt *NetworkBase) layoutBoundsUpdate() {
	en := nt.EmerNetwork
	nlay := en.NumLayers()
	mn := math32.Vector3Scalar(math32.Infinity)
	mx := math32.Vector3{}
	for li := range nlay {
		ly := en.EmerLayer(li).AsEmer()
		sz := ly.DisplaySize()
		ru := ly.Pos.Pos
		ru.X += sz.X
		ru.Y += sz.Y
		mn.SetMax(ly.Pos.Pos)
		mx.SetMax(ru)
	}
	nt.MaxPos = mn
	nt.MaxPos = mx
}
