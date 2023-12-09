// Copyright (c) 2021 The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chem

// Enz models an enzyme-catalyzed reaction based on the Michaelis-Menten kinetics
// that transforms S = substrate into P product via SE-bound C complex
//
//	K1         K3
//
// S + E --> C(SE) ---> P + E
//
//	<-- K2
//
// S = substrate, E = enzyme, C = SE complex, P = product
// The source K constants are in terms of concentrations μM-1 and sec-1
// but calculations take place using N's, and the forward direction has
// two factors while reverse only has one, so a corrective volume factor needs
// to be divided out to set the actual forward factor.
type Enz struct {

	// S+E forward rate constant, in μM-1 msec-1
	K1 float64

	// SE backward rate constant, in μM-1 msec-1
	K2 float64

	// SE -> P + E catalyzed rate constant, in μM-1 msec-1
	K3 float64

	// Michaelis constant = (K2 + K3) / K1
	Km float64 `inactive:"+"`
}

func (rt *Enz) Update() {
	rt.Km = (rt.K2 + rt.K3) / rt.K1
}

// SetKmVol sets time constants in seconds using Km, K2, K3
// dividing forward K1 by volume to compensate for 2 volume-based concentrations
// occurring in forward component (s * e), vs just 1 in back
func (rt *Enz) SetKmVol(km, vol, k2, k3 float64) {
	k1 := (k2 + k3) / km
	rt.K1 = CoFmN(k1, vol)
	rt.K2 = k2
	rt.K3 = k3
	rt.Update()
}

// SetKm sets time constants in seconds using Km, K2, K3
func (rt *Enz) SetKm(km, k2, k3 float64) {
	k1 := (k2 + k3) / km
	rt.K1 = k1
	rt.K2 = k2
	rt.K3 = k3
	rt.Update()
}

// Set sets time constants in seconds directly
func (rt *Enz) Set(k1, k2, k3 float64) {
	rt.K1 = k1
	rt.K2 = k2
	rt.K3 = k3
	rt.Update()
}

// Step computes delta values based on current S, E, C, and P values
func (rt *Enz) Step(cs, ce, cc, cp float64, ds, de, dc, dp *float64) {
	df := rt.K1 * cs * ce // forward
	db := rt.K2 * cc      // backward
	do := rt.K3 * cc      // out to product
	*dp += do
	*dc += df - (do + db) // complex = forward - back - output
	*de += (do + db) - df // e is released with product and backward from complex, consumed by forward
	*ds -= (df - db)      // substrate = back - forward
}

// StepK computes delta values based on current S, E, C, and P values
// K version has additional rate multiplier for Kf = K1
func (rt *Enz) StepK(kf, cs, ce, cc, cp float64, ds, de, dc, dp *float64) {
	df := kf * rt.K1 * cs * ce // forward
	db := rt.K2 * cc           // backward
	do := rt.K3 * cc           // out to product
	*dp += do
	*dc += df - (do + db) // complex = forward - back - output
	*de += (do + db) - df // e is released with product and backward from complex, consumed by forward
	*ds -= (df - db)      // substrate = back - forward
}
