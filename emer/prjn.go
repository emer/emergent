// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

import (
	"io"

	"github.com/emer/emergent/prjn"
	"github.com/goki/ki/kit"
)

// Prjn defines the basic interface for a projection which connects two layers
type Prjn interface {
	// Init MUST be called to initialize the prjn's pointer to itself as an emer.Prjn
	// which enables the proper interface methods to be called.
	Init(prjn Prjn)

	// RecvLay returns the receiving layer for this projection
	RecvLay() Layer

	// SendLay returns the sending layer for this projection
	SendLay() Layer

	// Pattern returns the pattern of connectivity for interconnecting the layers
	Pattern() prjn.Pattern

	// PrjType returns the functional type of projection according to PrjnType (extensible in
	// more specialized versions of Leabra)
	PrjType() PrjnType

	// SetType sets the functional type of projection according to PrjnType
	SetType(typ PrjnType)

	// Connect sets the basic connection parameters for this projection (send, recv, pattern, and type)
	Connect(send, recv Layer, pat prjn.Pattern, typ PrjnType)

	// PrjnClass is for applying parameter styles, CSS-style -- can be space-separated multple tags
	PrjnClass() string

	// SetClass sets CSS-style class name(s) for this projection (space-separated if multiple)
	SetClass(cls string)

	// PrjnName is the automatic name of projection: RecvLay().LayName() + "Fm" + SendLay().LayName()
	PrjnName() string

	// Label satisfies the gi.Labeler interface for getting the name of objects generically
	Label() string

	// IsOff returns true if projection or either send or recv layer has been turned Off -- for experimentation
	IsOff() bool

	// SetOff sets the projection Off status (i.e., lesioned)
	SetOff(off bool)

	// SynVarNames returns the names of all the variables on the synapse
	SynVarNames() []string

	// SynVals returns values of given variable name on synapses for each synapse in the projection
	// using the natural ordering of the synapses (sender based for Leabra)
	SynVals(varnm string) []float32

	// SynVal returns value of given variable name on the synapse between given recv unit index
	// and send unit index
	SynVal(varnm string, ridx, sidx int) (float32, error)

	// Defaults sets default parameter values for all Prjn parameters
	Defaults()

	// UpdateParams() updates parameter values for all Prjn parameters,
	// based on any other params that might have changed.
	UpdateParams()

	// SetParams sets given parameters to this prjn, if the target type is Prjn
	// calls UpdateParams to ensure derived parameters are all updated.
	// If setMsg is true, then a message is printed to confirm each parameter that is set.
	// it always prints a message if a parameter fails to be set.
	SetParams(pars Params, setMsg bool) bool

	// StyleParam applies a given style to this projection
	// depending on the style specification (.Class, #Name, Type) and target value of params.
	// .PrjType is automatically recognized as a .Class type (e.g., .Forward vs. .Back etc)
	// If setMsg is true, then a message is printed to confirm each parameter that is set.
	// it always prints a message if a parameter fails to be set.
	StyleParam(sty string, pars Params, setMsg bool) bool

	// StyleParams applies a given ParamStyle style sheet to the projections
	// depending on the style specification (.Class, #Name, Type) and target value of params.
	// .PrjType is automatically recognized as a .Class type (e.g., .Forward vs. .Back etc)
	// If setMsg is true, then a message is printed to confirm each parameter that is set.
	// it always prints a message if a parameter fails to be set.
	StyleParams(psty ParamStyle, setMsg bool)

	// WriteWtsJSON writes the weights from this projection from the receiver-side perspective
	// in a JSON text format.  We build in the indentation logic to make it much faster and
	// more efficient.
	WriteWtsJSON(w io.Writer, depth int)

	// ReadWtsJSON reads the weights from this projection from the receiver-side perspective
	// in a JSON text format.
	ReadWtsJSON(r io.Reader) error

	// Build constructs the full connectivity among the layers as specified in this projection.
	Build() error
}

// PrjnList is a slice of projections
type PrjnList []Prjn

// Add adds a projection to the list
func (pl *PrjnList) Add(p Prjn) {
	(*pl) = append(*pl, p)
}

// FindSend finds the projection with given send layer
func (pl *PrjnList) FindSend(send Layer) (Prjn, bool) {
	for _, pj := range *pl {
		if pj.SendLay() == send {
			return pj, true
		}
	}
	return nil, false
}

// FindRecv finds the projection with given recv layer
func (pl *PrjnList) FindRecv(recv Layer) (Prjn, bool) {
	for _, pj := range *pl {
		if pj.RecvLay() == recv {
			return pj, true
		}
	}
	return nil, false
}

// FindSendName finds the projection with given send layer name
func (pl *PrjnList) FindSendName(sender string) (Prjn, bool) {
	for _, pj := range *pl {
		if pj.SendLay().LayName() == sender {
			return pj, true
		}
	}
	return nil, false
}

// FindRecvName finds the projection with given recv layer name
func (pl *PrjnList) FindRecvName(recv string) (Prjn, bool) {
	for _, pj := range *pl {
		if pj.RecvLay().LayName() == recv {
			return pj, true
		}
	}
	return nil, false
}

//////////////////////////////////////////////////////////////////////////////////////
//  PrjnType

// PrjnType is the type of the projection (extensible for more specialized algorithms).
// Class parameter styles automatically key off of these types.
type PrjnType int32

//go:generate stringer -type=PrjnType

var KiT_PrjnType = kit.Enums.AddEnum(PrjnTypeN, false, nil)

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
