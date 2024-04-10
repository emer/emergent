// Copyright (c) 2021 The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chem

// SimpleEnz models a simple enzyme-catalyzed reaction
// that transforms S = substrate into P product via E which is not consumed
// assuming there is much more E than S and P -- E effectively acts as a
// rate constant multiplier
//
//	Kf*E
//
// S ----> P
//
// S = substrate, E = enzyme, P = product, Kf is the rate of the reaction
type SimpleEnz struct {

	// S->P forward rate constant, in μM-1 msec-1
	Kf float64
}

// SetVol sets reaction forward / backward time constants in seconds,
// dividing forward Kf by volume to compensate for 2 volume-based concentrations
// occurring in forward component, vs just 1 in back
func (rt *SimpleEnz) SetVol(f, vol float64) {
	rt.Kf = CoFromN(f, vol)
}

// Step computes delta S and P values based on current S, E values
func (rt *SimpleEnz) Step(cs, ce float64, ds, dp *float64) {
	df := rt.Kf * cs * ce // forward
	*ds -= df
	*dp += df
}

// StepCo computes delta S and P values based on current S, E values
// based on concentration
func (rt *SimpleEnz) StepCo(cs, ce, vol float64, ds, dp *float64) {
	df := rt.Kf * CoFromN(cs, vol) * CoFromN(ce, vol) // forward
	*ds -= df
	*dp += df
}

// StepK computes delta S and P values based on current S, E values
// K version has additional rate multiplier for Kf
func (rt *SimpleEnz) StepK(kf, cs, ce float64, ds, dp *float64) {
	df := kf * rt.Kf * cs * ce // forward
	*ds -= df
	*dp += df
}
