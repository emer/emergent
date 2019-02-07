// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package deep provides the DeepLeabra variant of Leabra, which performs predictive
learning by attempting to predict the activation states over the Pulvinar nucleus
of the thalamus (in posterior sensory cortex), which are driven phasically every
100 msec by deep layer 5 intrinsic bursting (5IB) neurons that have strong focal
(essentially 1-to-1) connections onto the Pulvinar Thalamic Relay Cell (TRC)
neurons.

This package allows you to specify layer types as Super, Deep, and TRC
which in turn drives specific forms of computation associated with each
of those layer types.

DeepLeabra captures both the predictive learning and attentional modulation
functions of the deep layer and thalamocortical circuitry.

* Super layer neurons reflect the superficial layers of the neocortex, but they
also are the basis for directly computing the DeepBurst activation signal that
reflects the deep layer 5 IB bursting activation, via thresholding of the superficial
layer activations (Bursting is thought to have a higher threshold).

* Deep layer neurons reflect the layer 6 regular spiking CT corticothalamic neurons
that project into the thalamus, and back up to all the other lamina within a
microcolumn, driving a multiplicative attentional modulation signal.  These neurons
receive the DeepBurst activation, typically once every 100 msec, and integrate
that in the DeepCtxt value, which is added to other excitatory conductance inputs
to drive the overall activation (Act) of these neurons.  Due to the bursting nature
of the DeepBurst inputs, this causes these Deep layer neurons to reflect what
the superficial layers encoded on the *previous* timestep -- thus they represent
a temporally-delayed context state.

* Deep layer neurons project to the TRC (Pulvinar) neurons, to drive predictive
learning, and they also can project back to the Super layer neurons, to drive
attentional modulation of activity there.

* TRC layer neurons receive a DeepBurst projection from the Super layer (typically
a one-to-one projection), which drives the plus-phase "outcome" activation state
of these Pulvinar layers (Super actually computes the 5IB DeepBurst activation).
These layers also receive regular connections from Deep layers, which drive the
prediction of this plus-phase outcome state, based on the temporally-delayed deep
layer context information.

* The attentional effects are implemented via projections from Deep to Super layers,
typically fixed, non-learning, one-to-one kinds of projections, that drive the
DeepAttnGe excitatory condutance inputs, which are then translated into DeepAttn and
DeepLrn values that modulate (i.e., multiply) the activation (DeepAttn) or learning
rate (DeepLrn) of these superficial neurons.

*/
package deep
