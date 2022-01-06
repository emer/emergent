// Copyright (c) 2021 The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chem

// EnzRate models an enzyme-catalyzed reaction based on the Michaelis-Menten kinetics
// that transforms S = substrate into P product via SE bound C complex
//       K1         K3
// S + E --> C(SE) ---> P + E
//      <-- K2
// S = substrate, E = enzyme, C = SE complex, P = product
// This version does NOT consume the E enzyme or directly use the C complex
// as an accumulated factor: instead it directly computes an overall rate
// for the end-to-end S <-> P reaction based on the K constants:
// rate = S * E * K3 / (S + Km)
// This amount is added to the P and subtracted from the S, and recorded
// in the C complex variable as rate / K3 -- it is just directly set.
// In some situations this C variable can be used for other things.
// The source K constants are in terms of concentrations μM-1 and sec-1
// but calculations take place using N's, and the forward direction has
// two factors while reverse only has one, so a corrective volume factor needs
// to be divided out to set the actual forward factor.
type EnzRate struct {
	K1 float64 `desc:"S+E forward rate constant, in μM-1 msec-1"`
	K2 float64 `desc:"SE backward rate constant, in μM-1 msec-1"`
	K3 float64 `desc:"SE -> P + E catalyzed rate constant, in μM-1 msec-1"`
	Km float64 `inactive:"+" desc:"Michaelis constant = (K2 + K3) / K1 -- goes into the rate"`
}

func (rt *EnzRate) Update() {
	rt.Km = (rt.K2 + rt.K3) / rt.K1
}

// SetKmVol sets time constants in seconds using Km, K2, K3
// dividing forward K1 by volume to compensate for 2 volume-based concentrations
// occurring in forward component (s * e), vs just 1 in back
func (rt *EnzRate) SetKmVol(km, vol, k2, k3 float64) {
	k1 := (k2 + k3) / km
	rt.K1 = CoFmN(k1, vol)
	rt.K2 = k2
	rt.K3 = k3
	rt.Update()
}

// SetKm sets time constants in seconds using Km, K2, K3
func (rt *EnzRate) SetKm(km, k2, k3 float64) {
	k1 := (k2 + k3) / km
	rt.K1 = k1
	rt.K2 = k2
	rt.K3 = k3
	rt.Update()
}

// Set sets time constants in seconds directly
func (rt *EnzRate) Set(k1, k2, k3 float64) {
	rt.K1 = k1
	rt.K2 = k2
	rt.K3 = k3
	rt.Update()
}

// Step computes delta values based on current S, E values, setting dS, dP and C = rate
func (rt *EnzRate) Step(cs, ce float64, ds, dp, cc *float64) {
	rate := (rt.K3 * cs * ce) / (cs + rt.Km)
	*dp += rate
	*ds -= rate
	*cc += rate // directly stored
}

// Step computes delta values based on current S, E values, setting dS, dP and C = rate
// K version has additional rate multiplier for Kf = K1
func (rt *EnzRate) StepK(kf, cs, ce float64, ds, dp, cc *float64) {
	km := (rt.K2 + rt.K3) / (rt.K1 * kf)
	rate := (rt.K3 * cs * ce) / (cs + km)
	*dp += rate
	*ds -= rate
	*cc += rate // directly stored
}
