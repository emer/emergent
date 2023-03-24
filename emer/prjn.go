// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

import (
	"fmt"
	"io"

	"github.com/emer/emergent/params"
	"github.com/emer/emergent/prjn"
	"github.com/emer/emergent/weights"
	"github.com/goki/ki/kit"
)

// Prjn defines the basic interface for a projection which connects two layers.
// Name is set automatically to: SendLay().Name() + "To" + RecvLay().Name()
type Prjn interface {
	params.Styler // TypeName, Name, and Class methods for parameter styling

	// Init MUST be called to initialize the prjn's pointer to itself as an emer.Prjn
	// which enables the proper interface methods to be called.
	Init(prjn Prjn)

	// SendLay returns the sending layer for this projection
	SendLay() Layer

	// RecvLay returns the receiving layer for this projection
	RecvLay() Layer

	// Pattern returns the pattern of connectivity for interconnecting the layers
	Pattern() prjn.Pattern

	// SetPattern sets the pattern of connectivity for interconnecting the layers.
	// Returns Prjn so it can be chained to set other properties too
	SetPattern(pat prjn.Pattern) Prjn

	// Type returns the functional type of projection according to PrjnType (extensible in
	// more specialized algorithms)
	Type() PrjnType

	// SetType sets the functional type of projection according to PrjnType
	// Returns Prjn so it can be chained to set other properties too
	SetType(typ PrjnType) Prjn

	// PrjnTypeName returns the string rep of functional type of projection
	// according to PrjnType (extensible in more specialized algorithms, by
	// redefining this method as needed).
	PrjnTypeName() string

	// Connect sets the basic connection parameters for this projection (send, recv, pattern, and type)
	// Connect(send, recv Layer, pat prjn.Pattern, typ PrjnType)

	// SetClass sets CSS-style class name(s) for this projection (space-separated if multiple)
	// Returns Prjn so it can be chained to set other properties too
	SetClass(cls string) Prjn

	// Label satisfies the gi.Labeler interface for getting the name of objects generically
	Label() string

	// IsOff returns true if projection or either send or recv layer has been turned Off.
	// Useful for experimentation
	IsOff() bool

	// SetOff sets the projection Off status (i.e., lesioned). Careful: Layer.SetOff(true) will
	// reactivate that layer's projections, so projection-level lesioning should always be called
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

	// SynIdx returns the index of the synapse between given send, recv unit indexes
	// (1D, flat indexes). Returns -1 if synapse not found between these two neurons.
	// This requires searching within connections for receiving unit (a bit slow).
	SynIdx(sidx, ridx int) int

	// SynVarIdx returns the index of given variable within the synapse,
	// according to *this prjn's* SynVarNames() list (using a map to lookup index),
	// or -1 and error message if not found.
	SynVarIdx(varNm string) (int, error)

	// SynVarNum returns the number of synapse-level variables
	// for this prjn.  This is needed for extending indexes in derived types.
	SynVarNum() int

	// Syn1DNum returns the number of synapses for this prjn as a 1D array.
	// This is the max idx for SynVal1D and the number of vals set by SynVals.
	Syn1DNum() int

	// SynVal1D returns value of given variable index (from SynVarIdx) on given SynIdx.
	// Returns NaN on invalid index.
	// This is the core synapse var access method used by other methods,
	// so it is the only one that needs to be updated for derived layer types.
	SynVal1D(varIdx int, synIdx int) float32

	// SynVals sets values of given variable name for each synapse, using the natural ordering
	// of the synapses (sender based for Leabra),
	// into given float32 slice (only resized if not big enough).
	// Returns error on invalid var name.
	SynVals(vals *[]float32, varNm string) error

	// SynVal returns value of given variable name on the synapse
	// between given send, recv unit indexes (1D, flat indexes).
	// Returns mat32.NaN() for access errors.
	SynVal(varNm string, sidx, ridx int) float32

	// SetSynVal sets value of given variable name on the synapse
	// between given send, recv unit indexes (1D, flat indexes).
	// Typically only supports base synapse variables and is not extended
	// for derived types.
	// Returns error for access errors.
	SetSynVal(varNm string, sidx, ridx int, val float32) error

	// Defaults sets default parameter values for all Prjn parameters
	Defaults()

	// UpdateParams() updates parameter values for all Prjn parameters,
	// based on any other params that might have changed.
	UpdateParams()

	// ApplyParams applies given parameter style Sheet to this projection.
	// Calls UpdateParams if anything set to ensure derived parameters are all updated.
	// If setMsg is true, then a message is printed to confirm each parameter that is set.
	// it always prints a message if a parameter fails to be set.
	// returns true if any params were set, and error if there were any errors.
	ApplyParams(pars *params.Sheet, setMsg bool) (bool, error)

	// NonDefaultParams returns a listing of all parameters in the Projection that
	// are not at their default values -- useful for setting param styles etc.
	NonDefaultParams() string

	// AllParams returns a listing of all parameters in the Projection
	AllParams() string

	// WriteWtsJSON writes the weights from this projection from the receiver-side perspective
	// in a JSON text format.  We build in the indentation logic to make it much faster and
	// more efficient.
	WriteWtsJSON(w io.Writer, depth int)

	// ReadWtsJSON reads the weights from this projection from the receiver-side perspective
	// in a JSON text format.  This is for a set of weights that were saved *for one prjn only*
	// and is not used for the network-level ReadWtsJSON, which reads into a separate
	// structure -- see SetWts method.
	ReadWtsJSON(r io.Reader) error

	// SetWts sets the weights for this projection from weights.Prjn decoded values
	SetWts(pw *weights.Prjn) error

	// Build constructs the full connectivity among the layers as specified in this projection.
	Build() error
}

