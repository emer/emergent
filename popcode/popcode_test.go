// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package popcode

import (
	"testing"

	"github.com/emer/etable/etensor"
	"github.com/goki/mat32"
)

// difTol is the numerical difference tolerance for comparing vs. target values
const difTol = float32(1.0e-6)
const difTolWeak = float32(1.0e-4)
const difTolMulti = float32(1.0e-2)

func CmprFloats(out, cor []float32, msg string, t *testing.T) {
	for i := range out {
		dif := mat32.Abs(out[i] - cor[i])
		if dif > difTol { // allow for small numerical diffs
			t.Errorf("%v err: out: %v, cor: %v, dif: %v\n", msg, out[i], cor[i], dif)
		}
	}
}

func TestPopCode1D(t *testing.T) {
	pc := OneD{}
	pc.Defaults()
	var vals []float32
	pc.Values(&vals, 11)
	// fmt.Printf("vals: %v\n", vals)

	corVals := []float32{-0.5, -0.3, -0.1, 0.1, 0.3, 0.5, 0.7, 0.9, 1.1, 1.3, 1.5}

	CmprFloats(vals, corVals, "vals for 11 units", t)

	var pat []float32
	pc.Encode(&pat, 0.5, 11, Set)
	// fmt.Printf("pat for 0.5: %v\n", pat)

	corPat := []float32{0.0019304542, 0.018315637, 0.10539923, 0.3678795, 0.7788008, 1, 0.77880067, 0.3678795, 0.10539923, 0.01831562, 0.0019304542}

	CmprFloats(pat, corPat, "pattern for 0.5 over 11 units", t)

	val := pc.Decode(pat)
	//fmt.Printf("decode pat for 0.5: %v\n", val)
	if mat32.Abs(val-0.5) > difTol {
		t.Errorf("did not decode properly: val: %v != 0.5", val)
	}
}

func TestPopCode1DMulti(t *testing.T) {
	pc := OneD{}
	pc.Defaults()
	var pat []float32
	// note: usually you'd use a larger pattern size for multiple values
	pc.Encode(&pat, 0.1, 11, Set)
	pc.Encode(&pat, 0.9, 11, Add)
	// fmt.Printf("pat for 0.1, 0.9: %v\n", pat)

	corPat := []float32{0.10540401, 0.36800286, 0.78073126, 1.0183157, 0.8842, 0.73575896, 0.8842002, 1.0183157, 0.78073114, 0.36800268, 0.10540401}

	CmprFloats(pat, corPat, "pattern for 0.25, 0.75 over 11 units", t)

	vals := pc.DecodeNPeaks(pat, 2, 1)
	// fmt.Printf("decode pat for 0.25, 0.75: %v\n", vals)
	for _, val := range vals {
		if val > 0.5 {
			if mat32.Abs(val-0.9) > difTolMulti {
				t.Errorf("did not decode properly: val: %v != 0.9", val)
			}
		} else {
			if mat32.Abs(val-0.1) > difTolMulti {
				t.Errorf("did not decode properly: val: %v != 0.1", val)
			}
		}
	}
}

