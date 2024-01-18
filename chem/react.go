// Copyright (c) 2021 The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chem

//go:generate core generate -add-types

// React models a basic chemical reaction:
//
//	Kf
//
// A + B --> AB
//
//	<-- Kb
//
// where Kf is the forward and Kb is the backward time constant.
// The source Kf and Kb constants are in terms of concentrations μM-1 and sec-1
// but calculations take place using N's, and the forward direction has
// two factors while reverse only has one, so a corrective volume factor needs
// to be divided out to set the actual forward factor.
type React struct {

	// forward rate constant for N / sec assuming 2 forward factors
	Kf float64

	// backward rate constant for N / sec assuming 1 backward factor
	Kb float64
}

// SetVol sets reaction forward / backward time constants in seconds,
// dividing forward Kf by volume to compensate for 2 volume-based concentrations
// occurring in forward component, vs just 1 in back
func (rt *React) SetVol(f, vol, b float64) {
	rt.Kf = CoFmN(f, vol)
	rt.Kb = b
}

// Set sets reaction forward / backward time constants in seconds
func (rt *React) Set(f, b float64) {
	rt.Kf = f
	rt.Kb = b
}

// Step computes delta A, B, AB values based on current A, B, and AB values
func (rt *React) Step(ca, cb, cab float64, da, db, dab *float64) {
	df := rt.Kf*ca*cb - rt.Kb*cab
	*dab += df
	*da -= df
	*db -= df
}

// StepK computes delta A, B, AB values based on current A, B, and AB values
// K version has additional rate multiplier for Kf
func (rt *React) StepK(kf, ca, cb, cab float64, da, db, dab *float64) {
	df := kf*rt.Kf*ca*cb - rt.Kb*cab
	*dab += df
	*da -= df
	*db -= df
}

// StepCB computes delta A, AB values based on current A, B, and AB values
// assumes B does not change -- does not compute db
func (rt *React) StepCB(ca, cb, cab float64, da, dab *float64) {
	df := rt.Kf*ca*cb - rt.Kb*cab
	*dab += df
	*da -= df
}
