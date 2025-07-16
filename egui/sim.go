// Copyright (c) 2025, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

import "cogentcore.org/core/core"

// Sim is an interface implemented by all sim types.
// It is parameterized by the config type C. *C must implement [Config].
type Sim[C any] interface {

	// SetConfig sets the sim config.
	SetConfig(cfg *C)

	ConfigSim()
	Init()
	ConfigGUI()

	// Body returns the [core.Body] used by the sim.
	Body() *core.Body

	RunNoGUI()
}
