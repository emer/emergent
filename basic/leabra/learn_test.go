// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import (
	"testing"

	"github.com/chewxy/math32"
)

func TestXCal(t *testing.T) {
	xcal := XCalParams{}
	xcal.Defaults()

	// note: these values have been validated against emergent v8.5.6 svn 11492 in
	// demo/leabra/basic_leabra_test.proj, TestLearn program

	tstSr := []float32{.01, .02, .03, .04, .05, .1, .2, .3, .4, .5, .6, .7, .8}
	tstThrp := []float32{.1, .1, .1, .1, .1, .1, .2, .2, .2, .2, .2, .3, .3}
	cory := []float32{-0.089999996, -0.08, -0.07, -0.060000002, -0.05, 0, 0, 0.10000001, 0.2, 0.3, 0.40000004, 0.39999998, 0.5}
	ny := make([]float32, len(tstSr))

	for i := range tstSr {
		ny[i] = xcal.DWt(tstSr[i], tstThrp[i])
		dif := math32.Abs(ny[i] - cory[i])
		if dif > difTol { // allow for small numerical diffs
			t.Errorf("XCal err: i: %v, Sr: %v, thrP: %v, got: %v, cor y: %v, dif: %v\n", i, tstSr[i], tstThrp[i], ny[i], cory[i], dif)
		}
	}
	// fmt.Printf("ny vals: %v\n", ny)
}
