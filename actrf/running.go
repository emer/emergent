// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package actrf

import "github.com/emer/etable/etensor"

// RunningAvg computes a running-average activation-based receptive field
// for activities act relative to target activations trg (the thing we're projecting rf onto)
// accumulating into output out, with time constant tau.
// act and trg must each be 2D (flatten to 2D if not otherwise) and
// resulting out is 4D of act outer and trg inner.
func RunningAvg(out, act, trg *etensor.Float32, tau float32) {
	dt := 1 / tau
	cdt := 1 - dt
	aNy := act.Dim(0)
	aNx := act.Dim(1)
	tNy := act.Dim(0)
	tNx := act.Dim(1)
	oshp := []int{aNy, aNx, tNy, tNx}
	out.SetShape(oshp, nil, []string{"ActY", "ActX", "TrgY", "TrgX"})
	for ay := 0; ay < aNy; ay++ {
		for ax := 0; ax < aNx; ax++ {
			av := act.Value([]int{ay, ax})
			for ty := 0; ty < tNy; ty++ {
				for tx := 0; tx < tNx; tx++ {
					tv := trg.Value([]int{ty, tx})
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
