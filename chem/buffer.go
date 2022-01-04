// Copyright (c) 2021 The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chem

// Buffer provides a soft buffering driving deltas relative to a target N
// which can be set by concentration and volume.
type Buffer struct {
	K    float64 `desc:"rate of buffering (akin to permeability / conductance of a channel)"`
	Targ float64 `desc:"buffer target concentration -- drives delta relative to this"`
}

func (bf *Buffer) SetTargVol(targ, vol float64) {
	bf.Targ = CoToN(targ, vol)
}

// Step computes da delta for current value ca relative to target value Targ
func (bf *Buffer) Step(ca float64, da *float64) {
	*da += bf.K * (bf.Targ - ca)
}
