// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

import (
	"fmt"
	"io"

	"cogentcore.org/core/base/reflectx"
	"github.com/emer/emergent/v2/params"
	"github.com/emer/emergent/v2/paths"
	"github.com/emer/emergent/v2/weights"
)

// Path defines the basic interface for a pathway which connects two layers.
// Name is set automatically to: SendLay().Name() + "To" + RecvLay().Name()
type Path interface {
	params.Styler // TypeName, Name, and Class methods for parameter styling

	// Init MUST be called to initialize the path's pointer to itself as an emer.Path
	// which enables the proper interface methods to be called.
	Init(path Path)

	// SendLay returns the sending layer for this pathway
	SendLay() Layer

	// RecvLay returns the receiving layer for this pathway
	RecvLay() Layer

	// Pattern returns the pattern of connectivity for interconnecting the layers
	Pattern() paths.Pattern

	// SetPattern sets the pattern of connectivity for interconnecting the layers.
	// Returns Path so it can be chained to set other properties too
	SetPattern(pat paths.Pattern) Path

	// Type returns the functional type of pathway according to PathType (extensible in
	// more specialized algorithms)
	Type() PathType

	// SetType sets the functional type of pathway according to PathType
	// Returns Path so it can be chained to set other properties too
	SetType(typ PathType) Path

	// PathTypeName returns the string rep of functional type of pathway
	// according to PathType (extensible in more specialized algorithms, by
	// redefining this method as needed).
	PathTypeName() string

	// AddClass adds a CSS-style class name(s) for this path,
	// ensuring that it is not a duplicate, and properly space separated.
	// Returns Path so it can be chained to set other properties too
	AddClass(cls ...string) Path

	// Label satisfies the core.Labeler interface for getting the name of objects generically
	Label() string

	// IsOff returns true if pathway or either send or recv layer has been turned Off.
	// Useful for experimentation
	IsOff() bool

	// SetOff sets the pathway Off status (i.e., lesioned). Careful: Layer.SetOff(true) will
	// reactivate that layer's pathways, so pathway-level lesioning should always be called
	// after layer-level lesioning.
	SetOff(off bool)

	// SynVarNames returns the names of all the variables on the synapse
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
	// Note: this is a global list so do not modify!
	SynVarProps() map[string]string

	// SynIndex returns the index of the synapse between given send, recv unit indexes
	// (1D, flat indexes). Returns -1 if synapse not found between these two neurons.
	// This requires searching within connections for receiving unit (a bit slow).
	SynIndex(sidx, ridx int) int

	// SynVarIndex returns the index of given variable within the synapse,
	// according to *this path's* SynVarNames() list (using a map to lookup index),
	// or -1 and error message if not found.
	SynVarIndex(varNm string) (int, error)

	// SynVarNum returns the number of synapse-level variables
	// for this paths.  This is needed for extending indexes in derived types.
	SynVarNum() int

	// Syn1DNum returns the number of synapses for this path as a 1D array.
	// This is the max idx for SynVal1D and the number of vals set by SynValues.
	Syn1DNum() int

	// SynVal1D returns value of given variable index (from SynVarIndex) on given SynIndex.
	// Returns NaN on invalid index.
	// This is the core synapse var access method used by other methods,
	// so it is the only one that needs to be updated for derived layer types.
	SynVal1D(varIndex int, synIndex int) float32

	// SynValues sets values of given variable name for each synapse, using the natural ordering
	// of the synapses (sender based for Leabra),
	// into given float32 slice (only resized if not big enough).
	// Returns error on invalid var name.
	SynValues(vals *[]float32, varNm string) error

	// SynVal returns value of given variable name on the synapse
	// between given send, recv unit indexes (1D, flat indexes).
	// Returns math32.NaN() for access errors.
	SynValue(varNm string, sidx, ridx int) float32

	// SetSynVal sets value of given variable name on the synapse
	// between given send, recv unit indexes (1D, flat indexes).
	// Typically only supports base synapse variables and is not extended
	// for derived types.
	// Returns error for access errors.
	SetSynValue(varNm string, sidx, ridx int, val float32) error

	// Defaults sets default parameter values for all Path parameters
	Defaults()

	// UpdateParams() updates parameter values for all Path parameters,
	// based on any other params that might have changed.
	UpdateParams()

	// ApplyParams applies given parameter style Sheet to this pathway.
	// Calls UpdateParams if anything set to ensure derived parameters are all updated.
	// If setMsg is true, then a message is printed to confirm each parameter that is set.
	// it always prints a message if a parameter fails to be set.
	// returns true if any params were set, and error if there were any errors.
	ApplyParams(pars *params.Sheet, setMsg bool) (bool, error)

	// SetParam sets parameter at given path to given value.
	// returns error if path not found or value cannot be set.
	SetParam(path, val string) error

	// NonDefaultParams returns a listing of all parameters in the Projection that
	// are not at their default values -- useful for setting param styles etc.
	NonDefaultParams() string

	// AllParams returns a listing of all parameters in the Projection
	AllParams() string

	// WriteWtsJSON writes the weights from this pathway from the receiver-side perspective
	// in a JSON text format.  We build in the indentation logic to make it much faster and
	// more efficient.
	WriteWtsJSON(w io.Writer, depth int)

	// ReadWtsJSON reads the weights from this pathway from the receiver-side perspective
	// in a JSON text format.  This is for a set of weights that were saved *for one path only*
	// and is not used for the network-level ReadWtsJSON, which reads into a separate
	// structure -- see SetWts method.
	ReadWtsJSON(r io.Reader) error

	// SetWts sets the weights for this pathway from weights.Path decoded values
	SetWts(pw *weights.Path) error

	// Build constructs the full connectivity among the layers as specified in this pathway.
	Build() error
}

