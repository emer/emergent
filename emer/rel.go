// Copyright (c) 2018, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

// Rel defines a relationship among layers for example
type Rel struct {
	Rel   Relation // LeftOf, RightOf, etc..
	Name  string   // who we are in relation to - always use names instead of pointers!
	Space int      // spacing
}
