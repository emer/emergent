// Copyright (c) 2025, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

import (
	_ "embed"

	"cogentcore.org/core/cli"
	"cogentcore.org/core/core"
	"cogentcore.org/core/tree"
)

// appIcon is the default emergent app icon, installed by default.
//
//go:embed icon.svg
var appIcon string

// Sim is an interface implemented by all sim types.
// It is parameterized by the config type C. *C must implement [Config].
//
// See [Run], [RunSim], and [Embed].
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

// Run runs a sim of the given type S with config type C. *S must implement [Sim][C]
// (interface [Sim] parameterized by config type C), and *C must implement [Config].
//
// This is a high-level helper function designed to be called as one-liner
// from the main() function of the sim's command subdirectory with package main.
// This subdirectory has the same name as the sim name itself, ex: sims/ra25
// has the package with the sim logic, and sims/ra25/ra25 has the compilable main().
//
// Run uses the config type C to make a new [Config] object and set its default values
// with [Config.Defaults].
func Run[S, C any]() {
	core.AppIcon = appIcon
	cfgC, cfg := NewConfig[C]()

	bc := cfg.AsBaseConfig()
	opts := cli.DefaultOptions(bc.Name, bc.Title)
	opts.DefaultFiles = append(opts.DefaultFiles, "config.toml")
	opts.SearchUp = true // so that the sim can be run from the command subdirectory
	opts.IncludePaths = append(opts.IncludePaths, "../configs")

	cli.Run(opts, cfgC, RunSim[S, C])
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

// Embed runs a sim with the default config, embedding it under the given parent node.
// It returns the resulting sim. *S must implement [Sim][C] (interface [Sim]
// parameterized by config type C). The configFun is an optional configuration function
// run before calling ConfigSim with the new configuration, allowing the embedding to
// set configuration and other parameters, which otherwise happen via config.toml in
// the non-embed case.
//
// The GUI is made in a frame that, when it is destroyed (as when the page of
// a content app that shows it is left), stops the sim and calls its Destroy
// method once it has stopped running, if it is a [Destroyer], to release
// resources such as those of the GPU. The sim is made once all of the sims
// that are being released are, in another goroutine if any are.
//
// See also [Run] and [RunSim].
func Embed[S, C any](parent tree.Node, configFun func(sim *S, cfg *C)) *S {
	simS := new(S)
	sf := tree.New[simFrame](parent)
	sf.sim = simS
	sf.released = make(chan struct{})
	newSim := func() {
		cfgC, cfg := NewConfig[C]()
		cfg.AsBaseConfig().GUI = true // force GUI on
		sim := any(simS).(Sim[C])
		sim.SetConfig(cfgC)
		if configFun != nil {
			configFun(simS, cfgC)
		}
		sim.ConfigSim()
		sim.Init()
		sim.ConfigGUI(sf)
	}

	// The sim is made after the sims that are being released are.
	pending := pendingReleases()
	if len(pending) == 0 {
		newSim()
		return simS
	}
	go func() {
		for _, c := range pending {
			<-c
		}
		sf.AsyncLock()
		newSim()
		sf.Update()
		sf.AsyncUnlock()
	}()
	return simS
}
