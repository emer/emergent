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
const difTolWeak = float32(1.0e-4)

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

func TestRing(t *testing.T) {
	pc := Ring{}
	pc.Defaults()
	pc.Min = 0
	pc.Max = 360
	pc.Sigma = .15 // a bit tighter
	var vals []float32
	pc.Values(&vals, 24)
	// fmt.Printf("vals: %v\n", vals)

	corVals := []float32{0, 15, 30, 45, 60, 75, 90, 105, 120, 135, 150, 165, 180, 195, 210, 225, 240, 255, 270, 285, 300, 315, 330, 345}

	CmprFloats(vals, corVals, "vals for 24 units", t)

	var pat []float32
	pc.Encode(&pat, 180, 24)
	// fmt.Printf("pat for 180: %v\n", pat)

	corPat := []float32{1.4945374e-05, 8.815469e-05, 0.00044561853, 0.001930456, 0.007166979, 0.022802997, 0.06217656, 0.1452917, 0.2909605, 0.49935186, 0.73444366, 0.92574126, 1, 0.92574126, 0.73444366, 0.49935186, 0.2909605, 0.1452917, 0.06217656, 0.022802997, 0.0071669817, 0.0019304849, 0.0004458889, 9.0326124e-05}

	CmprFloats(pat, corPat, "pattern for 180 over 24 units", t)

	val := pc.Decode(pat)
	// fmt.Printf("decode pat for 180: %v\n", val)
	if math32.Abs(val-180) > difTolWeak {
		t.Errorf("did not decode properly: val: %v != 180", val)
	}

	///////// 330

	pc.Encode(&pat, 330, 24)
	// fmt.Printf("pat for 330: %v\n", pat)

	corPat = []float32{0.73444366, 0.49935186, 0.2909605, 0.1452917, 0.06217656, 0.022802997, 0.0071669817, 0.0019304849, 0.0004458889, 9.0326124e-05, 2.9890747e-05, 9.0326124e-05, 0.0004458889, 0.0019304849, 0.0071669817, 0.022802997, 0.06217656, 0.1452917, 0.2909605, 0.49935186, 0.73444366, 0.92574126, 1, 0.92574126}

	val = pc.Decode(pat)
	// fmt.Printf("decode pat for 330: %v\n", val)
	if math32.Abs(val-330) > difTolWeak {
		t.Errorf("did not decode properly: val: %v != 330", val)
	}

	///////// 30

	pc.Encode(&pat, 30, 24)
	// fmt.Printf("pat for 30: %v\n", pat)

	corPat = []float32{0.73444366, 0.92574126, 1, 0.92574126, 0.73444366, 0.49935186, 0.2909605, 0.1452917, 0.06217656, 0.022802997, 0.0071669817, 0.0019304849, 0.0004458889, 9.0326124e-05, 2.9890747e-05, 9.0326124e-05, 0.0004458889, 0.0019304849, 0.0071669817, 0.022802997, 0.06217656, 0.1452917, 0.2909605, 0.49935186}

	val = pc.Decode(pat)
	// fmt.Printf("decode pat for 30: %v\n", val)
	if math32.Abs(val-30) > difTolWeak {
		t.Errorf("did not decode properly: val: %v != 30", val)
	}
}
