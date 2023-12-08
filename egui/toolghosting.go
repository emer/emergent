// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

import "goki.dev/ki/v2/kit"

// ToolGhosting the mode enum
type ToolGhosting int32

//go:generate stringer -type=ToolGhosting

var KiT_ToolGhosting = kit.Enums.AddEnum(ToolGhostingN, kit.BitFlag, nil)

func (ev ToolGhosting) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *ToolGhosting) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }

// The evaluation modes for when a tool bar can be clicked
const (
	ActiveStopped ToolGhosting = iota

	ActiveRunning

	ActiveAlways

	ToolGhostingN
)
