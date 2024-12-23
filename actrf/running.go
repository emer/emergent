// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package actrf

import "cogentcore.org/lab/tensor"

// RunningAvg computes a running-average activation-based receptive field
// for activities act relative to source activations src (the thing we're projecting rf onto)
// accumulating into output out, with time constant tau.
// act and src are projected into a 2D space (tensor.Projection2D* methods), and
// resulting out is 4D of act outer and src inner.
func RunningAvg(out *tensor.Float32, act, src tensor.Tensor, tau float32) {
	dt := 1 / tau
	cdt := 1 - dt
	aNy, aNx, _, _ := tensor.Projection2DShape(act.Shape(), false)
	tNy, tNx, _, _ := tensor.Projection2DShape(src.Shape(), false)
	oshp := []int{aNy, aNx, tNy, tNx}
	out.SetShapeSizes(oshp...)
	for ay := 0; ay < aNy; ay++ {
		for ax := 0; ax < aNx; ax++ {
			av := float32(tensor.Projection2DValue(act, false, ay, ax))
			for ty := 0; ty < tNy; ty++ {
				for tx := 0; tx < tNx; tx++ {
					tv := float32(tensor.Projection2DValue(src, false, ty, tx))
					oo := out.Shape().IndexTo1D(ay, ax, ty, tx)
					ov := out.Values[oo]
					nv := cdt*ov + dt*tv*av
					out.Values[oo] = nv
				}
			}
		}
	}
}
