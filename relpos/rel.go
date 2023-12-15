// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package relpos defines a position relationship among layers,
in terms of X,Y width and height of layer
and associated position within a given X-Y plane,
and Z vertical stacking of layers above and below each other.
*/
package relpos

//go:generate goki generate -add-types

import (
	"goki.dev/mat32/v2"
)

// Rel defines a position relationship among layers, in terms of X,Y width and height of layer
// and associated position within a given X-Y plane,
// and Z vertical stacking of layers above and below each other.
type Rel struct { //git:add

	// spatial relationship between this layer and the other layer
	Rel Relations

	// ] horizontal (x-axis) alignment relative to other
	XAlign XAligns `viewif:"Rel=[FrontOf,Behind,Above,Below]"`

	// ] vertical (y-axis) alignment relative to other
	YAlign YAligns `viewif:"Rel=[LeftOf,RightOf,Above,Below]"`

	// name of the other layer we are in relationship to
	Other string

	// scaling factor applied to layer size for displaying
	Scale float32

	// number of unit-spaces between us
	Space float32

	// for vertical (y-axis) alignment, amount we are offset relative to perfect alignment
	XOffset float32

	// for horizontial (x-axis) alignment, amount we are offset relative to perfect alignment
	YOffset float32
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

// NewRightOf returns a RightOf relationship with default YAlign: Front alignment and given spacing
func NewRightOf(other string, space float32) Rel {
	return Rel{Rel: RightOf, Other: other, YAlign: Front, Space: space, Scale: 1}
}

// NewBehind returns a Behind relationship with default XAlign: Left alignment and given spacing
func NewBehind(other string, space float32) Rel {
	return Rel{Rel: Behind, Other: other, XAlign: Left, Space: space, Scale: 1}
}

// NewAbove returns an Above relationship with default XAlign: Left, YAlign: Front alignment
func NewAbove(other string) Rel {
	return Rel{Rel: Above, Other: other, XAlign: Left, YAlign: Front, YOffset: 1, Scale: 1}
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
type Relations int32 //enums:enum

// The relations
const (
	NoRel Relations = iota
	RightOf
	LeftOf
	Behind
	FrontOf
	Above
	Below
)

// XAligns are different horizontal alignments
type XAligns int32 //enums:enum

const (
	Left XAligns = iota
	Middle
	Right
)

// YAligns are different vertical alignments
type YAligns int32 //enums:enum

const (
	Front YAligns = iota
	Center
	Back
)
