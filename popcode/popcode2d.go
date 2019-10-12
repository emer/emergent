// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package popcode

import (
	"fmt"
	"log"

	"github.com/chewxy/math32"
	"github.com/emer/etable/etensor"
	"github.com/goki/gi/mat32"
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
func (pc *TwoD) Encode(pat etensor.Tensor, val mat32.Vec2) error {
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
	nf := mat32.Vec2{float32(nx), float32(ny)}
	incr := rng.Div(nf)
	for yi := 0; yi < ny; yi++ {
		for xi := 0; xi < nx; xi++ {
			fi := mat32.Vec2{float32(xi), float32(yi)}
			trg := pc.Min.Add(incr.Mul(fi))
			act := float32(0)
			switch pc.Code {
			case GaussBump:
				dist := trg.Sub(val).Mul(gnrm)
				act = math32.Exp(-dist.LengthSq())
			case Localist:
				dist := trg.Sub(val)
				dist.X = math32.Abs(dist.X)
				dist.Y = math32.Abs(dist.Y)
				if dist.X > incr.X || dist.Y > incr.Y {
					act = 0
				} else {
					nd := dist.Div(incr)
					act = 1.0 - 0.5*(nd.X+nd.Y)
				}
			}
			pat.SetFloat([]int{yi, xi}, float64(act))
		}
	}
	return nil
}

// Decode decodes value from a pattern of activation
// as the activation-weighted-average of the unit's preferred
// tuning values.
// must have 2 or more values in pattern pat.
// TODO: TBD
func (pc *TwoD) Decode(pat []float32) float32 {
	// n := len(pat)
	// if n < 2 {
	// 	return 0
	// }
	// rng := pc.Max - pc.Min
	// incr := rng / float32(n-1)
	// avg := float32(0)
	// sum := float32(0)
	// for i, act := range pat {
	// 	trg := pc.Min + incr*float32(i)
	// 	if act < pc.Thr {
	// 		act = 0
	// 	}
	// 	avg += trg * act
	// 	sum += act
	// }
	// sum = math32.Max(sum, pc.MinSum)
	// avg /= sum
	// return avg
	return 0
}

// Values sets the vals 2D tensor to the target preferred tuning values
// for each unit, for a distribution of given size for shape of tensor.
// tensor must have 2D shape with dims 2 or larger (error if not)
// TODO: TBD
// requires 2 separate float vecs it seems
func (pc *TwoD) Values(vals etensor.Tensor) error {
	// if pat.NumDims() != 2 {
	// 	err := fmt.Errorf("popcode.TwoD Encode: pattern must have 2 dimensions")
	// 	log.Println(err)
	// 	return err
	// }
	// rng := pc.Max.Sub(pc.Min)
	// ny := pat.Dim(0)
	// nx := pat.Dim(1)
	// nf := mat32.Vec2{float32(nx), float32(ny)}
	// incr := rng.Div(nf)
	// for yi := 0; yi < ny; yi++ {
	// 	for xi := 0; xi < nx; xi++ {
	// 		fi := mat32.Vec2{float32(xi), float32(yi)}
	// 		trg := pc.Min.Add(incr.Mul(fi))
	// 		(*vals)[i] = trg
	// }
	return nil
}
