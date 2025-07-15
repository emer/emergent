// Copyright (c) 2025, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

import "cogentcore.org/core/cli"

// RunCmd runs a sim using the given function that runs the sim.
// This is a high-level helper function designed to be called as one-liner
// from the main() function of the sim's command subdirectory with package main.
//
// This subdirectory has the same name as the sim name itself, ex: sims/ra25
// has the package with the sim logic, and sims/ra25/ra25 has the compilable main().
//
// RunCmd uses the config type T determined from the runSim function to make a new
// [Config] object and set its default values with [Config.Defaults]. The given runSim
// function MUST take a single argument that is a pointer to the [Config] type for the
// sim.
func RunCmd[T any](runSim func(c *T) error) {
	cfgT := new(T)
	cfg := any(cfgT).(Config)
	cfg.Defaults()

	bc := cfg.AsBaseConfig()
	opts := cli.DefaultOptions(bc.Name, bc.Title)
	opts.DefaultFiles = append(opts.DefaultFiles, "config.toml")
	opts.SearchUp = true // so that the sim can be run from the command subdirectory

	cli.Run(opts, cfgT, runSim)
}
