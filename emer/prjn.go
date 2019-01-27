// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

import "github.com/emer/emergent/prjn"

// Prjn defines the basic interface for a projection which connects two layers
type Prjn interface {
	// RecvLay returns the receiving layer for this projection
	RecvLay() Layer

	// SendLay returns the sending layer for this projection
	SendLay() Layer

	// Pattern returns the pattern of connectivity for interconnecting the layers
	Pattern() prjn.Pattern

	// PrjnClass is for applying parameter styles, CSS-style -- can be space-separated multple tags
	PrjnClass() string

	// PrjnName is the automatic name of projection: RecvLay().LayName() + "Fm" + SendLay().LayName()
	PrjnName() string

	// IsOff returns true if projection or either send or recv layer has been turned Off -- for experimentation
	IsOff() bool
}
