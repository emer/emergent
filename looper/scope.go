// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import "cogentcore.org/core/enums"

// Scope is a combined Mode + Level value.
// Mode is encoded by multiples of 1000 and Level is added to that.
type Scope int

func (sc Scope) ModeLevel() (mode, level int64) {
	mode = int64(sc / 1000)
	level = int64(sc % 1000)
	return
}

func ToScope(mode, level enums.Enum) Scope {
	return Scope(mode.Int64()*1000 + level.Int64())
}
