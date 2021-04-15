// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package erand

import (
	"math/rand"

	"github.com/goki/ki/kit"
	"gonum.org/v1/gonum/stat/distuv"
)

// RndParams provides parameterized random number generation according to different distributions
// and variance, mean params
type RndParams struct {
	Dist RndDists `desc:"distribution to generate random numbers from"`
	Mean float64  `desc:"mean of random distribution -- typically added to generated random variants"`
	Var  float64  `desc:"variability parameter for the random numbers (gauss = standard deviation, not variance; uniform = half-range, others as noted in RndDists)"`
	Par  float64  `view:"if Dist=Gamma,Binomial,Beta" desc:"extra parameter for distribution (depends on each one)"`
}

// Gen generates a random variable according to current parameters.
// (0 <= thr < 100) specifies thread or dmem proc number for parallel safe random sequences
// (-1 = taMisc::dmem_proc for auto-safe dmem)
func (rp *RndParams) Gen(thr int) float64 {
	switch rp.Dist {
	case Uniform:
		return UniformMeanRange(rp.Mean, rp.Var, thr)
	case Binomial:
		return rp.Mean + Binom(rp.Par, rp.Var, thr)
	case Poisson:
		return rp.Mean + Poiss(rp.Var, thr)
	case Gamma:
		return rp.Mean + Gam(rp.Var, rp.Par, thr)
	case Gaussian:
		return rp.Mean + Gauss(rp.Var, thr)
	case Beta:
		return rp.Mean + Bet(rp.Var, rp.Par, thr)
	}
	return rp.Mean
}

// Density returns density of random variable according to current params, at given
// x value
func (rp *RndParams) Density(s float64) float64 {
	return 0
}

// RndDists are different random number distributions
type RndDists int

//go:generate stringer -type=RndDists

var KiT_RndDists = kit.Enums.AddEnum(RndDistsN, kit.NotBitFlag, nil)

func (ev RndDists) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *RndDists) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }

// The random number distributions
const (
	// Uniform has a uniform probability distribution over var = range on either side of the mean
	Uniform RndDists = iota

	// Binomial represents number of 1's in n (Par) random (Bernouli) trials of probability p (Var)
	Binomial

	// Poisson represents number of events in interval, with event rate (lambda = Var) plus mean
	Poisson

	// Gamma represents maximum entropy distribution with two parameters: scaling parameter (Var)
	// and shape parameter k (Par) plus mean
	Gamma

	// Gaussian normal with Var = stddev plus mean
	Gaussian

	// Beta with var = a and par = b shape parameters
	Beta

	// Mean is just the constant mean, no randomness
	Mean

	RndDistsN
)

// IntZeroN returns uniform random integer in the range between 0 and n, exclusive of n: [0,n).
func IntZeroN(n int64, thr int) int64 {
	return rand.Int63n(n)
}

// IntMinMax returns uniform random integer in range between min and max, exclusive of max: [min,max).
func IntMinMax(min, max int64, thr int) int64 {
	return min + IntZeroN(max-min, thr)
}

// IntMeanRange returns uniform random integer with given range on either side of the mean:
// [mean - range, mean + range]
func IntMeanRange(mean, rnge int64, thr int) int64 {
	return mean + (IntZeroN(2*rnge+1, thr) - rnge)
}

// Discrete samples from a discrete distribution with probabilities given
// (automatically renormalizes the values).  Returns the index of the element of dist.
func Discrete(dist []float64, thr int) int {
	return 0
}

// ZeroOne returns a uniform random number between zero and one (exclusive of 1)
func ZeroOne(thr int) float64 {
	return rand.Float64()
}

// UniformMinMax returns uniform random number between min and max values inclusive
// (Do not use for generating integers - will not include max!)
func UniformMinMax(min, max float64, thr int) float64 {
	return min + (max-min)*ZeroOne(thr)
}

// UniformMeanRange returns uniform random number with given range on either size of the mean:
// [mean - range, mean + range]
func UniformMeanRange(mean, rnge float64, thr int) float64 {
	return mean + rnge*2.0*(ZeroOne(thr)-0.5)
}

// Binom returns binomial with n trials (par) each of probability p (var)
func Binom(n, p float64, thr int) float64 {
	pd := distuv.Binomial{N: n, P: p}
	return pd.Rand()
}

// Poiss returns poisson variable, as number of events in interval,
// with event rate (lmb = Var) plus mean
func Poiss(lmb float64, thr int) float64 {
	pd := distuv.Poisson{Lambda: lmb}
	return pd.Rand()
}

// Gam represents maximum entropy distribution with two parameters:
// scaling parameter (Var, Beta) and shape parameter k (Par, Alpha)
func Gam(v, k float64, thr int) float64 {
	pd := distuv.Gamma{Alpha: k, Beta: v}
	return pd.Rand()
}

// Gauss returns gaussian (normal) random number with given standard deviation
func Gauss(stdev float64, thr int) float64 {
	return stdev * rand.NormFloat64()
}

// Beta returns beta random number with two shape parameters a > 0 and b > 0
func Bet(a, b float64, thr int) float64 {
	x1 := Gam(a, 1, thr)
	x2 := Gam(b, 1, thr)

	return x1 / (x1 + x2)
}

// BoolProp returns boolean true/false with given probability.
func BoolProb(p float64, thr int) bool {
	return (ZeroOne(thr) < p)
}
