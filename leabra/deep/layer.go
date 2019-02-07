// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package deep

import (
	"github.com/emer/emergent/emer"
	"github.com/goki/ki/kit"
)

//////////////////////////////////////////////////////////////////////////////////////
//  LayerType

// DeepLeabra extensions to the emer.LayerType types

//go:generate stringer -type=LayerType

var KiT_LayerType = kit.Enums.AddEnum(LayerTypeN, false, nil)

func (ev LayerType) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *LayerType) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }

// The DeepLeabra layer types
const (
	// Super are superficial-layer neurons
	Super emer.LayerType = emer.LayerTypeN + iota

	// Deep are deep-layer neurons, reflecting activation of layer 6 CT corticothalamic
	// regular spiking neurons, which drive both attention in Super and predictions in Pulvinar / TRC
	Deep

	// TRC are thalamic relay cell neurons, typically in the Pulvinar, which alternately reflect
	// predictions driven by Deep layer projections, and actual outcomes driven by DeepBurst
	// projections from corresponding Super layer neurons that provide strong driving inputs to
	// TRC neurons.
	TRC

	LayerTypeN
)
