// Copyright (c) 2025, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

import (
	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/reflectx"
	"cogentcore.org/core/core"
	"cogentcore.org/core/system"
)

// Config is an interface implemented by all [Sim] config types.
// To implement Config, you must embed [BaseConfig]. You must
// implement [Config.Defaults] yourself.
type Config interface {

	// AsBaseConfig returns the embedded [BaseConfig].
	AsBaseConfig() *BaseConfig

	// Defaults sets default values for config fields.
	// Helper functions such as [Run], [Embed], and [NewConfig] already set defaults
	// based on struct tags, so you only need to set non-tag-based defaults here.
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

	// GPU indicates to use the GPU for computation. This is on by default, except
	// on web, where it is currently off by default.
	GPU bool
}

func (bc *BaseConfig) AsBaseConfig() *BaseConfig { return bc }

func (bc *BaseConfig) IncludesPtr() *[]string { return &bc.Includes }

// NewConfig makes a new [Config] of type *C with defaults set.
func NewConfig[C any]() (*C, Config) { //yaegi:add
	cfgC := new(C)
	cfg := any(cfgC).(Config)

	errors.Log(reflectx.SetFromDefaultTags(cfg))
	cfg.AsBaseConfig().GPU = core.TheApp.Platform() != system.Web // GPU compute not fully working on web yet
	cfg.Defaults()
	return cfgC, cfg
}
