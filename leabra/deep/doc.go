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

* The alpha-trial quarter(s) when DeepBurst is updated and broadcast is set in
DeepBurstParams.BurstQtr (defaults to Q4, can also be e.g., Q2 and Q4 for beta
frequency updating).  During this quarter(s), the DeepBurst from Super layers is
continuously sent via BurstTRC projections to TRC layers (using efficient
delta-based computation) to drive plus-phase outcome states in those layers.
At the end of the burst quarter(s), BurstCtxt projections convey the DeepBurst
signal to Deep layer neurons, where it is integrated into the DeepCtxt value
representing the temporally-delayed context information.  Note: Deep layers also
compute a DeepBurst value themselves, which can be sent via self projections to
relfect the extensive deep-to-deep lateral connectivity that provides more extensive
temporal context information.

* Deep layer neurons reflect the layer 6 regular spiking CT corticothalamic neurons
that project into the thalamus, and back up to all the other lamina within a
microcolumn, where they drive a multiplicative attentional modulation signal.
These neurons receive the DeepBurst activation via a BurstCtxt projection type,
typically once every 100 msec, and integrate that in the DeepCtxt value,
which is added to other excitatory conductance inputs to drive the overall
activation (Act) of these neurons.  Due to the bursting nature
of the DeepBurst inputs, this causes these Deep layer neurons to reflect what
the superficial layers encoded on the *previous* timestep -- thus they represent
a temporally-delayed context state.

* Deep layer neurons project to the TRC (Pulvinar) neurons via standard Act-driven
projections that integrate into standard Ge excitatory input in TRC neurons,
to drive the prediction aspect of predictive learning.
They also can project back to the Super layer neurons via a DeepAttn projection to
drive attentional modulation of activity there.

* TRC layer neurons receive a BurstTRC projection from the Super layer (typically
a one-to-one projection), which drives the plus-phase "outcome" activation state
of these Pulvinar layers (Super actually computes the 5IB DeepBurst activation).
These layers also receive regular connections from Deep layers, which drive the
prediction of this plus-phase outcome state, based on the temporally-delayed deep
layer context information.

* The attentional effects are implemented via DeepAttn projections from
Deep to Super layers, which are typically fixed, non-learning, one-to-one projections,
that drive the AttnGe excitatory condutance in Super layers.
AttnGe then drives the computation of DeepAttn and DeepLrn values that modulate
(i.e., multiply) the activation (DeepAttn) or learning rate (DeepLrn) of these
superficial neurons.

All of the relevant parameters are in the deep.go file, in the Deep*Params classes,
which are then fields in the deep.Layer.

* DeepBurstParams (layer DeepBurst) has the BurstQtr when DeepBurst is updated,
and the thresholding parameters.

* DeepCtxtParams (layer DeepCtxt) has parameters for integrating DeepCtxt input

* DeepTRCParams (layer DeepTRC) has parameters for how to compute TRC plus phase
activation states based on the TRCBurstGe excitatory input from the BurstTRC projections.

* DeepAttnParams (layer DeepAttn) has the parameters for computing DeepAttn and
DeepLrn from AttnGe

*/
package deep
