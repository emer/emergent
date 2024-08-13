// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

import (
	"io"

	"cogentcore.org/core/math32"
	"github.com/emer/emergent/v2/params"
	"github.com/emer/emergent/v2/paths"
	"github.com/emer/emergent/v2/weights"
)

// Path defines the minimal interface for a pathway
// which connects two layers, using a specific Pattern
// of connectivity, and with its own set of parameters.
// This supports visualization (NetView), I/O,
// and parameter setting functionality provided by emergent.
// Most of the standard expected functionality is defined in the
// PathBase struct, and this interface only has methods that must be
// implemented specifically for a given algorithmic implementation,
type Path interface {
	// StyleType, StyleClass, and StyleName methods for parameter styling.
	params.StylerObject

	// AsEmer returns the path as an *emer.PathBase,
	// to access base functionality.
	AsEmer() *PathBase

	// Label satisfies the core.Labeler interface for getting
	// the name of objects generically.
	Label() string

	// TypeName is the type or category of path, defined
	// by the algorithm (and usually set by an enum).
	TypeName() string

	// SendLayer returns the sending layer for this pathway,
	// as an emer.Layer interface.  The actual Path implmenetation
	// can use a Send field with the actual Layer struct type.
	SendLayer() Layer

	// RecvLayer returns the receiving layer for this pathway,
	// as an emer.Layer interface.  The actual Path implmenetation
	// can use a Recv field with the actual Layer struct type.
	RecvLayer() Layer

	// NumSyns returns the number of synapses for this path.
	// This is the max idx for SynValue1D and the number
	// of vals set by SynValues.
	NumSyns() int

	// SynIndex returns the index of the synapse between given send, recv unit indexes
	// (1D, flat indexes). Returns -1 if synapse not found between these two neurons.
	// This requires searching within connections for receiving unit (a bit slow).
	SynIndex(sidx, ridx int) int

	// SynVarNames returns the names of all the variables on the synapse
	// This is typically a global list so do not modify!
	SynVarNames() []string

	// SynVarNum returns the number of synapse-level variables
	// for this paths.  This is needed for extending indexes in derived types.
	SynVarNum() int

	// SynVarIndex returns the index of given variable within the synapse,
	// according to *this path's* SynVarNames() list (using a map to lookup index),
	// or -1 and error message if not found.
	SynVarIndex(varNm string) (int, error)

	// SynValues sets values of given variable name for each synapse,
	// using the natural ordering of the synapses (sender based for Axon),
	// into given float32 slice (only resized if not big enough).
	// Returns error on invalid var name.
	SynValues(vals *[]float32, varNm string) error

	// SynValue1D returns value of given variable index
	// (from SynVarIndex) on given SynIndex.
	// Returns NaN on invalid index.
	// This is the core synapse var access method used by other methods,
	// so it is the only one that needs to be updated for derived types.
	SynValue1D(varIndex int, synIndex int) float32

	// UpdateParams() updates parameter values for all Path parameters,
	// based on any other params that might have changed.
	UpdateParams()

	// SetParam sets parameter at given path to given value.
	// returns error if path not found or value cannot be set.
	SetParam(path, val string) error

	// AllParams returns a listing of all parameters in the Pathway.
	AllParams() string

	// WriteWeightsJSON writes the weights from this pathway
	// from the receiver-side perspective in a JSON text format.
	WriteWeightsJSON(w io.Writer, depth int)

	// SetWeights sets the weights for this pathway from weights.Path
	// decoded values
	SetWeights(pw *weights.Path) error
}

