// Copyright (c) 2024, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

import (
	"cogentcore.org/core/math32"
)

// TweakTypes are the types of param tweak logic supported.
type TweakTypes int32 //enums:enum

const (
	// Increment increments around current value, e.g., if .5, generates .4 and .6
	Increment TweakTypes = iota

	// Log uses the quasi-log scheme: 1, 2, 5, 10 etc, which only applies if value
	// is one of those numbers.
	Log
)

func tweakValue(msd, fact, exp10 float32, isRmdr bool) float32 {
	if isRmdr {
		return math32.Truncate(msd+fact*math32.Pow(10, exp10), 3)
	}
	return math32.Truncate(fact*math32.Pow(10, exp10), 3)
}

// Tweak returns parameter [Search] values to try,
// below and above the given value.
// Log: use the quasi-log scheme: 1, 2, 5, 10 etc. Only if val is one of these vals.
// Increment: use increments around current value: e.g., if .5, returns .4 and .6.
// These apply to the 2nd significant digit (remainder after most significant digit)
// if that is present in the original value.
func Tweak(v float32, typ TweakTypes) []float32 {
	ex := math32.Floor(math32.Log10(v))
	base := math32.Pow(10, ex)
	basem1 := math32.Pow(10, ex-1)
	fact := math32.Round(v / base)
	msd := tweakValue(0, fact, ex, false)
	rmdr := math32.Round((v - msd) / basem1)
	var vals []float32
	sv := fact
	isRmdr := false
	if rmdr != 0 {
		if rmdr < 0 {
			msd = tweakValue(0, fact-1, ex, false)
			rmdr = math32.Round((v - msd) / basem1)
		}
		sv = rmdr
		ex--
		isRmdr = true
	}
	switch sv {
	case 1:
		if typ == Log {
			vals = append(vals, tweakValue(msd, 5, ex-1, isRmdr), tweakValue(msd, 2, ex, isRmdr))
		} else {
			vals = append(vals, tweakValue(msd, 9, ex-1, isRmdr), tweakValue(msd, 1.1, ex, isRmdr))
		}
	case 2:
		if typ == Log {
			vals = append(vals, tweakValue(msd, 1, ex, isRmdr), tweakValue(msd, 5, ex, isRmdr))
		} else {
			vals = append(vals, tweakValue(msd, 1, ex, isRmdr), tweakValue(msd, 3, ex, isRmdr))
		}
	case 5:
		if typ == Log {
			vals = append(vals, tweakValue(msd, 2, ex, isRmdr), tweakValue(msd, 1, ex+1, isRmdr))
		} else {
			vals = append(vals, tweakValue(msd, 4, ex, isRmdr), tweakValue(msd, 6, ex, isRmdr))
		}
	case 9:
		vals = append(vals, tweakValue(msd, 8, ex, isRmdr), tweakValue(msd, 1, ex+1, isRmdr))
	default:
		vals = append(vals, tweakValue(msd, sv-1, ex, isRmdr), tweakValue(msd, sv+1, ex, isRmdr))
	}
	return vals
}

// TweakPct returns parameter [Search] values to try, as given given percent
// below and above the given value.
func TweakPct(v, pct float32) []float32 {
	trunc := 6
	return []float32{math32.Truncate(v*(1-pct), trunc), math32.Truncate(v*(1+pct), trunc)}
}
