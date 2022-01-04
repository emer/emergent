// Copyright (c) 2021 The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chem

// SimpleEnz models a simple enzyme-catalyzed reaction
// that transforms S = substrate into P product via E which is not consumed
// assuming there is much more E than S and P -- E effectively acts as a
// rate constant multiplier
//
//   Kf*E
// S ----> P
//
// S = substrate, E = enzyme, P = product, Kf is the rate of the reaction
type SimpleEnz struct {
	Kf float64 `desc:"S->P forward rate constant, in μM-1 msec-1"`
}

// Step computes delta S and P values based on current S, E values
func (rt *SimpleEnz) Step(cs, ce float64, ds, dp *float64) {
	df := rt.Kf * cs * ce // forward
	*ds -= df
	*dp += df
}
