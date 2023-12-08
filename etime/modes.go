// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etime

//go:generate goki generate

//gosl: start etime

// Modes are evaluation modes (Training, Testing, etc)
type Modes int32 //enums:enum

// The evaluation modes
const (
	NoEvalMode Modes = iota

	// AllModes indicates that the log should occur over all modes present in other items.
	AllModes

	// Train is when the network is learning
	Train

	// Test is when testing, typically without learning
	Test

	// Validate is typically for a special held-out testing set
	Validate

	// Analyze is when analyzing the representations and behavior of the network
	Analyze

	// Debug is for recording info particularly useful for debugging
	Debug
)

//gosl: end etime

// ModeFromString returns Mode int value from string name
func ModeFromString(str string) Modes {
	var mode Modes
	mode.SetString(str)
	return mode
}
