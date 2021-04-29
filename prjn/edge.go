// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prjn

import "github.com/goki/mat32"

// Edge returns coordinate value based on either wrapping or clipping at the edge
// and if not wrapping, if it should be clipped (ignored)
func Edge(ci, max int, wrap bool) (int, bool) {
	if ci < 0 {
		if wrap {
			return (max + ci) % max, false
		}
		return 0, true
	}
	if ci >= max {
		if wrap {
			return (ci - max) % max, false
		}
		return max - 1, true
	}
	return ci, false
}

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
