// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package popcode

import (
	"github.com/goki/mat32"
)

// Ring is a OneD popcode that encodes a circular value such as an angle
// that wraps around at the ends.  It uses two internal vectors
// to render the wrapped-around values into, and then adds them into
// the final result.  Unlike regular PopCodes, the Min and Max should
// represent the exact range of the value (e.g., 0 to 360 for angle)
// with no extra on the ends, as that extra will wrap around to
// the other side in this case.
type Ring struct {
	OneD
	LowVec  []float32 `view:"-" desc:"low-end encoding vector"`
	HighVec []float32 `view:"-" desc:"high-end encoding vector"`
}

// AllocVecs allocates internal LowVec, HighVec storage,
// allowing for variable lengths to be encoded using same object,
// growing capacity to max, but using exact amount each time
func (pc *Ring) AllocVecs(n int) {
	if cap(pc.LowVec) < n {
		pc.LowVec = make([]float32, n)
		pc.HighVec = make([]float32, n)
	}
	pc.LowVec = pc.LowVec[:n]
	pc.HighVec = pc.HighVec[:n]
}

// Encode generates a pattern of activation of given size to encode given value.
// n must be 2 or more.
// pat slice will be constructed if len != n
func (pc *Ring) Encode(pat *[]float32, val float32, n int) {
	pc.Clip = false // doesn't work with clip!
	if len(*pat) != n {
		*pat = make([]float32, n)
	}
	pc.AllocVecs(n)
	half := (pc.Max + pc.Min) / 2
	if val > half {
		pc.EncodeImpl(&pc.LowVec, pc.Min+(val-pc.Max), n) // 0 + (340 - 360) = -20
		pc.EncodeImpl(&pc.HighVec, val, n)
	} else {
		pc.EncodeImpl(&pc.LowVec, val, n)                  // 0 + (340 - 360) = -20
		pc.EncodeImpl(&pc.HighVec, pc.Max+(val-pc.Min), n) // 360 + (20-0) = 380
	}
	for i := 0; i < n; i++ {
		(*pat)[i] = pc.LowVec[i] + pc.HighVec[i]
	}
}

// EncodeImpl generates a pattern of activation of given size to encode given value.
// n must be 2 or more.
// pat slice will be constructed if len != n
func (pc *Ring) EncodeImpl(pat *[]float32, val float32, n int) {
	if len(*pat) != n {
		*pat = make([]float32, n)
	}
	if pc.Clip {
		val = mat32.Clamp(val, pc.Min, pc.Max)
	}
	rng := pc.Max - pc.Min
	gnrm := 1 / (rng * pc.Sigma)
	incr := rng / float32(n) // n instead of n-1
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
		(*pat)[i] = act
	}
}

// Decode decodes value from a pattern of activation
// as the activation-weighted-average of the unit's preferred
// tuning values.
// must have 2 or more values in pattern pat.
func (pc *Ring) Decode(pat []float32) float32 {
	n := len(pat)
	sn := int(pc.Sigma * float32(n)) // amount on each end to blank
	hsn := (n - 1) - sn

	// and record activity in each end region
	lsum := float32(0)
	hsum := float32(0)
	for i := 0; i < n; i++ {
		v := pat[i]
		if i < sn {
			lsum += v
		} else if i >= hsn {
			hsum += v
		}
	}
	hn := n / 2
	half := (pc.Max + pc.Min) / 2
	rng := pc.Max - pc.Min
	incr := rng / float32(n) // n instead of n-1
	avg := float32(0)
	sum := float32(0)
	thr := float32(sn) * .1       // threshold activity to count as having something in tail
	if lsum < thr && hsum < thr { // neither has significant activity, use straight decode
		for i := 0; i < n; i++ {
			act := pat[i]
			trg := pc.Min + incr*float32(i)
			if act < pc.Thr {
				act = 0
			}
			avg += trg * act
			sum += act
		}
	} else if lsum > hsum { // lower is more active -- wrap high end below low end
		for i := 0; i < hn; i++ { // decode lower half as usual
			act := pat[i]
			trg := pc.Min + incr*float32(i)
			if act < pc.Thr {
				act = 0
			}
			avg += trg * act
			sum += act
		}
		min := pc.Min - half
		for i := hn; i < n; i++ { // decode upper half as starting below lower
			act := pat[i]
			trg := min + incr*float32(i-hn)
			if act < pc.Thr {
				act = 0
			}
			avg += trg * act
			sum += act
		}
	} else {
		for i := hn; i < n; i++ { // decode upper half as usual
			act := pat[i]
			trg := pc.Min + incr*float32(i)
			if act < pc.Thr {
				act = 0
			}
			avg += trg * act
			sum += act
		}
		min := pc.Max
		for i := 0; i < hn; i++ { // decode lower half as starting above upper
			act := pat[i]
			trg := min + incr*float32(i)
			if act < pc.Thr {
				act = 0
			}
			avg += trg * act
			sum += act
		}
	}
	sum = mat32.Max(sum, pc.MinSum)
	avg /= sum
	return avg
}

// Values sets the vals slice to the target preferred tuning values
// for each unit, for a distribution of given size n.
// n must be 2 or more.
// vals slice will be constructed if len != n
func (pc *Ring) Values(vals *[]float32, n int) {
	if len(*vals) != n {
		*vals = make([]float32, n)
	}
	rng := pc.Max - pc.Min
	incr := rng / float32(n) // n instead of n-1
	for i := 0; i < n; i++ {
		trg := pc.Min + incr*float32(i)
		(*vals)[i] = trg
	}
}
