// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import (
	"testing"

	"github.com/chewxy/math32"
)

// difTol is the numerical difference tolerance for comparing vs. target values
const difTol = float32(1.0e-10)

func TestXX1(t *testing.T) {
	xx1 := XX1Params{}
	xx1.Defaults()

	tstx := []float32{-0.05, -0.04, -0.03, -0.02, -0.01, 0, .01, .02, .03, .04, .05, .1, .2, .3, .4, .5}
	cory := []float32{1.7735989e-14, 7.155215e-12, 2.8866178e-09, 1.1645374e-06, 0.00046864923, 0.094767615, 0.47916666, 0.65277773, 0.742268, 0.7967479, 0.8333333, 0.90909094, 0.95238096, 0.96774197, 0.9756098, 0.98039216}
	ny := make([]float32, len(tstx))

	for i := range tstx {
		ny[i] = xx1.NoisyXX1(tstx[i])
		dif := math32.Abs(ny[i] - cory[i])
		if dif > difTol { // allow for small numerical diffs
			t.Errorf("XX1 err: dix: %v, x: %v, y: %v, cor y: %v, dif: %v\n", i, tstx[i], ny[i], cory[i], dif)
		}
	}
	// fmt.Printf("ny vals: %v\n", ny)
}

func TestActUpdt(t *testing.T) {
	// note: these values have been validated against emergent v8.5.6 svn 11473 in
	// demo/leabra/basic_leabra_test.proj, TestAct program
	geinc := []float32{.01, .02, .03, .04, .05, .1, .2, .3}
	corge := []float32{0.007142857, 0.023469387, 0.049562685, 0.085589334, 0.13159695, 0.21617055, 0.3831916, 0.64519763}
	ge := make([]float32, len(geinc))
	corinet := []float32{-0.015714284, -0.0048542274, 0.011293108, 0.032156322, 0.056659013, 0.09967137, 0.1782439, 0.275567}
	inet := make([]float32, len(geinc))
	corvm := []float32{0.3952381, 0.39376712, 0.39718926, 0.4069336, 0.424103, 0.45430642, 0.50831974, 0.5918249}
	vm := make([]float32, len(geinc))
	coract := []float32{2.8884673e-29, 3.2081596e-29, 1.1549086e-28, 3.2309342e-26, 9.598328e-22, 7.120265e-14, 0.29335475, 0.5022214}
	act := make([]float32, len(geinc))

	ac := ActParams{}
	ac.Defaults()

	nrn := &Neuron{}
	ac.InitActs(nrn)

	for i := range geinc {
		nrn.GeInc = geinc[i]
		ac.GeGiFmInc(nrn)
		ac.VmFmG(nrn)
		ac.ActFmG(nrn)
		ge[i] = nrn.Ge
		inet[i] = nrn.Inet
		vm[i] = nrn.Vm
		act[i] = nrn.Act
		difge := math32.Abs(ge[i] - corge[i])
		if difge > difTol { // allow for small numerical diffs
			t.Errorf("ge err: idx: %v, geinc: %v, ge: %v, corge: %v, dif: %v\n", i, geinc[i], ge[i], corge[i], difge)
		}
		difinet := math32.Abs(inet[i] - corinet[i])
		if difinet > difTol { // allow for small numerical diffs
			t.Errorf("Inet err: idx: %v, geinc: %v, inet: %v, corinet: %v, dif: %v\n", i, geinc[i], inet[i], corinet[i], difinet)
		}
		difvm := math32.Abs(vm[i] - corvm[i])
		if difvm > difTol { // allow for small numerical diffs
			t.Errorf("Vm err: idx: %v, geinc: %v, vm: %v, corvm: %v, dif: %v\n", i, geinc[i], vm[i], corvm[i], difvm)
		}
		difact := math32.Abs(act[i] - coract[i])
		if difact > difTol { // allow for small numerical diffs
			t.Errorf("Act err: idx: %v, geinc: %v, act: %v, coract: %v, dif: %v\n", i, geinc[i], act[i], coract[i], difact)
		}
	}
	// fmt.Printf("ge vals: %v\n", ge)
	// fmt.Printf("Inet vals: %v\n", inet)
	// fmt.Printf("vm vals: %v\n", vm)
	// fmt.Printf("act vals: %v\n", act)
}
