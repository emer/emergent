// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

import (
	"testing"
)

func TestTweak(t *testing.T) {
	logvals := []float32{.1, .2, .5, 1, 1.5, 12, .015}
	logtargs := []float32{.05, .2, .1, .5, .2, 1, .5, 2, 1.2, 2, 11, 15, .012, .02}
	for i, v := range logvals {
		ps := Tweak(v, Log)
		for j, p := range ps {
			tp := logtargs[i*2+j]
			if p != tp {
				t.Errorf("log mismatch for v=%g: got %g != target %g\n", v, p, tp)
			}
		}
	}
	incrvals := []float32{.1, .3, 1.5, 25, .008}
	incrtargs := []float32{.09, .11, .2, .4, 1.4, 1.6, 24, 26, .007, .009}
	for i, v := range incrvals {
		ps := Tweak(v, Increment)
		for j, p := range ps {
			tp := incrtargs[i*2+j]
			if p != tp {
				t.Errorf("incr mismatch for v=%g: got %g != target %g\n", v, p, tp)
			}
		}
	}
	pctincrtargs := []float32{0.08, 0.12, 0.24, 0.36, 1.2, 1.8, 20, 30, 0.0064, 0.0096}
	for i, v := range incrvals {
		ps := TweakPct(v, .2)
		for j, p := range ps {
			tp := pctincrtargs[i*2+j]
			if p != tp {
				t.Errorf("incr mismatch for v=%g: got %g != target %g\n", v, p, tp)
			}
		}
	}
}
