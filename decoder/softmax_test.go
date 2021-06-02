// Copyright (c) 2021, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decoder

import (
	"testing"
)

func TestSoftMax(t *testing.T) {
	dec := SoftMax{}
	dec.Init(2, 2)
	dec.Lrate = .1
	for i := 0; i < 100; i++ {
		trg := 0
		if i%2 == 0 {
			dec.Inputs[0] = 1
			dec.Inputs[1] = 0
		} else {
			trg = 1
			dec.Inputs[0] = 0
			dec.Inputs[1] = 1
		}
		dec.Forward()
		dec.Sort()
		// fmt.Printf("%d\t%d\t%v", i, trg, dec.Sorted)
		// for j := 0; j < 2; j++ {
		// 	fmt.Printf("\t%g", dec.Units[j].Act)
		// }
		// fmt.Printf("\n")
		if i > 2 {
			if dec.Sorted[0] != trg {
				t.Errorf("err: %d\t%d\t%v\n", i, trg, dec.Sorted)
			}
		}
		dec.Train(trg)
	}
}
