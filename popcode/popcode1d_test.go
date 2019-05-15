// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package popcode

import (
	"testing"

	"github.com/chewxy/math32"
)

// difTol is the numerical difference tolerance for comparing vs. target values
const difTol = float32(1.0e-6)

func CmprFloats(out, cor []float32, msg string, t *testing.T) {
	for i := range out {
		dif := math32.Abs(out[i] - cor[i])
		if dif > difTol { // allow for small numerical diffs
			t.Errorf("%v err: out: %v, cor: %v, dif: %v\n", msg, out[i], cor[i], dif)
		}
	}
}

func TestPopCode(t *testing.T) {
	pc := OneD{}
	pc.Defaults()
	var vals []float32
	pc.Values(&vals, 11)
	// fmt.Printf("vals: %v\n", vals)

	corVals := []float32{-0.5, -0.3, -0.1, 0.1, 0.3, 0.5, 0.7, 0.9, 1.1, 1.3, 1.5}

	CmprFloats(vals, corVals, "vals for 11 units", t)

	var pat []float32
	pc.Encode(&pat, 0.5, 11)
	// fmt.Printf("pat for 0.5: %v\n", pat)

	corPat := []float32{0.0019304542, 0.018315637, 0.10539923, 0.3678795, 0.7788008, 1, 0.77880067, 0.3678795, 0.10539923, 0.01831562, 0.0019304542}

	CmprFloats(pat, corPat, "pattern for 0.5 over 11 units", t)

	val := pc.Decode(pat)
	//fmt.Printf("decode pat for 0.5: %v\n", val)
	if math32.Abs(val-0.5) > difTol {
		t.Errorf("did not decode properly: val: %v != 0.5", val)
	}
}
