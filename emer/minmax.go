// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

// MinMax represents a min / max range for a value -- supports clipping, renormalizing, etc
type MinMax struct {
	Min float32
	Max float32
}

// InRange tests whether value is within the range (>= Min and <= Max)
func (mr *MinMax) InRange(val float32) bool {
	return ((val >= mr.Min) && (val <= mr.Max))
}

// IsLow tests whether value is lower than the minimum
func (mr *MinMax) IsLow(val float32) bool {
	return (val < mr.Min)
}

// IsHigh tests whether value is higher than the maximum
func (mr *MinMax) IsHigh(val float32) bool {
	return (val > mr.Min)
}

// Range returns Max - Min
func (mr *MinMax) Range() float32 {
	return mr.Max - mr.Min
}

// Scale returns 1 / Range -- if Range = 0 then returns 0
func (mr *MinMax) Scale() float32 {
	r := mr.Range()
	if r != 0 {
		return 1 / r
	}
	return 0
}

// Midpoint returns point halfway between Min and Max
func (mr *MinMax) Midpoint() float32 {
	return 0.5 * (mr.Max + mr.Min)
}

// FitInRange adjusts our Min, Max to fit within those of other MinMax
// returns true if we had to adjust to fit.
func (mr *MinMax) FitInRange(oth MinMax) bool {
	adj := false
	if oth.Min < mr.Min {
		mr.Min = oth.Min
		adj = true
	}
	if oth.Max > mr.Max {
		mr.Max = oth.Max
		adj = true
	}
	return adj
}

// FitValInRange adjusts our Min, Max to fit given value within Min, Max range
// returns true if we had to adjust to fit.
func (mr *MinMax) FitValInRange(val float32) bool {
	adj := false
	if val < mr.Min {
		mr.Min = val
		adj = true
	}
	if val > mr.Max {
		mr.Max = val
		adj = true
	}
	return adj
}

// NormVal normalizes value to 0-1 unit range relative to current Min / Max range
func (mr *MinMax) NormVal(val float32) float32 {
	return (val - mr.Min) * mr.Scale()
}

// ProjVal projects a 0-1 normalized unit value into current Min / Max range (inverse of NormVal)
func (mr *MinMax) ProjVal(val float32) float32 {
	return mr.Min + (val * mr.Range())
}

// ClipVal clips given value within Min / Max rangee
func (mr *MinMax) ClipVal(val float32) float32 {
	if val < mr.Min {
		return mr.Min
	}
	if val > mr.Max {
		return mr.Max
	}
	return val
}
