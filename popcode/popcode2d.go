// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package popcode

import (
	"fmt"
	"log"
	"sort"

	"cogentcore.org/core/mat32"
	"github.com/emer/etable/v2/etensor"
)

// popcode.TwoD provides encoding and decoding of population
// codes, used to represent two continuous (scalar) values
// across a 2D tensor, using row-major XY encoding:
// Y = outer, first dim, X = inner, second dim
type TwoD struct {

	// how to encode the value
	Code PopCodes

	// minimum value representable on each dim -- for GaussBump, typically include extra to allow mean with activity on either side to represent the lowest value you want to encode
	Min mat32.Vec2

	// maximum value representable on each dim -- for GaussBump, typically include extra to allow mean with activity on either side to represent the lowest value you want to encode
	Max mat32.Vec2

	// sigma parameters of a gaussian specifying the tuning width of the coarse-coded units, in normalized 0-1 range
	Sigma mat32.Vec2 `default:"0.2"`

	// ensure that encoded and decoded value remains within specified range -- generally not useful with wrap
	Clip bool

	// x axis wraps around (e.g., for periodic values such as angle) -- encodes and decodes relative to both the min and max values
	WrapX bool

	// y axis wraps around (e.g., for periodic values such as angle) -- encodes and decodes relative to both the min and max values
	WrapY bool

	// threshold to cut off small activation contributions to overall average value (i.e., if unit's activation is below this threshold, it doesn't contribute to weighted average computation)
	Thr float32 `default:"0.1"`

	// minimum total activity of all the units representing a value: when computing weighted average value, this is used as a minimum for the sum that you divide by
	MinSum float32 `default:"0.2"`
}

func (pc *TwoD) Defaults() {
	pc.Code = GaussBump
	pc.Min.Set(-0.5, -0.5)
	pc.Max.Set(1.5, 1.5)
	pc.Sigma.Set(0.2, 0.2)
	pc.Clip = true
	pc.Thr = 0.1
	pc.MinSum = 0.2
}

func (pc *TwoD) ShouldShow(field string) bool {
	switch field {
	case "Sigma":
		return pc.Code == GaussBump
	default:
		return true
	}
}

// SetRange sets the min, max and sigma values to the same scalar values
func (pc *TwoD) SetRange(min, max, sigma float32) {
	pc.Min.Set(min, min)
	pc.Max.Set(max, max)
	pc.Sigma.Set(sigma, sigma)
}

// Encode generates a pattern of activation on given tensor, which must already have
// appropriate 2D shape which is used for encoding sizes (error if not).
// If add == false (use Set const for clarity), values are set to pattern
// else if add == true (Add), then values are added to any existing,
// for encoding additional values in same pattern.
func (pc *TwoD) Encode(pat etensor.Tensor, val mat32.Vec2, add bool) error {
	if pat.NumDims() != 2 {
		err := fmt.Errorf("popcode.TwoD Encode: pattern must have 2 dimensions")
		log.Println(err)
		return err
	}
	if pc.Clip {
		val.Clamp(pc.Min, pc.Max)
	}
	rng := pc.Max.Sub(pc.Min)
	sr := pc.Sigma.Mul(rng)
	if pc.WrapX || pc.WrapY {
		err := pc.EncodeImpl(pat, val, add) // always render first
		ev := val
		// relative to min
		if pc.WrapX && mat32.Abs(pc.Max.X-val.X) < sr.X { // has significant vals near max
			ev.X = pc.Min.X - (pc.Max.X - val.X) // wrap extra above max around to min
		}
		if pc.WrapY && mat32.Abs(pc.Max.Y-val.Y) < sr.Y {
			ev.Y = pc.Min.Y - (pc.Max.Y - val.Y)
		}
		if ev != val {
			err = pc.EncodeImpl(pat, ev, Add) // always add
		}
		// pev := ev
		ev = val
		if pc.WrapX && mat32.Abs(val.X-pc.Min.X) < sr.X { // has significant vals near min
			ev.X = pc.Max.X - (val.X - pc.Min.X) // wrap extra below min around to max
		}
		if pc.WrapY && mat32.Abs(val.Y-pc.Min.Y) < sr.Y {
			ev.Y = pc.Max.Y - (val.Y - pc.Min.Y)
		}
		if ev != val {
			err = pc.EncodeImpl(pat, ev, Add) // always add
		}
		return err
	}
	return pc.EncodeImpl(pat, val, add)
}