func TestPopCode2D(t *testing.T) {
	pc := TwoD{}
	pc.Defaults()
	var valsX, valsY []float32
	pc.Values(&valsX, &valsY, 11, 11)
	// fmt.Printf("vals: %v\n", valsX)

	corVals := []float32{-0.5, -0.3, -0.1, 0.1, 0.3, 0.5, 0.7, 0.9, 1.1, 1.3, 1.5}

	CmprFloats(valsX, corVals, "valsX for 11 units", t)
	CmprFloats(valsY, corVals, "valsY for 11 units", t)

	var pat etensor.Float32
	pat.SetShape([]int{11, 11}, nil, nil)
	pc.Encode(&pat, mat32.Vec2{0.3, 0.9}, Set)
	// fmt.Printf("pat for 0.5: %v\n", pat)

	corPat := []float32{8.7642576e-08, 5.0434767e-07, 1.7603463e-06, 3.7266532e-06, 4.7851167e-06, 3.7266532e-06, 1.7603463e-06, 5.0434767e-07, 8.7642576e-08, 9.237448e-09, 5.905302e-10, 2.2603292e-06, 1.3007299e-05, 4.5399953e-05, 9.611166e-05, 0.0001234098, 9.611166e-05, 4.5399953e-05, 1.3007299e-05, 2.2603292e-06, 2.3823696e-07, 1.5229979e-08, 3.53575e-05, 0.00020346837, 0.0007101748, 0.0015034394, 0.0019304542, 0.0015034394, 0.0007101748, 0.00020346837, 3.53575e-05, 3.7266532e-06, 2.3823696e-07, 0.00033546257, 0.0019304551, 0.0067379503, 0.014264241, 0.018315647, 0.014264241, 0.006737947, 0.001930456, 0.00033546257, 3.53575e-05, 2.2603292e-06, 0.0019304542, 0.011109002, 0.038774215, 0.08208501, 0.10539925, 0.08208501, 0.038774207, 0.011109007, 0.0019304542, 0.00020346837, 1.3007299e-05, 0.006737947, 0.038774207, 0.1353353, 0.28650483, 0.3678795, 0.28650483, 0.13533528, 0.038774226, 0.006737947, 0.0007101748, 4.5399953e-05, 0.014264233, 0.08208501, 0.28650486, 0.6065308, 0.77880096, 0.6065308, 0.2865048, 0.08208503, 0.014264233, 0.0015034394, 9.611166e-05, 0.018315637, 0.10539923, 0.36787945, 0.7788008, 1, 0.7788008, 0.3678794, 0.10539925, 0.018315637, 0.0019304542, 0.0001234098, 0.014264233, 0.08208499, 0.28650478, 0.6065306, 0.77880067, 0.6065306, 0.2865047, 0.08208499, 0.014264233, 0.0015034394, 9.611166e-05, 0.0067379437, 0.03877419, 0.13533522, 0.28650466, 0.36787927, 0.28650466, 0.13533519, 0.038774196, 0.0067379437, 0.0007101744, 4.5399953e-05, 0.0019304542, 0.011109002, 0.038774207, 0.08208499, 0.10539923, 0.08208499, 0.038774196, 0.011109002, 0.0019304542, 0.00020346837, 1.3007299e-05}

	CmprFloats(pat.Values, corPat, "pattern for 0.3, 0.9 over 11x11 units", t)

	val, err := pc.Decode(&pat)
	// fmt.Printf("decode pat for 0.5: %v\n", val)
	if err != nil {
		t.Error(err)
	}
	if mat32.Abs(val.X-0.3) > difTol {
		t.Errorf("did not decode properly: val: %v != 0.3", val)
	}
	if mat32.Abs(val.Y-0.9) > difTol {
		t.Errorf("did not decode properly: val: %v != 0.9", val)
	}
}

