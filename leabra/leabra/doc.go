// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package leabra provides the basic reference leabra implementation, for rate-coded
activations and standard error-driven learning.  Other packages provide spiking
or deep leabra, PVLV, PBWM, etc.

The overall design seeks an "optimal" tradeoff between simplicity, transparency, ability to flexibly
recombine and extend elements, and avoiding having to rewrite a bunch of stuff.

The *Stru elements handle the core structural components of the network, and hold
emer.* interface pointers to elements such as emer.Layer, which provides a very minimal
interface for these elements.  Interfaces are automatically pointers, so think of these
as generic pointers to your specific Layers etc.

This design means the same *Stru infrastructure can be re-used across different variants
of the algorithm.  Because we're keeping this infrastructure minimal and algorithm-free
it should be much less confusing than dealing with the multiple levels of inheritance
in C++ emergent.  The actual algorithm-specific code is now fully self-contained,
and largely orthogonalized from the infrastructure.

One specific cost of this is the need to cast the emer.* interface pointers into
the specific types of interest, when accessing via the *Stru infrastructure.

The *Params elements contain all the (meta)parameters and associated methods for computing
various functions.  They are the equivalent of Specs from original emergent, but unlike specs
they are local to each place they are used, and styling is used to apply common parameters
across multiple layers etc.  Params seems like a more explicit, recognizable name compared
to specs, and this also helps avoid confusion about their different nature than old specs.
Pars is shorter but confusable with "Parents" so "Params" is more unambiguous.

Params are organized into four major categories, which are more clearly functionally
labeled as opposed to just structurally so, to keep things clearer and better organized
overall:
* ActParams -- activation params, at the Neuron level (in act.go)
* InhibParams -- inhibition params, at the Layer / Pool level (in inhib.go)
* LearnNeurParams -- learning parameters at the Neuron level (running-averages that drive learning)
* LearnSynParams -- learning parameters at the Synapse level (both in learn.go)

The levels of structure and state are:
* Network
*   .Layers
*      .Pools: pooled inhibition state -- 1 for layer plus 1 for each sub-pool (unit group) with inhibition
*      .RecvPrjns: receiving projections from other sending layers
*      .SendPrjns: sending projections from other receiving layers
*      .Neurons: neuron state variables

There are methods on the Network that perform initialization and overall computation, by
iterating over layers and calling methods there.  This is typically how most users will
run their models.

Parallel computation across multiple CPU cores (threading) is achieved through persistent
worker go routines that listen for functions to run on thread-specific channels.  Each
layer has a designated thread number, so you can experiment with different ways of
dividing up the computation.  Timing data is kept for per-thread time use -- see TimeReport()
on the network.

The Layer methods directly iterate over Neurons, Pools, and Prjns, and there is no
finer-grained level of computation (e.g., at the individual Neuron level), except for
the *Params methods that directly compute relevant functions.  Thus, looking directly at
the layer.go code should provide a clear sense of exactly how everything is computed --
you may need to the refer to act.go, learn.go etc to see the relevant details but at
least the overall organization should be clear in layer.go.

Computational methods are generally named: VarFmVar to specifically name what variable
is being computed from what other input variables.  e.g., ActFmG computes activation from
conductances G.

The Pools (type Pool, in pool.go) hold state used for computing pooled inhibition, but also are
used to hold overall aggregate pooled state variables -- the first element in Pools applies
to the layer itself, and subsequent ones are for each sub-pool (4D layers).
These pools play the same role as the LeabraUnGpState structures in C++ emergent.

Prjns directly support all synapse-level computation, and hold the LearnSynParams and
iterate directly over all of their synapses.  It is the exact same Prjn object that lives
in the RecvPrjns of the receiver-side, and the SendPrjns of the sender-side, and it maintains
and coordinates both sides of the state.  This clarifies and simplifies a lot of code.
There is no separate equivalent of LeabraConSpec / LeabraConState at the level of
connection groups per unit per projection.

The pattern of connectivity between units is specified by the prjn.Pattern interface
and all the different standard options are avail in that prjn package.  The Pattern
code generates a full tensor bitmap of binary 1's and 0's for connected (1's) and not
(0's) units, and can use any method to do so.  This full lookup-table approach is not the most
memory-efficient, but it is fully general and shouldn't be too-bad memory-wise overall (fully
bit-packed arrays are used, and these bitmaps don't need to be retained once connections have
been established).  This approach allows patterns to just focus on patterns, and they don't care
at all how they are used to allocate actual connections.

*/
package leabra
