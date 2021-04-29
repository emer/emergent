// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package popcode

import (
	"fmt"
	"log"
	"sort"

	"github.com/emer/etable/etensor"
	"github.com/goki/mat32"
)

// popcode.TwoD provides encoding and decoding of population
// codes, used to represent two continuous (scalar) values
// across a 2D population of units / neurons (2 dimensional)
type TwoD struct {
	Code   PopCodes   `desc:"how to encode the value"`
	Min    mat32.Vec2 `desc:"minimum value representable on each dim -- for GaussBump, typically include extra to allow mean with activity on either side to represent the lowest value you want to encode"`
	Max    mat32.Vec2 `desc:"maximum value representable on each dim -- for GaussBump, typically include extra to allow mean with activity on either side to represent the lowest value you want to encode"`
	Sigma  mat32.Vec2 `def:"0.2" viewif:"Code=GaussBump" desc:"sigma parameters of a gaussian specifying the tuning width of the coarse-coded units, in normalized 0-1 range"`
	Clip   bool       `desc:"ensure that encoded and decoded value remains within specified range"`
	Thr    float32    `def:"0.1" desc:"threshold to cut off small activation contributions to overall average value (i.e., if unit's activation is below this threshold, it doesn't contribute to weighted average computation)"`
	MinSum float32    `def:"0.2" desc:"minimum total activity of all the units representing a value: when computing weighted average value, this is used as a minimum for the sum that you divide by"`
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

	gnrm := mat32.NewVec2Scalar(1).Div(rng.Mul(pc.Sigma))
	ny := pat.Dim(0)
	nx := pat.Dim(1)
	nf := mat32.Vec2{float32(nx - 1), float32(ny - 1)}
	incr := rng.Div(nf)
	for yi := 0; yi < ny; yi++ {
		for xi := 0; xi < nx; xi++ {
			fi := mat32.Vec2{float32(xi), float32(yi)}
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
	avg := mat32.Vec2{}
	if pat.NumDims() != 2 {
		err := fmt.Errorf("popcode.TwoD Decode: pattern must have 2 dimensions")
		log.Println(err)
		return avg, err
	}
	rng := pc.Max.Sub(pc.Min)
	ny := pat.Dim(0)
	nx := pat.Dim(1)
	nf := mat32.Vec2{float32(nx - 1), float32(ny - 1)}
	incr := rng.Div(nf)
	sum := float32(0)
	for yi := 0; yi < ny; yi++ {
		for xi := 0; xi < nx; xi++ {
			idx := []int{yi, xi}
			act := float32(pat.FloatVal(idx))
			if act < pc.Thr {
				act = 0
			}
			fi := mat32.Vec2{float32(xi), float32(yi)}
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
	nf := mat32.Vec2{float32(nx - 1), float32(ny - 1)}
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
	nf := mat32.Vec2{float32(nx - 1), float32(ny - 1)}
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
				fi := mat32.Vec2{float32(x), float32(y)}
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