// EncodeImpl is the implementation of encoding -- e.g., used twice for Wrap
func (pc *TwoD) EncodeImpl(pat etensor.Tensor, val mat32.Vec2, add bool) error {
	rng := pc.Max.Sub(pc.Min)

	gnrm := mat32.V2Scalar(1).Div(rng.Mul(pc.Sigma))
	ny := pat.Dim(0)
	nx := pat.Dim(1)
	nf := mat32.V2(float32(nx-1), float32(ny-1))
	incr := rng.Div(nf)
	for yi := 0; yi < ny; yi++ {
		for xi := 0; xi < nx; xi++ {
			fi := mat32.V2(float32(xi), float32(yi))
			trg := pc.Min.Add(incr.Mul(fi))
			act := float32(0)
			switch pc.Code {
			case GaussBump:
				dist := trg.Sub(val).Mul(gnrm)
				act = mat32.Exp(-dist.LengthSq())
			case Localist:
				dist := trg.Sub(val)
				dist.X = mat32.Abs(dist.X)
				dist.Y = mat32.Abs(dist.Y)
				if dist.X > incr.X || dist.Y > incr.Y {
					act = 0
				} else {
					nd := dist.Div(incr)
					act = 1.0 - 0.5*(nd.X+nd.Y)
				}
			}
			idx := []int{yi, xi}
			if add {
				val := float64(act) + pat.FloatVal(idx)
				pat.SetFloat(idx, val)
			} else {
				pat.SetFloat(idx, float64(act))
			}
		}
	}
	return nil
}

// Decode decodes 2D value from a pattern of activation
// as the activation-weighted-average of the unit's preferred
// tuning values.
func (pc *TwoD) Decode(pat etensor.Tensor) (mat32.Vec2, error) {
	if pat.NumDims() != 2 {
		err := fmt.Errorf("popcode.TwoD Decode: pattern must have 2 dimensions")
		log.Println(err)
		return mat32.Vec2{}, err
	}
	if pc.WrapX || pc.WrapY {
		ny := pat.Dim(0)
		nx := pat.Dim(1)
		ys := make([]float32, ny)
		xs := make([]float32, nx)
		ydiv := 1.0 / (float32(nx) * pc.Sigma.X)
		xdiv := 1.0 / (float32(ny) * pc.Sigma.Y)
		for yi := 0; yi < ny; yi++ {
			for xi := 0; xi < nx; xi++ {
				idx := []int{yi, xi}
				act := float32(pat.FloatVal(idx))
				if act < pc.Thr {
					act = 0
				}
				ys[yi] += act * ydiv
				xs[xi] += act * xdiv
			}
		}
		var val mat32.Vec2
		switch {
		case pc.WrapX && pc.WrapY:
			dx := Ring{}
			dx.Defaults()
			dx.Min = pc.Min.X
			dx.Max = pc.Max.X
			dx.Sigma = pc.Sigma.X
			dx.Thr = pc.Thr
			dx.MinSum = pc.MinSum
			dx.Code = pc.Code
			dy := Ring{}
			dy.Defaults()
			dy.Min = pc.Min.Y
			dy.Max = pc.Max.Y
			dy.Sigma = pc.Sigma.Y
			dy.Thr = pc.Thr
			dy.MinSum = pc.MinSum
			dy.Code = pc.Code
			val.X = dx.Decode(xs)
			val.Y = dy.Decode(ys)
		case pc.WrapX:
			dx := Ring{}
			dx.Defaults()
			dx.Min = pc.Min.X
			dx.Max = pc.Max.X
			dx.Sigma = pc.Sigma.X
			dx.Thr = pc.Thr
			dx.MinSum = pc.MinSum
			dx.Code = pc.Code
			dy := OneD{}
			dy.Defaults()
			dy.Min = pc.Min.Y
			dy.Max = pc.Max.Y
			dy.Sigma = pc.Sigma.Y
			dy.Thr = pc.Thr
			dy.MinSum = pc.MinSum
			dy.Code = pc.Code
			val.X = dx.Decode(xs)
			val.Y = dy.Decode(ys)
		case pc.WrapY:
			dx := OneD{}
			dx.Defaults()
			dx.Min = pc.Min.X
			dx.Max = pc.Max.X
			dx.Sigma = pc.Sigma.X
			dx.Thr = pc.Thr
			dx.MinSum = pc.MinSum
			dx.Code = pc.Code
			dy := Ring{}
			dy.Defaults()
			dy.Min = pc.Min.Y
			dy.Max = pc.Max.Y
			dy.Sigma = pc.Sigma.Y
			dy.Thr = pc.Thr
			dy.MinSum = pc.MinSum
			dy.Code = pc.Code
			val.X = dx.Decode(xs)
			val.Y = dy.Decode(ys)
		}
		return val, nil
	} else {
		return pc.DecodeImpl(pat)
	}
}

