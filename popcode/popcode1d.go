// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package popcode

import "github.com/chewxy/math32"

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

// Encode generates a pattern of activation of given size to encode given value.
// n must be 2 or more.
// pat slice will be constructed if len != n
func (pc *OneD) Encode(pat *[]float32, val float32, n int) {
	if len(*pat) != n {
		*pat = make([]float32, n)
	}
	rng := pc.Max - pc.Min
	incr := rng / float32(n-1)
	for i := 0; i < n; i++ {
		trg := pc.Min + incr*float32(i)
		act := float32(0)
		switch pc.Code {
		case GaussBump:
			dist := (trg - val) / (rng * pc.Sigma)
			act = math32.Exp(-(dist * dist))
		case Localist:
			dist := math32.Abs(trg - val)
			if dist > incr {
				act = 0
			} else {
				act = 1.0 - (dist / incr)
			}
		}
		(*pat)[i] = act
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
		trg := pc.Min + incr*float32(i)
		if act < pc.Thr {
			act = 0
		}
		avg += trg * act
		sum += act
	}
	sum = math32.Max(sum, pc.MinSum)
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
