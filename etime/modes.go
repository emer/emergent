// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etime

import (
	"github.com/goki/ki/kit"
)

// Modes are evaluation modes (Training, Testing, etc)
type Modes int32

//go:generate stringer -type=Modes

var KiT_Modes = kit.Enums.AddEnum(ModesN, kit.NotBitFlag, nil)

func (ev Modes) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *Modes) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }

// The evaluation modes
const (
	NoEvalMode Modes = iota

	// AllModes indicates that the log should occur over all modes present in other items.
	AllModes

	// Train is this a training mode for the env
	Train

	// Test is this a test mode for the env
	Test

	// Validate is this a validation mode for the env
	Validate

	// Analyze when analyzing the representations and behavior of the network
	Analyze

	ModesN
)

// ModeFromString returns Mode int value from string name
func ModeFromString(str string) Modes {
	var mode Modes
	mode.FromString(str)
	return mode
}
