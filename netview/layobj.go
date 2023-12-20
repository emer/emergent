// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import (
	"goki.dev/xyz"
)

// LayObj is the Layer 3D object within the NetView
type LayObj struct { //gti:add
	xyz.Solid

	// name of the layer we represent
	LayName string

	// our netview
	NetView *NetView `copy:"-" json:"-" xml:"-" view:"-"`
}

// LayName is the Layer name as a Text2D within the NetView
type LayName struct {
	xyz.Text2D

	// our netview
	NetView *NetView `copy:"-" json:"-" xml:"-" view:"-"`
}
