// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package popcode

import (
	"sort"

	"github.com/goki/mat32"
)

type PopCodes int

const (
	// GaussBump = gaussian bump, with value = weighted average of tuned unit values
	GaussBump PopCodes = iota

	// Localist = each unit represents a distinct value; intermediate values represented by graded activity of neighbors; overall activity is weighted-average across all units
	Localist
)

// popcode.OneD provides encoding and decoding of population
// codes, used to represent a single continuous (scalar) value
// across a population of units / neurons (1 dimensional)
type OneD struct {
	Code   PopCodes `desc:"how to encode the value"`
	Min    float32  `desc:"minimum value representable -- for GaussBump, typically include extra to allow mean with activity on either side to represent the lowest value you want to encode"`
	Max    float32  `desc:"maximum value representable -- for GaussBump, typically include extra to allow mean with activity on either side to represent the lowest value you want to encode"`
	Sigma  float32  `def:"0.2" viewif:"Code=GaussBump" desc:"sigma parameter of a gaussian specifying the tuning width of the coarse-coded units, in normalized 0-1 range"`
	Clip   bool     `desc:"ensure that encoded and decoded value remains within specified range"`
	Thr    float32  `def:"0.1" desc:"threshold to cut off small activation contributions to overall average value (i.e., if unit's activation is below this threshold, it doesn't contribute to weighted average computation)"`
	MinSum float32  `def:"0.2" desc:"minimum total activity of all the units representing a value: when computing weighted average value, this is used as a minimum for the sum that you divide by"`
}

func (pc *OneD) Defaults() {
	pc.Code = GaussBump
	pc.Min = -0.5
	pc.Max = 1.5
	pc.Sigma = 0.2
	pc.Clip = true
	pc.Thr = 0.1
	pc.MinSum = 0.2
}

// SetRange sets the min, max and sigma values
func (pc *OneD) SetRange(min, max, sigma float32) {
	pc.Min = min
	pc.Max = max
	pc.Sigma = sigma
}

const (
	// Add is used for popcode Encode methods, add arg -- indicates to add values
	// to any existing values in the target vector / tensor:
	// used for encoding additional values (see DecodeN for decoding).
	Add = true

	// Set is used for popcode Encode methods, add arg -- indicates to set values
	// in any existing values in the target vector / tensor:
	// used for encoding first / only values.
	Set = false
)

// Encode generates a pattern of activation of given size to encode given value.
// n must be 2 or more. pat slice will be constructed if len != n.
// If add == false (use Set const for clarity), values are set to pattern
// else if add == true (Add), then values are added to any existing,
// for encoding additional values in same pattern.
func (pc *OneD) Encode(pat *[]float32, val float32, n int, add bool) {
	if len(*pat) != n {
		*pat = make([]float32, n)
	}
	if pc.Clip {
		val = mat32.Clamp(val, pc.Min, pc.Max)
	}
	rng := pc.Max - pc.Min
	gnrm := 1 / (rng * pc.Sigma)
	incr := rng / float32(n-1)
	for i := 0; i < n; i++ {
		trg := pc.Min + incr*float32(i)
		act := float32(0)
		switch pc.Code {
		case GaussBump:
			dist := gnrm * (trg - val)
			act = mat32.Exp(-(dist * dist))
		case Localist:
			dist := mat32.Abs(trg - val)
			if dist > incr {
				act = 0
			} else {
				act = 1.0 - (dist / incr)
			}
		}
		if add {
			(*pat)[i] += act
		} else {
			(*pat)[i] = act
		}
	}
}

// Decode decodes value from a pattern of activation
// as the activation-weighted-average of the unit's preferred
// tuning values.
// must have 2 or more values in pattern pat.
func (pc *OneD) Decode(pat []float32) float32 {
	n := len(pat)
	if n < 2 {
		return 0
	}
	rng := pc.Max - pc.Min
	incr := rng / float32(n-1)
	avg := float32(0)
	sum := float32(0)
	for i, act := range pat {
		if act < pc.Thr {
			act = 0
		}
		trg := pc.Min + incr*float32(i)
		avg += trg * act
		sum += act
	}
	sum = mat32.Max(sum, pc.MinSum)
	avg /= sum
	return avg
}

// Values sets the vals slice to the target preferred tuning values
// for each unit, for a distribution of given size n.
// n must be 2 or more.
// vals slice will be constructed if len != n
func (pc *OneD) Values(vals *[]float32, n int) {
	if len(*vals) != n {
		*vals = make([]float32, n)
	}
	rng := pc.Max - pc.Min
	incr := rng / float32(n-1)
	for i := 0; i < n; i++ {
		trg := pc.Min + incr*float32(i)
		(*vals)[i] = trg
	}
}

// DecodeNPeaks decodes N values from a pattern of activation
// using a neighborhood of specified width around local maxima,
// which is the amount on either side of the central point to
// accumulate (0 = localist, single points, 1 = +/- 1 point on
// either side, etc).
// Allocates a temporary slice of size pat, and sorts that: relatively expensive
func (pc *OneD) DecodeNPeaks(pat []float32, nvals, width int) []float32 {
	n := len(pat)
	if n < 2 {
		return nil
	}
	rng := pc.Max - pc.Min
	incr := rng / float32(n-1)

	type navg struct {
		avg float32
		idx int
	}
	avgs := make([]navg, n)

	for i := range pat {
		sum := float32(0)
		ns := 0
		for d := -width; d <= width; d++ {
			di := i + d
			if di < 0 || di >= n {
				continue
			}
			act := pat[di]
			if act < pc.Thr {
				continue
			}
			sum += pat[di]
			ns++
		}
		avgs[i].avg = sum / float32(ns)
		avgs[i].idx = i
	}

	// sort highest to lowest
	sort.Slice(avgs, func(i, j int) bool {
		return avgs[i].avg > avgs[j].avg
	})

	vals := make([]float32, nvals)
	for i := range vals {
		avg := float32(0)
		sum := float32(0)
		mxi := avgs[i].idx
		for d := -width; d <= width; d++ {
			di := mxi + d
			if di < 0 || di >= n {
				continue
			}
			act := pat[di]
			if act < pc.Thr {
				act = 0
			}
			trg := pc.Min + incr*float32(di)
			avg += trg * act
			sum += act
		}
		sum = mat32.Max(sum, pc.MinSum)
		vals[i] = avg / sum
	}

	return vals
}
