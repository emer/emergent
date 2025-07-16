// Copyright (c) 2025, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/tree"
)

// Sim is an interface implemented by all sim types.
// It is parameterized by the config type C. *C must implement [Config].
type Sim[C any] interface {

	// SetConfig sets the sim config.
	SetConfig(cfg *C)

	ConfigSim()
	Init()
	ConfigGUI(b tree.Node)

	// Body returns the [core.Body] used by the sim.
	Body() *core.Body

	RunNoGUI()
}

// RunSim runs a sim with the given config. *S must implement [Sim][C]
// (interface [Sim] parameterized by config type C).
//
// Unlike [Run], this does not handle command-line config parsing. End users
// should typically use [Run], which uses RunSim under the hood.
func RunSim[S, C any](cfg *C) error {
	simS := new(S)
	sim := any(simS).(Sim[C])

	bc := any(cfg).(Config).AsBaseConfig()

	sim.SetConfig(cfg)
	sim.ConfigSim()

	if bc.GUI {
		sim.Init()
		sim.ConfigGUI(nil)
		sim.Body().RunMainWindow()
	} else {
		sim.RunNoGUI()
	}
	return nil
}
