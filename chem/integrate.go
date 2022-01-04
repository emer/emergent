// Copyright (c) 2021 The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chem

// IntegrationDt is the time step of integration
// for Urakubo et al, 2008: uses 5e-5, 2e-4 is barely stable, 5e-4 is not
// The AC1act dynamics in particular are not stable due to large ATP, AMP numbers
const IntegrationDt = 5e-5

// Integrate adds delta to current value with integration rate constant IntegrationDt
// new value cannot go below 0
func Integrate(c *float64, d float64) {
	*c += IntegrationDt * d
	if *c < 0 {
		*c = 0
	}
}

// note: genesis kkit uses exponential Euler which requires separate A - B deltas
// advantages are unclear.
// if *c > 1e-10 && d > 1e-10 { // note: exponential Euler requires separate A - B deltas
// 	dd := math.Exp()
// } else {
