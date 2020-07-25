// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env

import "github.com/emer/etable/etensor"

// Env defines an interface for environments, which determine the nature and
// sequence of States that can be used as inputs to a model, and the Env
// also can accept Action responses from the model that affect state evolution.
//
// The Env encapsulates all of the counter management logic to advance
// the temporal state of the environment, using TimeScales standard
// intervals.
//
// State is comprised of one or more Elements, each of which consists of an
// etensor.Tensor chunk of values that can be obtained by the model.
// Likewise, Actions can also have Elements.  The Step method is the main
// interface for advancing the Env state.  Counters should be queried
// after calling Step to see if any relevant values have changed, to trigger
// functions in the model (e.g., logging of prior statistics, etc).
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
	// case calls to Counters(), States(), and Actions() should all
	// return valid data.  It is essential that a model *always* check
	// this as a first step, because the Env will not generally check
	// for errors on any subsequent calls (for greater efficiency
	// and simplicity) and this call can also establish certain general
	// initialization settings that are not run-specific and thus make
	// sense to do once at this point, not every time during Init().
	Validate() error

	// Init initializes the environment for a given run of the model.
	// The environment may not care about the run number, but may implement
	// different parameterizations for different runs (e.g., between-subject
	// manipulations).  In general the Env can expect that the model will likely
	// have established a different random seed per run, prior to calling this
	// method, and that may be sufficient to enable different run-level behavior.
	// All other initialization / updating beyond this outer-most Run level must
	// be managed internally by the Env itself, and the model can query the
	// Counter state information to determine when things have updated at different
	// time scales.  See Step() for important info about state of env after Init
	// but prior to first Step() call.
	Init(run int)

	// Step generates the next step of environment state.
	// This is the main API for how the model interacts with the environment --
	// the env should update all other levels of state internally over
	// repeated calls to the Step method.
	// If there are no further inputs available, it returns false (most envs
	// typically only return true and just continue running as long as needed).
	//
	// The Env thus always reflects the *current* state of things, and this
	// call increments that current state, such that subsequent calls to
	// State(), Counter() etc will return this current state.
	// This implies that the state just after Init and prior to first Step
	// call should be an *initialized* state that then allows the first Step
	// call to establish the proper *first* state.  Typically this means that
	// one or more counters will be set to -1 during Init and then get incremented
	// to 0 on the first Step call.
	Step() bool

	// Counter(scale TimeScales) returns current counter state for given time scale,
	// the immediate previous counter state, and whether that time scale changed
	// during the last Step() function call (this may be true even if cur == prv, if
	// the Max = 1).  Use the Ctr struct for each counter, which manages all of this.
	// See external Counter* methods for Python-safe single-return-value versions.
	Counter(scale TimeScales) (cur, prv int, changed bool)

	// State returns the given element's worth of tensor data from the environment
	// based on the current state of the env, as a function of having called Step().
	// If no output is available on that element, then nil is returned.
	// The returned tensor must be treated as read-only as it likely points to original
	// source data -- please make a copy before modifying (e.g., Clone() methdod)
	State(element string) etensor.Tensor

	// Action sends tensor data about e.g., responses from model back to act
	// on the environment and influence its subsequent evolution.
	// The nature and timing of this input is paradigm dependent.
	Action(element string, input etensor.Tensor)
}

// EnvDesc is an interface that defines methods that describe an Env.
// These are optional for basic Env, but in cases where an Env
// should be fully self-describing, these methods can be implemented.
type EnvDesc interface {
	// Counters returns []TimeScales list of counters supported by this env.
	// These should be consistent within a paradigm and most models
	// will just expect particular sets of counters, but this can be
	// useful for sanity checking that a suitable env has been selected.
	// See SchemaFromScales function that takes this list of time
	// scales and returns an etable.Schema for Table columns to record
	// these counters in a log.
	Counters() []TimeScales

	// States returns a list of Elements of tensor outputs that this env
	// generates, specifying the unique Name and Shape of the data.
	// This information can be derived directly from an etable.Schema
	// and used for configuring model input / output pathways to fit
	// with those provided by the environment.  Depending on the
	// env paradigm, all elements may not be always available at every
	// point in time e.g., an env might alternate between Action and Reward
	// elements.  This may return nil if Env has not been properly
	// configured.
	States() Elements

	// Actions returns a list of elements of tensor inputs that this env
	// accepts, specifying the unique Name and Shape of the data.
	// Specific paradigms of envs can establish the timing and function
	// of these inputs, and how they then affect subsequent outputs
	// e.g., if the model is required to make a particular choice
	// response and then it can receive a reward or not contingent
	// on that choice.
	Actions() Elements
}

// CounterCur returns current counter state for given time scale
// this Counter for Python because it cannot process multiple return values
func CounterCur(en Env, scale TimeScales) int {
	cur, _, _ := en.Counter(scale)
	return cur
}

// CounterPrv returns previous counter state for given time scale
// this Counter for Python because it cannot process multiple return values
func CounterPrv(en Env, scale TimeScales) int {
	_, prv, _ := en.Counter(scale)
	return prv
}

// CounterChg returns whether counter changed during last Step()
// this Counter for Python because it cannot process multiple return values
func CounterChg(en Env, scale TimeScales) bool {
	_, _, chg := en.Counter(scale)
	return chg
}
