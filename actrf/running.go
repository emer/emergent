// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package actrf

import "github.com/emer/etable/etensor"

// RunningAvg computes a running-average activation-based receptive field
// for activities act relative to source activations src (the thing we're projecting rf onto)
// accumulating into output out, with time constant tau.
// act and src are projected into a 2D space (etensor.Prjn2D* methods), and
// resulting out is 4D of act outer and src inner.
func RunningAvg(out *etensor.Float32, act, src etensor.Tensor, tau float32) {
	dt := 1 / tau
	cdt := 1 - dt
	aNy, aNx, _, _ := etensor.Prjn2DShape(act.ShapeObj(), false)
	tNy, tNx, _, _ := etensor.Prjn2DShape(src.ShapeObj(), false)
	oshp := []int{aNy, aNx, tNy, tNx}
	out.SetShape(oshp, nil, []string{"ActY", "ActX", "SrcY", "SrcX"})
	for ay := 0; ay < aNy; ay++ {
		for ax := 0; ax < aNx; ax++ {
			av := float32(etensor.Prjn2DVal(act, false, ay, ax))
			for ty := 0; ty < tNy; ty++ {
				for tx := 0; tx < tNx; tx++ {
					tv := float32(etensor.Prjn2DVal(src, false, ty, tx))
					oi := []int{ay, ax, ty, tx}
					oo := out.Offset(oi)
					ov := out.Values[oo]
					nv := cdt*ov + dt*tv*av
					out.Values[oo] = nv
				}
			}
		}
	}
}
