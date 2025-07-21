// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env

import (
	"fmt"

	"cogentcore.org/core/enums"
)

// Envs is a map of environments organized according
// to the evaluation mode string (recommended key value)
type Envs map[string]Env

// Init initializes the map if not yet
func (es *Envs) Init() {
	if *es == nil {
		*es = make(map[string]Env)
	}
}

// Add adds Env(s), using its Label as the key
func (es *Envs) Add(evs ...Env) {
	es.Init()
	for _, ev := range evs {
		(*es)[ev.Label()] = ev
	}
}

// ByMode returns env by Modes evaluation mode as the map key.
// returns nil if not found.
func (es *Envs) ByMode(mode enums.Enum) Env {
	return (*es)[mode.String()]
}

// ModeDi returns the string of the given mode appended with
// _di data index with leading zero.
func ModeDi(mode enums.Enum, di int) string {
	return fmt.Sprintf("%s_%02d", mode.String(), di)
}

// ByModeDi returns env by etime.Modes evaluation mode and
// data parallel index as the map key, using ModeDi function.
// returns nil if not found.
func (es *Envs) ByModeDi(mode enums.Enum, di int) Env {
	return (*es)[ModeDi(mode, di)]
}
