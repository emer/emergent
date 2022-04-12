// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import "github.com/emer/emergent/etime"

// StdTrainTest adds standard Train: Run, Epoch, Trial, Test: Epoch, Trial
func StdTrainTest() *Set {
	set := NewSet()

	trn := NewStack(etime.Train.String(), etime.Run, etime.Epoch, etime.Trial)
	tst := NewStack(etime.Test.String(), etime.Epoch, etime.Trial)

	set.AddStack(trn)
	set.AddStack(tst)
	return set
}
