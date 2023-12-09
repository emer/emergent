// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

//go:generate goki generate

// ToolGhosting the mode enum
type ToolGhosting int32 //enums:enum

// The evaluation modes for when a tool bar can be clicked
const (
	ActiveStopped ToolGhosting = iota

	ActiveRunning

	ActiveAlways
)