func TestPopCode2DMulti(t *testing.T) {
	pc := TwoD{}
	pc.Defaults()

	var pat etensor.Float32
	// note: usually you'd use a larger pattern size for multiple values
	pat.SetShape([]int{11, 11}, nil, nil)
	pc.Encode(&pat, mat32.Vec2{0.1, 0.9}, Set)
	pc.Encode(&pat, mat32.Vec2{0.9, 0.1}, Add)

	// fmt.Printf("pat for 0.1, 0.9: %v\n", pat)

	corPat := []float32{1.0086953e-06, 1.4767645e-05, 0.00020719503, 0.0019352402, 0.011112728, 0.03877597, 0.08208552, 0.10539932, 0.082085, 0.03877419, 0.011109002, 1.4767645e-05, 9.0799906e-05, 0.00080628647, 0.0068613603, 0.038870327, 0.1353807, 0.28651786, 0.36788172, 0.286505, 0.13533524, 0.038774207, 0.00020719503, 0.00080628647, 0.0030068788, 0.016194696, 0.08358845, 0.28721502, 0.6067343, 0.77883613, 0.60653436, 0.2865049, 0.082085, 0.0019352402, 0.0068613603, 0.016194696, 0.036631294, 0.11966349, 0.37461746, 0.78073144, 1.0003355, 0.778836, 0.36788154, 0.10539932, 0.011112728, 0.038870327, 0.08358845, 0.11966349, 0.16417003, 0.32527903, 0.6176397, 0.7807312, 0.60673404, 0.28651765, 0.0820855, 0.03877597, 0.1353807, 0.28721502, 0.37461746, 0.32527903, 0.2706706, 0.32527906, 0.3746174, 0.28721496, 0.13538063, 0.03877597, 0.08208552, 0.28651786, 0.6067343, 0.78073144, 0.6176397, 0.32527906, 0.16417003, 0.11966347, 0.08358843, 0.0388703, 0.011112728, 0.10539932, 0.36788172, 0.77883613, 1.0003355, 0.7807312, 0.3746174, 0.11966347, 0.036631294, 0.016194696, 0.006861357, 0.0019352402, 0.082085, 0.286505, 0.60653436, 0.778836, 0.60673404, 0.28721496, 0.08358843, 0.016194696, 0.0030068788, 0.00080628606, 0.00020719503, 0.03877419, 0.13533524, 0.2865049, 0.36788154, 0.28651765, 0.13538063, 0.0388703, 0.006861357, 0.00080628606, 9.0799906e-05, 1.4767645e-05, 0.011109002, 0.038774207, 0.082085, 0.10539932, 0.0820855, 0.03877597, 0.011112728, 0.0019352402, 0.00020719503, 1.4767645e-05, 1.0086953e-06}

	CmprFloats(pat.Values, corPat, "pattern for 0.1, 0.9; 0.9; 0.1 over 11x11 units", t)

	vals, err := pc.DecodeNPeaks(&pat, 2, 1)
	if err != nil {
		t.Error(err)
	}

	// fmt.Printf("decode pat for 0.1, 0.9; 0.9, 0.1: %v\n", vals)
	for _, valv := range vals {
		for d := 0; d < 2; d++ {
			val := valv.Dim(mat32.Dims(d))
			if val > 0.5 {
				if mat32.Abs(val-0.9) > difTolMulti {
					t.Errorf("did not decode properly: val: %v != 0.9", val)
				}
			} else {
				if mat32.Abs(val-0.1) > difTolMulti {
					t.Errorf("did not decode properly: val: %v != 0.1", val)
				}
			}
		}
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
	if mat32.Abs(val-180) > difTolWeak {
		t.Errorf("did not decode properly: val: %v != 180", val)
	}

	///////// 330

	pc.Encode(&pat, 330, 24)
	// fmt.Printf("pat for 330: %v\n", pat)

	corPat = []float32{0.73444366, 0.49935186, 0.2909605, 0.1452917, 0.06217656, 0.022802997, 0.0071669817, 0.0019304849, 0.0004458889, 9.0326124e-05, 2.9890747e-05, 9.0326124e-05, 0.0004458889, 0.0019304849, 0.0071669817, 0.022802997, 0.06217656, 0.1452917, 0.2909605, 0.49935186, 0.73444366, 0.92574126, 1, 0.92574126}

	val = pc.Decode(pat)
	// fmt.Printf("decode pat for 330: %v\n", val)
	if mat32.Abs(val-330) > difTolWeak {
		t.Errorf("did not decode properly: val: %v != 330", val)
	}

	///////// 30

	pc.Encode(&pat, 30, 24)
	// fmt.Printf("pat for 30: %v\n", pat)

	corPat = []float32{0.73444366, 0.92574126, 1, 0.92574126, 0.73444366, 0.49935186, 0.2909605, 0.1452917, 0.06217656, 0.022802997, 0.0071669817, 0.0019304849, 0.0004458889, 9.0326124e-05, 2.9890747e-05, 9.0326124e-05, 0.0004458889, 0.0019304849, 0.0071669817, 0.022802997, 0.06217656, 0.1452917, 0.2909605, 0.49935186}

	val = pc.Decode(pat)
	// fmt.Printf("decode pat for 30: %v\n", val)
	if mat32.Abs(val-30) > difTolWeak {
		t.Errorf("did not decode properly: val: %v != 30", val)
	}
}
