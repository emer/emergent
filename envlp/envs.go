// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package envlp

// Envs is a map of environments organized according
// to the evaluation mode string (recommended key value)
type Envs map[string]Env

// Init initializes the map if not yet
func (es *Envs) Init() {
	if *es == nil {
		*es = make(map[string]Env)
	}
}

// Add adds Env(s), using its Mode as the key
func (es *Envs) Add(evs ...Env) {
	es.Init()
	for _, ev := range evs {
		(*es)[ev.Mode()] = ev
	}
}
