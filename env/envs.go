// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env

import "github.com/emer/emergent/etime"

// Envs is a map of environments organized according
// to the evaluation mode string (recommended key value)
type Envs map[string]Env

// Init initializes the map if not yet
func (es *Envs) Init() {
	if *es == nil {
		*es = make(map[string]Env)
	}
}

// Add adds Env(s), using its Name as the key
func (es *Envs) Add(evs ...Env) {
	es.Init()
	for _, ev := range evs {
		(*es)[ev.Name()] = ev
	}
}

// ByMode returns env by etime.Modes evaluation mode as the map key.
// returns nil if not found
func (es *Envs) ByMode(mode etime.Modes) Env {
	return (*es)[mode.String()]
}
