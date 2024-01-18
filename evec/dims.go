// Copyright 2019 The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package evec

//go:generate core generate -add-types

// Dims is a list of vector dimension (component) names
type Dims int32 //enums:enum

const (
	X Dims = iota
	Y
	Z
	W
)
