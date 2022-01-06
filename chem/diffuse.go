// Copyright (c) 2021 The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chem

// Diffuse models diffusion between two compartments A and B as
// a function of concentration in each and potentially asymmetric
// rate constants: A Kf -> B and B Kb -> A
// computes the difference between each direction and applies to each
type Diffuse struct {
	Kf float64 `desc:"A -> B forward diffusion rate constant, sec-1"`
	Kb float64 `desc:"B -> A backward diffusion rate constant, sec-1"`
}

// Set sets both diffusion rates
func (rt *Diffuse) Set(kf, kb float64) {
	rt.Kf = kf
	rt.Kb = kb
}

// SetSym sets symmetric diffusion rate (Kf == Kb)
func (rt *Diffuse) SetSym(kfb float64) {
	rt.Kf = kfb
	rt.Kb = kfb
}

// Step computes delta A and B values based on current A, B values
// inputs are numbers, converted to concentration to drive rate
func (rt *Diffuse) Step(ca, cb, va, vb float64, da, db *float64) {
	df := rt.Kf*(ca/va) - rt.Kb*(cb/vb)
	*da -= df
	*db += df
}
