// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

import "github.com/goki/ki/kit"

type Vec3i struct {
	X, Y, Z float32
}

// Rel defines a relationship among layers for example
type Rel struct {
	Rel    Relations `desc:"spatial relationship between this layer and the other layer"`
	Other  string    `desc:"name of the other layer we are in relationship to"`
	Space  int       `desc:"number of unit-spaces between us"`
	Offset int       `desc:"for alignment, amount we are offset relative to perfect alignment"`
	XAlign XAligns   `view:"if Rel=FrontOf,Behind,Above,Below" horizontal (x-axis) alignment relative to other"`
	YAlign YAligns   `view:"if Rel=LeftOf,RightOf,Above,Below" vertical (y-axis) alignment relative to other"`
}

// Relations are different spatial relationships (of layers)
type Relations int

//go:generate stringer -type=Relations

var KiT_Relations = kit.Enums.AddEnum(RelationsN, false, nil)

func (ev Relations) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *Relations) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }

// The relations
const (
	NoRel Relations = iota
	RightOf
	LeftOf
	Behind
	FrontOf
	Above
	Below

	RelationsN
)

// XAligns are different horizontal alignments
type XAligns int

//go:generate stringer -type=XAligns

var KiT_XAligns = kit.Enums.AddEnum(XAlignsN, false, nil)

func (ev XAligns) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *XAligns) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }

const (
	Left XAligns = iota
	Middle
	Right

	XAlignsN
)

// YAligns are different vertical alignments
type YAligns int

//go:generate stringer -type=YAligns

var KiT_YAligns = kit.Enums.AddEnum(YAlignsN, false, nil)

func (ev YAligns) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *YAligns) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }

const (
	Front YAligns = iota
	Center
	Back

	YAlignsN
)