// DecodeImpl does direct decoding of x, y simultaneously -- for non-wrap
func (pc *TwoD) DecodeImpl(pat etensor.Tensor) (mat32.Vec2, error) {
	avg := mat32.Vec2{}
	rng := pc.Max.Sub(pc.Min)
	ny := pat.Dim(0)
	nx := pat.Dim(1)
	nf := mat32.V2(float32(nx-1), float32(ny-1))
	incr := rng.Div(nf)
	sum := float32(0)
	for yi := 0; yi < ny; yi++ {
		for xi := 0; xi < nx; xi++ {
			idx := []int{yi, xi}
			act := float32(pat.FloatVal(idx))
			if act < pc.Thr {
				act = 0
			}
			fi := mat32.V2(float32(xi), float32(yi))
			trg := pc.Min.Add(incr.Mul(fi))
			avg = avg.Add(trg.MulScalar(act))
			sum += act
		}
	}
	sum = mat32.Max(sum, pc.MinSum)
	return avg.DivScalar(sum), nil
}

// Values sets the vals slices to the target preferred tuning values
// for each unit, for a distribution of given dimensions.
// n's must be 2 or more in each dim.
// vals slice will be constructed if len != n
func (pc *TwoD) Values(valsX, valsY *[]float32, nx, ny int) {
	rng := pc.Max.Sub(pc.Min)
	nf := mat32.V2(float32(nx-1), float32(ny-1))
	incr := rng.Div(nf)

	// X
	if len(*valsX) != nx {
		*valsX = make([]float32, nx)
	}
	for i := 0; i < nx; i++ {
		trg := pc.Min.X + incr.X*float32(i)
		(*valsX)[i] = trg
	}

	// Y
	if len(*valsY) != ny {
		*valsY = make([]float32, ny)
	}
	for i := 0; i < ny; i++ {
		trg := pc.Min.Y + incr.Y*float32(i)
		(*valsY)[i] = trg
	}
}

// DecodeNPeaks decodes N values from a pattern of activation
// using a neighborhood of specified width around local maxima,
// which is the amount on either side of the central point to
// accumulate (0 = localist, single points, 1 = +/- 1 points on
// either side in a square around central point, etc)
// Allocates a temporary slice of size pat, and sorts that: relatively expensive
func (pc *TwoD) DecodeNPeaks(pat etensor.Tensor, nvals, width int) ([]mat32.Vec2, error) {
	if pat.NumDims() != 2 {
		err := fmt.Errorf("popcode.TwoD DecodeNPeaks: pattern must have 2 dimensions")
		log.Println(err)
		return nil, err
	}
	rng := pc.Max.Sub(pc.Min)
	ny := pat.Dim(0)
	nx := pat.Dim(1)
	nf := mat32.V2(float32(nx-1), float32(ny-1))
	incr := rng.Div(nf)

	type navg struct {
		avg  float32
		x, y int
	}
	avgs := make([]navg, nx*ny) // expensive

	idx := 0
	for yi := 0; yi < ny; yi++ {
		for xi := 0; xi < nx; xi++ {
			sum := float32(0)
			ns := 0
			for dy := -width; dy <= width; dy++ {
				y := yi + dy
				if y < 0 || y >= ny {
					continue
				}
				for dx := -width; dx <= width; dx++ {
					x := xi + dx
					if x < 0 || x >= nx {
						continue
					}
					idx := []int{y, x}
					act := float32(pat.FloatVal(idx))
					sum += act
					ns++
				}
			}
			avgs[idx].avg = sum / float32(ns)
			avgs[idx].x = xi
			avgs[idx].y = yi
			idx++
		}
	}

	// sort highest to lowest
	sort.Slice(avgs, func(i, j int) bool {
		return avgs[i].avg > avgs[j].avg
	})

	vals := make([]mat32.Vec2, nvals)
	for i := range vals {
		avg := mat32.Vec2{}
		sum := float32(0)
		mxi := avgs[i].x
		myi := avgs[i].y
		for dy := -width; dy <= width; dy++ {
			y := myi + dy
			if y < 0 || y >= ny {
				continue
			}
			for dx := -width; dx <= width; dx++ {
				x := mxi + dx
				if x < 0 || x >= nx {
					continue
				}
				idx := []int{y, x}
				act := float32(pat.FloatVal(idx))
				if act < pc.Thr {
					act = 0
				}
				fi := mat32.V2(float32(x), float32(y))
				trg := pc.Min.Add(incr.Mul(fi))
				avg = avg.Add(trg.MulScalar(act))
				sum += act
			}
		}
		sum = mat32.Max(sum, pc.MinSum)
		vals[i] = avg.DivScalar(sum)
	}

	return vals, nil
}
