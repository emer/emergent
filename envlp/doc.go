// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package envlp defines an interface for environments, which determine the
nature and sequence of States that can be used as inputs to a model
and it can also accept Action responses from the model that affect
how the enviroment evolves in the future.

This version uses the looper control framework to manage the incrementing
of counters on the Env, instead of the Env automatically incrementing counters
on its own, which is the behavior of the original `env.Env` environment.

By adhering to this interface, it is then easier to mix-and-match
environments with models.
*/
package envlp
