// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package deep

import (
	"github.com/emer/emergent/leabra/leabra"
)

// DeepLayer defines the essential algorithmic API for DeepLeabra at the layer level.
type DeepLayer interface {
	leabra.LeabraLayer

	// AsDeep returns this layer as a deep.Layer -- all derived layers must redefine
	// this to return the deep Layer type, so that the DeepLayer interface does not
	// need to include accessors to all the fields.
	AsDeep() *Layer

	// AvgMaxAttnGe computes the average and max AttnGe stats
	AvgMaxAttnGe(ltime *leabra.Time)

	// DeepAttnFmG computes DeepAttn and DeepLrn from AttnGe input,
	// and then applies the DeepAttn modulation to the Act activation value.
	DeepAttnFmG(ltime *leabra.Time)

	// AvgMaxActNoAttn computes the average and max ActNoAttn stats
	AvgMaxActNoAttn(ltime *leabra.Time)

	// DeepBurstFmAct updates DeepBurst layer 5 IB bursting value from current Act (superficial activation)
	// Subject to thresholding.
	DeepBurstFmAct(ltime *leabra.Time)

	// SendTRCBurstGeDelta sends change in DeepBurst activation since last sent, over BurstTRC
	// projections.
	SendTRCBurstGeDelta(ltime *leabra.Time)

	// TRCBurstGeFmInc computes the TRCBurstGe input from sent values
	TRCBurstGeFmInc(ltime *leabra.Time)

	// AvgMaxTRCBurstGe computes the average and max TRCBurstGe stats
	AvgMaxTRCBurstGe(ltime *leabra.Time)

	// SendDeepCtxtGe sends full DeepBurst activation over BurstCtxt projections to integrate
	// DeepCtxtGe excitatory conductance on deep layers.
	// This must be called at the end of the DeepBurst quarter for this layer.
	SendDeepCtxtGe(ltime *leabra.Time)

	// DeepCtxtFmGe integrates new DeepCtxtGe excitatory conductance from projections, and computes
	// overall DeepCtxt value.  This must be called at the end of the DeepBurst quarter for this layer,
	// after SendDeepCtxtGe.
	DeepCtxtFmGe(ltime *leabra.Time)

	// DeepBurstPrv saves DeepBurst as DeepBurstPrv
	DeepBurstPrv(ltime *leabra.Time)
}

// DeepPrjn defines the essential algorithmic API for DeepLeabra at the projection level.
type DeepPrjn interface {
	leabra.LeabraPrjn

	// SendDeepCtxtGe sends the full DeepBurst activation from sending neuron index si,
	// to integrate DeepCtxtGe excitatory conductance on receivers
	SendDeepCtxtGe(si int, dburst float32)

	// SendTRCBurstGeDelta sends the delta-DeepBurst activation from sending neuron index si,
	// to integrate TRCBurstGe excitatory conductance on receivers
	SendTRCBurstGeDelta(si int, delta float32)

	// SendAttnGeDelta sends the delta-activation from sending neuron index si,
	// to integrate into AttnGeInc excitatory conductance on receivers
	SendAttnGeDelta(si int, delta float32)

	// RecvDeepCtxtGeInc increments the receiver's DeepCtxtGe from that of all the projections
	RecvDeepCtxtGeInc()

	// RecvTRCBurstGeInc increments the receiver's TRCBurstGe from that of all the projections
	RecvTRCBurstGeInc()

	// RecvAttnGeInc increments the receiver's AttnGe from that of all the projections
	RecvAttnGeInc()

	// DWtDeepCtxt computes the weight change (learning) -- for DeepCtxt projections
	DWtDeepCtxt()
}
