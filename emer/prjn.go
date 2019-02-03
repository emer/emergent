// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

import "github.com/emer/emergent/prjn"

// Prjn defines the basic interface for a projection which connects two layers
type Prjn interface {
	// RecvLay returns the receiving layer for this projection
	RecvLay() Layer

	// SendLay returns the sending layer for this projection
	SendLay() Layer

	// Pattern returns the pattern of connectivity for interconnecting the layers
	Pattern() prjn.Pattern

	// PrjnClass is for applying parameter styles, CSS-style -- can be space-separated multple tags
	PrjnClass() string

	// PrjnName is the automatic name of projection: RecvLay().LayName() + "Fm" + SendLay().LayName()
	PrjnName() string

	// Label satisfies the gi.Labeler interface for getting the name of objects generically
	Label() string

	// IsOff returns true if projection or either send or recv layer has been turned Off -- for experimentation
	IsOff() bool

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

	// StyleParams applies a given ParamStyle style sheet to the projections
	// If setMsg is true, then a message is printed to confirm each parameter that is set.
	// it always prints a message if a parameter fails to be set.
	StyleParams(psty ParamStyle, setMsg bool)
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
