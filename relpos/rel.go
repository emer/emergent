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

//go:generate core generate -add-types

import (
	"cogentcore.org/core/math32"
)

// Pos specifies the relative spatial relationship to another
// layer, which determines positioning.  Every layer except one
// "anchor" layer should be positioned relative to another,
// e.g., RightOf, Above, etc.  This provides robust positioning
// in the face of layer size changes etc.
// Layers are arranged in X-Y planes, stacked vertically along the Z axis.
type Pos struct { //git:add
	// spatial relationship between this layer and the other layer
	Rel Relations

	// ] horizontal (x-axis) alignment relative to other
	XAlign XAligns

	// ] vertical (y-axis) alignment relative to other
	YAlign YAligns

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

	// Pos is the computed position of lower-left-hand corner of layer
	// in 3D space, computed from the relation to other layer.
	Pos math32.Vector3 `edit:"-"`
}

// Defaults sets default scale, space, offset values.
// The relationship and align must be set specifically.
// These are automatically applied if Scale = 0
func (rp *Pos) Defaults() {
	if rp.Scale == 0 {
		rp.Scale = 1
	}
	if rp.Space == 0 {
		rp.Space = 5
	}
}

func (rp *Pos) ShouldDisplay(field string) bool {
	switch field {
	case "XAlign":
		return rp.Rel == FrontOf || rp.Rel == Behind || rp.Rel == Above || rp.Rel == Below
	case "YAlign":
		return rp.Rel == LeftOf || rp.Rel == RightOf || rp.Rel == Above || rp.Rel == Below
	default:
		return true
	}
}

// SetRightOf sets a RightOf relationship with default YAlign:
// Front alignment and given spacing.
func (rp *Pos) SetRightOf(other string, space float32) {
	rp.Rel = RightOf
	rp.Other = other
	rp.YAlign = Front
	rp.Space = space
	rp.Scale = 1
}

// SetBehind sets a Behind relationship with default XAlign:
// Left alignment and given spacing.
func (rp *Pos) SetBehind(other string, space float32) {
	rp.Rel = Behind
	rp.Other = other
	rp.XAlign = Left
	rp.Space = space
	rp.Scale = 1
}

// SetAbove returns an Above relationship with default XAlign:
// Left, YAlign: Front alignment
func (rp *Pos) SetAbove(other string) {
	rp.Rel = Above
	rp.Other = other
	rp.XAlign = Left
	rp.YAlign = Front
	rp.YOffset = 1
	rp.Scale = 1
}

// SetPos sets the relative position based on other layer
// position and size, using current settings.
// osz and sz must both have already been scaled by
// relevant Scale factor.
func (rp *Pos) SetPos(op math32.Vector3, osz math32.Vector2, sz math32.Vector2) {
	if rp.Scale == 0 {
		rp.Defaults()
	}
	rp.Pos = op
	switch rp.Rel {
	case NoRel:
		return
	case RightOf:
		rp.Pos.X = op.X + osz.X + rp.Space
		rp.Pos.Y = rp.AlignYPos(op.Y, osz.Y, sz.Y)
	case LeftOf:
		rp.Pos.X = op.X - sz.X - rp.Space
		rp.Pos.Y = rp.AlignYPos(op.Y, osz.Y, sz.Y)
	case Behind:
		rp.Pos.Y = op.Y + osz.Y + rp.Space
		rp.Pos.X = rp.AlignXPos(op.X, osz.X, sz.X)
	case FrontOf:
		rp.Pos.Y = op.Y - sz.Y - rp.Space
		rp.Pos.X = rp.AlignXPos(op.X, osz.X, sz.X)
	case Above:
		rp.Pos.Z += 1
		rp.Pos.X = rp.AlignXPos(op.X, osz.X, sz.X)
		rp.Pos.Y = rp.AlignYPos(op.Y, osz.Y, sz.Y)
	case Below:
		rp.Pos.Z -= 1
		rp.Pos.X = rp.AlignXPos(op.X, osz.X, sz.X)
		rp.Pos.Y = rp.AlignYPos(op.Y, osz.Y, sz.Y)
	}
}

// AlignYPos returns the Y-axis (within-plane vertical or height)
// position according to alignment factors.
func (rp *Pos) AlignYPos(yop, yosz, ysz float32) float32 {
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

// AlignXPos returns the X-axis (within-plane horizontal or width)
// position according to alignment factors.
func (rp *Pos) AlignXPos(xop, xosz, xsz float32) float32 {
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
