// Copyright (c) 2021, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decoder

import (
	"fmt"
	"testing"
)

func TestSigmoid(t *testing.T) {
	dec := Sigmoid{}
	dec.Init(2, 2)
	dec.Lrate = .1
	trgs := []float32{0, 1}
	outs := []float32{0, 0}
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
		// fmt.Printf("%d\t%v\t%v", i, trgs, outs)
		// for j := 0; j < 2; j++ {
		// 	fmt.Printf("\t%g", dec.Units[j].Act)
		// }
		fmt.Printf("\n")
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
		dec.Train(trgs)
	}
}
