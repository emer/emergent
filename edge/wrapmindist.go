// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package edge

import "github.com/goki/mat32"

// WrapMinDist returns the wrapped coordinate value that is closest to ctr
// i.e., if going out beyond max is closer, then returns that coordinate
// else if going below 0 is closer than not, then returns that coord
func WrapMinDist(ci, max, ctr float32) float32 {
	nwd := mat32.Abs(ci - ctr) // no-wrap dist
	if mat32.Abs((ci+max)-ctr) < nwd {
		return ci + max
	}
	if mat32.Abs((ci-max)-ctr) < nwd {
		return ci - max
	}
	return ci
}