// PathBase defines the basic shared data for a pathway
// which connects two layers, using a specific Pattern
// of connectivity, and with its own set of parameters.
// The same struct token is added to the Recv and Send
// layer path lists,
type PathBase struct {
	// EmerPath provides access to the emer.Path interface
	// methods for functions defined in the PathBase type.
	// Must set this with a pointer to the actual instance
	// when created, using InitPath function.
	EmerPath Path

	// Name of the path, which can be automatically set to
	// SendLayer().Name + "To" + RecvLayer().Name via
	// SetStandardName method.
	Name string

	// Class is for applying parameter styles across multiple paths
	// that all get the same parameters.  This can be space separated
	// with multple classes.
	Class string

	// Info contains descriptive information about the pathway.
	// This is displayed in a tooltip in the network view.
	Info string

	// can record notes about this pathway here.
	Notes string

	// Pattern specifies the pattern of connectivity
	// for interconnecting the sending and receiving layers.
	Pattern paths.Pattern

	// Off inactivates this pathway, allowing for easy experimentation.
	Off bool

	// provides a history of parameters applied to the layer
	ParamsHistory params.HistoryImpl `table:"-"`
}

// InitPath initializes the path, setting the EmerPath interface
// to provide access to it for PathBase methods.
func InitPath(pt Path) {
	pb := pt.AsEmer()
	pb.EmerPath = pt
}

func (pt *PathBase) AsEmer() *PathBase { return pt }

// params.Styler:
func (pt *PathBase) StyleType() string  { return "Path" }
func (pt *PathBase) StyleClass() string { return pt.EmerPath.TypeName() + " " + pt.Class }
func (pt *PathBase) StyleName() string  { return pt.Name }
func (pt *PathBase) Label() string      { return pt.Name }

// AddClass adds a CSS-style class name(s) for this path,
// ensuring that it is not a duplicate, and properly space separated.
// Returns Path so it can be chained to set other properties too.
func (pt *PathBase) AddClass(cls ...string) *PathBase {
	pt.Class = params.AddClass(pt.Class, cls...)
	return pt
}

// SynValue returns value of given variable name on the synapse
// between given send, recv unit indexes (1D, flat indexes).
// Returns math32.NaN() for access errors.
func (pt *PathBase) SynValue(varNm string, sidx, ridx int) float32 {
	vidx, err := pt.EmerPath.SynVarIndex(varNm)
	if err != nil {
		return math32.NaN()
	}
	syi := pt.EmerPath.SynIndex(sidx, ridx)
	return pt.EmerPath.SynValue1D(vidx, syi)
}

////////////////////////////////////////////////////////////////////
//		Params

// ParamsHistoryReset resets parameter application history
func (pt *PathBase) ParamsHistoryReset() {
	pt.ParamsHistory.ParamsHistoryReset()
}

// ParamsApplied is just to satisfy History interface so reset can be applied
func (pt *PathBase) ParamsApplied(sel *params.Sel) {
	pt.ParamsHistory.ParamsApplied(sel)
}

// SetParam sets parameter at given path to given value.
// returns error if path not found or value cannot be set.
func (pt *PathBase) SetParam(path, val string) error {
	return params.SetParam(pt.EmerPath.StyleObject(), path, val)
}

// ApplyParams applies given parameter style Sheet to this pathway.
// Calls UpdateParams if anything set to ensure derived parameters are all updated.
// If setMsg is true, then a message is printed to confirm each parameter that is set.
// it always prints a message if a parameter fails to be set.
// returns true if any params were set, and error if there were any errors.
func (pt *PathBase) ApplyParams(pars *params.Sheet, setMsg bool) (bool, error) {
	app, err := pars.Apply(pt.EmerPath, setMsg)
	// note: must use EmerPath to get to actual Path, which then uses Styler interface
	// to return the Params struct.
	if app {
		pt.EmerPath.UpdateParams()
	}
	return app, err
}

// NonDefaultParams returns a listing of all parameters in the Layer that
// are not at their default values -- useful for setting param styles etc.
func (pt *PathBase) NonDefaultParams() string {
	// nds := reflectx.NonDefaultFields(ly.Params) // todo:
	nds := "non default field strings todo"
	//Str(ly.AxonLay.AsAxon().Params, ly.Name)
	return nds
}
