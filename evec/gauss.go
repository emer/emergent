// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package evec

import (
	"github.com/chewxy/math32"
	"github.com/goki/gi/mat32"
)

// GaussVecDistNoNorm returns the gaussian of the distance between two 2D vectors
// using given sigma standard deviation, without normalizing area under gaussian
// (i.e., max value is 1 at dist = 0)
func GaussVecDistNoNorm(a, b mat32.Vec2, sigma float32) float32 {
	dsq := a.DistToSquared(b)
	return math32.Exp((-0.5 * dsq) / (sigma * sigma))
}
