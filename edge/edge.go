// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package edge

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
