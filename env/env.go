// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env

import (
	"cogentcore.org/core/base/labels"
	"cogentcore.org/core/tensor"
)

//go:generate core generate -add-types

// Env defines an interface for environments, which determine the nature and
// sequence of States that can be used as inputs to a model, and the Env
// also can accept Action responses from the model that affect state evolution.
//
// The Env manages [Counter] values to advance the temporal state of the
// environment, using [etime.Times] standard intervals.
//
// State is comprised of one or more Elements, each of which consists of an
// tensor.Tensor chunk of values that can be obtained by the model.
// Likewise, Actions can also have Elements.  The Step method is the main
// interface for advancing the Env state.  Counters should be queried
// after calling Step to see if any relevant values have changed, to trigger
// functions in the model (e.g., logging of prior statistics, etc).
//
// Typically each specific implementation of this Env interface will have
// multiple parameters etc that can be modified to control env behavior --
// all of this is paradigm-specific and outside the scope of this basic interface.
type Env interface {
	labels.Labeler

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
	// This is the main API for how the model interacts with the environment.
	// The env should update all other levels of state internally over
	// repeated calls to the Step method.
	// If there are no further inputs available, it returns false (most envs
	// typically only return true and just continue running as long as needed).
	//
	// The Env thus always reflects the *current* state of things, and this
	// call increments that current state, such that subsequent calls to
	// State() will return this current state.
	//
	// This implies that the state just after Init and prior to first Step
	// call should be an *initialized* state that then allows the first Step
	// call to establish the proper *first* state.  Typically this means that
	// one or more counters will be set to -1 during Init and then get incremented
	// to 0 on the first Step call.
	Step() bool

	// State returns the given element's worth of tensor data from the environment
	// based on the current state of the env, as a function of having called Step().
	// If no output is available on that element, then nil is returned.
	// The returned tensor must be treated as read-only as it likely points to original
	// source data -- please make a copy before modifying (e.g., Clone() methdod).
	State(element string) tensor.Tensor

	// Action sends tensor data about e.g., responses from model back to act
	// on the environment and influence its subsequent evolution.
	// The nature and timing of this input is paradigm dependent.
	Action(element string, input tensor.Tensor)
}
