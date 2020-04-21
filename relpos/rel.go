// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package relpos defines a position relationship among layers, in terms of X,Y width and height of layer
and associated position within a given X-Y plane, and Z vertical stacking of layers above and below each other.
*/
package relpos

import (
	"github.com/goki/ki/kit"
	"github.com/goki/mat32"
)

// Rel defines a position relationship among layers, in terms of X,Y width and height of layer
// and associated position within a given X-Y plane,
// and Z vertical stacking of layers above and below each other.
type Rel struct {
	Rel     Relations `desc:"spatial relationship between this layer and the other layer"`
	XAlign  XAligns   `viewif:"Rel=FrontOf,Behind,Above,Below" desc:"horizontal (x-axis) alignment relative to other"`
	YAlign  YAligns   `viewif:"Rel=LeftOf,RightOf,Above,Below" desc:"vertical (y-axis) alignment relative to other"`
	Other   string    `desc:"name of the other layer we are in relationship to"`
	Scale   float32   `desc:"scaling factor applied to layer size for displaying"`
	Space   float32   `desc:"number of unit-spaces between us"`
	XOffset float32   `desc:"for vertical (y-axis) alignment, amount we are offset relative to perfect alignment"`
	YOffset float32   `desc:"for horizontial (x-axis) alignment, amount we are offset relative to perfect alignment"`
}

// Defaults sets default scale, space, offset values -- rel, align must be set specifically
// These are automatically applied if Scale = 0
func (rp *Rel) Defaults() {
	if rp.Scale == 0 {
		rp.Scale = 1
	}
	if rp.Space == 0 {
		rp.Space = 5
	}
}

// Pos returns the relative position compared to other position and size, based on settings
// osz and sz must both have already been scaled by relevant Scale factor
func (rp *Rel) Pos(op mat32.Vec3, osz mat32.Vec2, sz mat32.Vec2) mat32.Vec3 {
	if rp.Scale == 0 {
		rp.Defaults()
	}
	rs := op
	switch rp.Rel {
	case NoRel:
		return op
	case RightOf:
		rs.X = op.X + osz.X + rp.Space
		rs.Y = rp.AlignYPos(op.Y, osz.Y, sz.Y)
	case LeftOf:
		rs.X = op.X - sz.X - rp.Space
		rs.Y = rp.AlignYPos(op.Y, osz.Y, sz.Y)
	case Behind:
		rs.Y = op.Y + osz.Y + rp.Space
		rs.X = rp.AlignXPos(op.X, osz.X, sz.X)
	case FrontOf:
		rs.Y = op.Y - sz.Y - rp.Space
		rs.X = rp.AlignXPos(op.X, osz.X, sz.X)
	case Above:
		rs.Z += 1
		rs.X = rp.AlignXPos(op.X, osz.X, sz.X)
		rs.Y = rp.AlignYPos(op.Y, osz.Y, sz.Y)
	case Below:
		rs.Z -= 1
		rs.X = rp.AlignXPos(op.X, osz.X, sz.X)
		rs.Y = rp.AlignYPos(op.Y, osz.Y, sz.Y)
	}
	return rs
}

// AlignYPos returns the Y-axis (within-plane vertical or height) position according to alignment factors
func (rp *Rel) AlignYPos(yop, yosz, ysz float32) float32 {
	switch rp.YAlign {
	case Front:
		return yop + rp.YOffset
	case Center:
		return yop + 0.5*yosz - 0.5*ysz + rp.YOffset
	case Back:
		return yop + yosz - ysz + rp.YOffset
	}
	return yop
}

// AlignXPos returns the X-axis (within-plane horizontal or width) position according to alignment factors
func (rp *Rel) AlignXPos(xop, xosz, xsz float32) float32 {
	switch rp.XAlign {
	case Left:
		return xop + rp.XOffset
	case Middle:
		return xop + 0.5*xosz - 0.5*xsz + rp.XOffset
	case Right:
		return xop + xosz - xsz + rp.XOffset
	}
	return xop
}

// Relations are different spatial relationships (of layers)
type Relations int

//go:generate stringer -type=Relations

var KiT_Relations = kit.Enums.AddEnum(RelationsN, kit.NotBitFlag, nil)

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

var KiT_XAligns = kit.Enums.AddEnum(XAlignsN, kit.NotBitFlag, nil)

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

var KiT_YAligns = kit.Enums.AddEnum(YAlignsN, kit.NotBitFlag, nil)

func (ev YAligns) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *YAligns) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }

const (
	Front YAligns = iota
	Center
	Back

	YAlignsN
)
