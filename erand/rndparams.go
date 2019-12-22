// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package erand

import (
	"math/rand"

	"github.com/goki/ki/kit"
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
		return rp.Mean + Binom(int(rp.Par), rp.Var, thr)
	case Poisson:
		return rp.Mean + Poiss(rp.Var, thr)
	case Gamma:
		return rp.Mean + Gam(rp.Var, int(rp.Par), thr)
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
func Binom(n int, p float64, thr int) float64 {
	/*
	  var j int
	  static int nold=(-1);
	   double am,em,g,angle,p,bnl,sq,t,y;
	  static double pold=(-1.0),pc,plog,pclog,en,oldg;

	  p=(pp <= 0.5 ? pp : 1.0-pp);
	  am=n*p;
	  if (n < 25) {
	    bnl=0.0;
	    for (j=1;j<=n;j++)
	      if (MTRnd::GenRandRes53(thr_no) < p) bnl += 1.0;
	  }
	  else if (am < 1.0) {
	    g=exp(-am);
	    t=1.0;
	    for (j=0;j<=n;j++) {
	      t *= MTRnd::GenRandRes53(thr_no);
	      if (t < g) break;
	    }
	    bnl=(j <= n ? j : n);
	  }
	  else {
	    if (n != nold) {
	      en=n;
	      oldg=gamma_ln(en+1.0);
	      nold=n;
	    }
	    if (p != pold) {
	      pc=1.0-p;
	      plog=log(p);
	      pclog=log(pc);
	      pold=p;
	    }
	    sq=sqrt(2.0*am*pc);
	    do {
	      do {
	        angle=pi*MTRnd::GenRandRes53(thr_no);
	        y=tan(angle);
	        em=sq*y+am;
	      } while (em < 0.0 || em >= (en+1.0));
	      em=floor(em);
	      t=1.2*sq*(1.0+y*y)*exp(oldg-gamma_ln(em+1.0)
	                             -gamma_ln(en-em+1.0)+em*plog+(en-em)*pclog);
	    } while (MTRnd::GenRandRes53(thr_no) > t);
	    bnl=em;
	  }
	  if (p != pp) bnl=n-bnl;
	  return bnl;
	*/
	return 0
}

// Poiss returns poisson variable, as number of events in interval, with event rate (lmb = Var) plus mean
func Poiss(lmb float64, thr int) float64 {
	/*  static double sq,alxm,g,oldm=(-1.0);
	    double em,t,y;

	    if (xm < 12.0) {
	      if (xm != oldm) {
	        oldm=xm;
	        g=exp(-xm);
	      }
	      em = -1;
	      t=1.0;
	      do {
	        em += 1.0;
	        t *= MTRnd::GenRandRes53(thr_no);
	      } while (t > g);
	    }
	    else {
	      if (xm != oldm) {
	        oldm=xm;
	        sq=sqrt(2.0*xm);
	        alxm=log(xm);
	        g=xm*alxm-gamma_ln(xm+1.0);
	      }
	      do {
	        do {
	          y=tan(pi*MTRnd::GenRandRes53(thr_no));
	          em=sq*y+xm;
	        } while (em < 0.0);
	        em=floor(em);
	        t=0.9*(1.0+y*y)*exp(em*alxm-gamma_ln(em+1.0)-g);
	      } while (MTRnd::GenRandRes53(thr_no) > t);
	    }
	    return em; */
	return 0
}

// Gam represents maximum entropy distribution with two parameters: scaling parameter (Var)
// and shape parameter k (Par) plus mean
func Gam(v float64, k int, thr int) float64 {
	/*
	  if (a < 1) {
	    double u = MTRnd::GenRandRes53(thr_no);
	    return gamma_dev(1.0 + a, b, thr_no) * pow (u, 1.0 / a);
	  }

	  {
	    double x, v, u;
	    double d = a - 1.0 / 3.0;
	    double c = (1.0 / 3.0) / sqrt (d);

	    while (true) {
	      do {
	        x = gauss_dev(thr_no);
	        v = 1.0 + c * x;
	      }
	      while (v <= 0);

	      v = v * v * v;
	      u = MTRnd::GenRandRes53(thr_no);

	      if (u < 1 - 0.0331 * x * x * x * x)
	        break;

	      if (log (u) < 0.5 * x * x + d * (1 - v + log (v)))
	        break;
	    }

	    return b * d * v;
	  }
	*/
	return 0.1
}

// Gauss returns gaussian (normal) random number with given standard deviation
func Gauss(stdev float64, thr int) float64 {
	return stdev * rand.NormFloat64()
}

// Beta returns beta random number with two shape parameters a > 0 and b > 0
func Bet(a, b float64, thr int) float64 {
	x1 := Gam(a, 1.0, thr)
	x2 := Gam(b, 1.0, thr)

	return x1 / (x1 + x2)
}

//   static double UniformDen(double x, double range)
//   { double rval = 0.0; if(fabs(x) <= range) rval = 1.0 / (2.0 * range); return rval; }
//   // #CAT_Float uniform density at x with given range on either size of 0 (subtr mean from x before)
//   static double BinomDen(int n, int j, double p);
//   // #CAT_Float binomial density at j with n trials (par) each of probability p (var)
//   static double PoissonDen(int j, double l);
//   // #CAT_Float poisson density with parameter l (var)
//   static double GammaDen(int j, double l, double t);
//   // #CAT_Float gamma density at time t with given number of stages (par), lmb (var)
//   static double GaussDen(double x, double stdev);
//   // #CAT_Float gaussian (normal) density for given standard deviation (0 mean)
//   static double BetaDen(double x, double a, double b);
//   // #CAT_Float beta density at value 0 < x < 1 for shape parameters a, b
//

// BoolProp returns boolean true/false with given probability.
func BoolProb(p float64, thr int) bool {
	return (ZeroOne(thr) < p)
}
