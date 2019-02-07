// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package deep

import "github.com/emer/emergent/leabra/leabra"

// deep.Neuron holds the extra neuron (unit) level variables for DeepLeabra computation.
// DeepLeabra includes both attentional and predictive learning functions of the deep layers
// and thalamocortical circuitry.
type Neuron struct {
	leabra.Neuron
	ActNoAttn    float32 `desc:"non-attention modulated activation of the superficial-layer neurons -- i.e., the activation prior to any modulation by the DeepAttn modulatory signal.  Using this as a driver of DeepBurst when there is DeepAttn modulation of superficial-layer activations prevents a positive-feedback loop that can be problematic."`
	DeepBurst    float32 `desc:"Deep layer bursting activation values, reflecting activity of layer 5b intrinsic bursting (5IB) neurons, which project into the thalamus.  Somewhat confusingly, this is computed on the Superficial layer neurons, as a thresholded function of the unit activation.  DeepBurst is only updated during the bursting quarter (typically the 4th quarter) of the alpha cycle, and it drives the DeepCtxt inputs at the end of that quarter, which then provide the new temporal context state for the separate deep layer (which reflects activation of layer 6 CT corticothalamic neurons)."`
	DeepBurstPrv float32 `desc:"DeepBurst from the previous alpha trial -- this is typically used for learning in the DeepContext projection into "`
	DeepBurstGe  float32 `desc:"Total excitatory conductance received from DeepBurst activations.  This is then integrated to drive DeepCtxt temporal context and attentional modulation signal."`
	DeepCtxt     float32 `desc:"Temporally-delayed local integration of DeepBurst signals sent via DeepCtxt projection into separate Deep layer neurons, which reflect the activation of layer 6 CT corticothalamic neurons.  This is computed from DeepBurstGe which is the total conductance from deep bursting inputs."`
	DeepAttn     float32 `desc:"Current attention modulatory value of deep layer 6 CT corticothalamic, regular spiking neurons that represents the net attentional filter applied to the superficial layers.  This value directly multiplies the superficial layer activations (Act) (ActNoAttn represents value prior to this multiplication).  Value is computed from DeepAttnGe received via DeepAttn projections from Deep layers."`
	DeepLrn      float32 `desc:"Version of DeepAttn that modulates learning rates instead of activations -- learning is assumed to be more strongly affected than activation, so this value, computed from DeepAttnGe, typically has a lower range than DeepAttn."`
	DeepAttnGe   float32 `desc:"Total excitatory conductance received from from deep layer activations (representing layer 6 regular spiking CT corticothalamic neurons) via DeepAttn projections.  This drives both DeepAttn and DeepLrn values."`
}
