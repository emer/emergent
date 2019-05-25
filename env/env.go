// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env

import "github.com/emer/etable/etensor"

// Env Defines an interface for environments, which determine the nature and
// sequence of inputs to a model (and responses from the model that are inputs
// to the environment).
//
// By adhering to this interface, it is then easier to mix-and-match environments
// with models.
//
// The overall division of labor is that the model keeps track of the outer-most
// Run time-scale depending on its own parameters and learning trajectory
// and the environment is responsible for generating patterns for each run.
//
// Multiple different environments will typically be used in a model, e.g.,
// one for training and other(s) for testing.  Even if these envs all share
// a common database of patterns, a different Env should be used for each
// case where different counters and sequences of events etc are presented.
//
// Thus, the Env encapsulates all of the counter management logic for each
// aspect of model training and testing, so that the model itself just
// needs to manange which Env to use, when, and manage the connection of
// the Env Outputs as inputs to the model, and vice-versa for Inputs to the
// Env coming from the model.
//
// The channel allows annotation about multiple possible channels of I/O
// and any given I/O event can specify which channels are provided.
// Particular paradigms of environments must establish naming conventions
// for these channels which then allow the model to use the information
// appropriately -- the Env interface only provides the most basic framework
// for establishing these paradigms, and ultimately a given model will only
// work within a particular paradigm of environments following specific
// conventions.
//
// See e.g., env.FixedTable for particular implementation of a fixed Table
// of patterns, for one example of a widely-used paradigm.
//
// Typically each specific implementation of this Env interface will have
// multiple parameters etc that can be modified to control env behavior --
// all of this is paradigm-specific and outside the scope of this basic interface.
type Env interface {
	// Name returns a name for this environment, which can be useful
	// for selecting from a list of options etc.
	Name() string

	// Desc returns an (optional) brief description of this particular
	// environment
	Desc() string

	// Validate checks if the various specific parameters for this
	// Env have been properly set -- if not, error message(s) will
	// be returned.  If everything is OK, nil is returned, in which
	// case calls to Counters(), Outputs(), and Inputs() should all
	// return valid data.  It is essential that a model *always* check
	// this as a first step, because the Env will not generally check
	// for errors on any subsequent calls (for greater efficiency
	// and simplicity) and this call can also establish certain general
	// initialization settings that are not run-specific and thus make
	// sense to do once at this point, not every time during Init().
	Validate() error

	// Counters returns []TimeScales list of counters supported by this env.
	// These should be consistent with in a paradigm and most models
	// will just expect particular sets of counters, but this can be
	// useful for sanity checking that a suitable env has been selected.
	// See SchemaFromScales for function that takes this list of time
	// scales and returns an etable.Schema for Table columns to record
	// these counters in a log.
	Counters() []TimeScales

	// Outputs returns a list of channels of tensor outputs that this env
	// generates, specifying the unique Name and Shape of the data.
	// This information can be derived directly from an etable.Schema
	// and used for configuring model input / output pathways to fit
	// with those provided by the environment.  Depending on the
	// env paradigm, all channels may not be always available at every
	// point in time e.g., an env my alternate between Input and Reward
	// channels.  This may return nil if Env has not been properly
	// configured.
	Outputs() Channels

	// Inputs returns a list of channels of tensor inputs that this env
	// accepts, specifying the unique Name and Shape of the data (could be nil).
	// Specific paradigms of envs can establish the timing and function
	// of these inputs, and how they then affect subsequent outputs
	// e.g., if the model is required to make a particular choice
	// response and then it can receive a reward or not contingent
	// on that choice.
	Inputs() Channels

	// Init initializes the environment for a given run of the model.
	// The environment may not care about the run number, but may implement
	// different parameterizations for different runs (e.g., between-subject
	// manipulations).  In general the Env can expect that the model will likely
	// have established a different random seed per run, prior to calling this
	// method, and that may be sufficient to enable different run-level behavior.
	// All other initialization / updating beyond this outer-most Run level must
	// be managed internally by the Env itself, and the model can query the
	// Counter state information to determine when things have updated at different
	// time scales.  See Next() for important info about state of env after Init
	// but prior to first Next() call.
	Init(run int)

	// Next generates the next step of environment state.
	// This is the main API for how the model interacts with the environment --
	// the env should update all other levels of state internally over
	// repeated calls to the Next method.
	// If there are no further inputs available, it returns false (most envs
	// typically only return true and just continue running as long as needed).
	//
	// The Env thus always reflects the *current* state of things, and this
	// call increments that current state, such that subsequent calls to
	// Output(), Counter() etc will return this current state.
	// This implies that the state just after Init and prior to first Next
	// call should be an *initialized* state that then allows the first Next
	// call to establish the proper *first* state.  Typically this means that
	// one or more counters will be set to -1 during Init and then get incremented
	// to 0 on the first Next call.
	Next() bool

	// Output returns the given channel's worth of tensor data from the environment
	// based on the current state of the env, as a function of having called Next().
	// If no output is available on that channel, then nil is returned.
	// The returned tensor must be treated as read-only as it likely points to original
	// source data -- please make a copy before modifying.
	Output(channel string) etensor.Tensor

	// Input sends tensor data about e.g., responses from model back to environment.
	// The nature and timing of this input is paradigm dependent.
	Input(channel string, input etensor.Tensor)

	// Counter(scale TimeScales) returns current counter state for given time scale,
	// and whether that time scale has changed since the last time it was queried.
	// Thus, each Env must store a record of last-queried counters -- use the
	// Ctr struct for each counter which handles this.
	Counter(scale TimeScales) (val int, changed bool)
}
