// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import "github.com/emer/emergent/etime"

func StdTrainTestPhaseCycle() *Set {
	set := NewSet()

	trn := NewStack(etime.Train.String(), etime.Run, etime.Epoch, etime.Trial, etime.Phase, etime.Cycle)
	tst := NewStack(etime.Test.String(), etime.Epoch, etime.Trial, etime.Phase, etime.Cycle)

	set.AddStack(trn)
	set.AddStack(tst)
	return set
}
