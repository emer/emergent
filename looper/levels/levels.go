// Copyright (c) 2024, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package levels

//go:generate core generate

type Modes int32 //enums:enum
const (
	Train Modes = iota
	Test
)

type Levels int32 //enums:enum
const (
	Cycle Levels = iota
	Trial
	Epoch
	Run
)