// Prjns is a slice of projections
type Prjns []Prjn

// ElemLabel satisfies the gi.SliceLabeler interface to provide labels for slice elements
func (pl *Prjns) ElemLabel(idx int) string {
	if len(*pl) == 0 {
		return "(empty)"
	}
	if idx < 0 || idx >= len(*pl) {
		return ""
	}
	pj := (*pl)[idx]
	if kit.IfaceIsNil(pj) {
		return "nil"
	}
	return pj.Name()
}

// Add adds a projection to the list
func (pl *Prjns) Add(p Prjn) {
	(*pl) = append(*pl, p)
}

// Send finds the projection with given send layer
func (pl *Prjns) Send(send Layer) (Prjn, bool) {
	for _, pj := range *pl {
		if pj.SendLay() == send {
			return pj, true
		}
	}
	return nil, false
}

// Recv finds the projection with given recv layer
func (pl *Prjns) Recv(recv Layer) (Prjn, bool) {
	for _, pj := range *pl {
		if pj.RecvLay() == recv {
			return pj, true
		}
	}
	return nil, false
}

// SendName finds the projection with given send layer name, nil if not found
// see Try version for error checking.
func (pl *Prjns) SendName(sender string) Prjn {
	pj, _ := pl.SendNameTry(sender)
	return pj
}

// RecvName finds the projection with given recv layer name, nil if not found
// see Try version for error checking.
func (pl *Prjns) RecvName(recv string) Prjn {
	pj, _ := pl.RecvNameTry(recv)
	return pj
}

// SendNameTry finds the projection with given send layer name.
// returns error message if not found
func (pl *Prjns) SendNameTry(sender string) (Prjn, error) {
	for _, pj := range *pl {
		if pj.SendLay().Name() == sender {
			return pj, nil
		}
	}
	return nil, fmt.Errorf("sending layer: %v not found in list of projections", sender)
}

// SendNameTypeTry finds the projection with given send layer name and Type string.
// returns error message if not found.
func (pl *Prjns) SendNameTypeTry(sender, typ string) (Prjn, error) {
	for _, pj := range *pl {
		if pj.SendLay().Name() == sender {
			tstr := pj.PrjnTypeName()
			if tstr == typ {
				return pj, nil
			}
		}
	}
	return nil, fmt.Errorf("sending layer: %v, type: %v not found in list of projections", sender, typ)
}

// RecvNameTry finds the projection with given recv layer name.
// returns error message if not found
func (pl *Prjns) RecvNameTry(recv string) (Prjn, error) {
	for _, pj := range *pl {
		if pj.RecvLay().Name() == recv {
			return pj, nil
		}
	}
	return nil, fmt.Errorf("receiving layer: %v not found in list of projections", recv)
}

// RecvNameTypeTry finds the projection with given recv layer name and Type string.
// returns error message if not found.
func (pl *Prjns) RecvNameTypeTry(recv, typ string) (Prjn, error) {
	for _, pj := range *pl {
		if pj.RecvLay().Name() == recv {
			tstr := pj.PrjnTypeName()
			if tstr == typ {
				return pj, nil
			}
		}
	}
	return nil, fmt.Errorf("receiving layer: %v, type: %v not found in list of projections", recv, typ)
}

//////////////////////////////////////////////////////////////////////////////////////
//  PrjnType

// PrjnType is the type of the projection (extensible for more specialized algorithms).
// Class parameter styles automatically key off of these types.
type PrjnType int32

//go:generate stringer -type=PrjnType

var KiT_PrjnType = kit.Enums.AddEnum(PrjnTypeN, kit.NotBitFlag, nil)

func (ev PrjnType) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *PrjnType) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }

// The projection types
const (
	// Forward is a feedforward, bottom-up projection from sensory inputs to higher layers
	Forward PrjnType = iota

	// Back is a feedback, top-down projection from higher layers back to lower layers
	Back

	// Lateral is a lateral projection within the same layer / area
	Lateral

	// Inhib is an inhibitory projection that drives inhibitory synaptic inputs instead of excitatory
	Inhib

	PrjnTypeN
)
