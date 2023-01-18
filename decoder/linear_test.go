// Copyright (c) 2021, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decoder

import (
	"testing"
)

func testLinear(t *testing.T, activationFn ActivationFunc) {
	const tol = 1.0e-6

	dec := Linear{}
	dec.Init(2, 2, activationFn)
	trgs := []float32{0, 1}
	outs := []float32{0, 0}
	var lastSSE float32
	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			dec.Inputs[0] = 1
			dec.Inputs[1] = 0
			trgs[0] = 1
			trgs[1] = 0
		} else {
			dec.Inputs[0] = 0
			dec.Inputs[1] = 1
			trgs[0] = 0
			trgs[1] = 1
		}
		dec.Forward()
		dec.Output(&outs)
		if i > 2 {
			if i%2 == 0 {
				if outs[0] < outs[1] {
					t.Errorf("err: %d\t output: %g !> other: %g\n", i, outs[0], outs[1])
				}
			} else {
				if outs[1] < outs[0] {
					t.Errorf("err: %d\t output: %g !> other: %g\n", i, outs[1], outs[0])
				}
			}
		}
		sse, err := dec.Train(trgs)
		if err != nil {
			t.Error(err)
		}
		if i > 2 {
			if (sse - lastSSE) > tol {
				t.Errorf("error: %d\t sse now is *larger* than previoust: %g > %g\n", i, sse, lastSSE)
			}
		}
		lastSSE = sse
	}
}

func TestLinearIdentity(t *testing.T) {
	testLinear(t, IdentityFunc)
}

func TestLinearLogistic(t *testing.T) {
	testLinear(t, LogisticFunc)
}
