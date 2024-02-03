// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

import (
	"cogentcore.org/core/mat32"
)

func tweakVal(fact, exp10 float32) float32 {
	return mat32.Truncate(fact*mat32.Pow(10, exp10), 3)
}

// Tweak returns parameters to try below and above the given value.
// log: use the quasi-log scheme: 1, 2, 5, 10 etc.  Only if val is one of these vals.
// incr: use increments around current value: e.g., if .5, returns .4 and .6
func Tweak(v float32, log, incr bool) []float32 {
	ex := mat32.Floor(mat32.Log10(v))
	base := mat32.Pow(10, ex)
	fact := mat32.Round(v / base)
	var vals []float32
	switch fact {
	case 1:
		if log {
			vals = append(vals, tweakVal(5, ex-1), tweakVal(2, ex))
		}
		if incr {
			vals = append(vals, tweakVal(9, ex-1))
			vals = append(vals, tweakVal(1.1, ex))
		}
	case 2:
		if log {
			vals = append(vals, tweakVal(1, ex), tweakVal(5, ex))
		}
		if incr {
			if !log {
				vals = append(vals, tweakVal(1, ex))
			}
			vals = append(vals, tweakVal(3, ex))
		}
	case 5:
		if log {
			vals = append(vals, tweakVal(2, ex), tweakVal(1, ex+1))
		}
		if incr {
			vals = append(vals, tweakVal(4, ex), tweakVal(6, ex))
		}
	case 9:
		vals = append(vals, tweakVal(8, ex), tweakVal(1, ex+1))
	default:
		vals = append(vals, tweakVal(fact-1, ex), tweakVal(fact+1, ex))
	}
	return vals
}
