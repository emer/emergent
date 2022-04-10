// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package envlp

import (
	"github.com/emer/emergent/estats"
	"github.com/emer/emergent/etime"
	"github.com/emer/etable/etensor"
)

// Env defines an interface for environments, which determine the nature and
// sequence of States that can be used as inputs to a model, and the Env
// also can accept Action responses from the model that affect state evolution.
//
// The Env holds a set of Ctr counters for different time scales according
// to the etime scoping system, which are assumed to be managed by
// the looper.Stack looping control system.  The Env should only increment
// the inner-most counter that tracks the Stepping of the environment.
//
// State is comprised of one or more elements, each of which consists of an
// etensor.Tensor chunk of values that can be obtained by the model.
// Likewise, Actions provide tensor elements as input to the Env.
type Env interface {
	// Name returns a name for this environment, which can be useful
	// for selecting from a list of options etc.
	Name() string

	// Desc returns an (optional) brief description of this environment.
	Desc() string

	// Mode returns the evaluation mode (etime.Modes) for this environment
	// (Train, Test, etc). This is used for the Scope of the counters.
	Mode() string

	// Validate checks if the various specific parameters for this
	// Env have been properly set -- if not, error message(s) will
	// be returned.  If everything is OK, nil is returned, in which
	// case calls to Counters(), States(), and Actions() should all
	// return valid data.  It is essential that a model *always* check
	// this as a first step, because the Env will not generally check
	// for errors on any subsequent calls (for greater efficiency
	// and simplicity).
	Validate() error

	// Init initializes the environment at start of a new Run, preserving
	// the current Run level counter value, if that counter is present, but
	// resetting all other counters to 0.
	// In general the Env can expect that the Sim will have established a
	// different random seed per run, prior to calling this method,
	// sufficient to enable different run-level behavior.
	// The current State() must be updated to reflect the first step of
	// the environment, consistent with the post-increment model, where
	// Step is called *after* the current state is used.
	Init()

	// Step advances to the next step of environment state,
	// rendering any new State values as needed, and incrementing the
	// counter associated with stepping (e.g., Trial).
	// This is called *after* using the current State, making it ready
	// for the next iteration.
	// The looper control system will detect when the Trial is over Max
	// and reset that back to 0, while updating other higher counters as needed.
	// The Env should expect this and prepare a next state consistent with
	// the Trial (stepping level) counter reset back to 0.
	Step()

	// Counters returns the full map of counters used in the Env.
	// A specific scope counter can be accessed as Counters()[scope]
	Counters() Ctrs

	// Counter returns counter for given standard etime.Times value, using
	// the Mode set for this environment to generate a ScopeKey string.
	Counter(time etime.Times) *Ctr

	// String returns a string representation of the current step contents.
	// This is typically saved as a TrialName in the CtrsToStats function,
	// and satisfies the standard Stringer interface.
	String() string

	// State returns the given element's worth of tensor data from the environment
	// based on the current state of the env (prepared by Init or the last Step).
	// If no output is available on that element, then nil is returned.
	// The returned tensor must be treated as read-only as it likely points to original
	// source data -- please make a copy before modifying (e.g., Clone() method)
	State(element string) etensor.Tensor

	// Action sends tensor data about e.g., responses from model back to act
	// on the environment and influence its subsequent evolution.
	// The nature and timing of this input is paradigm dependent, but
	// in general it should happen prior to the Step() call, so that
	// Action sets values on the Env that are then used in the Step call
	// to generate the appropriate next state values.
	Action(element string, input etensor.Tensor)

	// CtrsToStats sets the current counter values to estats Int values
	// by their time names only (no eval Mode).  These values can then
	// be read by elog LogItems to record the counters in logs.
	// Typically, a TrialName string is also expected to be set,
	// to describe the current trial (Step) contents in a useful way,
	// and other relevant info (e.g., group / category info) can also be set.
	CtrsToStats(stats *estats.Stats)
}
