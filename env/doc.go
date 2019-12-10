// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package env defines an interface for environments, which determine the
nature and sequence of States that can be used as inputs to a model
and it can also accept Action responses from the model that affect
how the enviroment evolves in the future.

By adhering to this interface, it is then easier to mix-and-match
environments with models.

The overall division of labor is that the model keeps track of the outer-most
Run time-scale depending on its own parameters and learning trajectory
and the environment is responsible for generating patterns for each run.

Multiple different environments will typically be used in a model, e.g.,
one for training and other(s) for testing.  Even if these envs all share
a common database of patterns, a different Env should be used for each
case where different counters and sequences of events etc are presented,
which keeps them from interfering with each other.  Also, the etable.IdxView
can be used to allow multiple different Env's to all present different
indexed views into a shared common etable.Table (e.g., train / test splits).
The basic FixedTable env implementation uses this.

Thus, the Env encapsulates all of the counter management logic for each
aspect of model training and testing, so that the model itself just
needs to manange which Env to use, when, and manage the connection of
the Env States as inputs to the model, and vice-versa for Actions on the
Env coming from the model.

Each Element of the overall State allows annotation about the different
elements of state that are available in general, and the `Step` should
update all relevant state elements as appropriate, so these can be queried
by the user. Particular paradigms of environments must establish naming
conventions for these state elements which then allow the model to use
the information appropriately -- the Env interface only provides the most
basic framework for establishing these paradigms, and ultimately a given
model will only work within a particular paradigm of environments following
specific conventions.

See e.g., env.FixedTable for particular implementation of a fixed Table
of patterns, for one example of a widely-used paradigm.

Typically each specific implementation of this Env interface will have
multiple parameters etc that can be modified to control env behavior --
all of this is paradigm-specific and outside the scope of this basic interface.

See the emergent github wiki for more info:
https://github.com/emer/emergent/wiki/Env

*/
package env
