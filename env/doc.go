// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package env defines an interface for environments, which determine the nature and
sequence of States as inputs to a model. Action responses from the model
can also drive state evolution.

State is comprised of one or more Elements, each of which consists of an
tensor.Values chunk of values that can be obtained by the model.
Likewise, Actions can also have Elements. The Step method is the main
interface for advancing the Env state.

The standard String() string fmt.Stringer method must be defined to return
a string description of the current environment state, e.g., as a TrialName.
A Label() string method must be defined to return the Name of the environment,
which is typically the Mode of usage (Train vs. Test).

Typically each specific implementation of this Env interface will have
multiple parameters etc that can be modified to control env behavior:
all of this is paradigm-specific and outside the scope of this basic interface.

See e.g., env.FixedTable for particular implementation of a fixed Table
of patterns, for one example of a widely used paradigm.
*/
package env
