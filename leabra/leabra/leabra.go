// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import (
	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/etensor"
)

// LeabraLayer defines the essential algorithmic API for Leabra, at the layer level.
// These are the methods that the leabra.Network calls on its layers at each step
// of processing.  Other Layer types can selectively re-implement (override) these methods
// to modify the computation, while inheriting the basic behavior for non-overridden methods.
//
// All of the structural API is in emer.Layer, which this interface also inherits for
// convenience.
type LeabraLayer interface {
	emer.Layer

	// AsLeabra returns this layer as a leabra.Layer -- all derived layers must redefine
	// this to return the base Layer type, so that the LeabraLayer interface does not
	// need to include accessors to all the basic stuff
	AsLeabra() *Layer

	// InitWts initializes the weight values in the network, i.e., resetting learning
	// Also calls InitActs
	InitWts()

	// InitActAvg initializes the running-average activation values that drive learning.
	InitActAvg()

	// InitActs fully initializes activation state -- only called automatically during InitWts
	InitActs()

	// InitWtsSym initializes the weight symmetry -- higher layers copy weights from lower layers
	InitWtSym()

	// InitExt initializes external input state -- called prior to apply ext
	InitExt()

	// ApplyExt applies external input in the form of an etensor.Float32
	// If the layer is a Target or Compare layer type, then it goes in Targ
	// otherwise it goes in Ext.
	ApplyExt(ext *etensor.Float32)

	// TrialInit handles all initialization at start of new input pattern, including computing
	// netinput scaling from running average activation etc.
	// should already have presented the external input to the network at this point.
	TrialInit()

	// AvgLFmAvgM updates AvgL long-term running average activation that drives BCM Hebbian learning
	AvgLFmAvgM()

	// GScaleFmAvgAct computes the scaling factor for synaptic conductance input
	// based on sending layer average activation.
	// This attempts to automatically adjust for overall differences in raw activity coming into the units
	// to achieve a general target of around .5 to 1 for the integrated G values.
	GScaleFmAvgAct()

	// GenNoise generates random noise for all neurons
	GenNoise()

	// DecayState decays activation state by given proportion (default is on ly.Act.Init.Decay)
	DecayState(decay float32)

	// HardClamp hard-clamps the activations in the layer -- called during TrialInit
	// for hard-clamped Input layers
	HardClamp()

	//////////////////////////////////////////////////////////////////////////////////////
	//  Cycle Methods

	// InitGInc initializes synaptic conductance increments -- optional
	InitGInc()

	// SendGDelta sends change in activation since last sent, to increment recv
	// synaptic conductances G, if above thresholds
	SendGDelta(ltime *Time)

	// GFmInc integrates new synaptic conductances from increments sent during last SendGDelta
	GFmInc(ltime *Time)

	// AvgMaxGe computes the average and max Ge stats, used in inhibition
	AvgMaxGe(ltime *Time)

	// InhibiFmGeAct computes inhibition Gi from Ge and Act averages within relevant Pools
	InhibFmGeAct(ltime *Time)

	// ActFmG computes rate-code activation from Ge, Gi, Gl conductances
	// and updates learning running-average activations from that Act
	ActFmG(ltime *Time)

	// AvgMaxAct computes the average and max Act stats, used in inhibition
	AvgMaxAct(ltime *Time)

	//////////////////////////////////////////////////////////////////////////////////////
	//  Quarter Methods

	// QuarterFinal does updating after end of a quarter
	QuarterFinal(ltime *Time)

	// CosDiffFmActs computes the cosine difference in activation state between minus and plus phases.
	// this is also used for modulating the amount of BCM hebbian learning
	CosDiffFmActs()

	// DWt computes the weight change (learning) -- calls DWt method on sending projections
	DWt()

	// WtFmDWt updates the weights from delta-weight changes -- on the sending projections
	WtFmDWt()

	// WtBalFmWt computes the Weight Balance factors based on average recv weights
	WtBalFmWt()
}

// LeabraPrjn defines the essential algorithmic API for Leabra, at the projection level.
// These are the methods that the leabra.Layer calls on its prjns at each step
// of processing.  Other Prjn types can selectively re-implement (override) these methods
// to modify the computation, while inheriting the basic behavior for non-overridden methods.
//
// All of the structural API is in emer.Prjn, which this interface also inherits for
// convenience.
type LeabraPrjn interface {
	emer.Prjn

	// AsLeabra returns this prjn as a leabra.Prjn -- all derived prjns must redefine
	// this to return the base Prjn type, so that the LeabraPrjn interface does not
	// need to include accessors to all the basic stuff.
	AsLeabra() *Prjn

	// InitWts initializes weight values according to Learn.WtInit params
	InitWts()

	// InitWtSym initializes weight symmetry -- is given the reciprocal projection where
	// the Send and Recv layers are reversed.
	InitWtSym(rpj LeabraPrjn)

	// InitGInc initializes the per-projection synaptic conductance threadsafe increments.
	// This is not typically needed (called during InitWts only) but can be called when needed
	InitGInc()

	// SendGDelta sends the delta-activation from sending neuron index si,
	// to integrate synaptic conductances on receivers
	SendGDelta(si int, delta float32)

	// RecvGInc increments the receiver's synaptic conductances from those of all the projections.
	RecvGInc()

	// DWt computes the weight change (learning) -- on sending projections
	DWt()

	// WtFmDWt updates the synaptic weight values from delta-weight changes -- on sending projections
	WtFmDWt()

	// WtBalFmWt computes the Weight Balance factors based on average recv weights
	WtBalFmWt()
}
