// Copyright (c) 2025, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

import "cogentcore.org/core/cli"

// Config is an interface implemented by all sim config types.
// To implement Config, you must embed [BaseConfig]. You must
// implement [Config.Defaults] yourself.
type Config interface {

	// AsBaseConfig returns the embedded [BaseConfig].
	AsBaseConfig() *BaseConfig

	// Defaults sets default values for config fields.
	Defaults()
}

// BaseConfig contains the basic configuration parameters common to all sims.
type BaseConfig struct {

	// Name is the short name of the sim.
	Name string `display:"-"`

	// Title is the longer title of the sim.
	Title string `display:"-"`

	// URL is a link to the online README or other documentation for this sim.
	URL string `display:"-"`

	// Doc is brief documentation of the sim.
	Doc string `display:"-"`

	// Includes has a list of additional config files to include.
	// After configuration, it contains list of include files added.
	Includes []string

	// GUI indicates to open the GUI. Otherwise it runs automatically and quits,
	// saving results to log files.
	GUI bool `default:"true"`

	// Debug indicates to report debugging information.
	Debug bool
}

func (bc *BaseConfig) AsBaseConfig() *BaseConfig { return bc }

func (bc *BaseConfig) IncludesPtr() *[]string { return &bc.Includes }

//////////// Run

// Run runs a sim using the given function that runs the sim.
// This is a high-level helper function designed to be called as one-liner
// from the main() function of the sim's command subdirectory with package main.
//
// This subdirectory has the same name as the sim name itself, ex: sims/ra25
// has the package with the sim logic, and sims/ra25/ra25 has the compilable main().
//
// Run uses the config type C determined from the runSim function to make a new
// [Config] object and set its default values with [Config.Defaults]. The given runSim
// function MUST take a single argument that is a pointer to the [Config] type for the
// sim. If its argument type does not implement [Config], Run will panic.
func Run[C any](runSim func(cfg *C) error) {
	cfgT := new(C)
	cfg := any(cfgT).(Config)
	cfg.Defaults()

	bc := cfg.AsBaseConfig()
	opts := cli.DefaultOptions(bc.Name, bc.Title)
	opts.DefaultFiles = append(opts.DefaultFiles, "config.toml")
	opts.SearchUp = true // so that the sim can be run from the command subdirectory

	cli.Run(opts, cfgT, runSim)
}