// Paths is a slice of pathways
type Paths []Path

// ElemLabel satisfies the core.SliceLabeler interface to provide labels for slice elements
func (pl *Paths) ElemLabel(idx int) string {
	if len(*pl) == 0 {
		return "(empty)"
	}
	if idx < 0 || idx >= len(*pl) {
		return ""
	}
	pj := (*pl)[idx]
	if reflectx.AnyIsNil(pj) {
		return "nil"
	}
	return pj.Name()
}

// Add adds a pathway to the list
func (pl *Paths) Add(p Path) {
	(*pl) = append(*pl, p)
}

// Send finds the pathway with given send layer
func (pl *Paths) Send(send Layer) (Path, bool) {
	for _, pj := range *pl {
		if pj.SendLay() == send {
			return pj, true
		}
	}
	return nil, false
}

// Recv finds the pathway with given recv layer
func (pl *Paths) Recv(recv Layer) (Path, bool) {
	for _, pj := range *pl {
		if pj.RecvLay() == recv {
			return pj, true
		}
	}
	return nil, false
}

// SendName finds the pathway with given send layer name, nil if not found
// see Try version for error checking.
func (pl *Paths) SendName(sender string) Path {
	pj, _ := pl.SendNameTry(sender)
	return pj
}

// RecvName finds the pathway with given recv layer name, nil if not found
// see Try version for error checking.
func (pl *Paths) RecvName(recv string) Path {
	pj, _ := pl.RecvNameTry(recv)
	return pj
}

// SendNameTry finds the pathway with given send layer name.
// returns error message if not found
func (pl *Paths) SendNameTry(sender string) (Path, error) {
	for _, pj := range *pl {
		if pj.SendLay().Name() == sender {
			return pj, nil
		}
	}
	return nil, fmt.Errorf("sending layer: %v not found in list of pathways", sender)
}

// SendNameTypeTry finds the pathway with given send layer name and Type string.
// returns error message if not found.
func (pl *Paths) SendNameTypeTry(sender, typ string) (Path, error) {
	for _, pj := range *pl {
		if pj.SendLay().Name() == sender {
			tstr := pj.PathTypeName()
			if tstr == typ {
				return pj, nil
			}
		}
	}
	return nil, fmt.Errorf("sending layer: %v, type: %v not found in list of pathways", sender, typ)
}

// RecvNameTry finds the pathway with given recv layer name.
// returns error message if not found
func (pl *Paths) RecvNameTry(recv string) (Path, error) {
	for _, pj := range *pl {
		if pj.RecvLay().Name() == recv {
			return pj, nil
		}
	}
	return nil, fmt.Errorf("receiving layer: %v not found in list of pathways", recv)
}

// RecvNameTypeTry finds the pathway with given recv layer name and Type string.
// returns error message if not found.
func (pl *Paths) RecvNameTypeTry(recv, typ string) (Path, error) {
	for _, pj := range *pl {
		if pj.RecvLay().Name() == recv {
			tstr := pj.PathTypeName()
			if tstr == typ {
				return pj, nil
			}
		}
	}
	return nil, fmt.Errorf("receiving layer: %v, type: %v not found in list of pathways", recv, typ)
}

//////////////////////////////////////////////////////////////////////////////////////
//  PathType

// PathType is the type of the pathway (extensible for more specialized algorithms).
// Class parameter styles automatically key off of these types.
type PathType int32 //enums:enum

// The pathway types
const (
	// Forward is a feedforward, bottom-up pathway from sensory inputs to higher layers
	Forward PathType = iota

	// Back is a feedback, top-down pathway from higher layers back to lower layers
	Back

	// Lateral is a lateral pathway within the same layer / area
	Lateral

	// Inhib is an inhibitory pathway that drives inhibitory synaptic inputs instead of excitatory
	Inhib
)
