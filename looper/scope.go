// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import "cogentcore.org/core/enums"

// Scope is a combined Mode + Time value.
// Mode is encoded by multiples of 1000 and Time is added to that.
type Scope int

func (sc Scope) ModeTime() (mode, time int64) {
	mode = int64(sc / 1000)
	time = int64(sc % 1000)
	return
}

func ToScope(mode, time enums.Enum) Scope {
	return Scope(mode.Int64()*1000 + time.Int64())
}
